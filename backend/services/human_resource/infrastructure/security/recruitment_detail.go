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

type recruitmentInterviewLockReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Interview, error)
}
type interviewerScheduleReader interface {
	HasInterviewerScheduleOverlap(context.Context, string, string, time.Time, time.Time) (bool, error)
}

type recruitmentPanelistReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.InterviewPanelist, error)
}

type recruitmentFeedbackReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.InterviewFeedback, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type recruitmentOfferReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Offer, error)
}

type recruitmentAssessmentReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.ApplicationAssessment, error)
}

type recruitmentWorkerLookup interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Worker, error)
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type recruitmentJobLockReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Job, error)
}

type interviewPanelistValidator struct {
	schedules    interviewerScheduleReader
	interviews   recruitmentInterviewLockReader
	applications interviewApplicationReader
	workers      recruitmentWorkerLookup
	panelists    recruitmentPanelistReader
	feedback     recruitmentFeedbackReader
}

type interviewFeedbackValidator struct {
	interviews   recruitmentInterviewLockReader
	applications interviewApplicationReader
	workers      recruitmentWorkerLookup
	panelists    recruitmentPanelistReader
	feedback     recruitmentFeedbackReader
}

type offerValidator struct {
	applications interviewApplicationReader
	candidates   interviewCandidateReader
	requisitions interviewRequisitionReader
	jobs         recruitmentJobLockReader
	offers       recruitmentOfferReader
	foundation   FoundationVerifier
}

type applicationAssessmentValidator struct {
	applications interviewApplicationReader
	candidates   interviewCandidateReader
	requisitions interviewRequisitionReader
	workers      recruitmentWorkerLookup
	assessments  recruitmentAssessmentReader
}

// RecruitmentDetailValidators menjaga integritas panel wawancara, feedback, penawaran kerja, dan asesmen lamaran.
func RecruitmentDetailValidators(repos *repository.Registry, foundation FoundationVerifier) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Interview) != nil ||
		operation.RequireDependency(repos.Application) != nil || operation.RequireDependency(repos.Candidate) != nil ||
		operation.RequireDependency(repos.Requisition) != nil || operation.RequireDependency(repos.Job) != nil ||
		operation.RequireDependency(repos.Worker) != nil || operation.RequireDependency(repos.InterviewPanelist) != nil ||
		operation.RequireDependency(repos.InterviewFeedback) != nil || operation.RequireDependency(repos.Offer) != nil ||
		operation.RequireDependency(repos.ApplicationAssessment) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_interview_panelists": &interviewPanelistValidator{
			schedules:  repos.Interview,
			interviews: repos.Interview, applications: repos.Application, workers: repos.Worker,
			panelists: repos.InterviewPanelist, feedback: repos.InterviewFeedback,
		},
		"hr_interview_feedback": &interviewFeedbackValidator{
			interviews: repos.Interview, applications: repos.Application, workers: repos.Worker,
			panelists: repos.InterviewPanelist, feedback: repos.InterviewFeedback,
		},
		"hr_offers": &offerValidator{
			applications: repos.Application, candidates: repos.Candidate, requisitions: repos.Requisition,
			jobs: repos.Job, offers: repos.Offer, foundation: foundation,
		},
		"hr_application_assessments": &applicationAssessmentValidator{
			applications: repos.Application, candidates: repos.Candidate, requisitions: repos.Requisition,
			workers: repos.Worker, assessments: repos.ApplicationAssessment,
		},
	}, nil
}

