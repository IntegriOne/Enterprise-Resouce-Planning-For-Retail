package security

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type interviewApplicationReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Application, error)
}

type interviewCandidateReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Candidate, error)
}

type interviewRequisitionReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Requisition, error)
}

type interviewReader interface {
	HasCandidateScheduleOverlap(context.Context, string, string, time.Time, time.Time) (bool, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Interview, error)
}

type interviewChildReader[T any] interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]T, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type interviewValidator struct {
	applications interviewApplicationReader
	candidates   interviewCandidateReader
	requisitions interviewRequisitionReader
	interviews   interviewReader
	panelists    interviewChildReader[entity.InterviewPanelist]
	feedback     interviewChildReader[entity.InterviewFeedback]
}

// RecruitmentInterviewValidators membuka draf wawancara pada lamaran yang masih sah.
func RecruitmentInterviewValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Application) != nil ||
		operation.RequireDependency(repos.Candidate) != nil || operation.RequireDependency(repos.Requisition) != nil ||
		operation.RequireDependency(repos.Interview) != nil || operation.RequireDependency(repos.InterviewPanelist) != nil ||
		operation.RequireDependency(repos.InterviewFeedback) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_interviews": &interviewValidator{
			applications: repos.Application, candidates: repos.Candidate, requisitions: repos.Requisition,
			interviews: repos.Interview, panelists: repos.InterviewPanelist, feedback: repos.InterviewFeedback,
		},
	}, nil
}

