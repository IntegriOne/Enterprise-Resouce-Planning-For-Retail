// Package repository merakit implementasi konkret GORM PostgreSQL untuk seluruh kontrak repositori domain Enterprise Foundation.
package repository

import (
	"gorm.io/gorm"

	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/address"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/legalentity"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/location"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/organizationunit"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/outbox"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/party"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

// Registry menampung seluruh repositori domain dan unit transaksi yang siap disuntikkan ke lapisan service Enterprise Foundation.
type Registry struct {
	Transactor        domainrepo.Transactor
	Addresses         domainrepo.AddressRepository
	LegalEntities     domainrepo.LegalEntityRepository
	Locations         domainrepo.LocationRepository
	OrganizationUnits domainrepo.OrganizationUnitRepository
	Parties           domainrepo.PartyRepository
	PartyRoles        domainrepo.PartyRoleRepository
	PartyContacts     domainrepo.PartyContactRepository
	PartyAddresses    domainrepo.PartyAddressRepository
	Outbox            domainrepo.OutboxRepository
}

// New menginisialisasi seluruh repositori PostgreSQL GORM secara fail-closed apabila koneksi DB belum tersedia.
func New(db *gorm.DB) (*Registry, error) {
	tx, err := pgstore.NewTransactor(db)
	if err != nil {
		return nil, err
	}
	addressRepo, err := address.NewRepository(db)
	if err != nil {
		return nil, err
	}
	legalEntityRepo, err := legalentity.NewRepository(db)
	if err != nil {
		return nil, err
	}
	locationRepo, err := location.NewRepository(db)
	if err != nil {
		return nil, err
	}
	orgUnitRepo, err := organizationunit.NewRepository(db)
	if err != nil {
		return nil, err
	}
	partyRepo, err := party.NewPartyRepository(db)
	if err != nil {
		return nil, err
	}
	partyRoleRepo, err := party.NewPartyRoleRepository(db)
	if err != nil {
		return nil, err
	}
	partyContactRepo, err := party.NewPartyContactRepository(db)
	if err != nil {
		return nil, err
	}
	partyAddressRepo, err := party.NewPartyAddressRepository(db)
	if err != nil {
		return nil, err
	}
	outboxRepo, err := outbox.NewRepository(db)
	if err != nil {
		return nil, err
	}

	return &Registry{
		Transactor:        tx,
		Addresses:         addressRepo,
		LegalEntities:     legalEntityRepo,
		Locations:         locationRepo,
		OrganizationUnits: orgUnitRepo,
		Parties:           partyRepo,
		PartyRoles:        partyRoleRepo,
		PartyContacts:     partyContactRepo,
		PartyAddresses:    partyAddressRepo,
		Outbox:            outboxRepo,
	}, nil
}
