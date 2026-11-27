package security

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	talentdomain "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/talent"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type talentCycleStore interface {
	FindByIDForUpdate(context.Context, string) (*entity.TalentReviewCycle, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.TalentReviewCycle, error)
}

type talentAssessmentStore interface {
	FindByIDForUpdate(context.Context, string) (*entity.TalentAssessment, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.TalentAssessment, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type talentScoreStore interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.TalentAssessmentScore, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type talentSessionStore interface {
	FindByIDForUpdate(context.Context, string) (*entity.TalentCalibrationSession, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.TalentCalibrationSession, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type talentDecisionStore interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.TalentCalibrationDecision, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type talentPoolStore interface {
	FindByIDForUpdate(context.Context, string) (*entity.TalentPool, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.TalentPool, error)
}

type talentMembershipStore interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.TalentPoolMembership, error)
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type talentWorkerLookup interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Worker, error)
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type talentPerformanceReviewReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.PerformanceReview, error)
}

type talentDimensionLocker interface {
	FindByIDForUpdate(context.Context, string) (*entity.TalentDimension, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.TalentDimension, error)
}

type talentFrameworkValidator struct{}

type talentReviewCycleValidator struct {
	frameworks  talentFrameworkReader
	dimensions  talentDimensionReader
	cycles      talentCycleStore
	assessments talentAssessmentStore
	sessions    talentSessionStore
	workers     talentWorkerLookup
	foundation  FoundationVerifier
}

type talentAssessmentValidator struct {
	frameworks  talentFrameworkReader
	dimensions  talentDimensionReader
	cycles      talentCycleStore
	assessments talentAssessmentStore
	scores      talentScoreStore
	workers     talentWorkerLookup
	reviews     talentPerformanceReviewReader
}

type talentAssessmentScoreValidator struct {
	assessments talentAssessmentStore
	dimensions  talentDimensionLocker
	cycles      talentCycleStore
	scores      talentScoreStore
	decisions   talentDecisionStore
}

type talentCalibrationSessionValidator struct {
	cycles    talentCycleStore
	sessions  talentSessionStore
	decisions talentDecisionStore
	workers   talentWorkerLookup
}

type talentCalibrationDecisionValidator struct {
	sessions    talentSessionStore
	cycles      talentCycleStore
	assessments talentAssessmentStore
	scores      talentScoreStore
	decisions   talentDecisionStore
	workers     talentWorkerLookup
}

type talentPoolValidator struct {
	pools       talentPoolStore
	memberships talentMembershipStore
	jobs        recruitmentJobLockReader
	workers     talentWorkerLookup
	foundation  FoundationVerifier
}

type talentPoolMembershipValidator struct {
	pools       talentPoolStore
	workers     talentWorkerLookup
	assessments talentAssessmentStore
	memberships talentMembershipStore
}