func (v *interviewPanelistValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_interview_panelists" || operation.RequireDependency(v.interviews) != nil ||
		operation.RequireDependency(v.applications) != nil || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.panelists) != nil || operation.RequireDependency(v.feedback) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.InterviewPanelist
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.InterviewPanelist)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	panelist := previous
	if after != nil {
		var ok bool
		panelist, ok = after.(*entity.InterviewPanelist)
		if !ok || panelist == nil {
			return hrerrors.ErrInvalidInput
		}
		if panelist.ScorecardTemplateReferenceID != nil {
			return hrerrors.ErrUnavailable
		}
		if previous != nil && (previous.ID != panelist.ID || previous.InterviewID != panelist.InterviewID ||
			previous.InterviewerPrincipalID != panelist.InterviewerPrincipalID) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	interview, _, err := lockEditableInterviewAndApplication(ctx, v.interviews, v.applications, panelist.InterviewID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if interview.Status == "in_progress" ||
		(interview.ScheduledStart != nil && !interview.ScheduledStart.After(now)) {
		return hrerrors.ErrWorkflowRequired
	}
	if previous != nil {
		feedbackCount, err := v.feedback.CountWhere(ctx, []domainrepo.Filter{
			{Column: "interview_id", Values: []any{previous.InterviewID}},
			{Column: "interviewer_principal_id", Values: []any{previous.InterviewerPrincipalID}},
		})
		if err != nil {
			return err
		}
		if feedbackCount != 0 {
			return hrerrors.ErrReferenceConflict
		}
		if after == nil {
			return nil
		}
	}
	worker, err := activeRecruitmentPrincipalWorker(ctx, v.workers, panelist.InterviewerPrincipalID, now)
	if err != nil {
		return err
	}
	if interview.ScheduledStart != nil && interview.ScheduledEnd != nil {
		if worker.TerminationDate != nil && entity.CompareDates(interview.ScheduledEnd.UTC(), *worker.TerminationDate) > 0 {
			return hrerrors.ErrReferenceConflict
		}
		if operation.RequireDependency(v.schedules) != nil {
			return hrerrors.ErrUnavailable
		}
		overlaps, err := v.schedules.HasInterviewerScheduleOverlap(ctx, panelist.InterviewerPrincipalID, interview.ID,
			*interview.ScheduledStart, *interview.ScheduledEnd)
		if err != nil {
			return err
		}
		if overlaps {
			return hrerrors.ErrReferenceConflict
		}
	}
	if panelist.FeedbackDueAt != nil && interview.ScheduledEnd != nil && panelist.FeedbackDueAt.Before(*interview.ScheduledEnd) {
		return hrerrors.Invalid("feedback_due_at", "must not precede interview end")
	}
	matches, err := v.panelists.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "interview_id", Values: []any{panelist.InterviewID}},
		{Column: "interviewer_principal_id", Values: []any{panelist.InterviewerPrincipalID}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != panelist.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *interviewFeedbackValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_interview_feedback" || operation.RequireDependency(v.interviews) != nil ||
		operation.RequireDependency(v.applications) != nil || operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.panelists) != nil || operation.RequireDependency(v.feedback) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.InterviewFeedback
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.InterviewFeedback)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	feedback := previous
	if after != nil {
		var ok bool
		feedback, ok = after.(*entity.InterviewFeedback)
		if !ok || feedback == nil {
			return hrerrors.ErrInvalidInput
		}
		if strings.TrimSpace(feedback.OverallRecommendation) == "" {
			return hrerrors.Invalid("overall_recommendation", "must not be empty")
		}
		if !supportedFeedbackVisibilityClass(feedback.VisibilityClass) {
			return hrerrors.Invalid("visibility_class", "unsupported visibility class")
		}
		if previous != nil && (previous.ID != feedback.ID || previous.InterviewID != feedback.InterviewID ||
			previous.InterviewerPrincipalID != feedback.InterviewerPrincipalID ||
			previous.FeedbackVersion != feedback.FeedbackVersion ||
			!previous.SubmittedAt.Equal(feedback.SubmittedAt)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	interview, application, err := lockEditableInterviewAndApplication(ctx, v.interviews, v.applications, feedback.InterviewID)
	if err != nil {
		return err
	}
	if after != nil && (interview.Status != "in_progress" || application.Status != entity.ApplicationStatusInterview ||
		interview.ScheduledStart == nil || feedback.SubmittedAt.Before(*interview.ScheduledStart)) {
		return hrerrors.ErrWorkflowRequired
	}
	assigned, err := v.panelists.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "interview_id", Values: []any{feedback.InterviewID}},
		{Column: "interviewer_principal_id", Values: []any{feedback.InterviewerPrincipalID}},
	}, 2, 0)
	if err != nil {
		return err
	}
	if len(assigned) != 1 {
		return hrerrors.ErrReferenceConflict
	}
	filters := []domainrepo.Filter{
		{Column: "interview_id", Values: []any{feedback.InterviewID}},
		{Column: "interviewer_principal_id", Values: []any{feedback.InterviewerPrincipalID}},
	}
	versionCount, err := v.feedback.CountWhere(ctx, filters)
	if err != nil {
		return err
	}
	existingVersions, err := operation.CollectComplete(ctx, versionCount, func(limit, offset int) (operation.Page[entity.InterviewFeedback], error) {
		rows, err := v.feedback.SelectWhere(ctx, filters, limit, offset)
		return operation.Page[entity.InterviewFeedback]{Items: rows}, err
	})
	if err != nil {
		return err
	}
	for _, existing := range existingVersions {
		if existing.ID != feedback.ID && existing.FeedbackVersion == feedback.FeedbackVersion {
			return hrerrors.ErrAlreadyExists
		}
		if previous != nil && existing.FeedbackVersion > feedback.FeedbackVersion {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		return nil
	}
	_, err = activeRecruitmentPrincipalWorker(ctx, v.workers, feedback.InterviewerPrincipalID, time.Now())
	return err
}

