package role

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
)

// Dependencies mengelompokkan port persistensi dan utilitas domain yang dibutuhkan layanan peran.
type Dependencies struct {
	Transactor      domainrepo.Transactor
	Roles           domainrepo.RoleRepository
	RolePermissions domainrepo.RolePermissionRepository
	Permissions     domainrepo.PermissionRepository
	Users           domainrepo.UserRepository
	UserRoles       domainrepo.UserRoleRepository
	SecurityRuntime domainrepo.SecurityRuntimeRepository
	Clock           domainservice.Clock
	IDGen           domainservice.IDGenerator
	Redirect        domainservice.RedirectResolver
	OutboxCoord     *redisstore.Coordinator
	Projector       domainservice.RuntimeSecurityProjector
}

// CreateInput membawa data pembentukan peran RBAC baru.
type CreateInput struct {
	Code                string
	Name                string
	Description         string
	ManagementDomain    string
	DefaultRedirectPath string
	RedirectPriority    int
	IsSystem            bool
}

// UpdateInput membawa pembaruan metadata peran yang sudah ada.
type UpdateInput struct {
	ID                  string
	Name                string
	Description         string
	ManagementDomain    string
	DefaultRedirectPath string
	RedirectPriority    int
}

// AssignUserRoleInput membawa data penugasan peran berscope ke akun pengguna.
type AssignUserRoleInput struct {
	UserID    string
	RoleID    string
	ScopeType entity.ScopeType
	ScopeID   *string
	GrantedBy *string
	ExpiresAt *time.Time
}
