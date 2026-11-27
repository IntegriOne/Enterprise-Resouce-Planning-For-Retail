package human_resource

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// CreateWorker mendaftarkan pekerja pada tahap preboarding.
func (a *Adapter) CreateWorker(ctx context.Context, input model.CreateWorkerInput) (*model.Worker, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.workers.create")
	if err != nil {
		return nil, err
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.CreateWorker(rpcCtx, &personnel.CreateWorkerRequest{Input: &personnel.WorkerInput{
		WorkerNumber: &input.WorkerNumber, WorkerCategory: &input.WorkerCategory,
		PersonPartyReferenceId: optionalRawNodeID(&input.PersonPartyReferenceID, "party"), PrimaryLegalEntityReferenceId: optionalRawNodeID(&input.PrimaryLegalEntityReferenceID, "legal-entity"),
		HireDate: input.HireDate, ServiceStartDate: input.ServiceStartDate,
	}})
	return workerMutationResponse(response, err, "", "preboarding")
}

// UpdateWorker mengoreksi tanggal preboarding sesuai aturan referensi HRM.
func (a *Adapter) UpdateWorker(ctx context.Context, rawID string, input model.UpdateWorkerInput) (*model.Worker, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.workers.update")
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(strings.TrimPrefix(rawID, workerGIDPrefix))
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker ID", nil)
	}
	patch := &personnel.WorkerInput{}
	mask := &fieldmaskpb.FieldMask{}
	if value, set := input.HireDate.ValueOK(); set {
		patch.HireDate = value
		mask.Paths = append(mask.Paths, "hire_date")
	}
	if value, set := input.ServiceStartDate.ValueOK(); set {
		patch.ServiceStartDate = value
		mask.Paths = append(mask.Paths, "service_start_date")
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("worker update is empty", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.UpdateWorker(rpcCtx, &personnel.UpdateWorkerRequest{
		Id: id.String(), Input: patch, UpdateMask: mask,
	})
	return workerMutationResponse(response, err, id.String(), "preboarding")
}

func workerMutationResponse(response *personnel.WorkerResponse, err error, expectedID, expectedStatus string) (*model.Worker, error) {
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return nil, middleware.ErrConflict("worker conflicts with existing state", err)
	case codes.OK:
	default:
		return nil, mapReadError(err, "worker")
	}
	worker := response.GetItem()
	id, err := uuid.Parse(worker.GetId())
	if err != nil || id == uuid.Nil || worker.GetStatus() != expectedStatus || (expectedID != "" && worker.GetId() != expectedID) {
		return nil, middleware.ErrUnavailable("invalid human resource worker response", nil)
	}
	return toModelWorker(worker), nil
}

// ActivateWorker menjalankan transisi preboarding ke aktif melalui HRM.
func (a *Adapter) ActivateWorker(ctx context.Context, rawID string) (*model.Worker, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.workers.activate")
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(strings.TrimPrefix(rawID, workerGIDPrefix))
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker ID", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.ActivateWorker(rpcCtx, &common.IDRequest{Id: id.String()})
	return workerMutationResponse(response, err, id.String(), "active")
}

// TerminateWorker menerapkan keputusan offboarding melalui HRM.
func (a *Adapter) TerminateWorker(ctx context.Context, rawID, rawPlanID string) (*model.Worker, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.workers.terminate")
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(strings.TrimPrefix(rawID, workerGIDPrefix))
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker ID", nil)
	}
	planID, err := parseNodeID(rawPlanID, "offboarding-plan")
	if err != nil || planID == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid offboarding plan ID", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.TerminateWorker(rpcCtx, &personnel.TerminateWorkerRequest{
		WorkerId: id.String(), OffboardingPlanId: planID.String(),
	})
	return workerMutationResponse(response, err, id.String(), "terminated")
}

// ApplyPrimaryAssignment menetapkan assignment primer yang disetujui HRM.
func (a *Adapter) ApplyPrimaryAssignment(ctx context.Context, rawID, rawAssignmentID string) (*model.Worker, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.workers.apply_primary_assignment")
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(strings.TrimPrefix(rawID, workerGIDPrefix))
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid worker ID", nil)
	}
	assignmentID, err := parseNodeID(rawAssignmentID, "worker-assignment")
	if err != nil || assignmentID == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid assignment ID", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.personnel.ApplyPrimaryAssignment(rpcCtx, &personnel.ApplyPrimaryAssignmentRequest{
		WorkerId: id.String(), AssignmentId: assignmentID.String(),
	})
	if err == nil && response.GetItem().GetPrimaryAssignmentId() != assignmentID.String() {
		return nil, middleware.ErrUnavailable("invalid human resource assignment response", nil)
	}
	return workerMutationResponse(response, err, id.String(), "active")
}
