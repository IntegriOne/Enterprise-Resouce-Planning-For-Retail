package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type accessReqRepoStub struct {
	rows []entity.AccessRequest
}

func (s *accessReqRepoStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.AccessRequest, error) {
	var out []entity.AccessRequest
	for _, row := range s.rows {
		if len(filters) == 1 && filters[0].Column == "idempotency_key" && row.IdempotencyKey == filters[0].Values[0] {
			out = append(out, row)
		}
		if len(filters) == 3 && filters[0].Column == "worker_id" && row.WorkerID == filters[0].Values[0] &&
			row.RequestKind == "provision" && (row.Status == "pending" || row.Status == "submitted" || row.Status == "approved") {
			out = append(out, row)
		}
	}
	return out, nil
}

func TestAccessRequestValidator(t *testing.T) {
	t.Parallel()
	requester := "00000000-0000-4000-8000-000000000601"
	ctx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: requester, Authenticated: true})
	active := entity.WorkerStatusActive
	events := &promoEventStub{byID: map[int64]entity.WorkerLifecycleEvent{
		1: {ID: 1, WorkerID: "worker-1", EventType: "hire", NewStatus: &active},
		2: {ID: 2, WorkerID: "other-worker"},
		3: {ID: 3, WorkerID: "worker-1", EventType: "principal_linked"},
	}}
	validator := &accessRequestValidator{
		workers:  &lockedWorkerStub{worker: &entity.Worker{ID: "worker-1", Status: entity.WorkerStatusActive, PrimaryLegalEntityReferenceID: "le-1"}, order: &[]string{}},
		events:   events,
		requests: &accessReqRepoStub{},
	}
	req := &entity.AccessRequest{
		ID: "ar-1", WorkerID: "worker-1", LifecycleEventID: 1, RequestKind: "provision",
		RequestedScopeType: "legal_entity", RequestedScopeReferenceID: "le-1", Status: "pending",
		RequestedByPrincipalID: requester, RequestedAt: time.Now().UTC(), IdempotencyKey: "idem-ar-1",
	}
	if err := validator.ValidateMutation(ctx, "hr_access_requests", nil, req); err != nil {
		t.Fatalf("create valid pending access request: %v", err)
	}
	wrongEntity := *req
	wrongEntity.RequestedScopeReferenceID = "le-2"
	if err := validator.ValidateMutation(ctx, "hr_access_requests", nil, &wrongEntity); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("a provisioning request for a different legal entity must fail: %v", err)
	}
	linked := "existing-account"
	worker := validator.workers.(*lockedWorkerStub).worker
	worker.IAMPrincipalReferenceID = &linked
	if err := validator.ValidateMutation(ctx, "hr_access_requests", nil, req); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("a worker with an account must not request provisioning: %v", err)
	}
	worker.IAMPrincipalReferenceID = nil
	duplicate := *req
	duplicate.ID = "ar-2"
	duplicate.IdempotencyKey = "idem-ar-2"
	validator.requests = &accessReqRepoStub{rows: []entity.AccessRequest{duplicate}}
	if err := validator.ValidateMutation(ctx, "hr_access_requests", nil, req); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("a second open request must fail: %v", err)
	}
	validator.requests = &accessReqRepoStub{}
	crossWorker := *req
	crossWorker.LifecycleEventID = 2
	if err := validator.ValidateMutation(ctx, "hr_access_requests", nil, &crossWorker); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("mismatched worker lifecycle event must fail: %v", err)
	}
	otherEvent := *req
	otherEvent.LifecycleEventID = 3
	if err := validator.ValidateMutation(ctx, "hr_access_requests", nil, &otherEvent); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("a non-activation event must not allow provisioning: %v", err)
	}
	futureEvent := events.byID[1]
	futureEvent.OccurredAt = req.RequestedAt.Add(time.Second)
	events.byID[1] = futureEvent
	if err := validator.ValidateMutation(ctx, "hr_access_requests", nil, req); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("access request referring to a future lifecycle event: %v", err)
	}
	futureEvent.OccurredAt = req.RequestedAt
	events.byID[1] = futureEvent
	if err := validator.ValidateMutation(ctx, "hr_access_requests", nil, req); err != nil {
		t.Fatalf("access request at the lifecycle event time: %v", err)
	}
	badScope := *req
	badScope.RequestedScopeType = "unknown_scope"
	if err := validator.ValidateMutation(ctx, "hr_access_requests", nil, &badScope); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("invalid IAM scope type must fail: %v", err)
	}
	approved := *req
	approved.Status = "approved"
	if err := validator.ValidateMutation(ctx, "hr_access_requests", req, &approved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("direct approval of access request must fail: %v", err)
	}
}
