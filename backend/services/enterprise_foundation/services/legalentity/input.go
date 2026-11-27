package legalentity

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

// Dependencies mengelompokkan repositori, kebijakan domain, dan utilitas yang dibutuhkan layanan badan hukum.
type Dependencies struct {
	Transactor    domainrepo.Transactor
	LegalEntities domainrepo.LegalEntityRepository
	Addresses     domainrepo.AddressRepository
	Outbox        domainrepo.OutboxRepository
	OutboxBuilder *domainservice.OutboxBuilder
	Hierarchy     domainservice.HierarchyPolicy
	Scope         domainservice.ScopePolicy
	Clock         domainservice.Clock
	IDGen         domainservice.IDGenerator
}

// CreateInput membawa payload pembentukan badan hukum internal baru.
type CreateInput struct {
	ParentLegalEntityID      *string
	EntityCode               string
	LegalName                string
	TradingName              *string
	LegalForm                *string
	IsPrimaryOperatingEntity bool
	Status                   entity.LegalEntityStatus
	JurisdictionCountryCode  string
	NIB                      *string
	TaxIDNumber              *string
	IsPKP                    bool
	RegisteredAddressID      *string
	FiscalYearStartMonth     int16
	IncorporationDate        *time.Time
	DissolutionDate          *time.Time
}

// UpdateInput membawa pembaruan profil dan atribut kepatuhan badan hukum yang sudah ada.
type UpdateInput struct {
	ID                      string
	ParentLegalEntityID     *string
	LegalName               string
	TradingName             *string
	LegalForm               *string
	JurisdictionCountryCode string
	NIB                     *string
	TaxIDNumber             *string
	IsPKP                   bool
	RegisteredAddressID     *string
	FiscalYearStartMonth    int16
	IncorporationDate       *time.Time
	DissolutionDate         *time.Time
}
