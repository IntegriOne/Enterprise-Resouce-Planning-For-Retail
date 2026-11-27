package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestOrganizationUnit_Lifecycle(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 15, 11, 30, 0, 0, time.Local)
	unit, err := entity.NewOrganizationUnit(testUUID1, entity.OrganizationUnitParams{
		LegalEntityID:     testUUID2,
		PrimaryLocationID: strPtr(testUUID3),
		UnitCode:          " OPS-WH ",
		Name:              " Operasional Gudang ",
		Kind:              entity.OrgUnitKindDepartment,
		CostCenterCode:    strPtr(" CC-OPS-01 "),
		HeadPersonID:      strPtr(testUUID4),
	}, now)
	if err != nil {
		t.Fatalf("NewOrganizationUnit() error = %v", err)
	}
	if unit.UnitCode != "ops-wh" || !unit.IsEligibleScope() || *unit.CostCenterCode != "CC-OPS-01" {
		t.Fatalf("unexpected OrganizationUnit fields: %+v", unit)
	}

	if err := unit.AssignHeadPerson(nil, now.Add(time.Minute)); err != nil {
		t.Fatalf("AssignHeadPerson(nil) error = %v", err)
	}
	if unit.HeadPersonID != nil {
		t.Fatal("HeadPersonID should be nil after unassign")
	}

	if err := unit.Deactivate(now.Add(2 * time.Minute)); err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}
	if err := unit.CanOperate(); !errors.Is(err, eferrors.ErrOrgUnitInactive) {
		t.Fatalf("CanOperate() error = %v, want ErrOrgUnitInactive", err)
	}

	if err := unit.SoftDelete(now.Add(3 * time.Minute)); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	if err := unit.UpdateDetails(entity.OrganizationUnitUpdateParams{
		Name: "After Delete",
		Kind: entity.OrgUnitKindTeam,
	}, now.Add(4*time.Minute)); !errors.Is(err, eferrors.ErrOrgUnitDeleted) {
		t.Fatalf("UpdateDetails() after delete = %v, want ErrOrgUnitDeleted", err)
	}
}

func TestNewOrganizationUnit_SelfParentRejected(t *testing.T) {
	t.Parallel()

	_, err := entity.NewOrganizationUnit(testUUID1, entity.OrganizationUnitParams{
		LegalEntityID: testUUID2,
		ParentUnitID:  strPtr(testUUID1),
		UnitCode:      "fin",
		Name:          "Finance",
		Kind:          entity.OrgUnitKindDepartment,
	}, time.Now())
	if !errors.Is(err, eferrors.ErrOrgUnitSelfParent) {
		t.Fatalf("expected ErrOrgUnitSelfParent, got %v", err)
	}
}
