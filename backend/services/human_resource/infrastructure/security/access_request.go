package security

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type accessRequestReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.AccessRequest, error)
}

type accessRequestLifecycleReader interface {
	FindByIDForUpdate(context.Context, int64) (*entity.WorkerLifecycleEvent, error)
}

type accessRequestValidator struct {
	workers  lockedWorkerReader
	events   accessRequestLifecycleReader
	requests accessRequestReader
}

// AccessRequestValidators mendaftarkan validator usulan permintaan akses pekerja HRM ke IAM.
func AccessRequestValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil ||
		operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.WorkerLifecycleEvent) != nil ||
		operation.RequireDependency(repos.AccessRequest) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_access_requests": &accessRequestValidator{
			workers:  repos.Worker,
			events:   repos.WorkerLifecycleEvent,
			requests: repos.AccessRequest,
		},
	}, nil
}

// ValidateMutation membatasi mutasi lokal ke pengajuan berstatus pending tanpa menulis keputusan/grant IAM.
func (v *accessRequestValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_access_requests" ||
		operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.events) != nil ||
		operation.RequireDependency(v.requests) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.AccessRequest
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.AccessRequest)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status != "pending" || previous.IAMRequestReferenceID != nil ||
			previous.IAMDecisionEventReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return nil
	}
	req, ok := after.(*entity.AccessRequest)
	if !ok || req == nil || (previous != nil && previous.ID != req.ID) {
		return hrerrors.ErrInvalidInput
	}
	if req.Status != "pending" || req.IAMRequestReferenceID != nil || req.IAMDecisionEventReferenceID != nil {
		return hrerrors.ErrWorkflowRequired
	}
	if !supportedIAMScopeType(req.RequestedScopeType) {
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
	} else if previous.WorkerID != req.WorkerID || previous.LifecycleEventID != req.LifecycleEventID ||
		previous.IdempotencyKey != req.IdempotencyKey ||
		previous.RequestedByPrincipalID != req.RequestedByPrincipalID || !previous.RequestedAt.Equal(req.RequestedAt) {
		return hrerrors.ErrWorkflowRequired
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, req.WorkerID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if worker == nil {
		return hrerrors.ErrReferenceConflict
	}
	if req.RequestKind == "provision" {
		if worker.Status != entity.WorkerStatusActive || worker.IAMPrincipalReferenceID != nil || worker.TerminationDate != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if req.RequestedScopeType == "legal_entity" && req.RequestedScopeReferenceID != worker.PrimaryLegalEntityReferenceID {
			return hrerrors.ErrReferenceConflict
		}
		// Kunci worker di atas menyerialkan pengajuan provisioning untuk pekerja yang sama.
		open, err := v.requests.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "worker_id", Values: []any{req.WorkerID}},
			{Column: "request_kind", Values: []any{"provision"}},
			{Column: "status", Values: []any{"pending", "submitted", "approved"}},
		}, 2, 0)
		if err != nil {
			return err
		}
		for _, existing := range open {
			if existing.ID != req.ID {
				return hrerrors.ErrAlreadyExists
			}
		}
	}
	event, err := v.events.FindByIDForUpdate(ctx, req.LifecycleEventID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if event == nil || event.WorkerID != req.WorkerID || event.OccurredAt.After(req.RequestedAt) {
		return hrerrors.ErrReferenceConflict
	}
	if req.RequestKind == "provision" && (event.EventType != "hire" || event.NewStatus == nil || *event.NewStatus != entity.WorkerStatusActive) {
		return hrerrors.ErrWorkflowRequired
	}
	matches, err := v.requests.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "idempotency_key", Values: []any{req.IdempotencyKey}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != req.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func supportedIAMScopeType(value string) bool {
	switch value {
	case "global", "legal_entity", "organization_unit", "facility", "supplier_party", "vendor_party", "self":
		return true
	default:
		return false
	}
}
