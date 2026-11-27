package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestNewPartyContact_Invariants(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		id      string
		params  entity.PartyContactParams
		wantErr error
		check   func(t *testing.T, c *entity.PartyContact)
	}{
		{
			name: "directory contact without iam user",
			id:   testUUID1,
			params: entity.PartyContactParams{
				PartyID:            testUUID2,
				FullName:           " Budi Santoso ",
				JobTitle:           strPtr(" Purchasing PIC "),
				ContactRole:        "",
				Email:              strPtr(" Budi@Vendor.ID "),
				PortalAccessStatus: entity.PortalAccessStatusNone,
				IsPrimary:          true,
			},
			check: func(t *testing.T, c *entity.PartyContact) {
				t.Helper()
				if c.ContactRole != entity.DefaultContactRole {
					t.Errorf("ContactRole = %q, want default %q", c.ContactRole, entity.DefaultContactRole)
				}
				if c.Email == nil || *c.Email != "budi@vendor.id" {
					t.Errorf("Email = %v, want budi@vendor.id", c.Email)
				}
				if c.HasActivePortalAccess() {
					t.Errorf("directory contact must not have active portal access")
				}
				if err := c.ValidatePortalLinkInvariant(); err != nil {
					t.Errorf("ValidatePortalLinkInvariant() unexpected error: %v", err)
				}
			},
		},
		{
			name: "portal active contact with iam user only",
			id:   testUUID1,
			params: entity.PartyContactParams{
				PartyID:            testUUID2,
				IAMUserID:          strPtr(testUUID3),
				FullName:           "Siti Aminah",
				ContactRole:        "portal_admin",
				PortalAccessStatus: entity.PortalAccessStatusActive,
			},
			check: func(t *testing.T, c *entity.PartyContact) {
				t.Helper()
				if !c.HasActivePortalAccess() {
					t.Errorf("expected HasActivePortalAccess() == true")
				}
				if err := c.CanUsePortalAccess(); err != nil {
					t.Errorf("CanUsePortalAccess() unexpected error: %v", err)
				}
			},
		},
		{
			name: "violates contact channel check constraint when email phone iamUserID all empty",
			id:   testUUID1,
			params: entity.PartyContactParams{
				PartyID:            testUUID2,
				FullName:           "No Channel PIC",
				PortalAccessStatus: entity.PortalAccessStatusNone,
			},
			wantErr: eferrors.ErrContactChannelRequired,
		},
		{
			name: "violates portal link invariant nil iam_user_id with pending_link",
			id:   testUUID1,
			params: entity.PartyContactParams{
				PartyID:            testUUID2,
				FullName:           "Broken Pair",
				Email:              strPtr("pic@vendor.id"),
				PortalAccessStatus: entity.PortalAccessStatusPendingLink,
			},
			wantErr: eferrors.ErrPortalLinkInvariantViolation,
		},
		{
			name: "violates portal link invariant non-nil iam_user_id with none status",
			id:   testUUID1,
			params: entity.PartyContactParams{
				PartyID:            testUUID2,
				IAMUserID:          strPtr(testUUID3),
				FullName:           "Broken Pair 2",
				Email:              strPtr("pic@vendor.id"),
				PortalAccessStatus: entity.PortalAccessStatusNone,
			},
			wantErr: eferrors.ErrPortalLinkInvariantViolation,
		},
		{
			name: "invalid portal access status",
			id:   testUUID1,
			params: entity.PartyContactParams{
				PartyID:            testUUID2,
				FullName:           "PIC",
				Email:              strPtr("pic@vendor.id"),
				PortalAccessStatus: entity.PortalAccessStatus("approved"),
			},
			wantErr: eferrors.ErrInvalidPortalAccessStatus,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := entity.NewPartyContact(tc.id, tc.params, now)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("NewPartyContact() error = %v, want errors.Is(_, %v)", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewPartyContact() unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}

func TestPartyContact_PortalTransitionsAndGuards(t *testing.T) {
	t.Parallel()

	t0 := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)

	c, err := entity.NewPartyContact(testUUID1, entity.PartyContactParams{
		PartyID:            testUUID2,
		FullName:           "PIC Operasional",
		Email:              strPtr("ops@vendor.id"),
		PortalAccessStatus: entity.PortalAccessStatusNone,
	}, t0)
	if err != nil {
		t.Fatalf("setup NewPartyContact failed: %v", err)
	}

	if err := c.RequestPortalLink(testUUID3, t1); err != nil {
		t.Fatalf("RequestPortalLink() failed: %v", err)
	}
	if c.PortalAccessStatus != entity.PortalAccessStatusPendingLink || c.IAMUserID == nil || *c.IAMUserID != testUUID3 {
		t.Fatalf("unexpected state after RequestPortalLink: %+v", c)
	}
	if err := c.CanUsePortalAccess(); !errors.Is(err, eferrors.ErrPortalAccessNotActive) {
		t.Errorf("pending_link CanUsePortalAccess() = %v, want ErrPortalAccessNotActive", err)
	}

	// Menautkan user IAM berbeda harus ditolak dengan conflict.
	if err := c.ActivatePortalAccess(testUUID4, t1); !errors.Is(err, eferrors.ErrConflict) {
		t.Errorf("ActivatePortalAccess with another iamUserID = %v, want ErrConflict", err)
	}

	if err := c.ActivatePortalAccess(testUUID3, t1); err != nil {
		t.Fatalf("ActivatePortalAccess() failed: %v", err)
	}
	if !c.HasActivePortalAccess() {
		t.Errorf("expected active portal access after activation")
	}

	// Tidak boleh RequestPortalLink lagi saat sudah active.
	if err := c.RequestPortalLink(testUUID3, t1); !errors.Is(err, eferrors.ErrConflict) {
		t.Errorf("RequestPortalLink when already active = %v, want ErrConflict", err)
	}

	// Hapus semua email & telepon harus gagal jika UnlinkPortalUser (melanggar checkChannelInvariant).
	if err := c.UpdateDetails(entity.PartyContactUpdateParams{
		FullName:    "PIC Operasional",
		ContactRole: "operations",
		Email:       nil,
		Phone:       nil,
	}, t1); err != nil {
		t.Fatalf("UpdateDetails removing email/phone while linked to IAM should pass: %v", err)
	}
	if err := c.UnlinkPortalUser(t1); !errors.Is(err, eferrors.ErrContactChannelRequired) {
		t.Errorf("UnlinkPortalUser without email/phone = %v, want ErrContactChannelRequired", err)
	}

	// Kembalikan email agar bisa unlink secara sah.
	if err := c.UpdateDetails(entity.PartyContactUpdateParams{
		FullName:    "PIC Operasional",
		ContactRole: "operations",
		Email:       strPtr("ops@vendor.id"),
	}, t1); err != nil {
		t.Fatalf("UpdateDetails restoring email failed: %v", err)
	}

	if err := c.RevokePortalAccess(t1); err != nil {
		t.Fatalf("RevokePortalAccess() failed: %v", err)
	}
	if c.PortalAccessStatus != entity.PortalAccessStatusRevoked || c.IAMUserID == nil {
		t.Errorf("RevokePortalAccess must keep IAMUserID audit trail and set revoked status")
	}

	if err := c.UnlinkPortalUser(t1); err != nil {
		t.Fatalf("UnlinkPortalUser() failed: %v", err)
	}
	if c.IAMUserID != nil || c.PortalAccessStatus != entity.PortalAccessStatusNone {
		t.Errorf("UnlinkPortalUser expected nil IAMUserID and status=none")
	}
	if err := c.RevokePortalAccess(t1); !errors.Is(err, eferrors.ErrPortalLinkInvariantViolation) {
		t.Errorf("RevokePortalAccess when unlinked = %v, want ErrPortalLinkInvariantViolation", err)
	}

	// Tautkan kembali lalu uji Deactivate & SoftDelete mencabut portal aktif.
	if err := c.ActivatePortalAccess(testUUID3, t1); err != nil {
		t.Fatalf("re-ActivatePortalAccess failed: %v", err)
	}
	if err := c.Deactivate(t1); err != nil {
		t.Fatalf("Deactivate() failed: %v", err)
	}
	if err := c.CanUsePortalAccess(); !errors.Is(err, eferrors.ErrPartyContactInactive) {
		t.Errorf("inactive contact CanUsePortalAccess() = %v, want ErrPartyContactInactive", err)
	}
	if err := c.Activate(t1); err != nil {
		t.Fatalf("Activate() failed: %v", err)
	}

	clone := c.Clone()
	if clone == nil || clone == c || *clone.IAMUserID != *c.IAMUserID {
		t.Fatalf("Clone() failed")
	}

	if err := c.SoftDelete(t1); err != nil {
		t.Fatalf("SoftDelete() failed: %v", err)
	}
	if !c.IsDeleted() || c.IsActive || c.PortalAccessStatus != entity.PortalAccessStatusRevoked {
		t.Errorf("SoftDelete should mark deleted, deactivate, and revoke active portal link: %+v", c)
	}
	if err := c.CanUsePortalAccess(); !errors.Is(err, eferrors.ErrPartyContactDeleted) {
		t.Errorf("deleted contact CanUsePortalAccess() = %v, want ErrPartyContactDeleted", err)
	}
}
