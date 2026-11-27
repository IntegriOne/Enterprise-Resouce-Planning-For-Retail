package entity

import (
	"strings"
	"time"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// PartyContact merepresentasikan PIC / perwakilan manusia milik mitra pada tabel party_contacts serta tautan portal opsional ke IAM (iam_user_id).
type PartyContact struct {
	ID                 string
	PartyID            string
	IAMUserID          *string
	FullName           string
	JobTitle           *string
	ContactRole        string
	Email              *string
	Phone              *string
	PortalAccessStatus PortalAccessStatus
	IsPrimary          bool
	IsActive           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

// PartyContactParams mengelompokkan data pembentukan kontak/PIC mitra baru.
type PartyContactParams struct {
	PartyID            string
	IAMUserID          *string
	FullName           string
	JobTitle           *string
	ContactRole        string
	Email              *string
	Phone              *string
	PortalAccessStatus PortalAccessStatus
	IsPrimary          bool
}

// PartyContactUpdateParams mengelompokkan data pembaruan profil kontak/PIC tanpa memutasikan tautan portal IAM secara diam-diam.
type PartyContactUpdateParams struct {
	FullName    string
	JobTitle    *string
	ContactRole string
	Email       *string
	Phone       *string
	IsPrimary   bool
}

// NewPartyContact membentuk entitas PartyContact baru dengan penegakan invarian kanal kontak dan pasangan iam_user_id <-> portal_access_status.
func NewPartyContact(id string, params PartyContactParams, now time.Time) (*PartyContact, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if err := ValidateUUID("party_id", params.PartyID); err != nil {
		return nil, err
	}
	iamUserID, err := ValidateOptionalUUID("iam_user_id", params.IAMUserID)
	if err != nil {
		return nil, err
	}
	if !params.PortalAccessStatus.IsValid() {
		return nil, eferrors.ErrInvalidPortalAccessStatus
	}

	details, err := validatePartyContactDetails(PartyContactUpdateParams{
		FullName:    params.FullName,
		JobTitle:    params.JobTitle,
		ContactRole: params.ContactRole,
		Email:       params.Email,
		Phone:       params.Phone,
		IsPrimary:   params.IsPrimary,
	})
	if err != nil {
		return nil, err
	}

	if err := checkChannelInvariant(details.Email, details.Phone, iamUserID); err != nil {
		return nil, err
	}
	if err := checkPortalLinkPair(iamUserID, params.PortalAccessStatus); err != nil {
		return nil, err
	}

	ts := normalizeLocalTime(now)
	return &PartyContact{
		ID:                 NormalizeUUID(id),
		PartyID:            NormalizeUUID(params.PartyID),
		IAMUserID:          iamUserID,
		FullName:           details.FullName,
		JobTitle:           details.JobTitle,
		ContactRole:        details.ContactRole,
		Email:              details.Email,
		Phone:              details.Phone,
		PortalAccessStatus: params.PortalAccessStatus,
		IsPrimary:          details.IsPrimary,
		IsActive:           true,
		CreatedAt:          ts,
		UpdatedAt:          ts,
	}, nil
}

// IsDeleted melaporkan apakah kontak mitra telah dihapus secara lunak.
func (c *PartyContact) IsDeleted() bool {
	return c != nil && c.DeletedAt != nil
}

// ValidatePortalLinkInvariant memverifikasi konsistensi antara iam_user_id dan portal_access_status sesuai constraint PostgreSQL.
func (c *PartyContact) ValidatePortalLinkInvariant() error {
	if c == nil {
		return eferrors.ErrPartyContactNotFound
	}
	if !c.PortalAccessStatus.IsValid() {
		return eferrors.ErrInvalidPortalAccessStatus
	}
	return checkPortalLinkPair(c.IAMUserID, c.PortalAccessStatus)
}

// HasActivePortalAccess melaporkan secara fail-closed apakah kontak masih aktif dan memiliki tautan portal aktif ke IAM.
func (c *PartyContact) HasActivePortalAccess() bool {
	return c != nil &&
		!c.IsDeleted() &&
		c.IsActive &&
		c.IAMUserID != nil &&
		c.PortalAccessStatus == PortalAccessStatusActive
}

// CanUsePortalAccess memvalidasi kelayakan akses portal kontak secara fail-closed dan mengembalikan error spesifik bila ditolak.
func (c *PartyContact) CanUsePortalAccess() error {
	if c == nil {
		return eferrors.ErrPartyContactNotFound
	}
	if c.IsDeleted() {
		return eferrors.ErrPartyContactDeleted
	}
	if !c.IsActive {
		return eferrors.ErrPartyContactInactive
	}
	if c.IAMUserID == nil || c.PortalAccessStatus != PortalAccessStatusActive {
		return eferrors.ErrPortalAccessNotActive
	}
	return nil
}

// UpdateDetails memperbarui nama, jabatan, peran kontak, serta email/telepon sambil menjaga invarian minimal satu kanal kontak atau tautan IAM.
func (c *PartyContact) UpdateDetails(params PartyContactUpdateParams, now time.Time) error {
	if c == nil {
		return eferrors.ErrPartyContactNotFound
	}
	if c.IsDeleted() {
		return eferrors.ErrPartyContactDeleted
	}
	details, err := validatePartyContactDetails(params)
	if err != nil {
		return err
	}
	if err := checkChannelInvariant(details.Email, details.Phone, c.IAMUserID); err != nil {
		return err
	}

	c.FullName = details.FullName
	c.JobTitle = details.JobTitle
	c.ContactRole = details.ContactRole
	c.Email = details.Email
	c.Phone = details.Phone
	c.IsPrimary = details.IsPrimary
	c.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// RequestPortalLink memulai permintaan penautan PIC ke principal IAM dalam status pending_link.
func (c *PartyContact) RequestPortalLink(iamUserID string, now time.Time) error {
	if c == nil {
		return eferrors.ErrPartyContactNotFound
	}
	if c.IsDeleted() {
		return eferrors.ErrPartyContactDeleted
	}
	if !c.IsActive {
		return eferrors.ErrPartyContactInactive
	}
	if err := ValidateUUID("iam_user_id", iamUserID); err != nil {
		return err
	}
	normUserID := NormalizeUUID(iamUserID)
	if c.IAMUserID != nil && !equalUUIDPtr(c.IAMUserID, normUserID) {
		return eferrors.WithField(eferrors.ErrConflict, "iam_user_id", "contact is already linked to another iam user")
	}
	if c.PortalAccessStatus == PortalAccessStatusActive {
		return eferrors.WithField(eferrors.ErrConflict, "portal_access_status", "portal access is already active")
	}

	c.IAMUserID = &normUserID
	c.PortalAccessStatus = PortalAccessStatusPendingLink
	c.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// ActivatePortalAccess mengaktifkan akses portal PIC terhadap principal IAM terverifikasi.
func (c *PartyContact) ActivatePortalAccess(iamUserID string, now time.Time) error {
	if c == nil {
		return eferrors.ErrPartyContactNotFound
	}
	if c.IsDeleted() {
		return eferrors.ErrPartyContactDeleted
	}
	if !c.IsActive {
		return eferrors.ErrPartyContactInactive
	}
	if err := ValidateUUID("iam_user_id", iamUserID); err != nil {
		return err
	}
	normUserID := NormalizeUUID(iamUserID)
	if c.IAMUserID != nil && !equalUUIDPtr(c.IAMUserID, normUserID) {
		return eferrors.WithField(eferrors.ErrConflict, "iam_user_id", "contact is already linked to another iam user")
	}

	c.IAMUserID = &normUserID
	c.PortalAccessStatus = PortalAccessStatusActive
	c.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// RevokePortalAccess mencabut akses portal PIC dengan tetap mempertahankan jejak audit iam_user_id pada baris kontak.
func (c *PartyContact) RevokePortalAccess(now time.Time) error {
	if c == nil {
		return eferrors.ErrPartyContactNotFound
	}
	if c.IsDeleted() {
		return eferrors.ErrPartyContactDeleted
	}
	if c.IAMUserID == nil {
		return eferrors.ErrPortalLinkInvariantViolation
	}

	c.PortalAccessStatus = PortalAccessStatusRevoked
	c.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// UnlinkPortalUser melepaskan tautan akun IAM secara penuh sekaligus mengembalikan portal_access_status ke none.
func (c *PartyContact) UnlinkPortalUser(now time.Time) error {
	if c == nil {
		return eferrors.ErrPartyContactNotFound
	}
	if c.IsDeleted() {
		return eferrors.ErrPartyContactDeleted
	}
	if err := checkChannelInvariant(c.Email, c.Phone, nil); err != nil {
		return err
	}

	c.IAMUserID = nil
	c.PortalAccessStatus = PortalAccessStatusNone
	c.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Activate mengaktifkan kembali kontak/PIC mitra.
func (c *PartyContact) Activate(now time.Time) error {
	if c == nil {
		return eferrors.ErrPartyContactNotFound
	}
	if c.IsDeleted() {
		return eferrors.ErrPartyContactDeleted
	}
	c.IsActive = true
	c.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Deactivate menonaktifkan kontak/PIC mitra sementara.
func (c *PartyContact) Deactivate(now time.Time) error {
	if c == nil {
		return eferrors.ErrPartyContactNotFound
	}
	if c.IsDeleted() {
		return eferrors.ErrPartyContactDeleted
	}
	c.IsActive = false
	c.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// SoftDelete menandai penghapusan lunak kontak sekaligus mencabut akses portal aktif bila masih tertaut.
func (c *PartyContact) SoftDelete(now time.Time) error {
	if c == nil {
		return eferrors.ErrPartyContactNotFound
	}
	if c.IsDeleted() {
		return eferrors.ErrPartyContactDeleted
	}
	ts := normalizeLocalTime(now)
	c.IsActive = false
	if c.IAMUserID != nil && c.PortalAccessStatus != PortalAccessStatusRevoked {
		c.PortalAccessStatus = PortalAccessStatusRevoked
	}
	c.DeletedAt = &ts
	c.UpdatedAt = ts
	return nil
}

// Clone mengembalikan salinan mendalam PartyContact.
func (c *PartyContact) Clone() *PartyContact {
	if c == nil {
		return nil
	}
	return &PartyContact{
		ID:                 c.ID,
		PartyID:            c.PartyID,
		IAMUserID:          cloneStringPtr(c.IAMUserID),
		FullName:           c.FullName,
		JobTitle:           cloneStringPtr(c.JobTitle),
		ContactRole:        c.ContactRole,
		Email:              cloneStringPtr(c.Email),
		Phone:              cloneStringPtr(c.Phone),
		PortalAccessStatus: c.PortalAccessStatus,
		IsPrimary:          c.IsPrimary,
		IsActive:           c.IsActive,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
		DeletedAt:          cloneTimePtr(c.DeletedAt),
	}
}

func validatePartyContactDetails(params PartyContactUpdateParams) (PartyContactUpdateParams, error) {
	fullName, err := ValidateRequiredText(params.FullName, MaxContactFullNameLength, eferrors.ErrInvalidPartyName)
	if err != nil {
		return PartyContactUpdateParams{}, eferrors.WithField(err, "full_name", "contact full name must be non-empty and within 200 characters")
	}
	jobTitle, err := ValidateOptionalText("job_title", params.JobTitle, MaxJobTitleLength, eferrors.ErrInvalidInput)
	if err != nil {
		return PartyContactUpdateParams{}, err
	}
	rawContactRole := strings.TrimSpace(params.ContactRole)
	if rawContactRole == "" {
		rawContactRole = DefaultContactRole
	}
	contactRole, err := ValidateCode(rawContactRole, MaxCodeLength, eferrors.NewValidation("contact_role", eferrors.Code("enterprise_foundation.party_contact.invalid_contact_role"), "invalid contact role"))
	if err != nil {
		return PartyContactUpdateParams{}, err
	}
	email, err := ValidateOptionalEmail(params.Email)
	if err != nil {
		return PartyContactUpdateParams{}, eferrors.WithField(err, "email", "invalid contact email format")
	}
	phone, err := ValidateOptionalPhone(params.Phone)
	if err != nil {
		return PartyContactUpdateParams{}, eferrors.WithField(err, "phone", "invalid contact phone format")
	}

	return PartyContactUpdateParams{
		FullName:    fullName,
		JobTitle:    jobTitle,
		ContactRole: contactRole,
		Email:       email,
		Phone:       phone,
		IsPrimary:   params.IsPrimary,
	}, nil
}

func checkChannelInvariant(email, phone, iamUserID *string) error {
	if email == nil && phone == nil && iamUserID == nil {
		return eferrors.ErrContactChannelRequired
	}
	return nil
}

func checkPortalLinkPair(iamUserID *string, status PortalAccessStatus) error {
	if iamUserID == nil && status != PortalAccessStatusNone {
		return eferrors.ErrPortalLinkInvariantViolation
	}
	if iamUserID != nil && status == PortalAccessStatusNone {
		return eferrors.ErrPortalLinkInvariantViolation
	}
	return nil
}
