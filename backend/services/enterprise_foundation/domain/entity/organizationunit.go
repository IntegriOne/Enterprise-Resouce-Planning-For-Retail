package entity

import (
	"time"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// OrganizationUnit merepresentasikan unit organisasi internal ber-UUID stabil pada tabel organization_units sekaligus target scope organization_unit IAM.
type OrganizationUnit struct {
	ID                string
	LegalEntityID     string
	ParentUnitID      *string
	PrimaryLocationID *string
	UnitCode          string
	Name              string
	Kind              OrgUnitKind
	CostCenterCode    *string
	HeadPersonID      *string
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// OrganizationUnitParams mengelompokkan data pembentukan unit organisasi baru.
type OrganizationUnitParams struct {
	LegalEntityID     string
	ParentUnitID      *string
	PrimaryLocationID *string
	UnitCode          string
	Name              string
	Kind              OrgUnitKind
	CostCenterCode    *string
	HeadPersonID      *string
}

// OrganizationUnitUpdateParams mengelompokkan data pembaruan unit organisasi internal.
type OrganizationUnitUpdateParams struct {
	ParentUnitID      *string
	PrimaryLocationID *string
	Name              string
	Kind              OrgUnitKind
	CostCenterCode    *string
	HeadPersonID      *string
}

// NewOrganizationUnit membentuk entitas OrganizationUnit baru dalam status aktif setelah memvalidasi kode, nama, induk, dan referensi UUID.
func NewOrganizationUnit(id string, params OrganizationUnitParams, now time.Time) (*OrganizationUnit, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	normID := NormalizeUUID(id)
	if err := ValidateUUID("legal_entity_id", params.LegalEntityID); err != nil {
		return nil, err
	}
	code, err := ValidateCode(params.UnitCode, MaxCodeLength, eferrors.ErrInvalidOrgUnitCode)
	if err != nil {
		return nil, err
	}

	upd, err := validateOrgUnitFields(normID, OrganizationUnitUpdateParams{
		ParentUnitID:      params.ParentUnitID,
		PrimaryLocationID: params.PrimaryLocationID,
		Name:              params.Name,
		Kind:              params.Kind,
		CostCenterCode:    params.CostCenterCode,
		HeadPersonID:      params.HeadPersonID,
	})
	if err != nil {
		return nil, err
	}

	ts := normalizeLocalTime(now)
	return &OrganizationUnit{
		ID:                normID,
		LegalEntityID:     NormalizeUUID(params.LegalEntityID),
		ParentUnitID:      upd.ParentUnitID,
		PrimaryLocationID: upd.PrimaryLocationID,
		UnitCode:          code,
		Name:              upd.Name,
		Kind:              upd.Kind,
		CostCenterCode:    upd.CostCenterCode,
		HeadPersonID:      upd.HeadPersonID,
		IsActive:          true,
		CreatedAt:         ts,
		UpdatedAt:         ts,
	}, nil
}

// IsDeleted melaporkan apakah unit organisasi telah dihapus secara lunak.
func (u *OrganizationUnit) IsDeleted() bool {
	return u != nil && u.DeletedAt != nil
}

// IsEligibleScope melaporkan apakah unit organisasi aktif untuk dipakai sebagai scope IAM atau penempatan HRM.
func (u *OrganizationUnit) IsEligibleScope() bool {
	return u != nil && !u.IsDeleted() && u.IsActive
}

// CanOperate memeriksa kelayakan operasional unit organisasi secara fail-closed.
func (u *OrganizationUnit) CanOperate() error {
	if u == nil {
		return eferrors.ErrOrgUnitNotFound
	}
	if u.IsDeleted() {
		return eferrors.ErrOrgUnitDeleted
	}
	if !u.IsActive {
		return eferrors.ErrOrgUnitInactive
	}
	return nil
}

// UpdateDetails memperbarui unit induk, lokasi utama, nama, jenis, kode pusat biaya, dan pimpinan logis unit.
func (u *OrganizationUnit) UpdateDetails(params OrganizationUnitUpdateParams, now time.Time) error {
	if u == nil {
		return eferrors.ErrOrgUnitNotFound
	}
	if u.IsDeleted() {
		return eferrors.ErrOrgUnitDeleted
	}

	validated, err := validateOrgUnitFields(u.ID, params)
	if err != nil {
		return err
	}

	u.ParentUnitID = validated.ParentUnitID
	u.PrimaryLocationID = validated.PrimaryLocationID
	u.Name = validated.Name
	u.Kind = validated.Kind
	u.CostCenterCode = validated.CostCenterCode
	u.HeadPersonID = validated.HeadPersonID
	u.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// AssignHeadPerson memperbarui atau mengosongkan referensi logis ID pimpinan unit (hr_workers.id di HRM).
func (u *OrganizationUnit) AssignHeadPerson(headPersonID *string, now time.Time) error {
	if u == nil {
		return eferrors.ErrOrgUnitNotFound
	}
	if u.IsDeleted() {
		return eferrors.ErrOrgUnitDeleted
	}
	validatedHead, err := ValidateOptionalUUID("head_person_id", headPersonID)
	if err != nil {
		return err
	}
	u.HeadPersonID = validatedHead
	u.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Activate mengaktifkan kembali unit organisasi.
func (u *OrganizationUnit) Activate(now time.Time) error {
	if u == nil {
		return eferrors.ErrOrgUnitNotFound
	}
	if u.IsDeleted() {
		return eferrors.ErrOrgUnitDeleted
	}
	u.IsActive = true
	u.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Deactivate menonaktifkan unit organisasi sementara.
func (u *OrganizationUnit) Deactivate(now time.Time) error {
	if u == nil {
		return eferrors.ErrOrgUnitNotFound
	}
	if u.IsDeleted() {
		return eferrors.ErrOrgUnitDeleted
	}
	u.IsActive = false
	u.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// SoftDelete menandai penghapusan lunak unit organisasi sekaligus menonaktifkannya.
func (u *OrganizationUnit) SoftDelete(now time.Time) error {
	if u == nil {
		return eferrors.ErrOrgUnitNotFound
	}
	if u.IsDeleted() {
		return eferrors.ErrOrgUnitDeleted
	}
	ts := normalizeLocalTime(now)
	u.IsActive = false
	u.DeletedAt = &ts
	u.UpdatedAt = ts
	return nil
}

// Clone mengembalikan salinan mendalam OrganizationUnit.
func (u *OrganizationUnit) Clone() *OrganizationUnit {
	if u == nil {
		return nil
	}
	return &OrganizationUnit{
		ID:                u.ID,
		LegalEntityID:     u.LegalEntityID,
		ParentUnitID:      cloneStringPtr(u.ParentUnitID),
		PrimaryLocationID: cloneStringPtr(u.PrimaryLocationID),
		UnitCode:          u.UnitCode,
		Name:              u.Name,
		Kind:              u.Kind,
		CostCenterCode:    cloneStringPtr(u.CostCenterCode),
		HeadPersonID:      cloneStringPtr(u.HeadPersonID),
		IsActive:          u.IsActive,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
		DeletedAt:         cloneTimePtr(u.DeletedAt),
	}
}

func validateOrgUnitFields(selfID string, params OrganizationUnitUpdateParams) (OrganizationUnitUpdateParams, error) {
	parentID, err := ValidateOptionalUUID("parent_unit_id", params.ParentUnitID)
	if err != nil {
		return OrganizationUnitUpdateParams{}, err
	}
	if parentID != nil && *parentID == selfID {
		return OrganizationUnitUpdateParams{}, eferrors.ErrOrgUnitSelfParent
	}
	primaryLocID, err := ValidateOptionalUUID("primary_location_id", params.PrimaryLocationID)
	if err != nil {
		return OrganizationUnitUpdateParams{}, err
	}
	headID, err := ValidateOptionalUUID("head_person_id", params.HeadPersonID)
	if err != nil {
		return OrganizationUnitUpdateParams{}, err
	}
	name, err := ValidateRequiredText(params.Name, MaxOrgUnitNameLength, eferrors.ErrInvalidOrgUnitName)
	if err != nil {
		return OrganizationUnitUpdateParams{}, eferrors.WithField(err, "name", "organization unit name must be non-empty and within 180 characters")
	}
	if !params.Kind.IsValid() {
		return OrganizationUnitUpdateParams{}, eferrors.ErrInvalidOrgUnitKind
	}
	costCenter, err := ValidateOptionalText("cost_center_code", params.CostCenterCode, MaxCodeLength, eferrors.ErrInvalidInput)
	if err != nil {
		return OrganizationUnitUpdateParams{}, err
	}

	return OrganizationUnitUpdateParams{
		ParentUnitID:      parentID,
		PrimaryLocationID: primaryLocID,
		Name:              name,
		Kind:              params.Kind,
		CostCenterCode:    costCenter,
		HeadPersonID:      headID,
	}, nil
}
