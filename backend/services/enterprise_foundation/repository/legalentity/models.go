// Package legalentity mengimplementasikan repositori PostgreSQL berbasis GORM untuk tabel legal_entities.
package legalentity

import "time"

type legalEntityModel struct {
	ID                       string     `gorm:"column:id;type:uuid;primaryKey"`
	ParentLegalEntityID      *string    `gorm:"column:parent_legal_entity_id;type:uuid"`
	EntityCode               string     `gorm:"column:entity_code"`
	LegalName                string     `gorm:"column:legal_name"`
	TradingName              *string    `gorm:"column:trading_name"`
	LegalForm                *string    `gorm:"column:legal_form"`
	IsPrimaryOperatingEntity bool       `gorm:"column:is_primary_operating_entity"`
	Status                   string     `gorm:"column:status"`
	JurisdictionCountryCode  string     `gorm:"column:jurisdiction_country_code"`
	NIB                      *string    `gorm:"column:nib"`
	TaxIDNumber              *string    `gorm:"column:tax_id_number"`
	IsPKP                    bool       `gorm:"column:is_pkp"`
	RegisteredAddressID      *string    `gorm:"column:registered_address_id;type:uuid"`
	FiscalYearStartMonth     int16      `gorm:"column:fiscal_year_start_month"`
	IncorporationDate        *time.Time `gorm:"column:incorporation_date;type:date"`
	DissolutionDate          *time.Time `gorm:"column:dissolution_date;type:date"`
	CreatedAt                time.Time  `gorm:"column:created_at"`
	UpdatedAt                time.Time  `gorm:"column:updated_at"`
	DeletedAt                *time.Time `gorm:"column:deleted_at"`
}

func (legalEntityModel) TableName() string {
	return "legal_entities"
}
