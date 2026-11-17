package security

import (
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type postingRequisitionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Requisition, error)
}

type postingApplicationCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type jobPostingValidator struct {
	requisitions postingRequisitionReader
	applications postingApplicationCounter
}

// RecruitmentJobPostingValidators menjaga transisi posting pada requisition yang terbuka.
func RecruitmentJobPostingValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Requisition) != nil ||
		operation.RequireDependency(repos.Application) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_job_postings": &jobPostingValidator{requisitions: repos.Requisition, applications: repos.Application},
	}, nil
}

func (v *jobPostingValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_job_postings" || operation.RequireDependency(v.requisitions) != nil ||
		operation.RequireDependency(v.applications) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.JobPosting
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.JobPosting)
		if !ok || previous == nil || previous.ClosesAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous.Status == "draft" && (previous.PublishedAt != nil || previous.ExternalPostingReference != nil) {
			return hrerrors.ErrWorkflowRequired
		}
		if previous.Status == "published" && previous.PublishedAt == nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous.Status != "draft" && previous.Status != "published" {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	posting := previous
	publishing := false
	closing := false
	if after != nil {
		var ok bool
		posting, ok = after.(*entity.JobPosting)
		if !ok || posting == nil {
			return hrerrors.ErrWorkflowRequired
		}
		publishing = previous != nil && previous.Status == "draft" && posting.Status == "published"
		closing = previous != nil && previous.Status == "published" && posting.Status == "closed"
		if publishing {
			original := *posting
			original.Status = "draft"
			original.PublishedAt = nil
			if posting.PublishedAt == nil || posting.ClosesAt != nil || posting.ExternalPostingReference != nil || len(posting.ContentDigest) != 32 ||
				!reflect.DeepEqual(previous, &original) {
				return hrerrors.ErrWorkflowRequired
			}
		} else if closing {
			original := *posting
			original.Status = "published"
			original.ClosesAt = nil
			if posting.PublishedAt == nil || posting.ClosesAt == nil || !posting.ClosesAt.After(*posting.PublishedAt) ||
				!reflect.DeepEqual(previous, &original) {
				return hrerrors.ErrWorkflowRequired
			}
		} else if previous != nil && previous.Status == "published" {
			return hrerrors.ErrWorkflowRequired
		} else if posting.Status != "draft" || posting.PublishedAt != nil || posting.ClosesAt != nil || posting.ExternalPostingReference != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != posting.ID || previous.RequisitionID != posting.RequisitionID ||
			previous.CreatedByPrincipalID != posting.CreatedByPrincipalID ||
			previous.ChannelCode != posting.ChannelCode || previous.PostingVersion != posting.PostingVersion ||
			(!publishing && !reflect.DeepEqual(previous.PublishedAt, posting.PublishedAt))) {
			return hrerrors.ErrWorkflowRequired
		}
	} else if previous != nil && previous.Status != "draft" {
		return hrerrors.ErrWorkflowRequired
	}
	requisition, err := v.requisitions.FindByIDForUpdate(ctx, posting.RequisitionID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if requisition == nil || requisition.ApprovedAt == nil || requisition.ApprovedByPrincipalID == nil ||
		requisition.OpenedAt == nil || requisition.BudgetReferenceID == nil ||
		requisition.OpenedAt.Before(*requisition.ApprovedAt) {
		return hrerrors.ErrReferenceConflict
	}
	if after != nil && (requisition.Status != "open" || requisition.ClosedAt != nil) {
		return hrerrors.ErrWorkflowRequired
	}
	if publishing && posting.PublishedAt.Before(*requisition.OpenedAt) {
		return hrerrors.ErrWorkflowRequired
	}
	if after == nil && requisition.Status != "open" && requisition.Status != "closed" {
		return hrerrors.ErrWorkflowRequired
	}
	if previous != nil && previous.Status == "draft" {
		count, err := v.applications.CountWhere(ctx, []domainrepo.Filter{{Column: "source_posting_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if count != 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
