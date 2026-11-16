package payroll

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func payComponent(v *entity.PayComponent) *pb.PayComponent {
	return &pb.PayComponent{
		Id: v.ID, ComponentCode: v.ComponentCode, ComponentVersion: v.ComponentVersion, ComponentName: v.ComponentName, ComponentType: v.ComponentType,
		CalculationMethod: v.CalculationMethod, TaxTreatmentCode: v.TaxTreatmentCode, JurisdictionCode: v.JurisdictionCode, CurrencyCode: v.CurrencyCode,
		RuleSetId: v.RuleSetID, FormulaReferenceId: v.FormulaReferenceID, Taxable: v.Taxable, Pensionable: v.Pensionable,
		EffectiveFrom: v.EffectiveFrom.Format("2006-01-02"), EffectiveTo: dateString(v.EffectiveTo), IsActive: v.IsActive,
	}
}
func payComponents(page operation.Page[entity.PayComponent], err error) (*pb.PayComponentsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PayComponent, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, payComponent(&page.Items[i]))
	}
	return &pb.PayComponentsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreatePayComponent(ctx context.Context, r *pb.CreatePayComponentRequest) (*pb.PayComponentResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	from, err := response.Date(in.GetEffectiveFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.CreatePayComponent(ctx, service.CreatePayComponentInput{
		ComponentCode: in.GetComponentCode(), ComponentVersion: in.GetComponentVersion(), ComponentName: in.GetComponentName(), ComponentType: in.GetComponentType(),
		CalculationMethod: in.GetCalculationMethod(), TaxTreatmentCode: in.TaxTreatmentCode, JurisdictionCode: in.GetJurisdictionCode(), CurrencyCode: in.CurrencyCode,
		RuleSetID: in.RuleSetId, FormulaReferenceID: in.FormulaReferenceId, Taxable: in.Taxable, Pensionable: in.Pensionable, EffectiveFrom: from, EffectiveTo: to, IsActive: in.IsActive,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayComponentResponse{Item: payComponent(v)}, nil
}
func (h *Handler) UpdatePayComponent(ctx context.Context, r *pb.UpdatePayComponentRequest) (*pb.PayComponentResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "component_code", "component_version", "component_name", "component_type", "calculation_method", "tax_treatment_code", "jurisdiction_code", "currency_code", "rule_set_id", "formula_reference_id", "taxable", "pensionable", "effective_from", "effective_to", "is_active")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	from, err := dateField(m["effective_from"], in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(m["effective_to"], in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.UpdatePayComponent(ctx, r.Id, service.UpdatePayComponentInput{
		ComponentCode: field(m["component_code"], in.ComponentCode), ComponentVersion: field(m["component_version"], in.ComponentVersion), ComponentName: field(m["component_name"], in.ComponentName),
		ComponentType: field(m["component_type"], in.ComponentType), CalculationMethod: field(m["calculation_method"], in.CalculationMethod), TaxTreatmentCode: field(m["tax_treatment_code"], in.TaxTreatmentCode),
		JurisdictionCode: field(m["jurisdiction_code"], in.JurisdictionCode), CurrencyCode: field(m["currency_code"], in.CurrencyCode), RuleSetID: field(m["rule_set_id"], in.RuleSetId),
		FormulaReferenceID: field(m["formula_reference_id"], in.FormulaReferenceId), Taxable: field(m["taxable"], in.Taxable), Pensionable: field(m["pensionable"], in.Pensionable),
		EffectiveFrom: from, EffectiveTo: to, IsActive: field(m["is_active"], in.IsActive),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayComponentResponse{Item: payComponent(v)}, nil
}
func (h *Handler) DeletePayComponent(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeletePayComponent(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetPayComponentByID(ctx context.Context, r *common.IDRequest) (*pb.PayComponentResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetPayComponentByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayComponentResponse{Item: payComponent(v)}, nil
}
func (h *Handler) GetAllPayComponents(ctx context.Context, r *common.PageRequest) (*pb.PayComponentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payComponents(h.service.GetAllPayComponents(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayComponentByComponentCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayComponentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payComponents(h.service.FindPayComponentByComponentCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayComponentByComponentType(ctx context.Context, r *common.StringFinderRequest) (*pb.PayComponentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payComponents(h.service.FindPayComponentByComponentType(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayComponentByTaxTreatmentCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayComponentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payComponents(h.service.FindPayComponentByTaxTreatmentCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayComponentByJurisdictionCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayComponentsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payComponents(h.service.FindPayComponentByJurisdictionCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayComponentByCurrencyCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayComponentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payComponents(h.service.FindPayComponentByCurrencyCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayComponentByRuleSetID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayComponentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payComponents(h.service.FindPayComponentByRuleSetID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayComponentByFormulaReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayComponentsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payComponents(h.service.FindPayComponentByFormulaReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
