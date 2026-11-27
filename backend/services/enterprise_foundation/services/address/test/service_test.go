package address_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/address"
)

type seqClock struct {
	now time.Time
}

func (c *seqClock) Now() time.Time { return c.now }

type seqIDGen struct {
	seq int
}

func (g *seqIDGen) NewUUID() (string, error) {
	g.seq++
	return fmt.Sprintf("10000000-0000-4000-8000-%012d", g.seq), nil
}

type fakeAddressRepo struct {
	rows        map[string]*entity.Address
	order       []string
	referenced  map[string]bool
	insertCalls int
}

func newFakeAddressRepo() *fakeAddressRepo {
	return &fakeAddressRepo{
		rows:       make(map[string]*entity.Address),
		referenced: make(map[string]bool),
	}
}

func (r *fakeAddressRepo) SelectByID(_ context.Context, id string) (*entity.Address, error) {
	row, ok := r.rows[entity.NormalizeUUID(id)]
	if !ok {
		return nil, eferrors.ErrAddressNotFound
	}
	return row.Clone(), nil
}

func (r *fakeAddressRepo) SelectByIDs(_ context.Context, ids []string) ([]*entity.Address, error) {
	out := make([]*entity.Address, 0, len(ids))
	for _, id := range ids {
		if row, ok := r.rows[entity.NormalizeUUID(id)]; ok {
			out = append(out, row.Clone())
		}
	}
	return out, nil
}

func (r *fakeAddressRepo) SelectList(_ context.Context, filter domainrepo.AddressFilter) (domainrepo.PageResult[*entity.Address], error) {
	f := filter.Normalize()
	filtered := make([]*entity.Address, 0, len(r.order))
	for _, id := range r.order {
		row := r.rows[id]
		if f.CountryCode != "" && !strings.EqualFold(row.CountryCode, f.CountryCode) {
			continue
		}
		if f.City != "" && !strings.Contains(strings.ToLower(row.City), strings.ToLower(f.City)) {
			continue
		}
		if f.Province != "" && (row.Province == nil || !strings.Contains(strings.ToLower(*row.Province), strings.ToLower(f.Province))) {
			continue
		}
		if f.PostalCode != "" && (row.PostalCode == nil || *row.PostalCode != f.PostalCode) {
			continue
		}
		if f.Query != "" {
			q := strings.ToLower(f.Query)
			label := ""
			if row.Label != nil {
				label = strings.ToLower(*row.Label)
			}
			if !strings.Contains(strings.ToLower(row.AddressLine1), q) && !strings.Contains(label, q) {
				continue
			}
		}
		filtered = append(filtered, row.Clone())
	}
	total := int64(len(filtered))
	start := min(f.Page.Offset, len(filtered))
	end := min(start+f.Page.Limit, len(filtered))
	return domainrepo.NewPageResult(filtered[start:end], total, f.Page), nil
}

func (r *fakeAddressRepo) CheckReferenced(_ context.Context, id string) (bool, error) {
	return r.referenced[entity.NormalizeUUID(id)], nil
}

func (r *fakeAddressRepo) Insert(_ context.Context, addr *entity.Address) error {
	r.insertCalls++
	r.rows[addr.ID] = addr.Clone()
	r.order = append(r.order, addr.ID)
	return nil
}

func (r *fakeAddressRepo) Update(_ context.Context, addr *entity.Address) error {
	if _, ok := r.rows[addr.ID]; !ok {
		return eferrors.ErrAddressNotFound
	}
	r.rows[addr.ID] = addr.Clone()
	return nil
}

func (r *fakeAddressRepo) DeleteByID(_ context.Context, id string) error {
	norm := entity.NormalizeUUID(id)
	if _, ok := r.rows[norm]; !ok {
		return eferrors.ErrAddressNotFound
	}
	delete(r.rows, norm)
	r.order = slices.DeleteFunc(r.order, func(cur string) bool { return cur == norm })
	return nil
}

func strPtr(v string) *string { return &v }

func TestAddressService_FailClosed(t *testing.T) {
	t.Parallel()

	if _, err := address.NewService(address.Dependencies{IDGen: &seqIDGen{}}); !errors.Is(err, eferrors.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for nil repo, got %v", err)
	}
	if _, err := address.NewService(address.Dependencies{Addresses: newFakeAddressRepo()}); !errors.Is(err, eferrors.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for nil idGen, got %v", err)
	}
	var nilSvc *address.Service
	if _, err := nilSvc.GetByID(context.Background(), "10000000-0000-4000-8000-000000000001"); !errors.Is(err, eferrors.ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable on nil receiver, got %v", err)
	}
}

