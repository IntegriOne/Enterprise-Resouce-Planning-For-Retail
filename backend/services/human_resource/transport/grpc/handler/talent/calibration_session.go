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

func calibrationSession(v *entity.TalentCalibrationSession) *pb.TalentCalibrationSession {
	if v == nil {
		return nil
	}
	return &pb.TalentCalibrationSession{Id: v.ID, TalentReviewCycleId: v.TalentReviewCycleID,
		SessionCode: v.SessionCode, ChairPrincipalId: v.ChairPrincipalID, ScheduledAt: timestamppb.New(v.ScheduledAt),
		Status: v.Status, ClosedAt: timestamp(v.ClosedAt), MinutesDocumentVersionReferenceId: v.MinutesDocumentVersionReferenceID}
}
func sessionResult(v *entity.TalentCalibrationSession, err error) (*pb.TalentCalibrationSessionResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.TalentCalibrationSessionResponse{Value: calibrationSession(v)}, nil
}
func sessions(p operation.Page[entity.TalentCalibrationSession], err error) (*pb.TalentCalibrationSessionsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.TalentCalibrationSession, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, calibrationSession(&p.Items[i]))
	}
	return &pb.TalentCalibrationSessionsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateTalentCalibrationSession(ctx context.Context, r *pb.CreateTalentCalibrationSessionRequest) (*pb.TalentCalibrationSessionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	scheduled, err := response.Timestamp(in.ScheduledAt)
	if err != nil {
		return nil, response.Error(err)
	}
	closed, err := response.OptionalTimestamp(in.ClosedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return sessionResult(s.service.CreateTalentCalibrationSession(ctx, service.CreateTalentCalibrationSessionInput{
		TalentReviewCycleID: in.GetTalentReviewCycleId(), SessionCode: in.GetSessionCode(), ChairPrincipalID: in.GetChairPrincipalId(),
		ScheduledAt: scheduled, Status: in.Status, ClosedAt: closed, MinutesDocumentVersionReferenceID: in.MinutesDocumentVersionReferenceId}))
}
func (s *server) UpdateTalentCalibrationSession(ctx context.Context, r *pb.UpdateTalentCalibrationSessionRequest) (*pb.TalentCalibrationSessionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "talent_review_cycle_id", "session_code", "chair_principal_id", "scheduled_at", "status", "closed_at", "minutes_document_version_reference_id")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	scheduled, err := timeField(f["scheduled_at"], in.ScheduledAt)
	if err != nil {
		return nil, response.Error(err)
	}
	closed, err := timeField(f["closed_at"], in.ClosedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return sessionResult(s.service.UpdateTalentCalibrationSession(ctx, r.Id, service.UpdateTalentCalibrationSessionInput{
		TalentReviewCycleID: operation.Field[string]{Set: f["talent_review_cycle_id"], Value: in.TalentReviewCycleId}, SessionCode: operation.Field[string]{Set: f["session_code"], Value: in.SessionCode},
		ChairPrincipalID: operation.Field[string]{Set: f["chair_principal_id"], Value: in.ChairPrincipalId}, ScheduledAt: scheduled,
		Status: operation.Field[string]{Set: f["status"], Value: in.Status}, ClosedAt: closed,
		MinutesDocumentVersionReferenceID: operation.Field[string]{Set: f["minutes_document_version_reference_id"], Value: in.MinutesDocumentVersionReferenceId}}))
}
func (s *server) DeleteTalentCalibrationSession(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteTalentCalibrationSession(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetTalentCalibrationSessionByID(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentCalibrationSessionResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return sessionResult(s.service.GetTalentCalibrationSessionByID(ctx, r.Id))
}
func (s *server) OpenTalentCalibrationSession(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentCalibrationSessionResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return sessionResult(s.service.OpenTalentCalibrationSession(ctx, r.Id))
}
func (s *server) GetAllTalentCalibrationSessions(ctx context.Context, r *commonv1.PageRequest) (*pb.TalentCalibrationSessionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return sessions(s.service.GetAllTalentCalibrationSessions(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationSessionByTalentReviewCycleID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationSessionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return sessions(s.service.FindTalentCalibrationSessionByTalentReviewCycleID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationSessionBySessionCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationSessionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return sessions(s.service.FindTalentCalibrationSessionBySessionCode(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationSessionByChairPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationSessionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return sessions(s.service.FindTalentCalibrationSessionByChairPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationSessionByStatus(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationSessionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return sessions(s.service.FindTalentCalibrationSessionByStatus(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentCalibrationSessionByMinutesDocumentVersionReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentCalibrationSessionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return sessions(s.service.FindTalentCalibrationSessionByMinutesDocumentVersionReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
