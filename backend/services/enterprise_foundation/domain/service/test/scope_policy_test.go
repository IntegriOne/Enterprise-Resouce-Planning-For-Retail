package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	efservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

func TestParseScopeTargetKind(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"legal_entity", " ORGANIZATION_UNIT ", "facility", "supplier_party", "vendor_party"} {
		got, err := efservice.ParseScopeTargetKind(raw)
		if err != nil || !got.IsValid() {
			t.Errorf("ParseScopeTargetKind(%q) = (%q, %v), want valid", raw, got, err)
		}
	}
	if _, err := efservice.ParseScopeTargetKind("unknown"); !errors.Is(err, eferrors.ErrInvalidScopeTarget) {
		t.Errorf("invalid ParseScopeTargetKind error = %v, want ErrInvalidScopeTarget", err)
	}
}

func TestScopePolicy_Evaluations(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	policy := efservice.NewScopePolicy()

	activeLE, err := entity.NewLegalEntity(svcUUID1, entity.LegalEntityParams{
		EntityCode:           "pt-ione",
		LegalName:            "PT IntegriOne",
		Status:               entity.LegalEntityStatusActive,
		FiscalYearStartMonth: entity.DefaultFiscalYearStartMonth,
	}, now)
	if err != nil {
		t.Fatalf("NewLegalEntity failed: %v", err)
	}

	pendingLE, err := entity.NewLegalEntity(svcUUID2, entity.LegalEntityParams{
		EntityCode:           "pt-pending",
		LegalName:            "PT Pending",
		Status:               entity.LegalEntityStatusPending,
		FiscalYearStartMonth: entity.DefaultFiscalYearStartMonth,
	}, now)
	if err != nil {
		t.Fatalf("NewLegalEntity(pending) failed: %v", err)
	}

	whLoc, err := entity.NewLocation(svcUUID3, entity.LocationParams{
		LegalEntityID: svcUUID1,
		LocationCode:  "wh-01",
		Name:          "Gudang Utama",
		Kind:          entity.LocationKindWarehouse,
	}, now)
	if err != nil {
		t.Fatalf("NewLocation(warehouse) failed: %v", err)
	}

	hqLoc, err := entity.NewLocation(svcUUID4, entity.LocationParams{
		LegalEntityID: svcUUID1,
		LocationCode:  "hq-01",
		Name:          "Kantor Pusat",
		Kind:          entity.LocationKindHeadquarters,
	}, now)
	if err != nil {
		t.Fatalf("NewLocation(hq) failed: %v", err)
	}

	unit, err := entity.NewOrganizationUnit(svcUUID5, entity.OrganizationUnitParams{
		LegalEntityID: svcUUID1,
		UnitCode:      "ops-01",
		Name:          "Operasional",
		Kind:          entity.OrgUnitKindDepartment,
	}, now)
	if err != nil {
		t.Fatalf("NewOrganizationUnit failed: %v", err)
	}

	t.Run("legal entity eligibility", func(t *testing.T) {
		t.Parallel()
		dec, err := policy.EvaluateLegalEntity(activeLE)
		if err != nil || !dec.Eligible || dec.ScopeKind != efservice.ScopeTargetLegalEntity || dec.Code != "pt-ione" {
			t.Fatalf("EvaluateLegalEntity(active) = (%+v, %v)", dec, err)
		}
		if _, err := policy.EvaluateLegalEntity(pendingLE); !errors.Is(err, eferrors.ErrLegalEntityInactive) {
			t.Errorf("EvaluateLegalEntity(pending) = %v, want ErrLegalEntityInactive", err)
		}
		if _, err := policy.EvaluateLegalEntity(nil); !errors.Is(err, eferrors.ErrLegalEntityNotFound) {
			t.Errorf("EvaluateLegalEntity(nil) = %v, want ErrLegalEntityNotFound", err)
		}
	})

	t.Run("organization unit and owner checks", func(t *testing.T) {
		t.Parallel()
		dec, err := policy.EvaluateOrganizationUnit(unit, activeLE)
		if err != nil || !dec.Eligible || dec.ScopeKind != efservice.ScopeTargetOrganizationUnit {
			t.Fatalf("EvaluateOrganizationUnit(active) = (%+v, %v)", dec, err)
		}
		if _, err := policy.EvaluateOrganizationUnit(unit, pendingLE); !errors.Is(err, eferrors.ErrLegalEntityInactive) {
			t.Errorf("inactive owner should reject unit: got %v", err)
		}
		otherOwner, err := entity.NewLegalEntity(svcUUID2, entity.LegalEntityParams{
			EntityCode:           "pt-other",
			LegalName:            "PT Other",
			Status:               entity.LegalEntityStatusActive,
			FiscalYearStartMonth: entity.DefaultFiscalYearStartMonth,
		}, now)
		if err != nil {
			t.Fatalf("NewLegalEntity(otherOwner) failed: %v", err)
		}
		if _, err := policy.EvaluateOrganizationUnit(unit, otherOwner); !errors.Is(err, eferrors.ErrScopeTargetNotEligible) {
			t.Errorf("owner mismatch got %v, want ErrScopeTargetNotEligible", err)
		}
	})

	t.Run("facility and warehouse checks", func(t *testing.T) {
		t.Parallel()
		if dec, err := policy.EvaluateFacility(hqLoc, activeLE); err != nil || dec.ScopeKind != efservice.ScopeTargetFacility {
			t.Fatalf("EvaluateFacility(hq) = (%+v, %v)", dec, err)
		}
		if dec, err := policy.EvaluateWarehouse(whLoc, activeLE); err != nil || !dec.Eligible {
			t.Fatalf("EvaluateWarehouse(wh) = (%+v, %v)", dec, err)
		}
		if _, err := policy.EvaluateWarehouse(hqLoc, activeLE); !errors.Is(err, eferrors.ErrScopeTargetNotEligible) {
			t.Errorf("EvaluateWarehouse(hq) = %v, want ErrScopeTargetNotEligible", err)
		}
	})

	t.Run("party role evaluation for supplier and vendor", func(t *testing.T) {
		t.Parallel()
		party, err := entity.NewParty(svcUUID1, entity.PartyParams{
			PartnerCode: "mitra-01",
			Kind:        entity.PartyKindOrganization,
			DisplayName: "PT Mitra Bahan & Jahit",
			Status:      entity.PartyStatusActive,
		}, now)
		if err != nil {
			t.Fatalf("NewParty failed: %v", err)
		}
		supRole, _ := entity.NewPartyRole(svcUUID2, svcUUID1, entity.PartyRoleKindSupplier, nil, now)
		venRole, _ := entity.NewPartyRole(svcUUID3, svcUUID1, entity.PartyRoleKindVendor, nil, now)
		_ = venRole.Deactivate(now)

		dec, err := policy.EvaluatePartyRole(party, []*entity.PartyRole{supRole, venRole}, entity.PartyRoleKindSupplier)
		if err != nil || !dec.Eligible || dec.ScopeKind != efservice.ScopeTargetSupplierParty {
			t.Fatalf("EvaluatePartyRole(supplier) = (%+v, %v)", dec, err)
		}
		if _, err := policy.EvaluatePartyRole(party, []*entity.PartyRole{supRole, venRole}, entity.PartyRoleKindVendor); !errors.Is(err, eferrors.ErrPartyRoleInactive) {
			t.Errorf("EvaluatePartyRole(inactive vendor) = %v, want ErrPartyRoleInactive", err)
		}
		if _, err := policy.EvaluatePartyRole(party, []*entity.PartyRole{supRole}, entity.PartyRoleKindCarrier); !errors.Is(err, eferrors.ErrPartyRoleNotFound) {
			t.Errorf("EvaluatePartyRole(missing carrier) = %v, want ErrPartyRoleNotFound", err)
		}
		if _, err := policy.EvaluatePartyRole(party, nil, entity.PartyRoleKind("bad")); !errors.Is(err, eferrors.ErrInvalidPartyRoleKind) {
			t.Errorf("EvaluatePartyRole(invalid kind) = %v, want ErrInvalidPartyRoleKind", err)
		}

		suspendedParty := party.Clone()
		_ = suspendedParty.Suspend(now)
		if _, err := policy.EvaluatePartyRole(suspendedParty, []*entity.PartyRole{supRole}, entity.PartyRoleKindSupplier); !errors.Is(err, eferrors.ErrPartyInactive) {
			t.Errorf("EvaluatePartyRole(suspended party) = %v, want ErrPartyInactive", err)
		}
	})
}
