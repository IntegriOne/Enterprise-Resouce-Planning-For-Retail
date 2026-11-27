package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	benefits "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/benefits/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelBenefitPlan(row *benefits.BenefitPlan) (*model.BenefitPlan, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource benefit-plan response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource benefit-plan ID", nil)
	}
	out := &model.BenefitPlan{ID: nodeID("benefit-plan", row.GetId()), RawID: row.GetId()}
	out.PlanCode = row.GetPlanCode()
	out.PlanVersion = int(row.GetPlanVersion())
	out.PlanName = row.GetPlanName()
	out.BenefitType = row.GetBenefitType()
	out.ProviderPartyReferenceID = row.ProviderPartyReferenceId
	out.JurisdictionCode = row.GetJurisdictionCode()
	if doc := row.GetEligibilityRule(); doc != nil {
		out.EligibilityRule = pbJSONString(doc.GetUtf8())
	}
	out.EmployeeContributionRuleReferenceID = row.EmployeeContributionRuleReferenceId
	out.EmployerContributionRuleReferenceID = row.EmployerContributionRuleReferenceId
	out.ValidFrom = row.GetValidFrom()
	out.ValidTo = row.ValidTo
	out.IsActive = row.GetIsActive()
	return out, nil
}

// BenefitPlan membaca detail benefit-plan berdasarkan ID.
func (a *Adapter) BenefitPlan(ctx context.Context, rawID string) (*model.BenefitPlan, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_benefit_plans", "human_resource.benefit_plans.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "benefit-plan", "benefit-plan")
	if err != nil {
		return nil, err
	}
	if a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.benefits.GetBenefitPlanByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "benefit-plan")
	}
	return toModelBenefitPlan(resp.GetItem())
}

// BenefitPlans mendaftar dan memfilter benefit-plan dengan paginasi.
func (a *Adapter) BenefitPlans(ctx context.Context, limit, offset int, f_plan_code *string, f_benefit_type *string, f_provider_party_reference_id *string, f_jurisdiction_code *string, f_employee_contribution_rule_reference_id *string, f_employer_contribution_rule_reference_id *string) (*Page[model.BenefitPlan], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_benefit_plans", "human_resource.benefit_plans.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid benefit-plan pagination", nil)
	}
	if a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_plan_code != nil, f_benefit_type != nil, f_provider_party_reference_id != nil, f_jurisdiction_code != nil, f_employee_contribution_rule_reference_id != nil, f_employer_contribution_rule_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *benefits.BenefitPlansResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_plan_code, "plan_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.benefits.FindBenefitPlanByPlanCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_benefit_type, "benefit_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.benefits.FindBenefitPlanByBenefitType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_provider_party_reference_id, "provider_party_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.benefits.FindBenefitPlanByProviderPartyReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_jurisdiction_code, "jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.benefits.FindBenefitPlanByJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_employee_contribution_rule_reference_id, "employee_contribution_rule_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.benefits.FindBenefitPlanByEmployeeContributionRuleReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_employer_contribution_rule_reference_id, "employer_contribution_rule_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.benefits.FindBenefitPlanByEmployerContributionRuleReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.benefits.GetAllBenefitPlans(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "benefit-plan")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "benefit-plan", toModelBenefitPlan)
}

// CreateBenefitPlan membuat benefit-plan baru melalui HRM.
func (a *Adapter) CreateBenefitPlan(ctx context.Context, input model.CreateBenefitPlanInput) (*model.BenefitPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_benefit_plans", "create", "human_resource.benefit_plans.create")
	if err != nil {
		return nil, err
	}
	if a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &benefits.BenefitPlanInput{}
	patch.PlanCode = input.PlanCode
	patch.PlanVersion = optionalInt32Ptr(input.PlanVersion)
	patch.PlanName = input.PlanName
	patch.BenefitType = input.BenefitType
	patch.ProviderPartyReferenceId = input.ProviderPartyReferenceID
	patch.JurisdictionCode = input.JurisdictionCode
	if b, jErr := optionalJSONBytes(input.EligibilityRule, "eligibility_rule"); jErr != nil {
		return nil, jErr
	} else if b != nil {
		patch.EligibilityRule = &common.JsonDocument{Utf8: b}
	}
	patch.EmployeeContributionRuleReferenceId = input.EmployeeContributionRuleReferenceID
	patch.EmployerContributionRuleReferenceId = input.EmployerContributionRuleReferenceID
	patch.ValidFrom = input.ValidFrom
	patch.ValidTo = input.ValidTo
	patch.IsActive = input.IsActive
	resp, err := a.benefits.CreateBenefitPlan(rpcCtx, &benefits.CreateBenefitPlanRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "benefit-plan")
	}
	return toModelBenefitPlan(resp.GetItem())
}

