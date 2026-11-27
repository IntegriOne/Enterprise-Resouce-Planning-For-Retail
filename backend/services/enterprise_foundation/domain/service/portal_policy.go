package service

import (
	"slices"
	"strings"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// PortalAccessEvaluation merangkum hasil evaluasi hak perwakilan portal PIC mitra terhadap principal IAM.
type PortalAccessEvaluation struct {
	Allowed     bool
	PartyID     string
	ContactID   string
	IAMUserID   string
	ActiveRoles []entity.PartyRoleKind
}

// PortalAccessPolicy menegakkan aturan keabsahan tautan PIC party_contacts dengan principal IAM secara fail-closed.
type PortalAccessPolicy struct{}

// NewPortalAccessPolicy membentuk evaluator kebijakan akses portal PIC mitra.
func NewPortalAccessPolicy() PortalAccessPolicy {
	return PortalAccessPolicy{}
}

// EvaluateRepresentativeAccess memeriksa apakah iamUserID berhak bertindak sebagai perwakilan aktif mitra pada peran bisnis yang diminta.
func (PortalAccessPolicy) EvaluateRepresentativeAccess(
	party *entity.Party,
	roles []*entity.PartyRole,
	contact *entity.PartyContact,
	iamUserID string,
	requiredRole entity.PartyRoleKind,
) (PortalAccessEvaluation, error) {
	if err := entity.ValidateUUID("iam_user_id", iamUserID); err != nil {
		return PortalAccessEvaluation{}, err
	}
	normUserID := entity.NormalizeUUID(iamUserID)

	if err := party.IsEligibleForTransactions(); err != nil {
		return PortalAccessEvaluation{}, err
	}
	if err := contact.CanUsePortalAccess(); err != nil {
		return PortalAccessEvaluation{}, err
	}
	if !strings.EqualFold(entity.NormalizeUUID(contact.PartyID), entity.NormalizeUUID(party.ID)) {
		return PortalAccessEvaluation{}, eferrors.ErrPartyContactNotFound
	}
	if contact.IAMUserID == nil || !strings.EqualFold(entity.NormalizeUUID(*contact.IAMUserID), normUserID) {
		return PortalAccessEvaluation{}, eferrors.ErrPortalAccessNotActive
	}

	activeRoles := collectActivePartyRoles(party.ID, roles)
	if requiredRole != "" {
		if !requiredRole.IsValid() {
			return PortalAccessEvaluation{}, eferrors.ErrInvalidPartyRoleKind
		}
		if err := ensureRequiredRoleActive(party.ID, roles, requiredRole); err != nil {
			return PortalAccessEvaluation{}, err
		}
	}

	return PortalAccessEvaluation{
		Allowed:     true,
		PartyID:     party.ID,
		ContactID:   contact.ID,
		IAMUserID:   normUserID,
		ActiveRoles: slices.Clone(activeRoles),
	}, nil
}

// CanTransitionPortalStatus memvalidasi apakah transisi status portal PIC dan penggantian principal IAM diperbolehkan.
func (PortalAccessPolicy) CanTransitionPortalStatus(
	current entity.PortalAccessStatus,
	next entity.PortalAccessStatus,
	sameUser bool,
) error {
	if !current.IsValid() || !next.IsValid() {
		return eferrors.ErrInvalidPortalAccessStatus
	}
	switch current {
	case entity.PortalAccessStatusNone:
		if next == entity.PortalAccessStatusPendingLink || next == entity.PortalAccessStatusActive {
			return nil
		}
	case entity.PortalAccessStatusPendingLink:
		if !sameUser && next != entity.PortalAccessStatusNone {
			return eferrors.WithField(eferrors.ErrConflict, "iam_user_id", "contact already has a pending iam user link")
		}
		if next == entity.PortalAccessStatusActive || next == entity.PortalAccessStatusRevoked || next == entity.PortalAccessStatusNone {
			return nil
		}
	case entity.PortalAccessStatusActive:
		if !sameUser {
			return eferrors.WithField(eferrors.ErrConflict, "iam_user_id", "active iam user link must be revoked or unlinked first")
		}
		if next == entity.PortalAccessStatusRevoked || next == entity.PortalAccessStatusNone {
			return nil
		}
	case entity.PortalAccessStatusRevoked:
		if next == entity.PortalAccessStatusPendingLink || next == entity.PortalAccessStatusActive || next == entity.PortalAccessStatusNone {
			return nil
		}
	}

	return eferrors.WithField(eferrors.ErrConflict, "portal_access_status", "invalid portal access status transition")
}

func collectActivePartyRoles(partyID string, roles []*entity.PartyRole) []entity.PartyRoleKind {
	normPartyID := entity.NormalizeUUID(partyID)
	active := make([]entity.PartyRoleKind, 0, len(roles))
	for _, r := range roles {
		if r == nil || !r.IsActive || !strings.EqualFold(entity.NormalizeUUID(r.PartyID), normPartyID) {
			continue
		}
		if !slices.Contains(active, r.Role) {
			active = append(active, r.Role)
		}
	}
	return active
}

func ensureRequiredRoleActive(partyID string, roles []*entity.PartyRole, required entity.PartyRoleKind) error {
	normPartyID := entity.NormalizeUUID(partyID)
	var seen bool
	for _, r := range roles {
		if r == nil || !strings.EqualFold(entity.NormalizeUUID(r.PartyID), normPartyID) || r.Role != required {
			continue
		}
		seen = true
		if r.IsActive {
			return nil
		}
	}
	if seen {
		return eferrors.ErrPartyRoleInactive
	}
	return eferrors.ErrPartyRoleNotFound
}
