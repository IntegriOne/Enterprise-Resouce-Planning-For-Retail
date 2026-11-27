package location_test

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
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/location"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type seqIDGen struct{ seq int }

func (g *seqIDGen) NewUUID() (string, error) {
	g.seq++
	return fmt.Sprintf("33333333-3333-4333-8333-%012d", g.seq), nil
}

type noopTransactor struct{}

func (noopTransactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type stubAddressRepo struct {
	items map[string]*entity.Address
}

func (r *stubAddressRepo) SelectByID(_ context.Context, id string) (*entity.Address, error) {
	if addr, ok := r.items[id]; ok {
		return addr.Clone(), nil
	}
	return nil, eferrors.ErrAddressNotFound
}
func (r *stubAddressRepo) SelectByIDs(context.Context, []string) ([]*entity.Address, error) {
	return nil, nil
}
func (r *stubAddressRepo) SelectList(context.Context, domainrepo.AddressFilter) (domainrepo.PageResult[*entity.Address], error) {
	return domainrepo.PageResult[*entity.Address]{}, nil
}
func (r *stubAddressRepo) CheckReferenced(context.Context, string) (bool, error) { return false, nil }
func (r *stubAddressRepo) Insert(context.Context, *entity.Address) error         { return nil }
func (r *stubAddressRepo) Update(context.Context, *entity.Address) error         { return nil }
func (r *stubAddressRepo) DeleteByID(context.Context, string) error              { return nil }

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

type memLocationRepo struct {
	order          []string
	items          map[string]*entity.Location
	dependentsByID map[string]int64
	insertCalls    int
}

func newMemLocationRepo() *memLocationRepo {
	return &memLocationRepo{
		items:          make(map[string]*entity.Location),
		dependentsByID: make(map[string]int64),
	}
}

func (r *memLocationRepo) seed(loc *entity.Location) {
	r.order = append(r.order, loc.ID)
	r.items[loc.ID] = loc.Clone()
}

func (r *memLocationRepo) SelectByID(_ context.Context, id string) (*entity.Location, error) {
	loc, ok := r.items[id]
	if !ok || loc.IsDeleted() {
		return nil, eferrors.ErrLocationNotFound
	}
	return loc.Clone(), nil
}

func (r *memLocationRepo) SelectByIDs(_ context.Context, ids []string) ([]*entity.Location, error) {
	out := make([]*entity.Location, 0, len(ids))
	for _, id := range ids {
		if loc, ok := r.items[id]; ok && !loc.IsDeleted() {
			out = append(out, loc.Clone())
		}
	}
	return out, nil
}

func (r *memLocationRepo) SelectByCode(_ context.Context, code string) (*entity.Location, error) {
	for _, id := range r.order {
		loc := r.items[id]
		if !loc.IsDeleted() && loc.LocationCode == code {
			return loc.Clone(), nil
		}
	}
	return nil, eferrors.ErrLocationNotFound
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

func (r *memLocationRepo) SelectByLegalEntityAndKind(_ context.Context, legalEntityID string, kind entity.LocationKind, activeOnly bool, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	var out []*entity.Location
	for _, id := range r.order {
		loc := r.items[id]
		if loc.IsDeleted() || loc.LegalEntityID != legalEntityID || loc.Kind != kind {
			continue
		}
		if activeOnly && !loc.IsActive {
			continue
		}
		out = append(out, loc.Clone())
	}
	return paginateMem(out, page), nil
}

func (r *memLocationRepo) SelectByAddressID(_ context.Context, addressID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	var out []*entity.Location
	for _, id := range r.order {
		loc := r.items[id]
		if !loc.IsDeleted() && loc.AddressID != nil && *loc.AddressID == addressID {
			out = append(out, loc.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memLocationRepo) SelectList(_ context.Context, filter domainrepo.LocationFilter) (domainrepo.PageResult[*entity.Location], error) {
	norm := filter.Normalize()
	var matched []*entity.Location
	q := strings.ToLower(strings.TrimSpace(norm.Query))
	for _, id := range r.order {
		loc := r.items[id]
		if !norm.IncludeDeleted && loc.IsDeleted() {
			continue
		}
		if norm.LegalEntityID != "" && loc.LegalEntityID != norm.LegalEntityID {
			continue
		}
		if norm.Kind != nil && loc.Kind != *norm.Kind {
			continue
		}
		if norm.ActiveOnly && !loc.IsActive {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(loc.LocationCode+" "+loc.Name), q) {
			continue
		}
		matched = append(matched, loc.Clone())
	}
	return paginateMem(matched, norm.Page), nil
}

func (r *memLocationRepo) SelectParentID(_ context.Context, id string) (*string, error) {
	loc, ok := r.items[id]
	if !ok || loc.IsDeleted() {
		return nil, eferrors.ErrLocationNotFound
	}
	if loc.ParentLocationID == nil {
		return nil, nil
	}
	v := *loc.ParentLocationID
	return &v, nil
}

func (r *memLocationRepo) SelectChildrenByParentID(_ context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	var out []*entity.Location
	for _, id := range r.order {
		loc := r.items[id]
		if !loc.IsDeleted() && loc.ParentLocationID != nil && *loc.ParentLocationID == parentID {
			out = append(out, loc.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memLocationRepo) CountActiveDependents(_ context.Context, id string) (int64, error) {
	return r.dependentsByID[id], nil
}

func (r *memLocationRepo) CheckCodeExists(_ context.Context, code, excludeID string) (bool, error) {
	for _, id := range r.order {
		loc := r.items[id]
		if !loc.IsDeleted() && loc.ID != excludeID && loc.LocationCode == code {
			return true, nil
		}
	}
	return false, nil
}

func (r *memLocationRepo) Insert(_ context.Context, loc *entity.Location) error {
	r.insertCalls++
	r.seed(loc)
	return nil
}

func (r *memLocationRepo) Update(_ context.Context, loc *entity.Location) error {
	if _, ok := r.items[loc.ID]; !ok {
		return eferrors.ErrLocationNotFound
	}
	r.items[loc.ID] = loc.Clone()
	return nil
}

func (r *memLocationRepo) SoftDeleteByID(_ context.Context, id string, deletedAt time.Time) error {
	loc, ok := r.items[id]
	if !ok {
		return eferrors.ErrLocationNotFound
	}
	loc.DeletedAt = &deletedAt
	loc.IsActive = false
	loc.UpdatedAt = deletedAt
	return nil
}

func newTestService(t *testing.T) (*location.Service, *memLocationRepo, *stubLegalEntityRepo, *stubAddressRepo, *stubOutboxRepo) {
	t.Helper()
	repo := newMemLocationRepo()
	leRepo := &stubLegalEntityRepo{items: make(map[string]*entity.LegalEntity)}
	addrRepo := &stubAddressRepo{items: make(map[string]*entity.Address)}
	outbox := &stubOutboxRepo{}
	svc, err := location.NewService(location.Dependencies{
		Transactor:    noopTransactor{},
		Locations:     repo,
		LegalEntities: leRepo,
		Addresses:     addrRepo,
		Outbox:        outbox,
		Clock:         fixedClock{now: time.Date(2026, 11, 15, 11, 0, 0, 0, time.UTC)},
		IDGen:         &seqIDGen{},
	})
	if err != nil {
		t.Fatalf("NewService unexpected error: %v", err)
	}
	return svc, repo, leRepo, addrRepo, outbox
}

func seedLegalEntity(t *testing.T, leRepo *stubLegalEntityRepo, id, code string, status entity.LegalEntityStatus) *entity.LegalEntity {
	t.Helper()
	le, err := entity.NewLegalEntity(id, entity.LegalEntityParams{
		EntityCode:              code,
		LegalName:               "PT " + code,
		Status:                  status,
		JurisdictionCountryCode: "ID",
		FiscalYearStartMonth:    1,
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("NewLegalEntity failed: %v", err)
	}
	leRepo.items[id] = le
	return le
}

func TestNewServiceFailClosedAndNilReceiver(t *testing.T) {
	base := location.Dependencies{
		Transactor:    noopTransactor{},
		Locations:     newMemLocationRepo(),
		LegalEntities: &stubLegalEntityRepo{},
		Addresses:     &stubAddressRepo{},
		Outbox:        &stubOutboxRepo{},
		IDGen:         &seqIDGen{},
	}

	cases := []struct {
		name  string
		mut   func(d *location.Dependencies)
		field string
	}{
		{"nil transactor", func(d *location.Dependencies) { d.Transactor = nil }, "transactor"},
		{"nil locations", func(d *location.Dependencies) { d.Locations = nil }, "locations"},
		{"nil legal_entities", func(d *location.Dependencies) { d.LegalEntities = nil }, "legal_entities"},
		{"nil addresses", func(d *location.Dependencies) { d.Addresses = nil }, "addresses"},
		{"nil outbox", func(d *location.Dependencies) { d.Outbox = nil }, "outbox"},
		{"nil id_gen", func(d *location.Dependencies) { d.IDGen = nil }, "id_gen"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := base
			tc.mut(&deps)
			_, err := location.NewService(deps)
			if err == nil || eferrors.KindOf(err) != eferrors.KindInvalidInput || eferrors.FieldOf(err) != tc.field {
				t.Fatalf("expected invalid input for %s, got %v", tc.field, err)
			}
		})
	}

	var nilSvc *location.Service
	if _, err := nilSvc.GetAll(context.Background(), domainrepo.PageRequest{}); eferrors.KindOf(err) != eferrors.KindUnavailable {
		t.Fatalf("expected KindUnavailable on nil receiver, got %v", err)
	}
}

func TestLocationCRUDHierarchyDuplicateScopeAndGetAll(t *testing.T) {
	ctx := context.Background()
	svc, repo, leRepo, addrRepo, outbox := newTestService(t)

	le1ID := "44444444-4444-4444-8444-444444444441"
	le2ID := "44444444-4444-4444-8444-444444444442"
	leSuspendedID := "44444444-4444-4444-8444-444444444443"
	seedLegalEntity(t, leRepo, le1ID, "pt-one", entity.LegalEntityStatusActive)
	seedLegalEntity(t, leRepo, le2ID, "pt-two", entity.LegalEntityStatusActive)
	seedLegalEntity(t, leRepo, leSuspendedID, "pt-susp", entity.LegalEntityStatusSuspended)

	addrID := "55555555-5555-4555-8555-555555555551"
	addr, err := entity.NewAddress(addrID, entity.AddressParams{
		AddressLine1: "Jl. Cakung Industri",
		City:         "Jakarta Timur",
		CountryCode:  "ID",
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("NewAddress error: %v", err)
	}
	addrRepo.items[addrID] = addr

	// Pemilik badan hukum nonaktif ditolak sebelum Insert.
	_, err = svc.Create(ctx, location.CreateInput{
		LegalEntityID: leSuspendedID,
		LocationCode:  "HQ-SUSP",
		Name:          "Kantor Beku",
		Kind:          entity.LocationKindHeadquarters,
		Timezone:      "Asia/Jakarta",
	})
	if !errors.Is(err, eferrors.ErrLegalEntityInactive) {
		t.Fatalf("expected ErrLegalEntityInactive, got %v", err)
	}

	// Buat HQ dan Warehouse.
	hq, err := svc.Create(ctx, location.CreateInput{
		LegalEntityID: le1ID,
		AddressID:     &addrID,
		LocationCode:  "HQ-JKT",
		Name:          "Kantor Pusat Jakarta",
		Kind:          entity.LocationKindHeadquarters,
		Timezone:      "Asia/Jakarta",
	})
	if err != nil {
		t.Fatalf("Create HQ failed: %v", err)
	}

	wh, err := svc.Create(ctx, location.CreateInput{
		LegalEntityID:    le1ID,
		ParentLocationID: &hq.ID,
		AddressID:        &addrID,
		LocationCode:     "WH-CKG",
		Name:             "Gudang Utama Cakung",
		Kind:             entity.LocationKindWarehouse,
		Timezone:         "Asia/Jakarta",
	})
	if err != nil {
		t.Fatalf("Create Warehouse failed: %v", err)
	}

	callsBefore := repo.insertCalls

	// Duplikasi location_code (normalisasi huruf kecil) ditolak sebelum Insert.
	_, err = svc.Create(ctx, location.CreateInput{
		LegalEntityID: le2ID,
		LocationCode:  " wh-ckg ",
		Name:          "Gudang Duplikat",
		Kind:          entity.LocationKindWarehouse,
		Timezone:      "Asia/Jakarta",
	})
	if !errors.Is(err, eferrors.ErrLocationCodeExists) {
		t.Fatalf("expected ErrLocationCodeExists, got %v", err)
	}

	// Lokasi induk beda badan hukum ditolak sebelum Insert.
	_, err = svc.Create(ctx, location.CreateInput{
		LegalEntityID:    le2ID,
		ParentLocationID: &hq.ID,
		LocationCode:     "ST-SUB",
		Name:             "Toko Lintas Entitas",
		Kind:             entity.LocationKindRetailStore,
		Timezone:         "Asia/Jakarta",
	})
	if !errors.Is(err, eferrors.ErrLocationParentMismatch) {
		t.Fatalf("expected ErrLocationParentMismatch, got %v", err)
	}
	if repo.insertCalls != callsBefore {
		t.Fatalf("expected Insert not called on invalid creates, before=%d after=%d", callsBefore, repo.insertCalls)
	}

	// Cycle check pada Update (hq menunjuk anak wh).
	_, err = svc.Update(ctx, location.UpdateInput{
		ID:               hq.ID,
		ParentLocationID: &wh.ID,
		AddressID:        &addrID,
		Name:             hq.Name,
		Kind:             hq.Kind,
		Timezone:         hq.Timezone,
	})
	if !errors.Is(err, eferrors.ErrLocationCycleDetected) {
		t.Fatalf("expected ErrLocationCycleDetected, got %v", err)
	}

	// Valid Update pada wh.
	whUpdated, err := svc.Update(ctx, location.UpdateInput{
		ID:               wh.ID,
		ParentLocationID: &hq.ID,
		AddressID:        &addrID,
		Name:             "Gudang Fulfillment Cakung",
		Kind:             entity.LocationKindWarehouse,
		Timezone:         "Asia/Jakarta",
	})
	if err != nil || whUpdated.Name != "Gudang Fulfillment Cakung" {
		t.Fatalf("Update failed: got=%+v err=%v", whUpdated, err)
	}

	// Scope checks: facility ok untuk HQ & WH; warehouse scope hanya untuk WH.
	if dec, err := svc.ValidateFacilityScope(ctx, hq.ID); err != nil || !dec.Eligible {
		t.Fatalf("ValidateFacilityScope HQ failed: dec=%+v err=%v", dec, err)
	}
	if dec, err := svc.ValidateWarehouseScope(ctx, wh.ID); err != nil || !dec.Eligible {
		t.Fatalf("ValidateWarehouseScope WH failed: dec=%+v err=%v", dec, err)
	}
	if _, err := svc.ValidateWarehouseScope(ctx, hq.ID); eferrors.KindOf(err) != eferrors.KindInvalidInput && eferrors.KindOf(err) != eferrors.KindForbidden && eferrors.KindOf(err) != eferrors.KindConflict {
		t.Fatalf("expected scope rejection for non-warehouse HQ, got %v", err)
	}

	// Query helpers.
	if got, err := svc.GetByCode(ctx, "wh-ckg"); err != nil || got.ID != wh.ID {
		t.Fatalf("GetByCode failed: got=%+v err=%v", got, err)
	}
	if list, err := svc.ListByLegalEntityAndKind(ctx, le1ID, entity.LocationKindWarehouse, true, domainrepo.PageRequest{}); err != nil || list.Total != 1 || len(list.Items) != 1 {
		t.Fatalf("ListByLegalEntityAndKind failed: page=%+v err=%v", list, err)
	}
	if byAddr, err := svc.ListByAddressID(ctx, addrID, domainrepo.PageRequest{}); err != nil || byAddr.Total != 2 || len(byAddr.Items) != 2 {
		t.Fatalf("ListByAddressID failed: page=%+v err=%v", byAddr, err)
	}
	if addrP1, err := svc.ListByAddressID(ctx, addrID, domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || len(addrP1.Items) != 1 || addrP1.Total != 2 || !addrP1.HasMore {
		t.Fatalf("ListByAddressID page 1 failed: page=%+v err=%v", addrP1, err)
	}
	if children, err := svc.ListChildren(ctx, hq.ID, domainrepo.PageRequest{}); err != nil || children.Total != 1 || len(children.Items) != 1 {
		t.Fatalf("ListChildren failed: page=%+v err=%v", children, err)
	}

	// Delete blocked jika ada tanggungan aktif.
	repo.dependentsByID[hq.ID] = 1
	if err := svc.Delete(ctx, hq.ID); err == nil || eferrors.KindOf(err) != eferrors.KindConflict {
		t.Fatalf("expected conflict error when location has dependents, got %v", err)
	}

	// Deactivate, Re-activate gagal bila legal entity disuspend.
	if _, err := svc.Deactivate(ctx, wh.ID); err != nil {
		t.Fatalf("Deactivate failed: %v", err)
	}
	if err := leRepo.items[le1ID].Suspend(time.Now().UTC()); err != nil {
		t.Fatalf("Suspend owner failed: %v", err)
	}
	if _, err := svc.Activate(ctx, wh.ID); !errors.Is(err, eferrors.ErrLegalEntityInactive) {
		t.Fatalf("expected ErrLegalEntityInactive on reactivate under suspended entity, got %v", err)
	}
	if err := leRepo.items[le1ID].Activate(time.Now().UTC()); err != nil {
		t.Fatalf("Reactivate owner failed: %v", err)
	}
	if _, err := svc.Activate(ctx, wh.ID); err != nil {
		t.Fatalf("Activate failed: %v", err)
	}

	// Soft-delete lokasi bebas tanggungan.
	repo.dependentsByID[wh.ID] = 0
	if err := svc.Delete(ctx, wh.ID); err != nil {
		t.Fatalf("Delete WH failed: %v", err)
	}
	if _, err := svc.GetByID(ctx, wh.ID); !errors.Is(err, eferrors.ErrLocationNotFound) {
		t.Fatalf("expected deleted location not found, got %v", err)
	}
	if len(outbox.events) < 5 {
		t.Fatalf("expected outbox events recorded, got %d", len(outbox.events))
	}

	// GetAll paginasi multi-halaman (>100 baris).
	for i := 0; i < 110; i++ {
		id := fmt.Sprintf("66666666-6666-4666-8666-%012d", i+1)
		loc, err := entity.NewLocation(id, entity.LocationParams{
			LegalEntityID: le1ID,
			LocationCode:  fmt.Sprintf("st-bulk-%03d", i+1),
			Name:          fmt.Sprintf("Toko Bulk %03d", i+1),
			Kind:          entity.LocationKindRetailStore,
			Timezone:      "Asia/Jakarta",
		}, time.Now().UTC())
		if err != nil {
			t.Fatalf("NewLocation seed error: %v", err)
		}
		repo.seed(loc)
	}
	p1, err := svc.GetAll(ctx, domainrepo.PageRequest{Limit: 50, Offset: 0})
	if err != nil || p1.Total != 111 || len(p1.Items) != 50 || !p1.HasMore {
		t.Fatalf("GetAll page 1 failed: page=%+v err=%v", p1, err)
	}
	p3, err := svc.GetAll(ctx, domainrepo.PageRequest{Limit: 50, Offset: 100})
	if err != nil || p3.Total != 111 || len(p3.Items) != 11 || p3.HasMore {
		t.Fatalf("GetAll page 3 failed: page=%+v err=%v", p3, err)
	}
}