// UpdateBenefitPlan memperbarui benefit-plan berdasarkan FieldMask.
func (a *Adapter) UpdateBenefitPlan(ctx context.Context, rawID string, input model.UpdateBenefitPlanInput) (*model.BenefitPlan, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_benefit_plans", "update", "human_resource.benefit_plans.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "benefit-plan", "benefit-plan")
	if err != nil {
		return nil, err
	}
	if a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &benefits.BenefitPlanInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "plan_code", input.PlanCode, &patch.PlanCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "plan_version", input.PlanVersion, &patch.PlanVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "plan_name", input.PlanName, &patch.PlanName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "benefit_type", input.BenefitType, &patch.BenefitType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "provider_party_reference_id", input.ProviderPartyReferenceID, &patch.ProviderPartyReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "jurisdiction_code", input.JurisdictionCode, &patch.JurisdictionCode, true); err != nil {
		return nil, err
	}
	if val, set := input.EligibilityRule.ValueOK(); set {
		if val == nil {
			patch.EligibilityRule = nil
		} else if b, jErr := requiredJSONBytes(*val, "eligibility_rule"); jErr != nil {
			return nil, jErr
		} else {
			patch.EligibilityRule = &common.JsonDocument{Utf8: b}
		}
		mask.Paths = append(mask.Paths, "eligibility_rule")
	}
	if err := applyStringMask(mask, "employee_contribution_rule_reference_id", input.EmployeeContributionRuleReferenceID, &patch.EmployeeContributionRuleReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "employer_contribution_rule_reference_id", input.EmployerContributionRuleReferenceID, &patch.EmployerContributionRuleReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "valid_from", input.ValidFrom, &patch.ValidFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "valid_to", input.ValidTo, &patch.ValidTo, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_active", input.IsActive, &patch.IsActive, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("benefit-plan update is empty", nil)
	}
	resp, err := a.benefits.UpdateBenefitPlan(rpcCtx, &benefits.UpdateBenefitPlanRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "benefit-plan")
	}
	return toModelBenefitPlan(resp.GetItem())
}

// DeleteBenefitPlan menghapus benefit-plan berdasarkan ID.
func (a *Adapter) DeleteBenefitPlan(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_benefit_plans", "human_resource.benefit_plans.delete", "benefit-plan", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.benefits.DeleteBenefitPlan(c, req)
	})
}

func toModelBenefitEnrollment(row *benefits.BenefitEnrollment) (*model.BenefitEnrollment, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource benefit-enrollment response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource benefit-enrollment ID", nil)
	}
	out := &model.BenefitEnrollment{ID: nodeID("benefit-enrollment", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.BenefitPlanID = row.GetBenefitPlanId()
	out.EnrollmentStatus = row.GetEnrollmentStatus()
	out.EffectiveFrom = row.GetEffectiveFrom()
	out.EffectiveTo = row.EffectiveTo
	out.EnrollmentReference = row.EnrollmentReference
	out.ProviderMemberReference = row.ProviderMemberReference
	out.EmployeeContributionAmount = row.EmployeeContributionAmount
	out.EmployerContributionAmount = row.EmployerContributionAmount
	out.CurrencyCode = row.CurrencyCode
	out.ElectedAt = pbTimePtr(row.GetElectedAt())
	out.VerifiedAt = pbTimePtr(row.GetVerifiedAt())
	return out, nil
}

// BenefitEnrollment membaca detail benefit-enrollment berdasarkan ID.
func (a *Adapter) BenefitEnrollment(ctx context.Context, rawID string) (*model.BenefitEnrollment, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_benefit_enrollments", "human_resource.benefit_enrollments.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "benefit-enrollment", "benefit-enrollment")
	if err != nil {
		return nil, err
	}
	if a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.benefits.GetBenefitEnrollmentByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "benefit-enrollment")
	}
	return toModelBenefitEnrollment(resp.GetItem())
}

