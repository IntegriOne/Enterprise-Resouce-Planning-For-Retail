// Package repository mendefinisikan kontrak port persistensi Enterprise Foundation, unit transaksi, dan filter pencarian tanpa ketergantungan ORM atau driver SQL.
package repository

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
)

const (
	// DefaultPageLimit adalah ukuran halaman bawaan saat batas tidak ditentukan atau bernilai non-positif.
	DefaultPageLimit = 20
	// MaxPageLimit adalah batas maksimum baris per halaman untuk melindungi kueri dari pemuatan berlebihan.
	MaxPageLimit = 100
)

// PageRequest mendefinisikan parameter paginasi berbasis limit dan offset.
type PageRequest struct {
	Limit  int
	Offset int
}

// Normalize menjamin nilai Limit selalu berada di rentang 1..100 dan Offset non-negatif.
func (p PageRequest) Normalize() PageRequest {
	limit := p.Limit
	if limit <= 0 {
		limit = DefaultPageLimit
	} else if limit > MaxPageLimit {
		limit = MaxPageLimit
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}
	return PageRequest{
		Limit:  limit,
		Offset: offset,
	}
}

// PageResult membungkus hasil kueri berhalaman beserta metadata total baris.
type PageResult[T any] struct {
	Items   []T
	Total   int64
	Limit   int
	Offset  int
	HasMore bool
}

// NewPageResult membentuk PageResult yang aman dengan slice non-nil dan perhitungan HasMore konsisten.
func NewPageResult[T any](items []T, total int64, page PageRequest) PageResult[T] {
	norm := page.Normalize()
	if items == nil {
		items = []T{}
	}
	if total < 0 {
		total = 0
	}
	remaining := total - int64(norm.Offset)
	return PageResult[T]{
		Items:   items,
		Total:   total,
		Limit:   norm.Limit,
		Offset:  norm.Offset,
		HasMore: remaining > int64(len(items)),
	}
}

// AddressFilter menyaring pencarian master alamat berdasarkan kata kunci, kota, provinsi, kode pos, dan negara.
type AddressFilter struct {
	Query       string
	CountryCode string
	City        string
	Province    string
	PostalCode  string
	Page        PageRequest
}

// Normalize menormalkan parameter filter pencarian alamat dan paginasi.
func (f AddressFilter) Normalize() AddressFilter {
	f.Query = strings.TrimSpace(f.Query)
	f.CountryCode = strings.ToUpper(strings.TrimSpace(f.CountryCode))
	f.City = strings.TrimSpace(f.City)
	f.Province = strings.TrimSpace(f.Province)
	f.PostalCode = strings.TrimSpace(f.PostalCode)
	f.Page = f.Page.Normalize()
	return f
}

// LegalEntityFilter menyaring pencarian badan hukum berdasarkan kata kunci, induk, status, dan penanda utama.
type LegalEntityFilter struct {
	Query               string
	ParentLegalEntityID *string
	Status              *entity.LegalEntityStatus
	PrimaryOnly         *bool
	IncludeDeleted      bool
	Page                PageRequest
}

// Normalize menormalkan filter pencarian badan hukum dan paginasi.
func (f LegalEntityFilter) Normalize() LegalEntityFilter {
	f.Query = strings.TrimSpace(f.Query)
	f.ParentLegalEntityID = normalizeOptionalUUIDFilter(f.ParentLegalEntityID)
	f.Page = f.Page.Normalize()
	return f
}

// LocationFilter menyaring pencarian lokasi internal berdasarkan badan hukum, induk, alamat, jenis, dan status aktif.
type LocationFilter struct {
	Query            string
	LegalEntityID    string
	ParentLocationID *string
	AddressID        *string
	Kind             *entity.LocationKind
	ActiveOnly       bool
	IncludeDeleted   bool
	Page             PageRequest
}

// Normalize menormalkan filter pencarian lokasi dan paginasi.
func (f LocationFilter) Normalize() LocationFilter {
	f.Query = strings.TrimSpace(f.Query)
	f.LegalEntityID = entity.NormalizeUUID(f.LegalEntityID)
	f.ParentLocationID = normalizeOptionalUUIDFilter(f.ParentLocationID)
	f.AddressID = normalizeOptionalUUIDFilter(f.AddressID)
	f.Page = f.Page.Normalize()
	return f
}

