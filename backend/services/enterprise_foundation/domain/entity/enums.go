// Package entity mendefinisikan entitas domain murni, enum tervalidasi, dan invarian Enterprise Foundation tanpa ketergantungan ORM atau transport.
package entity

import (
	"fmt"
	"strings"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// Konstanta batas panjang kolom dan nilai baku yang selaras 1:1 dengan skema PostgreSQL Enterprise Foundation.
const (
	// DefaultCountryCode adalah kode negara ISO 3166-1 alpha-2 bawaan skema.
	DefaultCountryCode = "ID"
	// DefaultTimezone adalah zona waktu operasional bawaan untuk lokasi baru.
	DefaultTimezone = "Asia/Jakarta"
	// DefaultPreferredLanguage adalah bahasa komunikasi bawaan mitra eksternal.
	DefaultPreferredLanguage = "id"
	// DefaultContactRole adalah peran kontak bawaan saat belum diklasifikasikan khusus.
	DefaultContactRole = "general"
	// DefaultFiscalYearStartMonth adalah bulan awal tahun buku bawaan (Januari).
	DefaultFiscalYearStartMonth int16 = 1

	// MaxAddressLabelLength membatasi panjang kolom addresses.label.
	MaxAddressLabelLength = 120
	// MaxAddressLineLength membatasi panjang kolom addresses.address_line_1 dan address_line_2.
	MaxAddressLineLength = 240
	// MaxCityLength membatasi panjang kolom addresses.city.
	MaxCityLength = 120
	// MaxProvinceLength membatasi panjang kolom addresses.province.
	MaxProvinceLength = 120
	// MaxPostalCodeLength membatasi panjang kolom addresses.postal_code.
	MaxPostalCodeLength = 20
	// CountryCodeLength menetapkan panjang wajib kode negara ISO alpha-2.
	CountryCodeLength = 2

	// MaxCodeLength membatasi panjang kolom kode entitas bisnis (entity_code, location_code, unit_code, partner_code, nib, phone, contact_role).
	MaxCodeLength = 60
	// MaxLegalNameLength membatasi panjang kolom legal_name, trading_name, display_name, dan tax_registered_name.
	MaxLegalNameLength = 240
	// MaxLegalFormLength membatasi panjang kolom legal_entities.legal_form.
	MaxLegalFormLength = 40
	// MaxTaxIDLength membatasi panjang kolom NPWP/Tax ID (tax_id_number).
	MaxTaxIDLength = 32

	// MaxLocationNameLength membatasi panjang kolom locations.name.
	MaxLocationNameLength = 200
	// MaxTimezoneLength membatasi panjang kolom locations.timezone.
	MaxTimezoneLength = 80
	// MaxOrgUnitNameLength membatasi panjang kolom organization_units.name.
	MaxOrgUnitNameLength = 180

	// MaxEmailLength membatasi panjang kolom email pada parties dan party_contacts.
	MaxEmailLength = 255
	// MaxWebsiteURLLength membatasi panjang kolom parties.website_url.
	MaxWebsiteURLLength = 500
	// MaxPreferredLanguageLength membatasi panjang kolom parties.preferred_language.
	MaxPreferredLanguageLength = 20
	// MaxReferenceNoteLength membatasi panjang kolom party_roles.reference_note.
	MaxReferenceNoteLength = 200
	// MaxContactFullNameLength membatasi panjang kolom party_contacts.full_name.
	MaxContactFullNameLength = 200
	// MaxJobTitleLength membatasi panjang kolom party_contacts.job_title.
	MaxJobTitleLength = 120

	// MaxOutboxEventTypeLength membatasi panjang kolom foundation_outbox_events.event_type.
	MaxOutboxEventTypeLength = 80
	// MaxOutboxAggregateTypeLength membatasi panjang kolom foundation_outbox_events.aggregate_type.
	MaxOutboxAggregateTypeLength = 40

	// MinLatitude dan MaxLatitude menetapkan rentang valid garis lintang bumi.
	MinLatitude = -90.0
	MaxLatitude = 90.0
	// MinLongitude dan MaxLongitude menetapkan rentang valid garis bujur bumi.
	MinLongitude = -180.0
	MaxLongitude = 180.0
)

// LegalEntityStatus merepresentasikan siklus hidup badan hukum sesuai enum PostgreSQL legal_entity_status.
type LegalEntityStatus string

const (
	// LegalEntityStatusPending menandakan badan hukum baru belum diaktifkan untuk transaksi.
	LegalEntityStatusPending LegalEntityStatus = "pending"
	// LegalEntityStatusActive menandakan badan hukum aktif dan sah digunakan lintas modul.
	LegalEntityStatusActive LegalEntityStatus = "active"
	// LegalEntityStatusSuspended menandakan badan hukum dibekukan sementara.
	LegalEntityStatusSuspended LegalEntityStatus = "suspended"
	// LegalEntityStatusDissolved menandakan badan hukum telah dibubarkan secara permanen.
	LegalEntityStatusDissolved LegalEntityStatus = "dissolved"
)

// String mengembalikan representasi string enum legal_entity_status.
func (s LegalEntityStatus) String() string {
	return string(s)
}

// IsValid memeriksa apakah LegalEntityStatus sesuai enum PostgreSQL.
func (s LegalEntityStatus) IsValid() bool {
	switch s {
	case LegalEntityStatusPending, LegalEntityStatusActive, LegalEntityStatusSuspended, LegalEntityStatusDissolved:
		return true
	default:
		return false
	}
}

// ParseLegalEntityStatus menormalkan dan memvalidasi nilai LegalEntityStatus.
func ParseLegalEntityStatus(raw string) (LegalEntityStatus, error) {
	s := LegalEntityStatus(normalizeEnumInput(raw))
	if !s.IsValid() {
		return "", fmt.Errorf("%w: %q", eferrors.ErrInvalidLegalEntityStatus, raw)
	}
	return s, nil
}

// OrgUnitKind merepresentasikan kategori unit organisasi sesuai enum PostgreSQL org_unit_kind.
type OrgUnitKind string

const (
	// OrgUnitKindDivision merepresentasikan divisi organisasi.
	OrgUnitKindDivision OrgUnitKind = "division"
	// OrgUnitKindDepartment merepresentasikan departemen fungsional.
	OrgUnitKindDepartment OrgUnitKind = "department"
	// OrgUnitKindTeam merepresentasikan tim operasional kecil.
	OrgUnitKindTeam OrgUnitKind = "team"
	// OrgUnitKindBranch merepresentasikan unit cabang organisasi.
	OrgUnitKindBranch OrgUnitKind = "branch"
	// OrgUnitKindOther merepresentasikan kategori unit organisasi lain.
	OrgUnitKindOther OrgUnitKind = "other"
)

// String mengembalikan representasi string enum org_unit_kind.
func (k OrgUnitKind) String() string {
	return string(k)
}

// IsValid memeriksa apakah OrgUnitKind sesuai enum PostgreSQL.
func (k OrgUnitKind) IsValid() bool {
	switch k {
	case OrgUnitKindDivision, OrgUnitKindDepartment, OrgUnitKindTeam, OrgUnitKindBranch, OrgUnitKindOther:
		return true
	default:
		return false
	}
}

// ParseOrgUnitKind menormalkan dan memvalidasi nilai OrgUnitKind.
func ParseOrgUnitKind(raw string) (OrgUnitKind, error) {
	k := OrgUnitKind(normalizeEnumInput(raw))
	if !k.IsValid() {
		return "", fmt.Errorf("%w: %q", eferrors.ErrInvalidOrgUnitKind, raw)
	}
	return k, nil
}

// LocationKind merepresentasikan jenis lokasi fisik perusahaan sesuai enum PostgreSQL location_kind.
type LocationKind string

const (
	// LocationKindHeadquarters adalah kantor pusat utama perusahaan.
	LocationKindHeadquarters LocationKind = "headquarters"
	// LocationKindWarehouse adalah gudang penyimpanan dan fulfillment WMS.
	LocationKindWarehouse LocationKind = "warehouse"
	// LocationKindRetailStore adalah toko ritel atau titik penjualan fisik.
	LocationKindRetailStore LocationKind = "retail_store"
	// LocationKindBranchOffice adalah kantor cabang pendukung operasional.
	LocationKindBranchOffice LocationKind = "branch_office"
)

// String mengembalikan representasi string enum location_kind.
func (k LocationKind) String() string {
	return string(k)
}

// IsValid memeriksa apakah LocationKind sesuai enum PostgreSQL.
func (k LocationKind) IsValid() bool {
	switch k {
	case LocationKindHeadquarters, LocationKindWarehouse, LocationKindRetailStore, LocationKindBranchOffice:
		return true
	default:
		return false
	}
}

// ParseLocationKind menormalkan dan memvalidasi nilai LocationKind.
func ParseLocationKind(raw string) (LocationKind, error) {
	k := LocationKind(normalizeEnumInput(raw))
	if !k.IsValid() {
		return "", fmt.Errorf("%w: %q", eferrors.ErrInvalidLocationKind, raw)
	}
	return k, nil
}

// PartyKind merepresentasikan bentuk entitas mitra eksternal sesuai enum PostgreSQL party_kind.
type PartyKind string

const (
	// PartyKindOrganization adalah mitra berbentuk badan usaha atau perusahaan.
	PartyKindOrganization PartyKind = "organization"
	// PartyKindIndividual adalah mitra perorangan atau freelancer B2B.
	PartyKindIndividual PartyKind = "individual"
	// PartyKindFinancialEntity adalah lembaga keuangan atau bank mitra.
	PartyKindFinancialEntity PartyKind = "financial_entity"
	// PartyKindGovernmentBody adalah instansi atau lembaga pemerintah.
	PartyKindGovernmentBody PartyKind = "government_body"
)

// String mengembalikan representasi string enum party_kind.
func (k PartyKind) String() string {
	return string(k)
}

// IsValid memeriksa apakah PartyKind sesuai enum PostgreSQL.
func (k PartyKind) IsValid() bool {
	switch k {
	case PartyKindOrganization, PartyKindIndividual, PartyKindFinancialEntity, PartyKindGovernmentBody:
		return true
	default:
		return false
	}
}

// ParsePartyKind menormalkan dan memvalidasi nilai PartyKind.
func ParsePartyKind(raw string) (PartyKind, error) {
	k := PartyKind(normalizeEnumInput(raw))
	if !k.IsValid() {
		return "", fmt.Errorf("%w: %q", eferrors.ErrInvalidPartyKind, raw)
	}
	return k, nil
}

// PartyStatus merepresentasikan status siklus hidup mitra eksternal sesuai enum PostgreSQL party_status.
type PartyStatus string

const (
	// PartyStatusPendingVerification menandakan data legal/pajak mitra menunggu verifikasi internal.
	PartyStatusPendingVerification PartyStatus = "pending_verification"
	// PartyStatusActive menandakan mitra aktif dan dapat terlibat pada transaksi bisnis.
	PartyStatusActive PartyStatus = "active"
	// PartyStatusSuspended menandakan mitra ditangguhkan sementara dari transaksi baru.
	PartyStatusSuspended PartyStatus = "suspended"
	// PartyStatusInactive menandakan hubungan mitra diarsipkan atau nonaktif.
	PartyStatusInactive PartyStatus = "inactive"
)

// String mengembalikan representasi string enum party_status.
func (s PartyStatus) String() string {
	return string(s)
}

// IsValid memeriksa apakah PartyStatus sesuai enum PostgreSQL.
func (s PartyStatus) IsValid() bool {
	switch s {
	case PartyStatusPendingVerification, PartyStatusActive, PartyStatusSuspended, PartyStatusInactive:
		return true
	default:
		return false
	}
}

// ParsePartyStatus menormalkan dan memvalidasi nilai PartyStatus.
func ParsePartyStatus(raw string) (PartyStatus, error) {
	s := PartyStatus(normalizeEnumInput(raw))
	if !s.IsValid() {
		return "", fmt.Errorf("%w: %q", eferrors.ErrInvalidPartyStatus, raw)
	}
	return s, nil
}

// PartyRoleKind merepresentasikan peran bisnis mitra terhadap perusahaan sesuai enum PostgreSQL party_role_kind.
type PartyRoleKind string

const (
	// PartyRoleKindSupplier adalah pemasok bahan baku, kain, atau barang jadi (terhubung ke scope supplier_party IAM).
	PartyRoleKindSupplier PartyRoleKind = "supplier"
	// PartyRoleKindVendor adalah vendor layanan atau makloon/produksi eksternal umum (terhubung ke scope vendor_party IAM).
	PartyRoleKindVendor PartyRoleKind = "vendor"
	// PartyRoleKindCustomerB2B adalah pelanggan grosir atau korporat B2B.
	PartyRoleKindCustomerB2B PartyRoleKind = "customer_b2b"
	// PartyRoleKindCarrier adalah mitra ekspedisi atau logistik pengiriman.
	PartyRoleKindCarrier PartyRoleKind = "carrier"
	// PartyRoleKindFinancialInstitution adalah institusi pembayaran atau perbankan mitra.
	PartyRoleKindFinancialInstitution PartyRoleKind = "financial_institution"
	// PartyRoleKindOther adalah peran bisnis eksternal khusus lainnya.
	PartyRoleKindOther PartyRoleKind = "other"
)

// String mengembalikan representasi string enum party_role_kind.
func (r PartyRoleKind) String() string {
	return string(r)
}

// IsValid memeriksa apakah PartyRoleKind sesuai enum PostgreSQL.
func (r PartyRoleKind) IsValid() bool {
	switch r {
	case PartyRoleKindSupplier,
		PartyRoleKindVendor,
		PartyRoleKindCustomerB2B,
		PartyRoleKindCarrier,
		PartyRoleKindFinancialInstitution,
		PartyRoleKindOther:
		return true
	default:
		return false
	}
}

// IAMScopeType mengembalikan nama scope IAM yang dipetakan untuk peran mitra portal eksternal.
func (r PartyRoleKind) IAMScopeType() (string, bool) {
	switch r {
	case PartyRoleKindSupplier:
		return "supplier_party", true
	case PartyRoleKindVendor:
		return "vendor_party", true
	default:
		return "", false
	}
}

// ParsePartyRoleKind menormalkan dan memvalidasi nilai PartyRoleKind.
func ParsePartyRoleKind(raw string) (PartyRoleKind, error) {
	r := PartyRoleKind(normalizeEnumInput(raw))
	if !r.IsValid() {
		return "", fmt.Errorf("%w: %q", eferrors.ErrInvalidPartyRoleKind, raw)
	}
	return r, nil
}

// PortalAccessStatus merepresentasikan status tautan akses portal PIC mitra terhadap akun IAM sesuai enum portal_access_status.
type PortalAccessStatus string

const (
	// PortalAccessStatusNone menandakan kontak murni direktori tanpa akun portal IAM.
	PortalAccessStatusNone PortalAccessStatus = "none"
	// PortalAccessStatusPendingLink menandakan kontak menunggu penyelesaian tautan/persetujuan portal IAM.
	PortalAccessStatusPendingLink PortalAccessStatus = "pending_link"
	// PortalAccessStatusActive menandakan kontak berstatus aktif mewakili mitra pada portal IAM.
	PortalAccessStatusActive PortalAccessStatus = "active"
	// PortalAccessStatusRevoked menandakan akses portal PIC telah dicabut oleh mitra atau admin.
	PortalAccessStatusRevoked PortalAccessStatus = "revoked"
)

// String mengembalikan representasi string enum portal_access_status.
func (s PortalAccessStatus) String() string {
	return string(s)
}

// IsValid memeriksa apakah PortalAccessStatus sesuai enum PostgreSQL.
func (s PortalAccessStatus) IsValid() bool {
	switch s {
	case PortalAccessStatusNone, PortalAccessStatusPendingLink, PortalAccessStatusActive, PortalAccessStatusRevoked:
		return true
	default:
		return false
	}
}

// ParsePortalAccessStatus menormalkan dan memvalidasi nilai PortalAccessStatus.
func ParsePortalAccessStatus(raw string) (PortalAccessStatus, error) {
	s := PortalAccessStatus(normalizeEnumInput(raw))
	if !s.IsValid() {
		return "", fmt.Errorf("%w: %q", eferrors.ErrInvalidPortalAccessStatus, raw)
	}
	return s, nil
}

// AddressUsageKind merepresentasikan fungsi penggunaan alamat mitra sesuai enum PostgreSQL address_usage_kind.
type AddressUsageKind string

const (
	// AddressUsageKindRegisteredOffice adalah alamat legal terdaftar pada akta/NIB mitra.
	AddressUsageKindRegisteredOffice AddressUsageKind = "registered_office"
	// AddressUsageKindHeadquarters adalah alamat kantor utama operasional mitra.
	AddressUsageKindHeadquarters AddressUsageKind = "headquarters"
	// AddressUsageKindBilling adalah alamat tujuan penagihan faktur.
	AddressUsageKindBilling AddressUsageKind = "billing"
	// AddressUsageKindShipping adalah alamat tujuan pengiriman atau penjemputan barang.
	AddressUsageKindShipping AddressUsageKind = "shipping"
	// AddressUsageKindOther adalah peruntukan alamat tambahan lainnya.
	AddressUsageKindOther AddressUsageKind = "other"
)

// String mengembalikan representasi string enum address_usage_kind.
func (u AddressUsageKind) String() string {
	return string(u)
}

// IsValid memeriksa apakah AddressUsageKind sesuai enum PostgreSQL.
func (u AddressUsageKind) IsValid() bool {
	switch u {
	case AddressUsageKindRegisteredOffice,
		AddressUsageKindHeadquarters,
		AddressUsageKindBilling,
		AddressUsageKindShipping,
		AddressUsageKindOther:
		return true
	default:
		return false
	}
}

// ParseAddressUsageKind menormalkan dan memvalidasi nilai AddressUsageKind.
func ParseAddressUsageKind(raw string) (AddressUsageKind, error) {
	u := AddressUsageKind(normalizeEnumInput(raw))
	if !u.IsValid() {
		return "", fmt.Errorf("%w: %q", eferrors.ErrInvalidAddressUsage, raw)
	}
	return u, nil
}

// OutboxStatus merepresentasikan status pemrosesan event outbox sesuai enum PostgreSQL outbox_status.
type OutboxStatus string

const (
	// OutboxStatusPending menandakan event siap dikirim atau menunggu jadwal ulang relay.
	OutboxStatusPending OutboxStatus = "pending"
	// OutboxStatusEffective menandakan event telah berhasil diterbitkan/diproyeksikan ke target lintas layanan.
	OutboxStatusEffective OutboxStatus = "effective"
	// OutboxStatusFailed menandakan event gagal diproses permanen atau melampaui batas percobaan.
	OutboxStatusFailed OutboxStatus = "failed"
)

// String mengembalikan representasi string enum outbox_status.
func (s OutboxStatus) String() string {
	return string(s)
}

// IsValid memeriksa apakah OutboxStatus sesuai enum PostgreSQL.
func (s OutboxStatus) IsValid() bool {
	switch s {
	case OutboxStatusPending, OutboxStatusEffective, OutboxStatusFailed:
		return true
	default:
		return false
	}
}

// ParseOutboxStatus menormalkan dan memvalidasi nilai OutboxStatus.
func ParseOutboxStatus(raw string) (OutboxStatus, error) {
	s := OutboxStatus(normalizeEnumInput(raw))
	if !s.IsValid() {
		return "", fmt.Errorf("%w: %q", eferrors.ErrInvalidOutboxStatus, raw)
	}
	return s, nil
}

func normalizeEnumInput(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}
