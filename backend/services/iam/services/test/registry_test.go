package services_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	iamrepo "github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services"
)

type stubTx struct{}

func (stubTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type stubUserRepo struct{}

func (stubUserRepo) SelectByID(context.Context, string) (*entity.User, error) { return nil, nil }
func (stubUserRepo) SelectList(context.Context, domainrepo.UserFilter) (domainrepo.PageResult[*entity.User], error) {
	return domainrepo.PageResult[*entity.User]{}, nil
}
func (stubUserRepo) SelectAccessSnapshotByUserID(context.Context, string, time.Time) (*entity.AccessSnapshot, error) {
	return nil, nil
}
func (stubUserRepo) Insert(context.Context, *entity.User) error { return nil }
func (stubUserRepo) InsertWithIdentities(context.Context, *entity.User, []*entity.UserIdentity) error {
	return nil
}
func (stubUserRepo) Update(context.Context, *entity.User) error               { return nil }
func (stubUserRepo) UpdateLastLogin(context.Context, string, time.Time) error { return nil }
func (stubUserRepo) SoftDelete(context.Context, string, time.Time) error      { return nil }

type stubIdentRepo struct{}

func (stubIdentRepo) SelectByID(context.Context, string) (*entity.UserIdentity, error) {
	return nil, nil
}
func (stubIdentRepo) SelectByKindAndValue(context.Context, entity.IdentityKind, string) (*entity.UserIdentity, error) {
	return nil, nil
}
func (stubIdentRepo) FindUserIdentityByEmailKind(context.Context, string) (*entity.UserIdentity, error) {
	return nil, nil
}
func (stubIdentRepo) FindUserIdentityByUsernameKind(context.Context, string) (*entity.UserIdentity, error) {
	return nil, nil
}
func (stubIdentRepo) SelectByEmailAndUsername(context.Context, string, string) ([]*entity.UserIdentity, error) {
	return nil, nil
}
func (stubIdentRepo) SelectByUserID(context.Context, string) ([]*entity.UserIdentity, error) {
	return nil, nil
}
func (stubIdentRepo) CheckExistsByKindAndValue(context.Context, entity.IdentityKind, string) (bool, error) {
	return false, nil
}
func (stubIdentRepo) Insert(context.Context, *entity.UserIdentity) error { return nil }
func (stubIdentRepo) Update(context.Context, *entity.UserIdentity) error { return nil }
func (stubIdentRepo) Delete(context.Context, string) error               { return nil }

type stubCredRepo struct{}

func (stubCredRepo) SelectByUserID(context.Context, string) (*entity.UserCredential, error) {
	return nil, nil
}
func (stubCredRepo) Insert(context.Context, *entity.UserCredential) error { return nil }
func (stubCredRepo) Update(context.Context, *entity.UserCredential) error { return nil }
func (stubCredRepo) DeleteByUserID(context.Context, string) error         { return nil }
func (stubCredRepo) RecordFailedAttemptAtomic(context.Context, string, int, time.Duration, time.Time) (int, *time.Time, error) {
	return 0, nil, nil
}
func (stubCredRepo) ResetFailedAttemptsAtomic(context.Context, string, time.Time) error { return nil }

type stubTOTPRepo struct{}

func (stubTOTPRepo) SelectByUserID(context.Context, string) (*entity.UserTOTPFactor, error) {
	return nil, nil
}
func (stubTOTPRepo) Insert(context.Context, *entity.UserTOTPFactor) error { return nil }
func (stubTOTPRepo) Update(context.Context, *entity.UserTOTPFactor) error { return nil }
func (stubTOTPRepo) UpdateLastUsedStepIfGreater(context.Context, string, int64, time.Time) (bool, error) {
	return false, nil
}
func (stubTOTPRepo) DeleteByUserID(context.Context, string) error { return nil }

type stubRecoveryRepo struct{}

func (stubRecoveryRepo) SelectUnusedByUserID(context.Context, string) ([]*entity.UserRecoveryCode, error) {
	return nil, nil
}
func (stubRecoveryRepo) CountUnusedByUserID(context.Context, string) (int64, error) { return 0, nil }
func (stubRecoveryRepo) InsertBatch(context.Context, []*entity.UserRecoveryCode) error {
	return nil
}
func (stubRecoveryRepo) UpdateUsedAtIfUnused(context.Context, string, time.Time) (bool, error) {
	return false, nil
}
func (stubRecoveryRepo) DeleteByUserID(context.Context, string) error { return nil }

type stubTokenRepo struct{}

func (stubTokenRepo) SelectByDigestAndPurpose(context.Context, []byte, entity.TokenPurpose) (*entity.UserToken, error) {
	return nil, nil
}
func (stubTokenRepo) Insert(context.Context, *entity.UserToken) error { return nil }
func (stubTokenRepo) UpdateActiveStatusByUserAndPurpose(context.Context, string, entity.TokenPurpose, entity.TokenStatus) (int64, error) {
	return 0, nil
}
func (stubTokenRepo) UpdateUsedAtIfUnused(context.Context, string, time.Time) (bool, error) {
	return false, nil
}
func (stubTokenRepo) DeleteExpiredOrUsed(context.Context, time.Time) (int64, error) { return 0, nil }

type stubSessionRepo struct{}

func (stubSessionRepo) SelectByID(context.Context, string) (*entity.Session, error) { return nil, nil }
func (stubSessionRepo) SelectByTokenDigest(context.Context, []byte) (*entity.Session, error) {
	return nil, nil
}
func (stubSessionRepo) SelectList(context.Context, domainrepo.SessionFilter) (domainrepo.PageResult[*entity.Session], error) {
	return domainrepo.PageResult[*entity.Session]{}, nil
}
func (stubSessionRepo) Insert(context.Context, *entity.Session) error { return nil }
func (stubSessionRepo) UpdateActivity(context.Context, string, time.Time, time.Time) error {
	return nil
}
func (stubSessionRepo) UpdateMFAVerified(context.Context, string, time.Time, []byte) error {
	return nil
}
func (stubSessionRepo) UpdateEnded(context.Context, string, entity.SessionEndReason, time.Time) error {
	return nil
}
func (stubSessionRepo) UpdateEndedByUserID(context.Context, string, entity.SessionEndReason, time.Time) (int64, error) {
	return 0, nil
}
func (stubSessionRepo) DeleteEndedBefore(context.Context, time.Time) (int64, error) { return 0, nil }

type stubRoleRepo struct{}

func (stubRoleRepo) SelectByID(context.Context, string) (*entity.Role, error)   { return nil, nil }
func (stubRoleRepo) SelectByCode(context.Context, string) (*entity.Role, error) { return nil, nil }
func (stubRoleRepo) SelectList(context.Context, domainrepo.RoleFilter) (domainrepo.PageResult[*entity.Role], error) {
	return domainrepo.PageResult[*entity.Role]{}, nil
}
func (stubRoleRepo) SelectEffectiveByUserID(context.Context, string, time.Time) ([]entity.Role, error) {
	return nil, nil
}
func (stubRoleRepo) Insert(context.Context, *entity.Role) error { return nil }
func (stubRoleRepo) Update(context.Context, *entity.Role) error { return nil }
func (stubRoleRepo) Delete(context.Context, string) error       { return nil }

type stubRolePermRepo struct{}

func (stubRolePermRepo) CheckExists(context.Context, string, string) (bool, error) { return false, nil }
func (stubRolePermRepo) Insert(context.Context, *entity.RolePermission) error      { return nil }
func (stubRolePermRepo) Delete(context.Context, string, string) error              { return nil }

type stubUserRoleRepo struct{}

func (stubUserRoleRepo) SelectByUserID(context.Context, string, bool, time.Time) ([]*entity.UserRole, error) {
	return nil, nil
}
func (stubUserRoleRepo) CheckExists(context.Context, string, string, entity.Scope) (bool, error) {
	return false, nil
}
func (stubUserRoleRepo) Insert(context.Context, *entity.UserRole) error { return nil }
func (stubUserRoleRepo) DeleteByID(context.Context, string) error       { return nil }

type stubPermRepo struct{}

func (stubPermRepo) SelectByID(context.Context, string) (*entity.Permission, error) { return nil, nil }
func (stubPermRepo) SelectByCode(context.Context, string) (*entity.Permission, error) {
	return nil, nil
}
func (stubPermRepo) SelectList(context.Context, domainrepo.PermissionFilter) (domainrepo.PageResult[*entity.Permission], error) {
	return domainrepo.PageResult[*entity.Permission]{}, nil
}
func (stubPermRepo) SelectByRoleID(context.Context, string) ([]*entity.Permission, error) {
	return nil, nil
}
func (stubPermRepo) Insert(context.Context, *entity.Permission) error { return nil }
func (stubPermRepo) Update(context.Context, *entity.Permission) error { return nil }
func (stubPermRepo) Delete(context.Context, string) error             { return nil }

type stubGroupRepo struct{}

func (stubGroupRepo) SelectByID(context.Context, string) (*entity.Group, error)   { return nil, nil }
func (stubGroupRepo) SelectByCode(context.Context, string) (*entity.Group, error) { return nil, nil }
func (stubGroupRepo) SelectList(context.Context, domainrepo.GroupFilter) (domainrepo.PageResult[*entity.Group], error) {
	return domainrepo.PageResult[*entity.Group]{}, nil
}
func (stubGroupRepo) Insert(context.Context, *entity.Group) error { return nil }
func (stubGroupRepo) Update(context.Context, *entity.Group) error { return nil }
func (stubGroupRepo) Delete(context.Context, string) error        { return nil }

type stubGroupMemberRepo struct{}

func (stubGroupMemberRepo) SelectUsersByGroupID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.User], error) {
	return domainrepo.PageResult[*entity.User]{}, nil
}
func (stubGroupMemberRepo) SelectGroupsByUserID(context.Context, string) ([]*entity.Group, error) {
	return nil, nil
}
func (stubGroupMemberRepo) CheckExists(context.Context, string, string) (bool, error) {
	return false, nil
}
func (stubGroupMemberRepo) Insert(context.Context, *entity.GroupMember) error { return nil }
func (stubGroupMemberRepo) Delete(context.Context, string, string) error      { return nil }

