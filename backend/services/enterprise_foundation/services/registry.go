// Package services menghimpun dan mengorkestrasi seluruh application service domain Enterprise Foundation.
package services

import (
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/pkg/uuid"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/address"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/legalentity"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/location"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/organizationunit"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/party"
)

// Options memungkinkan penyuntikan implementasi clock, generator ID, atau kebijakan domain khusus ke seluruh layanan.
type Options struct {
	Clock        domainservice.Clock
	IDGen        domainservice.IDGenerator
	Hierarchy    domainservice.HierarchyPolicy
	Scope        domainservice.ScopePolicy
	PortalPolicy domainservice.PortalAccessPolicy
}

// Registry menyatukan seluruh layanan aplikasi domain Enterprise Foundation yang siap digunakan oleh layer transport.
type Registry struct {
	Addresses         *address.Service
	LegalEntities     *legalentity.Service
	Locations         *location.Service
	OrganizationUnits *organizationunit.Service
	Parties           *party.Service
}

// NewRegistry merakit seluruh layanan domain Enterprise Foundation secara fail-closed berdasarkan Registry repositori.
func NewRegistry(repos *repository.Registry, opts Options) (*Registry, error) {
	if repos == nil {
		return nil, eferrors.NewValidation("repos", eferrors.Code("enterprise_foundation.services.nil_repositories"), "repository registry is required")
	}

	clock := opts.Clock
	if clock == nil {
		clock = domainservice.SystemClock{}
	}
	idGen := opts.IDGen
	if idGen == nil {
		idGen = uuid.NewGenerator()
	}
	hierarchy := opts.Hierarchy
	if hierarchy.MaxDepth() <= 0 {
		hierarchy = domainservice.NewDefaultHierarchyPolicy()
	}

	outboxBuilder, err := domainservice.NewOutboxBuilder(idGen, clock)
	if err != nil {
		return nil, err
	}

	addrSvc, err := address.NewService(address.Dependencies{
		Addresses: repos.Addresses,
		Clock:     clock,
		IDGen:     idGen,
	})
	if err != nil {
		return nil, err
	}

	leSvc, err := legalentity.NewService(legalentity.Dependencies{
		Transactor:    repos.Transactor,
		LegalEntities: repos.LegalEntities,
		Addresses:     repos.Addresses,
		Outbox:        repos.Outbox,
		OutboxBuilder: outboxBuilder,
		Hierarchy:     hierarchy,
		Scope:         opts.Scope,
		Clock:         clock,
		IDGen:         idGen,
	})
	if err != nil {
		return nil, err
	}

	locSvc, err := location.NewService(location.Dependencies{
		Transactor:    repos.Transactor,
		Locations:     repos.Locations,
		LegalEntities: repos.LegalEntities,
		Addresses:     repos.Addresses,
		Outbox:        repos.Outbox,
		OutboxBuilder: outboxBuilder,
		Hierarchy:     hierarchy,
		Scope:         opts.Scope,
		Clock:         clock,
		IDGen:         idGen,
	})
	if err != nil {
		return nil, err
	}

	orgSvc, err := organizationunit.NewService(organizationunit.Dependencies{
		Transactor:        repos.Transactor,
		OrganizationUnits: repos.OrganizationUnits,
		LegalEntities:     repos.LegalEntities,
		Locations:         repos.Locations,
		Outbox:            repos.Outbox,
		OutboxBuilder:     outboxBuilder,
		Hierarchy:         hierarchy,
		Scope:             opts.Scope,
		Clock:             clock,
		IDGen:             idGen,
	})
	if err != nil {
		return nil, err
	}

	partySvc, err := party.NewService(party.Dependencies{
		Transactor:     repos.Transactor,
		Parties:        repos.Parties,
		PartyRoles:     repos.PartyRoles,
		PartyContacts:  repos.PartyContacts,
		PartyAddresses: repos.PartyAddresses,
		Addresses:      repos.Addresses,
		Outbox:         repos.Outbox,
		OutboxBuilder:  outboxBuilder,
		Scope:          opts.Scope,
		PortalPolicy:   opts.PortalPolicy,
		Clock:          clock,
		IDGen:          idGen,
	})
	if err != nil {
		return nil, err
	}

	reg := &Registry{
		Addresses:         addrSvc,
		LegalEntities:     leSvc,
		Locations:         locSvc,
		OrganizationUnits: orgSvc,
		Parties:           partySvc,
	}
	if err := reg.Validate(); err != nil {
		return nil, err
	}
	return reg, nil
}

// Validate memastikan seluruh lima layanan inti pada Registry telah terinisialisasi secara lengkap.
func (r *Registry) Validate() error {
	if r == nil {
		return eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.services.nil_registry"), "service registry is nil")
	}
	checks := []struct {
		ok  bool
		msg string
	}{
		{r.Addresses != nil, "address service is not initialized"},
		{r.LegalEntities != nil, "legal entity service is not initialized"},
		{r.Locations != nil, "location service is not initialized"},
		{r.OrganizationUnits != nil, "organization unit service is not initialized"},
		{r.Parties != nil, "party service is not initialized"},
	}
	for _, c := range checks {
		if !c.ok {
			return eferrors.Wrap(eferrors.ErrUnavailable, eferrors.KindUnavailable, eferrors.Code("enterprise_foundation.services.incomplete_registry"), c.msg)
		}
	}
	return nil
}
