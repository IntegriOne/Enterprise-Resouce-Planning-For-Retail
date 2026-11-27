package human_resource

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	recruitment "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const requisitionApprovePermission = "human_resource.requisitions.approve"

// ApproveRequisition meminta HRM mengesahkan pengajuan pada unit peninjau.
func (a *Adapter) ApproveRequisition(ctx context.Context, id string) (*model.Requisition, error) {
	rpcCtx, err := a.authorizeRequisitionApprover(ctx)
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return requisitionDecision(rpcCtx, id, "approved", a.recruitment.ApproveRequisition)
}

// OpenRequisition meminta HRM membuka requisition setelah memeriksa ulang bukti.
func (a *Adapter) OpenRequisition(ctx context.Context, id string) (*model.Requisition, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.requisitions.open")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return requisitionDecision(rpcCtx, id, "open", a.recruitment.OpenRequisition)
}

// CloseRequisition menutup requisition yang seluruh proses aktifnya selesai.
func (a *Adapter) CloseRequisition(ctx context.Context, id string) (*model.Requisition, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.requisitions.close")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return requisitionDecision(rpcCtx, id, "closed", a.recruitment.CloseRequisition)
}

func (a *Adapter) authorizeRequisitionApprover(ctx context.Context) (context.Context, error) {
	return a.authorizeScopedSensitiveEmployee(ctx, requisitionApprovePermission, "organization_unit", "organization unit approval permission required")
}

func (a *Adapter) authorizeScopedSensitiveEmployee(ctx context.Context, permission, scope, missingGrantMessage string) (context.Context, error) {
	if a == nil || a.auth == nil || a.authorizer == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	session, err := a.auth.RequireAuthenticated(ctx)
	if err != nil {
		return nil, err
	}
	if session == nil || session.Token == nil || session.Snapshot == nil {
		return nil, middleware.ErrUnauthenticated("authentication required", nil)
	}
	for _, grant := range session.Snapshot.Permissions {
		id, parseErr := uuid.Parse(grant.Scope.ID)
		if !strings.EqualFold(grant.Code, permission) ||
			grant.Scope.Type != scope || parseErr != nil || id == uuid.Nil {
			continue
		}
		// HRM mencocokkan scope dengan record sebelum menjalankan operasi.
		checked, err := a.authorizer.RequirePermission(ctx, permission, grant.Scope, true)
		if err != nil {
			return nil, err
		}
		if checked.Caller.UserKind != "employee" || !checked.Token.MFAAuthenticated {
			return nil, middleware.ErrForbidden("employee MFA authentication required", nil)
		}
		return middleware.WithResolvedSession(ctx, checked), nil
	}
	return nil, middleware.ErrForbidden(missingGrantMessage, nil)
}

func requisitionDecision(ctx context.Context, rawID, expectedStatus string,
	command func(context.Context, *common.IDRequest, ...grpc.CallOption) (*recruitment.RequisitionResponse, error),
) (*model.Requisition, error) {
	id, err := parseNodeID(rawID, "requisition")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid requisition ID", nil)
	}
	response, err := command(ctx, &common.IDRequest{Id: id.String()})
	if status.Code(err) == codes.FailedPrecondition {
		return nil, middleware.ErrConflict("requisition does not meet decision requirements", err)
	}
	if err != nil {
		return nil, mapReadError(err, "requisition")
	}
	item := response.GetItem()
	if item.GetId() != id.String() || item.GetStatus() != expectedStatus {
		return nil, middleware.ErrUnavailable("invalid human resource requisition response", nil)
	}
	return toModelRequisition(item)
}
