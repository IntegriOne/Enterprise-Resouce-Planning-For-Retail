package group

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
)

// Dependencies mengelompokkan port persistensi dan utilitas domain untuk layanan grup IAM.
type Dependencies struct {
	Transactor      domainrepo.Transactor
	Groups          domainrepo.GroupRepository
	GroupMembers    domainrepo.GroupMemberRepository
	GroupRoles      domainrepo.GroupRoleRepository
	Roles           domainrepo.RoleRepository
	Users           domainrepo.UserRepository
	SecurityRuntime domainrepo.SecurityRuntimeRepository
	Clock           domainservice.Clock
	IDGen           domainservice.IDGenerator
	OutboxCoord     *redisstore.Coordinator
	Projector       domainservice.RuntimeSecurityProjector
}

// CreateInput membawa data pembentukan grup pengguna baru.
type CreateInput struct {
	Code        string
	Name        string
	Description string
}

// UpdateInput membawa pembaruan nama dan deskripsi grup pengguna.
type UpdateInput struct {
	ID          string
	Name        string
	Description string
}

// AssignGroupRoleInput membawa data penugasan peran berscope ke grup pengguna.
type AssignGroupRoleInput struct {
	GroupID   string
	RoleID    string
	ScopeType entity.ScopeType
	ScopeID   *string
	// GrantedBy dipertahankan untuk kompatibilitas pemanggil lama; service memakai principal dari context.
	GrantedBy *string
	ExpiresAt *time.Time
}
