package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// DeleteWorker menghapus data pekerja berdasarkan ID.
func (a *Adapter) DeleteWorker(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_workers", "human_resource.workers.delete", "worker", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.personnel.DeleteWorker(c, req)
	})
}

// UpdateWorkerAssignment memperbarui penugasan pekerja berdasarkan FieldMask.
func (a *Adapter) UpdateWorkerAssignment(ctx context.Context, rawID string, input model.UpdateWorkerAssignmentInput) (*model.WorkerAssignment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_worker_assignments", "update", "human_resource.worker_assignments.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "worker-assignment", "worker-assignment")
	if err != nil {
		return nil, err
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &personnel.WorkerAssignmentInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "position_id", input.PositionID, &patch.PositionId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "manager_worker_id", input.ManagerWorkerID, &patch.ManagerWorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assignment_type", input.AssignmentType, &patch.AssignmentType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "employment_status_snapshot", input.EmploymentStatusSnapshot, &patch.EmploymentStatusSnapshot, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "fte_percent", input.FtePercent, &patch.FtePercent, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "valid_from", input.ValidFrom, &patch.ValidFrom, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "valid_to", input.ValidTo, &patch.ValidTo, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assignment_reason", input.AssignmentReason, &patch.AssignmentReason, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("worker-assignment update is empty", nil)
	}
	resp, err := a.personnel.UpdateWorkerAssignment(rpcCtx, &personnel.UpdateWorkerAssignmentRequest{
		Id:         id,
		Input:      patch,
		UpdateMask: mask,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "worker-assignment")
	}
	return toModelWorkerAssignment(resp.GetItem())
}

// DeleteWorkerAssignment menghapus penugasan pekerja berdasarkan ID.
func (a *Adapter) DeleteWorkerAssignment(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_worker_assignments", "human_resource.worker_assignments.delete", "worker-assignment", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.personnel.DeleteWorkerAssignment(c, req)
	})
}
