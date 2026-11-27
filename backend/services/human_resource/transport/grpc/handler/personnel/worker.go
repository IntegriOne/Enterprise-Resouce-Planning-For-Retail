package personnel

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func worker(value *entity.Worker) *pb.Worker {
	return &pb.Worker{Id: value.ID, WorkerNumber: value.WorkerNumber, PersonPartyReferenceId: value.PersonPartyReferenceID,
		WorkerCategory: string(value.WorkerCategory), Status: string(value.Status), HireDate: dateString(value.HireDate),
		ServiceStartDate: dateString(value.ServiceStartDate), TerminationDate: dateString(value.TerminationDate), TerminationReasonCode: value.TerminationReasonCode,
		PrimaryLegalEntityReferenceId: value.PrimaryLegalEntityReferenceID, PrimaryAssignmentId: value.PrimaryAssignmentID, IamPrincipalReferenceId: value.IAMPrincipalReferenceID,
		CreatedAt: timestamppb.New(value.CreatedAt), UpdatedAt: timestamppb.New(value.UpdatedAt), CreatedByPrincipalId: value.CreatedByPrincipalID}
}
func workerResponse(value *entity.Worker, err error) (*pb.WorkerResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.WorkerResponse{Item: worker(value)}, nil
}
func workers(page operation.Page[entity.Worker], err error) (*pb.WorkersResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.Worker, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, worker(&page.Items[index]))
	}
	return &pb.WorkersResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateWorker(ctx context.Context, request *pb.CreateWorkerRequest) (*pb.WorkerResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	hire, err := optionalDate(in.HireDate)
	if err != nil {
		return nil, response.Error(err)
	}
	start, err := optionalDate(in.ServiceStartDate)
	if err != nil {
		return nil, response.Error(err)
	}
	termination, err := optionalDate(in.TerminationDate)
	if err != nil {
		return nil, response.Error(err)
	}
	return workerResponse(h.service.CreateWorker(ctx, service.CreateWorkerInput{WorkerNumber: in.GetWorkerNumber(), PersonPartyReferenceID: in.GetPersonPartyReferenceId(),
		WorkerCategory: entity.WorkerCategory(in.GetWorkerCategory()), Status: statusPointer(in.Status), HireDate: hire, ServiceStartDate: start, TerminationDate: termination,
		TerminationReasonCode: in.TerminationReasonCode, PrimaryLegalEntityReferenceID: in.GetPrimaryLegalEntityReferenceId(), PrimaryAssignmentID: in.PrimaryAssignmentId, IAMPrincipalReferenceID: in.IamPrincipalReferenceId}))
}

// ActivateWorker meneruskan aktivasi hubungan kerja ke command HRM.
func (h *Handler) ActivateWorker(ctx context.Context, request *common.IDRequest) (*pb.WorkerResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workerResponse(h.service.ActivateWorker(ctx, request.Id))
}

// TerminateWorker mencatat terminasi dari rencana offboarding yang disetujui.
func (h *Handler) TerminateWorker(ctx context.Context, request *pb.TerminateWorkerRequest) (*pb.WorkerResponse, error) {
	return workerResponse(h.service.TerminateWorker(ctx, request.GetWorkerId(), request.GetOffboardingPlanId()))
}

// ApplyPrimaryAssignment meneruskan penetapan assignment primer ke command HRM.
func (h *Handler) ApplyPrimaryAssignment(ctx context.Context, request *pb.ApplyPrimaryAssignmentRequest) (*pb.WorkerResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workerResponse(h.service.ApplyPrimaryAssignment(ctx, request.WorkerId, request.AssignmentId))
}
func (h *Handler) UpdateWorker(ctx context.Context, request *pb.UpdateWorkerRequest) (*pb.WorkerResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_number", "person_party_reference_id", "worker_category", "status", "hire_date", "service_start_date", "termination_date", "termination_reason_code", "primary_legal_entity_reference_id", "primary_assignment_id", "iam_principal_reference_id")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	hire, err := dateField(mask["hire_date"], in.HireDate)
	if err != nil {
		return nil, response.Error(err)
	}
	start, err := dateField(mask["service_start_date"], in.ServiceStartDate)
	if err != nil {
		return nil, response.Error(err)
	}
	termination, err := dateField(mask["termination_date"], in.TerminationDate)
	if err != nil {
		return nil, response.Error(err)
	}
	var category *entity.WorkerCategory
	if in.WorkerCategory != nil {
		value := entity.WorkerCategory(*in.WorkerCategory)
		category = &value
	}
	return workerResponse(h.service.UpdateWorker(ctx, request.Id, service.UpdateWorkerInput{
		WorkerNumber: field(mask["worker_number"], in.WorkerNumber), PersonPartyReferenceID: field(mask["person_party_reference_id"], in.PersonPartyReferenceId),
		WorkerCategory: field(mask["worker_category"], category), Status: field(mask["status"], statusPointer(in.Status)), HireDate: hire, ServiceStartDate: start, TerminationDate: termination,
		TerminationReasonCode: field(mask["termination_reason_code"], in.TerminationReasonCode), PrimaryLegalEntityReferenceID: field(mask["primary_legal_entity_reference_id"], in.PrimaryLegalEntityReferenceId),
		PrimaryAssignmentID: field(mask["primary_assignment_id"], in.PrimaryAssignmentId), IAMPrincipalReferenceID: field(mask["iam_principal_reference_id"], in.IamPrincipalReferenceId),
	}))
}
func (h *Handler) DeleteWorker(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteWorker(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetWorkerByID(ctx context.Context, request *common.IDRequest) (*pb.WorkerResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workerResponse(h.service.GetWorkerByID(ctx, request.Id))
}
func (h *Handler) GetAllWorkers(ctx context.Context, request *common.PageRequest) (*pb.WorkersResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workers(h.service.GetAllWorkers(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerByWorkerNumber(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkersResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workers(h.service.FindWorkerByWorkerNumber(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerByPersonPartyReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkersResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workers(h.service.FindWorkerByPersonPartyReferenceID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerByStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkersResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workers(h.service.FindWorkerByStatus(ctx, entity.WorkerStatus(*request.Value), int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerByTerminationReasonCode(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkersResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workers(h.service.FindWorkerByTerminationReasonCode(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerByPrimaryLegalEntityReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkersResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workers(h.service.FindWorkerByPrimaryLegalEntityReferenceID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerByPrimaryAssignmentID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkersResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workers(h.service.FindWorkerByPrimaryAssignmentID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerByIAMPrincipalReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkersResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workers(h.service.FindWorkerByIAMPrincipalReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerByCreatedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkersResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return workers(h.service.FindWorkerByCreatedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
