package service

import (
	"context"
	"reflect"
	"time"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Command hanya dipakai command domain bernama tetap yang memvalidasi transisinya sendiri.
// Action tidak boleh diteruskan langsung dari input transport.
func (r *Resource[T, K, P]) Command(
	ctx context.Context,
	key K,
	action string,
	change func(context.Context, *T, Actor, time.Time) (*T, error),
) (*T, error) {
	if r == nil {
		return nil, hrerrors.ErrUnavailable
	}
	endsBenefitEnrollment := r.spec.Table == "hr_benefit_enrollments" && action == "end"
	approvesBenefitEnrollment := r.spec.Table == "hr_benefit_enrollments" && action == "approve"
	closesLeaveEntitlement := r.spec.Table == "hr_leave_entitlements" && action == "close"
	approvesCompensationRecord := r.spec.Table == "hr_compensation_records" && action == "approve"
	changesLearningEnrollment := r.spec.Table == "hr_learning_enrollments" && (action == "start" || action == "complete" || action == "withdraw")
	changesServiceRequest := r.spec.Table == "hr_service_requests" && (action == "resolve" || action == "close" || action == "assign" || action == "cancel")
	postsLeave := r.spec.Table == "hr_leave_entitlements" && action == "post_leave_entry"
	rejectsLeaveAdjustment := r.spec.Table == "hr_leave_adjustments" && action == "reject"
	approvesLeaveAdjustment := r.spec.Table == "hr_leave_adjustments" && action == "approve"
	changesPayroll := r.spec.Table == "hr_payroll_runs" && (action == "approve" || action == "submit_review" || action == "request_posting" || action == "calculate" || action == "cancel")
	completesDevelopmentPlan := r.spec.Table == "hr_development_plans" && action == "complete"
	completesDevelopmentAction := r.spec.Table == "hr_development_actions" && action == "complete"
	approvesDevelopmentPlan := r.spec.Table == "hr_development_plans" && action == "approve"
	approvesSuccessionPlan := r.spec.Table == "hr_succession_plans" && action == "approve"
	approvesCareerPath := r.spec.Table == "hr_career_paths" && action == "approve"
	approvesWorkerAssignment := r.spec.Table == "hr_worker_assignments" && action == "approve"
	supersedesWorkforcePlan := r.spec.Table == "hr_workforce_plans" && action == "supersede"
	approvesWorkforcePlan := r.spec.Table == "hr_workforce_plans" && action == "approve"
	cancelsWorkforcePlan := r.spec.Table == "hr_workforce_plans" && action == "cancel"
	submitsWorkforcePlan := r.spec.Table == "hr_workforce_plans" && action == "submit"
	activatesPosition := r.spec.Table == "hr_positions" && action == "activate"
	linksWorkerPrincipal := r.spec.Table == "hr_workers" && action == "link_principal"
	signsEmploymentContract := r.spec.Table == "hr_employment_contracts" && action == "record_signature"
	activatesWorker := r.spec.Table == "hr_workers" && action == "activate"
	appliesPrimaryAssignment := r.spec.Table == "hr_workers" && action == "apply_primary_assignment"
	launchesPerformanceCycle := r.spec.Table == "hr_performance_cycles" && action == "launch"
	completesOffboardingPlan := r.spec.Table == "hr_offboarding_plans" && action == "complete"
	completesOffboardingTask := r.spec.Table == "hr_offboarding_tasks" && action == "complete"
	waivesOffboardingTask := r.spec.Table == "hr_offboarding_tasks" && action == "waive"
	cancelsOffboardingPlan := r.spec.Table == "hr_offboarding_plans" && action == "cancel"
	startsOffboardingPlan := r.spec.Table == "hr_offboarding_plans" && action == "start"
	approvesOffboardingPlan := r.spec.Table == "hr_offboarding_plans" && action == "approve"
	startsOnboardingPlan := r.spec.Table == "hr_onboarding_plans" && action == "start"
	completesOnboardingPlan := r.spec.Table == "hr_onboarding_plans" && action == "complete"
	completesOnboardingTask := r.spec.Table == "hr_onboarding_tasks" && action == "complete"
	finalizesTalent := r.spec.Table == "hr_talent_assessments" && action == "finalize"
	changesInterview := r.spec.Table == "hr_interviews" && (action == "cancel" || action == "start" || action == "complete")
	startsApplicationScreening := r.spec.Table == "hr_applications" && action == "start_screening"
	shortlistsApplication := r.spec.Table == "hr_applications" && action == "advance"
	rejectsApplication := r.spec.Table == "hr_applications" && action == "reject"
	submitsRequisition := r.spec.Table == "hr_requisitions" && action == "submit"
	approvesRequisition := r.spec.Table == "hr_requisitions" && action == "approve"
	rejectsRequisition := r.spec.Table == "hr_requisitions" && action == "reject"
	opensRequisition := r.spec.Table == "hr_requisitions" && action == "open"
	closesRequisition := r.spec.Table == "hr_requisitions" && action == "close"
	publishesJobPosting := r.spec.Table == "hr_job_postings" && action == "publish"
	closesJobPosting := r.spec.Table == "hr_job_postings" && action == "close"
	changesOffer := r.spec.Table == "hr_offers" && (action == "withdraw" || action == "approve" || action == "issue" || action == "accept" || action == "decline")
	withdrawsJobPosting := r.spec.Table == "hr_job_postings" && action == "withdraw"
	completesApplicationAssessment := r.spec.Table == "hr_application_assessments" && action == "complete"
	launchesTalentCycle := r.spec.Table == "hr_talent_review_cycles" && action == "launch"
	startsTalentCalibration := r.spec.Table == "hr_talent_review_cycles" && action == "start_calibration"
	closesTalentCycle := r.spec.Table == "hr_talent_review_cycles" && action == "close"
	submitsTalentAssessment := r.spec.Table == "hr_talent_assessments" && action == "submit"
	withdrawsTalentAssessment := r.spec.Table == "hr_talent_assessments" && action == "withdraw"
	opensCalibrationSession := r.spec.Table == "hr_talent_calibration_sessions" && action == "open"
	approvesCalibrationSession := r.spec.Table == "hr_talent_calibration_sessions" && action == "approve"
	approvesPoolMembership := r.spec.Table == "hr_talent_pool_memberships" && action == "approve"
	rejectsPoolMembership := r.spec.Table == "hr_talent_pool_memberships" && action == "reject"
	exitsPoolMembership := r.spec.Table == "hr_talent_pool_memberships" && action == "exit"
	submitsPerfGoal := r.spec.Table == "hr_performance_goals" && action == "submit"
	approvesPerfGoal := r.spec.Table == "hr_performance_goals" && action == "approve"
	rejectsPerfGoal := r.spec.Table == "hr_performance_goals" && action == "reject"
	submitsPerfReview := r.spec.Table == "hr_performance_reviews" && action == "submit"
	finalizesPerfReview := r.spec.Table == "hr_performance_reviews" && action == "finalize"
	closesPerfCycle := r.spec.Table == "hr_performance_cycles" && action == "close"
	submitsPromotion := r.spec.Table == "hr_promotion_recommendations" && action == "submit"
	approvesPromotion := r.spec.Table == "hr_promotion_recommendations" && action == "approve"
	rejectsPromotion := r.spec.Table == "hr_promotion_recommendations" && action == "reject"
	returnsPromotion := r.spec.Table == "hr_promotion_recommendations" && action == "return"
	appliesPromotion := r.spec.Table == "hr_promotion_recommendations" && action == "apply"
	acknowledgesPolicy := r.spec.Table == "hr_policy_acknowledgements" && action == "acknowledge"
	allowed := endsBenefitEnrollment || approvesBenefitEnrollment || closesLeaveEntitlement || changesServiceRequest || changesLearningEnrollment || approvesCompensationRecord || postsLeave || rejectsLeaveAdjustment || approvesLeaveAdjustment || changesPayroll || completesDevelopmentPlan || completesDevelopmentAction || approvesDevelopmentPlan || approvesCareerPath || approvesSuccessionPlan || approvesWorkerAssignment || supersedesWorkforcePlan || approvesWorkforcePlan || cancelsWorkforcePlan || submitsWorkforcePlan || activatesPosition || activatesWorker || signsEmploymentContract || linksWorkerPrincipal || appliesPrimaryAssignment || launchesPerformanceCycle || completesOffboardingPlan || completesOffboardingTask || waivesOffboardingTask || startsOffboardingPlan || cancelsOffboardingPlan || approvesOffboardingPlan || startsOnboardingPlan || completesOnboardingPlan || completesOnboardingTask || finalizesTalent || changesInterview || startsApplicationScreening || shortlistsApplication || rejectsApplication || submitsRequisition || approvesRequisition || rejectsRequisition || opensRequisition || closesRequisition || publishesJobPosting || closesJobPosting || withdrawsJobPosting || changesOffer || completesApplicationAssessment || launchesTalentCycle || startsTalentCalibration || closesTalentCycle || submitsTalentAssessment || withdrawsTalentAssessment || opensCalibrationSession || approvesCalibrationSession || approvesPoolMembership || rejectsPoolMembership || exitsPoolMembership || submitsPerfGoal || approvesPerfGoal || rejectsPerfGoal || submitsPerfReview || finalizesPerfReview || closesPerfCycle || submitsPromotion || approvesPromotion || rejectsPromotion || returnsPromotion || appliesPromotion || acknowledgesPolicy
	if !allowed || change == nil {
		return nil, hrerrors.ErrWorkflowRequired
	}
	if err := r.spec.ValidateKey(key); err != nil {
		return nil, err
	}
	actor, err := r.authorize(ctx, action, nil)
	if err != nil {
		return nil, err
	}
	var result *T
	err = r.deps.Transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		stored, err := r.repo.FindByIDForUpdate(txCtx, key)
		if err != nil {
			return err
		}
		before, err := clone(stored)
		if err != nil {
			return err
		}
		checked, err := r.authorize(txCtx, action, before)
		if err != nil {
			return err
		}
		if err := sameActor(actor, checked); err != nil {
			return err
		}
		working, err := clone(before)
		if err != nil {
			return err
		}
		now := r.deps.Clock.Now()
		if now.IsZero() {
			return hrerrors.ErrInvalidInput
		}
		after, err := change(txCtx, working, actor, now)
		if err != nil {
			return err
		}
		if after == nil || !reflect.DeepEqual(r.spec.Key(before), r.spec.Key(after)) {
			return hrerrors.ErrInvalidInput
		}
		if err := P(after).Validate(); err != nil {
			return err
		}
		checked, err = r.authorize(txCtx, action, after)
		if err != nil {
			return err
		}
		if err := sameActor(actor, checked); err != nil {
			return err
		}
		if err := r.CheckIntegrity(txCtx, before, after); err != nil {
			return err
		}
		if err := r.repo.Update(txCtx, after); err != nil {
			return err
		}
		if err := r.emit(txCtx, action, after, actor); err != nil {
			return err
		}
		result = after
		return r.finish(txCtx, action, after, actor)
	})
	if err != nil {
		return nil, err
	}
	return clone(result)
}

