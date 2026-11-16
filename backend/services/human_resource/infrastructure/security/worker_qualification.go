package security

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type workerQualificationValidator struct {
	workers lockedWorkerReader
}

// ValidateMutation membatasi CRUD kualifikasi pada draf yang belum diverifikasi.
func (v *workerQualificationValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_worker_qualifications" || operation.RequireDependency(v.workers) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.WorkerQualification
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.WorkerQualification)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.VerificationStatus != "unverified" || previous.VerifiedByPrincipalID != nil || previous.VerifiedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil && previous == nil {
		return hrerrors.ErrInvalidInput
	}
	qualification := previous
	if after != nil {
		var ok bool
		qualification, ok = after.(*entity.WorkerQualification)
		if !ok || qualification == nil {
			return hrerrors.ErrInvalidInput
		}
		if qualification.VerificationStatus != "unverified" || qualification.VerifiedByPrincipalID != nil || qualification.VerifiedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != qualification.ID || previous.WorkerID != qualification.WorkerID ||
			previous.QualificationType != qualification.QualificationType) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if qualification.IssuerPartyReferenceID != nil || qualification.CredentialReferenceID != nil {
		return hrerrors.ErrUnavailable
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, qualification.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || (after != nil && worker.Status != entity.WorkerStatusActive) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}
