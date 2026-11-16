package talent

import (
	"context"
	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/talent/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/talent"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func membership(v *entity.TalentPoolMembership) *pb.TalentPoolMembership {
	if v == nil {
		return nil
	}
	return &pb.TalentPoolMembership{Id: v.ID, TalentPoolId: v.TalentPoolID, WorkerId: v.WorkerID,
		AssessmentId: v.AssessmentID, NominatedByPrincipalId: v.NominatedByPrincipalID, Status: v.Status,
		ValidFrom: timestamppb.New(v.ValidFrom), ValidTo: timestamp(v.ValidTo), ApprovedByPrincipalId: v.ApprovedByPrincipalID,
		ApprovedAt: timestamp(v.ApprovedAt), ExitReasonCode: v.ExitReasonCode}
}
func membershipResult(v *entity.TalentPoolMembership, err error) (*pb.TalentPoolMembershipResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.TalentPoolMembershipResponse{Value: membership(v)}, nil
}
func memberships(p operation.Page[entity.TalentPoolMembership], err error) (*pb.TalentPoolMembershipsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.TalentPoolMembership, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, membership(&p.Items[i]))
	}
	return &pb.TalentPoolMembershipsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateTalentPoolMembership(ctx context.Context, r *pb.CreateTalentPoolMembershipRequest) (*pb.TalentPoolMembershipResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	from, err := response.Timestamp(in.ValidFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := response.OptionalTimestamp(in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return membershipResult(s.service.CreateTalentPoolMembership(ctx, service.CreateTalentPoolMembershipInput{
		TalentPoolID: in.GetTalentPoolId(), WorkerID: in.GetWorkerId(), AssessmentID: in.GetAssessmentId(), NominatedByPrincipalID: in.GetNominatedByPrincipalId(),
		Status: in.Status, ValidFrom: from, ValidTo: to, ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approved, ExitReasonCode: in.ExitReasonCode}))
}
func (s *server) UpdateTalentPoolMembership(ctx context.Context, r *pb.UpdateTalentPoolMembershipRequest) (*pb.TalentPoolMembershipResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "talent_pool_id", "worker_id", "assessment_id", "nominated_by_principal_id", "status", "valid_from", "valid_to", "approved_by_principal_id", "approved_at", "exit_reason_code")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	from, err := timeField(f["valid_from"], in.ValidFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := timeField(f["valid_to"], in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := timeField(f["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return membershipResult(s.service.UpdateTalentPoolMembership(ctx, r.Id, service.UpdateTalentPoolMembershipInput{
		TalentPoolID: operation.Field[string]{Set: f["talent_pool_id"], Value: in.TalentPoolId}, WorkerID: operation.Field[string]{Set: f["worker_id"], Value: in.WorkerId},
		AssessmentID:           operation.Field[string]{Set: f["assessment_id"], Value: in.AssessmentId},
		NominatedByPrincipalID: operation.Field[string]{Set: f["nominated_by_principal_id"], Value: in.NominatedByPrincipalId},
		Status:                 operation.Field[string]{Set: f["status"], Value: in.Status}, ValidFrom: from, ValidTo: to,
		ApprovedByPrincipalID: operation.Field[string]{Set: f["approved_by_principal_id"], Value: in.ApprovedByPrincipalId}, ApprovedAt: approved,
		ExitReasonCode: operation.Field[string]{Set: f["exit_reason_code"], Value: in.ExitReasonCode}}))
}
func (s *server) DeleteTalentPoolMembership(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteTalentPoolMembership(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetTalentPoolMembershipByID(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentPoolMembershipResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return membershipResult(s.service.GetTalentPoolMembershipByID(ctx, r.Id))
}
func (s *server) GetAllTalentPoolMemberships(ctx context.Context, r *commonv1.PageRequest) (*pb.TalentPoolMembershipsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return memberships(s.service.GetAllTalentPoolMemberships(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolMembershipByTalentPoolID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolMembershipsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return memberships(s.service.FindTalentPoolMembershipByTalentPoolID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolMembershipByWorkerID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolMembershipsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return memberships(s.service.FindTalentPoolMembershipByWorkerID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolMembershipByAssessmentID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolMembershipsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return memberships(s.service.FindTalentPoolMembershipByAssessmentID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolMembershipByNominatedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolMembershipsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return memberships(s.service.FindTalentPoolMembershipByNominatedByPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolMembershipByStatus(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolMembershipsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return memberships(s.service.FindTalentPoolMembershipByStatus(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolMembershipByApprovedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolMembershipsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return memberships(s.service.FindTalentPoolMembershipByApprovedByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolMembershipByExitReasonCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolMembershipsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return memberships(s.service.FindTalentPoolMembershipByExitReasonCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
