package entity_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func strPtr(s string) *string {
	return &s
}

func float64Ptr(v float64) *float64 {
	return &v
}

const (
	testUUID1 = "11111111-1111-4111-8111-111111111111"
	testUUID2 = "22222222-2222-4222-8222-222222222222"
	testUUID3 = "33333333-3333-4333-8333-333333333333"
	testUUID4 = "44444444-4444-4444-8444-444444444444"
)

func TestEnums_ParseAndValidate(t *testing.T) {
	t.Parallel()

	t.Run("LegalEntityStatus", func(t *testing.T) {
		t.Parallel()
		for _, v := range []entity.LegalEntityStatus{
			entity.LegalEntityStatusPending,
			entity.LegalEntityStatusActive,
			entity.LegalEntityStatusSuspended,
			entity.LegalEntityStatusDissolved,
		} {
			parsed, err := entity.ParseLegalEntityStatus(" " + v.String() + " ")
			if err != nil || parsed != v {
				t.Fatalf("ParseLegalEntityStatus(%q) = (%q, %v)", v, parsed, err)
			}
		}
		if _, err := entity.ParseLegalEntityStatus("closed"); !errors.Is(err, eferrors.ErrInvalidLegalEntityStatus) {
			t.Fatalf("expected ErrInvalidLegalEntityStatus, got %v", err)
		}
	})

	t.Run("OrgUnitKind", func(t *testing.T) {
		t.Parallel()
		for _, v := range []entity.OrgUnitKind{
			entity.OrgUnitKindDivision,
			entity.OrgUnitKindDepartment,
			entity.OrgUnitKindTeam,
			entity.OrgUnitKindBranch,
			entity.OrgUnitKindOther,
		} {
			parsed, err := entity.ParseOrgUnitKind(v.String())
			if err != nil || parsed != v {
				t.Fatalf("ParseOrgUnitKind(%q) = (%q, %v)", v, parsed, err)
			}
		}
		if _, err := entity.ParseOrgUnitKind("squad"); !errors.Is(err, eferrors.ErrInvalidOrgUnitKind) {
			t.Fatalf("expected ErrInvalidOrgUnitKind, got %v", err)
		}
	})

	t.Run("LocationKind", func(t *testing.T) {
		t.Parallel()
		for _, v := range []entity.LocationKind{
			entity.LocationKindHeadquarters,
			entity.LocationKindWarehouse,
			entity.LocationKindRetailStore,
			entity.LocationKindBranchOffice,
		} {
			parsed, err := entity.ParseLocationKind(v.String())
			if err != nil || parsed != v {
				t.Fatalf("ParseLocationKind(%q) = (%q, %v)", v, parsed, err)
			}
		}
		if _, err := entity.ParseLocationKind("factory"); !errors.Is(err, eferrors.ErrInvalidLocationKind) {
			t.Fatalf("expected ErrInvalidLocationKind, got %v", err)
		}
	})

	t.Run("PartyKindAndStatus", func(t *testing.T) {
		t.Parallel()
		for _, k := range []entity.PartyKind{
			entity.PartyKindOrganization,
			entity.PartyKindIndividual,
			entity.PartyKindFinancialEntity,
			entity.PartyKindGovernmentBody,
		} {
			parsed, err := entity.ParsePartyKind(k.String())
			if err != nil || parsed != k {
				t.Fatalf("ParsePartyKind(%q) = (%q, %v)", k, parsed, err)
			}
		}
		if _, err := entity.ParsePartyKind("bot"); !errors.Is(err, eferrors.ErrInvalidPartyKind) {
			t.Fatalf("expected ErrInvalidPartyKind, got %v", err)
		}

		for _, s := range []entity.PartyStatus{
			entity.PartyStatusPendingVerification,
			entity.PartyStatusActive,
			entity.PartyStatusSuspended,
			entity.PartyStatusInactive,
		} {
			parsed, err := entity.ParsePartyStatus(s.String())
			if err != nil || parsed != s {
				t.Fatalf("ParsePartyStatus(%q) = (%q, %v)", s, parsed, err)
			}
		}
		if _, err := entity.ParsePartyStatus("banned"); !errors.Is(err, eferrors.ErrInvalidPartyStatus) {
			t.Fatalf("expected ErrInvalidPartyStatus, got %v", err)
		}
	})

	t.Run("PartyRoleKindAndIAMScope", func(t *testing.T) {
		t.Parallel()
		for _, r := range []entity.PartyRoleKind{
			entity.PartyRoleKindSupplier,
			entity.PartyRoleKindVendor,
			entity.PartyRoleKindCustomerB2B,
			entity.PartyRoleKindCarrier,
			entity.PartyRoleKindFinancialInstitution,
			entity.PartyRoleKindOther,
		} {
			parsed, err := entity.ParsePartyRoleKind(r.String())
			if err != nil || parsed != r {
				t.Fatalf("ParsePartyRoleKind(%q) = (%q, %v)", r, parsed, err)
			}
		}
		if _, err := entity.ParsePartyRoleKind("vendor_cmt"); !errors.Is(err, eferrors.ErrInvalidPartyRoleKind) {
			t.Fatalf("expected ErrInvalidPartyRoleKind for legacy vendor_cmt, got %v", err)
		}
		if scope, ok := entity.PartyRoleKindSupplier.IAMScopeType(); !ok || scope != "supplier_party" {
			t.Fatalf("supplier IAMScopeType = (%q, %v)", scope, ok)
		}
		if scope, ok := entity.PartyRoleKindVendor.IAMScopeType(); !ok || scope != "vendor_party" {
			t.Fatalf("vendor IAMScopeType = (%q, %v)", scope, ok)
		}
		if scope, ok := entity.PartyRoleKindCarrier.IAMScopeType(); ok || scope != "" {
			t.Fatalf("carrier IAMScopeType = (%q, %v), want empty false", scope, ok)
		}
	})

	t.Run("PortalAccessAddressUsageAndOutboxStatus", func(t *testing.T) {
		t.Parallel()
		for _, p := range []entity.PortalAccessStatus{
			entity.PortalAccessStatusNone,
			entity.PortalAccessStatusPendingLink,
			entity.PortalAccessStatusActive,
			entity.PortalAccessStatusRevoked,
		} {
			parsed, err := entity.ParsePortalAccessStatus(p.String())
			if err != nil || parsed != p {
				t.Fatalf("ParsePortalAccessStatus(%q) = (%q, %v)", p, parsed, err)
			}
		}
		if _, err := entity.ParsePortalAccessStatus("invited"); !errors.Is(err, eferrors.ErrInvalidPortalAccessStatus) {
			t.Fatalf("expected ErrInvalidPortalAccessStatus, got %v", err)
		}

		for _, u := range []entity.AddressUsageKind{
			entity.AddressUsageKindRegisteredOffice,
			entity.AddressUsageKindHeadquarters,
			entity.AddressUsageKindBilling,
			entity.AddressUsageKindShipping,
			entity.AddressUsageKindOther,
		} {
			parsed, err := entity.ParseAddressUsageKind(u.String())
			if err != nil || parsed != u {
				t.Fatalf("ParseAddressUsageKind(%q) = (%q, %v)", u, parsed, err)
			}
		}
		if _, err := entity.ParseAddressUsageKind("warehouse"); !errors.Is(err, eferrors.ErrInvalidAddressUsage) {
			t.Fatalf("expected ErrInvalidAddressUsage, got %v", err)
		}

		for _, o := range []entity.OutboxStatus{
			entity.OutboxStatusPending,
			entity.OutboxStatusEffective,
			entity.OutboxStatusFailed,
		} {
			parsed, err := entity.ParseOutboxStatus(o.String())
			if err != nil || parsed != o {
				t.Fatalf("ParseOutboxStatus(%q) = (%q, %v)", o, parsed, err)
			}
		}
		if _, err := entity.ParseOutboxStatus("dead"); !errors.Is(err, eferrors.ErrInvalidOutboxStatus) {
			t.Fatalf("expected ErrInvalidOutboxStatus, got %v", err)
		}
	})
}
