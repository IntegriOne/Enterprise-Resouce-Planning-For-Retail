package legalentity

import (
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

func legalEntityToModel(e *entity.LegalEntity) (legalEntityModel, error) {
	if e == nil {
		return legalEntityModel{}, eferrors.WithField(eferrors.ErrInvalidInput, "legal_entity", "legal entity must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", e.ID)
	if err != nil {
		return legalEntityModel{}, err
	}
	parentID, err := pgstore.ValidateOptionalUUIDParam("parent_legal_entity_id", e.ParentLegalEntityID)
	if err != nil {
		return legalEntityModel{}, err
	}
	if parentID != nil && *parentID == id {
		return legalEntityModel{}, eferrors.ErrLegalEntitySelfParent
	}
	code, err := entity.ValidateCode(e.EntityCode, entity.MaxCodeLength, eferrors.ErrInvalidLegalEntityCode)
	if err != nil {
		return legalEntityModel{}, err
	}
	legalName, err := entity.ValidateRequiredText(e.LegalName, entity.MaxLegalNameLength, eferrors.ErrInvalidLegalEntityName)
	if err != nil {
		return legalEntityModel{}, err
	}
	if !e.Status.IsValid() {
		return legalEntityModel{}, eferrors.ErrInvalidLegalEntityStatus
	}
	countryCode, err := entity.NormalizeCountryCodeOrDefault("jurisdiction_country_code", e.JurisdictionCountryCode)
	if err != nil {
		return legalEntityModel{}, err
	}
	regAddrID, err := pgstore.ValidateOptionalUUIDParam("registered_address_id", e.RegisteredAddressID)
	if err != nil {
		return legalEntityModel{}, err
	}
	fiscalMonth := e.FiscalYearStartMonth
	if fiscalMonth == 0 {
		fiscalMonth = entity.DefaultFiscalYearStartMonth
	}
	if err := entity.ValidateFiscalMonth(fiscalMonth); err != nil {
		return legalEntityModel{}, err
	}
	incDate := pgstore.NormalizeDatePtr(e.IncorporationDate)
	dissDate := pgstore.NormalizeDatePtr(e.DissolutionDate)
	if incDate != nil && dissDate != nil && dissDate.Before(*incDate) {
		return legalEntityModel{}, eferrors.ErrInvalidEntityDates
	}

	return legalEntityModel{
		ID:                       id,
		ParentLegalEntityID:      parentID,
		EntityCode:               code,
		LegalName:                legalName,
		TradingName:              pgstore.CloneStringPtr(e.TradingName),
		LegalForm:                pgstore.CloneStringPtr(e.LegalForm),
		IsPrimaryOperatingEntity: e.IsPrimaryOperatingEntity,
		Status:                   e.Status.String(),
		JurisdictionCountryCode:  countryCode,
		NIB:                      pgstore.CloneStringPtr(e.NIB),
		TaxIDNumber:              pgstore.CloneStringPtr(e.TaxIDNumber),
		IsPKP:                    e.IsPKP,
		RegisteredAddressID:      regAddrID,
		FiscalYearStartMonth:     fiscalMonth,
		IncorporationDate:        incDate,
		DissolutionDate:          dissDate,
		CreatedAt:                pgstore.MustLocalNow(e.CreatedAt),
		UpdatedAt:                pgstore.MustLocalNow(e.UpdatedAt),
		DeletedAt:                pgstore.NormalizeLocalTimePtr(e.DeletedAt),
	}, nil
}

func legalEntityToDomain(m *legalEntityModel) (*entity.LegalEntity, error) {
	if m == nil {
		return nil, eferrors.ErrLegalEntityNotFound
	}
	id, err := pgstore.ValidateUUIDParam("id", m.ID)
	if err != nil {
		return nil, err
	}
	parentID, err := pgstore.ValidateOptionalUUIDParam("parent_legal_entity_id", m.ParentLegalEntityID)
	if err != nil {
		return nil, err
	}
	status, err := entity.ParseLegalEntityStatus(m.Status)
	if err != nil {
		return nil, err
	}
	countryCode, err := entity.ValidateCountryCode("jurisdiction_country_code", m.JurisdictionCountryCode)
	if err != nil {
		return nil, err
	}
	regAddrID, err := pgstore.ValidateOptionalUUIDParam("registered_address_id", m.RegisteredAddressID)
	if err != nil {
		return nil, err
	}

	return &entity.LegalEntity{
		ID:                       id,
		ParentLegalEntityID:      parentID,
		EntityCode:               entity.NormalizeCode(m.EntityCode),
		LegalName:                m.LegalName,
		TradingName:              pgstore.CloneStringPtr(m.TradingName),
		LegalForm:                pgstore.CloneStringPtr(m.LegalForm),
		IsPrimaryOperatingEntity: m.IsPrimaryOperatingEntity,
		Status:                   status,
		JurisdictionCountryCode:  countryCode,
		NIB:                      pgstore.CloneStringPtr(m.NIB),
		TaxIDNumber:              pgstore.CloneStringPtr(m.TaxIDNumber),
		IsPKP:                    m.IsPKP,
		RegisteredAddressID:      regAddrID,
		FiscalYearStartMonth:     m.FiscalYearStartMonth,
		IncorporationDate:        pgstore.NormalizeDatePtr(m.IncorporationDate),
		DissolutionDate:          pgstore.NormalizeDatePtr(m.DissolutionDate),
		CreatedAt:                pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:                pgstore.NormalizeLocalTime(m.UpdatedAt),
		DeletedAt:                pgstore.NormalizeLocalTimePtr(m.DeletedAt),
	}, nil
}
