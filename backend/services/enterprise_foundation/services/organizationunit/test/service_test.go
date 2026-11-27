package organizationunit_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/organizationunit"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type seqIDGen struct{ seq int }

func (g *seqIDGen) NewUUID() (string, error) {
	g.seq++
	return fmt.Sprintf("77777777-7777-4777-8777-%012d", g.seq), nil
}

type noopTransactor struct{}

func (noopTransactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type stubLegalEntityRepo struct {
	items map[string]*entity.LegalEntity
}

func (r *stubLegalEntityRepo) SelectByID(_ context.Context, id string) (*entity.LegalEntity, error) {
	if le, ok := r.items[id]; ok && !le.IsDeleted() {
		return le.Clone(), nil
	}
	return nil, eferrors.ErrLegalEntityNotFound
}
func (r *stubLegalEntityRepo) SelectByIDs(context.Context, []string) ([]*entity.LegalEntity, error) {
	return nil, nil
}
func (r *stubLegalEntityRepo) SelectByCode(context.Context, string) (*entity.LegalEntity, error) {
	return nil, eferrors.ErrLegalEntityNotFound
}
func (r *stubLegalEntityRepo) SelectByTaxIDNumber(context.Context, string) (*entity.LegalEntity, error) {
	return nil, eferrors.ErrLegalEntityNotFound
}
func paginateMem[T any](matched []T, page domainrepo.PageRequest) domainrepo.PageResult[T] {
	norm := page.Normalize()
	total := int64(len(matched))
	if norm.Offset >= len(matched) {
		return domainrepo.NewPageResult([]T{}, total, norm)
	}
	end := norm.Offset + norm.Limit
	if end > len(matched) {
		end = len(matched)
	}
	return domainrepo.NewPageResult(matched[norm.Offset:end], total, norm)
}

func (r *stubLegalEntityRepo) SelectByNIB(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	return domainrepo.PageResult[*entity.LegalEntity]{}, nil
}
func (r *stubLegalEntityRepo) SelectPrimaryOperating(context.Context) (*entity.LegalEntity, error) {
	return nil, eferrors.ErrLegalEntityNotFound
}
func (r *stubLegalEntityRepo) SelectList(context.Context, domainrepo.LegalEntityFilter) (domainrepo.PageResult[*entity.LegalEntity], error) {
	return domainrepo.PageResult[*entity.LegalEntity]{}, nil
}
func (r *stubLegalEntityRepo) SelectParentID(context.Context, string) (*string, error) {
	return nil, nil
}
func (r *stubLegalEntityRepo) SelectChildrenByParentID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	return domainrepo.PageResult[*entity.LegalEntity]{}, nil
}
func (r *stubLegalEntityRepo) CountActiveDependents(context.Context, string) (int64, error) {
	return 0, nil
}
func (r *stubLegalEntityRepo) CheckCodeExists(context.Context, string, string) (bool, error) {
	return false, nil
}
func (r *stubLegalEntityRepo) CheckTaxIDExists(context.Context, string, string) (bool, error) {
	return false, nil
}
func (r *stubLegalEntityRepo) Insert(context.Context, *entity.LegalEntity) error { return nil }
func (r *stubLegalEntityRepo) Update(context.Context, *entity.LegalEntity) error { return nil }
func (r *stubLegalEntityRepo) UpdateClearPrimaryOperatingExcept(context.Context, string, time.Time) error {
	return nil
}
func (r *stubLegalEntityRepo) SoftDeleteByID(context.Context, string, time.Time) error { return nil }

type stubLocationRepo struct {
	items map[string]*entity.Location
}

