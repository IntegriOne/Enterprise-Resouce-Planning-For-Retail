package party

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

// Dependencies mengelompokkan seluruh port repositori, kebijakan domain, dan utilitas yang dibutuhkan layanan mitra eksternal.
type Dependencies struct {
	Transactor     domainrepo.Transactor
	Parties        domainrepo.PartyRepository
	PartyRoles     domainrepo.PartyRoleRepository
	PartyContacts  domainrepo.PartyContactRepository
	PartyAddresses domainrepo.PartyAddressRepository
	Addresses      domainrepo.AddressRepository
	Outbox         domainrepo.OutboxRepository
	OutboxBuilder  *domainservice.OutboxBuilder
	Scope          domainservice.ScopePolicy
	PortalPolicy   domainservice.PortalAccessPolicy
	Clock          domainservice.Clock
	IDGen          domainservice.IDGenerator
}

// CreateInput membawa payload pendaftaran entitas mitra eksternal baru.
type CreateInput struct {
	PartnerCode        string
	Kind               entity.PartyKind
	DisplayName        string
	LegalName          *string
	Email              *string
	Phone              *string
	WebsiteURL         *string
	TaxIDNumber        *string
	TaxRegisteredName  *string
	IsPKP              bool
	IsTaxExempt        bool
	TaxExemptionReason *string
	TaxVerifiedAt      *time.Time
	PreferredLanguage  string
	Status             entity.PartyStatus
	Notes              *string
}

// UpdateProfileInput membawa payload pembaruan profil umum entitas mitra eksternal.
type UpdateProfileInput struct {
	ID                string
	Kind              entity.PartyKind
	DisplayName       string
	LegalName         *string
	Email             *string
	Phone             *string
	WebsiteURL        *string
	PreferredLanguage string
	Notes             *string
}

// UpdateTaxProfileInput membawa payload pembaruan profil perpajakan dan status PKP mitra.
type UpdateTaxProfileInput struct {
	ID                 string
	TaxIDNumber        *string
	TaxRegisteredName  *string
	IsPKP              bool
	IsTaxExempt        bool
	TaxExemptionReason *string
	TaxVerifiedAt      *time.Time
}

// AssignRoleInput membawa payload penambahan peran bisnis baru pada mitra eksternal.
type AssignRoleInput struct {
	PartyID       string
	Role          entity.PartyRoleKind
	ReferenceNote *string
}

// UpdateRoleInput membawa payload pembaruan catatan referensi peran bisnis mitra.
type UpdateRoleInput struct {
	ID            string
	ReferenceNote *string
}

// AddContactInput membawa payload penambahan kontak/PIC mitra dan tautan portal IAM opsional.
type AddContactInput struct {
	PartyID            string
	IAMUserID          *string
	FullName           string
	JobTitle           *string
	ContactRole        string
	Email              *string
	Phone              *string
	PortalAccessStatus entity.PortalAccessStatus
	IsPrimary          bool
}

// UpdateContactInput membawa payload pembaruan detail profil kontak/PIC mitra.
type UpdateContactInput struct {
	ID          string
	FullName    string
	JobTitle    *string
	ContactRole string
	Email       *string
	Phone       *string
	IsPrimary   bool
}

// LinkAddressInput membawa payload penautan alamat master ke mitra sesuai peruntukannya.
type LinkAddressInput struct {
	PartyID        string
	AddressID      string
	Usage          entity.AddressUsageKind
	IsPrimary      bool
	ReplacePrimary bool
}

// UpdateAddressLinkInput membawa payload pembaruan peruntukan dan penanda utama tautan alamat mitra.
type UpdateAddressLinkInput struct {
	ID             string
	Usage          entity.AddressUsageKind
	IsPrimary      bool
	ReplacePrimary bool
}
