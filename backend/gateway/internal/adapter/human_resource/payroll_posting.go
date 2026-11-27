package human_resource

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const payrollPostingPermission = "human_resource.payroll_runs.request_posting"

// RequestPayrollPosting mengajukan jurnal untuk run yang telah disetujui.
func (a *Adapter) RequestPayrollPosting(ctx context.Context, rawID string) (*model.PayrollPostingRequest, error) {
	run, err := a.runPayrollCommand(ctx, rawID, payrollPostingPermission, "approved", payroll.PayrollServiceClient.RequestPayrollPosting)
	if err != nil {
		return nil, err
	}
	at := run.GetPostingRequestedAt()
	if at.CheckValid() != nil {
		return nil, middleware.ErrUnavailable("invalid payroll posting request response", nil)
	}
	return &model.PayrollPostingRequest{PayrollRunID: run.GetId(), Status: run.GetStatus(), PostingRequestedAt: at.AsTime()}, nil
}

// CalculatePayrollRun menjalankan kalkulasi draf melalui adapter HRM yang terkonfigurasi.
func (a *Adapter) CalculatePayrollRun(ctx context.Context, id string) (*model.PayrollRun, error) {
	run, err := a.runPayrollCommand(ctx, id, "human_resource.payroll_runs.calculate", "calculated", payroll.PayrollServiceClient.CalculatePayrollRun)
	if err != nil {
		return nil, err
	}
	if run.GetCalculatedAt() == nil {
		return nil, middleware.ErrUnavailable("invalid payroll calculation response", nil)
	}
	return toModelPayrollRun(run)
}

// SubmitPayrollRunForReview mengajukan hasil kalkulasi untuk review.
func (a *Adapter) SubmitPayrollRunForReview(ctx context.Context, id string) (*model.PayrollRun, error) {
	run, err := a.runPayrollCommand(ctx, id, "human_resource.payroll_runs.submit_review", "review", payroll.PayrollServiceClient.SubmitPayrollRunForReview)
	if err != nil {
		return nil, err
	}
	return toModelPayrollRun(run)
}

// ApprovePayrollRun meminta keputusan persetujuan setelah review HRM.
func (a *Adapter) ApprovePayrollRun(ctx context.Context, id string) (*model.PayrollRun, error) {
	run, err := a.runPayrollCommand(ctx, id, "human_resource.payroll_runs.approve", "approved", payroll.PayrollServiceClient.ApprovePayrollRun)
	if err != nil {
		return nil, err
	}
	return toModelPayrollRun(run)
}

// CancelPayrollRun meminta pembatalan run yang sedang direview.
func (a *Adapter) CancelPayrollRun(ctx context.Context, id, reason string) (*model.PayrollRun, error) {
	run, err := a.runPayrollCommand(ctx, id, "human_resource.payroll_runs.cancel", "cancelled",
		func(client payroll.PayrollServiceClient, ctx context.Context, request *common.IDRequest, _ ...grpc.CallOption) (*payroll.PayrollRunResponse, error) {
			reason := strings.TrimSpace(reason)
			if reason == "" {
				return nil, status.Error(codes.InvalidArgument, "cancellation reason is required")
			}
			return client.CancelPayrollRun(ctx, &payroll.CancelPayrollRunRequest{Id: request.GetId(), Reason: reason})
		})
	if err != nil {
		return nil, err
	}
	if run.GetCancelledAt() == nil {
		return nil, middleware.ErrUnavailable("invalid payroll cancellation response", nil)
	}
	return toModelPayrollRun(run)
}

func (a *Adapter) runPayrollCommand(
	ctx context.Context,
	rawID, permission, expectedStatus string,
	command func(payroll.PayrollServiceClient, context.Context, *common.IDRequest, ...grpc.CallOption) (*payroll.PayrollRunResponse, error),
) (*payroll.PayrollRun, error) {
	rpcCtx, err := a.authorizeSensitivePayroll(ctx, permission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "payroll-run")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll run ID", nil)
	}
	response, err := command(a.payroll, rpcCtx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("payroll run does not meet workflow requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "payroll run")
	}
	run := response.GetItem()
	if run.GetId() != id.String() || run.GetStatus() != expectedStatus {
		return nil, middleware.ErrUnavailable("invalid human resource payroll run response", nil)
	}
	return run, nil
}

func (a *Adapter) authorizeSensitivePayroll(ctx context.Context, permission string) (context.Context, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, permission)
	if err != nil {
		return nil, err
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return rpcCtx, nil
}

func (a *Adapter) authorizeSensitiveEmployee(ctx context.Context, permission string) (context.Context, error) {
	if a == nil || a.authorizer == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	session, err := a.authorizer.RequirePermission(ctx, permission, authz.ScopeGrant{Type: authz.ScopeGlobal}, true)
	if err != nil {
		return nil, err
	}
	if session.Caller.UserKind != "employee" || !session.Token.MFAAuthenticated {
		return nil, middleware.ErrForbidden("employee MFA authentication required", nil)
	}
	return middleware.WithResolvedSession(ctx, session), nil
}
