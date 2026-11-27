package entity

import (
	"time"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// Party merepresentasikan entitas master mitra bisnis eksternal (supplier, vendor, customer B2B, carrier, institusi keuangan) pada tabel parties.
type Party struct {
	ID                 string
	PartnerCode        string
	Kind               PartyKind
	DisplayName        string
	LegalName          *string
	Email              *string
	Phone              *string
	WebsiteURL         *string
	TaxIDNumber        *string
	TaxRegisteredName  *string
	IsPKP              bool
	IsTaxExempt        bool
	TaxExemptionReason *string
	TaxVerifiedAt      *time.Time
	PreferredLanguage  string
	Status             PartyStatus
	Notes              *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

// PartyParams mengelompokkan data pembentukan mitra eksternal baru.
type PartyParams struct {
	PartnerCode        string
	Kind               PartyKind
	DisplayName        string
	LegalName          *string
	Email              *string
	Phone              *string
	WebsiteURL         *string
	TaxIDNumber        *string
	TaxRegisteredName  *string
	IsPKP              bool
	IsTaxExempt        bool
	TaxExemptionReason *string
	TaxVerifiedAt      *time.Time
	PreferredLanguage  string
	Status             PartyStatus
	Notes              *string
}

// PartyProfileParams mengelompokkan data pembaruan profil umum mitra tanpa mengubah profil pajak atau status.
type PartyProfileParams struct {
	Kind              PartyKind
	DisplayName       string
	LegalName         *string
	Email             *string
	Phone             *string
	WebsiteURL        *string
	PreferredLanguage string
	Notes             *string
}

// PartyTaxParams mengelompokkan data pembaruan profil perpajakan mitra beserta alasan pembebasan pajak.
type PartyTaxParams struct {
	TaxIDNumber        *string
	TaxRegisteredName  *string
	IsPKP              bool
	IsTaxExempt        bool
	TaxExemptionReason *string
	TaxVerifiedAt      *time.Time
}

// NewParty membentuk entitas Party tervalidasi sesuai invarian kode, profil pajak, dan bahasa komunikasi.
func NewParty(id string, params PartyParams, now time.Time) (*Party, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	code, err := ValidateCode(params.PartnerCode, MaxCodeLength, eferrors.ErrInvalidPartyCode)
	if err != nil {
		return nil, err
	}
	if !params.Status.IsValid() {
		return nil, eferrors.ErrInvalidPartyStatus
	}

	profile, err := validatePartyProfileParams(PartyProfileParams{
		Kind:              params.Kind,
		DisplayName:       params.DisplayName,
		LegalName:         params.LegalName,
		Email:             params.Email,
		Phone:             params.Phone,
		WebsiteURL:        params.WebsiteURL,
		PreferredLanguage: params.PreferredLanguage,
		Notes:             params.Notes,
	})
	if err != nil {
		return nil, err
	}

	tax, err := validatePartyTaxParams(PartyTaxParams{
		TaxIDNumber:        params.TaxIDNumber,
		TaxRegisteredName:  params.TaxRegisteredName,
		IsPKP:              params.IsPKP,
		IsTaxExempt:        params.IsTaxExempt,
		TaxExemptionReason: params.TaxExemptionReason,
		TaxVerifiedAt:      params.TaxVerifiedAt,
	})
	if err != nil {
		return nil, err
	}

	ts := normalizeLocalTime(now)
	return &Party{
		ID:                 NormalizeUUID(id),
		PartnerCode:        code,
		Kind:               profile.Kind,
		DisplayName:        profile.DisplayName,
		LegalName:          profile.LegalName,
		Email:              profile.Email,
		Phone:              profile.Phone,
		WebsiteURL:         profile.WebsiteURL,
		TaxIDNumber:        tax.TaxIDNumber,
		TaxRegisteredName:  tax.TaxRegisteredName,
		IsPKP:              tax.IsPKP,
		IsTaxExempt:        tax.IsTaxExempt,
		TaxExemptionReason: tax.TaxExemptionReason,
		TaxVerifiedAt:      tax.TaxVerifiedAt,
		PreferredLanguage:  profile.PreferredLanguage,
		Status:             params.Status,
		Notes:              profile.Notes,
		CreatedAt:          ts,
		UpdatedAt:          ts,
	}, nil
}

// IsDeleted melaporkan apakah mitra eksternal telah dihapus secara lunak.
func (p *Party) IsDeleted() bool {
	return p != nil && p.DeletedAt != nil
}

// IsEligibleForTransactions memeriksa secara fail-closed apakah mitra aktif dan sah bertransaksi lintas modul.
func (p *Party) IsEligibleForTransactions() error {
	if p == nil {
		return eferrors.ErrPartyNotFound
	}
	if p.IsDeleted() {
		return eferrors.ErrPartyDeleted
	}
	if p.Status != PartyStatusActive {
		return eferrors.ErrPartyInactive
	}
	return nil
}

// UpdateProfile memperbarui jenis entitas, nama, kontak umum, situs web, bahasa, dan catatan mitra.
func (p *Party) UpdateProfile(params PartyProfileParams, now time.Time) error {
	if p == nil {
		return eferrors.ErrPartyNotFound
	}
	if p.IsDeleted() {
		return eferrors.ErrPartyDeleted
	}

	profile, err := validatePartyProfileParams(params)
	if err != nil {
		return err
	}

	p.Kind = profile.Kind
	p.DisplayName = profile.DisplayName
	p.LegalName = profile.LegalName
	p.Email = profile.Email
	p.Phone = profile.Phone
	p.WebsiteURL = profile.WebsiteURL
	p.PreferredLanguage = profile.PreferredLanguage
	p.Notes = profile.Notes
	p.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// UpdateTaxProfile memperbarui NPWP, status PKP, pembebasan pajak beserta alasannya, dan waktu verifikasi pajak mitra.
func (p *Party) UpdateTaxProfile(params PartyTaxParams, now time.Time) error {
	if p == nil {
		return eferrors.ErrPartyNotFound
	}
	if p.IsDeleted() {
		return eferrors.ErrPartyDeleted
	}

	tax, err := validatePartyTaxParams(params)
	if err != nil {
		return err
	}

	p.TaxIDNumber = tax.TaxIDNumber
	p.TaxRegisteredName = tax.TaxRegisteredName
	p.IsPKP = tax.IsPKP
	p.IsTaxExempt = tax.IsTaxExempt
	p.TaxExemptionReason = tax.TaxExemptionReason
	p.TaxVerifiedAt = tax.TaxVerifiedAt
	p.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// VerifyAndActivate menyetujui verifikasi awal atau mengaktifkan kembali mitra ke status active.
func (p *Party) VerifyAndActivate(now time.Time) error {
	if p == nil {
		return eferrors.ErrPartyNotFound
	}
	if p.IsDeleted() {
		return eferrors.ErrPartyDeleted
	}
	p.Status = PartyStatusActive
	p.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Suspend membekukan mitra sementara agar tidak menerima transaksi baru.
func (p *Party) Suspend(now time.Time) error {
	if p == nil {
		return eferrors.ErrPartyNotFound
	}
	if p.IsDeleted() {
		return eferrors.ErrPartyDeleted
	}
	p.Status = PartyStatusSuspended
	p.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// Deactivate menandai hubungan bisnis dengan mitra sudah nonaktif.
func (p *Party) Deactivate(now time.Time) error {
	if p == nil {
		return eferrors.ErrPartyNotFound
	}
	if p.IsDeleted() {
		return eferrors.ErrPartyDeleted
	}
	p.Status = PartyStatusInactive
	p.UpdatedAt = normalizeLocalTime(now)
	return nil
}

// SoftDelete menandai penghapusan lunak mitra dan menurunkan statusnya ke inactive.
func (p *Party) SoftDelete(now time.Time) error {
	if p == nil {
		return eferrors.ErrPartyNotFound
	}
	if p.IsDeleted() {
		return eferrors.ErrPartyDeleted
	}
	ts := normalizeLocalTime(now)
	p.Status = PartyStatusInactive
	p.DeletedAt = &ts
	p.UpdatedAt = ts
	return nil
}

// Clone mengembalikan salinan mendalam Party.
func (p *Party) Clone() *Party {
	if p == nil {
		return nil
	}
	return &Party{
		ID:                 p.ID,
		PartnerCode:        p.PartnerCode,
		Kind:               p.Kind,
		DisplayName:        p.DisplayName,
		LegalName:          cloneStringPtr(p.LegalName),
		Email:              cloneStringPtr(p.Email),
		Phone:              cloneStringPtr(p.Phone),
		WebsiteURL:         cloneStringPtr(p.WebsiteURL),
		TaxIDNumber:        cloneStringPtr(p.TaxIDNumber),
		TaxRegisteredName:  cloneStringPtr(p.TaxRegisteredName),
		IsPKP:              p.IsPKP,
		IsTaxExempt:        p.IsTaxExempt,
		TaxExemptionReason: cloneStringPtr(p.TaxExemptionReason),
		TaxVerifiedAt:      cloneTimePtr(p.TaxVerifiedAt),
		PreferredLanguage:  p.PreferredLanguage,
		Status:             p.Status,
		Notes:              cloneStringPtr(p.Notes),
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
		DeletedAt:          cloneTimePtr(p.DeletedAt),
	}
}

func validatePartyProfileParams(params PartyProfileParams) (PartyProfileParams, error) {
	if !params.Kind.IsValid() {
		return PartyProfileParams{}, eferrors.ErrInvalidPartyKind
	}
	displayName, err := ValidateRequiredText(params.DisplayName, MaxLegalNameLength, eferrors.ErrInvalidPartyName)
	if err != nil {
		return PartyProfileParams{}, eferrors.WithField(err, "display_name", "display name must be non-empty and within 240 characters")
	}
	legalName, err := ValidateOptionalText("legal_name", params.LegalName, MaxLegalNameLength, eferrors.ErrInvalidPartyName)
	if err != nil {
		return PartyProfileParams{}, err
	}
	email, err := ValidateOptionalEmail(params.Email)
	if err != nil {
		return PartyProfileParams{}, eferrors.WithField(err, "email", "invalid party email format")
	}
	phone, err := ValidateOptionalPhone(params.Phone)
	if err != nil {
		return PartyProfileParams{}, eferrors.WithField(err, "phone", "invalid party phone format")
	}
	websiteURL, err := ValidateOptionalURL("website_url", params.WebsiteURL, MaxWebsiteURLLength)
	if err != nil {
		return PartyProfileParams{}, err
	}
	lang, err := ValidateLanguageCode(params.PreferredLanguage)
	if err != nil {
		return PartyProfileParams{}, err
	}
	notes, err := ValidateOptionalMultilineText("notes", params.Notes, eferrors.ErrInvalidInput)
	if err != nil {
		return PartyProfileParams{}, err
	}

	return PartyProfileParams{
		Kind:              params.Kind,
		DisplayName:       displayName,
		LegalName:         legalName,
		Email:             email,
		Phone:             phone,
		WebsiteURL:        websiteURL,
		PreferredLanguage: lang,
		Notes:             notes,
	}, nil
}

func validatePartyTaxParams(params PartyTaxParams) (PartyTaxParams, error) {
	taxID, err := ValidateOptionalText("tax_id_number", params.TaxIDNumber, MaxTaxIDLength, eferrors.ErrInvalidInput)
	if err != nil {
		return PartyTaxParams{}, err
	}
	taxName, err := ValidateOptionalText("tax_registered_name", params.TaxRegisteredName, MaxLegalNameLength, eferrors.ErrInvalidPartyName)
	if err != nil {
		return PartyTaxParams{}, err
	}
	reason, err := ValidateOptionalMultilineText("tax_exemption_reason", params.TaxExemptionReason, eferrors.ErrTaxExemptionReasonRequired)
	if err != nil {
		return PartyTaxParams{}, err
	}
	if params.IsTaxExempt && reason == nil {
		return PartyTaxParams{}, eferrors.WithField(eferrors.ErrTaxExemptionReasonRequired, "tax_exemption_reason", "tax exemption reason is required when tax exempt is true")
	}

	return PartyTaxParams{
		TaxIDNumber:        taxID,
		TaxRegisteredName:  taxName,
		IsPKP:              params.IsPKP,
		IsTaxExempt:        params.IsTaxExempt,
		TaxExemptionReason: reason,
		TaxVerifiedAt:      cloneTimePtr(params.TaxVerifiedAt),
	}, nil
}
