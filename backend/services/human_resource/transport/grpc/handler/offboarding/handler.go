// Package offboarding menerjemahkan RPC ke workflow keluar kerja tanpa memintas guard domain.
package offboarding

import (
	"context"
	"time"

	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/offboarding/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/offboarding"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedOffboardingServiceServer
	service *service.Service
}

// Register memasang service offboarding pada registrar privat.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterOffboardingServiceServer(registrar, &server{service: svc})
	return nil
}
func timestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(*value)
}
func timeField(selected bool, value *timestamppb.Timestamp) (operation.Field[time.Time], error) {
	if !selected {
		return operation.Field[time.Time]{}, nil
	}
	parsed, err := response.OptionalTimestamp(value)
	return operation.Field[time.Time]{Set: true, Value: parsed}, err
}
func plan(value *entity.OffboardingPlan) *pb.OffboardingPlan {
	if value == nil {
		return nil
	}
	return &pb.OffboardingPlan{
		Id: value.ID, WorkerId: value.WorkerID, PlanNumber: value.PlanNumber, ReasonCode: value.ReasonCode,
		NoticeDate: timestamppb.New(value.NoticeDate), LastWorkingDate: timestamppb.New(value.LastWorkingDate), TerminationEffectiveDate: timestamppb.New(value.TerminationEffectiveDate),
		Status: value.Status, ExitInterviewDocumentVersionReferenceId: value.ExitInterviewDocumentVersionReferenceID,
		FinalPayrollRunId: value.FinalPayrollRunID, FinanceClearanceReferenceId: value.FinanceClearanceReferenceID,
		ApprovedByPrincipalId: value.ApprovedByPrincipalID, ApprovedAt: timestamp(value.ApprovedAt), CompletedAt: timestamp(value.CompletedAt),
		CreatedByPrincipalId: value.CreatedByPrincipalID, CreatedAt: timestamppb.New(value.CreatedAt),
	}
}
func task(value *entity.OffboardingTask) *pb.OffboardingTask {
	if value == nil {
		return nil
	}
	return &pb.OffboardingTask{
		Id: value.ID, OffboardingPlanId: value.OffboardingPlanID, TaskCode: value.TaskCode, TaskName: value.TaskName, OwnerManagement: value.OwnerManagement,
		AssignedPrincipalId: value.AssignedPrincipalID, Status: value.Status, IsMandatory: value.IsMandatory, DueAt: timestamp(value.DueAt), CompletedAt: timestamp(value.CompletedAt),
		OwnerConfirmationReferenceId: value.OwnerConfirmationReferenceID, WaiverReason: value.WaiverReason, WaivedByPrincipalId: value.WaivedByPrincipalID,
	}
}
func planResponse(value *entity.OffboardingPlan, err error) (*pb.OffboardingPlanResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.OffboardingPlanResponse{Value: plan(value)}, nil
}
func taskResponse(value *entity.OffboardingTask, err error) (*pb.OffboardingTaskResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.OffboardingTaskResponse{Value: task(value)}, nil
}
func plans(page operation.Page[entity.OffboardingPlan], err error) (*pb.OffboardingPlansResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.OffboardingPlan, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, plan(&page.Items[index]))
	}
	return &pb.OffboardingPlansResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func tasks(page operation.Page[entity.OffboardingTask], err error) (*pb.OffboardingTasksResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.OffboardingTask, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, task(&page.Items[index]))
	}
	return &pb.OffboardingTasksResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}

