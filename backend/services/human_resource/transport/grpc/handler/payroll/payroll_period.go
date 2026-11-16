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

func payrollPeriod(v *entity.PayrollPeriod) *pb.PayrollPeriod {
	return &pb.PayrollPeriod{
		Id: v.ID, PayrollGroupId: v.PayrollGroupID, EmployingLegalEntityReferenceId: v.EmployingLegalEntityReferenceID, CurrencyCode: v.CurrencyCode,
		JurisdictionCode: v.JurisdictionCode, PeriodCode: v.PeriodCode, PeriodStart: v.PeriodStart.Format("2006-01-02"), PeriodEnd: v.PeriodEnd.Format("2006-01-02"),
		PayDate: v.PayDate.Format("2006-01-02"), PeriodStatus: v.PeriodStatus, CalendarReferenceId: v.CalendarReferenceID,
	}
}
func payrollPeriods(page operation.Page[entity.PayrollPeriod], err error) (*pb.PayrollPeriodsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PayrollPeriod, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, payrollPeriod(&page.Items[i]))
	}
	return &pb.PayrollPeriodsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreatePayrollPeriod(ctx context.Context, r *pb.CreatePayrollPeriodRequest) (*pb.PayrollPeriodResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	start, err := response.Date(in.GetPeriodStart())
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := response.Date(in.GetPeriodEnd())
	if err != nil {
		return nil, response.Error(err)
	}
	pay, err := response.Date(in.GetPayDate())
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.CreatePayrollPeriod(ctx, service.CreatePayrollPeriodInput{
		PayrollGroupID: in.GetPayrollGroupId(), EmployingLegalEntityReferenceID: in.GetEmployingLegalEntityReferenceId(), CurrencyCode: in.GetCurrencyCode(),
		JurisdictionCode: in.GetJurisdictionCode(), PeriodCode: in.GetPeriodCode(), PeriodStart: start, PeriodEnd: end, PayDate: pay, PeriodStatus: in.PeriodStatus, CalendarReferenceID: in.CalendarReferenceId,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollPeriodResponse{Item: payrollPeriod(v)}, nil
}
func (h *Handler) UpdatePayrollPeriod(ctx context.Context, r *pb.UpdatePayrollPeriodRequest) (*pb.PayrollPeriodResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "payroll_group_id", "employing_legal_entity_reference_id", "currency_code", "jurisdiction_code", "period_code", "period_start", "period_end", "pay_date", "period_status", "calendar_reference_id")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	start, err := dateField(m["period_start"], in.PeriodStart)
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := dateField(m["period_end"], in.PeriodEnd)
	if err != nil {
		return nil, response.Error(err)
	}
	pay, err := dateField(m["pay_date"], in.PayDate)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.UpdatePayrollPeriod(ctx, r.Id, service.UpdatePayrollPeriodInput{
		PayrollGroupID: field(m["payroll_group_id"], in.PayrollGroupId), EmployingLegalEntityReferenceID: field(m["employing_legal_entity_reference_id"], in.EmployingLegalEntityReferenceId),
		CurrencyCode: field(m["currency_code"], in.CurrencyCode), JurisdictionCode: field(m["jurisdiction_code"], in.JurisdictionCode), PeriodCode: field(m["period_code"], in.PeriodCode),
		PeriodStart: start, PeriodEnd: end, PayDate: pay, PeriodStatus: field(m["period_status"], in.PeriodStatus), CalendarReferenceID: field(m["calendar_reference_id"], in.CalendarReferenceId),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollPeriodResponse{Item: payrollPeriod(v)}, nil
}
func (h *Handler) DeletePayrollPeriod(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeletePayrollPeriod(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetPayrollPeriodByID(ctx context.Context, r *common.IDRequest) (*pb.PayrollPeriodResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetPayrollPeriodByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollPeriodResponse{Item: payrollPeriod(v)}, nil
}
func (h *Handler) GetAllPayrollPeriods(ctx context.Context, r *common.PageRequest) (*pb.PayrollPeriodsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollPeriods(h.service.GetAllPayrollPeriods(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollPeriodByPayrollGroupID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollPeriodsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollPeriods(h.service.FindPayrollPeriodByPayrollGroupID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollPeriodByEmployingLegalEntityReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollPeriodsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollPeriods(h.service.FindPayrollPeriodByEmployingLegalEntityReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollPeriodByCurrencyCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollPeriodsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollPeriods(h.service.FindPayrollPeriodByCurrencyCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollPeriodByJurisdictionCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollPeriodsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollPeriods(h.service.FindPayrollPeriodByJurisdictionCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollPeriodByPeriodCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollPeriodsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollPeriods(h.service.FindPayrollPeriodByPeriodCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollPeriodByPeriodStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollPeriodsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollPeriods(h.service.FindPayrollPeriodByPeriodStatus(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollPeriodByCalendarReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollPeriodsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollPeriods(h.service.FindPayrollPeriodByCalendarReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
