package human_resource

import (
	"bytes"
	"context"
	"encoding/base64"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/services/human_resource/contract"
	payroll "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/payroll/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const payslipReadSelf = authz.PermPayslipsReadSelf

func (a *Adapter) authorizeMyPayslip(ctx context.Context) (context.Context, error) {
	if a == nil || a.authorizer == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	session, err := a.authorizer.RequireSelfPermission(ctx, payslipReadSelf, false)
	if err != nil {
		return nil, err
	}
	scope := authz.ScopeGrant{Type: authz.ScopeSelf, ID: session.Token.UserID}
	if session.Caller.UserKind != "employee" || !slices.ContainsFunc(session.Snapshot.Permissions, func(grant authz.PermissionGrant) bool {
		return strings.EqualFold(grant.Code, payslipReadSelf) && grant.Scope.Type == authz.ScopeSelf && grant.Scope.Matches(scope)
	}) {
		return nil, middleware.ErrForbidden("employee self permission required", nil)
	}
	return middleware.WithResolvedSession(ctx, session), nil
}

// MyPayslip membaca metadata slip yang telah diperiksa oleh HRM.
func (a *Adapter) MyPayslip(ctx context.Context, resultID string) (*model.MyPayslip, error) {
	response, err := a.readMyPayslip(ctx, resultID, false)
	if err != nil {
		return nil, err
	}
	return toModelPayslip(response)
}

// MyPayslipPDF mengembalikan konten versi yang lolos pemeriksaan pemilik di HRM.
func (a *Adapter) MyPayslipPDF(ctx context.Context, resultID string) (*model.MyPayslipPDF, error) {
	response, err := a.readMyPayslip(ctx, resultID, true)
	if err != nil {
		return nil, err
	}
	if _, err := toModelPayslip(response); err != nil {
		return nil, err
	}
	pdf := response.GetPdf()
	if len(pdf) > contract.MaxPayslipPDFBytes || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		return nil, middleware.ErrUnavailable("invalid human resource payslip PDF", nil)
	}
	return &model.MyPayslipPDF{ContentBase64: base64.StdEncoding.EncodeToString(pdf)}, nil
}

func (a *Adapter) readMyPayslip(ctx context.Context, resultID string, includePDF bool) (*payroll.GetMyPayslipResponse, error) {
	rpcCtx, err := a.authorizeMyPayslip(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(resultID, "payroll-result")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payroll result ID", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.payroll.GetMyPayslip(rpcCtx, &payroll.GetMyPayslipRequest{PayrollResultId: id.String(), IncludePdf: includePDF})
	if err != nil {
		return nil, mapReadError(err, "payslip")
	}
	if response.GetPayrollResultId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource payslip response", nil)
	}
	return response, nil
}