// OrganizationUnitFilter menyaring pencarian unit organisasi berdasarkan badan hukum, induk, lokasi utama, jenis, dan status aktif.
type OrganizationUnitFilter struct {
	Query             string
	LegalEntityID     string
	ParentUnitID      *string
	PrimaryLocationID *string
	Kind              *entity.OrgUnitKind
	ActiveOnly        bool
	IncludeDeleted    bool
	Page              PageRequest
}

// Normalize menormalkan filter pencarian unit organisasi dan paginasi.
func (f OrganizationUnitFilter) Normalize() OrganizationUnitFilter {
	f.Query = strings.TrimSpace(f.Query)
	f.LegalEntityID = entity.NormalizeUUID(f.LegalEntityID)
	f.ParentUnitID = normalizeOptionalUUIDFilter(f.ParentUnitID)
	f.PrimaryLocationID = normalizeOptionalUUIDFilter(f.PrimaryLocationID)
	f.Page = f.Page.Normalize()
	return f
}

// PartyFilter menyaring pencarian mitra eksternal berdasarkan kata kunci, jenis entitas, status, dan peran bisnis aktif.
type PartyFilter struct {
	Query          string
	Kind           *entity.PartyKind
	Status         *entity.PartyStatus
	Role           *entity.PartyRoleKind
	IncludeDeleted bool
	Page           PageRequest
}

// Normalize menormalkan filter pencarian mitra eksternal dan paginasi.
func (f PartyFilter) Normalize() PartyFilter {
	f.Query = strings.TrimSpace(f.Query)
	f.Page = f.Page.Normalize()
	return f
}

// PartyContactFilter menyaring pencarian daftar PIC/kontak mitra berdasarkan mitra, kata kunci, peran kontak, dan status portal IAM.
type PartyContactFilter struct {
	PartyID            string
	Query              string
	ContactRole        string
	PortalAccessStatus *entity.PortalAccessStatus
	ActiveOnly         bool
	IncludeDeleted     bool
	Page               PageRequest
}

// Normalize menormalkan filter pencarian kontak mitra dan paginasi.
func (f PartyContactFilter) Normalize() PartyContactFilter {
	f.PartyID = entity.NormalizeUUID(f.PartyID)
	f.Query = strings.TrimSpace(f.Query)
	f.ContactRole = entity.NormalizeCode(f.ContactRole)
	f.Page = f.Page.Normalize()
	return f
}

// PartyRoleFilter menyaring pencarian daftar peran bisnis mitra berdasarkan mitra, jenis peran, kata kunci catatan, dan status aktif.
type PartyRoleFilter struct {
	PartyID    string
	Role       *entity.PartyRoleKind
	ActiveOnly bool
	Query      string
	Page       PageRequest
}

// Normalize menormalkan filter pencarian peran bisnis mitra dan paginasi.
func (f PartyRoleFilter) Normalize() PartyRoleFilter {
	f.PartyID = entity.NormalizeUUID(f.PartyID)
	f.Query = strings.TrimSpace(f.Query)
	f.Page = f.Page.Normalize()
	return f
}

// PartyAddressFilter menyaring pencarian daftar relasi alamat mitra berdasarkan mitra, alamat master, peruntukan, dan penanda utama.
type PartyAddressFilter struct {
	PartyID     string
	AddressID   string
	Usage       *entity.AddressUsageKind
	PrimaryOnly *bool
	Page        PageRequest
}

// Normalize menormalkan filter pencarian relasi alamat mitra dan paginasi.
func (f PartyAddressFilter) Normalize() PartyAddressFilter {
	f.PartyID = entity.NormalizeUUID(f.PartyID)
	f.AddressID = entity.NormalizeUUID(f.AddressID)
	f.Page = f.Page.Normalize()
	return f
}

func normalizeOptionalUUIDFilter(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := entity.NormalizeUUID(*raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
