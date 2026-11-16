package security

import (
	"context"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type applicationCandidateReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Candidate, error)
}

type applicationRequisitionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Requisition, error)
}

type applicationValidator struct {
	candidates   applicationCandidateReader
	requisitions applicationRequisitionReader
}

// RecruitmentApplicationValidators membuka penerimaan lamaran tanpa mengaktifkan keputusan seleksi.
func RecruitmentApplicationValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Candidate) != nil ||
		operation.RequireDependency(repos.Requisition) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_applications": &applicationValidator{candidates: repos.Candidate, requisitions: repos.Requisition},
	}, nil
}

func (v *applicationValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_applications" || operation.RequireDependency(v.candidates) != nil ||
		operation.RequireDependency(v.requisitions) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before != nil || after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	application, ok := after.(*entity.Application)
	if !ok || application == nil {
		return hrerrors.ErrInvalidInput
	}
	if application.Status != entity.ApplicationStatusReceived || application.CurrentStageCode != nil ||
		application.AssignedRecruiterPrincipalID != nil || application.ScreeningScore != nil ||
		application.ScorecardVersion != nil || application.DispositionReasonCode != nil ||
		application.DecisionAt != nil || application.DecisionByPrincipalID != nil ||
		application.HiredWorkerID != nil {
		return hrerrors.ErrWorkflowRequired
	}
	if application.SourcePostingID != nil {
		return hrerrors.ErrUnavailable
	}
	candidate, err := v.candidates.FindByIDForUpdate(ctx, application.CandidateID)
	if err != nil {
		return err
	}
	if candidate == nil || candidate.CandidateStatus != "active" || candidate.NoticeAcknowledgedAt == nil ||
		candidate.DataRetentionUntil == nil || strings.TrimSpace(candidate.PrivacyNoticeVersion) == "" {
		return hrerrors.ErrWorkflowRequired
	}
	if candidate.NoticeAcknowledgedAt.After(application.AppliedAt) ||
		entity.CompareDates(*candidate.DataRetentionUntil, application.AppliedAt) <= 0 {
		return hrerrors.ErrWorkflowRequired
	}
	requisition, err := v.requisitions.FindByIDForUpdate(ctx, application.RequisitionID)
	if err != nil {
		return err
	}
	if requisition == nil || requisition.Status != "open" || requisition.ApprovedAt == nil ||
		requisition.OpenedAt == nil || requisition.ApprovedByPrincipalID == nil ||
		requisition.ApprovedAt.After(*requisition.OpenedAt) ||
		requisition.OpenedAt.After(application.AppliedAt) ||
		(requisition.ClosedAt != nil && !application.AppliedAt.Before(*requisition.ClosedAt)) {
		return hrerrors.ErrWorkflowRequired
	}
	return nil
}
