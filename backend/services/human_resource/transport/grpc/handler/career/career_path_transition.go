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
)

func transition(v *entity.CareerPathTransition) *pb.CareerPathTransition {
	if v == nil {
		return nil
	}
	return &pb.CareerPathTransition{Id: v.ID, CareerPathId: v.CareerPathID, FromJobId: v.FromJobID,
		ToJobId: v.ToJobID, MovementType: v.MovementType, MinimumTenureMonths: v.MinimumTenureMonths,
		RequirementsDocumentVersionReferenceId: v.RequirementsDocumentVersionReferenceID, RequirementsDigest: v.RequirementsDigest}
}
func transitionResult(v *entity.CareerPathTransition, err error) (*pb.CareerPathTransitionResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.CareerPathTransitionResponse{Value: transition(v)}, nil
}
func transitions(p operation.Page[entity.CareerPathTransition], err error) (*pb.CareerPathTransitionsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.CareerPathTransition, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, transition(&p.Items[i]))
	}
	return &pb.CareerPathTransitionsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateCareerPathTransition(ctx context.Context, r *pb.CreateCareerPathTransitionRequest) (*pb.CareerPathTransitionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	return transitionResult(s.service.CreateCareerPathTransition(ctx, service.CreateCareerPathTransitionInput{
		CareerPathID: in.GetCareerPathId(), FromJobID: in.GetFromJobId(), ToJobID: in.GetToJobId(), MovementType: in.GetMovementType(),
		MinimumTenureMonths: in.MinimumTenureMonths, RequirementsDocumentVersionReferenceID: in.GetRequirementsDocumentVersionReferenceId(), RequirementsDigest: in.RequirementsDigest}))
}
func (s *server) UpdateCareerPathTransition(ctx context.Context, r *pb.UpdateCareerPathTransitionRequest) (*pb.CareerPathTransitionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "career_path_id", "from_job_id", "to_job_id", "movement_type", "minimum_tenure_months", "requirements_document_version_reference_id", "requirements_digest")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	return transitionResult(s.service.UpdateCareerPathTransition(ctx, r.Id, service.UpdateCareerPathTransitionInput{
		CareerPathID: operation.Field[string]{Set: f["career_path_id"], Value: in.CareerPathId}, FromJobID: operation.Field[string]{Set: f["from_job_id"], Value: in.FromJobId},
		ToJobID: operation.Field[string]{Set: f["to_job_id"], Value: in.ToJobId}, MovementType: operation.Field[string]{Set: f["movement_type"], Value: in.MovementType},
		MinimumTenureMonths:                    operation.Field[int32]{Set: f["minimum_tenure_months"], Value: in.MinimumTenureMonths},
		RequirementsDocumentVersionReferenceID: operation.Field[string]{Set: f["requirements_document_version_reference_id"], Value: in.RequirementsDocumentVersionReferenceId},
		RequirementsDigest:                     bytesField(f["requirements_digest"], in.RequirementsDigest)}))
}
func (s *server) DeleteCareerPathTransition(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteCareerPathTransition(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetCareerPathTransitionByID(ctx context.Context, r *commonv1.IDRequest) (*pb.CareerPathTransitionResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return transitionResult(s.service.GetCareerPathTransitionByID(ctx, r.Id))
}
func (s *server) GetAllCareerPathTransitions(ctx context.Context, r *commonv1.PageRequest) (*pb.CareerPathTransitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return transitions(s.service.GetAllCareerPathTransitions(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindCareerPathTransitionByCareerPathID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CareerPathTransitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return transitions(s.service.FindCareerPathTransitionByCareerPathID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCareerPathTransitionByFromJobID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CareerPathTransitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return transitions(s.service.FindCareerPathTransitionByFromJobID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCareerPathTransitionByToJobID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CareerPathTransitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return transitions(s.service.FindCareerPathTransitionByToJobID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCareerPathTransitionByMovementType(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CareerPathTransitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return transitions(s.service.FindCareerPathTransitionByMovementType(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCareerPathTransitionByRequirementsDocumentVersionReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CareerPathTransitionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return transitions(s.service.FindCareerPathTransitionByRequirementsDocumentVersionReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
