package location

import (
	"encoding/json"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

func locationToModel(l *entity.Location) (locationModel, error) {
	if l == nil {
		return locationModel{}, eferrors.WithField(eferrors.ErrInvalidInput, "location", "location entity must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", l.ID)
	if err != nil {
		return locationModel{}, err
	}
	legalEntityID, err := pgstore.ValidateUUIDParam("legal_entity_id", l.LegalEntityID)
	if err != nil {
		return locationModel{}, err
	}
	parentID, err := pgstore.ValidateOptionalUUIDParam("parent_location_id", l.ParentLocationID)
	if err != nil {
		return locationModel{}, err
	}
	if parentID != nil && *parentID == id {
		return locationModel{}, eferrors.ErrLocationSelfParent
	}
	addressID, err := pgstore.ValidateOptionalUUIDParam("address_id", l.AddressID)
	if err != nil {
		return locationModel{}, err
	}
	code, err := entity.ValidateCode(l.LocationCode, entity.MaxCodeLength, eferrors.ErrInvalidLocationCode)
	if err != nil {
		return locationModel{}, err
	}
	name, err := entity.ValidateRequiredText(l.Name, entity.MaxLocationNameLength, eferrors.ErrInvalidLocationName)
	if err != nil {
		return locationModel{}, err
	}
	if !l.Kind.IsValid() {
		return locationModel{}, eferrors.ErrInvalidLocationKind
	}
	tz, err := entity.ValidateTimezone(l.Timezone)
	if err != nil {
		return locationModel{}, err
	}
	attrs, err := entity.ValidateJSONObject("operational_attributes", l.OperationalAttributes)
	if err != nil {
		return locationModel{}, err
	}

	return locationModel{
		ID:                    id,
		LegalEntityID:         legalEntityID,
		ParentLocationID:      parentID,
		AddressID:             addressID,
		LocationCode:          code,
		Name:                  name,
		Kind:                  l.Kind.String(),
		Phone:                 pgstore.CloneStringPtr(l.Phone),
		Timezone:              tz,
		OperationalAttributes: string(attrs),
		IsActive:              l.IsActive,
		CreatedAt:             pgstore.MustLocalNow(l.CreatedAt),
		UpdatedAt:             pgstore.MustLocalNow(l.UpdatedAt),
		DeletedAt:             pgstore.NormalizeLocalTimePtr(l.DeletedAt),
	}, nil
}

func locationToDomain(m *locationModel) (*entity.Location, error) {
	if m == nil {
		return nil, eferrors.ErrLocationNotFound
	}
	id, err := pgstore.ValidateUUIDParam("id", m.ID)
	if err != nil {
		return nil, err
	}
	legalEntityID, err := pgstore.ValidateUUIDParam("legal_entity_id", m.LegalEntityID)
	if err != nil {
		return nil, err
	}
	parentID, err := pgstore.ValidateOptionalUUIDParam("parent_location_id", m.ParentLocationID)
	if err != nil {
		return nil, err
	}
	addressID, err := pgstore.ValidateOptionalUUIDParam("address_id", m.AddressID)
	if err != nil {
		return nil, err
	}
	kind, err := entity.ParseLocationKind(m.Kind)
	if err != nil {
		return nil, err
	}
	attrs, err := entity.ValidateJSONObject("operational_attributes", []byte(m.OperationalAttributes))
	if err != nil {
		return nil, err
	}

	return &entity.Location{
		ID:                    id,
		LegalEntityID:         legalEntityID,
		ParentLocationID:      parentID,
		AddressID:             addressID,
		LocationCode:          entity.NormalizeCode(m.LocationCode),
		Name:                  m.Name,
		Kind:                  kind,
		Phone:                 pgstore.CloneStringPtr(m.Phone),
		Timezone:              m.Timezone,
		OperationalAttributes: json.RawMessage(attrs),
		IsActive:              m.IsActive,
		CreatedAt:             pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:             pgstore.NormalizeLocalTime(m.UpdatedAt),
		DeletedAt:             pgstore.NormalizeLocalTimePtr(m.DeletedAt),
	}, nil
}