// MyPayslips membaca daftar slip milik identitas dari sesi terverifikasi.
func (a *Adapter) MyPayslips(ctx context.Context, periodID *string, limit, offset int) (*Page[model.MyPayslip], error) {
	rpcCtx, err := a.authorizeMyPayslip(ctx)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid payslip pagination", nil)
	}
	request := &payroll.ListMyPayslipsRequest{Limit: int32(limit), Offset: int32(offset)}
	if periodID != nil {
		id, err := parseNodeID(*periodID, "payroll-period")
		if err != nil || id == uuid.Nil {
			return nil, middleware.ErrInvalidArgument("invalid payroll period ID", nil)
		}
		canonical := id.String()
		request.PayrollPeriodId = &canonical
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.payroll.ListMyPayslips(rpcCtx, request)
	if err != nil {
		return nil, mapReadError(err, "payslip")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPublishedPayslipPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource payslip page", nil)
	}
	result := &Page[model.MyPayslip]{
		Items: make([]*model.MyPayslip, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	seenResults := make(map[uuid.UUID]struct{}, len(items))
	for _, slip := range items {
		if request.PayrollPeriodId != nil && slip.GetPayrollPeriodId() != request.GetPayrollPeriodId() {
			return nil, middleware.ErrUnavailable("invalid human resource payslip period", nil)
		}
		item, err := toModelPayslip(slip)
		if err != nil {
			return nil, err
		}
		resultID, _ := uuid.Parse(item.PayrollResultID)
		if _, exists := seenResults[resultID]; exists {
			return nil, middleware.ErrUnavailable("duplicate human resource payslip result", nil)
		}
		seenResults[resultID] = struct{}{}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func toModelPayslip(slip *payroll.GetMyPayslipResponse) (*model.MyPayslip, error) {
	for _, raw := range []string{slip.GetPayrollResultId(), slip.GetDocumentVersionId(), slip.GetPayrollPeriodId()} {
		if id, err := uuid.Parse(raw); err != nil || id == uuid.Nil {
			return nil, middleware.ErrUnavailable("invalid human resource payslip reference", nil)
		}
	}
	if err := slip.GetPublishedAt().CheckValid(); err != nil {
		return nil, middleware.ErrUnavailable("invalid human resource payslip publication", nil)
	}
	start, startErr := time.Parse(time.DateOnly, slip.GetPeriodStart())
	end, endErr := time.Parse(time.DateOnly, slip.GetPeriodEnd())
	if startErr != nil || endErr != nil || end.Before(start) {
		return nil, middleware.ErrUnavailable("invalid human resource payslip period dates", nil)
	}
	if _, err := time.Parse(time.DateOnly, slip.GetScheduledPayDate()); err != nil {
		return nil, middleware.ErrUnavailable("invalid human resource payslip scheduled pay date", nil)
	}
	switch slip.GetPaymentStatus() {
	case "scheduled", "processing", "paid", "failed":
	default:
		return nil, middleware.ErrUnavailable("invalid human resource payslip payment status", nil)
	}
	return &model.MyPayslip{
		PayrollResultID: slip.GetPayrollResultId(), DocumentVersionID: slip.GetDocumentVersionId(),
		PublishedAt: slip.GetPublishedAt().AsTime(), ScheduledPayDate: slip.GetScheduledPayDate(),
		PaymentStatus: slip.GetPaymentStatus(), PayrollPeriodID: slip.GetPayrollPeriodId(),
		PeriodCode: slip.GetPeriodCode(), PeriodStart: slip.GetPeriodStart(), PeriodEnd: slip.GetPeriodEnd(),
	}, nil
}

// ReissueEmployeePayslip meneruskan cetak ulang slip melalui RPC privat HRM.
func (a *Adapter) ReissueEmployeePayslip(ctx context.Context, workerID, resultID string) (*model.MyPayslipPDF, error) {
	rpcCtx, err := a.authorizeScopedSensitiveEmployee(ctx, "human_resource.payslips.reissue", "legal_entity", "legal entity payslip reissue permission required")
	if err != nil {
		return nil, err
	}
	worker, workerErr := uuid.Parse(strings.TrimPrefix(workerID, workerGIDPrefix))
	result, resultErr := parseNodeID(resultID, "payroll-result")
	if workerErr != nil || resultErr != nil || worker == uuid.Nil || result == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid payslip reissue references", nil)
	}
	if a.payroll == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.payroll.ReissueEmployeePayslip(rpcCtx, &payroll.ReissueEmployeePayslipRequest{
		WorkerId: worker.String(), PayrollResultId: result.String(),
	})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("payslip does not meet reissue requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "payslip")
	}
	version, err := uuid.Parse(response.GetDocumentVersionId())
	pdf := response.GetPdf()
	if err != nil || version == uuid.Nil || len(pdf) > contract.MaxPayslipPDFBytes || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		return nil, middleware.ErrUnavailable("invalid human resource payslip PDF", nil)
	}
	return &model.MyPayslipPDF{ContentBase64: base64.StdEncoding.EncodeToString(pdf)}, nil
}
