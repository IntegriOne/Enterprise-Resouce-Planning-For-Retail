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

func aspiration(v *entity.WorkerCareerAspiration) *pb.WorkerCareerAspiration {
	if v == nil {
		return nil
	}
	return &pb.WorkerCareerAspiration{Id: v.ID, WorkerId: v.WorkerID, TargetJobId: v.TargetJobID,
		CareerPathTransitionId: v.CareerPathTransitionID, PreferenceRank: v.PreferenceRank, RecordedAt: timestamppb.New(v.RecordedAt),
		RecordedByPrincipalId: v.RecordedByPrincipalID, TargetDate: timestamp(v.TargetDate), MobilityPreference: v.MobilityPreference,
		PreferredLocationReferenceId: v.PreferredLocationReferenceID, WithdrawnAt: timestamp(v.WithdrawnAt)}
}
func aspirationResult(v *entity.WorkerCareerAspiration, err error) (*pb.WorkerCareerAspirationResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.WorkerCareerAspirationResponse{Value: aspiration(v)}, nil
}
func aspirations(p operation.Page[entity.WorkerCareerAspiration], err error) (*pb.WorkerCareerAspirationsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.WorkerCareerAspiration, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, aspiration(&p.Items[i]))
	}
	return &pb.WorkerCareerAspirationsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateWorkerCareerAspiration(ctx context.Context, r *pb.CreateWorkerCareerAspirationRequest) (*pb.WorkerCareerAspirationResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	target, err := response.OptionalTimestamp(in.TargetDate)
	if err != nil {
		return nil, response.Error(err)
	}
	withdrawn, err := response.OptionalTimestamp(in.WithdrawnAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return aspirationResult(s.service.CreateWorkerCareerAspiration(ctx, service.CreateWorkerCareerAspirationInput{
		WorkerID: in.GetWorkerId(), TargetJobID: in.GetTargetJobId(), CareerPathTransitionID: in.CareerPathTransitionId,
		PreferenceRank: in.GetPreferenceRank(), TargetDate: target, MobilityPreference: in.MobilityPreference,
		PreferredLocationReferenceID: in.PreferredLocationReferenceId, WithdrawnAt: withdrawn}))
}
func (s *server) UpdateWorkerCareerAspiration(ctx context.Context, r *pb.UpdateWorkerCareerAspirationRequest) (*pb.WorkerCareerAspirationResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "worker_id", "target_job_id", "career_path_transition_id", "preference_rank", "target_date", "mobility_preference", "preferred_location_reference_id", "withdrawn_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	target, err := timeField(f["target_date"], in.TargetDate)
	if err != nil {
		return nil, response.Error(err)
	}
	withdrawn, err := timeField(f["withdrawn_at"], in.WithdrawnAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return aspirationResult(s.service.UpdateWorkerCareerAspiration(ctx, r.Id, service.UpdateWorkerCareerAspirationInput{
		WorkerID: operation.Field[string]{Set: f["worker_id"], Value: in.WorkerId}, TargetJobID: operation.Field[string]{Set: f["target_job_id"], Value: in.TargetJobId},
		CareerPathTransitionID: operation.Field[string]{Set: f["career_path_transition_id"], Value: in.CareerPathTransitionId},
		PreferenceRank:         operation.Field[int32]{Set: f["preference_rank"], Value: in.PreferenceRank}, TargetDate: target,
		MobilityPreference:           operation.Field[string]{Set: f["mobility_preference"], Value: in.MobilityPreference},
		PreferredLocationReferenceID: operation.Field[string]{Set: f["preferred_location_reference_id"], Value: in.PreferredLocationReferenceId}, WithdrawnAt: withdrawn}))
}
func (s *server) DeleteWorkerCareerAspiration(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteWorkerCareerAspiration(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetWorkerCareerAspirationByID(ctx context.Context, r *commonv1.IDRequest) (*pb.WorkerCareerAspirationResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return aspirationResult(s.service.GetWorkerCareerAspirationByID(ctx, r.Id))
}
func (s *server) GetAllWorkerCareerAspirations(ctx context.Context, r *commonv1.PageRequest) (*pb.WorkerCareerAspirationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return aspirations(s.service.GetAllWorkerCareerAspirations(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindWorkerCareerAspirationByWorkerID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.WorkerCareerAspirationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return aspirations(s.service.FindWorkerCareerAspirationByWorkerID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindWorkerCareerAspirationByTargetJobID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.WorkerCareerAspirationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return aspirations(s.service.FindWorkerCareerAspirationByTargetJobID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindWorkerCareerAspirationByCareerPathTransitionID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.WorkerCareerAspirationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return aspirations(s.service.FindWorkerCareerAspirationByCareerPathTransitionID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindWorkerCareerAspirationByRecordedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.WorkerCareerAspirationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return aspirations(s.service.FindWorkerCareerAspirationByRecordedByPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindWorkerCareerAspirationByPreferredLocationReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.WorkerCareerAspirationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return aspirations(s.service.FindWorkerCareerAspirationByPreferredLocationReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
