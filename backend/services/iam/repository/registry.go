// Package repository merakit implementasi konkret GORM PostgreSQL untuk seluruh kontrak repositori domain IAM.
package repository

import (
	"gorm.io/gorm"

	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/repository/group"
	"github.com/integrione/erp-retail/backend/services/iam/repository/outbox"
	"github.com/integrione/erp-retail/backend/services/iam/repository/permission"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
	"github.com/integrione/erp-retail/backend/services/iam/repository/provisioning"
	"github.com/integrione/erp-retail/backend/services/iam/repository/role"
	"github.com/integrione/erp-retail/backend/services/iam/repository/session"
	"github.com/integrione/erp-retail/backend/services/iam/repository/user"
)

// Registry menampung seluruh repositori domain dan unit transaksi yang siap disuntikkan ke service/usecase IAM.
type Registry struct {
	Transactor           domainrepo.Transactor
	Users                domainrepo.UserRepository
	UserIdentities       domainrepo.UserIdentityRepository
	UserCredentials      domainrepo.UserCredentialRepository
	UserTOTPFactors      domainrepo.UserTOTPFactorRepository
	UserRecoveryCodes    domainrepo.UserRecoveryCodeRepository
	UserTokens           domainrepo.UserTokenRepository
	Sessions             domainrepo.SessionRepository
	Roles                domainrepo.RoleRepository
	RolePermissions      domainrepo.RolePermissionRepository
	UserRoles            domainrepo.UserRoleRepository
	Permissions          domainrepo.PermissionRepository
	Groups               domainrepo.GroupRepository
	GroupMembers         domainrepo.GroupMemberRepository
	GroupRoles           domainrepo.GroupRoleRepository
	SecurityRuntime      domainrepo.SecurityRuntimeRepository
	EmployeeProvisioning domainrepo.EmployeeProvisioningRepository
}

// New menginisialisasi seluruh repositori PostgreSQL GORM secara fail-closed bila koneksi DB belum tersedia.
func New(db *gorm.DB) (*Registry, error) {
	tx, err := pgstore.NewTransactor(db)
	if err != nil {
		return nil, err
	}
	userRepo, err := user.NewUserRepository(db)
	if err != nil {
		return nil, err
	}
	identityRepo, err := user.NewIdentityRepository(db)
	if err != nil {
		return nil, err
	}
	credRepo, err := user.NewCredentialRepository(db)
	if err != nil {
		return nil, err
	}
	totpRepo, err := user.NewTOTPFactorRepository(db)
	if err != nil {
		return nil, err
	}
	recoveryRepo, err := user.NewRecoveryCodeRepository(db)
	if err != nil {
		return nil, err
	}
	tokenRepo, err := user.NewTokenRepository(db)
	if err != nil {
		return nil, err
	}
	sessionRepo, err := session.NewRepository(db)
	if err != nil {
		return nil, err
	}
	roleRepo, err := role.NewRoleRepository(db)
	if err != nil {
		return nil, err
	}
	rolePermRepo, err := role.NewRolePermissionRepository(db)
	if err != nil {
		return nil, err
	}
	userRoleRepo, err := role.NewUserRoleRepository(db)
	if err != nil {
		return nil, err
	}
	permRepo, err := permission.NewRepository(db)
	if err != nil {
		return nil, err
	}
	groupRepo, err := group.NewGroupRepository(db)
	if err != nil {
		return nil, err
	}
	groupMemberRepo, err := group.NewMemberRepository(db)
	if err != nil {
		return nil, err
	}
	groupRoleRepo, err := group.NewGroupRoleRepository(db)
	if err != nil {
		return nil, err
	}
	securityRuntimeRepo, err := outbox.NewRepository(db)
	if err != nil {
		return nil, err
	}

	employeeProvisioningRepo, err := provisioning.NewEmployeeRepository(db)
	if err != nil {
		return nil, err
	}

	return &Registry{
		Transactor:           tx,
		Users:                userRepo,
		UserIdentities:       identityRepo,
		UserCredentials:      credRepo,
		UserTOTPFactors:      totpRepo,
		UserRecoveryCodes:    recoveryRepo,
		UserTokens:           tokenRepo,
		Sessions:             sessionRepo,
		Roles:                roleRepo,
		RolePermissions:      rolePermRepo,
		UserRoles:            userRoleRepo,
		Permissions:          permRepo,
		Groups:               groupRepo,
		GroupMembers:         groupMemberRepo,
		GroupRoles:           groupRoleRepo,
		SecurityRuntime:      securityRuntimeRepo,
		EmployeeProvisioning: employeeProvisioningRepo,
	}, nil
}
