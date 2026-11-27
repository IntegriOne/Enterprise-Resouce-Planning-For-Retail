package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
)

const payrollRuleSetReadPermission = "human_resource.payroll_rule_sets.read"

// PayrollRuleSet membaca detail berdasarkan ID.
func (a *Adapter) PayrollRuleSet(ctx context.Context, rawID string) (*model.PayrollRuleSet, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, payrollRuleSetReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "payroll-rule-set")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll rule set ID", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.payroll.GetPayrollRuleSetByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "payroll rule set")
	}
	item := response.GetItem()
	if item.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource payroll rule set response", nil)
	}
	return toModelPayrollRuleSet(item)
}

// PayrollRuleSets mendaftar aturan payroll dengan izin baca tersendiri.
func (a *Adapter) PayrollRuleSets(ctx context.Context, limit, offset int) (*Page[model.PayrollRuleSet], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, payrollRuleSetReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll rule set pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.payroll.GetAllPayrollRuleSets(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, mapReadError(err, "payroll rule set")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource payroll rule set page", nil)
	}
	result := &Page[model.PayrollRuleSet]{
		Items: make([]*model.PayrollRuleSet, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, rule := range items {
		item, err := toModelPayrollRuleSet(rule)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelPayrollRuleSet(rule *payroll.PayrollRuleSet) (*model.PayrollRuleSet, error) {
	if id, err := uuid.Parse(rule.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource payroll rule set reference", nil)
	}
	item := &model.PayrollRuleSet{
		ID: nodeID("payroll-rule-set", rule.GetId()), RawID: rule.GetId(), RuleCode: rule.GetRuleCode(), RuleVersion: rule.GetRuleVersion(),
		JurisdictionCode: rule.GetJurisdictionCode(), EffectiveFrom: rule.GetEffectiveFrom(), EffectiveTo: rule.EffectiveTo,
	}
	if approved := rule.GetApprovedAt(); approved != nil {
		if err := approved.CheckValid(); err != nil {
			return nil, middleware.ErrUnavailable("invalid human resource payroll rule set approval", nil)
		}
		at := approved.AsTime()
		item.ApprovedAt = &at
	}
	return item, nil
}