func (r *stubLocationRepo) SelectByID(_ context.Context, id string) (*entity.Location, error) {
	if loc, ok := r.items[id]; ok && !loc.IsDeleted() {
		return loc.Clone(), nil
	}
	return nil, eferrors.ErrLocationNotFound
}
func (r *stubLocationRepo) SelectByIDs(context.Context, []string) ([]*entity.Location, error) {
	return nil, nil
}
func (r *stubLocationRepo) SelectByCode(context.Context, string) (*entity.Location, error) {
	return nil, eferrors.ErrLocationNotFound
}
func (r *stubLocationRepo) SelectByLegalEntityAndKind(context.Context, string, entity.LocationKind, bool, domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	return domainrepo.PageResult[*entity.Location]{}, nil
}
func (r *stubLocationRepo) SelectByAddressID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	return domainrepo.PageResult[*entity.Location]{}, nil
}
func (r *stubLocationRepo) SelectList(context.Context, domainrepo.LocationFilter) (domainrepo.PageResult[*entity.Location], error) {
	return domainrepo.PageResult[*entity.Location]{}, nil
}
func (r *stubLocationRepo) SelectParentID(context.Context, string) (*string, error) {
	return nil, nil
}
func (r *stubLocationRepo) SelectChildrenByParentID(context.Context, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	return domainrepo.PageResult[*entity.Location]{}, nil
}
func (r *stubLocationRepo) CountActiveDependents(context.Context, string) (int64, error) {
	return 0, nil
}
func (r *stubLocationRepo) CheckCodeExists(context.Context, string, string) (bool, error) {
	return false, nil
}
func (r *stubLocationRepo) Insert(context.Context, *entity.Location) error          { return nil }
func (r *stubLocationRepo) Update(context.Context, *entity.Location) error          { return nil }
func (r *stubLocationRepo) SoftDeleteByID(context.Context, string, time.Time) error { return nil }

type stubOutboxRepo struct {
	events []*entity.OutboxEvent
}

func (r *stubOutboxRepo) Insert(_ context.Context, event *entity.OutboxEvent) error {
	r.events = append(r.events, event)
	return nil
}
func (r *stubOutboxRepo) SelectByID(context.Context, string) (*entity.OutboxEvent, error) {
	return nil, eferrors.ErrOutboxEventNotFound
}
func (r *stubOutboxRepo) SelectByAggregate(context.Context, string, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.OutboxEvent], error) {
	return domainrepo.PageResult[*entity.OutboxEvent]{}, nil
}
func (r *stubOutboxRepo) SelectDuePending(context.Context, time.Time, int) ([]*entity.OutboxEvent, error) {
	return nil, nil
}
func (r *stubOutboxRepo) UpdateEffective(context.Context, string, time.Time) error { return nil }
func (r *stubOutboxRepo) UpdateRetry(context.Context, string, int, time.Time, string) error {
	return nil
}
func (r *stubOutboxRepo) UpdateFailed(context.Context, string, string) error { return nil }
func (r *stubOutboxRepo) CountPending(context.Context) (int64, error)        { return 0, nil }
func (r *stubOutboxRepo) CountFailed(context.Context) (int64, error)         { return 0, nil }
func (r *stubOutboxRepo) DeleteEffectiveBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type memOrgUnitRepo struct {
	order          []string
	items          map[string]*entity.OrganizationUnit
	dependentsByID map[string]int64
	insertCalls    int
}

func newMemOrgUnitRepo() *memOrgUnitRepo {
	return &memOrgUnitRepo{
		items:          make(map[string]*entity.OrganizationUnit),
		dependentsByID: make(map[string]int64),
	}
}

func (r *memOrgUnitRepo) seed(u *entity.OrganizationUnit) {
	r.order = append(r.order, u.ID)
	r.items[u.ID] = u.Clone()
}

func (r *memOrgUnitRepo) SelectByID(_ context.Context, id string) (*entity.OrganizationUnit, error) {
	u, ok := r.items[id]
	if !ok || u.IsDeleted() {
		return nil, eferrors.ErrOrgUnitNotFound
	}
	return u.Clone(), nil
}

func (r *memOrgUnitRepo) SelectByIDs(_ context.Context, ids []string) ([]*entity.OrganizationUnit, error) {
	out := make([]*entity.OrganizationUnit, 0, len(ids))
	for _, id := range ids {
		if u, ok := r.items[id]; ok && !u.IsDeleted() {
			out = append(out, u.Clone())
		}
	}
	return out, nil
}

func (r *memOrgUnitRepo) SelectByEntityAndCode(_ context.Context, legalEntityID, unitCode string) (*entity.OrganizationUnit, error) {
	for _, id := range r.order {
		u := r.items[id]
		if !u.IsDeleted() && u.LegalEntityID == legalEntityID && u.UnitCode == unitCode {
			return u.Clone(), nil
		}
	}
	return nil, eferrors.ErrOrgUnitNotFound
}

