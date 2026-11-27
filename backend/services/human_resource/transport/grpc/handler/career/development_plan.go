package career

import (
	"context"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/career/v1"
	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/career"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func developmentPlan(v *entity.DevelopmentPlan) *pb.DevelopmentPlan {
	if v == nil {
		return nil
	}
	return &pb.DevelopmentPlan{Id: v.ID, WorkerId: v.WorkerID, PlanCode: v.PlanCode,
		PlanVersion: v.PlanVersion, TargetJobId: v.TargetJobID, StartDate: timestamppb.New(v.StartDate), TargetDate: timestamppb.New(v.TargetDate),
		Status: v.Status, MentorWorkerId: v.MentorWorkerID, ApprovedByPrincipalId: v.ApprovedByPrincipalID, ApprovedAt: timestamp(v.ApprovedAt)}
}
func planResult(v *entity.DevelopmentPlan, err error) (*pb.DevelopmentPlanResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.DevelopmentPlanResponse{Value: developmentPlan(v)}, nil
}
func plans(p operation.Page[entity.DevelopmentPlan], err error) (*pb.DevelopmentPlansResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.DevelopmentPlan, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, developmentPlan(&p.Items[i]))
	}
	return &pb.DevelopmentPlansResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateDevelopmentPlan(ctx context.Context, r *pb.CreateDevelopmentPlanRequest) (*pb.DevelopmentPlanResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	start, err := response.Timestamp(in.StartDate)
	if err != nil {
		return nil, response.Error(err)
	}
	target, err := response.Timestamp(in.TargetDate)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return planResult(s.service.CreateDevelopmentPlan(ctx, service.CreateDevelopmentPlanInput{
		WorkerID: in.GetWorkerId(), PlanCode: in.GetPlanCode(), PlanVersion: in.GetPlanVersion(), TargetJobID: in.TargetJobId,
		StartDate: start, TargetDate: target, Status: in.Status, MentorWorkerID: in.MentorWorkerId,
		ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approved}))
}
func (s *server) UpdateDevelopmentPlan(ctx context.Context, r *pb.UpdateDevelopmentPlanRequest) (*pb.DevelopmentPlanResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "worker_id", "plan_code", "plan_version", "target_job_id", "start_date", "target_date", "status", "mentor_worker_id", "approved_by_principal_id", "approved_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	start, err := timeField(f["start_date"], in.StartDate)
	if err != nil {
		return nil, response.Error(err)
	}
	target, err := timeField(f["target_date"], in.TargetDate)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := timeField(f["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return planResult(s.service.UpdateDevelopmentPlan(ctx, r.Id, service.UpdateDevelopmentPlanInput{
		WorkerID: operation.Field[string]{Set: f["worker_id"], Value: in.WorkerId}, PlanCode: operation.Field[string]{Set: f["plan_code"], Value: in.PlanCode},
		PlanVersion: operation.Field[int32]{Set: f["plan_version"], Value: in.PlanVersion}, TargetJobID: operation.Field[string]{Set: f["target_job_id"], Value: in.TargetJobId},
		StartDate: start, TargetDate: target, Status: operation.Field[string]{Set: f["status"], Value: in.Status},
		MentorWorkerID:        operation.Field[string]{Set: f["mentor_worker_id"], Value: in.MentorWorkerId},
		ApprovedByPrincipalID: operation.Field[string]{Set: f["approved_by_principal_id"], Value: in.ApprovedByPrincipalId}, ApprovedAt: approved}))
}
func (s *server) DeleteDevelopmentPlan(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteDevelopmentPlan(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetDevelopmentPlanByID(ctx context.Context, r *commonv1.IDRequest) (*pb.DevelopmentPlanResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return planResult(s.service.GetDevelopmentPlanByID(ctx, r.Id))
}
func (s *server) GetAllDevelopmentPlans(ctx context.Context, r *commonv1.PageRequest) (*pb.DevelopmentPlansResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.GetAllDevelopmentPlans(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentPlanByWorkerID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentPlansResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindDevelopmentPlanByWorkerID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentPlanByPlanCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentPlansResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindDevelopmentPlanByPlanCode(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentPlanByTargetJobID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentPlansResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindDevelopmentPlanByTargetJobID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentPlanByStatus(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentPlansResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindDevelopmentPlanByStatus(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentPlanByMentorWorkerID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentPlansResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindDevelopmentPlanByMentorWorkerID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentPlanByApprovedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentPlansResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return plans(s.service.FindDevelopmentPlanByApprovedByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}

func (s *server) ApproveDevelopmentPlan(ctx context.Context, request *commonv1.IDRequest) (*pb.DevelopmentPlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return planResult(s.service.ApproveDevelopmentPlan(ctx, request.Id))
}

func (s *server) CompleteDevelopmentPlan(ctx context.Context, request *commonv1.IDRequest) (*pb.DevelopmentPlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return planResult(s.service.CompleteDevelopmentPlan(ctx, request.Id))
}