type stubGroupRoleRepo struct{}

func (stubGroupRoleRepo) SelectByGroupID(context.Context, string, bool, time.Time) ([]*entity.GroupRole, error) {
	return nil, nil
}
func (stubGroupRoleRepo) CheckExists(context.Context, string, string, entity.Scope) (bool, error) {
	return false, nil
}
func (stubGroupRoleRepo) Insert(context.Context, *entity.GroupRole) error { return nil }
func (stubGroupRoleRepo) DeleteByID(context.Context, string) error        { return nil }

type stubSecurityRuntime struct {
	domainrepo.SecurityRuntimeRepository
}

func fullStubRepoRegistry() *iamrepo.Registry {
	return &iamrepo.Registry{
		Transactor:        stubTx{},
		Users:             stubUserRepo{},
		UserIdentities:    stubIdentRepo{},
		UserCredentials:   stubCredRepo{},
		UserTOTPFactors:   stubTOTPRepo{},
		UserRecoveryCodes: stubRecoveryRepo{},
		UserTokens:        stubTokenRepo{},
		Sessions:          stubSessionRepo{},
		Roles:             stubRoleRepo{},
		RolePermissions:   stubRolePermRepo{},
		UserRoles:         stubUserRoleRepo{},
		Permissions:       stubPermRepo{},
		Groups:            stubGroupRepo{},
		GroupMembers:      stubGroupMemberRepo{},
		GroupRoles:        stubGroupRoleRepo{},
		SecurityRuntime:   stubSecurityRuntime{},
	}
}

