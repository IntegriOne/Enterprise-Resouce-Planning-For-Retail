// Package succession menerjemahkan RPC bertipe ke application service suksesi HRM.
package succession

import (
	"context"
	"time"

	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/succession/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/succession"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	// Kontrak privat wajib memiliki implementasi eksplisit untuk setiap RPC.
	pb.UnsafeSuccessionServiceServer
	service *service.Service
}

// Register hanya mendaftarkan service suksesi; keamanan transport dipasang pada server pemilik.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterSuccessionServiceServer(registrar, &server{service: svc})
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

func plan(value *entity.SuccessionPlan) *pb.SuccessionPlan {
	if value == nil {
		return nil
	}
	return &pb.SuccessionPlan{
		Id: value.ID, PositionId: value.PositionID, PlanCode: value.PlanCode,
		PlanVersion: value.PlanVersion, ReviewDate: timestamppb.New(value.ReviewDate),
		ConfidentialityClass: value.ConfidentialityClass, ApprovedByPrincipalId: value.ApprovedByPrincipalID,
		ApprovedAt: timestamp(value.ApprovedAt), ApprovalMandateReferenceId: value.ApprovalMandateReferenceID, CreatedByPrincipalId: value.CreatedByPrincipalID, CreatedAt: timestamp(value.CreatedAt), SuccessionRisk: value.SuccessionRisk,
	}
}

func candidate(value *entity.SuccessionCandidate) *pb.SuccessionCandidate {
	if value == nil {
		return nil
	}
	return &pb.SuccessionCandidate{
		Id: value.ID, SuccessionPlanId: value.SuccessionPlanID, CandidateWorkerId: value.CandidateWorkerID,
		ReadinessCode: value.ReadinessCode, ReadinessAssessedAt: timestamppb.New(value.ReadinessAssessedAt),
		ReadinessRationale: value.ReadinessRationale, TalentAssessmentId: value.TalentAssessmentID,
		DevelopmentPlanId: value.DevelopmentPlanID, RiskOfLoss: value.RiskOfLoss,
	}
}

func plans(page operation.Page[entity.SuccessionPlan]) *pb.SuccessionPlansResponse {
	items := make([]*pb.SuccessionPlan, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, plan(&page.Items[index]))
	}
	return &pb.SuccessionPlansResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}
}

func candidates(page operation.Page[entity.SuccessionCandidate]) *pb.SuccessionCandidatesResponse {
	items := make([]*pb.SuccessionCandidate, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, candidate(&page.Items[index]))
	}
	return &pb.SuccessionCandidatesResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}
}

