package service_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	efservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

func TestPortalAccessPolicy_EvaluateRepresentativeAccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	policy := efservice.NewPortalAccessPolicy()

	activeParty, err := entity.NewParty(svcUUID1, entity.PartyParams{
		PartnerCode: "pt-kain",
		Kind:        entity.PartyKindOrganization,
		DisplayName: "PT Kain",
		Status:      entity.PartyStatusActive,
	}, now)
	if err != nil {
		t.Fatalf("NewParty failed: %v", err)
	}

	supRole, _ := entity.NewPartyRole(svcUUID2, svcUUID1, entity.PartyRoleKindSupplier, nil, now)
	venRole, _ := entity.NewPartyRole(svcUUID3, svcUUID1, entity.PartyRoleKindVendor, nil, now)
	_ = venRole.Deactivate(now)

	activeContact, err := entity.NewPartyContact(svcUUID4, entity.PartyContactParams{
		PartyID:            svcUUID1,
		IAMUserID:          strPtr(svcUUID5),
		FullName:           "Budi PIC",
		Email:              strPtr("budi@kain.id"),
		PortalAccessStatus: entity.PortalAccessStatusActive,
	}, now)
	if err != nil {
		t.Fatalf("NewPartyContact failed: %v", err)
	}

	t.Run("valid representative with required supplier role", func(t *testing.T) {
		t.Parallel()
		eval, err := policy.EvaluateRepresentativeAccess(
			activeParty,
			[]*entity.PartyRole{supRole, venRole},
			activeContact,
			strings.ToUpper(svcUUID5),
			entity.PartyRoleKindSupplier,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !eval.Allowed || eval.PartyID != svcUUID1 || eval.ContactID != svcUUID4 || eval.IAMUserID != svcUUID5 {
			t.Errorf("unexpected evaluation: %+v", eval)
		}
		if len(eval.ActiveRoles) != 1 || eval.ActiveRoles[0] != entity.PartyRoleKindSupplier {
			t.Errorf("ActiveRoles = %v, want [supplier]", eval.ActiveRoles)
		}
	})

	t.Run("rejects mismatched iam user id", func(t *testing.T) {
		t.Parallel()
		_, err := policy.EvaluateRepresentativeAccess(activeParty, []*entity.PartyRole{supRole}, activeContact, svcUUID2, entity.PartyRoleKindSupplier)
		if !errors.Is(err, eferrors.ErrPortalAccessNotActive) {
			t.Errorf("mismatched iam user = %v, want ErrPortalAccessNotActive", err)
		}
	})

	t.Run("rejects contact belonging to another party", func(t *testing.T) {
		t.Parallel()
		otherContact, _ := entity.NewPartyContact(svcUUID4, entity.PartyContactParams{
			PartyID:            svcUUID2,
			IAMUserID:          strPtr(svcUUID5),
			FullName:           "Other PIC",
			PortalAccessStatus: entity.PortalAccessStatusActive,
		}, now)
		_, err := policy.EvaluateRepresentativeAccess(activeParty, []*entity.PartyRole{supRole}, otherContact, svcUUID5, entity.PartyRoleKindSupplier)
		if !errors.Is(err, eferrors.ErrPartyContactNotFound) {
			t.Errorf("cross-party contact = %v, want ErrPartyContactNotFound", err)
		}
	})

	t.Run("rejects inactive required role and missing required role", func(t *testing.T) {
		t.Parallel()
		_, err := policy.EvaluateRepresentativeAccess(activeParty, []*entity.PartyRole{supRole, venRole}, activeContact, svcUUID5, entity.PartyRoleKindVendor)
		if !errors.Is(err, eferrors.ErrPartyRoleInactive) {
			t.Errorf("inactive vendor role = %v, want ErrPartyRoleInactive", err)
		}
		_, err = policy.EvaluateRepresentativeAccess(activeParty, []*entity.PartyRole{supRole}, activeContact, svcUUID5, entity.PartyRoleKindCarrier)
		if !errors.Is(err, eferrors.ErrPartyRoleNotFound) {
			t.Errorf("missing carrier role = %v, want ErrPartyRoleNotFound", err)
		}
	})

	t.Run("rejects inactive or deleted party", func(t *testing.T) {
		t.Parallel()
		inactiveParty := activeParty.Clone()
		_ = inactiveParty.Suspend(now)
		_, err := policy.EvaluateRepresentativeAccess(inactiveParty, []*entity.PartyRole{supRole}, activeContact, svcUUID5, "")
		if !errors.Is(err, eferrors.ErrPartyInactive) {
			t.Errorf("suspended party = %v, want ErrPartyInactive", err)
		}
	})

	t.Run("rejects revoked contact", func(t *testing.T) {
		t.Parallel()
		revoked := activeContact.Clone()
		_ = revoked.RevokePortalAccess(now)
		_, err := policy.EvaluateRepresentativeAccess(activeParty, []*entity.PartyRole{supRole}, revoked, svcUUID5, entity.PartyRoleKindSupplier)
		if !errors.Is(err, eferrors.ErrPortalAccessNotActive) {
			t.Errorf("revoked contact = %v, want ErrPortalAccessNotActive", err)
		}
	})
}

func TestPortalAccessPolicy_CanTransitionPortalStatus(t *testing.T) {
	t.Parallel()

	policy := efservice.NewPortalAccessPolicy()

	tests := []struct {
		name     string
		current  entity.PortalAccessStatus
		next     entity.PortalAccessStatus
		sameUser bool
		wantErr  error
	}{
		{
			name:     "none to pending_link allowed",
			current:  entity.PortalAccessStatusNone,
			next:     entity.PortalAccessStatusPendingLink,
			sameUser: false,
		},
		{
			name:     "pending_link to active same user allowed",
			current:  entity.PortalAccessStatusPendingLink,
			next:     entity.PortalAccessStatusActive,
			sameUser: true,
		},
		{
			name:     "pending_link to active different user rejected",
			current:  entity.PortalAccessStatusPendingLink,
			next:     entity.PortalAccessStatusActive,
			sameUser: false,
			wantErr:  eferrors.ErrConflict,
		},
		{
			name:     "active to active different user rejected",
			current:  entity.PortalAccessStatusActive,
			next:     entity.PortalAccessStatusActive,
			sameUser: false,
			wantErr:  eferrors.ErrConflict,
		},
		{
			name:     "active to revoked same user allowed",
			current:  entity.PortalAccessStatusActive,
			next:     entity.PortalAccessStatusRevoked,
			sameUser: true,
		},
		{
			name:     "revoked to active allowed",
			current:  entity.PortalAccessStatusRevoked,
			next:     entity.PortalAccessStatusActive,
			sameUser: true,
		},
		{
			name:     "invalid status rejected",
			current:  entity.PortalAccessStatus("unknown"),
			next:     entity.PortalAccessStatusActive,
			sameUser: true,
			wantErr:  eferrors.ErrInvalidPortalAccessStatus,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := policy.CanTransitionPortalStatus(tc.current, tc.next, tc.sameUser)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("CanTransitionPortalStatus() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CanTransitionPortalStatus() unexpected error: %v", err)
			}
		})
	}
}
