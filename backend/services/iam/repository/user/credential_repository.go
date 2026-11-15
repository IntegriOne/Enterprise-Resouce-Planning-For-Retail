package user

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

var _ domainrepo.UserCredentialRepository = (*CredentialRepository)(nil)

// CredentialRepository mengimplementasikan kontrak UserCredentialRepository menggunakan PostgreSQL dan GORM.
type CredentialRepository struct {
	db *gorm.DB
}

// NewCredentialRepository membangun CredentialRepository dengan validasi koneksi database fail-closed.
func NewCredentialRepository(db *gorm.DB) (*CredentialRepository, error) {
	if err := pgstore.CheckDB(db); err != nil {
		return nil, err
	}
	return &CredentialRepository{db: db}, nil
}

// SelectByUserID mengambil kredensial kata sandi aktif berdasarkan UUID pengguna.
func (r *CredentialRepository) SelectByUserID(ctx context.Context, userID string) (*entity.UserCredential, error) {
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
	var m userCredentialModel
	if err := db.Where("user_id = ?", uid).Take(&m).Error; err != nil {
		return nil, pgstore.MapError(err, iamerrors.ErrCredentialNotFound)
	}
	return credentialToDomain(&m)
}

// Insert menyimpan kredensial kata sandi pengguna baru tanpa mengisi kolom id agar dihasilkan oleh gen_random_uuid().
func (r *CredentialRepository) Insert(ctx context.Context, credential *entity.UserCredential) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := credentialToModel(credential)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	return pgstore.MapError(db.Omit("id").Create(&model).Error, nil)
}

// Update memperbarui hash kata sandi, status rotasi wajib, penghitung gagal login, dan kunci sementara.
func (r *CredentialRepository) Update(ctx context.Context, credential *entity.UserCredential) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	model, err := credentialToModel(credential)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	res := db.Model(&userCredentialModel{}).
		Where("user_id = ?", model.UserID).
		Updates(map[string]any{
			"password_hash":         model.PasswordHash,
			"password_changed_at":   model.PasswordChangedAt,
			"must_change_password":  model.MustChangePassword,
			"failed_login_attempts": model.FailedLoginAttempts,
			"locked_until":          model.LockedUntil,
			"updated_at":            model.UpdatedAt,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrCredentialNotFound)
	}
	if res.RowsAffected == 0 {
		return iamerrors.ErrCredentialNotFound
	}
	return nil
}

// RecordFailedAttemptAtomic menambah penghitung gagal login secara atomik di PostgreSQL untuk mencegah race condition lost-update.
func (r *CredentialRepository) RecordFailedAttemptAtomic(
	ctx context.Context,
	userID string,
	maxAttempts int,
	lockDuration time.Duration,
	now time.Time,
) (int, *time.Time, error) {
	if r == nil {
		return 0, nil, pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return 0, nil, err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return 0, nil, err
	}

	ts := pgstore.MustLocalNow(now)
	var lockExpr any
	if maxAttempts > 0 && lockDuration > 0 {
		lockUntil := pgstore.MustLocalNow(ts.Add(lockDuration))
		lockExpr = gorm.Expr(
			"CASE WHEN (CASE WHEN locked_until IS NOT NULL AND locked_until <= ? THEN 1 ELSE failed_login_attempts + 1 END) >= ? THEN ? WHEN locked_until IS NOT NULL AND locked_until <= ? THEN NULL ELSE locked_until END",
			ts, maxAttempts, lockUntil, ts,
		)
	} else {
		lockExpr = gorm.Expr(
			"CASE WHEN locked_until IS NOT NULL AND locked_until <= ? THEN NULL ELSE locked_until END",
			ts,
		)
	}

	var updated userCredentialModel
	res := db.Model(&updated).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "failed_login_attempts"}, {Name: "locked_until"}}}).
		Where("user_id = ?", uid).
		Updates(map[string]any{
			"failed_login_attempts": gorm.Expr("CASE WHEN locked_until IS NOT NULL AND locked_until <= ? THEN 1 ELSE failed_login_attempts + 1 END", ts),
			"locked_until":          lockExpr,
			"updated_at":            ts,
		})
	if res.Error != nil {
		return 0, nil, pgstore.MapError(res.Error, iamerrors.ErrCredentialNotFound)
	}
	if res.RowsAffected == 0 {
		return 0, nil, iamerrors.ErrCredentialNotFound
	}
	var untilPtr *time.Time
	if updated.LockedUntil != nil {
		normalized := updated.LockedUntil.Local()
		untilPtr = &normalized
	}
	return updated.FailedLoginAttempts, untilPtr, nil
}

// ResetFailedAttemptsAtomic membersihkan penghitung kegagalan login dan kunci sementara secara atomik bila masih ada.
func (r *CredentialRepository) ResetFailedAttemptsAtomic(ctx context.Context, userID string, now time.Time) error {
	if r == nil {
		return pgstore.CheckDB(nil)
	}
	uid, err := pgstore.ValidateUUIDParam("user_id", userID)
	if err != nil {
		return err
	}
	db, err := pgstore.DBFromContext(ctx, r.db)
	if err != nil {
		return err
	}
	ts := pgstore.MustLocalNow(now)
	res := db.Model(&userCredentialModel{}).
		Where("user_id = ? AND (failed_login_attempts > 0 OR locked_until IS NOT NULL)", uid).
		Updates(map[string]any{
			"failed_login_attempts": 0,
			"locked_until":          nil,
			"updated_at":            ts,
		})
	if res.Error != nil {
		return pgstore.MapError(res.Error, iamerrors.ErrCredentialNotFound)
	}
	return nil
}
