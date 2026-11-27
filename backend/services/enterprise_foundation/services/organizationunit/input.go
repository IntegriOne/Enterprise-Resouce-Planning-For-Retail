package organizationunit

import (
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

// Dependencies mengelompokkan repositori, kebijakan domain, dan utilitas yang dibutuhkan layanan unit organisasi.
type Dependencies struct {
	Transactor        domainrepo.Transactor
	OrganizationUnits domainrepo.OrganizationUnitRepository
	LegalEntities     domainrepo.LegalEntityRepository
	Locations         domainrepo.LocationRepository
	Outbox            domainrepo.OutboxRepository
	OutboxBuilder     *domainservice.OutboxBuilder
	Hierarchy         domainservice.HierarchyPolicy
	Scope             domainservice.ScopePolicy
	Clock             domainservice.Clock
	IDGen             domainservice.IDGenerator
}

// CreateInput membawa payload pembentukan unit organisasi internal baru.
type CreateInput struct {
	LegalEntityID     string
	ParentUnitID      *string
	PrimaryLocationID *string
	UnitCode          string
	Name              string
	Kind              entity.OrgUnitKind
	CostCenterCode    *string
	HeadPersonID      *string
}

// UpdateInput membawa payload pembaruan struktur dan relasi unit organisasi yang sudah terdaftar.
type UpdateInput struct {
	ID                 string
	ParentUnitID       *string
	PrimaryLocationID  *string
	Name               string
	Kind               entity.OrgUnitKind
	CostCenterCode     *string
	HeadPersonID       *string
	PreserveHeadPerson bool
}