// NewID menyediakan ID server untuk record turunan di dalam command transaksi.
func (r *Resource[T, K, P]) NewID() (string, error) {
	if r == nil {
		return "", hrerrors.ErrUnavailable
	}
	return r.deps.IDs.NewID()
}

// GetForUpdate mempertahankan pemeriksaan akses saat command mengunci agregat terkait.
func (r *Resource[T, K, P]) GetForUpdate(ctx context.Context, key K) (*T, error) {
	if r == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := r.spec.ValidateKey(key); err != nil {
		return nil, err
	}
	actor, err := r.authorize(ctx, "read", nil)
	if err != nil {
		return nil, err
	}
	value, err := r.repo.FindByIDForUpdate(ctx, key)
	if err != nil {
		return nil, err
	}
	checked, err := r.authorize(ctx, "read", value)
	if err != nil {
		return nil, err
	}
	if err := sameActor(actor, checked); err != nil {
		return nil, err
	}
	return clone(value)
}

// CollectComplete menolak kumpulan parsial akibat scope atau batas memori.
// Count sumber hanya boleh diminta setelah akses terhadap root agregat disetujui.
func CollectComplete[T any](ctx context.Context, expected int64, fetch func(int, int) (Page[T], error)) ([]T, error) {
	if expected < 0 || expected > 10000 || fetch == nil {
		return nil, hrerrors.ErrConflict
	}
	items := make([]T, 0)
	for offset := 0; ; {
		if err := checkContext(ctx); err != nil {
			return nil, err
		}
		page, err := fetch(domainrepo.MaxLimit, offset)
		if err != nil {
			return nil, err
		}
		if len(items)+len(page.Items) > 10000 {
			return nil, hrerrors.ErrConflict
		}
		if int64(len(items)+len(page.Items)) > expected {
			return nil, hrerrors.ErrForbidden
		}
		items = append(items, page.Items...)
		if len(page.Items) < domainrepo.MaxLimit {
			break
		}
		offset += len(page.Items)
	}
	if int64(len(items)) != expected {
		return nil, hrerrors.ErrForbidden
	}
	return items, nil
}
