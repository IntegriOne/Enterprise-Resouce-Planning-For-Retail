// Package repository mendefinisikan kontrak port persistensi IAM dan unit transaksi tanpa ketergantungan driver database.
package repository

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
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
	return PageResult[T]{
		Items:   items,
		Total:   total,
		Limit:   norm.Limit,
		Offset:  norm.Offset,
		HasMore: int64(norm.Offset)+int64(len(items)) < total,
	}
}

// UserFilter menyaring pencarian daftar pengguna.
type UserFilter struct {
	Query          string
	Kind           *entity.UserKind
	Status         *entity.UserStatus
	IncludeDeleted bool
	Page           PageRequest
}

// Normalize merapikan kata kunci pencarian dan batas halaman filter pengguna.
func (f UserFilter) Normalize() UserFilter {
	f.Query = strings.TrimSpace(f.Query)
	f.Page = f.Page.Normalize()
	return f
}

// SessionFilter menyaring pencarian daftar sesi pengguna.
type SessionFilter struct {
	UserID     string
	ActiveOnly bool
	Page       PageRequest
}

// Normalize merapikan ID pengguna dan parameter halaman filter sesi.
func (f SessionFilter) Normalize() SessionFilter {
	f.UserID = strings.TrimSpace(f.UserID)
	f.Page = f.Page.Normalize()
	return f
}

// RoleFilter menyaring daftar peran berdasarkan kata kunci, domain management, dan status aktif.
type RoleFilter struct {
	Query            string
	ManagementDomain string
	ActiveOnly       bool
	Page             PageRequest
}

// Normalize menormalkan filter peran dan paginasi.
func (f RoleFilter) Normalize() RoleFilter {
	f.Query = strings.TrimSpace(f.Query)
	f.ManagementDomain = entity.NormalizeCode(f.ManagementDomain)
	f.Page = f.Page.Normalize()
	return f
}

// PermissionFilter menyaring katalog izin atomik berdasarkan kata kunci, resource, atau sifat sensitivitas.
type PermissionFilter struct {
	Query       string
	Resource    string
	IsSensitive *bool
	Page        PageRequest
}

// Normalize menormalkan filter izin dan paginasi.
func (f PermissionFilter) Normalize() PermissionFilter {
	f.Query = strings.TrimSpace(f.Query)
	f.Resource = entity.NormalizeCode(f.Resource)
	f.Page = f.Page.Normalize()
	return f
}

// GroupFilter menyaring daftar grup berdasarkan kata kunci atau status aktif.
type GroupFilter struct {
	Query      string
	ActiveOnly bool
	Page       PageRequest
}

// Normalize menormalkan filter grup dan paginasi.
func (f GroupFilter) Normalize() GroupFilter {
	f.Query = strings.TrimSpace(f.Query)
	f.Page = f.Page.Normalize()
	return f
}
