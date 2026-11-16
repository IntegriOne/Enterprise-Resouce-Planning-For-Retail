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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func payrollRun(v *entity.PayrollRun) *pb.PayrollRun {
	return &pb.PayrollRun{
		Id: v.ID, PayrollPeriodId: v.PayrollPeriodID, PayrollGroupId: v.PayrollGroupID, EmployingLegalEntityReferenceId: v.EmployingLegalEntityReferenceID,
		CurrencyCode: v.CurrencyCode, RuleSetId: v.RuleSetID, OriginalRunId: v.OriginalRunID, SettlementDirection: v.SettlementDirection,
		RunNumber: v.RunNumber, RunType: v.RunType, Status: string(v.Status), JurisdictionCode: v.JurisdictionCode, RuleEngineName: v.RuleEngineName,
		RuleEngineVersion: v.RuleEngineVersion, InputSnapshotReferenceId: v.InputSnapshotReferenceID, InputDigest: append([]byte(nil), v.InputDigest...),
		OutputDigest: append([]byte(nil), v.OutputDigest...), CalculatedAt: timestamp(v.CalculatedAt), ApprovedByPrincipalId: v.ApprovedByPrincipalID,
		ApprovedAt: timestamp(v.ApprovedAt), FinanceJournalReferenceId: v.FinanceJournalReferenceID, PaymentBatchReferenceId: v.PaymentBatchReferenceID,
		FinanceSettlementEventReferenceId: v.FinanceSettlementEventReferenceID, FinanceSettledAt: timestamp(v.FinanceSettledAt),
		StartedByPrincipalId: v.StartedByPrincipalID, CreatedAt: timestamppb.New(v.CreatedAt), IdempotencyKey: v.IdempotencyKey,
	}
}
func payrollRuns(page operation.Page[entity.PayrollRun], err error) (*pb.PayrollRunsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.PayrollRun, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, payrollRun(&page.Items[i]))
	}
	return &pb.PayrollRunsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreatePayrollRun(ctx context.Context, r *pb.CreatePayrollRunRequest) (*pb.PayrollRunResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	calculated, err := response.OptionalTimestamp(in.CalculatedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	settled, err := response.OptionalTimestamp(in.FinanceSettledAt)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.CreatePayrollRun(ctx, service.CreatePayrollRunInput{
		PayrollPeriodID: in.GetPayrollPeriodId(), PayrollGroupID: in.GetPayrollGroupId(), EmployingLegalEntityReferenceID: in.GetEmployingLegalEntityReferenceId(),
		CurrencyCode: in.GetCurrencyCode(), RuleSetID: in.GetRuleSetId(), OriginalRunID: in.OriginalRunId, SettlementDirection: in.SettlementDirection,
		RunNumber: in.GetRunNumber(), RunType: in.GetRunType(), Status: runStatus(in.Status), JurisdictionCode: in.GetJurisdictionCode(),
		RuleEngineName: in.GetRuleEngineName(), RuleEngineVersion: in.GetRuleEngineVersion(), InputSnapshotReferenceID: in.GetInputSnapshotReferenceId(),
		InputDigest: append([]byte(nil), in.InputDigest...), OutputDigest: append([]byte(nil), in.OutputDigest...), CalculatedAt: calculated,
		ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approved, FinanceJournalReferenceID: in.FinanceJournalReferenceId,
		PaymentBatchReferenceID: in.PaymentBatchReferenceId, FinanceSettlementEventReferenceID: in.FinanceSettlementEventReferenceId, FinanceSettledAt: settled, IdempotencyKey: in.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollRunResponse{Item: payrollRun(v)}, nil
}
func (h *Handler) UpdatePayrollRun(ctx context.Context, r *pb.UpdatePayrollRunRequest) (*pb.PayrollRunResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask,
		"payroll_period_id", "payroll_group_id", "employing_legal_entity_reference_id", "currency_code", "rule_set_id", "original_run_id",
		"settlement_direction", "run_number", "run_type", "status", "jurisdiction_code", "rule_engine_name", "rule_engine_version",
		"input_snapshot_reference_id", "input_digest", "output_digest", "calculated_at", "approved_by_principal_id", "approved_at",
		"finance_journal_reference_id", "payment_batch_reference_id", "finance_settlement_event_reference_id", "finance_settled_at", "idempotency_key",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	calculated, err := timeField(m["calculated_at"], in.CalculatedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := timeField(m["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	settled, err := timeField(m["finance_settled_at"], in.FinanceSettledAt)
	if err != nil {
		return nil, response.Error(err)
	}
	v, err := h.service.UpdatePayrollRun(ctx, r.Id, service.UpdatePayrollRunInput{
		PayrollPeriodID: field(m["payroll_period_id"], in.PayrollPeriodId), PayrollGroupID: field(m["payroll_group_id"], in.PayrollGroupId),
		EmployingLegalEntityReferenceID: field(m["employing_legal_entity_reference_id"], in.EmployingLegalEntityReferenceId), CurrencyCode: field(m["currency_code"], in.CurrencyCode),
		RuleSetID: field(m["rule_set_id"], in.RuleSetId), OriginalRunID: field(m["original_run_id"], in.OriginalRunId), SettlementDirection: field(m["settlement_direction"], in.SettlementDirection),
		RunNumber: field(m["run_number"], in.RunNumber), RunType: field(m["run_type"], in.RunType), Status: field(m["status"], runStatus(in.Status)),
		JurisdictionCode: field(m["jurisdiction_code"], in.JurisdictionCode), RuleEngineName: field(m["rule_engine_name"], in.RuleEngineName), RuleEngineVersion: field(m["rule_engine_version"], in.RuleEngineVersion),
		InputSnapshotReferenceID: field(m["input_snapshot_reference_id"], in.InputSnapshotReferenceId), InputDigest: bytesField(m["input_digest"], in.InputDigest), OutputDigest: bytesField(m["output_digest"], in.OutputDigest),
		CalculatedAt: calculated, ApprovedByPrincipalID: field(m["approved_by_principal_id"], in.ApprovedByPrincipalId), ApprovedAt: approved,
		FinanceJournalReferenceID: field(m["finance_journal_reference_id"], in.FinanceJournalReferenceId), PaymentBatchReferenceID: field(m["payment_batch_reference_id"], in.PaymentBatchReferenceId),
		FinanceSettlementEventReferenceID: field(m["finance_settlement_event_reference_id"], in.FinanceSettlementEventReferenceId), FinanceSettledAt: settled, IdempotencyKey: field(m["idempotency_key"], in.IdempotencyKey),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollRunResponse{Item: payrollRun(v)}, nil
}
func (h *Handler) DeletePayrollRun(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeletePayrollRun(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetPayrollRunByID(ctx context.Context, r *common.IDRequest) (*pb.PayrollRunResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.GetPayrollRunByID(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollRunResponse{Item: payrollRun(v)}, nil
}
func (h *Handler) GetAllPayrollRuns(ctx context.Context, r *common.PageRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.GetAllPayrollRuns(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) ApprovePayrollRun(ctx context.Context, r *common.IDRequest) (*pb.PayrollRunResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	v, err := h.service.ApprovePayrollRun(ctx, r.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PayrollRunResponse{Item: payrollRun(v)}, nil
}
func (h *Handler) FindPayrollRunByPayrollPeriodID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByPayrollPeriodID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByPayrollGroupID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByPayrollGroupID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByEmployingLegalEntityReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByEmployingLegalEntityReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByCurrencyCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByCurrencyCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByRuleSetID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByRuleSetID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByOriginalRunID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByOriginalRunID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByRunNumber(ctx context.Context, r *common.Int32FinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByRunNumber(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByRunType(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByRunType(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByStatus(ctx, entity.PayrollRunStatus(*r.Value), int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByJurisdictionCode(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByJurisdictionCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByInputSnapshotReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByInputSnapshotReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByApprovedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByApprovedByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByFinanceJournalReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByFinanceJournalReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByPaymentBatchReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByPaymentBatchReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByFinanceSettlementEventReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByFinanceSettlementEventReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByStartedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByStartedByPrincipalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindPayrollRunByIdempotencyKey(ctx context.Context, r *common.StringFinderRequest) (*pb.PayrollRunsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return payrollRuns(h.service.FindPayrollRunByIdempotencyKey(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