func (r *memOrgUnitRepo) SelectByHeadPersonID(_ context.Context, headPersonID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	var out []*entity.OrganizationUnit
	for _, id := range r.order {
		u := r.items[id]
		if !u.IsDeleted() && u.HeadPersonID != nil && *u.HeadPersonID == headPersonID {
			out = append(out, u.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memOrgUnitRepo) SelectByPrimaryLocationID(_ context.Context, primaryLocationID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	var out []*entity.OrganizationUnit
	for _, id := range r.order {
		u := r.items[id]
		if !u.IsDeleted() && u.PrimaryLocationID != nil && *u.PrimaryLocationID == primaryLocationID {
			out = append(out, u.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memOrgUnitRepo) SelectList(_ context.Context, filter domainrepo.OrganizationUnitFilter) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	norm := filter.Normalize()
	var matched []*entity.OrganizationUnit
	q := strings.ToLower(strings.TrimSpace(norm.Query))
	for _, id := range r.order {
		u := r.items[id]
		if !norm.IncludeDeleted && u.IsDeleted() {
			continue
		}
		if norm.LegalEntityID != "" && u.LegalEntityID != norm.LegalEntityID {
			continue
		}
		if norm.Kind != nil && u.Kind != *norm.Kind {
			continue
		}
		if norm.ActiveOnly && !u.IsActive {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(u.UnitCode+" "+u.Name), q) {
			continue
		}
		matched = append(matched, u.Clone())
	}
	return paginateMem(matched, norm.Page), nil
}

func (r *memOrgUnitRepo) SelectParentID(_ context.Context, id string) (*string, error) {
	u, ok := r.items[id]
	if !ok || u.IsDeleted() {
		return nil, eferrors.ErrOrgUnitNotFound
	}
	if u.ParentUnitID == nil {
		return nil, nil
	}
	v := *u.ParentUnitID
	return &v, nil
}

func (r *memOrgUnitRepo) SelectChildrenByParentID(_ context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	var out []*entity.OrganizationUnit
	for _, id := range r.order {
		u := r.items[id]
		if !u.IsDeleted() && u.ParentUnitID != nil && *u.ParentUnitID == parentID {
			out = append(out, u.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memOrgUnitRepo) CountActiveDependents(_ context.Context, id string) (int64, error) {
	return r.dependentsByID[id], nil
}

func (r *memOrgUnitRepo) CheckCodeExistsInEntity(_ context.Context, legalEntityID, unitCode, excludeID string) (bool, error) {
	for _, id := range r.order {
		u := r.items[id]
		if !u.IsDeleted() && u.ID != excludeID && u.LegalEntityID == legalEntityID && u.UnitCode == unitCode {
			return true, nil
		}
	}
	return false, nil
}

func (r *memOrgUnitRepo) Insert(_ context.Context, u *entity.OrganizationUnit) error {
	r.insertCalls++
	r.seed(u)
	return nil
}

func (r *memOrgUnitRepo) Update(_ context.Context, u *entity.OrganizationUnit) error {
	if _, ok := r.items[u.ID]; !ok {
		return eferrors.ErrOrgUnitNotFound
	}
	r.items[u.ID] = u.Clone()
	return nil
}

func (r *memOrgUnitRepo) SoftDeleteByID(_ context.Context, id string, deletedAt time.Time) error {
	u, ok := r.items[id]
	if !ok {
		return eferrors.ErrOrgUnitNotFound
	}
	u.DeletedAt = &deletedAt
	u.IsActive = false
	u.UpdatedAt = deletedAt
	return nil
}

func newTestService(t *testing.T) (*organizationunit.Service, *memOrgUnitRepo, *stubLegalEntityRepo, *stubLocationRepo, *stubOutboxRepo) {
	t.Helper()
	repo := newMemOrgUnitRepo()
	leRepo := &stubLegalEntityRepo{items: make(map[string]*entity.LegalEntity)}
	locRepo := &stubLocationRepo{items: make(map[string]*entity.Location)}
	outbox := &stubOutboxRepo{}
	svc, err := organizationunit.NewService(organizationunit.Dependencies{
		Transactor:        noopTransactor{},
		OrganizationUnits: repo,
		LegalEntities:     leRepo,
		Locations:         locRepo,
		Outbox:            outbox,
		Clock:             fixedClock{now: time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)},
		IDGen:             &seqIDGen{},
	})
	if err != nil {
		t.Fatalf("NewService unexpected error: %v", err)
	}
	return svc, repo, leRepo, locRepo, outbox
}

func TestNewServiceFailClosedAndNilReceiver(t *testing.T) {
	base := organizationunit.Dependencies{
		Transactor:        noopTransactor{},
		OrganizationUnits: newMemOrgUnitRepo(),
		LegalEntities:     &stubLegalEntityRepo{},
		Locations:         &stubLocationRepo{},
		Outbox:            &stubOutboxRepo{},
		IDGen:             &seqIDGen{},
	}

	cases := []struct {
		name  string
		mut   func(d *organizationunit.Dependencies)
		field string
	}{
		{"nil transactor", func(d *organizationunit.Dependencies) { d.Transactor = nil }, "transactor"},
		{"nil organization_units", func(d *organizationunit.Dependencies) { d.OrganizationUnits = nil }, "organization_units"},
		{"nil legal_entities", func(d *organizationunit.Dependencies) { d.LegalEntities = nil }, "legal_entities"},
		{"nil locations", func(d *organizationunit.Dependencies) { d.Locations = nil }, "locations"},
		{"nil outbox", func(d *organizationunit.Dependencies) { d.Outbox = nil }, "outbox"},
		{"nil id_gen", func(d *organizationunit.Dependencies) { d.IDGen = nil }, "id_gen"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := base
			tc.mut(&deps)
			_, err := organizationunit.NewService(deps)
			if err == nil || eferrors.KindOf(err) != eferrors.KindInvalidInput || eferrors.FieldOf(err) != tc.field {
				t.Fatalf("expected invalid input for %s, got %v", tc.field, err)
			}
		})
	}

	var nilSvc *organizationunit.Service
	if _, err := nilSvc.GetAll(context.Background(), domainrepo.PageRequest{}); eferrors.KindOf(err) != eferrors.KindUnavailable {
		t.Fatalf("expected KindUnavailable on nil receiver, got %v", err)
	}
}

func TestOrganizationUnitCRUDCrossEntityScopeAndGetAll(t *testing.T) {
	ctx := context.Background()
	svc, repo, leRepo, locRepo, outbox := newTestService(t)

	le1ID := "88888888-8888-4888-8888-888888888881"
	le2ID := "88888888-8888-4888-8888-888888888882"
	for _, spec := range []struct{ id, code string }{{le1ID, "le-1"}, {le2ID, "le-2"}} {
		le, err := entity.NewLegalEntity(spec.id, entity.LegalEntityParams{
			EntityCode:              spec.code,
			LegalName:               "PT " + spec.code,
			Status:                  entity.LegalEntityStatusActive,
			JurisdictionCountryCode: "ID",
			FiscalYearStartMonth:    1,
		}, time.Now().UTC())
		if err != nil {
			t.Fatalf("NewLegalEntity failed: %v", err)
		}
		leRepo.items[spec.id] = le
	}

	loc1ID := "99999999-9999-4999-8999-999999999991"
	loc2ID := "99999999-9999-4999-8999-999999999992"
	for _, spec := range []struct{ id, leID, code string }{{loc1ID, le1ID, "hq-1"}, {loc2ID, le2ID, "hq-2"}} {
		loc, err := entity.NewLocation(spec.id, entity.LocationParams{
			LegalEntityID: spec.leID,
			LocationCode:  spec.code,
			Name:          "Kantor " + spec.code,
			Kind:          entity.LocationKindHeadquarters,
			Timezone:      "Asia/Jakarta",
		}, time.Now().UTC())
		if err != nil {
			t.Fatalf("NewLocation failed: %v", err)
		}
		locRepo.items[spec.id] = loc
	}

	headID := "aaaaaaaa-0000-4000-8000-000000000001"
	div, err := svc.Create(ctx, organizationunit.CreateInput{
		LegalEntityID:     le1ID,
		PrimaryLocationID: &loc1ID,
		UnitCode:          "FIN-OPS",
		Name:              "Divisi Keuangan",
		Kind:              entity.OrgUnitKindDivision,
		HeadPersonID:      &headID,
	})
	if err != nil {
		t.Fatalf("Create division failed: %v", err)
	}

	// Kode sama pada badan hukum BERBEDA diizinkan.
	_, err = svc.Create(ctx, organizationunit.CreateInput{
		LegalEntityID:     le2ID,
		PrimaryLocationID: &loc2ID,
		UnitCode:          "FIN-OPS",
		Name:              "Divisi Keuangan Entitas 2",
		Kind:              entity.OrgUnitKindDivision,
	})
	if err != nil {
		t.Fatalf("expected same unit code allowed in different legal entity, got %v", err)
	}

	callsBefore := repo.insertCalls

	// Kode sama pada badan hukum yang SAMA ditolak sebelum Insert.
	_, err = svc.Create(ctx, organizationunit.CreateInput{
		LegalEntityID: le1ID,
		UnitCode:      " fin-ops ",
		Name:          "Duplikat Unit",
		Kind:          entity.OrgUnitKindDepartment,
	})
	if !errors.Is(err, eferrors.ErrOrgUnitCodeExists) {
		t.Fatalf("expected ErrOrgUnitCodeExists, got %v", err)
	}

	// Lokasi utama milik entitas lain ditolak sebelum Insert.
	_, err = svc.Create(ctx, organizationunit.CreateInput{
		LegalEntityID:     le1ID,
		PrimaryLocationID: &loc2ID,
		UnitCode:          "OPS-SUB",
		Name:              "Sub Unit",
		Kind:              entity.OrgUnitKindTeam,
	})
	if !errors.Is(err, eferrors.ErrOrgUnitLocationMismatch) {
		t.Fatalf("expected ErrOrgUnitLocationMismatch, got %v", err)
	}

	// Unit induk milik entitas lain ditolak sebelum Insert.
	_, err = svc.Create(ctx, organizationunit.CreateInput{
		LegalEntityID: le2ID,
		ParentUnitID:  &div.ID,
		UnitCode:      "TEAM-X",
		Name:          "Tim Lintas Entitas",
		Kind:          entity.OrgUnitKindTeam,
	})
	if !errors.Is(err, eferrors.ErrOrgUnitParentMismatch) {
		t.Fatalf("expected ErrOrgUnitParentMismatch, got %v", err)
	}
	if repo.insertCalls != callsBefore {
		t.Fatalf("expected Insert not called on invalid creates, before=%d after=%d", callsBefore, repo.insertCalls)
	}

	// Buat sub-unit valid lalu uji deteksi siklus pada Update.
	dept, err := svc.Create(ctx, organizationunit.CreateInput{
		LegalEntityID:     le1ID,
		ParentUnitID:      &div.ID,
		PrimaryLocationID: &loc1ID,
		UnitCode:          "ACC-AP",
		Name:              "Departemen Hutang Dagang",
		Kind:              entity.OrgUnitKindDepartment,
	})
	if err != nil {
		t.Fatalf("Create dept failed: %v", err)
	}

	_, err = svc.Update(ctx, organizationunit.UpdateInput{
		ID:           div.ID,
		ParentUnitID: &dept.ID,
		Name:         div.Name,
		Kind:         div.Kind,
	})
	if !errors.Is(err, eferrors.ErrOrgUnitCycleDetected) {
		t.Fatalf("expected ErrOrgUnitCycleDetected, got %v", err)
	}

	// AssignHeadPerson, query helpers, scope check.
	newHeadID := "aaaaaaaa-0000-4000-8000-000000000002"
	dept, err = svc.AssignHeadPerson(ctx, dept.ID, &newHeadID)
	if err != nil || dept.HeadPersonID == nil || *dept.HeadPersonID != newHeadID {
		t.Fatalf("AssignHeadPerson failed: got=%+v err=%v", dept, err)
	}
	if got, err := svc.GetByEntityAndCode(ctx, le1ID, "acc-ap"); err != nil || got.ID != dept.ID {
		t.Fatalf("GetByEntityAndCode failed: got=%+v err=%v", got, err)
	}
	if list, err := svc.ListByHeadPersonID(ctx, newHeadID, domainrepo.PageRequest{}); err != nil || list.Total != 1 || len(list.Items) != 1 {
		t.Fatalf("ListByHeadPersonID failed: page=%+v err=%v", list, err)
	}
	if list, err := svc.ListByPrimaryLocationID(ctx, loc1ID, domainrepo.PageRequest{}); err != nil || list.Total != 2 || len(list.Items) != 2 {
		t.Fatalf("ListByPrimaryLocationID failed: page=%+v err=%v", list, err)
	}
	if locP1, err := svc.ListByPrimaryLocationID(ctx, loc1ID, domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || len(locP1.Items) != 1 || locP1.Total != 2 || !locP1.HasMore {
		t.Fatalf("ListByPrimaryLocationID page 1 failed: page=%+v err=%v", locP1, err)
	}
	if children, err := svc.ListChildren(ctx, div.ID, domainrepo.PageRequest{}); err != nil || children.Total != 1 || len(children.Items) != 1 {
		t.Fatalf("ListChildren failed: page=%+v err=%v", children, err)
	}
	if dec, err := svc.ValidateScopeTarget(ctx, dept.ID); err != nil || !dec.Eligible {
		t.Fatalf("ValidateScopeTarget failed: dec=%+v err=%v", dec, err)
	}

	// Deactivate menggagalkan ValidateScopeTarget, Activate memulihkan, Delete menolak bila ada tanggungan.
	if _, err := svc.Deactivate(ctx, dept.ID); err != nil {
		t.Fatalf("Deactivate failed: %v", err)
	}
	if _, err := svc.ValidateScopeTarget(ctx, dept.ID); !errors.Is(err, eferrors.ErrOrgUnitInactive) {
		t.Fatalf("expected ErrOrgUnitInactive, got %v", err)
	}
	if _, err := svc.Activate(ctx, dept.ID); err != nil {
		t.Fatalf("Activate failed: %v", err)
	}

	repo.dependentsByID[div.ID] = 1
	if err := svc.Delete(ctx, div.ID); err == nil || eferrors.KindOf(err) != eferrors.KindConflict {
		t.Fatalf("expected conflict on delete with dependents, got %v", err)
	}
	repo.dependentsByID[dept.ID] = 0
	if err := svc.Delete(ctx, dept.ID); err != nil {
		t.Fatalf("Delete dept failed: %v", err)
	}
	if len(outbox.events) < 6 {
		t.Fatalf("expected outbox events recorded, got %d", len(outbox.events))
	}

	// GetAll paginasi multi-halaman (>100 baris).
	for i := 0; i < 105; i++ {
		id := fmt.Sprintf("bbbbbbbb-bbbb-4bbb-8bbb-%012d", i+1)
		u, err := entity.NewOrganizationUnit(id, entity.OrganizationUnitParams{
			LegalEntityID: le1ID,
			UnitCode:      fmt.Sprintf("ou-bulk-%03d", i+1),
			Name:          fmt.Sprintf("Unit Bulk %03d", i+1),
			Kind:          entity.OrgUnitKindTeam,
		}, time.Now().UTC())
		if err != nil {
			t.Fatalf("NewOrganizationUnit seed error: %v", err)
		}
		repo.seed(u)
	}
	p1, err := svc.GetAll(ctx, domainrepo.PageRequest{Limit: 50, Offset: 0})
	if err != nil || p1.Total != 107 || len(p1.Items) != 50 || !p1.HasMore {
		t.Fatalf("GetAll page 1 failed: page=%+v err=%v", p1, err)
	}
	p3, err := svc.GetAll(ctx, domainrepo.PageRequest{Limit: 50, Offset: 100})
	if err != nil || p3.Total != 107 || len(p3.Items) != 7 || p3.HasMore {
		t.Fatalf("GetAll page 3 failed: page=%+v err=%v", p3, err)
	}
}

func (r *memOrgUnitRepo) SelectByIDForUpdate(ctx context.Context, id string) (*entity.OrganizationUnit, error) {
	return r.SelectByID(ctx, id)
}
