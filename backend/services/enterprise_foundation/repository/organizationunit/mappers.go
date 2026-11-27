package organizationunit

import (
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

func organizationUnitToModel(u *entity.OrganizationUnit) (organizationUnitModel, error) {
	if u == nil {
		return organizationUnitModel{}, eferrors.WithField(eferrors.ErrInvalidInput, "organization_unit", "organization unit entity must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", u.ID)
	if err != nil {
		return organizationUnitModel{}, err
	}
	legalEntityID, err := pgstore.ValidateUUIDParam("legal_entity_id", u.LegalEntityID)
	if err != nil {
		return organizationUnitModel{}, err
	}
	parentID, err := pgstore.ValidateOptionalUUIDParam("parent_unit_id", u.ParentUnitID)
	if err != nil {
		return organizationUnitModel{}, err
	}
	if parentID != nil && *parentID == id {
		return organizationUnitModel{}, eferrors.ErrOrgUnitSelfParent
	}
	primaryLocID, err := pgstore.ValidateOptionalUUIDParam("primary_location_id", u.PrimaryLocationID)
	if err != nil {
		return organizationUnitModel{}, err
	}
	headPersonID, err := pgstore.ValidateOptionalUUIDParam("head_person_id", u.HeadPersonID)
	if err != nil {
		return organizationUnitModel{}, err
	}
	code, err := entity.ValidateCode(u.UnitCode, entity.MaxCodeLength, eferrors.ErrInvalidOrgUnitCode)
	if err != nil {
		return organizationUnitModel{}, err
	}
	name, err := entity.ValidateRequiredText(u.Name, entity.MaxOrgUnitNameLength, eferrors.ErrInvalidOrgUnitName)
	if err != nil {
		return organizationUnitModel{}, err
	}
	if !u.Kind.IsValid() {
		return organizationUnitModel{}, eferrors.ErrInvalidOrgUnitKind
	}

	return organizationUnitModel{
		ID:                id,
		LegalEntityID:     legalEntityID,
		ParentUnitID:      parentID,
		PrimaryLocationID: primaryLocID,
		UnitCode:          code,
		Name:              name,
		Kind:              u.Kind.String(),
		CostCenterCode:    pgstore.CloneStringPtr(u.CostCenterCode),
		HeadPersonID:      headPersonID,
		IsActive:          u.IsActive,
		CreatedAt:         pgstore.MustLocalNow(u.CreatedAt),
		UpdatedAt:         pgstore.MustLocalNow(u.UpdatedAt),
		DeletedAt:         pgstore.NormalizeLocalTimePtr(u.DeletedAt),
	}, nil
}

func organizationUnitToDomain(m *organizationUnitModel) (*entity.OrganizationUnit, error) {
	if m == nil {
		return nil, eferrors.ErrOrgUnitNotFound
	}
	id, err := pgstore.ValidateUUIDParam("id", m.ID)
	if err != nil {
		return nil, err
	}
	legalEntityID, err := pgstore.ValidateUUIDParam("legal_entity_id", m.LegalEntityID)
	if err != nil {
		return nil, err
	}
	parentID, err := pgstore.ValidateOptionalUUIDParam("parent_unit_id", m.ParentUnitID)
	if err != nil {
		return nil, err
	}
	primaryLocID, err := pgstore.ValidateOptionalUUIDParam("primary_location_id", m.PrimaryLocationID)
	if err != nil {
		return nil, err
	}
	headPersonID, err := pgstore.ValidateOptionalUUIDParam("head_person_id", m.HeadPersonID)
	if err != nil {
		return nil, err
	}
	kind, err := entity.ParseOrgUnitKind(m.Kind)
	if err != nil {
		return nil, err
	}

	return &entity.OrganizationUnit{
		ID:                id,
		LegalEntityID:     legalEntityID,
		ParentUnitID:      parentID,
		PrimaryLocationID: primaryLocID,
		UnitCode:          entity.NormalizeCode(m.UnitCode),
		Name:              m.Name,
		Kind:              kind,
		CostCenterCode:    pgstore.CloneStringPtr(m.CostCenterCode),
		HeadPersonID:      headPersonID,
		IsActive:          m.IsActive,
		CreatedAt:         pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:         pgstore.NormalizeLocalTime(m.UpdatedAt),
		DeletedAt:         pgstore.NormalizeLocalTimePtr(m.DeletedAt),
	}, nil
}