func (v *interviewValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_interviews" || operation.RequireDependency(v.applications) != nil ||
		operation.RequireDependency(v.candidates) != nil || operation.RequireDependency(v.requisitions) != nil ||
		operation.RequireDependency(v.interviews) != nil || operation.RequireDependency(v.panelists) != nil ||
		operation.RequireDependency(v.feedback) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.Interview
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.Interview)
		if !ok || previous == nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous != nil {
		if next, ok := after.(*entity.Interview); ok && next != nil && next.Status == "cancelled" {
			if previous.Status != "planned" && previous.Status != "scheduled" {
				return hrerrors.ErrWorkflowRequired
			}
			unchanged := *next
			unchanged.Status = previous.Status
			if !reflect.DeepEqual(previous, &unchanged) {
				return hrerrors.ErrWorkflowRequired
			}
			count, err := v.feedback.CountWhere(ctx, []domainrepo.Filter{{Column: "interview_id", Values: []any{previous.ID}}})
			if err != nil {
				return err
			}
			if count != 0 {
				return hrerrors.ErrReferenceConflict
			}
			return nil
		}
		if next, ok := after.(*entity.Interview); ok && next != nil && next.Status == "in_progress" {
			if (previous.Status != "planned" && previous.Status != "scheduled") ||
				previous.ScheduledStart == nil || previous.ScheduledEnd == nil {
				return hrerrors.ErrWorkflowRequired
			}
			unchanged := *next
			unchanged.Status = previous.Status
			if !reflect.DeepEqual(previous, &unchanged) {
				return hrerrors.ErrWorkflowRequired
			}
			application, _, _, err := lockRecruitmentApplicationContext(ctx, v.applications, v.candidates, v.requisitions, previous.ApplicationID, time.Now().UTC())
			if err != nil {
				return err
			}
			if application.Status != entity.ApplicationStatusInterview {
				return hrerrors.ErrWorkflowRequired
			}
			count, err := v.panelists.CountWhere(ctx, []domainrepo.Filter{{Column: "interview_id", Values: []any{previous.ID}}})
			if err != nil {
				return err
			}
			if count == 0 {
				return hrerrors.ErrReferenceConflict
			}
			return nil
		}
		if next, ok := after.(*entity.Interview); ok && next != nil && next.Status == "completed" {
			if previous.Status != "in_progress" || previous.ScheduledStart == nil {
				return hrerrors.ErrWorkflowRequired
			}
			unchanged := *next
			unchanged.Status = previous.Status
			if !reflect.DeepEqual(previous, &unchanged) {
				return hrerrors.ErrWorkflowRequired
			}
			return v.validateCompletion(ctx, previous)
		}
		if previous.Status != "planned" {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	interview := previous
	if after != nil {
		var ok bool
		interview, ok = after.(*entity.Interview)
		if !ok || interview == nil {
			return hrerrors.ErrInvalidInput
		}
		if interview.LocationReferenceID != nil || interview.MeetingReferenceID != nil {
			return hrerrors.ErrUnavailable
		}
		if interview.Status != "planned" || interview.CreatedByPrincipalID == nil || (previous != nil &&
			(previous.ID != interview.ID || previous.ApplicationID != interview.ApplicationID ||
				previous.InterviewRound != interview.InterviewRound || previous.InterviewType != interview.InterviewType ||
				!reflect.DeepEqual(previous.CreatedByPrincipalID, interview.CreatedByPrincipalID))) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	application, err := v.applications.FindByIDForUpdate(ctx, interview.ApplicationID)
	if err != nil {
		return err
	}
	if application == nil {
		return hrerrors.ErrReferenceConflict
	}
	if previous != nil {
		filters := []domainrepo.Filter{{Column: "interview_id", Values: []any{previous.ID}}}
		panelistCount, err := v.panelists.CountWhere(ctx, filters)
		if err != nil {
			return err
		}
		feedbackCount, err := v.feedback.CountWhere(ctx, filters)
		if err != nil {
			return err
		}
		if panelistCount != 0 || feedbackCount != 0 {
			return hrerrors.ErrReferenceConflict
		}
		if after == nil {
			return nil
		}
	}
	switch application.Status {
	case entity.ApplicationStatusReceived, entity.ApplicationStatusScreening,
		entity.ApplicationStatusShortlisted, entity.ApplicationStatusInterview:
	default:
		return hrerrors.ErrWorkflowRequired
	}
	now := time.Now()
	candidate, err := v.candidates.FindByIDForUpdate(ctx, application.CandidateID)
	if err != nil {
		return err
	}
	if candidate == nil || candidate.CandidateStatus != "active" || candidate.NoticeAcknowledgedAt == nil ||
		candidate.NoticeAcknowledgedAt.After(now) || candidate.DataRetentionUntil == nil ||
		entity.CompareDates(*candidate.DataRetentionUntil, now) <= 0 {
		return hrerrors.ErrReferenceConflict
	}
	requisition, err := v.requisitions.FindByIDForUpdate(ctx, application.RequisitionID)
	if err != nil {
		return err
	}
	if !recruitmentRequisitionOpenAt(requisition, now) {
		return hrerrors.ErrReferenceConflict
	}
	if (interview.ScheduledStart == nil) != (interview.ScheduledEnd == nil) {
		return hrerrors.Invalid("scheduled_start", "start and end must both be set")
	}
	if interview.ScheduledStart != nil {
		if interview.CandidateTimezone == nil || *interview.CandidateTimezone == "" || *interview.CandidateTimezone == "Local" {
			return hrerrors.Invalid("candidate_timezone", "an IANA timezone is required for scheduled interviews")
		}
		candidateLocation, err := time.LoadLocation(*interview.CandidateTimezone)
		if err != nil {
			return hrerrors.Invalid("candidate_timezone", "must be a valid IANA timezone")
		}
		if !interview.ScheduledStart.After(now) || !interview.ScheduledEnd.After(*interview.ScheduledStart) ||
			entity.CompareDates(*candidate.DataRetentionUntil, interview.ScheduledEnd.In(candidateLocation)) <= 0 {
			return hrerrors.Invalid("scheduled_start", "schedule must be future and within candidate retention")
		}
	}
	if previous == nil {
		matches, err := v.interviews.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "application_id", Values: []any{interview.ApplicationID}},
			{Column: "interview_round", Values: []any{interview.InterviewRound}},
			{Column: "interview_type", Values: []any{interview.InterviewType}},
		}, 1, 0)
		if err != nil {
			return err
		}
		if len(matches) != 0 {
			return hrerrors.ErrAlreadyExists
		}
	}
	if interview.ScheduledStart != nil {
		// Lock kandidat menjaga konsistensi jadwal antar-lamaran.
		overlaps, err := v.interviews.HasCandidateScheduleOverlap(ctx, candidate.ID, interview.ID, *interview.ScheduledStart, *interview.ScheduledEnd)
		if err != nil {
			return err
		}
		if overlaps {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}

// validateCompletion memeriksa versi feedback terakhir setiap panelis saat wawancara terkunci.
func (v *interviewValidator) validateCompletion(ctx context.Context, interview *entity.Interview) error {
	filters := []domainrepo.Filter{{Column: "interview_id", Values: []any{interview.ID}}}
	panelists, err := collectInterviewChildren(ctx, v.panelists, filters)
	if err != nil {
		return err
	}
	if len(panelists) == 0 {
		return hrerrors.ErrReferenceConflict
	}
	feedback, err := collectInterviewChildren(ctx, v.feedback, filters)
	if err != nil {
		return err
	}
	latest := make(map[string]entity.InterviewFeedback, len(panelists))
	for _, entry := range feedback {
		principal := strings.ToLower(entry.InterviewerPrincipalID)
		if entry.FeedbackVersion > latest[principal].FeedbackVersion {
			latest[principal] = entry
		}
	}
	now := time.Now().UTC()
	for _, panelist := range panelists {
		entry, found := latest[strings.ToLower(panelist.InterviewerPrincipalID)]
		if !found || entry.SubmittedAt.Before(*interview.ScheduledStart) || entry.SubmittedAt.After(now) ||
			strings.TrimSpace(entry.OverallRecommendation) == "" || !supportedFeedbackVisibilityClass(entry.VisibilityClass) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	return nil
}

func collectInterviewChildren[T any](ctx context.Context, reader interviewChildReader[T], filters []domainrepo.Filter) ([]T, error) {
	count, err := reader.CountWhere(ctx, filters)
	if err != nil {
		return nil, err
	}
	return operation.CollectComplete(ctx, count, func(limit, offset int) (operation.Page[T], error) {
		items, err := reader.SelectWhere(ctx, filters, limit, offset)
		return operation.Page[T]{Items: items}, err
	})
}
