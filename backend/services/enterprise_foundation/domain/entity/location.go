package entity

import (
	"bytes"
	"encoding/json"
	"time"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// Location merepresentasikan lokasi fisik operasional internal (kantor utama, gudang, toko ritel, cabang) sekaligus target scope facility IAM.
type Location struct {
	ID                    string
	LegalEntityID         string
	ParentLocationID      *string
	AddressID             *string
	LocationCode          string
	Name                  string
	Kind                  LocationKind
	Phone                 *string
	Timezone              string
	OperationalAttributes json.RawMessage
	IsActive              bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
}

// LocationParams mengelompokkan data pembentukan lokasi fisik baru.
type LocationParams struct {
	LegalEntityID         string
	ParentLocationID      *string
	AddressID             *string
	LocationCode          string
	Name                  string
	Kind                  LocationKind
	Phone                 *string
	Timezone              string
	OperationalAttributes json.RawMessage
}

// LocationUpdateParams mengelompokkan data pembaruan lokasi fisik internal.
type LocationUpdateParams struct {
	ParentLocationID      *string
	AddressID             *string
	Name                  string
	Kind                  LocationKind
	Phone                 *string
	Timezone              string
	OperationalAttributes json.RawMessage
}

// NewLocation membentuk entitas Location baru dalam status aktif setelah memvalidasi kode, jenis lokasi, zona waktu, dan atribut JSONB.
func NewLocation(id string, params LocationParams, now time.Time) (*Location, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	normID := NormalizeUUID(id)
	if err := ValidateUUID("legal_entity_id", params.LegalEntityID); err != nil {
		return nil, err
	}
	code, err := ValidateCode(params.LocationCode, MaxCodeLength, eferrors.ErrInvalidLocationCode)
	if err != nil {
		return nil, err
	}

	upd, err := validateLocationFields(normID, LocationUpdateParams{
		ParentLocationID:      params.ParentLocationID,
		AddressID:             params.AddressID,
		Name:                  params.Name,
		Kind:                  params.Kind,
		Phone:                 params.Phone,
		Timezone:              params.Timezone,
		OperationalAttributes: params.OperationalAttributes,
	})
	if err != nil {
		return nil, err
	}

	ts := normalizeLocalTime(now)
	return &Location{
		ID:                    normID,
		LegalEntityID:         NormalizeUUID(params.LegalEntityID),
		ParentLocationID:      upd.ParentLocationID,
		AddressID:             upd.AddressID,
		LocationCode:          code,
		Name:                  upd.Name,
		Kind:                  upd.Kind,
		Phone:                 upd.Phone,
		Timezone:              upd.Timezone,
		OperationalAttributes: upd.OperationalAttributes,
		IsActive:              true,
		CreatedAt:             ts,
		UpdatedAt:             ts,
	}, nil
}

// IsDeleted melaporkan apakah lokasi telah dihapus secara lunak.
func (l *Location) IsDeleted() bool {
	return l != nil && l.DeletedAt != nil
}

// IsEligibleScope melaporkan apakah lokasi aktif untuk dipakai pada scope facility IAM dan transaksi modul lain.
func (l *Location) IsEligibleScope() bool {
	return l != nil && !l.IsDeleted() && l.IsActive
}

// IsWarehouse melaporkan apakah lokasi berjenis gudang untuk kebutuhan fulfillment dan inventori WMS.
func (l *Location) IsWarehouse() bool {
	return l.IsEligibleScope() && l.Kind == LocationKindWarehouse
}

// CanOperate memeriksa kelayakan operasional lokasi secara fail-closed.
func (l *Location) CanOperate() error {
	if l == nil {
		return eferrors.ErrLocationNotFound
	}
	if l.IsDeleted() {
		return eferrors.ErrLocationDeleted
	}
	if !l.IsActive {
		return eferrors.ErrLocationInactive
	}
	return nil
}

// UpdateDetails memperbarui induk, alamat, nama, jenis, telepon, zona waktu, dan atribut operasional lokasi.
func (l *Location) UpdateDetails(params LocationUpdateParams, now time.Time) error {
	if l == nil {
		return eferrors.ErrLocationNotFound
	}
	if l.IsDeleted() {
		return eferrors.ErrLocationDeleted
	}

	validated, err := validateLocationFields(l.ID, params)
	if err != nil {
		return err
	}
	l.ParentLocationID = validated.ParentLocationID
	l.AddressID = validated.AddressID
	l.Name = validated.Name
	l.Kind = validated.Kind
	l.Phone = validated.Phone
	l.Timezone = validated.Timezone
	l.OperationalAttributes = validated.OperationalAttributes
	l.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Activate mengaktifkan kembali lokasi operasional yang belum dihapus lunak.
func (l *Location) Activate(now time.Time) error {
	if l == nil {
		return eferrors.ErrLocationNotFound
	}
	if l.IsDeleted() {
		return eferrors.ErrLocationDeleted
	}
	l.IsActive = true
	l.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Deactivate menonaktifkan lokasi operasional sementara.
func (l *Location) Deactivate(now time.Time) error {
	if l == nil {
		return eferrors.ErrLocationNotFound
	}
	if l.IsDeleted() {
		return eferrors.ErrLocationDeleted
	}
	l.IsActive = false
	l.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// SoftDelete menandai penghapusan lunak lokasi sekaligus menonaktifkannya.
func (l *Location) SoftDelete(now time.Time) error {
	if l == nil {
		return eferrors.ErrLocationNotFound
	}
	if l.IsDeleted() {
		return eferrors.ErrLocationDeleted
	}
	ts := normalizeLocalTime(now)
	l.IsActive = false
	l.DeletedAt = &ts
	l.UpdatedAt = ts
	return nil
}

// OperationalAttributesCopy mengembalikan salinan aman JSON operational_attributes agar pemanggil tidak memodifikasi slice internal.
func (l *Location) OperationalAttributesCopy() json.RawMessage {
	if l == nil || len(l.OperationalAttributes) == 0 {
		return json.RawMessage("{}")
	}
	return json.RawMessage(bytes.Clone(l.OperationalAttributes))
}

// Clone mengembalikan salinan mendalam Location.
func (l *Location) Clone() *Location {
	if l == nil {
		return nil
	}
	return &Location{
		ID:                    l.ID,
		LegalEntityID:         l.LegalEntityID,
		ParentLocationID:      cloneStringPtr(l.ParentLocationID),
		AddressID:             cloneStringPtr(l.AddressID),
		LocationCode:          l.LocationCode,
		Name:                  l.Name,
		Kind:                  l.Kind,
		Phone:                 cloneStringPtr(l.Phone),
		Timezone:              l.Timezone,
		OperationalAttributes: l.OperationalAttributesCopy(),
		IsActive:              l.IsActive,
		CreatedAt:             l.CreatedAt,
		UpdatedAt:             l.UpdatedAt,
		DeletedAt:             cloneTimePtr(l.DeletedAt),
	}
}

func validateLocationFields(selfID string, params LocationUpdateParams) (LocationUpdateParams, error) {
	parentID, err := ValidateOptionalUUID("parent_location_id", params.ParentLocationID)
	if err != nil {
		return LocationUpdateParams{}, err
	}
	if parentID != nil && *parentID == selfID {
		return LocationUpdateParams{}, eferrors.ErrLocationSelfParent
	}
	addrID, err := ValidateOptionalUUID("address_id", params.AddressID)
	if err != nil {
		return LocationUpdateParams{}, err
	}
	name, err := ValidateRequiredText(params.Name, MaxLocationNameLength, eferrors.ErrInvalidLocationName)
	if err != nil {
		return LocationUpdateParams{}, eferrors.WithField(err, "name", "location name must be non-empty and within 200 characters")
	}
	if !params.Kind.IsValid() {
		return LocationUpdateParams{}, eferrors.ErrInvalidLocationKind
	}
	phone, err := ValidateOptionalPhone(params.Phone)
	if err != nil {
		return LocationUpdateParams{}, eferrors.WithField(err, "phone", "invalid location phone format")
	}
	tz, err := ValidateTimezone(params.Timezone)
	if err != nil {
		return LocationUpdateParams{}, eferrors.WithField(err, "timezone", "invalid iana timezone")
	}
	attrs, err := ValidateJSONObject("operational_attributes", params.OperationalAttributes)
	if err != nil {
		return LocationUpdateParams{}, err
	}

	return LocationUpdateParams{
		ParentLocationID:      parentID,
		AddressID:             addrID,
		Name:                  name,
		Kind:                  params.Kind,
		Phone:                 phone,
		Timezone:              tz,
		OperationalAttributes: json.RawMessage(attrs),
	}, nil
}
