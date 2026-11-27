package entity

import (
	"time"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// LegalEntity merepresentasikan entitas badan hukum perusahaan internal pada tabel legal_entities sekaligus target scope legal_entity IAM.
type LegalEntity struct {
	ID                       string
	ParentLegalEntityID      *string
	EntityCode               string
	LegalName                string
	TradingName              *string
	LegalForm                *string
	IsPrimaryOperatingEntity bool
	Status                   LegalEntityStatus
	JurisdictionCountryCode  string
	NIB                      *string
	TaxIDNumber              *string
	IsPKP                    bool
	RegisteredAddressID      *string
	FiscalYearStartMonth     int16
	IncorporationDate        *time.Time
	DissolutionDate          *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
	DeletedAt                *time.Time
}

// LegalEntityParams membawa data pembentukan badan hukum baru.
type LegalEntityParams struct {
	ParentLegalEntityID      *string
	EntityCode               string
	LegalName                string
	TradingName              *string
	LegalForm                *string
	IsPrimaryOperatingEntity bool
	Status                   LegalEntityStatus
	JurisdictionCountryCode  string
	NIB                      *string
	TaxIDNumber              *string
	IsPKP                    bool
	RegisteredAddressID      *string
	FiscalYearStartMonth     int16
	IncorporationDate        *time.Time
	DissolutionDate          *time.Time
}

// LegalEntityUpdateParams membawa pembaruan profil dan atribut legal/pajak badan hukum.
type LegalEntityUpdateParams struct {
	ParentLegalEntityID     *string
	LegalName               string
	TradingName             *string
	LegalForm               *string
	JurisdictionCountryCode string
	NIB                     *string
	TaxIDNumber             *string
	IsPKP                   bool
	RegisteredAddressID     *string
	FiscalYearStartMonth    int16
	IncorporationDate       *time.Time
	DissolutionDate         *time.Time
}

// NewLegalEntity membentuk entitas LegalEntity tervalidasi sesuai invarian kode, parent, tanggal legal, dan tahun buku.
func NewLegalEntity(id string, params LegalEntityParams, now time.Time) (*LegalEntity, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	normID := NormalizeUUID(id)

	code, err := ValidateCode(params.EntityCode, MaxCodeLength, eferrors.ErrInvalidLegalEntityCode)
	if err != nil {
		return nil, err
	}
	if !params.Status.IsValid() {
		return nil, eferrors.ErrInvalidLegalEntityStatus
	}
	if params.IsPrimaryOperatingEntity && params.Status != LegalEntityStatusActive {
		return nil, eferrors.WithField(eferrors.ErrConflict, "is_primary_operating_entity", "primary operating entity must be active")
	}

	upd, err := validateLegalEntityFields(normID, LegalEntityUpdateParams{
		ParentLegalEntityID:     params.ParentLegalEntityID,
		LegalName:               params.LegalName,
		TradingName:             params.TradingName,
		LegalForm:               params.LegalForm,
		JurisdictionCountryCode: params.JurisdictionCountryCode,
		NIB:                     params.NIB,
		TaxIDNumber:             params.TaxIDNumber,
		IsPKP:                   params.IsPKP,
		RegisteredAddressID:     params.RegisteredAddressID,
		FiscalYearStartMonth:    params.FiscalYearStartMonth,
		IncorporationDate:       params.IncorporationDate,
		DissolutionDate:         params.DissolutionDate,
	})
	if err != nil {
		return nil, err
	}

	ts := normalizeLocalTime(now)
	return &LegalEntity{
		ID:                       normID,
		ParentLegalEntityID:      upd.ParentLegalEntityID,
		EntityCode:               code,
		LegalName:                upd.LegalName,
		TradingName:              upd.TradingName,
		LegalForm:                upd.LegalForm,
		IsPrimaryOperatingEntity: params.IsPrimaryOperatingEntity,
		Status:                   params.Status,
		JurisdictionCountryCode:  upd.JurisdictionCountryCode,
		NIB:                      upd.NIB,
		TaxIDNumber:              upd.TaxIDNumber,
		IsPKP:                    upd.IsPKP,
		RegisteredAddressID:      upd.RegisteredAddressID,
		FiscalYearStartMonth:     upd.FiscalYearStartMonth,
		IncorporationDate:        upd.IncorporationDate,
		DissolutionDate:          upd.DissolutionDate,
		CreatedAt:                ts,
		UpdatedAt:                ts,
	}, nil
}

// IsDeleted melaporkan apakah badan hukum telah dihapus secara lunak.
func (e *LegalEntity) IsDeleted() bool {
	return e != nil && e.DeletedAt != nil
}

// IsEligibleScope melaporkan secara fail-closed apakah badan hukum aktif dan sah dijadikan scope IAM atau referensi transaksi.
func (e *LegalEntity) IsEligibleScope() bool {
	return e != nil && !e.IsDeleted() && e.Status == LegalEntityStatusActive
}

// CanBeUsedInTransactions memeriksa kelayakan operasional badan hukum dan mengembalikan sentinel error spesifik bila ditolak.
func (e *LegalEntity) CanBeUsedInTransactions() error {
	if e == nil {
		return eferrors.ErrLegalEntityNotFound
	}
	if e.IsDeleted() {
		return eferrors.ErrLegalEntityDeleted
	}
	if e.Status != LegalEntityStatusActive {
		return eferrors.ErrLegalEntityInactive
	}
	return nil
}

// UpdateDetails memperbarui profil, induk, alamat terdaftar, NPWP, dan kalender fiskal badan hukum.
func (e *LegalEntity) UpdateDetails(params LegalEntityUpdateParams, now time.Time) error {
	if e == nil {
		return eferrors.ErrLegalEntityNotFound
	}
	if e.IsDeleted() {
		return eferrors.ErrLegalEntityDeleted
	}
	if e.Status == LegalEntityStatusDissolved {
		return eferrors.ErrConflict
	}

	validated, err := validateLegalEntityFields(e.ID, params)
	if err != nil {
		return err
	}

	e.ParentLegalEntityID = validated.ParentLegalEntityID
	e.LegalName = validated.LegalName
	e.TradingName = validated.TradingName
	e.LegalForm = validated.LegalForm
	e.JurisdictionCountryCode = validated.JurisdictionCountryCode
	e.NIB = validated.NIB
	e.TaxIDNumber = validated.TaxIDNumber
	e.IsPKP = validated.IsPKP
	e.RegisteredAddressID = validated.RegisteredAddressID
	e.FiscalYearStartMonth = validated.FiscalYearStartMonth
	e.IncorporationDate = validated.IncorporationDate
	e.DissolutionDate = validated.DissolutionDate
	e.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Activate mengaktifkan badan hukum dari status pending atau suspended.
func (e *LegalEntity) Activate(now time.Time) error {
	if e == nil {
		return eferrors.ErrLegalEntityNotFound
	}
	if e.IsDeleted() {
		return eferrors.ErrLegalEntityDeleted
	}
	if e.Status == LegalEntityStatusDissolved {
		return eferrors.WithField(eferrors.ErrConflict, "status", "dissolved legal entity cannot be reactivated")
	}
	e.Status = LegalEntityStatusActive
	e.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Suspend membekukan sementara badan hukum; entitas utama operasional harus melepaskan status utama terlebih dahulu atau ditangani terpisah.
func (e *LegalEntity) Suspend(now time.Time) error {
	if e == nil {
		return eferrors.ErrLegalEntityNotFound
	}
	if e.IsDeleted() {
		return eferrors.ErrLegalEntityDeleted
	}
	if e.Status == LegalEntityStatusDissolved {
		return eferrors.WithField(eferrors.ErrConflict, "status", "dissolved legal entity cannot be suspended")
	}
	if e.IsPrimaryOperatingEntity {
		return eferrors.ErrPrimaryLegalEntityExists
	}
	e.Status = LegalEntityStatusSuspended
	e.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Dissolve menandai badan hukum telah dibubarkan secara permanen serta memvalidasi tanggal pembubaran.
func (e *LegalEntity) Dissolve(dissolutionDate *time.Time, now time.Time) error {
	if e == nil {
		return eferrors.ErrLegalEntityNotFound
	}
	if e.IsDeleted() {
		return eferrors.ErrLegalEntityDeleted
	}
	if e.IsPrimaryOperatingEntity {
		return eferrors.ErrPrimaryLegalEntityExists
	}
	normDissolution := normalizeCalendarDatePtr(dissolutionDate)
	if normDissolution != nil && e.IncorporationDate != nil && normDissolution.Before(*e.IncorporationDate) {
		return eferrors.ErrInvalidEntityDates
	}
	e.Status = LegalEntityStatusDissolved
	if normDissolution != nil {
		e.DissolutionDate = normDissolution
	}
	e.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// SetPrimaryOperatingEntity menetapkan atau melepaskan status entitas operasional utama perusahaan.
func (e *LegalEntity) SetPrimaryOperatingEntity(primary bool, now time.Time) error {
	if e == nil {
		return eferrors.ErrLegalEntityNotFound
	}
	if e.IsDeleted() {
		return eferrors.ErrLegalEntityDeleted
	}
	if primary && e.Status != LegalEntityStatusActive {
		return eferrors.ErrLegalEntityInactive
	}
	e.IsPrimaryOperatingEntity = primary
	e.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// SoftDelete menandai penghapusan lunak badan hukum; entitas operasional utama tidak boleh langsung dihapus sebelum diganti.
func (e *LegalEntity) SoftDelete(now time.Time) error {
	if e == nil {
		return eferrors.ErrLegalEntityNotFound
	}
	if e.IsDeleted() {
		return eferrors.ErrLegalEntityDeleted
	}
	if e.IsPrimaryOperatingEntity {
		return eferrors.ErrPrimaryLegalEntityExists
	}
	ts := normalizeLocalTime(now)
	e.DeletedAt = &ts
	e.UpdatedAt = ts
	return nil
}

// Clone mengembalikan salinan mendalam LegalEntity.
func (e *LegalEntity) Clone() *LegalEntity {
	if e == nil {
		return nil
	}
	return &LegalEntity{
		ID:                       e.ID,
		ParentLegalEntityID:      cloneStringPtr(e.ParentLegalEntityID),
		EntityCode:               e.EntityCode,
		LegalName:                e.LegalName,
		TradingName:              cloneStringPtr(e.TradingName),
		LegalForm:                cloneStringPtr(e.LegalForm),
		IsPrimaryOperatingEntity: e.IsPrimaryOperatingEntity,
		Status:                   e.Status,
		JurisdictionCountryCode:  e.JurisdictionCountryCode,
		NIB:                      cloneStringPtr(e.NIB),
		TaxIDNumber:              cloneStringPtr(e.TaxIDNumber),
		IsPKP:                    e.IsPKP,
		RegisteredAddressID:      cloneStringPtr(e.RegisteredAddressID),
		FiscalYearStartMonth:     e.FiscalYearStartMonth,
		IncorporationDate:        cloneTimePtr(e.IncorporationDate),
		DissolutionDate:          cloneTimePtr(e.DissolutionDate),
		CreatedAt:                e.CreatedAt,
		UpdatedAt:                e.UpdatedAt,
		DeletedAt:                cloneTimePtr(e.DeletedAt),
	}
}

func validateLegalEntityFields(selfID string, params LegalEntityUpdateParams) (LegalEntityUpdateParams, error) {
	parentID, err := ValidateOptionalUUID("parent_legal_entity_id", params.ParentLegalEntityID)
	if err != nil {
		return LegalEntityUpdateParams{}, err
	}
	if parentID != nil && *parentID == selfID {
		return LegalEntityUpdateParams{}, eferrors.ErrLegalEntitySelfParent
	}

	legalName, err := ValidateRequiredText(params.LegalName, MaxLegalNameLength, eferrors.ErrInvalidLegalEntityName)
	if err != nil {
		return LegalEntityUpdateParams{}, eferrors.WithField(err, "legal_name", "legal name must be non-empty and within 240 characters")
	}
	tradingName, err := ValidateOptionalText("trading_name", params.TradingName, MaxLegalNameLength, eferrors.ErrInvalidLegalEntityName)
	if err != nil {
		return LegalEntityUpdateParams{}, err
	}
	legalForm, err := ValidateOptionalText("legal_form", params.LegalForm, MaxLegalFormLength, eferrors.ErrInvalidInput)
	if err != nil {
		return LegalEntityUpdateParams{}, err
	}
	countryCode, err := NormalizeCountryCodeOrDefault("jurisdiction_country_code", params.JurisdictionCountryCode)
	if err != nil {
		return LegalEntityUpdateParams{}, err
	}
	nib, err := ValidateOptionalText("nib", params.NIB, MaxCodeLength, eferrors.ErrInvalidInput)
	if err != nil {
		return LegalEntityUpdateParams{}, err
	}
	taxID, err := ValidateOptionalText("tax_id_number", params.TaxIDNumber, MaxTaxIDLength, eferrors.ErrInvalidInput)
	if err != nil {
		return LegalEntityUpdateParams{}, err
	}
	addrID, err := ValidateOptionalUUID("registered_address_id", params.RegisteredAddressID)
	if err != nil {
		return LegalEntityUpdateParams{}, err
	}
	if err := ValidateFiscalMonth(params.FiscalYearStartMonth); err != nil {
		return LegalEntityUpdateParams{}, err
	}

	incDate := normalizeCalendarDatePtr(params.IncorporationDate)
	dissDate := normalizeCalendarDatePtr(params.DissolutionDate)
	if incDate != nil && dissDate != nil && dissDate.Before(*incDate) {
		return LegalEntityUpdateParams{}, eferrors.ErrInvalidEntityDates
	}

	return LegalEntityUpdateParams{
		ParentLegalEntityID:     parentID,
		LegalName:               legalName,
		TradingName:             tradingName,
		LegalForm:               legalForm,
		JurisdictionCountryCode: countryCode,
		NIB:                     nib,
		TaxIDNumber:             taxID,
		IsPKP:                   params.IsPKP,
		RegisteredAddressID:     addrID,
		FiscalYearStartMonth:    params.FiscalYearStartMonth,
		IncorporationDate:       incDate,
		DissolutionDate:         dissDate,
	}, nil
}
