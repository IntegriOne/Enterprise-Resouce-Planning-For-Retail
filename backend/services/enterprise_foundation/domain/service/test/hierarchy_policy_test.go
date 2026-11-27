package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	efservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

const (
	svcUUID1 = "11111111-1111-4111-8111-111111111111"
	svcUUID2 = "22222222-2222-4222-8222-222222222222"
	svcUUID3 = "33333333-3333-4333-8333-333333333333"
	svcUUID4 = "44444444-4444-4444-8444-444444444444"
	svcUUID5 = "55555555-5555-4555-8555-555555555555"
)

func strPtr(s string) *string {
	return &s
}

func TestHierarchyPolicy_ValidateParentChain(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	policy := efservice.NewHierarchyPolicy(3)

	t.Run("nil proposed parent passes without lookup", func(t *testing.T) {
		t.Parallel()
		err := policy.ValidateParentChain(ctx, svcUUID1, nil, nil, eferrors.ErrOrgUnitSelfParent, eferrors.ErrOrgUnitCycleDetected)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("self parent returns self error", func(t *testing.T) {
		t.Parallel()
		err := policy.ValidateParentChain(ctx, svcUUID1, strPtr(svcUUID1), nil, eferrors.ErrOrgUnitSelfParent, eferrors.ErrOrgUnitCycleDetected)
		if !errors.Is(err, eferrors.ErrOrgUnitSelfParent) {
			t.Fatalf("got %v, want ErrOrgUnitSelfParent", err)
		}
	})

	t.Run("valid multi-hop ancestor chain within max depth", func(t *testing.T) {
		t.Parallel()
		parents := map[string]*string{
			svcUUID2: strPtr(svcUUID3),
			svcUUID3: nil,
		}
		err := policy.ValidateParentChain(
			ctx,
			svcUUID1,
			strPtr(svcUUID2),
			func(_ context.Context, id string) (*string, error) {
				return parents[id], nil
			},
			eferrors.ErrOrgUnitSelfParent,
			eferrors.ErrOrgUnitCycleDetected,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("multi-hop cycle back to target detects cycle", func(t *testing.T) {
		t.Parallel()
		parents := map[string]*string{
			svcUUID2: strPtr(svcUUID3),
			svcUUID3: strPtr(svcUUID1),
		}
		err := policy.ValidateParentChain(
			ctx,
			svcUUID1,
			strPtr(svcUUID2),
			func(_ context.Context, id string) (*string, error) {
				return parents[id], nil
			},
			eferrors.ErrLegalEntitySelfParent,
			eferrors.ErrLegalEntityCycleDetected,
		)
		if !errors.Is(err, eferrors.ErrLegalEntityCycleDetected) {
			t.Fatalf("got %v, want ErrLegalEntityCycleDetected", err)
		}
	})

	t.Run("depth overflow fails closed with cycle error", func(t *testing.T) {
		t.Parallel()
		parents := map[string]*string{
			svcUUID2: strPtr(svcUUID3),
			svcUUID3: strPtr(svcUUID4),
			svcUUID4: strPtr(svcUUID5),
			svcUUID5: nil,
		}
		err := policy.ValidateParentChain(
			ctx,
			svcUUID1,
			strPtr(svcUUID2),
			func(_ context.Context, id string) (*string, error) {
				return parents[id], nil
			},
			eferrors.ErrLocationSelfParent,
			eferrors.ErrLocationCycleDetected,
		)
		if !errors.Is(err, eferrors.ErrLocationCycleDetected) {
			t.Fatalf("depth overflow got %v, want ErrLocationCycleDetected", err)
		}
	})

	t.Run("canceled context and lookup failure propagate errors", func(t *testing.T) {
		t.Parallel()
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()
		err := policy.ValidateParentChain(
			canceledCtx,
			svcUUID1,
			strPtr(svcUUID2),
			func(_ context.Context, _ string) (*string, error) {
				return nil, nil
			},
			eferrors.ErrOrgUnitSelfParent,
			eferrors.ErrOrgUnitCycleDetected,
		)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("canceled context got %v, want context.Canceled", err)
		}

		boom := errors.New("db unavailable")
		err = policy.ValidateParentChain(
			ctx,
			svcUUID1,
			strPtr(svcUUID2),
			func(_ context.Context, _ string) (*string, error) {
				return nil, boom
			},
			eferrors.ErrOrgUnitSelfParent,
			eferrors.ErrOrgUnitCycleDetected,
		)
		if !errors.Is(err, boom) {
			t.Errorf("lookup error got %v, want boom", err)
		}
	})

	t.Run("default max depth fallback on non-positive config", func(t *testing.T) {
		t.Parallel()
		p := efservice.NewHierarchyPolicy(0)
		if p.MaxDepth() != efservice.DefaultMaxHierarchyDepth {
			t.Errorf("MaxDepth() = %d, want %d", p.MaxDepth(), efservice.DefaultMaxHierarchyDepth)
		}
	})
}

func TestHierarchyPolicy_CrossEntityAndRemovalGuards(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	policy := efservice.NewDefaultHierarchyPolicy()

	locSame, err := entity.NewLocation(svcUUID2, entity.LocationParams{
		LegalEntityID: svcUUID1,
		LocationCode:  "wh-jkt",
		Name:          "Gudang Jakarta",
		Kind:          entity.LocationKindWarehouse,
	}, now)
	if err != nil {
		t.Fatalf("NewLocation(locSame) failed: %v", err)
	}

	locOther, err := entity.NewLocation(svcUUID3, entity.LocationParams{
		LegalEntityID: svcUUID4,
		LocationCode:  "wh-bdg",
		Name:          "Gudang Bandung",
		Kind:          entity.LocationKindWarehouse,
	}, now)
	if err != nil {
		t.Fatalf("NewLocation(locOther) failed: %v", err)
	}

	unitOther, err := entity.NewOrganizationUnit(svcUUID5, entity.OrganizationUnitParams{
		LegalEntityID: svcUUID4,
		UnitCode:      "fin-01",
		Name:          "Finance",
		Kind:          entity.OrgUnitKindDepartment,
	}, now)
	if err != nil {
		t.Fatalf("NewOrganizationUnit(unitOther) failed: %v", err)
	}

	if err := policy.EnsureSameLegalEntityForLocation(svcUUID1, locSame); err != nil {
		t.Errorf("same entity location failed: %v", err)
	}
	if err := policy.EnsureSameLegalEntityForLocation(svcUUID1, locOther); !errors.Is(err, eferrors.ErrLocationParentMismatch) {
		t.Errorf("mismatched parent location = %v, want ErrLocationParentMismatch", err)
	}
	if err := policy.EnsureSameLegalEntityForOrgUnit(svcUUID1, unitOther, nil); !errors.Is(err, eferrors.ErrOrgUnitParentMismatch) {
		t.Errorf("mismatched parent org unit = %v, want ErrOrgUnitParentMismatch", err)
	}
	if err := policy.EnsureSameLegalEntityForOrgUnit(svcUUID1, nil, locOther); !errors.Is(err, eferrors.ErrOrgUnitLocationMismatch) {
		t.Errorf("mismatched primary location = %v, want ErrOrgUnitLocationMismatch", err)
	}

	if err := locOther.SoftDelete(now); err != nil {
		t.Fatalf("SoftDelete locOther failed: %v", err)
	}
	if err := policy.EnsureSameLegalEntityForLocation(svcUUID4, locOther); !errors.Is(err, eferrors.ErrLocationDeleted) {
		t.Errorf("deleted parent location = %v, want ErrLocationDeleted", err)
	}

	primaryLE, err := entity.NewLegalEntity(svcUUID1, entity.LegalEntityParams{
		EntityCode:               "pt-integri-one",
		LegalName:                "PT IntegriOne Ritel",
		IsPrimaryOperatingEntity: true,
		Status:                   entity.LegalEntityStatusActive,
		FiscalYearStartMonth:     entity.DefaultFiscalYearStartMonth,
	}, now)
	if err != nil {
		t.Fatalf("NewLegalEntity(primary) failed: %v", err)
	}
	if err := policy.ValidateLegalEntityRemoval(primaryLE, 0); !errors.Is(err, eferrors.ErrPrimaryLegalEntityExists) {
		t.Errorf("removing primary legal entity = %v, want ErrPrimaryLegalEntityExists", err)
	}

	subsidiaryLE, err := entity.NewLegalEntity(svcUUID2, entity.LegalEntityParams{
		EntityCode:               "pt-child",
		LegalName:                "PT Anak Usaha",
		IsPrimaryOperatingEntity: false,
		Status:                   entity.LegalEntityStatusActive,
		FiscalYearStartMonth:     entity.DefaultFiscalYearStartMonth,
	}, now)
	if err != nil {
		t.Fatalf("NewLegalEntity(subsidiary) failed: %v", err)
	}
	if err := policy.ValidateLegalEntityRemoval(subsidiaryLE, 2); !errors.Is(err, eferrors.ErrLegalEntityHasDependents) {
		t.Errorf("removing LE with active dependents = %v, want ErrLegalEntityHasDependents", err)
	}
	if err := policy.ValidateLegalEntityRemoval(subsidiaryLE, 0); err != nil {
		t.Errorf("removing non-primary LE without dependents failed: %v", err)
	}
}
