package location

import (
	"encoding/json"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

// Dependencies mengelompokkan repositori, kebijakan domain, dan utilitas yang dibutuhkan layanan lokasi fisik.
type Dependencies struct {
	Transactor    domainrepo.Transactor
	Locations     domainrepo.LocationRepository
	LegalEntities domainrepo.LegalEntityRepository
	Addresses     domainrepo.AddressRepository
	Outbox        domainrepo.OutboxRepository
	OutboxBuilder *domainservice.OutboxBuilder
	Hierarchy     domainservice.HierarchyPolicy
	Scope         domainservice.ScopePolicy
	Clock         domainservice.Clock
	IDGen         domainservice.IDGenerator
}

// CreateInput membawa payload pembentukan lokasi fisik perusahaan baru.
type CreateInput struct {
	LegalEntityID         string
	ParentLocationID      *string
	AddressID             *string
	LocationCode          string
	Name                  string
	Kind                  entity.LocationKind
	Phone                 *string
	Timezone              string
	OperationalAttributes json.RawMessage
}

// UpdateInput membawa payload pembaruan profil dan relasi lokasi fisik internal.
type UpdateInput struct {
	ID                    string
	ParentLocationID      *string
	AddressID             *string
	Name                  string
	Kind                  entity.LocationKind
	Phone                 *string
	Timezone              string
	OperationalAttributes json.RawMessage
}
