package party

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

func partyToModel(p *entity.Party) (partyModel, error) {
	if p == nil {
		return partyModel{}, eferrors.WithField(eferrors.ErrInvalidInput, "party", "party entity must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", p.ID)
	if err != nil {
		return partyModel{}, err
	}
	code, err := entity.ValidateCode(p.PartnerCode, entity.MaxCodeLength, eferrors.ErrInvalidPartyCode)
	if err != nil {
		return partyModel{}, err
	}
	if !p.Kind.IsValid() {
		return partyModel{}, eferrors.ErrInvalidPartyKind
	}
	displayName, err := entity.ValidateRequiredText(p.DisplayName, entity.MaxLegalNameLength, eferrors.ErrInvalidPartyName)
	if err != nil {
		return partyModel{}, err
	}
	if !p.Status.IsValid() {
		return partyModel{}, eferrors.ErrInvalidPartyStatus
	}
	lang, err := entity.ValidateLanguageCode(p.PreferredLanguage)
	if err != nil {
		return partyModel{}, err
	}
	if p.IsTaxExempt && (p.TaxExemptionReason == nil || strings.TrimSpace(*p.TaxExemptionReason) == "") {
		return partyModel{}, eferrors.ErrTaxExemptionReasonRequired
	}
	email, err := entity.ValidateOptionalEmail(p.Email)
	if err != nil {
		return partyModel{}, err
	}

	return partyModel{
		ID:                 id,
		PartnerCode:        code,
		Kind:               p.Kind.String(),
		DisplayName:        displayName,
		LegalName:          pgstore.CloneStringPtr(p.LegalName),
		Email:              email,
		Phone:              pgstore.CloneStringPtr(p.Phone),
		WebsiteURL:         pgstore.CloneStringPtr(p.WebsiteURL),
		TaxIDNumber:        pgstore.CloneStringPtr(p.TaxIDNumber),
		TaxRegisteredName:  pgstore.CloneStringPtr(p.TaxRegisteredName),
		IsPKP:              p.IsPKP,
		IsTaxExempt:        p.IsTaxExempt,
		TaxExemptionReason: pgstore.CloneStringPtr(p.TaxExemptionReason),
		TaxVerifiedAt:      pgstore.NormalizeLocalTimePtr(p.TaxVerifiedAt),
		PreferredLanguage:  lang,
		Status:             p.Status.String(),
		Notes:              pgstore.CloneStringPtr(p.Notes),
		CreatedAt:          pgstore.MustLocalNow(p.CreatedAt),
		UpdatedAt:          pgstore.MustLocalNow(p.UpdatedAt),
		DeletedAt:          pgstore.NormalizeLocalTimePtr(p.DeletedAt),
	}, nil
}

func partyToDomain(m *partyModel) (*entity.Party, error) {
	if m == nil {
		return nil, eferrors.ErrPartyNotFound
	}
	id, err := pgstore.ValidateUUIDParam("id", m.ID)
	if err != nil {
		return nil, err
	}
	kind, err := entity.ParsePartyKind(m.Kind)
	if err != nil {
		return nil, err
	}
	status, err := entity.ParsePartyStatus(m.Status)
	if err != nil {
		return nil, err
	}

	return &entity.Party{
		ID:                 id,
		PartnerCode:        entity.NormalizeCode(m.PartnerCode),
		Kind:               kind,
		DisplayName:        m.DisplayName,
		LegalName:          pgstore.CloneStringPtr(m.LegalName),
		Email:              pgstore.CloneStringPtr(m.Email),
		Phone:              pgstore.CloneStringPtr(m.Phone),
		WebsiteURL:         pgstore.CloneStringPtr(m.WebsiteURL),
		TaxIDNumber:        pgstore.CloneStringPtr(m.TaxIDNumber),
		TaxRegisteredName:  pgstore.CloneStringPtr(m.TaxRegisteredName),
		IsPKP:              m.IsPKP,
		IsTaxExempt:        m.IsTaxExempt,
		TaxExemptionReason: pgstore.CloneStringPtr(m.TaxExemptionReason),
		TaxVerifiedAt:      pgstore.NormalizeLocalTimePtr(m.TaxVerifiedAt),
		PreferredLanguage:  strings.ToLower(strings.TrimSpace(m.PreferredLanguage)),
		Status:             status,
		Notes:              pgstore.CloneStringPtr(m.Notes),
		CreatedAt:          pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:          pgstore.NormalizeLocalTime(m.UpdatedAt),
		DeletedAt:          pgstore.NormalizeLocalTimePtr(m.DeletedAt),
	}, nil
}

func partyRoleToModel(r *entity.PartyRole) (partyRoleModel, error) {
	if r == nil {
		return partyRoleModel{}, eferrors.WithField(eferrors.ErrInvalidInput, "party_role", "party role entity must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", r.ID)
	if err != nil {
		return partyRoleModel{}, err
	}
	partyID, err := pgstore.ValidateUUIDParam("party_id", r.PartyID)
	if err != nil {
		return partyRoleModel{}, err
	}
	if !r.Role.IsValid() {
		return partyRoleModel{}, eferrors.ErrInvalidPartyRoleKind
	}

	return partyRoleModel{
		ID:            id,
		PartyID:       partyID,
		Role:          r.Role.String(),
		IsActive:      r.IsActive,
		ReferenceNote: pgstore.CloneStringPtr(r.ReferenceNote),
		CreatedAt:     pgstore.MustLocalNow(r.CreatedAt),
		UpdatedAt:     pgstore.MustLocalNow(r.UpdatedAt),
	}, nil
}

func partyRoleToDomain(m *partyRoleModel) (*entity.PartyRole, error) {
	if m == nil {
		return nil, eferrors.ErrPartyRoleNotFound
	}
	id, err := pgstore.ValidateUUIDParam("id", m.ID)
	if err != nil {
		return nil, err
	}
	partyID, err := pgstore.ValidateUUIDParam("party_id", m.PartyID)
	if err != nil {
		return nil, err
	}
	role, err := entity.ParsePartyRoleKind(m.Role)
	if err != nil {
		return nil, err
	}

	return &entity.PartyRole{
		ID:            id,
		PartyID:       partyID,
		Role:          role,
		IsActive:      m.IsActive,
		ReferenceNote: pgstore.CloneStringPtr(m.ReferenceNote),
		CreatedAt:     pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:     pgstore.NormalizeLocalTime(m.UpdatedAt),
	}, nil
}

func partyContactToModel(c *entity.PartyContact) (partyContactModel, error) {
	if c == nil {
		return partyContactModel{}, eferrors.WithField(eferrors.ErrInvalidInput, "party_contact", "party contact entity must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", c.ID)
	if err != nil {
		return partyContactModel{}, err
	}
	partyID, err := pgstore.ValidateUUIDParam("party_id", c.PartyID)
	if err != nil {
		return partyContactModel{}, err
	}
	iamUserID, err := pgstore.ValidateOptionalUUIDParam("iam_user_id", c.IAMUserID)
	if err != nil {
		return partyContactModel{}, err
	}
	fullName, err := entity.ValidateRequiredText(c.FullName, entity.MaxContactFullNameLength, eferrors.ErrInvalidPartyName)
	if err != nil {
		return partyContactModel{}, err
	}
	contactRole := entity.NormalizeCode(c.ContactRole)
	if contactRole == "" {
		contactRole = entity.DefaultContactRole
	}
	email, err := entity.ValidateOptionalEmail(c.Email)
	if err != nil {
		return partyContactModel{}, err
	}
	if !c.PortalAccessStatus.IsValid() {
		return partyContactModel{}, eferrors.ErrInvalidPortalAccessStatus
	}
	if email == nil && (c.Phone == nil || strings.TrimSpace(*c.Phone) == "") && iamUserID == nil {
		return partyContactModel{}, eferrors.ErrContactChannelRequired
	}
	if iamUserID == nil && c.PortalAccessStatus != entity.PortalAccessStatusNone {
		return partyContactModel{}, eferrors.ErrPortalLinkInvariantViolation
	}
	if iamUserID != nil && c.PortalAccessStatus == entity.PortalAccessStatusNone {
		return partyContactModel{}, eferrors.ErrPortalLinkInvariantViolation
	}

	return partyContactModel{
		ID:                 id,
		PartyID:            partyID,
		IAMUserID:          iamUserID,
		FullName:           fullName,
		JobTitle:           pgstore.CloneStringPtr(c.JobTitle),
		ContactRole:        contactRole,
		Email:              email,
		Phone:              pgstore.CloneStringPtr(c.Phone),
		PortalAccessStatus: c.PortalAccessStatus.String(),
		IsPrimary:          c.IsPrimary,
		IsActive:           c.IsActive,
		CreatedAt:          pgstore.MustLocalNow(c.CreatedAt),
		UpdatedAt:          pgstore.MustLocalNow(c.UpdatedAt),
		DeletedAt:          pgstore.NormalizeLocalTimePtr(c.DeletedAt),
	}, nil
}

func partyContactToDomain(m *partyContactModel) (*entity.PartyContact, error) {
	if m == nil {
		return nil, eferrors.ErrPartyContactNotFound
	}
	id, err := pgstore.ValidateUUIDParam("id", m.ID)
	if err != nil {
		return nil, err
	}
	partyID, err := pgstore.ValidateUUIDParam("party_id", m.PartyID)
	if err != nil {
		return nil, err
	}
	iamUserID, err := pgstore.ValidateOptionalUUIDParam("iam_user_id", m.IAMUserID)
	if err != nil {
		return nil, err
	}
	portalStatus, err := entity.ParsePortalAccessStatus(m.PortalAccessStatus)
	if err != nil {
		return nil, err
	}

	return &entity.PartyContact{
		ID:                 id,
		PartyID:            partyID,
		IAMUserID:          iamUserID,
		FullName:           m.FullName,
		JobTitle:           pgstore.CloneStringPtr(m.JobTitle),
		ContactRole:        entity.NormalizeCode(m.ContactRole),
		Email:              pgstore.CloneStringPtr(m.Email),
		Phone:              pgstore.CloneStringPtr(m.Phone),
		PortalAccessStatus: portalStatus,
		IsPrimary:          m.IsPrimary,
		IsActive:           m.IsActive,
		CreatedAt:          pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:          pgstore.NormalizeLocalTime(m.UpdatedAt),
		DeletedAt:          pgstore.NormalizeLocalTimePtr(m.DeletedAt),
	}, nil
}

func partyAddressToModel(link *entity.PartyAddress) (partyAddressModel, error) {
	if link == nil {
		return partyAddressModel{}, eferrors.WithField(eferrors.ErrInvalidInput, "party_address", "party address link must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", link.ID)
	if err != nil {
		return partyAddressModel{}, err
	}
	partyID, err := pgstore.ValidateUUIDParam("party_id", link.PartyID)
	if err != nil {
		return partyAddressModel{}, err
	}
	addressID, err := pgstore.ValidateUUIDParam("address_id", link.AddressID)
	if err != nil {
		return partyAddressModel{}, err
	}
	if !link.Usage.IsValid() {
		return partyAddressModel{}, eferrors.ErrInvalidAddressUsage
	}

	return partyAddressModel{
		ID:        id,
		PartyID:   partyID,
		AddressID: addressID,
		Usage:     link.Usage.String(),
		IsPrimary: link.IsPrimary,
		CreatedAt: pgstore.MustLocalNow(link.CreatedAt),
		UpdatedAt: pgstore.MustLocalNow(link.UpdatedAt),
	}, nil
}

func partyAddressToDomain(m *partyAddressModel) (*entity.PartyAddress, error) {
	if m == nil {
		return nil, eferrors.ErrPartyAddressNotFound
	}
	id, err := pgstore.ValidateUUIDParam("id", m.ID)
	if err != nil {
		return nil, err
	}
	partyID, err := pgstore.ValidateUUIDParam("party_id", m.PartyID)
	if err != nil {
		return nil, err
	}
	addressID, err := pgstore.ValidateUUIDParam("address_id", m.AddressID)
	if err != nil {
		return nil, err
	}
	usage, err := entity.ParseAddressUsageKind(m.Usage)
	if err != nil {
		return nil, err
	}

	return &entity.PartyAddress{
		ID:        id,
		PartyID:   partyID,
		AddressID: addressID,
		Usage:     usage,
		IsPrimary: m.IsPrimary,
		CreatedAt: pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt: pgstore.NormalizeLocalTime(m.UpdatedAt),
	}, nil
}
