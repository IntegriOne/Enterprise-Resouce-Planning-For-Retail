package pgstore

import (
	"net/netip"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
)

var likeEscaper = strings.NewReplacer(
	`\`, `\\`,
	`%`, `\%`,
	`_`, `\_`,
)

// EscapeLike menormalisasi karakter wildcard LIKE/ILIKE agar kueri pencarian aman dengan klausul ESCAPE '\'.
func EscapeLike(raw string) string {
	return likeEscaper.Replace(raw)
}

// LikePattern membentuk pola pencarian substring aman yang cocok untuk operator PostgreSQL ILIKE ? ESCAPE '\'.
func LikePattern(raw string) string {
	return "%" + EscapeLike(strings.TrimSpace(raw)) + "%"
}

// ParseIPAddress mengubah representasi string/CIDR dari kolom inet PostgreSQL menjadi pointer netip.Addr.
func ParseIPAddress(raw string) *netip.Addr {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	if addr, err := netip.ParseAddr(trimmed); err == nil {
		unmapped := addr.Unmap()
		return &unmapped
	}
	if prefix, err := netip.ParsePrefix(trimmed); err == nil {
		unmapped := prefix.Addr().Unmap()
		return &unmapped
	}
	return nil
}

// ParseIPAddressPtr mengonversi pointer string dari kolom inet PostgreSQL menjadi pointer netip.Addr.
func ParseIPAddressPtr(raw *string) *netip.Addr {
	if raw == nil {
		return nil
	}
	return ParseIPAddress(*raw)
}

// FormatIPAddressPtr menormalkan pointer netip.Addr menjadi string bersih untuk kolom inet PostgreSQL.
func FormatIPAddressPtr(addr *netip.Addr) *string {
	if addr == nil || !addr.IsValid() {
		return nil
	}
	val := addr.Unmap().String()
	return &val
}

// NormalizeLocalTime mengonversi timestamp non-zero ke zona waktu lokal OS (time.Local).
func NormalizeLocalTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return t.In(time.Local)
}

// NormalizeLocalTimePtr menormalkan pointer timestamp opsional ke zona waktu lokal OS.
func NormalizeLocalTimePtr(t *time.Time) *time.Time {
	if t == nil || t.IsZero() {
		return nil
	}
	local := t.In(time.Local)
	return &local
}

// MustLocalNow mengembalikan waktu lokal OS terkini apabila parameter waktu bernilai zero.
func MustLocalNow(now time.Time) time.Time {
	if now.IsZero() {
		return time.Now()
	}
	return now.In(time.Local)
}

// ValidateUUIDParam memvalidasi dan menormalkan parameter UUID ke huruf kecil sebelum diteruskan ke kueri PostgreSQL.
func ValidateUUIDParam(field, raw string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if err := entity.ValidateUUID(field, trimmed); err != nil {
		return "", err
	}
	return trimmed, nil
}

// NullableStringPtr mengonversi string opsional ke pointer nullable untuk penyimpanan SQL NULL saat kosong.
func NullableStringPtr(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// DerefString mengambil nilai string dari pointer nullable secara aman.
func DerefString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// ApplyScopeWhere menerapkan filter PostgreSQL scope_type dan scope_id termasuk kondisi IS NULL untuk global.
func ApplyScopeWhere(q *gorm.DB, scope entity.Scope) (*gorm.DB, error) {
	if !scope.IsValid() {
		return nil, iamerrors.ErrInvalidScope
	}
	if scope.Type == entity.ScopeTypeGlobal || scope.ID == nil {
		return q.Where("scope_type = ? AND scope_id IS NULL", scope.Type.String()), nil
	}
	scopeID, err := ValidateUUIDParam("scope_id", *scope.ID)
	if err != nil {
		return nil, iamerrors.ErrInvalidScope
	}
	return q.Where("scope_type = ? AND scope_id = ?", scope.Type.String(), scopeID), nil
}

// ToScope merekonstruksi entity.Scope dari nilai kolom scope_type dan scope_id hasil pembacaan PostgreSQL.
func ToScope(scopeType string, scopeID *string) (entity.Scope, error) {
	parsedType, err := entity.ParseScopeType(scopeType)
	if err != nil {
		return entity.Scope{}, err
	}
	return entity.NewScope(parsedType, scopeID)
}

// Paginate menjalankan kueri berhalaman secara konsisten dengan penghitungan total terpisah dan pemetaan entitas domain.
func Paginate[M any, E any](
	q *gorm.DB,
	page domainrepo.PageRequest,
	orderBy string,
	mapFn func(*M) (E, error),
) (domainrepo.PageResult[E], error) {
	norm := page.Normalize()
	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return domainrepo.PageResult[E]{}, MapError(err, nil)
	}
	if total == 0 {
		return domainrepo.NewPageResult([]E{}, 0, norm), nil
	}

	var rows []M
	if err := q.Order(orderBy).Limit(norm.Limit).Offset(norm.Offset).Find(&rows).Error; err != nil {
		return domainrepo.PageResult[E]{}, MapError(err, nil)
	}

	items := make([]E, 0, len(rows))
	for i := range rows {
		mapped, err := mapFn(&rows[i])
		if err != nil {
			return domainrepo.PageResult[E]{}, err
		}
		items = append(items, mapped)
	}
	return domainrepo.NewPageResult(items, total, norm), nil
}
