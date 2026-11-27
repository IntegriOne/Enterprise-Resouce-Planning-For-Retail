package entity

import (
	"math"
	"time"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// Address merepresentasikan master alamat geografis bersama pada tabel addresses untuk badan hukum, lokasi internal, dan mitra eksternal.
type Address struct {
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
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// AddressParams mengelompokkan field masukan untuk membentuk atau memperbarui Address.
type AddressParams struct {
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

// NewAddress membentuk entitas Address tervalidasi sesuai constraint ck_addresses_* pada skema PostgreSQL.
func NewAddress(id string, params AddressParams, now time.Time) (*Address, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	validated, err := validateAddressParams(params)
	if err != nil {
		return nil, err
	}
	ts := normalizeLocalTime(now)
	return &Address{
		ID:               NormalizeUUID(id),
		Label:            validated.Label,
		AddressLine1:     validated.AddressLine1,
		AddressLine2:     validated.AddressLine2,
		City:             validated.City,
		Province:         validated.Province,
		PostalCode:       validated.PostalCode,
		CountryCode:      validated.CountryCode,
		FormattedAddress: validated.FormattedAddress,
		Latitude:         validated.Latitude,
		Longitude:        validated.Longitude,
		CreatedAt:        ts,
		UpdatedAt:        ts,
	}, nil
}

// Update memperbarui seluruh field alamat secara atomik setelah menegakkan invarian validasi.
func (a *Address) Update(params AddressParams, now time.Time) error {
	if a == nil {
		return eferrors.ErrAddressNotFound
	}
	validated, err := validateAddressParams(params)
	if err != nil {
		return err
	}
	a.Label = validated.Label
	a.AddressLine1 = validated.AddressLine1
	a.AddressLine2 = validated.AddressLine2
	a.City = validated.City
	a.Province = validated.Province
	a.PostalCode = validated.PostalCode
	a.CountryCode = validated.CountryCode
	a.FormattedAddress = validated.FormattedAddress
	a.Latitude = validated.Latitude
	a.Longitude = validated.Longitude
	a.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Clone mengembalikan salinan mendalam Address agar pemanggil tidak memodifikasi pointer internal secara langsung.
func (a *Address) Clone() *Address {
	if a == nil {
		return nil
	}
	return &Address{
		ID:               a.ID,
		Label:            cloneStringPtr(a.Label),
		AddressLine1:     a.AddressLine1,
		AddressLine2:     cloneStringPtr(a.AddressLine2),
		City:             a.City,
		Province:         cloneStringPtr(a.Province),
		PostalCode:       cloneStringPtr(a.PostalCode),
		CountryCode:      a.CountryCode,
		FormattedAddress: cloneStringPtr(a.FormattedAddress),
		Latitude:         cloneFloat64Ptr(a.Latitude),
		Longitude:        cloneFloat64Ptr(a.Longitude),
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
	}
}

func validateAddressParams(params AddressParams) (AddressParams, error) {
	line1, err := ValidateRequiredText(params.AddressLine1, MaxAddressLineLength, eferrors.ErrInvalidAddressLine)
	if err != nil {
		return AddressParams{}, eferrors.WithField(err, "address_line_1", "address line 1 must be non-empty and within 240 characters")
	}
	city, err := ValidateRequiredText(params.City, MaxCityLength, eferrors.ErrInvalidCity)
	if err != nil {
		return AddressParams{}, eferrors.WithField(err, "city", "city must be non-empty and within 120 characters")
	}
	label, err := ValidateOptionalText("label", params.Label, MaxAddressLabelLength, eferrors.ErrInvalidInput)
	if err != nil {
		return AddressParams{}, err
	}
	line2, err := ValidateOptionalText("address_line_2", params.AddressLine2, MaxAddressLineLength, eferrors.ErrInvalidAddressLine)
	if err != nil {
		return AddressParams{}, err
	}
	province, err := ValidateOptionalText("province", params.Province, MaxProvinceLength, eferrors.ErrInvalidInput)
	if err != nil {
		return AddressParams{}, err
	}
	postalCode, err := ValidateOptionalText("postal_code", params.PostalCode, MaxPostalCodeLength, eferrors.ErrInvalidInput)
	if err != nil {
		return AddressParams{}, err
	}

	countryCode, err := NormalizeCountryCodeOrDefault("country_code", params.CountryCode)
	if err != nil {
		return AddressParams{}, err
	}

	formatted, err := ValidateOptionalMultilineText("formatted_address", params.FormattedAddress, eferrors.ErrInvalidInput)
	if err != nil {
		return AddressParams{}, err
	}

	lat, err := validateCoordinate("latitude", params.Latitude, MinLatitude, MaxLatitude)
	if err != nil {
		return AddressParams{}, err
	}
	lon, err := validateCoordinate("longitude", params.Longitude, MinLongitude, MaxLongitude)
	if err != nil {
		return AddressParams{}, err
	}

	return AddressParams{
		Label:            label,
		AddressLine1:     line1,
		AddressLine2:     line2,
		City:             city,
		Province:         province,
		PostalCode:       postalCode,
		CountryCode:      countryCode,
		FormattedAddress: formatted,
		Latitude:         lat,
		Longitude:        lon,
	}, nil
}

func validateCoordinate(field string, val *float64, minVal, maxVal float64) (*float64, error) {
	if val == nil {
		return nil, nil
	}
	v := *val
	// Tolak NaN dan Infinity secara eksplisit karena perbandingan IEEE 754 terhadap NaN selalu bernilai false.
	if math.IsNaN(v) || math.IsInf(v, 0) || v < minVal || v > maxVal {
		return nil, eferrors.WithField(eferrors.ErrInvalidCoordinates, field, "coordinate is out of valid range")
	}
	return &v, nil
}
