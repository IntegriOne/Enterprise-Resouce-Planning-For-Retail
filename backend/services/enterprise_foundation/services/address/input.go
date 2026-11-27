package address

import (
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

// Dependencies mengelompokkan port persistensi dan utilitas domain untuk layanan master alamat.
type Dependencies struct {
	Addresses domainrepo.AddressRepository
	Clock     domainservice.Clock
	IDGen     domainservice.IDGenerator
}

// CreateInput membawa data pembentukan entitas alamat master baru.
type CreateInput struct {
	Label            *string
	AddressLine1     string
	AddressLine2     *string
	City             string
	Province         *string
	PostalCode       *string
	CountryCode      string
	FormattedAddress *string
	Latitude         *float64
	Longitude        *float64
}

// UpdateInput membawa pembaruan seluruh atribut master alamat yang sudah terdaftar.
type UpdateInput struct {
	ID               string
	Label            *string
	AddressLine1     string
	AddressLine2     *string
	City             string
	Province         *string
	PostalCode       *string
	CountryCode      string
	FormattedAddress *string
	Latitude         *float64
	Longitude        *float64
}
