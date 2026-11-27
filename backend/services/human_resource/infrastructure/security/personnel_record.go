package security

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type emergencyContactReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.EmergencyContact, error)
}

type workerDocumentValidator struct {
	workers lockedWorkerReader
}

type emergencyContactValidator struct {
	workers    lockedWorkerReader
	contacts   emergencyContactReader
	foundation FoundationVerifier
}

// PersonnelRecordValidators memasang aturan dokumen pekerja pending dan kontak darurat dengan verifikasi Party Foundation.
func PersonnelRecordValidators(repos *repository.Registry, foundation FoundationVerifier) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.EmergencyContact) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_worker_documents": &workerDocumentValidator{workers: repos.Worker},
		"hr_emergency_contacts": &emergencyContactValidator{
			workers: repos.Worker, contacts: repos.EmergencyContact, foundation: foundation,
		},
	}, nil
}

func (v *workerDocumentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_worker_documents" || operation.RequireDependency(v.workers) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.WorkerDocument
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.WorkerDocument)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.VerificationStatus != "pending" || previous.VerifiedByPrincipalID != nil || previous.VerifiedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	target := previous
	if after != nil {
		var ok bool
		target, ok = after.(*entity.WorkerDocument)
		if !ok || target == nil {
			return hrerrors.ErrInvalidInput
		}
		if target.VerificationStatus != "pending" || target.VerifiedByPrincipalID != nil || target.VerifiedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != target.ID || previous.WorkerID != target.WorkerID ||
			!previous.CreatedAt.Equal(target.CreatedAt)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, target.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil {
		return hrerrors.ErrReferenceConflict
	}
	if after != nil && (worker.Status == entity.WorkerStatusTerminated || worker.Status == entity.WorkerStatusDeceased) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func (v *emergencyContactValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_emergency_contacts" || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.contacts) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.EmergencyContact
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.EmergencyContact)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	target := previous
	if after != nil {
		var ok bool
		target, ok = after.(*entity.EmergencyContact)
		if !ok || target == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous != nil && (previous.ID != target.ID || previous.WorkerID != target.WorkerID) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, target.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || (after != nil && worker.Status == entity.WorkerStatusDeceased) {
		return hrerrors.ErrReferenceConflict
	}
	if after == nil {
		return nil
	}
	if target.AuthorizationConfirmedAt != nil && target.AuthorizationConfirmedAt.After(time.Now()) {
		return hrerrors.Invalid("authorization_confirmed_at", "must not be in the future")
	}
	if strings.EqualFold(target.ContactPartyReferenceID, worker.PersonPartyReferenceID) {
		return hrerrors.ErrReferenceConflict
	}
	if err := verifyFoundationParty(ctx, v.foundation, target.ContactPartyReferenceID, true); err != nil {
		return err
	}
	matches, err := v.contacts.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "worker_id", Values: []any{target.WorkerID}},
		{Column: "priority_order", Values: []any{target.PriorityOrder}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != target.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}
