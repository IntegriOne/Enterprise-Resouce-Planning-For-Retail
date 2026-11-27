package legalentity_test

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
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/legalentity"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type seqIDGen struct {
	seq int
}

func (g *seqIDGen) NewUUID() (string, error) {
	g.seq++
	return fmt.Sprintf("11111111-1111-4111-8111-%012d", g.seq), nil
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

type memLegalEntityRepo struct {
	order            []string
	items            map[string]*entity.LegalEntity
	dependentsByID   map[string]int64
	insertCalls      int
	clearExceptCalls int
}

func newMemLegalEntityRepo() *memLegalEntityRepo {
	return &memLegalEntityRepo{
		items:          make(map[string]*entity.LegalEntity),
		dependentsByID: make(map[string]int64),
	}
}

func (r *memLegalEntityRepo) seed(le *entity.LegalEntity) {
	r.order = append(r.order, le.ID)
	r.items[le.ID] = le.Clone()
}

func (r *memLegalEntityRepo) SelectByID(_ context.Context, id string) (*entity.LegalEntity, error) {
	le, ok := r.items[id]
	if !ok || le.IsDeleted() {
		return nil, eferrors.ErrLegalEntityNotFound
	}
	return le.Clone(), nil
}

func (r *memLegalEntityRepo) SelectByIDs(_ context.Context, ids []string) ([]*entity.LegalEntity, error) {
	out := make([]*entity.LegalEntity, 0, len(ids))
	for _, id := range ids {
		if le, ok := r.items[id]; ok && !le.IsDeleted() {
			out = append(out, le.Clone())
		}
	}
	return out, nil
}

func (r *memLegalEntityRepo) SelectByCode(_ context.Context, code string) (*entity.LegalEntity, error) {
	for _, id := range r.order {
		le := r.items[id]
		if !le.IsDeleted() && le.EntityCode == code {
			return le.Clone(), nil
		}
	}
	return nil, eferrors.ErrLegalEntityNotFound
}

func (r *memLegalEntityRepo) SelectByTaxIDNumber(_ context.Context, taxID string) (*entity.LegalEntity, error) {
	for _, id := range r.order {
		le := r.items[id]
		if !le.IsDeleted() && le.TaxIDNumber != nil && *le.TaxIDNumber == taxID {
			return le.Clone(), nil
		}
	}
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

func (r *memLegalEntityRepo) SelectByNIB(_ context.Context, nib string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	var out []*entity.LegalEntity
	for _, id := range r.order {
		le := r.items[id]
		if !le.IsDeleted() && le.NIB != nil && *le.NIB == nib {
			out = append(out, le.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memLegalEntityRepo) SelectPrimaryOperating(context.Context) (*entity.LegalEntity, error) {
	for _, id := range r.order {
		le := r.items[id]
		if !le.IsDeleted() && le.IsPrimaryOperatingEntity {
			return le.Clone(), nil
		}
	}
	return nil, eferrors.ErrLegalEntityNotFound
}

func (r *memLegalEntityRepo) SelectList(_ context.Context, filter domainrepo.LegalEntityFilter) (domainrepo.PageResult[*entity.LegalEntity], error) {
	norm := filter.Normalize()
	var matched []*entity.LegalEntity
	q := strings.ToLower(strings.TrimSpace(norm.Query))
	for _, id := range r.order {
		le := r.items[id]
		if !norm.IncludeDeleted && le.IsDeleted() {
			continue
		}
		if norm.Status != nil && le.Status != *norm.Status {
			continue
		}
		if norm.PrimaryOnly != nil && *norm.PrimaryOnly && !le.IsPrimaryOperatingEntity {
			continue
		}
		if norm.ParentLegalEntityID != nil && (le.ParentLegalEntityID == nil || *le.ParentLegalEntityID != *norm.ParentLegalEntityID) {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(le.EntityCode+" "+le.LegalName), q) {
			continue
		}
		matched = append(matched, le.Clone())
	}
	return paginateMem(matched, norm.Page), nil
}

func (r *memLegalEntityRepo) SelectParentID(_ context.Context, id string) (*string, error) {
	le, ok := r.items[id]
	if !ok || le.IsDeleted() {
		return nil, eferrors.ErrLegalEntityNotFound
	}
	if le.ParentLegalEntityID == nil {
		return nil, nil
	}
	v := *le.ParentLegalEntityID
	return &v, nil
}

func (r *memLegalEntityRepo) SelectChildrenByParentID(_ context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	var out []*entity.LegalEntity
	for _, id := range r.order {
		le := r.items[id]
		if !le.IsDeleted() && le.ParentLegalEntityID != nil && *le.ParentLegalEntityID == parentID {
			out = append(out, le.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memLegalEntityRepo) CountActiveDependents(_ context.Context, id string) (int64, error) {
	return r.dependentsByID[id], nil
}

func (r *memLegalEntityRepo) CheckCodeExists(_ context.Context, code, excludeID string) (bool, error) {
	for _, id := range r.order {
		le := r.items[id]
		if !le.IsDeleted() && le.ID != excludeID && le.EntityCode == code {
			return true, nil
		}
	}
	return false, nil
}

func (r *memLegalEntityRepo) CheckTaxIDExists(_ context.Context, taxID, excludeID string) (bool, error) {
	for _, id := range r.order {
		le := r.items[id]
		if !le.IsDeleted() && le.ID != excludeID && le.TaxIDNumber != nil && *le.TaxIDNumber == taxID {
			return true, nil
		}
	}
	return false, nil
}

func (r *memLegalEntityRepo) Insert(_ context.Context, le *entity.LegalEntity) error {
	r.insertCalls++
	r.seed(le)
	return nil
}

func (r *memLegalEntityRepo) Update(_ context.Context, le *entity.LegalEntity) error {
	if _, ok := r.items[le.ID]; !ok {
		return eferrors.ErrLegalEntityNotFound
	}
	r.items[le.ID] = le.Clone()
	return nil
}

func (r *memLegalEntityRepo) UpdateClearPrimaryOperatingExcept(_ context.Context, keepID string, updatedAt time.Time) error {
	r.clearExceptCalls++
	for _, id := range r.order {
		le := r.items[id]
		if le.ID != keepID && le.IsPrimaryOperatingEntity {
			le.IsPrimaryOperatingEntity = false
			le.UpdatedAt = updatedAt
		}
	}
	return nil
}

func (r *memLegalEntityRepo) SoftDeleteByID(_ context.Context, id string, deletedAt time.Time) error {
	le, ok := r.items[id]
	if !ok {
		return eferrors.ErrLegalEntityNotFound
	}
	le.DeletedAt = &deletedAt
	le.UpdatedAt = deletedAt
	return nil
}

func strPtr(v string) *string { return &v }

func newTestService(t *testing.T) (*legalentity.Service, *memLegalEntityRepo, *stubAddressRepo, *stubOutboxRepo) {
	t.Helper()
	repo := newMemLegalEntityRepo()
	addrRepo := &stubAddressRepo{items: make(map[string]*entity.Address)}
	outbox := &stubOutboxRepo{}
	svc, err := legalentity.NewService(legalentity.Dependencies{
		Transactor:    noopTransactor{},
		LegalEntities: repo,
		Addresses:     addrRepo,
		Outbox:        outbox,
		Clock:         fixedClock{now: time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)},
		IDGen:         &seqIDGen{},
	})
	if err != nil {
		t.Fatalf("NewService unexpected error: %v", err)
	}
	return svc, repo, addrRepo, outbox
}

func TestNewServiceFailClosedAndNilReceiver(t *testing.T) {
	base := legalentity.Dependencies{
		Transactor:    noopTransactor{},
		LegalEntities: newMemLegalEntityRepo(),
		Addresses:     &stubAddressRepo{},
		Outbox:        &stubOutboxRepo{},
		IDGen:         &seqIDGen{},
	}

	cases := []struct {
		name  string
		mut   func(d *legalentity.Dependencies)
		field string
	}{
		{"nil transactor", func(d *legalentity.Dependencies) { d.Transactor = nil }, "transactor"},
		{"nil legal entity repo", func(d *legalentity.Dependencies) { d.LegalEntities = nil }, "legal_entities"},
		{"nil address repo", func(d *legalentity.Dependencies) { d.Addresses = nil }, "addresses"},
		{"nil outbox repo", func(d *legalentity.Dependencies) { d.Outbox = nil }, "outbox"},
		{"nil id gen", func(d *legalentity.Dependencies) { d.IDGen = nil }, "id_gen"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deps := base
			tc.mut(&deps)
			_, err := legalentity.NewService(deps)
			if err == nil || eferrors.KindOf(err) != eferrors.KindInvalidInput || eferrors.FieldOf(err) != tc.field {
				t.Fatalf("expected invalid input for %s, got %v", tc.field, err)
			}
		})
	}

	var nilSvc *legalentity.Service
	if _, err := nilSvc.GetAll(context.Background(), domainrepo.PageRequest{}); eferrors.KindOf(err) != eferrors.KindUnavailable {
		t.Fatalf("expected KindUnavailable on nil receiver, got %v", err)
	}
}

func TestCreateDuplicatePreventionHierarchyOutboxAndScope(t *testing.T) {
	ctx := context.Background()
	svc, repo, addrRepo, outbox := newTestService(t)

	addrID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"
	addr, err := entity.NewAddress(addrID, entity.AddressParams{
		AddressLine1: "Jl. Sudirman No. 1",
		City:         "Jakarta Selatan",
		CountryCode:  "ID",
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("NewAddress error: %v", err)
	}
	addrRepo.items[addrID] = addr

	primary, err := svc.Create(ctx, legalentity.CreateInput{
		EntityCode:               "PT-HOLD",
		LegalName:                "PT IntegriOne Ritel Nusantara",
		IsPrimaryOperatingEntity: true,
		Status:                   entity.LegalEntityStatusActive,
		JurisdictionCountryCode:  "ID",
		NIB:                      strPtr("1234567890123"),
		TaxIDNumber:              strPtr("01.234.567.8-901.000"),
		IsPKP:                    true,
		RegisteredAddressID:      &addrID,
		FiscalYearStartMonth:     1,
	})
	if err != nil {
		t.Fatalf("Create primary failed: %v", err)
	}
	if primary.EntityCode != "pt-hold" {
		t.Fatalf("expected normalized code pt-hold, got %q", primary.EntityCode)
	}
	if len(outbox.events) != 1 || outbox.events[0].EventType != entity.EventTypeLegalEntityCreated {
		t.Fatalf("expected created outbox event, got %+v", outbox.events)
	}

	callsBefore := repo.insertCalls

	// Duplikasi kode badan hukum (case-insensitive) ditolak sebelum Insert.
	_, err = svc.Create(ctx, legalentity.CreateInput{
		EntityCode:              " pt-hold ",
		LegalName:               "Duplikat PT",
		Status:                  entity.LegalEntityStatusActive,
		JurisdictionCountryCode: "ID",
		FiscalYearStartMonth:    1,
	})
	if !errors.Is(err, eferrors.ErrLegalEntityCodeExists) {
		t.Fatalf("expected ErrLegalEntityCodeExists, got %v", err)
	}
	// Duplikasi NPWP ditolak sebelum Insert.
	_, err = svc.Create(ctx, legalentity.CreateInput{
		EntityCode:              "pt-sub-1",
		LegalName:               "Sub PT",
		Status:                  entity.LegalEntityStatusActive,
		JurisdictionCountryCode: "ID",
		TaxIDNumber:             strPtr("01.234.567.8-901.000"),
		FiscalYearStartMonth:    1,
	})
	if !errors.Is(err, eferrors.ErrLegalEntityTaxIDExists) {
		t.Fatalf("expected ErrLegalEntityTaxIDExists, got %v", err)
	}
	// Entitas utama ganda saat Create ditolak sebelum Insert.
	_, err = svc.Create(ctx, legalentity.CreateInput{
		EntityCode:               "pt-sub-2",
		LegalName:                "Sub PT 2",
		IsPrimaryOperatingEntity: true,
		Status:                   entity.LegalEntityStatusActive,
		JurisdictionCountryCode:  "ID",
		FiscalYearStartMonth:     1,
	})
	if !errors.Is(err, eferrors.ErrPrimaryLegalEntityExists) {
		t.Fatalf("expected ErrPrimaryLegalEntityExists, got %v", err)
	}
	if repo.insertCalls != callsBefore {
		t.Fatalf("expected Insert not called on duplicate checks; before=%d after=%d", callsBefore, repo.insertCalls)
	}

	// Buat entitas anak sah di bawah induk.
	child, err := svc.Create(ctx, legalentity.CreateInput{
		ParentLegalEntityID:     &primary.ID,
		EntityCode:              "pt-anak",
		LegalName:               "PT IntegriOne Anak",
		Status:                  entity.LegalEntityStatusPending,
		JurisdictionCountryCode: "ID",
		NIB:                     strPtr("1234567890123"),
		TaxIDNumber:             strPtr("02.345.678.9-012.000"),
		FiscalYearStartMonth:    1,
	})
	if err != nil {
		t.Fatalf("Create child failed: %v", err)
	}

	// Pending tidak lolos scope evaluation sampai diaktifkan.
	if _, err := svc.ValidateScopeTarget(ctx, child.ID); !errors.Is(err, eferrors.ErrLegalEntityInactive) {
		t.Fatalf("expected ErrLegalEntityInactive on pending scope check, got %v", err)
	}

	// Update induk primary menunjuk child -> harus terdeteksi siklus hierarki.
	_, err = svc.Update(ctx, legalentity.UpdateInput{
		ID:                      primary.ID,
		ParentLegalEntityID:     &child.ID,
		LegalName:               primary.LegalName,
		JurisdictionCountryCode: "ID",
		TaxIDNumber:             primary.TaxIDNumber,
		FiscalYearStartMonth:    1,
	})
	if !errors.Is(err, eferrors.ErrLegalEntityCycleDetected) {
		t.Fatalf("expected ErrLegalEntityCycleDetected, got %v", err)
	}

	// Lookup helpers & ListChildren.
	if got, err := svc.GetByCode(ctx, "PT-HOLD"); err != nil || got.ID != primary.ID {
		t.Fatalf("GetByCode failed: got=%+v err=%v", got, err)
	}
	if got, err := svc.GetByTaxIDNumber(ctx, "01.234.567.8-901.000"); err != nil || got.ID != primary.ID {
		t.Fatalf("GetByTaxIDNumber failed: got=%+v err=%v", got, err)
	}
	if nibList, err := svc.ListByNIB(ctx, "1234567890123", domainrepo.PageRequest{}); err != nil || nibList.Total != 2 || len(nibList.Items) != 2 || nibList.HasMore {
		t.Fatalf("ListByNIB failed: page=%+v err=%v", nibList, err)
	}
	if nibP1, err := svc.ListByNIB(ctx, "1234567890123", domainrepo.PageRequest{Limit: 1, Offset: 0}); err != nil || len(nibP1.Items) != 1 || nibP1.Total != 2 || !nibP1.HasMore {
		t.Fatalf("ListByNIB page 1 failed: page=%+v err=%v", nibP1, err)
	} else if nibP2, err := svc.ListByNIB(ctx, "1234567890123", domainrepo.PageRequest{Limit: 1, Offset: 1}); err != nil || len(nibP2.Items) != 1 || nibP2.Total != 2 || nibP2.HasMore || nibP2.Items[0].ID == nibP1.Items[0].ID {
		t.Fatalf("ListByNIB page 2 failed: page=%+v p1=%+v err=%v", nibP2, nibP1, err)
	}
	if children, err := svc.ListChildren(ctx, primary.ID, domainrepo.PageRequest{}); err != nil || children.Total != 1 || len(children.Items) != 1 || children.Items[0].ID != child.ID {
		t.Fatalf("ListChildren failed: page=%+v err=%v", children, err)
	}
}

func TestLifecyclePrimarySwitchDeleteAndGetAllPagination(t *testing.T) {
	ctx := context.Background()
	svc, repo, _, outbox := newTestService(t)

	primary, err := svc.Create(ctx, legalentity.CreateInput{
		EntityCode:               "pt-utama",
		LegalName:                "PT Utama",
		IsPrimaryOperatingEntity: true,
		Status:                   entity.LegalEntityStatusActive,
		JurisdictionCountryCode:  "ID",
		FiscalYearStartMonth:     1,
	})
	if err != nil {
		t.Fatalf("Create primary error: %v", err)
	}
	secondary, err := svc.Create(ctx, legalentity.CreateInput{
		EntityCode:              "pt-cabang",
		LegalName:               "PT Cabang",
		Status:                  entity.LegalEntityStatusPending,
		JurisdictionCountryCode: "ID",
		FiscalYearStartMonth:    1,
	})
	if err != nil {
		t.Fatalf("Create secondary error: %v", err)
	}

	// Aktifkan secondary -> scope valid.
	secondary, err = svc.Activate(ctx, secondary.ID)
	if err != nil || secondary.Status != entity.LegalEntityStatusActive {
		t.Fatalf("Activate secondary failed: got=%+v err=%v", secondary, err)
	}
	decision, err := svc.ValidateScopeTarget(ctx, secondary.ID)
	if err != nil || !decision.Eligible || decision.LegalEntityID != secondary.ID {
		t.Fatalf("ValidateScopeTarget failed: decision=%+v err=%v", decision, err)
	}

	// Pindahkan entitas utama operasional ke secondary.
	secondary, err = svc.SetPrimaryOperating(ctx, secondary.ID)
	if err != nil || !secondary.IsPrimaryOperatingEntity {
		t.Fatalf("SetPrimaryOperating failed: got=%+v err=%v", secondary, err)
	}
	if repo.clearExceptCalls != 1 {
		t.Fatalf("expected UpdateClearPrimaryOperatingExcept called once, got %d", repo.clearExceptCalls)
	}
	if curPrimary, err := svc.GetPrimaryOperating(ctx); err != nil || curPrimary.ID != secondary.ID {
		t.Fatalf("GetPrimaryOperating mismatch: got=%+v err=%v", curPrimary, err)
	}

	// Hapus entitas primary (lama) gagal bila masih punya tanggungan aktif.
	repo.dependentsByID[primary.ID] = 2
	if err := svc.Delete(ctx, primary.ID); !errors.Is(err, eferrors.ErrLegalEntityHasDependents) {
		t.Fatalf("expected ErrLegalEntityHasDependents, got %v", err)
	}

	// Bebaskan tanggungan lalu suspend dan soft-delete.
	repo.dependentsByID[primary.ID] = 0
	if _, err := svc.Suspend(ctx, primary.ID); err != nil {
		t.Fatalf("Suspend old primary failed: %v", err)
	}
	if err := svc.Delete(ctx, primary.ID); err != nil {
		t.Fatalf("Delete old primary failed: %v", err)
	}
	if _, err := svc.GetByID(ctx, primary.ID); !errors.Is(err, eferrors.ErrLegalEntityNotFound) {
		t.Fatalf("expected deleted entity not found, got %v", err)
	}
	if len(outbox.events) < 5 {
		t.Fatalf("expected outbox lifecycle events recorded, got %d", len(outbox.events))
	}

	// Uji GetAll multi-halaman (>100 entitas).
	for i := 0; i < 120; i++ {
		id := fmt.Sprintf("22222222-2222-4222-8222-%012d", i+1)
		le, err := entity.NewLegalEntity(id, entity.LegalEntityParams{
			EntityCode:              fmt.Sprintf("le-bulk-%03d", i+1),
			LegalName:               fmt.Sprintf("PT Bulk %03d", i+1),
			Status:                  entity.LegalEntityStatusActive,
			JurisdictionCountryCode: "ID",
			FiscalYearStartMonth:    1,
		}, time.Now().UTC())
		if err != nil {
			t.Fatalf("NewLegalEntity seed error: %v", err)
		}
		repo.seed(le)
	}
	p1, err := svc.GetAll(ctx, domainrepo.PageRequest{Limit: 50, Offset: 0})
	if err != nil || p1.Total != 121 || len(p1.Items) != 50 || !p1.HasMore {
		t.Fatalf("GetAll page 1 failed: page=%+v err=%v", p1, err)
	}
	p2, err := svc.GetAll(ctx, domainrepo.PageRequest{Limit: 50, Offset: 50})
	if err != nil || p2.Total != 121 || len(p2.Items) != 50 || !p2.HasMore {
		t.Fatalf("GetAll page 2 failed: page=%+v err=%v", p2, err)
	}
	p3, err := svc.GetAll(ctx, domainrepo.PageRequest{Limit: 50, Offset: 100})
	if err != nil || p3.Total != 121 || len(p3.Items) != 21 || p3.HasMore {
		t.Fatalf("GetAll page 3 failed: page=%+v err=%v", p3, err)
	}
}
