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

func reviewCycle(v *entity.TalentReviewCycle) *pb.TalentReviewCycle {
	if v == nil {
		return nil
	}
	return &pb.TalentReviewCycle{Id: v.ID, CycleCode: v.CycleCode, FrameworkId: v.FrameworkID,
		LegalEntityReferenceId: v.LegalEntityReferenceID, OrganizationUnitReferenceId: v.OrganizationUnitReferenceID,
		PeriodStart: timestamppb.New(v.PeriodStart), PeriodEnd: timestamppb.New(v.PeriodEnd), Status: v.Status, OwnerPrincipalId: v.OwnerPrincipalID, ClosedAt: timestamp(v.ClosedAt)}
}
func cycleResult(v *entity.TalentReviewCycle, err error) (*pb.TalentReviewCycleResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.TalentReviewCycleResponse{Value: reviewCycle(v)}, nil
}
func cycles(p operation.Page[entity.TalentReviewCycle], err error) (*pb.TalentReviewCyclesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.TalentReviewCycle, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, reviewCycle(&p.Items[i]))
	}
	return &pb.TalentReviewCyclesResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateTalentReviewCycle(ctx context.Context, r *pb.CreateTalentReviewCycleRequest) (*pb.TalentReviewCycleResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	start, err := response.Timestamp(in.PeriodStart)
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := response.Timestamp(in.PeriodEnd)
	if err != nil {
		return nil, response.Error(err)
	}
	closed, err := response.OptionalTimestamp(in.ClosedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return cycleResult(s.service.CreateTalentReviewCycle(ctx, service.CreateTalentReviewCycleInput{
		CycleCode: in.GetCycleCode(), FrameworkID: in.GetFrameworkId(), LegalEntityReferenceID: in.GetLegalEntityReferenceId(),
		OrganizationUnitReferenceID: in.OrganizationUnitReferenceId, PeriodStart: start, PeriodEnd: end, Status: in.Status, OwnerPrincipalID: in.GetOwnerPrincipalId(), ClosedAt: closed}))
}
func (s *server) UpdateTalentReviewCycle(ctx context.Context, r *pb.UpdateTalentReviewCycleRequest) (*pb.TalentReviewCycleResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "cycle_code", "framework_id", "legal_entity_reference_id", "organization_unit_reference_id", "period_start", "period_end", "status", "owner_principal_id", "closed_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	start, err := timeField(f["period_start"], in.PeriodStart)
	if err != nil {
		return nil, response.Error(err)
	}
	end, err := timeField(f["period_end"], in.PeriodEnd)
	if err != nil {
		return nil, response.Error(err)
	}
	closed, err := timeField(f["closed_at"], in.ClosedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return cycleResult(s.service.UpdateTalentReviewCycle(ctx, r.Id, service.UpdateTalentReviewCycleInput{
		CycleCode: operation.Field[string]{Set: f["cycle_code"], Value: in.CycleCode}, FrameworkID: operation.Field[string]{Set: f["framework_id"], Value: in.FrameworkId},
		LegalEntityReferenceID:      operation.Field[string]{Set: f["legal_entity_reference_id"], Value: in.LegalEntityReferenceId},
		OrganizationUnitReferenceID: operation.Field[string]{Set: f["organization_unit_reference_id"], Value: in.OrganizationUnitReferenceId},
		PeriodStart:                 start, PeriodEnd: end, Status: operation.Field[string]{Set: f["status"], Value: in.Status},
		OwnerPrincipalID: operation.Field[string]{Set: f["owner_principal_id"], Value: in.OwnerPrincipalId}, ClosedAt: closed}))
}
func (s *server) DeleteTalentReviewCycle(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteTalentReviewCycle(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetTalentReviewCycleByID(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentReviewCycleResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycleResult(s.service.GetTalentReviewCycleByID(ctx, r.Id))
}
func (s *server) GetAllTalentReviewCycles(ctx context.Context, r *commonv1.PageRequest) (*pb.TalentReviewCyclesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(s.service.GetAllTalentReviewCycles(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentReviewCycleByCycleCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentReviewCyclesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(s.service.FindTalentReviewCycleByCycleCode(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentReviewCycleByFrameworkID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentReviewCyclesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(s.service.FindTalentReviewCycleByFrameworkID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentReviewCycleByLegalEntityReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentReviewCyclesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(s.service.FindTalentReviewCycleByLegalEntityReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentReviewCycleByOrganizationUnitReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentReviewCyclesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(s.service.FindTalentReviewCycleByOrganizationUnitReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentReviewCycleByStatus(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentReviewCyclesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(s.service.FindTalentReviewCycleByStatus(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentReviewCycleByOwnerPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentReviewCyclesResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return cycles(s.service.FindTalentReviewCycleByOwnerPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
