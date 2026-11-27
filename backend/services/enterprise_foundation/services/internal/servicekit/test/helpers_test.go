package servicekit_test

import (
	"context"
	"errors"
	"testing"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/internal/servicekit"
)

func TestCollectAllPages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	calls := 0
	items, err := servicekit.CollectAllPages(ctx, func(_ context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[int], error) {
		calls++
		if page.Limit != domainrepo.MaxPageLimit {
			t.Fatalf("expected limit %d, got %d", domainrepo.MaxPageLimit, page.Limit)
		}
		if page.Offset == 0 {
			return domainrepo.NewPageResult([]int{1, 2}, 3, page), nil
		}
		return domainrepo.NewPageResult([]int{3}, 3, page), nil
	})
	if err != nil {
		t.Fatalf("CollectAllPages error = %v", err)
	}
	if calls != 2 || len(items) != 3 || items[0] != 1 || items[2] != 3 {
		t.Fatalf("unexpected CollectAllPages result: calls=%d items=%v", calls, items)
	}

	// Uji batas maksimum agar tidak terjebak loop tak terbatas saat stub salah konfigurasi.
	_, err = servicekit.CollectAllPages(ctx, func(_ context.Context, page domainrepo.PageRequest) (domainrepo.PageResult[int], error) {
		batch := make([]int, domainrepo.MaxPageLimit)
		return domainrepo.PageResult[int]{
			Items:   batch,
			Total:   servicekit.MaxCollectAllItems + 500,
			Limit:   page.Limit,
			Offset:  page.Offset,
			HasMore: true,
		}, nil
	})
	if !errors.Is(err, eferrors.ErrConflict) {
		t.Fatalf("expected ErrConflict when bound exceeded, got %v", err)
	}
}

func TestValidateOptionalUUIDList(t *testing.T) {
	t.Parallel()

	idA := "11111111-1111-4111-8111-111111111111"
	idB := "22222222-2222-4222-8222-222222222222"
	got, err := servicekit.ValidateOptionalUUIDList("ids", []string{" " + idA + " ", idB, idA})
	if err != nil || len(got) != 2 {
		t.Fatalf("unexpected ValidateOptionalUUIDList result: got=%v err=%v", got, err)
	}

	if _, err := servicekit.ValidateOptionalUUIDList("ids", []string{idA, "bad-uuid"}); !errors.Is(err, eferrors.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
