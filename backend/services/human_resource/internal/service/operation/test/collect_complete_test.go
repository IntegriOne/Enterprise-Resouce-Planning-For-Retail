package service_test

import (
	"errors"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func TestCollectCompleteStopsWhenPageExceedsCount(t *testing.T) {
	t.Parallel()
	calls := 0
	records, err := service.CollectComplete(t.Context(), 1, func(limit, offset int) (service.Page[int], error) {
		calls++
		if offset != 0 {
			t.Fatalf("unexpected page after count mismatch at offset %d", offset)
		}
		return service.Page[int]{Items: make([]int, domainrepo.MaxLimit)}, nil
	})
	if !errors.Is(err, hrerrors.ErrForbidden) || records != nil || calls != 1 {
		t.Fatalf("records=%v, err=%v, calls=%d", records, err, calls)
	}
}
