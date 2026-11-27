package human_resource

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const payrollRunReadPermission = "human_resource.payroll_runs.read"

func (a *Adapter) authorizeGlobalEmployee(ctx context.Context, permission string) (context.Context, error) {
	if a == nil || a.authorizer == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	session, err := a.authorizer.RequirePermission(ctx, permission, authz.ScopeGrant{Type: authz.ScopeGlobal}, false)
	if err != nil {
		return nil, err
	}
	if session.Caller.UserKind != "employee" {
		return nil, middleware.ErrForbidden("employee session required", nil)
	}
	return middleware.WithResolvedSession(ctx, session), nil
}

// PayrollRun membaca status agregat run dari HRM.
func (a *Adapter) PayrollRun(ctx context.Context, rawID string) (*model.PayrollRun, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, payrollRunReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "payroll-run")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll run ID", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.payroll.GetPayrollRunByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "payroll run")
	}
	run := response.GetItem()
	if run.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource payroll run response", nil)
	}
	return toModelPayrollRun(run)
}

// PayrollRuns membaca halaman status run dari HRM.
func (a *Adapter) PayrollRuns(ctx context.Context, limit, offset int, idempotencyKey, periodID *string) (*Page[model.PayrollRun], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, payrollRunReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payroll run pagination", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	if idempotencyKey != nil && periodID != nil {
		return nil, middleware.ErrInvalidArgument("use only one payroll run filter", nil)
	}
	var response *payroll.PayrollRunsResponse
	var periodFilter string
	switch {
	case periodID != nil:
		id, parseErr := parseNodeID(*periodID, "payroll-period")
		if parseErr != nil || id == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid payroll period ID", nil)
		}
		periodFilter = id.String()
		response, err = a.payroll.FindPayrollRunByPayrollPeriodID(rpcCtx, &common.StringFinderRequest{
			Value: &periodFilter, Limit: int32(limit), Offset: int32(offset),
		})
	case idempotencyKey != nil:
		if !validIdempotencyKey(*idempotencyKey) {
			return nil, middleware.ErrInvalidArgument("invalid payroll idempotency key", nil)
		}
		response, err = a.payroll.FindPayrollRunByIdempotencyKey(rpcCtx, &common.StringFinderRequest{
			Value: idempotencyKey, Limit: int32(limit), Offset: int32(offset),
		})
	default:
		response, err = a.payroll.GetAllPayrollRuns(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "payroll run")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource payroll run page", nil)
	}
	result := &Page[model.PayrollRun]{
		Items: make([]*model.PayrollRun, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, run := range items {
		if (idempotencyKey != nil && run.GetIdempotencyKey() != *idempotencyKey) ||
			(periodID != nil && run.GetPayrollPeriodId() != periodFilter) {
			return nil, middleware.ErrUnavailable("invalid human resource payroll run match", nil)
		}
		item, err := toModelPayrollRun(run)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelPayrollRun(run *payroll.PayrollRun) (*model.PayrollRun, error) {
	for _, raw := range []string{run.GetId(), run.GetPayrollPeriodId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource payroll run reference", nil)
		}
	}
	result := &model.PayrollRun{
		ID: nodeID("payroll-run", run.GetId()), RawID: run.GetId(), PayrollPeriodID: run.GetPayrollPeriodId(), RunNumber: int(run.GetRunNumber()),
		Status: run.GetStatus(), FinanceJournalReferenceID: run.FinanceJournalReferenceId,
	}
	for _, field := range []struct {
		source *timestamppb.Timestamp
		target **time.Time
	}{
		{run.GetCalculatedAt(), &result.CalculatedAt},
		{run.GetPostingRequestedAt(), &result.PostingRequestedAt},
		{run.GetFinanceSettledAt(), &result.FinanceSettledAt},
		{run.GetCancelledAt(), &result.CancelledAt},
	} {
		if field.source == nil {
			continue
		}
		if err := field.source.CheckValid(); err != nil {
			return nil, middleware.ErrUnavailable("invalid human resource payroll run timestamp", nil)
		}
		value := field.source.AsTime()
		*field.target = &value
	}
	return result, nil
}
