package security

import (
	"context"
	"reflect"
	"strings"
	"time"

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
	if after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	application, ok := after.(*entity.Application)
	if !ok || application == nil {
		return hrerrors.ErrInvalidInput
	}
	mutating := before != nil
	rejecting := false
	shortlisting := false
	if mutating {
		previous, ok := before.(*entity.Application)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		screening := previous.Status == entity.ApplicationStatusReceived && application.Status == entity.ApplicationStatusScreening
		shortlisting = previous.Status == entity.ApplicationStatusScreening && application.Status == entity.ApplicationStatusShortlisted
		rejecting = (previous.Status == entity.ApplicationStatusReceived || previous.Status == entity.ApplicationStatusScreening ||
			previous.Status == entity.ApplicationStatusShortlisted) &&
			application.Status == entity.ApplicationStatusRejected
		switch {
		case screening:
			if previous.CurrentStageCode != nil || previous.AssignedRecruiterPrincipalID != nil ||
				application.CurrentStageCode == nil || *application.CurrentStageCode != "screening" ||
				application.AssignedRecruiterPrincipalID == nil {
				return hrerrors.ErrWorkflowRequired
			}
			expected := *previous
			expected.Status = entity.ApplicationStatusScreening
			expected.CurrentStageCode = application.CurrentStageCode
			expected.AssignedRecruiterPrincipalID = application.AssignedRecruiterPrincipalID
			if !reflect.DeepEqual(&expected, application) {
				return hrerrors.ErrWorkflowRequired
			}
		case rejecting:
			if previous.Status == entity.ApplicationStatusReceived &&
				(previous.CurrentStageCode != nil || previous.AssignedRecruiterPrincipalID != nil) {
				return hrerrors.ErrWorkflowRequired
			}
			if previous.Status == entity.ApplicationStatusScreening &&
				(previous.CurrentStageCode == nil || *previous.CurrentStageCode != "screening" ||
					previous.AssignedRecruiterPrincipalID == nil) {
				return hrerrors.ErrWorkflowRequired
			}
			if previous.Status == entity.ApplicationStatusShortlisted &&
				(previous.CurrentStageCode == nil || *previous.CurrentStageCode != "shortlisted" ||
					previous.AssignedRecruiterPrincipalID == nil) {
				return hrerrors.ErrWorkflowRequired
			}
			if application.CurrentStageCode == nil || *application.CurrentStageCode != "rejected" ||
				application.DispositionReasonCode == nil || strings.TrimSpace(*application.DispositionReasonCode) == "" ||
				application.DecisionAt == nil || application.DecisionAt.Before(application.AppliedAt) ||
				application.DecisionByPrincipalID == nil {
				return hrerrors.ErrWorkflowRequired
			}
			expected := *previous
			expected.Status = entity.ApplicationStatusRejected
			expected.CurrentStageCode = application.CurrentStageCode
			expected.DispositionReasonCode = application.DispositionReasonCode
			expected.DecisionAt = application.DecisionAt
			expected.DecisionByPrincipalID = application.DecisionByPrincipalID
			if !reflect.DeepEqual(&expected, application) {
				return hrerrors.ErrWorkflowRequired
			}
		case shortlisting:
			if previous.CurrentStageCode == nil || *previous.CurrentStageCode != "screening" ||
				previous.AssignedRecruiterPrincipalID == nil || application.CurrentStageCode == nil ||
				*application.CurrentStageCode != "shortlisted" {
				return hrerrors.ErrWorkflowRequired
			}
			expected := *previous
			expected.Status = entity.ApplicationStatusShortlisted
			expected.CurrentStageCode = application.CurrentStageCode
			if !reflect.DeepEqual(&expected, application) {
				return hrerrors.ErrWorkflowRequired
			}
		default:
			return hrerrors.ErrWorkflowRequired
		}
	} else if application.Status != entity.ApplicationStatusReceived || application.CurrentStageCode != nil ||
		application.AssignedRecruiterPrincipalID != nil {
		return hrerrors.ErrWorkflowRequired
	}
	if application.ScreeningScore != nil || application.ScorecardVersion != nil ||
		(!rejecting && (application.DispositionReasonCode != nil || application.DecisionAt != nil || application.DecisionByPrincipalID != nil)) ||
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
		entity.CompareDates(*candidate.DataRetentionUntil, application.AppliedAt) <= 0 ||
		(mutating && entity.CompareDates(*candidate.DataRetentionUntil, time.Now()) <= 0) {
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
		(mutating && requisition.ClosedAt != nil) ||
		(requisition.ClosedAt != nil && !application.AppliedAt.Before(*requisition.ClosedAt)) {
		return hrerrors.ErrWorkflowRequired
	}
	return nil
}