func TestServicesRegistry_FailClosedAndWiring(t *testing.T) {
	t.Parallel()

	if _, err := services.NewRegistry(nil, services.SecurityOptions{AllowEphemeralDevKey: true}); err == nil {
		t.Fatal("expected error for nil repository registry")
	}
	if _, err := services.NewRegistry(&iamrepo.Registry{}, services.SecurityOptions{AllowEphemeralDevKey: true}); err == nil {
		t.Fatal("expected error for incomplete repository registry")
	}

	repos := fullStubRepoRegistry()
	// Tanpa TOTPKeyring dan AllowEphemeralDevKey=false wajib gagal tertutup.
	if _, err := services.NewRegistry(repos, services.SecurityOptions{}); err == nil {
		t.Fatal("expected fail-closed error when no cipher keyring provided and ephemeral disabled")
	}
	// Dengan TOTPKeyring valid tetapi tanpa CSRFKeyring dan AllowEphemeralCSRFKey=false tetap wajib gagal tertutup.
	if _, err := services.NewRegistry(repos, services.SecurityOptions{
		TOTPActiveKeyVersion: 1,
		TOTPKeyring:          map[int16][]byte{1: bytes.Repeat([]byte{0x55}, 32)},
	}); err == nil {
		t.Fatal("expected fail-closed error when no CSRF keyring provided and ephemeral CSRF disabled")
	}

	// Dengan TOTPKeyring dan CSRFKeyring valid berhasil membentuk seluruh 5 service.
	reg, err := services.NewRegistry(repos, services.SecurityOptions{
		TOTPActiveKeyVersion: 1,
		TOTPKeyring:          map[int16][]byte{1: bytes.Repeat([]byte{0x55}, 32)},
		CSRFActiveKeyVersion: 1,
		CSRFKeyring:          map[int][]byte{1: bytes.Repeat([]byte{0x66}, 32)},
	})
	if err != nil {
		t.Fatalf("NewRegistry(valid) error = %v", err)
	}
	if err := reg.Validate(); err != nil {
		t.Fatalf("Registry.Validate() error = %v", err)
	}
}
