package entity_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestLocation_LifecycleAndAttributes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 11, 0, 0, 0, time.Local)
	loc, err := entity.NewLocation(testUUID1, entity.LocationParams{
		LegalEntityID:         testUUID2,
		AddressID:             strPtr(testUUID3),
		LocationCode:          " WH-JKT-01 ",
		Name:                  " Gudang Utama Cikarang ",
		Kind:                  entity.LocationKindWarehouse,
		Phone:                 strPtr("+62 21 5550100"),
		Timezone:              "Asia/Jakarta",
		OperationalAttributes: json.RawMessage(`{"receiving_dock":2,"cold_storage":false}`),
	}, now)
	if err != nil {
		t.Fatalf("NewLocation() error = %v", err)
	}
	if loc.LocationCode != "wh-jkt-01" || !loc.IsWarehouse() || !loc.IsEligibleScope() {
		t.Fatalf("unexpected Location fields: %+v", loc)
	}

	attrsCopy := loc.OperationalAttributesCopy()
	attrsCopy[0] = 'X'
	if string(loc.OperationalAttributesCopy())[0] != '{' {
		t.Fatal("OperationalAttributesCopy() allowed internal mutation")
	}

	if err := loc.Deactivate(now.Add(time.Minute)); err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}
	if loc.IsEligibleScope() {
		t.Fatal("inactive location must not be eligible scope")
	}
	if err := loc.CanOperate(); !errors.Is(err, eferrors.ErrLocationInactive) {
		t.Fatalf("CanOperate() inactive error = %v, want ErrLocationInactive", err)
	}

	if err := loc.Activate(now.Add(2 * time.Minute)); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if err := loc.SoftDelete(now.Add(3 * time.Minute)); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	if err := loc.CanOperate(); !errors.Is(err, eferrors.ErrLocationDeleted) {
		t.Fatalf("CanOperate() deleted error = %v, want ErrLocationDeleted", err)
	}
}

func TestNewLocation_ValidationErrors(t *testing.T) {
	t.Parallel()

	now := time.Now()
	if _, err := entity.NewLocation(testUUID1, entity.LocationParams{
		LegalEntityID:    testUUID2,
		ParentLocationID: strPtr(testUUID1),
		LocationCode:     "hq-01",
		Name:             "HQ",
		Kind:             entity.LocationKindHeadquarters,
	}, now); !errors.Is(err, eferrors.ErrLocationSelfParent) {
		t.Fatalf("expected ErrLocationSelfParent, got %v", err)
	}

	if _, err := entity.NewLocation(testUUID1, entity.LocationParams{
		LegalEntityID: testUUID2,
		LocationCode:  "hq-01",
		Name:          "HQ",
		Kind:          entity.LocationKindHeadquarters,
		Timezone:      "Invalid/Timezone",
	}, now); !errors.Is(err, eferrors.ErrInvalidTimezone) {
		t.Fatalf("expected ErrInvalidTimezone, got %v", err)
	}
}
