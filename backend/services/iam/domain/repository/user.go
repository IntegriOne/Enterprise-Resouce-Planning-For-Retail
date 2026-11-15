package repository

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
)

// UserRepository mendefinisikan operasi baca dan tulis untuk tabel users serta snapshot akses teragregasi.
type UserRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.User, error)
	SelectList(ctx context.Context, filter UserFilter) (PageResult[*entity.User], error)
	SelectAccessSnapshotByUserID(ctx context.Context, userID string, now time.Time) (*entity.AccessSnapshot, error)
	Insert(ctx context.Context, user *entity.User) error
	// InsertWithIdentities memverifikasi ketiadaan duplikasi email dan username di user_identities sebelum menyimpan pengguna beserta identitasnya secara atomik.
	InsertWithIdentities(ctx context.Context, user *entity.User, identities []*entity.UserIdentity) error
	Update(ctx context.Context, user *entity.User) error
	UpdateLastLogin(ctx context.Context, userID string, loginAt time.Time) error
	SoftDelete(ctx context.Context, id string, deletedAt time.Time) error
}

// UserIdentityRepository mendefinisikan operasi baca dan tulis untuk tabel user_identities.
type UserIdentityRepository interface {
	SelectByID(ctx context.Context, id string) (*entity.UserIdentity, error)
	SelectByKindAndValue(ctx context.Context, kind entity.IdentityKind, value string) (*entity.UserIdentity, error)
	// FindUserIdentityByEmailKind mencari identitas pengguna berdasarkan kind='email' dan nilai email terstandardisasi.
	FindUserIdentityByEmailKind(ctx context.Context, email string) (*entity.UserIdentity, error)
	// FindUserIdentityByUsernameKind mencari identitas pengguna berdasarkan kind='username' dan nilai username terstandardisasi.
	FindUserIdentityByUsernameKind(ctx context.Context, username string) (*entity.UserIdentity, error)
	// SelectByEmailAndUsername mencari identitas email dan/atau username yang sudah ada di user_identities untuk mencegah duplikasi sebelum pembuatan pengguna.
	SelectByEmailAndUsername(ctx context.Context, email, username string) ([]*entity.UserIdentity, error)
	SelectByUserID(ctx context.Context, userID string) ([]*entity.UserIdentity, error)
	CheckExistsByKindAndValue(ctx context.Context, kind entity.IdentityKind, value string) (bool, error)
	Insert(ctx context.Context, identity *entity.UserIdentity) error
	Update(ctx context.Context, identity *entity.UserIdentity) error
	Delete(ctx context.Context, id string) error
}