// BenefitEnrollments mendaftar dan memfilter benefit-enrollment dengan paginasi.
func (a *Adapter) BenefitEnrollments(ctx context.Context, limit, offset int, f_worker_id *string, f_benefit_plan_id *string, f_enrollment_status *string, f_currency_code *string) (*Page[model.BenefitEnrollment], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_benefit_enrollments", "human_resource.benefit_enrollments.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid benefit-enrollment pagination", nil)
	}
	if a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_benefit_plan_id != nil, f_enrollment_status != nil, f_currency_code != nil)
	if err != nil {
		return nil, err
	}
	var resp *benefits.BenefitEnrollmentsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.benefits.FindBenefitEnrollmentByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_benefit_plan_id, "benefit_plan_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.benefits.FindBenefitEnrollmentByBenefitPlanID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_enrollment_status, "enrollment_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.benefits.FindBenefitEnrollmentByEnrollmentStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_currency_code, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.benefits.FindBenefitEnrollmentByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.benefits.GetAllBenefitEnrollments(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "benefit-enrollment")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "benefit-enrollment", toModelBenefitEnrollment)
}

// CreateBenefitEnrollment membuat benefit-enrollment baru melalui HRM.
func (a *Adapter) CreateBenefitEnrollment(ctx context.Context, input model.CreateBenefitEnrollmentInput) (*model.BenefitEnrollment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_benefit_enrollments", "create", "human_resource.benefit_enrollments.create")
	if err != nil {
		return nil, err
	}
	if a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &benefits.BenefitEnrollmentInput{}
	patch.WorkerId = input.WorkerID
	patch.BenefitPlanId = input.BenefitPlanID
	patch.EnrollmentStatus = input.EnrollmentStatus
	patch.EffectiveFrom = input.EffectiveFrom
	patch.EffectiveTo = input.EffectiveTo
	patch.EnrollmentReference = input.EnrollmentReference
	patch.ProviderMemberReference = input.ProviderMemberReference
	patch.EmployeeContributionAmount = input.EmployeeContributionAmount
	patch.EmployerContributionAmount = input.EmployerContributionAmount
	patch.CurrencyCode = input.CurrencyCode
	patch.ElectedAt = optionalTimestamp(input.ElectedAt)
	patch.VerifiedAt = optionalTimestamp(input.VerifiedAt)
	resp, err := a.benefits.CreateBenefitEnrollment(rpcCtx, &benefits.CreateBenefitEnrollmentRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "benefit-enrollment")
	}
	return toModelBenefitEnrollment(resp.GetItem())
}

// UpdateBenefitEnrollment memperbarui benefit-enrollment berdasarkan FieldMask.
func (a *Adapter) UpdateBenefitEnrollment(ctx context.Context, rawID string, input model.UpdateBenefitEnrollmentInput) (*model.BenefitEnrollment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_benefit_enrollments", "update", "human_resource.benefit_enrollments.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "benefit-enrollment", "benefit-enrollment")
	if err != nil {
		return nil, err
	}
	if a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &benefits.BenefitEnrollmentInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "benefit_plan_id", input.BenefitPlanID, &patch.BenefitPlanId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "enrollment_status", input.EnrollmentStatus, &patch.EnrollmentStatus, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_from", input.EffectiveFrom, &patch.EffectiveFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "effective_to", input.EffectiveTo, &patch.EffectiveTo, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "enrollment_reference", input.EnrollmentReference, &patch.EnrollmentReference, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "provider_member_reference", input.ProviderMemberReference, &patch.ProviderMemberReference, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "employee_contribution_amount", input.EmployeeContributionAmount, &patch.EmployeeContributionAmount, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "employer_contribution_amount", input.EmployerContributionAmount, &patch.EmployerContributionAmount, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "currency_code", input.CurrencyCode, &patch.CurrencyCode, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "elected_at", input.ElectedAt, &patch.ElectedAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "verified_at", input.VerifiedAt, &patch.VerifiedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("benefit-enrollment update is empty", nil)
	}
	resp, err := a.benefits.UpdateBenefitEnrollment(rpcCtx, &benefits.UpdateBenefitEnrollmentRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "benefit-enrollment")
	}
	return toModelBenefitEnrollment(resp.GetItem())
}

// DeleteBenefitEnrollment menghapus benefit-enrollment berdasarkan ID.
func (a *Adapter) DeleteBenefitEnrollment(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.benefits == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_benefit_enrollments", "human_resource.benefit_enrollments.delete", "benefit-enrollment", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.benefits.DeleteBenefitEnrollment(c, req)
	})
}
