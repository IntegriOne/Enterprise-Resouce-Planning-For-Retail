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

type hrsRequestStub struct {
	byID map[string]entity.ServiceRequest
	rows []entity.ServiceRequest
}

func (s *hrsRequestStub) FindByIDForUpdate(_ context.Context, id string) (*entity.ServiceRequest, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *hrsRequestStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.ServiceRequest, error) {
	var out []entity.ServiceRequest
	for _, row := range s.rows {
		match := true
		for _, f := range filters {
			switch f.Column {
			case "request_number":
				if row.RequestNumber != f.Values[0] {
					match = false
				}
			case "idempotency_key":
				if row.IdempotencyKey != f.Values[0] {
					match = false
				}
			}
		}
		if match {
			out = append(out, row)
		}
	}
	return out, nil
}

type hrsEventStub struct {
	rows  []entity.ServiceRequestEvent
	count int64
}

func (s *hrsEventStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.ServiceRequestEvent, error) {
	var out []entity.ServiceRequestEvent
	for _, row := range s.rows {
		if len(filters) == 1 && filters[0].Column == "idempotency_key" && row.IdempotencyKey == filters[0].Values[0] {
			out = append(out, row)
		}
	}
	return out, nil
}

func (s *hrsEventStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

func TestHRServiceValidators(t *testing.T) {
	t.Parallel()
	requester := "00000000-0000-4000-8000-000000000401"
	ctx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: requester, Authenticated: true})
	reqs := &hrsRequestStub{byID: map[string]entity.ServiceRequest{}}
	evs := &hrsEventStub{}
	reqVal := &serviceRequestValidator{
		workers:  &lockedWorkerStub{worker: &entity.Worker{ID: "worker-1", Status: entity.WorkerStatusActive}, order: &[]string{}},
		requests: reqs,
		events:   evs,
	}
	evVal := &serviceRequestEventValidator{requests: reqs, events: evs}

	req := &entity.ServiceRequest{
		ID: "req-1", RequestNumber: "SR-001", WorkerID: "worker-1", RequestType: "employment_letter",
		RequestedByPrincipalID: requester, Status: "open",
		RestrictedDetailsDocumentVersionReferenceID: "doc-1", OpenedAt: time.Now().UTC(), IdempotencyKey: "idem-sr-1",
	}
	if err := reqVal.ValidateMutation(ctx, "hr_service_requests", nil, req); err != nil {
		t.Fatalf("create valid service request: %v", err)
	}
	invalidDueAt := *req
	yesterday := req.OpenedAt.Add(-24 * time.Hour)
	invalidDueAt.DueAt = &yesterday
	if err := reqVal.ValidateMutation(ctx, "hr_service_requests", nil, &invalidDueAt); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Errorf("create with deadline earlier than opening must fail: %v", err)
	}
	if err := reqVal.ValidateMutation(ctx, "hr_service_requests", req, &invalidDueAt); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Errorf("update with deadline earlier than opening must fail: %v", err)
	}
	reqs.byID["req-1"] = *req

	ev := &entity.ServiceRequestEvent{
		ID: "ev-1", ServiceRequestID: "req-1", EventType: "commented", ActorPrincipalID: requester, OccurredAt: time.Now().UTC(), IdempotencyKey: "idem-ev-1",
	}
	if err := evVal.ValidateMutation(ctx, "hr_service_request_events", nil, ev); err != nil {
		t.Fatalf("create valid request event: %v", err)
	}
	backdated := *ev
	backdated.OccurredAt = req.OpenedAt.Add(-time.Second)
	if err := evVal.ValidateMutation(ctx, "hr_service_request_events", nil, &backdated); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Errorf("event earlier than request opening must fail: %v", err)
	}
	for _, eventType := range []string{"assigned", "resolved", "closed", "custom"} {
		transition := *ev
		transition.EventType = eventType
		if err := evVal.ValidateMutation(ctx, "hr_service_request_events", nil, &transition); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
			t.Errorf("generic %s event must fail: %v", eventType, err)
		}
	}
	decision := entity.ApprovalDecision("approved")
	withDecision := *ev
	withDecision.Decision = &decision
	if err := evVal.ValidateMutation(ctx, "hr_service_request_events", nil, &withDecision); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Errorf("generic event with decision must fail: %v", err)
	}
	if err := evVal.ValidateMutation(ctx, "hr_service_request_events", ev, ev); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("update request event must fail: %v", err)
	}
	evs.count = 1
	if err := reqVal.ValidateMutation(ctx, "hr_service_requests", req, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete service request with events must fail: %v", err)
	}
}
