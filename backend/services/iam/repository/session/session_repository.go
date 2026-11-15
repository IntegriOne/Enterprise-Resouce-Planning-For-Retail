package session

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.SessionRepository = (*Repository)(nil)

// Repository mengimplementasikan kontrak SessionRepository menggunakan PostgreSQL dan GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository membangun SessionRepository dengan validasi koneksi database fail-closed.
func NewRepository(db *gorm.DB) (*Repository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// SelectByID mengambil sesi login berdasarkan UUID sesi.
func (r *Repository) SelectByID(ctx context.Context, id string) (*entity.Session, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	sid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m sessionModel
	if err := db.Where("id = ?", sid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrSessionNotFound)
	}
	return sessionToDomain(&m)
}

// SelectByTokenDigest mengambil sesi login berdasarkan SHA-256 digest cookie sesi opaque.
func (r *Repository) SelectByTokenDigest(ctx context.Context, tokenDigest []byte) (*entity.Session, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	if len(tokenDigest) == 0 {
		return nil, iamerrors.NewValidation("token_digest", iamerrors.Code("iam.session.empty_token_digest"), "session token digest must not be empty")
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var m sessionModel
	if err := db.Where("token_digest = ?", tokenDigest).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrSessionNotFound)
	}
	return sessionToDomain(&m)
}

// SelectList menampilkan daftar sesi pengguna secara berhalaman berdasarkan filter UserID dan status aktif.
func (r *Repository) SelectList(ctx context.Context, filter domainrepo.SessionFilter) (domainrepo.PageResult[*entity.Session], error) {
	if r == nil {
		return domainrepo.PageResult[*entity.Session]{}, pgstore.CheckDB(nil)
	}
	f := filter.Normalize()
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return domainrepo.PageResult[*entity.Session]{}, err
	}

	q := db.Model(&sessionModel{})
	if f.UserID != "" {
		uid, err := pgstore.ValidateUUIDParam("user_id", f.UserID)
		if err != nil {
			return domainrepo.PageResult[*entity.Session]{}, err
		}
		q = q.Where("user_id = ?", uid)
	}
	if f.ActiveOnly {
		now := time.Now()
		q = q.Where("ended_at IS NULL AND absolute_expires_at > ? AND idle_expires_at > ?", now, now)
	}

	return pgstore.Paginate(q, f.Page, "authenticated_at DESC, id ASC", sessionToDomain)
}

// Insert menyimpan sesi autentikasi baru ke tabel sessions.
func (r *Repository) Insert(ctx context.Context, sess *entity.Session) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := sessionToModel(sess)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Create(&model).Error, nil)
}

// UpdateActivity memperbarui cap waktu aktivitas terakhir dan idle timeout (tertahan oleh absolute timeout) pada sesi aktif.
func (r *Repository) UpdateActivity(ctx context.Context, sessionID string, lastActiveAt, idleExpiresAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	sid, err := pgstore.ValidateUUIDParam("session_id", sessionID)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}

	lastSeen := pgstore.MustLocalNow(lastActiveAt)
	idleExp := pgstore.NormalizeLocalTime(idleExpiresAt)
	if idleExp.IsZero() {
		idleExp = lastSeen.Add(entity.SessionIdleTimeout)
	}
	res := db.Model(&sessionModel{}).
		Where("id = ? AND ended_at IS NULL", sid).
		Updates(map[string]any{
			"last_seen_at":    lastSeen,
			"idle_expires_at": gorm.Expr("LEAST(absolute_expires_at, ?)", idleExp),
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrSessionNotFound)
	}
	if res.RowsAffected == 0 {
		return checkMissingOrEndedSession(db, sid)
	}
	return nil
}

// UpdateMFAVerified menandai sesi terverifikasi 2FA sekaligus merotasi digest CSRF secara atomik.
func (r *Repository) UpdateMFAVerified(ctx context.Context, sessionID string, verifiedAt time.Time, rotatedCSRFDigest []byte) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	sid, err := pgstore.ValidateUUIDParam("session_id", sessionID)
	if err != nil {
		return err
	}
	if len(rotatedCSRFDigest) == 0 {
		return iamerrors.NewValidation("csrf_token_digest", iamerrors.Code("iam.session.empty_csrf_digest"), "csrf token digest must not be empty")
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}

	ts := pgstore.MustLocalNow(verifiedAt)
	res := db.Model(&sessionModel{}).
		Where("id = ? AND ended_at IS NULL", sid).
		Updates(map[string]any{
			"mfa_verified":      true,
			"mfa_verified_at":   ts,
			"csrf_token_digest": rotatedCSRFDigest,
			"generation":        gorm.Expr("generation + 1"),
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrSessionNotFound)
	}
	if res.RowsAffected == 0 {
		return checkMissingOrEndedSession(db, sid)
	}
	return nil
}

// UpdateEnded mengakhiri satu sesi aktif secara idempotent-safe sambil menolak sesi yang sudah berakhir sebelumnya.
func (r *Repository) UpdateEnded(ctx context.Context, sessionID string, reason entity.SessionEndReason, endedAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	sid, err := pgstore.ValidateUUIDParam("session_id", sessionID)
	if err != nil {
		return err
	}
	if !reason.IsValid() {
		return iamerrors.ErrInvalidSessionEndReason
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}

	ts := pgstore.MustLocalNow(endedAt)
	res := db.Model(&sessionModel{}).
		Where("id = ? AND ended_at IS NULL", sid).
		Updates(map[string]any{
			"ended_at":   ts,
			"end_reason": reason.String(),
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrSessionNotFound)
	}
	if res.RowsAffected == 0 {
		return checkMissingOrEndedSession(db, sid)
	}
	return nil
}

// UpdateEndedByUserID mencabut seluruh sesi aktif milik satu pengguna secara massal.
func (r *Repository) UpdateEndedByUserID(ctx context.Context, userID string, reason entity.SessionEndReason, endedAt time.Time) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return 0, err
	}
	if !reason.IsValid() {
		return 0, iamerrors.ErrInvalidSessionEndReason
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}

	ts := pgstore.MustLocalNow(endedAt)
	res := db.Model(&sessionModel{}).
		Where("user_id = ? AND ended_at IS NULL", uid).
		Updates(map[string]any{
			"ended_at":   ts,
			"end_reason": reason.String(),
		})
	if res.Error != nil {
		return 0, pgstore.MapError(res.Error, nil)
	}
	return res.RowsAffected, nil
}

// DeleteEndedBefore membersihkan arsip sesi yang telah berakhir sebelum batas waktu cutoff.
func (r *Repository) DeleteEndedBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	ts := pgstore.MustLocalNow(cutoff)
	res := db.Where("ended_at IS NOT NULL AND ended_at < ?", ts).Delete(&sessionModel{})
	if res.Error != nil {
		return 0, pgstore.MapError(res.Error, nil)
	}
	return res.RowsAffected, nil
}

func checkMissingOrEndedSession(db *gorm.DB, sid string) error {
	var row struct {
		EndedAt *time.Time `gorm:"column:ended_at"`
	}
	if err := db.Model(&sessionModel{}).Select("ended_at").Where("id = ?", sid).Take(&row).Error; err != nil {
		return pgstore.MapError(err, iamerrors.ErrSessionNotFound)
	}
	if row.EndedAt != nil {
		return iamerrors.ErrSessionAlreadyEnded
	}
	return iamerrors.ErrSessionNotFound
}