func (s *server) CreateOffboardingPlan(ctx context.Context, request *pb.CreateOffboardingPlanRequest) (*pb.OffboardingPlanResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	notice, err := response.Timestamp(in.NoticeDate)
	if err != nil {
		return nil, response.Error(err)
	}
	last, err := response.Timestamp(in.LastWorkingDate)
	if err != nil {
		return nil, response.Error(err)
	}
	termination, err := response.Timestamp(in.TerminationEffectiveDate)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := response.OptionalTimestamp(in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return planResponse(s.service.CreateOffboardingPlan(ctx, service.CreateOffboardingPlanInput{
		WorkerID: in.GetWorkerId(), PlanNumber: in.GetPlanNumber(), ReasonCode: in.GetReasonCode(), NoticeDate: notice, LastWorkingDate: last, TerminationEffectiveDate: termination,
		Status: in.Status, ExitInterviewDocumentVersionReferenceID: in.ExitInterviewDocumentVersionReferenceId, FinalPayrollRunID: in.FinalPayrollRunId, FinanceClearanceReferenceID: in.FinanceClearanceReferenceId,
		ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approved, CompletedAt: completed,
	}))
}
func (s *server) UpdateOffboardingPlan(ctx context.Context, request *pb.UpdateOffboardingPlanRequest) (*pb.OffboardingPlanResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"worker_id", "plan_number", "reason_code", "notice_date", "last_working_date", "termination_effective_date", "status",
		"exit_interview_document_version_reference_id", "final_payroll_run_id", "finance_clearance_reference_id", "approved_by_principal_id", "approved_at", "completed_at",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	notice, err := timeField(selected["notice_date"], in.NoticeDate)
	if err != nil {
		return nil, response.Error(err)
	}
	last, err := timeField(selected["last_working_date"], in.LastWorkingDate)
	if err != nil {
		return nil, response.Error(err)
	}
	termination, err := timeField(selected["termination_effective_date"], in.TerminationEffectiveDate)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := timeField(selected["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := timeField(selected["completed_at"], in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return planResponse(s.service.UpdateOffboardingPlan(ctx, request.Id, service.UpdateOffboardingPlanInput{
		WorkerID: operation.Field[string]{Set: selected["worker_id"], Value: in.WorkerId}, PlanNumber: operation.Field[string]{Set: selected["plan_number"], Value: in.PlanNumber}, ReasonCode: operation.Field[string]{Set: selected["reason_code"], Value: in.ReasonCode},
		NoticeDate: notice, LastWorkingDate: last, TerminationEffectiveDate: termination, Status: operation.Field[string]{Set: selected["status"], Value: in.Status},
		ExitInterviewDocumentVersionReferenceID: operation.Field[string]{Set: selected["exit_interview_document_version_reference_id"], Value: in.ExitInterviewDocumentVersionReferenceId}, FinalPayrollRunID: operation.Field[string]{Set: selected["final_payroll_run_id"], Value: in.FinalPayrollRunId},
		FinanceClearanceReferenceID: operation.Field[string]{Set: selected["finance_clearance_reference_id"], Value: in.FinanceClearanceReferenceId}, ApprovedByPrincipalID: operation.Field[string]{Set: selected["approved_by_principal_id"], Value: in.ApprovedByPrincipalId}, ApprovedAt: approved, CompletedAt: completed,
	}))
}
func (s *server) DeleteOffboardingPlan(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteOffboardingPlan(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetOffboardingPlanByID(ctx context.Context, request *commonv1.IDRequest) (*pb.OffboardingPlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return planResponse(s.service.GetOffboardingPlanByID(ctx, request.Id))
}
func (s *server) GetAllOffboardingPlans(ctx context.Context, request *commonv1.PageRequest) (*pb.OffboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.GetAllOffboardingPlans(ctx, int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingPlanByWorkerID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindOffboardingPlanByWorkerID(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingPlanByPlanNumber(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindOffboardingPlanByPlanNumber(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingPlanByReasonCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindOffboardingPlanByReasonCode(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingPlanByStatus(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindOffboardingPlanByStatus(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingPlanByExitInterviewDocumentVersionReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindOffboardingPlanByExitInterviewDocumentVersionReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingPlanByFinalPayrollRunID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindOffboardingPlanByFinalPayrollRunID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingPlanByFinanceClearanceReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindOffboardingPlanByFinanceClearanceReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingPlanByApprovedByPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindOffboardingPlanByApprovedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingPlanByCreatedByPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindOffboardingPlanByCreatedByPrincipalID(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}

func (s *server) CreateOffboardingTask(ctx context.Context, request *pb.CreateOffboardingTaskRequest) (*pb.OffboardingTaskResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	due, err := response.OptionalTimestamp(in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := response.OptionalTimestamp(in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return taskResponse(s.service.CreateOffboardingTask(ctx, service.CreateOffboardingTaskInput{
		OffboardingPlanID: in.GetOffboardingPlanId(), TaskCode: in.GetTaskCode(), TaskName: in.GetTaskName(), OwnerManagement: in.GetOwnerManagement(),
		AssignedPrincipalID: in.AssignedPrincipalId, Status: in.Status, IsMandatory: in.IsMandatory, DueAt: due, CompletedAt: completed,
		OwnerConfirmationReferenceID: in.OwnerConfirmationReferenceId, WaiverReason: in.WaiverReason, WaivedByPrincipalID: in.WaivedByPrincipalId,
	}))
}
func (s *server) UpdateOffboardingTask(ctx context.Context, request *pb.UpdateOffboardingTaskRequest) (*pb.OffboardingTaskResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"offboarding_plan_id", "task_code", "task_name", "owner_management", "assigned_principal_id", "status", "is_mandatory", "due_at", "completed_at", "owner_confirmation_reference_id", "waiver_reason", "waived_by_principal_id",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	due, err := timeField(selected["due_at"], in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := timeField(selected["completed_at"], in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return taskResponse(s.service.UpdateOffboardingTask(ctx, request.Id, service.UpdateOffboardingTaskInput{
		OffboardingPlanID: operation.Field[string]{Set: selected["offboarding_plan_id"], Value: in.OffboardingPlanId}, TaskCode: operation.Field[string]{Set: selected["task_code"], Value: in.TaskCode}, TaskName: operation.Field[string]{Set: selected["task_name"], Value: in.TaskName},
		OwnerManagement: operation.Field[string]{Set: selected["owner_management"], Value: in.OwnerManagement}, AssignedPrincipalID: operation.Field[string]{Set: selected["assigned_principal_id"], Value: in.AssignedPrincipalId}, Status: operation.Field[string]{Set: selected["status"], Value: in.Status},
		IsMandatory: operation.Field[bool]{Set: selected["is_mandatory"], Value: in.IsMandatory}, DueAt: due, CompletedAt: completed,
		OwnerConfirmationReferenceID: operation.Field[string]{Set: selected["owner_confirmation_reference_id"], Value: in.OwnerConfirmationReferenceId}, WaiverReason: operation.Field[string]{Set: selected["waiver_reason"], Value: in.WaiverReason}, WaivedByPrincipalID: operation.Field[string]{Set: selected["waived_by_principal_id"], Value: in.WaivedByPrincipalId},
	}))
}
func (s *server) DeleteOffboardingTask(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteOffboardingTask(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetOffboardingTaskByID(ctx context.Context, request *commonv1.IDRequest) (*pb.OffboardingTaskResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return taskResponse(s.service.GetOffboardingTaskByID(ctx, request.Id))
}
func (s *server) GetAllOffboardingTasks(ctx context.Context, request *commonv1.PageRequest) (*pb.OffboardingTasksResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(s.service.GetAllOffboardingTasks(ctx, int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingTaskByOffboardingPlanID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingTasksResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(s.service.FindOffboardingTaskByOffboardingPlanID(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingTaskByTaskCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingTasksResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(s.service.FindOffboardingTaskByTaskCode(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingTaskByAssignedPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingTasksResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(s.service.FindOffboardingTaskByAssignedPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingTaskByStatus(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingTasksResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(s.service.FindOffboardingTaskByStatus(ctx, request.GetValue(), int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingTaskByOwnerConfirmationReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingTasksResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(s.service.FindOffboardingTaskByOwnerConfirmationReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (s *server) FindOffboardingTaskByWaivedByPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.OffboardingTasksResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return tasks(s.service.FindOffboardingTaskByWaivedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}

func (s *server) ApproveOffboardingPlan(ctx context.Context, request *commonv1.IDRequest) (*pb.OffboardingPlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return planResponse(s.service.ApproveOffboardingPlan(ctx, request.Id))
}

func (s *server) StartOffboardingPlan(ctx context.Context, request *commonv1.IDRequest) (*pb.OffboardingPlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return planResponse(s.service.StartOffboardingPlan(ctx, request.Id))
}