func (v *offerValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_offers" || operation.RequireDependency(v.applications) != nil ||
		operation.RequireDependency(v.candidates) != nil || operation.RequireDependency(v.requisitions) != nil ||
		operation.RequireDependency(v.jobs) != nil || operation.RequireDependency(v.offers) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.Offer
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.Offer)
		if ok && previous != nil && previous.OfferStatus == "issued" {
			next, valid := after.(*entity.Offer)
			if valid && next != nil && (next.OfferStatus == "accepted" || next.OfferStatus == "declined") &&
				next.CandidateResponse != nil && next.RespondedAt != nil && next.CandidateResponseEvidenceID != nil {
				unchanged := *next
				unchanged.OfferStatus = previous.OfferStatus
				unchanged.CandidateResponse, unchanged.RespondedAt, unchanged.CandidateResponseEvidenceID = nil, nil, nil
				if reflect.DeepEqual(previous, &unchanged) {
					return nil
				}
			}
			if valid && next != nil && next.OfferStatus == "withdrawn" {
				unchanged := *next
				unchanged.OfferStatus = previous.OfferStatus
				unchanged.WithdrawnAt, unchanged.WithdrawnByPrincipalID, unchanged.WithdrawalReasonCode = nil, nil, nil
				if reflect.DeepEqual(previous, &unchanged) {
					return nil
				}
			}
		}
		if ok && previous != nil && previous.OfferStatus == "approved" {
			next, valid := after.(*entity.Offer)
			if valid && next != nil && next.OfferStatus == "issued" &&
				next.IssuedAt != nil && next.OfferDocumentReferenceID != nil &&
				next.OfferDocumentVersionReferenceID != nil && len(next.OfferContentDigest) == 32 {
				unchanged := *next
				unchanged.OfferStatus = previous.OfferStatus
				unchanged.IssuedAt, unchanged.OfferDocumentReferenceID, unchanged.OfferDocumentVersionReferenceID, unchanged.OfferContentDigest = nil, nil, nil, nil
				if reflect.DeepEqual(previous, &unchanged) {
					now := time.Now().UTC()
					application, _, _, err := lockRecruitmentApplicationContext(ctx, v.applications, v.candidates, v.requisitions, next.ApplicationID, now)
					if err != nil {
						return err
					}
					if application.Status != entity.ApplicationStatusOffer {
						return hrerrors.ErrWorkflowRequired
					}
					return v.validateOfferJobAndEmployer(ctx, next, now)
				}
			}
		}
		if !ok || previous == nil || !draftUnsignedOffer(previous) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	offer := previous
	if after != nil {
		var ok bool
		offer, ok = after.(*entity.Offer)
		if !ok || offer == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous != nil && offer.OfferStatus == "approved" {
			if offer.ApprovedByPrincipalID == nil || offer.ApprovedAt == nil ||
				offer.ProposedCompensationDocumentVersionReferenceID == nil {
				return hrerrors.ErrWorkflowRequired
			}
			unchanged := *offer
			unchanged.OfferStatus = previous.OfferStatus
			unchanged.ApprovedByPrincipalID, unchanged.ApprovedAt, unchanged.ProposedCompensationDocumentVersionReferenceID = nil, nil, nil
			if !reflect.DeepEqual(previous, &unchanged) {
				return hrerrors.ErrWorkflowRequired
			}
			offer = &unchanged
		}
		if offer.ProposedCompensationDocumentVersionReferenceID != nil || offer.OfferDocumentReferenceID != nil ||
			offer.OfferDocumentVersionReferenceID != nil || len(offer.OfferContentDigest) != 0 {
			return hrerrors.ErrUnavailable
		}
		if !draftUnsignedOffer(offer) {
			return hrerrors.ErrWorkflowRequired
		}
		if !supportedContractType(offer.ProposedContractType) {
			return hrerrors.Invalid("proposed_contract_type", "unsupported contract type")
		}
		if offer.CurrencyCode != nil && !validISO3UpperCode(*offer.CurrencyCode) {
			return hrerrors.Invalid("currency_code", "must be a 3-letter uppercase code")
		}
		now := time.Now()
		if offer.ExpiresAt != nil && !offer.ExpiresAt.After(now) {
			return hrerrors.Invalid("expires_at", "must be in the future")
		}
		if offer.ProposedStartDate != nil && entity.CompareDates(*offer.ProposedStartDate, now) < 0 {
			return hrerrors.ErrReferenceConflict
		}
		if previous == nil && (offer.CreatedByPrincipalID == nil || offer.CreatedAt == nil) {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != offer.ID || previous.ApplicationID != offer.ApplicationID ||
			previous.OfferNumber != offer.OfferNumber || previous.OfferVersion != offer.OfferVersion ||
			!reflect.DeepEqual(previous.CreatedByPrincipalID, offer.CreatedByPrincipalID) || !reflect.DeepEqual(previous.CreatedAt, offer.CreatedAt)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	now := time.Now()
	application, _, _, err := lockRecruitmentApplicationContext(ctx, v.applications, v.candidates, v.requisitions, offer.ApplicationID, now)
	if err != nil {
		return err
	}
	if after == nil {
		return nil
	}
	switch application.Status {
	case entity.ApplicationStatusShortlisted, entity.ApplicationStatusInterview,
		entity.ApplicationStatusAssessment, entity.ApplicationStatusOffer:
	default:
		return hrerrors.ErrWorkflowRequired
	}
	if err := v.validateOfferJobAndEmployer(ctx, offer, now); err != nil {
		return err
	}
	numberMatches, err := v.offers.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "offer_number", Values: []any{offer.OfferNumber}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range numberMatches {
		if existing.ID != offer.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	versionMatches, err := v.offers.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "application_id", Values: []any{offer.ApplicationID}},
		{Column: "offer_version", Values: []any{offer.OfferVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range versionMatches {
		if existing.ID != offer.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *offerValidator) validateOfferJobAndEmployer(ctx context.Context, offer *entity.Offer, now time.Time) error {
	job, err := v.jobs.FindByIDForUpdate(ctx, offer.ProposedJobID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if job == nil || !job.IsActive || entity.CompareDates(now, job.ValidFrom) < 0 ||
		(job.ValidTo != nil && (entity.CompareDates(now, *job.ValidTo) > 0 ||
			offer.ProposedStartDate != nil && entity.CompareDates(*offer.ProposedStartDate, *job.ValidTo) > 0)) {
		return hrerrors.ErrReferenceConflict
	}
	return verifyFoundationLegalEntity(ctx, v.foundation, offer.ProposedLegalEntityReferenceID)
}

func (v *applicationAssessmentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_application_assessments" || operation.RequireDependency(v.applications) != nil ||
		operation.RequireDependency(v.candidates) != nil || operation.RequireDependency(v.requisitions) != nil ||
		operation.RequireDependency(v.workers) != nil || operation.RequireDependency(v.assessments) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.ApplicationAssessment
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.ApplicationAssessment)
		if !ok || previous == nil || (previous.Status != "planned" && previous.Status != "in_progress") || previous.CompletedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	assessment := previous
	completing := false
	if after != nil {
		var ok bool
		assessment, ok = after.(*entity.ApplicationAssessment)
		if !ok || assessment == nil {
			return hrerrors.ErrInvalidInput
		}
		completing = previous != nil && validApplicationAssessmentComplete(previous, assessment)
		if completing {
			if len(assessment.EvidenceDigest) != 0 && len(assessment.EvidenceDigest) != 32 {
				return hrerrors.Invalid("evidence_digest", "must be a 32-byte digest")
			}
		} else {
			if assessment.EvidenceDocumentVersionReferenceID != nil || len(assessment.EvidenceDigest) != 0 {
				return hrerrors.ErrUnavailable
			}
			if assessment.Status != "planned" || assessment.CompletedAt != nil ||
				assessment.NormalizedScore != nil || assessment.RecommendationCode != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if previous != nil && (previous.ID != assessment.ID || previous.ApplicationID != assessment.ApplicationID ||
				previous.AssessmentCode != assessment.AssessmentCode || previous.AttemptNumber != assessment.AttemptNumber) {
				return hrerrors.ErrWorkflowRequired
			}
		}
	} else if previous.Status != "planned" {
		return hrerrors.ErrWorkflowRequired
	}
	now := time.Now()
	application, _, _, err := lockRecruitmentApplicationContext(ctx, v.applications, v.candidates, v.requisitions, assessment.ApplicationID, now)
	if err != nil {
		return err
	}
	if after == nil {
		return nil
	}
	switch application.Status {
	case entity.ApplicationStatusScreening, entity.ApplicationStatusShortlisted,
		entity.ApplicationStatusInterview, entity.ApplicationStatusAssessment:
	default:
		return hrerrors.ErrWorkflowRequired
	}
	if assessment.ScheduledAt != nil && assessment.ScheduledAt.Before(application.AppliedAt) {
		return hrerrors.Invalid("scheduled_at", "must not precede application submission")
	}
	if completing && assessment.CompletedAt.Before(application.AppliedAt) {
		return hrerrors.ErrWorkflowRequired
	}
	if assessment.AssessorPrincipalID != nil {
		if _, err := activeRecruitmentPrincipalWorker(ctx, v.workers, *assessment.AssessorPrincipalID, now); err != nil {
			return err
		}
	}
	if completing {
		return nil
	}
	matches, err := v.assessments.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "application_id", Values: []any{assessment.ApplicationID}},
		{Column: "assessment_code", Values: []any{assessment.AssessmentCode}},
		{Column: "attempt_number", Values: []any{assessment.AttemptNumber}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != assessment.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func lockEditableInterviewAndApplication(
	ctx context.Context,
	interviews recruitmentInterviewLockReader,
	applications interviewApplicationReader,
	interviewID string,
) (*entity.Interview, *entity.Application, error) {
	interview, err := interviews.FindByIDForUpdate(ctx, interviewID)
	if err != nil {
		return nil, nil, workerAssessmentReferenceError(err)
	}
	if interview == nil {
		return nil, nil, hrerrors.ErrReferenceConflict
	}
	if interview.Status != "planned" && interview.Status != "scheduled" && interview.Status != "in_progress" {
		return nil, nil, hrerrors.ErrWorkflowRequired
	}
	application, err := applications.FindByIDForUpdate(ctx, interview.ApplicationID)
	if err != nil {
		return nil, nil, workerAssessmentReferenceError(err)
	}
	if application == nil {
		return nil, nil, hrerrors.ErrReferenceConflict
	}
	switch application.Status {
	case entity.ApplicationStatusRejected, entity.ApplicationStatusWithdrawn, entity.ApplicationStatusHired:
		return nil, nil, hrerrors.ErrWorkflowRequired
	}
	return interview, application, nil
}

func lockRecruitmentApplicationContext(
	ctx context.Context,
	applications interviewApplicationReader,
	candidates interviewCandidateReader,
	requisitions interviewRequisitionReader,
	applicationID string,
	now time.Time,
) (*entity.Application, *entity.Candidate, *entity.Requisition, error) {
	application, err := applications.FindByIDForUpdate(ctx, applicationID)
	if err != nil {
		return nil, nil, nil, workerAssessmentReferenceError(err)
	}
	if application == nil {
		return nil, nil, nil, hrerrors.ErrReferenceConflict
	}
	candidate, err := candidates.FindByIDForUpdate(ctx, application.CandidateID)
	if err != nil {
		return nil, nil, nil, workerAssessmentReferenceError(err)
	}
	if candidate == nil || candidate.CandidateStatus != "active" || candidate.NoticeAcknowledgedAt == nil ||
		candidate.NoticeAcknowledgedAt.After(now) || candidate.DataRetentionUntil == nil ||
		entity.CompareDates(*candidate.DataRetentionUntil, now) <= 0 {
		return nil, nil, nil, hrerrors.ErrReferenceConflict
	}
	requisition, err := requisitions.FindByIDForUpdate(ctx, application.RequisitionID)
	if err != nil {
		return nil, nil, nil, workerAssessmentReferenceError(err)
	}
	if !recruitmentRequisitionOpenAt(requisition, now) {
		return nil, nil, nil, hrerrors.ErrReferenceConflict
	}
	return application, candidate, requisition, nil
}

func recruitmentRequisitionOpenAt(requisition *entity.Requisition, now time.Time) bool {
	return requisition != nil && requisition.Status == "open" && requisition.ApprovedAt != nil &&
		requisition.ApprovedByPrincipalID != nil && requisition.OpenedAt != nil &&
		!requisition.ApprovedAt.After(*requisition.OpenedAt) && !requisition.OpenedAt.After(now) &&
		requisition.ClosedAt == nil
}

func activeRecruitmentPrincipalWorker(ctx context.Context, workers recruitmentWorkerLookup, principalID string, now time.Time) (*entity.Worker, error) {
	rows, err := workers.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "iam_principal_reference_id", Values: []any{principalID}},
	}, 2, 0)
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, hrerrors.ErrReferenceConflict
	}
	worker, err := workers.FindByIDForUpdate(ctx, rows[0].ID)
	if err != nil {
		return nil, workerAssessmentReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive || worker.IAMPrincipalReferenceID == nil ||
		!strings.EqualFold(*worker.IAMPrincipalReferenceID, principalID) ||
		(worker.HireDate != nil && entity.CompareDates(*worker.HireDate, now) > 0) ||
		(worker.TerminationDate != nil && entity.CompareDates(*worker.TerminationDate, now) <= 0) {
		return nil, hrerrors.ErrReferenceConflict
	}
	return worker, nil
}

func validApplicationAssessmentComplete(before, after *entity.ApplicationAssessment) bool {
	if before == nil || after == nil || (before.Status != "planned" && before.Status != "in_progress") ||
		after.Status != "completed" || before.CompletedAt != nil || after.CompletedAt == nil ||
		after.AssessorPrincipalID == nil || (after.EvidenceDocumentVersionReferenceID == nil) != (len(after.EvidenceDigest) == 0) {
		return false
	}
	if before.AssessorPrincipalID != nil && !strings.EqualFold(*before.AssessorPrincipalID, *after.AssessorPrincipalID) {
		return false
	}
	unchanged := *after
	unchanged.Status = before.Status
	unchanged.AssessorPrincipalID = before.AssessorPrincipalID
	unchanged.CompletedAt = before.CompletedAt
	unchanged.NormalizedScore = before.NormalizedScore
	unchanged.RecommendationCode = before.RecommendationCode
	unchanged.EvidenceDocumentVersionReferenceID = before.EvidenceDocumentVersionReferenceID
	unchanged.EvidenceDigest = before.EvidenceDigest
	return reflect.DeepEqual(*before, unchanged)
}

func draftUnsignedOffer(offer *entity.Offer) bool {
	return offer != nil && offer.OfferStatus == "draft" && offer.ApprovedByPrincipalID == nil &&
		offer.ApprovedAt == nil && offer.IssuedAt == nil && offer.CandidateResponse == nil && offer.RespondedAt == nil
}

func supportedFeedbackVisibilityClass(value string) bool {
	switch value {
	case "restricted", "confidential", "panel_shared", "hiring_team":
		return true
	default:
		return false
	}
}

func validISO3UpperCode(value string) bool {
	if len(value) != 3 {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < 'A' || value[i] > 'Z' {
			return false
		}
	}
	return true
}
