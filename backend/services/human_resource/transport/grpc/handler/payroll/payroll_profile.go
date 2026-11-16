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

func payrollProfile(v *entity.PayrollProfile) *pb.PayrollProfile {
	return &pb.PayrollProfile{
		Id: v.ID, WorkerId: v.WorkerID, PayrollCountryCode: v.PayrollCountryCode, PayrollJurisdictionCode: v.PayrollJurisdictionCode,
		PayrollGroupId: v.PayrollGroupID, EmployingLegalEntityReferenceId: v.EmployingLegalEntityReferenceID, CurrencyCode: v.CurrencyCode,
		TaxProfileReferenceId: v.TaxProfileReferenceID, BankAccountTokenReferenceId: v.BankAccountTokenReferenceID, TaxStatusCode: v.TaxStatusCode,
		WithholdingElectionReferenceId: v.WithholdingElectionReferenceID, EffectiveFrom: v.EffectiveFrom.Format("2006-01-02"), EffectiveTo: dateString(v.EffectiveTo), IsActive: v.IsActive,
	}
}
func payrollProfiles(page operation.Page[entity.PayrollProfile], err error) (*pb.PayrollProfilesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PayrollProfile, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, payrollProfile(&page.Items[i]))
	}
	return &pb.PayrollProfilesResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreatePayrollProfile(ctx context.Context, r *pb.CreatePayrollProfileRequest) (*pb.PayrollProfileResponse, error) {
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
	v, err := h.service.CreatePayrollProfile(ctx, service.CreatePayrollProfileInput{
		WorkerID: in.GetWorkerId(), PayrollCountryCode: in.GetPayrollCountryCode(), PayrollJurisdictionCode: in.GetPayrollJurisdictionCode(),
		PayrollGroupID: in.GetPayrollGroupId(), EmployingLegalEntityReferenceID: in.GetEmployingLegalEntityReferenceId(), CurrencyCode: in.GetCurrencyCode(),
		TaxProfileReferenceID: in.TaxProfileReferenceId, BankAccountTokenReferenceID: in.BankAccountTokenReferenceId, TaxStatusCode: in.TaxStatusCode,
		WithholdingElectionReferenceID: in.WithholdingElectionReferenceId, EffectiveFrom: from, EffectiveTo: to, IsActive: in.IsActive,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollProfileResponse{Item: payrollProfile(v)}, nil
}
func (h *Handler) UpdatePayrollProfile(ctx context.Context, r *pb.UpdatePayrollProfileRequest) (*pb.PayrollProfileResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "worker_id", "payroll_country_code", "payroll_jurisdiction_code", "payroll_group_id", "employing_legal_entity_reference_id", "currency_code", "tax_profile_reference_id", "bank_account_token_reference_id", "tax_status_code", "withholding_election_reference_id", "effective_from", "effective_to", "is_active")
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
	v, err := h.service.UpdatePayrollProfile(ctx, r.Id, service.UpdatePayrollProfileInput{
		WorkerID: field(m["worker_id"], in.WorkerId), PayrollCountryCode: field(m["payroll_country_code"], in.PayrollCountryCode),
		PayrollJurisdictionCode: field(m["payroll_jurisdiction_code"], in.PayrollJurisdictionCode), PayrollGroupID: field(m["payroll_group_id"], in.PayrollGroupId),
		EmployingLegalEntityReferenceID: field(m["employing_legal_entity_reference_id"], in.EmployingLegalEntityReferenceId), CurrencyCode: field(m["currency_code"], in.CurrencyCode),
		TaxProfileReferenceID: field(m["tax_profile_reference_id"], in.TaxProfileReferenceId), BankAccountTokenReferenceID: field(m["bank_account_token_reference_id"], in.BankAccountTokenReferenceId),
		TaxStatusCode: field(m["tax_status_code"], in.TaxStatusCode), WithholdingElectionReferenceID: field(m["withholding_election_reference_id"], in.WithholdingElectionReferenceId),
		EffectiveFrom: from, EffectiveTo: to, IsActive: field(m["is_active"], in.IsActive),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollProfileResponse{Item: payrollProfile(v)}, nil
}
func (h *Handler) DeletePayrollProfile(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeletePayrollProfile(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetPayrollProfileByID(ctx context.Context, r *common.IDRequest) (*pb.PayrollProfileResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetPayrollProfileByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollProfileResponse{Item: payrollProfile(v)}, nil
}
func (h *Handler) GetAllPayrollProfiles(ctx context.Context, r *common.PageRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.GetAllPayrollProfiles(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollProfileByWorkerID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.FindPayrollProfileByWorkerID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollProfileByPayrollCountryCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.FindPayrollProfileByPayrollCountryCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollProfileByPayrollJurisdictionCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.FindPayrollProfileByPayrollJurisdictionCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollProfileByPayrollGroupID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.FindPayrollProfileByPayrollGroupID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollProfileByEmployingLegalEntityReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.FindPayrollProfileByEmployingLegalEntityReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollProfileByCurrencyCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.FindPayrollProfileByCurrencyCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollProfileByTaxProfileReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.FindPayrollProfileByTaxProfileReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollProfileByBankAccountTokenReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.FindPayrollProfileByBankAccountTokenReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollProfileByTaxStatusCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.FindPayrollProfileByTaxStatusCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollProfileByWithholdingElectionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollProfilesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollProfiles(h.service.FindPayrollProfileByWithholdingElectionReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
