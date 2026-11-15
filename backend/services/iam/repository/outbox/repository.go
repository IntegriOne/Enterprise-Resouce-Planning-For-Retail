// Package outbox mengimplementasikan repositori PostgreSQL GORM untuk kontrol recovery_epoch, antrean outbox Redis, dan pembaruan epoch keamanan monotonik.
package outbox

import (
	"bytes"
	"context"
	"net/netip"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.SecurityRuntimeRepository = (*Repository)(nil)

type securityControlModel struct {
	ID            int16     `gorm:"column:id;primaryKey"`
	RecoveryEpoch int64     `gorm:"column:recovery_epoch"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (securityControlModel) TableName() string {
	return "iam_security_control"
}

type outboxEventModel struct {
	ID            string     `gorm:"column:id;type:uuid;primaryKey"`
	EventType     string     `gorm:"column:event_type"`
	AggregateType string     `gorm:"column:aggregate_type"`
	AggregateID   string     `gorm:"column:aggregate_id;type:uuid"`
	Payload       string     `gorm:"column:payload;type:jsonb"`
	Status        string     `gorm:"column:status"`
	AttemptCount  int        `gorm:"column:attempt_count"`
	NextRetryAt   time.Time  `gorm:"column:next_retry_at"`
	LastError     *string    `gorm:"column:last_error"`
	ProcessedAt   *time.Time `gorm:"column:processed_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (outboxEventModel) TableName() string {
	return "iam_outbox_events"
}

type activeSessionRecoveryRow struct {
	ID                string     `gorm:"column:id"`
	UserID            string     `gorm:"column:user_id"`
	TokenDigest       []byte     `gorm:"column:token_digest"`
	CSRFTokenDigest   []byte     `gorm:"column:csrf_token_digest"`
	Generation        int64      `gorm:"column:generation"`
	CSRFKeyVersion    int        `gorm:"column:csrf_key_version"`
	MFAVerified       bool       `gorm:"column:mfa_verified"`
	MFAVerifiedAt     *time.Time `gorm:"column:mfa_verified_at"`
	IPAddress         *string    `gorm:"column:ip_address"`
	UserAgent         *string    `gorm:"column:user_agent"`
	AuthenticatedAt   time.Time  `gorm:"column:authenticated_at"`
	LastSeenAt        time.Time  `gorm:"column:last_seen_at"`
	IdleExpiresAt     time.Time  `gorm:"column:idle_expires_at"`
	AbsoluteExpiresAt time.Time  `gorm:"column:absolute_expires_at"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
}

// Repository mengimplementasikan kontrak SecurityRuntimeRepository menggunakan PostgreSQL dan GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository membentuk Repository runtime keamanan IAM secara fail-closed bila koneksi database nil.
func NewRepository(db *gorm.DB) (*Repository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// GetRecoveryEpoch mengambil recovery_epoch global saat ini dari tabel singleton iam_security_control.
func (r *Repository) GetRecoveryEpoch(ctx context.Context) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var m securityControlModel
	if err := db.Where("id = ?", 1).Take(&m).Error; err != nil {
		return 0, pgstore.MapError(err, nil)
	}
	if m.RecoveryEpoch < 1 {
		return 1, nil
	}
	return m.RecoveryEpoch, nil
}

// IncrementRecoveryEpoch menaikkan recovery_epoch global secara atomik dan mengembalikan nilai terbarunya.
func (r *Repository) IncrementRecoveryEpoch(ctx context.Context) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var next int64
	err = db.Raw(`
		INSERT INTO iam_security_control (id, recovery_epoch, updated_at)
		VALUES (1, 2, ?)
		ON CONFLICT (id) DO UPDATE
			SET recovery_epoch = iam_security_control.recovery_epoch + 1,
			    updated_at = EXCLUDED.updated_at
		RETURNING recovery_epoch
	`, pgstore.MustLocalNow(time.Now())).Scan(&next).Error
	if err != nil {
		return 0, pgstore.MapError(err, nil)
	}
	return next, nil
}

// Insert menyimpan event outbox keamanan baru ke iam_outbox_events dalam transaksi aktif.
func (r *Repository) Insert(ctx context.Context, ev *entity.OutboxEvent) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	if ev == nil {
		return iamerrors.NewValidation("outbox_event", iamerrors.Code("iam.outbox.nil"), "outbox event must not be nil")
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	payload := string(ev.Payload)
	if strings.TrimSpace(payload) == "" {
		payload = "{}"
	}
	var lastErr *string
	if strings.TrimSpace(ev.LastError) != "" {
		msg := strings.TrimSpace(ev.LastError)
		lastErr = &msg
	}
	m := outboxEventModel{
		ID:            strings.TrimSpace(ev.ID),
		EventType:     strings.TrimSpace(ev.EventType),
		AggregateType: strings.TrimSpace(ev.AggregateType),
		AggregateID:   strings.TrimSpace(ev.AggregateID),
		Payload:       payload,
		Status:        string(ev.Status),
		AttemptCount:  ev.AttemptCount,
		NextRetryAt:   pgstore.MustLocalNow(ev.NextRetryAt),
		LastError:     lastErr,
		ProcessedAt:   pgstore.NormalizeLocalTimePtr(ev.ProcessedAt),
		CreatedAt:     pgstore.MustLocalNow(ev.CreatedAt),
		UpdatedAt:     pgstore.MustLocalNow(ev.UpdatedAt),
	}
	return pgstore.MapError(db.Create(&m).Error, nil)
}

// SelectDuePending mengambil batch event outbox berstatus pending yang siap diproyeksikan ulang.
func (r *Repository) SelectDuePending(ctx context.Context, now time.Time, limit int) ([]*entity.OutboxEvent, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	ts := pgstore.MustLocalNow(now)
	var rows []outboxEventModel
	if err := db.Where("status = ? AND next_retry_at <= ?", string(entity.OutboxStatusPending), ts).
		Order("created_at ASC, id ASC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	out := make([]*entity.OutboxEvent, 0, len(rows))
	for i := range rows {
		m := &rows[i]
		out = append(out, &entity.OutboxEvent{
			ID:            strings.ToLower(strings.TrimSpace(m.ID)),
			EventType:     m.EventType,
			AggregateType: m.AggregateType,
			AggregateID:   strings.ToLower(strings.TrimSpace(m.AggregateID)),
			Payload:       []byte(m.Payload),
			Status:        entity.OutboxStatus(m.Status),
			AttemptCount:  m.AttemptCount,
			NextRetryAt:   pgstore.NormalizeLocalTime(m.NextRetryAt),
			LastError:     pgstore.DerefString(m.LastError),
			ProcessedAt:   pgstore.NormalizeLocalTimePtr(m.ProcessedAt),
			CreatedAt:     pgstore.NormalizeLocalTime(m.CreatedAt),
			UpdatedAt:     pgstore.NormalizeLocalTime(m.UpdatedAt),
		})
	}
	return out, nil
}

// MarkEffective menandai event outbox sudah efektif terproyeksi ke Redis.
func (r *Repository) MarkEffective(ctx context.Context, id string, processedAt time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	oid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	ts := pgstore.MustLocalNow(processedAt)
	return pgstore.MapError(db.Model(&outboxEventModel{}).
		Where("id = ?", oid).
		Updates(map[string]any{
			"status":       string(entity.OutboxStatusEffective),
			"processed_at": ts,
			"last_error":   nil,
			"updated_at":   ts,
		}).Error, nil)
}

// MarkRetry menjadwalkan ulang percobaan outbox yang gagal memproyeksikan ke Redis.
func (r *Repository) MarkRetry(ctx context.Context, id string, attemptCount int, nextRetryAt time.Time, errMsg string) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	oid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	now := pgstore.MustLocalNow(time.Now())
	return pgstore.MapError(db.Model(&outboxEventModel{}).
		Where("id = ? AND status = ?", oid, string(entity.OutboxStatusPending)).
		Updates(map[string]any{
			"attempt_count": attemptCount,
			"next_retry_at": pgstore.MustLocalNow(nextRetryAt),
			"last_error":    strings.TrimSpace(errMsg),
			"updated_at":    now,
		}).Error, nil)
}

// MarkFailed menandai event outbox gagal permanen setelah melampaui ambang percobaan.
func (r *Repository) MarkFailed(ctx context.Context, id string, errMsg string) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	oid, err := pgstore.ValidateUUIDParam("id", id)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	now := pgstore.MustLocalNow(time.Now())
	return pgstore.MapError(db.Model(&outboxEventModel{}).
		Where("id = ?", oid).
		Updates(map[string]any{
			"status":     string(entity.OutboxStatusFailed),
			"last_error": strings.TrimSpace(errMsg),
			"updated_at": now,
		}).Error, nil)
}

// CountPending menghitung jumlah event outbox yang masih menunggu proyeksi ke Redis.
func (r *Repository) CountPending(ctx context.Context) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var count int64
	if err := db.Model(&outboxEventModel{}).Where("status = ?", string(entity.OutboxStatusPending)).Count(&count).Error; err != nil {
		return 0, pgstore.MapError(err, nil)
	}
	return count, nil
}

// CountFailed menghitung jumlah event outbox yang gagal permanen dan memerlukan intervensi pemulihan sebelum gate dibuka.
func (r *Repository) CountFailed(ctx context.Context) (int64, error) {
	if r == nil {
		return 0, pgstore.CheckDB(nil)
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, err
	}
	var count int64
	if err := db.Model(&outboxEventModel{}).Where("status = ?", string(entity.OutboxStatusFailed)).Count(&count).Error; err != nil {
		return 0, pgstore.MapError(err, nil)
	}
	return count, nil
}

// IncrementUserSecurityCounters menaikkan auth_epoch dan/atau perm_revision user secara atomik dalam PostgreSQL.
func (r *Repository) IncrementUserSecurityCounters(
	ctx context.Context,
	userID string,
	bumpAuthEpoch bool,
	bumpPermRevision bool,
) (int64, int64, error) {
	if r == nil {
		return 0, 0, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return 0, 0, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, 0, err
	}
	authDelta := 0
	if bumpAuthEpoch {
		authDelta = 1
	}
	permDelta := 0
	if bumpPermRevision {
		permDelta = 1
	}
	var row struct {
		AuthEpoch    int64 `gorm:"column:auth_epoch"`
		PermRevision int64 `gorm:"column:perm_revision"`
	}
	err = db.Raw(`
		UPDATE users
		SET auth_epoch = auth_epoch + ?,
		    perm_revision = perm_revision + ?,
		    updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
		RETURNING auth_epoch, perm_revision
	`, authDelta, permDelta, pgstore.MustLocalNow(time.Now()), uid).Scan(&row).Error
	if err != nil {
		return 0, 0, pgstore.MapError(err, iamerrors.ErrUserNotFound)
	}
	if row.AuthEpoch == 0 && row.PermRevision == 0 {
		return 0, 0, iamerrors.ErrUserNotFound
	}
	return row.AuthEpoch, row.PermRevision, nil
}

// BumpPermRevisionForRole menaikkan perm_revision seluruh user aktif terdampak perubahan role (langsung maupun grup).
func (r *Repository) BumpPermRevisionForRole(ctx context.Context, roleID string) ([]string, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	rid, err := pgstore.ValidateUUIDParam("role_id", roleID)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var userIDs []string
	err = db.Raw(`
		UPDATE users
		SET perm_revision = perm_revision + 1,
		    updated_at = ?
		WHERE deleted_at IS NULL
		  AND id IN (
			SELECT user_id FROM user_roles WHERE role_id = ?
			UNION
			SELECT gm.user_id
			FROM group_members gm
			JOIN group_roles gr ON gr.group_id = gm.group_id
			WHERE gr.role_id = ?
		  )
		RETURNING id::text
	`, pgstore.MustLocalNow(time.Now()), rid, rid).Scan(&userIDs).Error
	if err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return normalizeUUIDSlice(userIDs), nil
}

// BumpPermRevisionForGroup menaikkan perm_revision seluruh anggota grup yang aktif terdampak perubahan peran grup.
func (r *Repository) BumpPermRevisionForGroup(ctx context.Context, groupID string) ([]string, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	gid, err := pgstore.ValidateUUIDParam("group_id", groupID)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var userIDs []string
	err = db.Raw(`
		UPDATE users
		SET perm_revision = perm_revision + 1,
		    updated_at = ?
		WHERE deleted_at IS NULL
		  AND id IN (SELECT user_id FROM group_members WHERE group_id = ?)
		RETURNING id::text
	`, pgstore.MustLocalNow(time.Now()), gid).Scan(&userIDs).Error
	if err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	return normalizeUUIDSlice(userIDs), nil
}

// SelectActiveSessionDigestsByUserID mengambil daftar SHA-256 token digest sesi user yang belum berakhir.
func (r *Repository) SelectActiveSessionDigestsByUserID(ctx context.Context, userID string) ([][]byte, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	var rows []struct {
		TokenDigest []byte `gorm:"column:token_digest"`
	}
	if err := db.Table("sessions").
		Select("token_digest").
		Where("user_id = ? AND ended_at IS NULL", uid).
		Scan(&rows).Error; err != nil {
		return nil, pgstore.MapError(err, nil)
	}
	digests := make([][]byte, 0, len(rows))
	for _, item := range rows {
		if len(item.TokenDigest) > 0 {
			digests = append(digests, bytes.Clone(item.TokenDigest))
		}
	}
	return digests, nil
}

// SelectActiveSessionsForRecovery mengambil sesi aktif yang belum kedaluwarsa untuk pemulihan Redis saat startup/failover.
func (r *Repository) SelectActiveSessionsForRecovery(ctx context.Context, now time.Time, limit int) ([]*entity.Session, error) {
	if r == nil {
		return nil, pgstore.CheckDB(nil)
	}
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return nil, err
	}
	ts := pgstore.MustLocalNow(now)
	var rows []activeSessionRecoveryRow
	err = db.Table("sessions").
		Select("id::text, user_id::text, token_digest, csrf_token_digest, generation, csrf_key_version, mfa_verified, mfa_verified_at, host(ip_address) AS ip_address, user_agent, authenticated_at, last_seen_at, idle_expires_at, absolute_expires_at, created_at").
		Where("ended_at IS NULL AND absolute_expires_at > ? AND idle_expires_at > ?", ts, ts).
		Order("authenticated_at DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, pgstore.MapError(err, nil)
	}

	out := make([]*entity.Session, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		var ipPtr *netip.Addr
		if row.IPAddress != nil && strings.TrimSpace(*row.IPAddress) != "" {
			if parsed, parseErr := netip.ParseAddr(strings.TrimSpace(*row.IPAddress)); parseErr == nil {
				u := parsed.Unmap()
				ipPtr = &u
			}
		}
		sess, err := entity.RehydrateSession(
			strings.ToLower(strings.TrimSpace(row.ID)),
			strings.ToLower(strings.TrimSpace(row.UserID)),
			row.TokenDigest,
			row.CSRFTokenDigest,
			row.MFAVerified,
			pgstore.NormalizeLocalTimePtr(row.MFAVerifiedAt),
			ipPtr,
			pgstore.DerefString(row.UserAgent),
			pgstore.NormalizeLocalTime(row.AuthenticatedAt),
			pgstore.NormalizeLocalTime(row.LastSeenAt),
			pgstore.NormalizeLocalTime(row.IdleExpiresAt),
			pgstore.NormalizeLocalTime(row.AbsoluteExpiresAt),
			nil,
			nil,
			pgstore.NormalizeLocalTime(row.CreatedAt),
		)
		if err != nil {
			continue
		}
		sess.SetSecurityMetadata(row.Generation, row.CSRFKeyVersion)
		out = append(out, sess)
	}
	return out, nil
}

// UpdateSessionSecurityMetadata memperbarui generation, csrf_key_version, dan digest CSRF pada sesi tertentu.
func (r *Repository) UpdateSessionSecurityMetadata(
	ctx context.Context,
	sessionID string,
	generation int64,
	csrfKeyVersion int,
	csrfDigest []byte,
) error {
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
	updates := map[string]any{
		"generation":       generation,
		"csrf_key_version": csrfKeyVersion,
	}
	if len(csrfDigest) > 0 {
		updates["csrf_token_digest"] = csrfDigest
	}
	return pgstore.MapError(db.Table("sessions").Where("id = ?", sid).Updates(updates).Error, iamerrors.ErrSessionNotFound)
}

// ResolveUserRoleUserID mengambil user_id pemilik penugasan user_roles sebelum dihapus.
func (r *Repository) ResolveUserRoleUserID(ctx context.Context, assignmentID string) (string, error) {
	if r == nil {
		return "", pgstore.CheckDB(nil)
	}
	aid, err := pgstore.ValidateUUIDParam("user_role_id", assignmentID)
	if err != nil {
		return "", err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return "", err
	}
	var userID string
	err = db.Table("user_roles").Select("user_id::text").Where("id = ?", aid).Take(&userID).Error
	if err != nil {
		return "", pgstore.MapError(err, iamerrors.ErrUserRoleNotFound)
	}
	return strings.ToLower(strings.TrimSpace(userID)), nil
}

// ResolveGroupRoleGroupID mengambil group_id pemilik penugasan group_roles sebelum dihapus.
func (r *Repository) ResolveGroupRoleGroupID(ctx context.Context, assignmentID string) (string, error) {
	if r == nil {
		return "", pgstore.CheckDB(nil)
	}
	aid, err := pgstore.ValidateUUIDParam("group_role_id", assignmentID)
	if err != nil {
		return "", err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return "", err
	}
	var groupID string
	err = db.Table("group_roles").Select("group_id::text").Where("id = ?", aid).Take(&groupID).Error
	if err != nil {
		return "", pgstore.MapError(err, iamerrors.ErrGroupRoleNotFound)
	}
	return strings.ToLower(strings.TrimSpace(groupID)), nil
}

func normalizeUUIDSlice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		trimmed := strings.ToLower(strings.TrimSpace(s))
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
