package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
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

type changingScopePolicy struct {
	*mockdb.Harness
	calls int
}

func (p *changingScopePolicy) ReadScope(context.Context, service.Access) (service.Scope, error) {
	p.calls++
	return service.Scope{Filters: []domainrepo.Filter{{Column: "id", Values: []any{mockdb.ID(p.calls)}}}}, nil
}

func TestListRejectsScopeChangeBeforeReturningTotal(t *testing.T) {
	t.Parallel()
	h := mockdb.New()
	deps := h.Runtime()
	deps.Policy = &changingScopePolicy{Harness: h}
	key := func(worker *entity.Worker) string { return worker.ID }
	store := mockdb.NewStore(h, key)
	store.Rows[mockdb.ID(1)] = entity.Worker{ID: mockdb.ID(1)}
	resource, err := service.NewResource[entity.Worker, string, *entity.Worker](
		store, deps, service.Specification[entity.Worker, string]{
			Domain: "personnel", Table: "hr_workers", AggregateID: key, Key: key,
			KeyString:   func(id string) string { return id },
			ValidateKey: func(id string) error { return entity.ValidateUUID("id", id) },
			Initialize:  func(*entity.Worker, service.Actor, time.Time, string) {},
			Touch:       func(*entity.Worker, service.Actor, time.Time) {},
			Attributes:  func(*entity.Worker) map[string]string { return nil },
			Guard:       func(string, *entity.Worker, *entity.Worker) error { return nil },
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resource.List(t.Context(), nil, 1, 1); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("list with changed scope = %v, want forbidden", err)
	}
}
