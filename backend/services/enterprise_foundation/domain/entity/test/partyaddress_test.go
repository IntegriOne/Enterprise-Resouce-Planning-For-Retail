package entity_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestNewPartyAddress_AndMutations(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 10, 30, 0, 0, time.UTC)

	pa, err := entity.NewPartyAddress(
		strings.ToUpper(testUUID1),
		strings.ToUpper(testUUID2),
		strings.ToUpper(testUUID3),
		entity.AddressUsageKindShipping,
		true,
		now,
	)
	if err != nil {
		t.Fatalf("NewPartyAddress() unexpected error: %v", err)
	}
	if pa.ID != testUUID1 || pa.PartyID != testUUID2 || pa.AddressID != testUUID3 {
		t.Errorf("UUIDs were not normalized: %+v", pa)
	}
	if pa.Usage != entity.AddressUsageKindShipping || !pa.IsPrimary {
		t.Errorf("unexpected usage/isPrimary: %+v", pa)
	}

	if err := pa.UpdateUsage(entity.AddressUsageKindBilling, false, now.Add(time.Minute)); err != nil {
		t.Fatalf("UpdateUsage() failed: %v", err)
	}
	if pa.Usage != entity.AddressUsageKindBilling || pa.IsPrimary {
		t.Errorf("UpdateUsage() did not update fields: %+v", pa)
	}

	if err := pa.SetPrimary(true, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("SetPrimary() failed: %v", err)
	}
	if !pa.IsPrimary {
		t.Errorf("expected IsPrimary=true")
	}

	clone := pa.Clone()
	if clone == nil || clone == pa || *clone != *pa {
		t.Errorf("Clone() mismatch: got %+v want %+v", clone, pa)
	}

	t.Run("validation and nil guards", func(t *testing.T) {
		t.Parallel()
		if _, err := entity.NewPartyAddress("bad", testUUID2, testUUID3, entity.AddressUsageKindBilling, false, now); !errors.Is(err, eferrors.ErrInvalidInput) {
			t.Errorf("invalid id error = %v, want ErrInvalidInput", err)
		}
		if _, err := entity.NewPartyAddress(testUUID1, "bad", testUUID3, entity.AddressUsageKindBilling, false, now); !errors.Is(err, eferrors.ErrInvalidInput) {
			t.Errorf("invalid party_id error = %v, want ErrInvalidInput", err)
		}
		if _, err := entity.NewPartyAddress(testUUID1, testUUID2, "bad", entity.AddressUsageKindBilling, false, now); !errors.Is(err, eferrors.ErrInvalidInput) {
			t.Errorf("invalid address_id error = %v, want ErrInvalidInput", err)
		}
		if _, err := entity.NewPartyAddress(testUUID1, testUUID2, testUUID3, entity.AddressUsageKind("unknown"), false, now); !errors.Is(err, eferrors.ErrInvalidAddressUsage) {
			t.Errorf("invalid usage error = %v, want ErrInvalidAddressUsage", err)
		}
		if err := pa.UpdateUsage(entity.AddressUsageKind("invalid"), true, now); !errors.Is(err, eferrors.ErrInvalidAddressUsage) {
			t.Errorf("UpdateUsage invalid error = %v, want ErrInvalidAddressUsage", err)
		}

		var nilPA *entity.PartyAddress
		if nilPA.Clone() != nil {
			t.Errorf("nil Clone() want nil")
		}
		if err := nilPA.UpdateUsage(entity.AddressUsageKindBilling, false, now); !errors.Is(err, eferrors.ErrPartyAddressNotFound) {
			t.Errorf("nil UpdateUsage = %v, want ErrPartyAddressNotFound", err)
		}
		if err := nilPA.SetPrimary(true, now); !errors.Is(err, eferrors.ErrPartyAddressNotFound) {
			t.Errorf("nil SetPrimary = %v, want ErrPartyAddressNotFound", err)
		}
	})
}