// ApproveSuccessionPlan meneruskan keputusan panel ke service suksesi.
func (s *server) ApproveSuccessionPlan(ctx context.Context, request *commonv1.IDRequest) (*pb.SuccessionPlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	result, err := s.service.ApproveSuccessionPlan(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.SuccessionPlanResponse{Value: plan(result)}, nil
}

func (s *server) CreateSuccessionPlan(ctx context.Context, request *pb.CreateSuccessionPlanRequest) (*pb.SuccessionPlanResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	reviewDate, err := response.Timestamp(in.ReviewDate)
	if err != nil {
		return nil, response.Error(err)
	}
	approvedAt, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.CreateSuccessionPlan(ctx, service.CreateSuccessionPlanInput{
		PositionID: in.GetPositionId(), PlanCode: in.GetPlanCode(), PlanVersion: in.GetPlanVersion(),
		ReviewDate: reviewDate, ConfidentialityClass: in.ConfidentialityClass,
		ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approvedAt, SuccessionRisk: in.SuccessionRisk,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.SuccessionPlanResponse{Value: plan(value)}, nil
}

func (s *server) UpdateSuccessionPlan(ctx context.Context, request *pb.UpdateSuccessionPlanRequest) (*pb.SuccessionPlanResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"position_id", "plan_code", "plan_version", "review_date", "confidentiality_class",
		"approved_by_principal_id", "approved_at", "succession_risk",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	reviewDate, err := timeField(selected["review_date"], in.ReviewDate)
	if err != nil {
		return nil, response.Error(err)
	}
	approvedAt, err := timeField(selected["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.UpdateSuccessionPlan(ctx, request.Id, service.UpdateSuccessionPlanInput{
		PositionID:  operation.Field[string]{Set: selected["position_id"], Value: in.PositionId},
		PlanCode:    operation.Field[string]{Set: selected["plan_code"], Value: in.PlanCode},
		PlanVersion: operation.Field[int32]{Set: selected["plan_version"], Value: in.PlanVersion},
		ReviewDate:  reviewDate, ApprovedAt: approvedAt,
		ConfidentialityClass:  operation.Field[string]{Set: selected["confidentiality_class"], Value: in.ConfidentialityClass},
		ApprovedByPrincipalID: operation.Field[string]{Set: selected["approved_by_principal_id"], Value: in.ApprovedByPrincipalId},
		SuccessionRisk:        operation.Field[string]{Set: selected["succession_risk"], Value: in.SuccessionRisk},
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.SuccessionPlanResponse{Value: plan(value)}, nil
}

func (s *server) DeleteSuccessionPlan(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteSuccessionPlan(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetSuccessionPlanByID(ctx context.Context, request *commonv1.IDRequest) (*pb.SuccessionPlanResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	value, err := s.service.GetSuccessionPlanByID(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.SuccessionPlanResponse{Value: plan(value)}, nil
}
func (s *server) GetAllSuccessionPlans(ctx context.Context, request *commonv1.PageRequest) (*pb.SuccessionPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.GetAllSuccessionPlans(ctx, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return plans(page), nil
}
func (s *server) FindSuccessionPlanByPositionID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.SuccessionPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindSuccessionPlanByPositionID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return plans(page), nil
}
func (s *server) FindSuccessionPlanByPlanCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.SuccessionPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindSuccessionPlanByPlanCode(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return plans(page), nil
}
func (s *server) FindSuccessionPlanByApprovedByPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.SuccessionPlansResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindSuccessionPlanByApprovedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return plans(page), nil
}

func (s *server) CreateSuccessionCandidate(ctx context.Context, request *pb.CreateSuccessionCandidateRequest) (*pb.SuccessionCandidateResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	assessedAt, err := response.Timestamp(in.ReadinessAssessedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.CreateSuccessionCandidate(ctx, service.CreateSuccessionCandidateInput{
		SuccessionPlanID: in.GetSuccessionPlanId(), CandidateWorkerID: in.GetCandidateWorkerId(),
		ReadinessCode: in.GetReadinessCode(), ReadinessAssessedAt: assessedAt,
		ReadinessRationale: in.ReadinessRationale, TalentAssessmentID: in.TalentAssessmentId,
		DevelopmentPlanID: in.DevelopmentPlanId, RiskOfLoss: in.RiskOfLoss,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.SuccessionCandidateResponse{Value: candidate(value)}, nil
}
func (s *server) UpdateSuccessionCandidate(ctx context.Context, request *pb.UpdateSuccessionCandidateRequest) (*pb.SuccessionCandidateResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"succession_plan_id", "candidate_worker_id", "readiness_code", "readiness_assessed_at",
		"readiness_rationale", "talent_assessment_id", "development_plan_id", "risk_of_loss",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	assessedAt, err := timeField(selected["readiness_assessed_at"], in.ReadinessAssessedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.UpdateSuccessionCandidate(ctx, request.Id, service.UpdateSuccessionCandidateInput{
		SuccessionPlanID:    operation.Field[string]{Set: selected["succession_plan_id"], Value: in.SuccessionPlanId},
		CandidateWorkerID:   operation.Field[string]{Set: selected["candidate_worker_id"], Value: in.CandidateWorkerId},
		ReadinessCode:       operation.Field[string]{Set: selected["readiness_code"], Value: in.ReadinessCode},
		ReadinessAssessedAt: assessedAt,
		ReadinessRationale:  operation.Field[string]{Set: selected["readiness_rationale"], Value: in.ReadinessRationale},
		TalentAssessmentID:  operation.Field[string]{Set: selected["talent_assessment_id"], Value: in.TalentAssessmentId},
		DevelopmentPlanID:   operation.Field[string]{Set: selected["development_plan_id"], Value: in.DevelopmentPlanId},
		RiskOfLoss:          operation.Field[string]{Set: selected["risk_of_loss"], Value: in.RiskOfLoss},
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.SuccessionCandidateResponse{Value: candidate(value)}, nil
}
func (s *server) DeleteSuccessionCandidate(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteSuccessionCandidate(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetSuccessionCandidateByID(ctx context.Context, request *commonv1.IDRequest) (*pb.SuccessionCandidateResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	value, err := s.service.GetSuccessionCandidateByID(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.SuccessionCandidateResponse{Value: candidate(value)}, nil
}
func (s *server) GetAllSuccessionCandidates(ctx context.Context, request *commonv1.PageRequest) (*pb.SuccessionCandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.GetAllSuccessionCandidates(ctx, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return candidates(page), nil
}
func (s *server) FindSuccessionCandidateBySuccessionPlanID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.SuccessionCandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindSuccessionCandidateBySuccessionPlanID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return candidates(page), nil
}
func (s *server) FindSuccessionCandidateByCandidateWorkerID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.SuccessionCandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindSuccessionCandidateByCandidateWorkerID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return candidates(page), nil
}
func (s *server) FindSuccessionCandidateByReadinessCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.SuccessionCandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindSuccessionCandidateByReadinessCode(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return candidates(page), nil
}
func (s *server) FindSuccessionCandidateByTalentAssessmentID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.SuccessionCandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindSuccessionCandidateByTalentAssessmentID(ctx, request.Value, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return candidates(page), nil
}
func (s *server) FindSuccessionCandidateByDevelopmentPlanID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.SuccessionCandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindSuccessionCandidateByDevelopmentPlanID(ctx, request.Value, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return candidates(page), nil
}
