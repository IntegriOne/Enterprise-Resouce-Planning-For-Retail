package pgstore

import (
	"bytes"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
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

// NormalizeDatePtr menormalkan pointer tanggal kalender opsional menjadi pukul 00:00:00 UTC sesuai model domain.
func NormalizeDatePtr(t *time.Time) *time.Time {
	if t == nil || t.IsZero() {
		return nil
	}
	y, m, d := t.Date()
	val := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return &val
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
	trimmed := entity.NormalizeUUID(raw)
	if err := entity.ValidateUUID(field, trimmed); err != nil {
		return "", eferrors.WithField(err, field, "must be a valid uuid")
	}
	return trimmed, nil
}

// ValidateOptionalUUIDParam memvalidasi pointer UUID opsional; string kosong dinormalkan menjadi nil kecuali invalid.
func ValidateOptionalUUIDParam(field string, raw *string) (*string, error) {
	return entity.ValidateOptionalUUID(field, raw)
}

// NormalizeUUIDSlice memvalidasi, menormalkan, dan menduplikasi-hapus daftar UUID sambil mempertahankan urutan kemunculan pertama.
func NormalizeUUIDSlice(field string, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, raw := range ids {
		uid, err := ValidateUUIDParam(field, raw)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[uid]; exists {
			continue
		}
		seen[uid] = struct{}{}
		out = append(out, uid)
	}
	return out, nil
}

// NullableStringPtr mengonversi string opsional ke pointer nullable untuk penyimpanan SQL NULL saat kosong.
func NullableStringPtr(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// CloneStringPtr menyalin pointer string opsional agar mutasi model/entitas tidak saling berbagi referensi.
func CloneStringPtr(ptr *string) *string {
	if ptr == nil {
		return nil
	}
	v := *ptr
	return &v
}

// CloneFloat64Ptr menyalin pointer float64 opsional secara aman.
func CloneFloat64Ptr(ptr *float64) *float64 {
	if ptr == nil {
		return nil
	}
	v := *ptr
	return &v
}

// CloneBytes menyalin slice byte secara aman agar payload JSONB tidak berbagi backing array.
func CloneBytes(src []byte) []byte {
	return bytes.Clone(src)
}

// DerefString mengambil nilai string dari pointer nullable secara aman.
func DerefString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// MapSlice mengonversi slice model GORM menjadi slice entitas domain menggunakan fungsi pemeta yang disediakan.
func MapSlice[M any, E any](rows []M, mapFn func(*M) (E, error)) ([]E, error) {
	out := make([]E, 0, len(rows))
	for i := range rows {
		mapped, err := mapFn(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

// Paginate menjalankan kueri berhalaman dengan penghitungan total terpisah dan pemetaan entitas domain secara konsisten.
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
	if total == 0 || int64(norm.Offset) >= total {
		return domainrepo.NewPageResult([]E{}, total, norm), nil
	}

	var rows []M
	if err := q.Session(&gorm.Session{}).Order(orderBy).Limit(norm.Limit).Offset(norm.Offset).Find(&rows).Error; err != nil {
		return domainrepo.PageResult[E]{}, MapError(err, nil)
	}

	items, err := MapSlice(rows, mapFn)
	if err != nil {
		return domainrepo.PageResult[E]{}, err
	}
	return domainrepo.NewPageResult(items, total, norm), nil
}
