package entity_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestNewPartyRole_AndDualRoleSupport(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC)

	supplierRole, err := entity.NewPartyRole(
		strings.ToUpper(testUUID1),
		strings.ToUpper(testUUID2),
		entity.PartyRoleKindSupplier,
		strPtr("  pemasok bahan baku katun "),
		now,
	)
	if err != nil {
		t.Fatalf("NewPartyRole(supplier) unexpected error: %v", err)
	}
	if supplierRole.ID != testUUID1 || supplierRole.PartyID != testUUID2 {
		t.Errorf("UUIDs were not normalized: id=%q partyID=%q", supplierRole.ID, supplierRole.PartyID)
	}
	if !supplierRole.IsEligibleForScope(entity.PartyRoleKindSupplier) {
		t.Errorf("expected supplierRole to be eligible for supplier scope")
	}
	if supplierRole.IsEligibleForScope(entity.PartyRoleKindVendor) {
		t.Errorf("supplierRole must not be eligible for vendor scope")
	}

	// Satu mitra dapat memegang peran supplier sekaligus vendor secara terpisah.
	vendorRole, err := entity.NewPartyRole(testUUID3, testUUID2, entity.PartyRoleKindVendor, strPtr("vendor jahit makloon"), now)
	if err != nil {
		t.Fatalf("NewPartyRole(vendor) unexpected error: %v", err)
	}
	if !vendorRole.IsEligibleForScope(entity.PartyRoleKindVendor) {
		t.Errorf("expected vendorRole to be eligible for vendor scope")
	}

	if err := vendorRole.UpdateReferenceNote(strPtr("  "), now.Add(time.Minute)); err != nil {
		t.Fatalf("UpdateReferenceNote() failed: %v", err)
	}
	if vendorRole.ReferenceNote != nil {
		t.Errorf("blank ReferenceNote should normalize to nil, got %v", vendorRole.ReferenceNote)
	}

	if err := vendorRole.Deactivate(now.Add(2 * time.Minute)); err != nil {
		t.Fatalf("Deactivate() failed: %v", err)
	}
	if vendorRole.IsEligibleForScope(entity.PartyRoleKindVendor) {
		t.Errorf("deactivated role must fail scope eligibility")
	}
	if err := vendorRole.Activate(now.Add(3 * time.Minute)); err != nil {
		t.Fatalf("Activate() failed: %v", err)
	}
	if !vendorRole.IsEligibleForScope(entity.PartyRoleKindVendor) {
		t.Errorf("reactivated role should be eligible")
	}

	clone := vendorRole.Clone()
	if clone == nil || clone == vendorRole {
		t.Fatalf("Clone() should return non-nil distinct copy")
	}

	t.Run("validation errors", func(t *testing.T) {
		t.Parallel()
		if _, err := entity.NewPartyRole("bad", testUUID2, entity.PartyRoleKindVendor, nil, now); !errors.Is(err, eferrors.ErrInvalidInput) {
			t.Errorf("invalid id error = %v, want ErrInvalidInput", err)
		}
		if _, err := entity.NewPartyRole(testUUID1, "bad", entity.PartyRoleKindVendor, nil, now); !errors.Is(err, eferrors.ErrInvalidInput) {
			t.Errorf("invalid party_id error = %v, want ErrInvalidInput", err)
		}
		if _, err := entity.NewPartyRole(testUUID1, testUUID2, entity.PartyRoleKind("vendor_cmt"), nil, now); !errors.Is(err, eferrors.ErrInvalidPartyRoleKind) {
			t.Errorf("invalid role error = %v, want ErrInvalidPartyRoleKind", err)
		}
		var nilRole *entity.PartyRole
		if nilRole.Clone() != nil {
			t.Errorf("nil Clone() want nil")
		}
		if nilRole.IsEligibleForScope(entity.PartyRoleKindVendor) {
			t.Errorf("nil IsEligibleForScope want false")
		}
		if err := nilRole.Activate(now); !errors.Is(err, eferrors.ErrPartyRoleNotFound) {
			t.Errorf("nil Activate() = %v, want ErrPartyRoleNotFound", err)
		}
	})
}
