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

func developmentAction(v *entity.DevelopmentAction) *pb.DevelopmentAction {
	if v == nil {
		return nil
	}
	return &pb.DevelopmentAction{Id: v.ID, DevelopmentPlanId: v.DevelopmentPlanID,
		ActionNumber: v.ActionNumber, CompetencyId: v.CompetencyID, CourseId: v.CourseID, ActionType: v.ActionType, Objective: v.Objective,
		TargetLevel: levelValue(v.TargetLevel), DueDate: timestamppb.New(v.DueDate), CompletedAt: timestamp(v.CompletedAt), EvidenceDocumentVersionReferenceId: v.EvidenceDocumentVersionReferenceID}
}
func actionResult(v *entity.DevelopmentAction, err error) (*pb.DevelopmentActionResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.DevelopmentActionResponse{Value: developmentAction(v)}, nil
}
func actions(p operation.Page[entity.DevelopmentAction], err error) (*pb.DevelopmentActionsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.DevelopmentAction, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, developmentAction(&p.Items[i]))
	}
	return &pb.DevelopmentActionsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateDevelopmentAction(ctx context.Context, r *pb.CreateDevelopmentActionRequest) (*pb.DevelopmentActionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	due, err := response.Timestamp(in.DueDate)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := response.OptionalTimestamp(in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	target, err := level(in.TargetLevel)
	if err != nil {
		return nil, response.Error(err)
	}
	return actionResult(s.service.CreateDevelopmentAction(ctx, service.CreateDevelopmentActionInput{
		DevelopmentPlanID: in.GetDevelopmentPlanId(), ActionNumber: in.GetActionNumber(), CompetencyID: in.CompetencyId, CourseID: in.CourseId,
		ActionType: in.GetActionType(), Objective: in.GetObjective(), TargetLevel: target, DueDate: due, CompletedAt: completed, EvidenceDocumentVersionReferenceID: in.EvidenceDocumentVersionReferenceId}))
}
func (s *server) UpdateDevelopmentAction(ctx context.Context, r *pb.UpdateDevelopmentActionRequest) (*pb.DevelopmentActionResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "development_plan_id", "action_number", "competency_id", "course_id", "action_type", "objective", "target_level", "due_date", "completed_at", "evidence_document_version_reference_id")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	due, err := timeField(f["due_date"], in.DueDate)
	if err != nil {
		return nil, response.Error(err)
	}
	completed, err := timeField(f["completed_at"], in.CompletedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	var target *int16
	if f["target_level"] {
		target, err = level(in.TargetLevel)
		if err != nil {
			return nil, response.Error(err)
		}
	}
	return actionResult(s.service.UpdateDevelopmentAction(ctx, r.Id, service.UpdateDevelopmentActionInput{
		DevelopmentPlanID: operation.Field[string]{Set: f["development_plan_id"], Value: in.DevelopmentPlanId},
		ActionNumber:      operation.Field[int32]{Set: f["action_number"], Value: in.ActionNumber}, CompetencyID: operation.Field[string]{Set: f["competency_id"], Value: in.CompetencyId},
		CourseID: operation.Field[string]{Set: f["course_id"], Value: in.CourseId}, ActionType: operation.Field[string]{Set: f["action_type"], Value: in.ActionType},
		Objective: operation.Field[string]{Set: f["objective"], Value: in.Objective}, TargetLevel: operation.Field[int16]{Set: f["target_level"], Value: target},
		DueDate: due, CompletedAt: completed, EvidenceDocumentVersionReferenceID: operation.Field[string]{Set: f["evidence_document_version_reference_id"], Value: in.EvidenceDocumentVersionReferenceId}}))
}
func (s *server) DeleteDevelopmentAction(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteDevelopmentAction(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetDevelopmentActionByID(ctx context.Context, r *commonv1.IDRequest) (*pb.DevelopmentActionResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return actionResult(s.service.GetDevelopmentActionByID(ctx, r.Id))
}
func (s *server) GetAllDevelopmentActions(ctx context.Context, r *commonv1.PageRequest) (*pb.DevelopmentActionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return actions(s.service.GetAllDevelopmentActions(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentActionByDevelopmentPlanID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentActionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return actions(s.service.FindDevelopmentActionByDevelopmentPlanID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentActionByActionNumber(ctx context.Context, r *commonv1.Int32FinderRequest) (*pb.DevelopmentActionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return actions(s.service.FindDevelopmentActionByActionNumber(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentActionByCompetencyID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentActionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return actions(s.service.FindDevelopmentActionByCompetencyID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentActionByCourseID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentActionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return actions(s.service.FindDevelopmentActionByCourseID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentActionByActionType(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentActionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return actions(s.service.FindDevelopmentActionByActionType(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindDevelopmentActionByEvidenceDocumentVersionReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.DevelopmentActionsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return actions(s.service.FindDevelopmentActionByEvidenceDocumentVersionReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
