package benefits

import (
	"context"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/benefits/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/benefits"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func plan(value *entity.BenefitPlan) (*pb.BenefitPlan, error) {
	rule, err := response.JSONValue(value.EligibilityRule)
	if err != nil {
		return nil, err
	}
	return &pb.BenefitPlan{
		Id: value.ID, PlanCode: value.PlanCode, PlanVersion: value.PlanVersion,
		PlanName: value.PlanName, BenefitType: value.BenefitType,
		ProviderPartyReferenceId: value.ProviderPartyReferenceID,
		JurisdictionCode:         value.JurisdictionCode, EligibilityRule: rule,
		EmployeeContributionRuleReferenceId: value.EmployeeContributionRuleReferenceID,
		EmployerContributionRuleReferenceId: value.EmployerContributionRuleReferenceID,
		ValidFrom:                           value.ValidFrom.Format("2006-01-02"), ValidTo: dateString(value.ValidTo),
		IsActive: value.IsActive,
	}, nil
}
func planResponse(value *entity.BenefitPlan, err error) (*pb.BenefitPlanResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	item, err := plan(value)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.BenefitPlanResponse{Item: item}, nil
}
func plans(page operation.Page[entity.BenefitPlan], err error) (*pb.BenefitPlansResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.BenefitPlan, 0, len(page.Items))
	for index := range page.Items {
		item, err := plan(&page.Items[index])
		if err != nil {
			return nil, response.Error(err)
		}
		items = append(items, item)
	}
	return &pb.BenefitPlansResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateBenefitPlan(ctx context.Context, request *pb.CreateBenefitPlanRequest) (*pb.BenefitPlanResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	from, err := response.Date(in.GetValidFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	rule, err := response.JSON(in.EligibilityRule)
	if err != nil {
		return nil, response.Error(err)
	}
	return planResponse(h.service.CreateBenefitPlan(ctx, service.CreateBenefitPlanInput{
		PlanCode: in.GetPlanCode(), PlanVersion: in.GetPlanVersion(), PlanName: in.GetPlanName(),
		BenefitType: in.GetBenefitType(), ProviderPartyReferenceID: in.ProviderPartyReferenceId,
		JurisdictionCode: in.GetJurisdictionCode(), EligibilityRule: rule,
		EmployeeContributionRuleReferenceID: in.EmployeeContributionRuleReferenceId,
		EmployerContributionRuleReferenceID: in.EmployerContributionRuleReferenceId,
		ValidFrom:                           from, ValidTo: to, IsActive: in.IsActive,
	}))
}
func (h *Handler) UpdateBenefitPlan(ctx context.Context, request *pb.UpdateBenefitPlanRequest) (*pb.BenefitPlanResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask,
		"plan_code", "plan_version", "plan_name", "benefit_type", "provider_party_reference_id",
		"jurisdiction_code", "eligibility_rule", "employee_contribution_rule_reference_id",
		"employer_contribution_rule_reference_id", "valid_from", "valid_to", "is_active",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	from, err := dateField(mask["valid_from"], in.ValidFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(mask["valid_to"], in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	rule := operation.Field[entity.JSON]{Set: mask["eligibility_rule"]}
	if rule.Set && in.EligibilityRule != nil {
		value, err := response.JSON(in.EligibilityRule)
		if err != nil {
			return nil, response.Error(err)
		}
		rule.Value = &value
	}
	return planResponse(h.service.UpdateBenefitPlan(ctx, request.Id, service.UpdateBenefitPlanInput{
		PlanCode:                            field(mask["plan_code"], in.PlanCode),
		PlanVersion:                         field(mask["plan_version"], in.PlanVersion),
		PlanName:                            field(mask["plan_name"], in.PlanName),
		BenefitType:                         field(mask["benefit_type"], in.BenefitType),
		ProviderPartyReferenceID:            field(mask["provider_party_reference_id"], in.ProviderPartyReferenceId),
		JurisdictionCode:                    field(mask["jurisdiction_code"], in.JurisdictionCode),
		EligibilityRule:                     rule,
		EmployeeContributionRuleReferenceID: field(mask["employee_contribution_rule_reference_id"], in.EmployeeContributionRuleReferenceId),
		EmployerContributionRuleReferenceID: field(mask["employer_contribution_rule_reference_id"], in.EmployerContributionRuleReferenceId),
		ValidFrom:                           from, ValidTo: to, IsActive: field(mask["is_active"], in.IsActive),
	}))
}
func (h *Handler) DeleteBenefitPlan(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteBenefitPlan(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetBenefitPlanByID(ctx context.Context, request *common.IDRequest) (*pb.BenefitPlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return planResponse(h.service.GetBenefitPlanByID(ctx, request.Id))
}
func (h *Handler) GetAllBenefitPlans(ctx context.Context, request *common.PageRequest) (*pb.BenefitPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.GetAllBenefitPlans(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindBenefitPlanByPlanCode(ctx context.Context, request *common.StringFinderRequest) (*pb.BenefitPlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindBenefitPlanByPlanCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindBenefitPlanByBenefitType(ctx context.Context, request *common.StringFinderRequest) (*pb.BenefitPlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindBenefitPlanByBenefitType(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindBenefitPlanByProviderPartyReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.BenefitPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindBenefitPlanByProviderPartyReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindBenefitPlanByJurisdictionCode(ctx context.Context, request *common.StringFinderRequest) (*pb.BenefitPlansResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindBenefitPlanByJurisdictionCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindBenefitPlanByEmployeeContributionRuleReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.BenefitPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindBenefitPlanByEmployeeContributionRuleReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindBenefitPlanByEmployerContributionRuleReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.BenefitPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(h.service.FindBenefitPlanByEmployerContributionRuleReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
