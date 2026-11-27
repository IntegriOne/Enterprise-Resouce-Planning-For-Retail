// Package party mengimplementasikan repositori PostgreSQL berbasis GORM untuk tabel parties, party_roles, party_contacts, dan party_addresses.
package party

import "time"

type partyModel struct {
	ID                 string     `gorm:"column:id;type:uuid;primaryKey"`
	PartnerCode        string     `gorm:"column:partner_code"`
	Kind               string     `gorm:"column:party_kind"`
	DisplayName        string     `gorm:"column:display_name"`
	LegalName          *string    `gorm:"column:legal_name"`
	Email              *string    `gorm:"column:email"`
	Phone              *string    `gorm:"column:phone"`
	WebsiteURL         *string    `gorm:"column:website_url"`
	TaxIDNumber        *string    `gorm:"column:tax_id_number"`
	TaxRegisteredName  *string    `gorm:"column:tax_registered_name"`
	IsPKP              bool       `gorm:"column:is_pkp"`
	IsTaxExempt        bool       `gorm:"column:is_tax_exempt"`
	TaxExemptionReason *string    `gorm:"column:tax_exemption_reason"`
	TaxVerifiedAt      *time.Time `gorm:"column:tax_verified_at"`
	PreferredLanguage  string     `gorm:"column:preferred_language"`
	Status             string     `gorm:"column:status"`
	Notes              *string    `gorm:"column:notes"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
	DeletedAt          *time.Time `gorm:"column:deleted_at"`
}

func (partyModel) TableName() string {
	return "parties"
}

type partyRoleModel struct {
	ID            string    `gorm:"column:id;type:uuid;primaryKey"`
	PartyID       string    `gorm:"column:party_id;type:uuid"`
	Role          string    `gorm:"column:role"`
	IsActive      bool      `gorm:"column:is_active"`
	ReferenceNote *string   `gorm:"column:reference_note"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (partyRoleModel) TableName() string {
	return "party_roles"
}

type partyContactModel struct {
	ID                 string     `gorm:"column:id;type:uuid;primaryKey"`
	PartyID            string     `gorm:"column:party_id;type:uuid"`
	IAMUserID          *string    `gorm:"column:iam_user_id;type:uuid"`
	FullName           string     `gorm:"column:full_name"`
	JobTitle           *string    `gorm:"column:job_title"`
	ContactRole        string     `gorm:"column:contact_role"`
	Email              *string    `gorm:"column:email"`
	Phone              *string    `gorm:"column:phone"`
	PortalAccessStatus string     `gorm:"column:portal_access_status"`
	IsPrimary          bool       `gorm:"column:is_primary"`
	IsActive           bool       `gorm:"column:is_active"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
	DeletedAt          *time.Time `gorm:"column:deleted_at"`
}

func (partyContactModel) TableName() string {
	return "party_contacts"
}

type partyAddressModel struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	PartyID   string    `gorm:"column:party_id;type:uuid"`
	AddressID string    `gorm:"column:address_id;type:uuid"`
	Usage     string    `gorm:"column:usage"`
	IsPrimary bool      `gorm:"column:is_primary"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (partyAddressModel) TableName() string {
	return "party_addresses"
}
