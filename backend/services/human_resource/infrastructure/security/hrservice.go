package security

import (
	"context"
	"reflect"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type hrServiceRequestReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.ServiceRequest, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.ServiceRequest, error)
}

type hrServiceRequestEventReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.ServiceRequestEvent, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type serviceRequestValidator struct {
	workers  lockedWorkerReader
	requests hrServiceRequestReader
	events   hrServiceRequestEventReader
}

type serviceRequestEventValidator struct {
	requests hrServiceRequestReader
	events   hrServiceRequestEventReader
}

// HRServiceValidators mendaftarkan validator tiket layanan HR internal dan riwayat kejadiannya.
func HRServiceValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil ||
		operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.ServiceRequest) != nil ||
		operation.RequireDependency(repos.ServiceRequestEvent) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_service_requests": &serviceRequestValidator{
			workers:  repos.Worker,
			requests: repos.ServiceRequest,
			events:   repos.ServiceRequestEvent,
		},
		"hr_service_request_events": &serviceRequestEventValidator{
			requests: repos.ServiceRequest,
			events:   repos.ServiceRequestEvent,
		},
	}, nil
}

// ValidateMutation menjaga draf permintaan layanan HR dan mencegah penghapusan tiket berriwayat.
func (v *serviceRequestValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_service_requests" ||
		operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.requests) != nil ||
		operation.RequireDependency(v.events) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.ServiceRequest
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.ServiceRequest)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status == "closed" || previous.Status == "cancelled" {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status != "open" {
			return hrerrors.ErrWorkflowRequired
		}
		count, err := v.events.CountWhere(ctx, []domainrepo.Filter{
			{Column: "service_request_id", Values: []any{previous.ID}},
		})
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	req, ok := after.(*entity.ServiceRequest)
	if !ok || req == nil || (previous != nil && previous.ID != req.ID) {
		return hrerrors.ErrInvalidInput
	}
	if previous != nil && (previous.RequestNumber != req.RequestNumber || previous.IdempotencyKey != req.IdempotencyKey) {
		return hrerrors.ErrWorkflowRequired
	}
	if !supportedServiceRequestType(req.RequestType) {
		return hrerrors.ErrInvalidInput
	}
	if previous == nil {
		caller, ok := authcontext.FromContext(ctx)
		if !ok || !caller.Authenticated {
			return hrerrors.ErrUnauthenticated
		}
		if !strings.EqualFold(caller.PrincipalID, req.RequestedByPrincipalID) {
			return hrerrors.ErrForbidden
		}
		if req.Status != "open" || req.AssignedPrincipalID != nil || req.ResolvedAt != nil || req.ResolutionDocumentVersionReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
	} else if previous.Status == "open" && req.Status == "cancelled" {
		caller, ok := authcontext.FromContext(ctx)
		if !ok || !caller.Authenticated || !strings.EqualFold(caller.PrincipalID, previous.RequestedByPrincipalID) {
			return hrerrors.ErrForbidden
		}
		expected := *previous
		expected.Status = "cancelled"
		if previous.AssignedPrincipalID != nil || previous.ResolvedAt != nil ||
			previous.ResolutionDocumentVersionReferenceID != nil || !reflect.DeepEqual(&expected, req) {
			return hrerrors.ErrWorkflowRequired
		}
		return nil
	} else if previous.Status == "open" && req.Status == "in_progress" {
		caller, ok := authcontext.FromContext(ctx)
		if !ok || !caller.Authenticated || req.AssignedPrincipalID == nil ||
			!strings.EqualFold(caller.PrincipalID, *req.AssignedPrincipalID) {
			return hrerrors.ErrForbidden
		}
		expected := *previous
		expected.Status = "in_progress"
		expected.AssignedPrincipalID = req.AssignedPrincipalID
		if previous.AssignedPrincipalID != nil || previous.ResolvedAt != nil ||
			previous.ResolutionDocumentVersionReferenceID != nil || !reflect.DeepEqual(&expected, req) {
			return hrerrors.ErrWorkflowRequired
		}
	} else if previous.Status == "in_progress" && req.Status == "resolved" {
		caller, ok := authcontext.FromContext(ctx)
		if !ok || !caller.Authenticated || previous.AssignedPrincipalID == nil ||
			!strings.EqualFold(caller.PrincipalID, *previous.AssignedPrincipalID) {
			return hrerrors.ErrForbidden
		}
		expected := *previous
		expected.Status = "resolved"
		expected.ResolvedAt = req.ResolvedAt
		expected.ResolutionDocumentVersionReferenceID = req.ResolutionDocumentVersionReferenceID
		if previous.ResolvedAt != nil || previous.ResolutionDocumentVersionReferenceID != nil ||
			req.ResolvedAt == nil || req.ResolvedAt.Before(req.OpenedAt) ||
			req.ResolutionDocumentVersionReferenceID == nil || !reflect.DeepEqual(&expected, req) {
			return hrerrors.ErrWorkflowRequired
		}
	} else if previous.Status == "resolved" && req.Status == "closed" {
		caller, ok := authcontext.FromContext(ctx)
		if !ok || !caller.Authenticated || !strings.EqualFold(caller.PrincipalID, previous.RequestedByPrincipalID) {
			return hrerrors.ErrForbidden
		}
		expected := *previous
		expected.Status = "closed"
		if previous.ResolvedAt == nil || previous.ResolutionDocumentVersionReferenceID == nil || !reflect.DeepEqual(&expected, req) {
			return hrerrors.ErrWorkflowRequired
		}
		return nil
	} else if previous.Status == "resolved" || !reflect.DeepEqual(req.AssignedPrincipalID, previous.AssignedPrincipalID) || req.Status != previous.Status || req.WorkerID != previous.WorkerID ||
		req.RequestedByPrincipalID != previous.RequestedByPrincipalID || !req.OpenedAt.Equal(previous.OpenedAt) ||
		req.ResolvedAt != nil || req.ResolutionDocumentVersionReferenceID != nil {
		return hrerrors.ErrWorkflowRequired
	}
	// Tenggat historis tetap dapat dipertahankan saat workflow mengubah status.
	if req.DueAt != nil && req.DueAt.Before(req.OpenedAt) &&
		(previous == nil || previous.DueAt == nil || !req.DueAt.Equal(*previous.DueAt)) {
		return hrerrors.Invalid("due_at", "must not precede request opening")
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, req.WorkerID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if worker == nil {
		return hrerrors.ErrReferenceConflict
	}
	byNumber, err := v.requests.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "request_number", Values: []any{req.RequestNumber}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range byNumber {
		if existing.ID != req.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	byIdem, err := v.requests.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "idempotency_key", Values: []any{req.IdempotencyKey}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range byIdem {
		if existing.ID != req.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

// ValidateMutation menjaga riwayat kejadian permintaan layanan tetap append-only.
func (v *serviceRequestEventValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_service_request_events" ||
		operation.RequireDependency(v.requests) != nil ||
		operation.RequireDependency(v.events) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if before != nil || after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	event, ok := after.(*entity.ServiceRequestEvent)
	if !ok || event == nil {
		return hrerrors.ErrInvalidInput
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return hrerrors.ErrUnauthenticated
	}
	if !strings.EqualFold(caller.PrincipalID, event.ActorPrincipalID) {
		return hrerrors.ErrForbidden
	}
	req, err := v.requests.FindByIDForUpdate(ctx, event.ServiceRequestID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if req == nil {
		return hrerrors.ErrReferenceConflict
	}
	// Keputusan dan transisi status hanya dicatat oleh operasi workflow.
	if event.EventType != "commented" || event.Decision != nil {
		return hrerrors.ErrWorkflowRequired
	}
	if req.Status == "closed" || req.Status == "cancelled" {
		return hrerrors.ErrWorkflowRequired
	}
	if event.OccurredAt.Before(req.OpenedAt) {
		return hrerrors.Invalid("occurred_at", "must not precede request opening")
	}
	// Lampiran menunggu pembuktian versi dokumen dari DRM.
	if event.EvidenceDocumentVersionReferenceID != nil {
		return hrerrors.ErrUnavailable
	}
	byIdem, err := v.events.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "idempotency_key", Values: []any{event.IdempotencyKey}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range byIdem {
		if existing.ID != event.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func supportedServiceRequestType(value string) bool {
	switch value {
	case "employment_letter", "data_correction", "benefit_query", "payroll_query", "policy_query", "other":
		return true
	default:
		return false
	}
}
