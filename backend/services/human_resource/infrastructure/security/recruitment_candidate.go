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

type candidateReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Candidate, error)
}

type candidateApplicationCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type CandidateFoundationVerifier interface {
	FoundationVerifier
	VerifyPartyContact(context.Context, string, string) error
}

type candidateValidator struct {
	candidates   candidateReader
	applications candidateApplicationCounter
	foundation   CandidateFoundationVerifier
}

// RecruitmentCandidateValidators memeriksa referensi identitas kandidat melalui Enterprise Foundation.
func RecruitmentCandidateValidators(repos *repository.Registry, foundation CandidateFoundationVerifier) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Candidate) != nil ||
		operation.RequireDependency(repos.Application) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_candidates": &candidateValidator{candidates: repos.Candidate, applications: repos.Application, foundation: foundation},
	}, nil
}

func (v *candidateValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_candidates" || operation.RequireDependency(v.candidates) != nil ||
		operation.RequireDependency(v.applications) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var old *entity.Candidate
	if before != nil {
		var ok bool
		old, ok = before.(*entity.Candidate)
		if !ok || old == nil || old.CandidateStatus != "active" {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if old == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	candidate := old
	if after != nil {
		var ok bool
		candidate, ok = after.(*entity.Candidate)
		if !ok || candidate == nil || candidate.CandidateStatus != "active" {
			return hrerrors.ErrWorkflowRequired
		}
		if candidate.SourceType != nil || candidate.SourceReference != nil {
			return hrerrors.ErrUnavailable
		}
		if old != nil && (old.ID != candidate.ID || old.CandidateNumber != candidate.CandidateNumber ||
			old.PrivacyNoticeVersion != candidate.PrivacyNoticeVersion ||
			!reflect.DeepEqual(old.PersonPartyReferenceID, candidate.PersonPartyReferenceID) ||
			!reflect.DeepEqual(old.ContactReferenceID, candidate.ContactReferenceID) ||
			!reflect.DeepEqual(old.NoticeAcknowledgedAt, candidate.NoticeAcknowledgedAt) ||
			!reflect.DeepEqual(old.DataRetentionUntil, candidate.DataRetentionUntil) ||
			!reflect.DeepEqual(old.CreatedByPrincipalID, candidate.CreatedByPrincipalID) ||
			!old.CreatedAt.Equal(candidate.CreatedAt)) {
			return hrerrors.ErrWorkflowRequired
		}
		if err := validateCandidateNotice(candidate); err != nil {
			return err
		}
	}
	if after == nil {
		count, err := v.applications.CountWhere(ctx, []domainrepo.Filter{{Column: "candidate_id", Values: []any{old.ID}}})
		if err != nil {
			return err
		}
		if count != 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	if old == nil {
		rows, err := v.candidates.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "candidate_number", Values: []any{candidate.CandidateNumber}},
		}, 2, 0)
		if err != nil {
			return err
		}
		if len(rows) != 0 {
			return hrerrors.ErrAlreadyExists
		}
		if candidate.PersonPartyReferenceID == nil {
			if candidate.ContactReferenceID != nil {
				return hrerrors.ErrReferenceConflict
			}
			return nil
		}
		partyID := *candidate.PersonPartyReferenceID
		if err := verifyFoundationParty(ctx, v.foundation, partyID, true); err != nil {
			return err
		}
		if candidate.ContactReferenceID != nil {
			return v.foundation.VerifyPartyContact(ctx, *candidate.ContactReferenceID, partyID)
		}
	}
	return nil
}

func validateCandidateNotice(candidate *entity.Candidate) error {
	if candidate.NoticeAcknowledgedAt == nil || candidate.DataRetentionUntil == nil {
		return hrerrors.Invalid("privacy_notice_version", "notice acknowledgement and retention date are required")
	}
	if candidate.NoticeAcknowledgedAt.After(candidate.CreatedAt) ||
		entity.CompareDates(*candidate.DataRetentionUntil, *candidate.NoticeAcknowledgedAt) <= 0 ||
		entity.CompareDates(*candidate.DataRetentionUntil, candidate.CreatedAt) <= 0 {
		return hrerrors.Invalid("data_retention_until", "retention date must follow notice acknowledgement and creation")
	}
	return nil
}