// TalentValidators menjaga integritas framework, siklus review, asesmen, kalibrasi, dan talent pool.
func TalentValidators(repos *repository.Registry, foundation FoundationVerifier) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.TalentFramework) != nil ||
		operation.RequireDependency(repos.TalentDimension) != nil || operation.RequireDependency(repos.TalentReviewCycle) != nil ||
		operation.RequireDependency(repos.TalentAssessment) != nil || operation.RequireDependency(repos.TalentAssessmentScore) != nil ||
		operation.RequireDependency(repos.TalentCalibrationSession) != nil || operation.RequireDependency(repos.TalentCalibrationDecision) != nil ||
		operation.RequireDependency(repos.TalentPool) != nil || operation.RequireDependency(repos.TalentPoolMembership) != nil ||
		operation.RequireDependency(repos.Worker) != nil || operation.RequireDependency(repos.PerformanceReview) != nil ||
		operation.RequireDependency(repos.Job) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_talent_frameworks": &talentFrameworkValidator{},
		"hr_talent_review_cycles": &talentReviewCycleValidator{
			frameworks: repos.TalentFramework, dimensions: repos.TalentDimension,
			cycles: repos.TalentReviewCycle, assessments: repos.TalentAssessment,
			sessions: repos.TalentCalibrationSession, workers: repos.Worker, foundation: foundation,
		},
		"hr_talent_assessments": &talentAssessmentValidator{
			frameworks: repos.TalentFramework, dimensions: repos.TalentDimension,
			cycles: repos.TalentReviewCycle, assessments: repos.TalentAssessment,
			scores: repos.TalentAssessmentScore, workers: repos.Worker, reviews: repos.PerformanceReview,
		},
		"hr_talent_assessment_scores": &talentAssessmentScoreValidator{
			assessments: repos.TalentAssessment, dimensions: repos.TalentDimension,
			cycles: repos.TalentReviewCycle, scores: repos.TalentAssessmentScore,
			decisions: repos.TalentCalibrationDecision,
		},
		"hr_talent_calibration_sessions": &talentCalibrationSessionValidator{
			cycles: repos.TalentReviewCycle, sessions: repos.TalentCalibrationSession,
			decisions: repos.TalentCalibrationDecision, workers: repos.Worker,
		},
		"hr_talent_calibration_decisions": &talentCalibrationDecisionValidator{
			sessions: repos.TalentCalibrationSession, cycles: repos.TalentReviewCycle,
			assessments: repos.TalentAssessment, scores: repos.TalentAssessmentScore,
			decisions: repos.TalentCalibrationDecision, workers: repos.Worker,
		},
		"hr_talent_pools": &talentPoolValidator{
			pools: repos.TalentPool, memberships: repos.TalentPoolMembership,
			jobs: repos.Job, workers: repos.Worker, foundation: foundation,
		},
		"hr_talent_pool_memberships": &talentPoolMembershipValidator{
			pools: repos.TalentPool, workers: repos.Worker,
			assessments: repos.TalentAssessment, memberships: repos.TalentPoolMembership,
		},
	}, nil
}

