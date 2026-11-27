package address

import (
	"math"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

func validateCoordinate(field string, val *float64, minVal, maxVal float64) (*float64, error) {
	if val == nil {
		return nil, nil
	}
	v := *val
	if math.IsNaN(v) || math.IsInf(v, 0) || v < minVal || v > maxVal {
		return nil, eferrors.WithField(eferrors.ErrInvalidCoordinates, field, "coordinate is out of valid range")
	}
	return &v, nil
}

func addressToModel(a *entity.Address) (addressModel, error) {
	if a == nil {
		return addressModel{}, eferrors.WithField(eferrors.ErrInvalidInput, "address", "address entity must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", a.ID)
	if err != nil {
		return addressModel{}, err
	}
	line1, err := entity.ValidateRequiredText(a.AddressLine1, entity.MaxAddressLineLength, eferrors.ErrInvalidAddressLine)
	if err != nil {
		return addressModel{}, err
	}
	city, err := entity.ValidateRequiredText(a.City, entity.MaxCityLength, eferrors.ErrInvalidCity)
	if err != nil {
		return addressModel{}, err
	}
	country, err := entity.NormalizeCountryCodeOrDefault("country_code", a.CountryCode)
	if err != nil {
		return addressModel{}, err
	}
	lat, err := validateCoordinate("latitude", a.Latitude, entity.MinLatitude, entity.MaxLatitude)
	if err != nil {
		return addressModel{}, err
	}
	lng, err := validateCoordinate("longitude", a.Longitude, entity.MinLongitude, entity.MaxLongitude)
	if err != nil {
		return addressModel{}, err
	}
	createdAt := pgstore.MustLocalNow(a.CreatedAt)
	updatedAt := pgstore.MustLocalNow(a.UpdatedAt)

	return addressModel{
		ID:               id,
		Label:            pgstore.CloneStringPtr(a.Label),
		AddressLine1:     line1,
		AddressLine2:     pgstore.CloneStringPtr(a.AddressLine2),
		City:             city,
		Province:         pgstore.CloneStringPtr(a.Province),
		PostalCode:       pgstore.CloneStringPtr(a.PostalCode),
		CountryCode:      country,
		FormattedAddress: pgstore.CloneStringPtr(a.FormattedAddress),
		Latitude:         lat,
		Longitude:        lng,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}

func addressToDomain(m *addressModel) (*entity.Address, error) {
	if m == nil {
		return nil, eferrors.ErrAddressNotFound
	}
	id, err := pgstore.ValidateUUIDParam("id", m.ID)
	if err != nil {
		return nil, err
	}
	country, err := entity.ValidateCountryCode("country_code", m.CountryCode)
	if err != nil {
		return nil, err
	}
	return &entity.Address{
		ID:               id,
		Label:            pgstore.CloneStringPtr(m.Label),
		AddressLine1:     m.AddressLine1,
		AddressLine2:     pgstore.CloneStringPtr(m.AddressLine2),
		City:             m.City,
		Province:         pgstore.CloneStringPtr(m.Province),
		PostalCode:       pgstore.CloneStringPtr(m.PostalCode),
		CountryCode:      country,
		FormattedAddress: pgstore.CloneStringPtr(m.FormattedAddress),
		Latitude:         pgstore.CloneFloat64Ptr(m.Latitude),
		Longitude:        pgstore.CloneFloat64Ptr(m.Longitude),
		CreatedAt:        pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:        pgstore.NormalizeLocalTime(m.UpdatedAt),
	}, nil
}