func TestAddressService_CRUD_DuplicateFindAndGetAll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := newFakeAddressRepo()
	svc, err := address.NewService(address.Dependencies{
		Addresses: repo,
		Clock:     &seqClock{now: time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)},
		IDGen:     &seqIDGen{},
	})
	if err != nil {
		t.Fatalf("NewService error = %v", err)
	}

	created, err := svc.Create(ctx, address.CreateInput{
		Label:        strPtr("Kantor Utama"),
		AddressLine1: "Jl. Asia Afrika No. 100",
		City:         "Bandung",
		Province:     strPtr("Jawa Barat"),
		PostalCode:   strPtr("40111"),
		CountryCode:  "id",
	})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}

	// Sebelum create wajib find agar alamat identik tidak dibuat dua kali.
	callsBefore := repo.insertCalls
	if _, err := svc.Create(ctx, address.CreateInput{
		Label:        strPtr("kantor utama"),
		AddressLine1: "  jl. asia afrika no. 100 ",
		City:         "BANDUNG",
		Province:     strPtr("jawa barat"),
		PostalCode:   strPtr("40111"),
		CountryCode:  "ID",
	}); !errors.Is(err, eferrors.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists on duplicate address, got %v", err)
	}
	if repo.insertCalls != callsBefore {
		t.Fatalf("expected Insert not called on duplicate address, calls before=%d after=%d", callsBefore, repo.insertCalls)
	}

	// FindOrCreate mengembalikan entitas lama saat identik, atau membuat baru bila berbeda.
	reused, wasCreated, err := svc.FindOrCreate(ctx, address.CreateInput{
		Label:        strPtr("Kantor Utama"),
		AddressLine1: "Jl. Asia Afrika No. 100",
		City:         "Bandung",
		Province:     strPtr("Jawa Barat"),
		PostalCode:   strPtr("40111"),
		CountryCode:  "ID",
	})
	if err != nil || wasCreated || reused == nil || reused.ID != created.ID {
		t.Fatalf("expected FindOrCreate reuse existing %s, got=%+v created=%v err=%v", created.ID, reused, wasCreated, err)
	}

	second, wasCreated2, err := svc.FindOrCreate(ctx, address.CreateInput{
		Label:        strPtr("Gudang Cakung"),
		AddressLine1: "Jl. Industri Raya No. 8",
		City:         "Jakarta Timur",
		CountryCode:  "ID",
	})
	if err != nil || !wasCreated2 || second.ID == created.ID {
		t.Fatalf("expected FindOrCreate create second address, got created=%v err=%v", wasCreated2, err)
	}

	// GetByID, GetByIDs, List, dan GetAll.
	got, err := svc.GetByID(ctx, created.ID)
	if err != nil || got.CountryCode != "ID" {
		t.Fatalf("GetByID failed: got=%+v err=%v", got, err)
	}
	batch, err := svc.GetByIDs(ctx, []string{created.ID, second.ID, created.ID})
	if err != nil || len(batch) != 2 {
		t.Fatalf("GetByIDs failed: len=%d err=%v", len(batch), err)
	}
	page, err := svc.List(ctx, domainrepo.AddressFilter{City: "Bandung"})
	if err != nil || page.Total != 1 {
		t.Fatalf("List failed: page=%+v err=%v", page, err)
	}
	all, err := svc.GetAll(ctx, domainrepo.PageRequest{})
	if err != nil || all.Total != 2 || len(all.Items) != 2 || all.HasMore {
		t.Fatalf("GetAll failed: page=%+v err=%v", all, err)
	}
	allP1, err := svc.GetAll(ctx, domainrepo.PageRequest{Limit: 1, Offset: 0})
	if err != nil || allP1.Total != 2 || len(allP1.Items) != 1 || !allP1.HasMore {
		t.Fatalf("GetAll page 1 failed: page=%+v err=%v", allP1, err)
	}
	allP2, err := svc.GetAll(ctx, domainrepo.PageRequest{Limit: 1, Offset: 1})
	if err != nil || allP2.Total != 2 || len(allP2.Items) != 1 || allP2.HasMore || allP2.Items[0].ID == allP1.Items[0].ID {
		t.Fatalf("GetAll page 2 failed: page=%+v p1=%+v err=%v", allP2, allP1, err)
	}

	// Update gagal jika berubah menjadi identik dengan alamat lain.
	if _, err := svc.Update(ctx, address.UpdateInput{
		ID:           second.ID,
		Label:        strPtr("Kantor Utama"),
		AddressLine1: "Jl. Asia Afrika No. 100",
		City:         "Bandung",
		Province:     strPtr("Jawa Barat"),
		PostalCode:   strPtr("40111"),
		CountryCode:  "ID",
	}); !errors.Is(err, eferrors.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists on colliding update, got %v", err)
	}

	updated, err := svc.Update(ctx, address.UpdateInput{
		ID:           second.ID,
		Label:        strPtr("Gudang Utama Cakung"),
		AddressLine1: "Jl. Industri Raya No. 18",
		City:         "Jakarta Timur",
		CountryCode:  "ID",
	})
	if err != nil || updated.AddressLine1 != "Jl. Industri Raya No. 18" {
		t.Fatalf("Update failed: updated=%+v err=%v", updated, err)
	}

	// Delete menolak jika masih dirujuk, lalu berhasil setelah dilepas.
	repo.referenced[second.ID] = true
	if ref, err := svc.IsReferenced(ctx, second.ID); err != nil || !ref {
		t.Fatalf("IsReferenced expected true, got (%v, %v)", ref, err)
	}
	if err := svc.Delete(ctx, second.ID); !errors.Is(err, eferrors.ErrAddressInUse) {
		t.Fatalf("expected ErrAddressInUse, got %v", err)
	}
	repo.referenced[second.ID] = false
	if err := svc.Delete(ctx, second.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := svc.GetByID(ctx, second.ID); !errors.Is(err, eferrors.ErrAddressNotFound) {
		t.Fatalf("expected ErrAddressNotFound after delete, got %v", err)
	}
}