func (v *talentFrameworkValidator) ValidateMutation(ctx context.Context, table string, _, _ entity.Validatable) error {
	if v == nil || table != "hr_talent_frameworks" {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Framework talenta bersifat immutable dan memerlukan otorisasi DRM tersendiri.
	return hrerrors.ErrWorkflowRequired
}

func (v *talentReviewCycleValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_talent_review_cycles" || operation.RequireDependency(v.frameworks) != nil ||
		operation.RequireDependency(v.dimensions) != nil || operation.RequireDependency(v.cycles) != nil ||
		operation.RequireDependency(v.assessments) != nil || operation.RequireDependency(v.sessions) != nil ||
		operation.RequireDependency(v.workers) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.TalentReviewCycle
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.TalentReviewCycle)
		if !ok || previous == nil || (previous.Status != "draft" && previous.Status != "open" && previous.Status != "calibration") {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if after == nil {
		if previous.Status != "draft" {
			return hrerrors.ErrWorkflowRequired
		}
		assessmentCount, err := v.assessments.CountWhere(ctx, []domainrepo.Filter{{Column: "talent_review_cycle_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		sessionCount, err := v.sessions.CountWhere(ctx, []domainrepo.Filter{{Column: "talent_review_cycle_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if assessmentCount != 0 || sessionCount != 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	cycle, ok := after.(*entity.TalentReviewCycle)
	if !ok || cycle == nil {
		return hrerrors.ErrInvalidInput
	}
	launching := previous != nil && previous.Status == "draft" && cycle.Status == "open"
	calibrating := previous != nil && previous.Status == "open" && cycle.Status == "calibration"
	closing := previous != nil && (previous.Status == "open" || previous.Status == "calibration") && cycle.Status == "closed"
	if launching || calibrating || closing {
		unchanged := *cycle
		unchanged.Status = previous.Status
		if closing {
			if cycle.ClosedAt == nil {
				return hrerrors.ErrWorkflowRequired
			}
			unchanged.ClosedAt = previous.ClosedAt
		}
		if !reflect.DeepEqual(*previous, unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
		if closing {
			pendingAssessments, err := v.assessments.CountWhere(ctx, []domainrepo.Filter{
				{Column: "talent_review_cycle_id", Values: []any{cycle.ID}},
				{Column: "status", Values: []any{"draft", "submitted"}},
			})
			if err != nil {
				return err
			}
			openSessions, err := v.sessions.CountWhere(ctx, []domainrepo.Filter{
				{Column: "talent_review_cycle_id", Values: []any{cycle.ID}},
				{Column: "status", Values: []any{"planned", "open"}},
			})
			if err != nil {
				return err
			}
			if pendingAssessments != 0 || openSessions != 0 {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		}
	} else {
		if (previous != nil && previous.Status != "draft") || cycle.Status != "draft" || cycle.ClosedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != cycle.ID || previous.CycleCode != cycle.CycleCode || previous.FrameworkID != cycle.FrameworkID) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	framework, err := v.frameworks.FindByIDForUpdate(ctx, cycle.FrameworkID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if framework == nil || entity.CompareDates(cycle.PeriodStart, framework.EffectiveFrom) < 0 ||
		(framework.EffectiveTo != nil && entity.CompareDates(cycle.PeriodEnd, *framework.EffectiveTo) > 0) {
		return hrerrors.ErrReferenceConflict
	}
	if err := talentFrameworkDimensionsComplete(ctx, v.dimensions, framework.ID); err != nil {
		return err
	}
	if err := verifyFoundationLegalEntity(ctx, v.foundation, cycle.LegalEntityReferenceID); err != nil {
		return err
	}
	if cycle.OrganizationUnitReferenceID != nil {
		if err := verifyFoundationOrganizationUnit(ctx, v.foundation, *cycle.OrganizationUnitReferenceID, cycle.LegalEntityReferenceID); err != nil {
			return err
		}
	}
	if _, err := activeRecruitmentPrincipalWorker(ctx, v.workers, cycle.OwnerPrincipalID, time.Now()); err != nil {
		return err
	}
	matches, err := v.cycles.SelectWhere(ctx, []domainrepo.Filter{{Column: "cycle_code", Values: []any{cycle.CycleCode}}}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != cycle.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *talentAssessmentValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_talent_assessments" || operation.RequireDependency(v.frameworks) != nil ||
		operation.RequireDependency(v.dimensions) != nil || operation.RequireDependency(v.cycles) != nil ||
		operation.RequireDependency(v.assessments) != nil || operation.RequireDependency(v.scores) != nil ||
		operation.RequireDependency(v.workers) != nil || operation.RequireDependency(v.reviews) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.TalentAssessment
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.TalentAssessment)
		if !ok || previous == nil || (previous.Status != "draft" && previous.Status != "submitted") {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if after == nil {
		if previous.Status != "draft" {
			return hrerrors.ErrWorkflowRequired
		}
		scoreCount, err := v.scores.CountWhere(ctx, []domainrepo.Filter{{Column: "assessment_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if scoreCount != 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	assessment, ok := after.(*entity.TalentAssessment)
	if !ok || assessment == nil {
		return hrerrors.ErrInvalidInput
	}
	submitting := previous != nil && previous.Status == "draft" && assessment.Status == "submitted"
	withdrawing := previous != nil && (previous.Status == "draft" || previous.Status == "submitted") && assessment.Status == "withdrawn"
	finalizing := previous != nil && (previous.Status == "draft" || previous.Status == "submitted") && assessment.Status == "finalized"
	if submitting || withdrawing || finalizing {
		unchanged := *assessment
		unchanged.Status = previous.Status
		if finalizing {
			if assessment.FinalizedAt == nil || assessment.FinalizedByPrincipalID == nil || assessment.FinalizedAt.Before(assessment.AssessedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			unchanged.FinalizedAt = previous.FinalizedAt
			unchanged.FinalizedByPrincipalID = previous.FinalizedByPrincipalID
		} else if assessment.FinalizedAt != nil || assessment.FinalizedByPrincipalID != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if !reflect.DeepEqual(*previous, unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
	} else {
		if (previous != nil && previous.Status != "draft") || assessment.Status != "draft" ||
			assessment.FinalizedAt != nil || assessment.FinalizedByPrincipalID != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != assessment.ID || previous.WorkerID != assessment.WorkerID ||
			previous.TalentReviewCycleID != assessment.TalentReviewCycleID ||
			previous.FrameworkID != assessment.FrameworkID || previous.AssessmentVersion != assessment.AssessmentVersion ||
			!strings.EqualFold(previous.AssessedByPrincipalID, assessment.AssessedByPrincipalID) ||
			!previous.AssessedAt.Equal(assessment.AssessedAt)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	cycle, err := v.cycles.FindByIDForUpdate(ctx, assessment.TalentReviewCycleID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if cycle == nil || (cycle.Status != "open" && cycle.Status != "calibration") ||
		!strings.EqualFold(cycle.FrameworkID, assessment.FrameworkID) ||
		entity.CompareDates(assessment.AssessedAt, cycle.PeriodStart) < 0 ||
		entity.CompareDates(assessment.AssessedAt, cycle.PeriodEnd) > 0 {
		return hrerrors.ErrReferenceConflict
	}
	framework, err := v.frameworks.FindByIDForUpdate(ctx, assessment.FrameworkID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if framework == nil || entity.CompareDates(assessment.AssessedAt, framework.EffectiveFrom) < 0 ||
		(framework.EffectiveTo != nil && entity.CompareDates(assessment.AssessedAt, *framework.EffectiveTo) > 0) {
		return hrerrors.ErrReferenceConflict
	}
	if withdrawing {
		return nil
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, assessment.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive ||
		(worker.TerminationDate != nil && entity.CompareDates(assessment.AssessedAt, *worker.TerminationDate) > 0) {
		return hrerrors.ErrReferenceConflict
	}
	if _, err := activeRecruitmentPrincipalWorker(ctx, v.workers, assessment.AssessedByPrincipalID, assessment.AssessedAt); err != nil {
		return err
	}
	if assessment.PerformanceReviewID != nil {
		review, err := v.reviews.FindByIDForUpdate(ctx, *assessment.PerformanceReviewID)
		if err != nil {
			return workerAssessmentReferenceError(err)
		}
		if review == nil || !strings.EqualFold(review.WorkerID, assessment.WorkerID) || review.FinalizedAt == nil {
			return hrerrors.ErrReferenceConflict
		}
	}
	if submitting || finalizing {
		return v.validateAssessmentScoresPresent(ctx, assessment)
	}
	matches, err := v.assessments.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "worker_id", Values: []any{assessment.WorkerID}},
		{Column: "talent_review_cycle_id", Values: []any{assessment.TalentReviewCycleID}},
		{Column: "assessment_version", Values: []any{assessment.AssessmentVersion}},
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

func (v *talentAssessmentValidator) validateAssessmentScoresPresent(ctx context.Context, assessment *entity.TalentAssessment) error {
	dimensions, err := v.dimensions.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "framework_id", Values: []any{assessment.FrameworkID}},
	}, 200, 0)
	if err != nil {
		return err
	}
	scores, err := v.scores.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "assessment_id", Values: []any{assessment.ID}},
	}, 200, 0)
	if err != nil {
		return err
	}
	if len(scores) == 0 {
		return hrerrors.Invalid("scores", "at least one dimension score is required")
	}
	scored := make(map[string]bool, len(scores))
	for _, sc := range scores {
		scored[strings.ToLower(sc.DimensionID)] = true
	}
	for _, dim := range dimensions {
		if dim.IsRequired && !scored[strings.ToLower(dim.ID)] {
			return hrerrors.Invalid("scores", "required dimension score is missing")
		}
	}
	return nil
}

func (v *talentAssessmentScoreValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_talent_assessment_scores" || operation.RequireDependency(v.assessments) != nil ||
		operation.RequireDependency(v.dimensions) != nil || operation.RequireDependency(v.cycles) != nil ||
		operation.RequireDependency(v.scores) != nil || operation.RequireDependency(v.decisions) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.TalentAssessmentScore
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.TalentAssessmentScore)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	score := previous
	if after != nil {
		var ok bool
		score, ok = after.(*entity.TalentAssessmentScore)
		if !ok || score == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous != nil && (previous.ID != score.ID || previous.AssessmentID != score.AssessmentID ||
			previous.FrameworkID != score.FrameworkID || previous.DimensionID != score.DimensionID) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	assessment, err := v.assessments.FindByIDForUpdate(ctx, score.AssessmentID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if assessment == nil || assessment.Status != "draft" {
		return hrerrors.ErrWorkflowRequired
	}
	if !strings.EqualFold(assessment.FrameworkID, score.FrameworkID) {
		return hrerrors.ErrReferenceConflict
	}
	cycle, err := v.cycles.FindByIDForUpdate(ctx, assessment.TalentReviewCycleID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if cycle == nil || (cycle.Status != "open" && cycle.Status != "calibration") {
		return hrerrors.ErrWorkflowRequired
	}
	dimension, err := v.dimensions.FindByIDForUpdate(ctx, score.DimensionID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if dimension == nil || !strings.EqualFold(dimension.FrameworkID, score.FrameworkID) {
		return hrerrors.ErrReferenceConflict
	}
	if previous != nil {
		calCount, err := v.decisions.CountWhere(ctx, []domainrepo.Filter{
			{Column: "assessment_id", Values: []any{previous.AssessmentID}},
			{Column: "dimension_id", Values: []any{previous.DimensionID}},
		})
		if err != nil {
			return err
		}
		if calCount != 0 {
			return hrerrors.ErrReferenceConflict
		}
		if after == nil {
			return nil
		}
	}
	matches, err := v.scores.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "assessment_id", Values: []any{score.AssessmentID}},
		{Column: "dimension_id", Values: []any{score.DimensionID}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != score.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *talentCalibrationSessionValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_talent_calibration_sessions" || operation.RequireDependency(v.cycles) != nil ||
		operation.RequireDependency(v.sessions) != nil || operation.RequireDependency(v.decisions) != nil ||
		operation.RequireDependency(v.workers) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.TalentCalibrationSession
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.TalentCalibrationSession)
		if !ok || previous == nil || (previous.Status != "planned" && previous.Status != "open") {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if after == nil {
		if previous.Status != "planned" {
			return hrerrors.ErrWorkflowRequired
		}
		decisionCount, err := v.decisions.CountWhere(ctx, []domainrepo.Filter{{Column: "session_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if decisionCount != 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	session, ok := after.(*entity.TalentCalibrationSession)
	if !ok || session == nil {
		return hrerrors.ErrInvalidInput
	}
	opening := previous != nil && previous.Status == "planned" && session.Status == "open"
	stayingOpen := previous != nil && previous.Status == "open" && reflect.DeepEqual(previous, session)
	if stayingOpen {
		cycle, err := v.cycles.FindByIDForUpdate(ctx, session.TalentReviewCycleID)
		if err != nil {
			return workerAssessmentReferenceError(err)
		}
		if cycle == nil || cycle.Status != "calibration" {
			return hrerrors.ErrWorkflowRequired
		}
		return nil
	}
	if opening {
		unchanged := *session
		unchanged.Status = previous.Status
		if !reflect.DeepEqual(*previous, unchanged) || session.ClosedAt != nil || session.MinutesDocumentVersionReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
	} else {
		if (previous != nil && previous.Status != "planned") || session.Status != "planned" ||
			session.ClosedAt != nil || session.MinutesDocumentVersionReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != session.ID || previous.TalentReviewCycleID != session.TalentReviewCycleID ||
			previous.SessionCode != session.SessionCode) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	cycle, err := v.cycles.FindByIDForUpdate(ctx, session.TalentReviewCycleID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if cycle == nil || (opening && cycle.Status != "calibration") || (!opening && cycle.Status != "open" && cycle.Status != "calibration") {
		return hrerrors.ErrWorkflowRequired
	}
	if _, err := activeRecruitmentPrincipalWorker(ctx, v.workers, session.ChairPrincipalID, time.Now()); err != nil {
		return err
	}
	matches, err := v.sessions.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "talent_review_cycle_id", Values: []any{session.TalentReviewCycleID}},
		{Column: "session_code", Values: []any{session.SessionCode}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != session.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *talentCalibrationDecisionValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_talent_calibration_decisions" || operation.RequireDependency(v.sessions) != nil ||
		operation.RequireDependency(v.cycles) != nil || operation.RequireDependency(v.assessments) != nil ||
		operation.RequireDependency(v.scores) != nil || operation.RequireDependency(v.decisions) != nil ||
		operation.RequireDependency(v.workers) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before != nil {
		return hrerrors.ErrWorkflowRequired
	}
	decision, ok := after.(*entity.TalentCalibrationDecision)
	if !ok || decision == nil {
		return hrerrors.ErrInvalidInput
	}
	session, err := v.sessions.FindByIDForUpdate(ctx, decision.SessionID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if session == nil || session.Status != "open" {
		return hrerrors.ErrWorkflowRequired
	}
	cycle, err := v.cycles.FindByIDForUpdate(ctx, decision.TalentReviewCycleID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if cycle == nil || cycle.Status != "calibration" || !strings.EqualFold(cycle.ID, session.TalentReviewCycleID) {
		return hrerrors.ErrWorkflowRequired
	}
	assessment, err := v.assessments.FindByIDForUpdate(ctx, decision.AssessmentID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if assessment == nil || assessment.Status != "finalized" {
		return hrerrors.ErrWorkflowRequired
	}
	if strings.EqualFold(assessment.AssessedByPrincipalID, decision.DecidedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, assessment.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || (worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, decision.DecidedByPrincipalID)) {
		return hrerrors.ErrForbidden
	}
	if _, err := activeRecruitmentPrincipalWorker(ctx, v.workers, decision.DecidedByPrincipalID, decision.DecidedAt); err != nil {
		return err
	}
	scores, err := v.scores.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "assessment_id", Values: []any{decision.AssessmentID}},
		{Column: "dimension_id", Values: []any{decision.DimensionID}},
	}, 2, 0)
	if err != nil {
		return err
	}
	if len(scores) != 1 {
		return hrerrors.ErrReferenceConflict
	}
	if err := talentdomain.ValidateCalibration(session, assessment, &scores[0], decision); err != nil {
		return err
	}
	existing, err := v.decisions.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "session_id", Values: []any{decision.SessionID}},
		{Column: "assessment_id", Values: []any{decision.AssessmentID}},
		{Column: "dimension_id", Values: []any{decision.DimensionID}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, row := range existing {
		if row.ID != decision.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *talentPoolValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_talent_pools" || operation.RequireDependency(v.pools) != nil ||
		operation.RequireDependency(v.memberships) != nil || operation.RequireDependency(v.jobs) != nil ||
		operation.RequireDependency(v.workers) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.TalentPool
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.TalentPool)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if after == nil {
		count, err := v.memberships.CountWhere(ctx, []domainrepo.Filter{{Column: "talent_pool_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if count != 0 {
			return hrerrors.ErrReferenceConflict
		}
		return nil
	}
	pool, ok := after.(*entity.TalentPool)
	if !ok || pool == nil {
		return hrerrors.ErrInvalidInput
	}
	if previous != nil && (previous.ID != pool.ID || previous.PoolCode != pool.PoolCode || previous.PoolVersion != pool.PoolVersion ||
		!strings.EqualFold(previous.LegalEntityReferenceID, pool.LegalEntityReferenceID) ||
		!reflect.DeepEqual(previous.TargetJobID, pool.TargetJobID) || entity.CompareDates(previous.ValidFrom, pool.ValidFrom) != 0) {
		count, err := v.memberships.CountWhere(ctx, []domainrepo.Filter{{Column: "talent_pool_id", Values: []any{previous.ID}}})
		if err != nil {
			return err
		}
		if count != 0 {
			return hrerrors.ErrReferenceConflict
		}
		if previous.ID != pool.ID {
			return hrerrors.ErrInvalidInput
		}
	}
	if err := verifyFoundationLegalEntity(ctx, v.foundation, pool.LegalEntityReferenceID); err != nil {
		return err
	}
	if pool.TargetJobID != nil {
		job, err := v.jobs.FindByIDForUpdate(ctx, *pool.TargetJobID)
		if err != nil {
			return workerAssessmentReferenceError(err)
		}
		if job == nil || !job.IsActive || !compensationPeriodWithin(pool.ValidFrom, pool.ValidTo, job.ValidFrom, job.ValidTo) {
			return hrerrors.ErrReferenceConflict
		}
	}
	if _, err := activeRecruitmentPrincipalWorker(ctx, v.workers, pool.OwnerPrincipalID, time.Now()); err != nil {
		return err
	}
	matches, err := v.pools.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "legal_entity_reference_id", Values: []any{pool.LegalEntityReferenceID}},
		{Column: "pool_code", Values: []any{pool.PoolCode}},
		{Column: "pool_version", Values: []any{pool.PoolVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != pool.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func (v *talentPoolMembershipValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_talent_pool_memberships" || operation.RequireDependency(v.pools) != nil ||
		operation.RequireDependency(v.workers) != nil || operation.RequireDependency(v.assessments) != nil ||
		operation.RequireDependency(v.memberships) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.TalentPoolMembership
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.TalentPoolMembership)
		if !ok || previous == nil || (previous.Status != "nominated" && previous.Status != "active") {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if after == nil {
		if previous.Status != "nominated" {
			return hrerrors.ErrWorkflowRequired
		}
		return nil
	}
	membership, ok := after.(*entity.TalentPoolMembership)
	if !ok || membership == nil {
		return hrerrors.ErrInvalidInput
	}
	approving := previous != nil && previous.Status == "nominated" && membership.Status == "active"
	rejecting := previous != nil && previous.Status == "nominated" && membership.Status == "rejected"
	exiting := previous != nil && previous.Status == "active" && membership.Status == "exited"
	switch {
	case approving:
		if membership.ApprovedByPrincipalID == nil || membership.ApprovedAt == nil || membership.ExitReasonCode != nil ||
			strings.EqualFold(membership.NominatedByPrincipalID, *membership.ApprovedByPrincipalID) {
			return hrerrors.ErrWorkflowRequired
		}
		unchanged := *membership
		unchanged.Status = previous.Status
		unchanged.ApprovedByPrincipalID = nil
		unchanged.ApprovedAt = nil
		if !reflect.DeepEqual(*previous, unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
	case rejecting:
		if membership.ExitReasonCode == nil || strings.TrimSpace(*membership.ExitReasonCode) == "" ||
			membership.ApprovedByPrincipalID != nil || membership.ApprovedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
		unchanged := *membership
		unchanged.Status = previous.Status
		unchanged.ExitReasonCode = previous.ExitReasonCode
		if !reflect.DeepEqual(*previous, unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
	case exiting:
		if membership.ValidTo == nil || entity.CompareDates(*membership.ValidTo, membership.ValidFrom) < 0 ||
			membership.ExitReasonCode == nil || strings.TrimSpace(*membership.ExitReasonCode) == "" {
			return hrerrors.ErrWorkflowRequired
		}
		unchanged := *membership
		unchanged.Status = previous.Status
		unchanged.ValidTo = previous.ValidTo
		unchanged.ExitReasonCode = previous.ExitReasonCode
		if !reflect.DeepEqual(*previous, unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
	default:
		if (previous != nil && previous.Status != "nominated") || membership.Status != "nominated" ||
			membership.ApprovedByPrincipalID != nil || membership.ApprovedAt != nil || membership.ExitReasonCode != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.ID != membership.ID || previous.TalentPoolID != membership.TalentPoolID ||
			previous.WorkerID != membership.WorkerID || previous.AssessmentID != membership.AssessmentID ||
			!strings.EqualFold(previous.NominatedByPrincipalID, membership.NominatedByPrincipalID)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	pool, err := v.pools.FindByIDForUpdate(ctx, membership.TalentPoolID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if pool == nil || !compensationPeriodWithin(membership.ValidFrom, membership.ValidTo, pool.ValidFrom, pool.ValidTo) {
		return hrerrors.ErrReferenceConflict
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, membership.WorkerID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive ||
		(worker.TerminationDate != nil && (membership.ValidTo == nil || entity.CompareDates(*membership.ValidTo, *worker.TerminationDate) > 0)) {
		return hrerrors.ErrReferenceConflict
	}
	if approving && worker.IAMPrincipalReferenceID != nil &&
		strings.EqualFold(*worker.IAMPrincipalReferenceID, *membership.ApprovedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	assessment, err := v.assessments.FindByIDForUpdate(ctx, membership.AssessmentID)
	if err != nil {
		return workerAssessmentReferenceError(err)
	}
	// Bandingkan pada zona waktu UTC agar instant TIMESTAMPTZ dari database konsisten dengan DATE server UTC.
	if assessment == nil || assessment.Status != "finalized" || !strings.EqualFold(assessment.WorkerID, membership.WorkerID) ||
		assessment.FinalizedAt == nil || entity.CompareDates(membership.ValidFrom, assessment.FinalizedAt.UTC()) < 0 {
		return hrerrors.ErrReferenceConflict
	}
	if _, err := activeRecruitmentPrincipalWorker(ctx, v.workers, membership.NominatedByPrincipalID, time.Now()); err != nil {
		return err
	}
	if approving || exiting || membership.Status == "nominated" {
		existing, err := v.memberships.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "talent_pool_id", Values: []any{membership.TalentPoolID}},
			{Column: "worker_id", Values: []any{membership.WorkerID}},
		}, 100, 0)
		if err != nil {
			return err
		}
		for _, row := range existing {
			if row.ID == membership.ID {
				continue
			}
			if entity.CompareDates(row.ValidFrom, membership.ValidFrom) == 0 {
				return hrerrors.ErrAlreadyExists
			}
			if (approving || exiting) && (row.Status == "active" || row.Status == "exited") &&
				datesOverlap(membership.ValidFrom, membership.ValidTo, row.ValidFrom, row.ValidTo) {
				return hrerrors.ErrAlreadyExists
			}
		}
	}
	return nil
}

func talentFrameworkDimensionsComplete(ctx context.Context, reader talentDimensionReader, frameworkID string) error {
	dims, err := reader.SelectWhere(ctx, []domainrepo.Filter{{Column: "framework_id", Values: []any{frameworkID}}}, 200, 0)
	if err != nil {
		return err
	}
	if len(dims) == 0 {
		return hrerrors.Invalid("dimensions", "framework requires at least one dimension")
	}
	weights := make([]entity.Decimal, 0, len(dims))
	for _, dim := range dims {
		weights = append(weights, dim.WeightPercent)
	}
	total, err := entity.DecimalSum(3, weights...)
	if err != nil {
		return err
	}
	rat, err := total.Rat()
	if err != nil {
		return err
	}
	expected, err := entity.Decimal("100").Rat()
	if err != nil {
		return err
	}
	if rat.Cmp(expected) != 0 {
		return hrerrors.Invalid("weight_percent", "dimension weights must sum to one hundred")
	}
	return nil
}
