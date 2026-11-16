package service_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func TestMissingDependenciesFailClosed(t *testing.T) {
	_, err := service.NewResource[entity.Worker, string, *entity.Worker](nil, service.Dependencies{}, service.Specification[entity.Worker, string]{})
	if !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("missing dependencies: %v", err)
	}
}

func TestTypedNilDependenciesFailClosed(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"repository", "transactor", "policy", "integrity", "outbox"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			h := mockdb.New()
			deps := h.Runtime()
			var missing *mockdb.Harness
			var repo service.Repository[entity.Worker, string] = mockdb.NewStore[entity.Worker, string](
				h,
				func(worker *entity.Worker) string { return worker.ID },
			)
			switch name {
			case "repository":
				var store *mockdb.Store[entity.Worker, string]
				repo = store
			case "transactor":
				deps.Transactor = missing
			case "policy":
				deps.Policy = missing
			case "integrity":
				deps.Integrity = missing
			case "outbox":
				deps.Outbox = missing
			}
			resource, err := service.NewResource[entity.Worker, string, *entity.Worker](
				repo,
				deps,
				service.Specification[entity.Worker, string]{},
			)
			if resource != nil || !errors.Is(err, hrerrors.ErrUnavailable) {
				t.Fatalf("typed nil %s: resource %v, error %v", name, resource, err)
			}
			if h.TransactionCalls != 0 || h.WriteCalls != 0 {
				t.Fatal("invalid constructor reached persistence")
			}
		})
	}
}
