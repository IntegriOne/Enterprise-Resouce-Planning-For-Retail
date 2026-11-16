package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

var decimalPattern = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)$`)

// Decimal menyimpan numeric PostgreSQL sebagai teks agar nilai uang tidak kehilangan presisi.
// Nilai kosong tidak sama dengan nol; kolom wajib harus diisi oleh pemanggil.
type Decimal string

// Value menghasilkan parameter numeric tanpa konversi float.
func (d Decimal) Value() (driver.Value, error) {
	if !decimalPattern.MatchString(string(d)) {
		return nil, fmt.Errorf("%w: invalid decimal value", hrerrors.ErrInvalidInput)
	}
	return string(d), nil
}

// Scan menyalin representasi numeric dari driver PostgreSQL secara persis.
func (d *Decimal) Scan(src any) error {
	if d == nil {
		return fmt.Errorf("decimal destination is nil")
	}
	var raw string
	switch v := src.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	case int64:
		raw = fmt.Sprint(v)
	default:
		return fmt.Errorf("unsupported decimal source %T", src)
	}
	if !decimalPattern.MatchString(raw) {
		return fmt.Errorf("%w: invalid decimal value", hrerrors.ErrInvalidInput)
	}
	*d = Decimal(raw)
	return nil
}

// JSON membedakan SQL NULL (nil) dari literal JSON null dan menjaga payload mentah.
type JSON json.RawMessage

// Value mengirim JSON sebagai teks, bukan parameter bytea.
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	if !json.Valid(j) {
		return nil, fmt.Errorf("%w: invalid json value", hrerrors.ErrInvalidInput)
	}
	return string(j), nil
}

// Scan menyalin payload agar tidak mengacu buffer milik driver.
func (j *JSON) Scan(src any) error {
	if j == nil {
		return fmt.Errorf("json destination is nil")
	}
	switch v := src.(type) {
	case nil:
		*j = nil
		return nil
	case string:
		*j = append((*j)[:0], v...)
	case []byte:
		*j = append((*j)[:0], v...)
	default:
		return fmt.Errorf("unsupported json source %T", src)
	}
	if !json.Valid(*j) {
		return fmt.Errorf("%w: invalid json value", hrerrors.ErrInvalidInput)
	}
	return nil
}

// MarshalJSON mempertahankan objek/array JSON saat entitas diserialisasi.
func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	if !json.Valid(j) {
		return nil, fmt.Errorf("%w: invalid json value", hrerrors.ErrInvalidInput)
	}
	return []byte(j), nil
}

// UnmarshalJSON memvalidasi dan menyalin dokumen JSON.
func (j *JSON) UnmarshalJSON(data []byte) error { return j.Scan(data) }
