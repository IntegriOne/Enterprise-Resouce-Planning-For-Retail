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

func framework(v *entity.TalentFramework) *pb.TalentFramework {
	if v == nil {
		return nil
	}
	return &pb.TalentFramework{Id: v.ID, FrameworkCode: v.FrameworkCode, FrameworkVersion: v.FrameworkVersion,
		FrameworkName: v.FrameworkName, RubricDocumentVersionReferenceId: v.RubricDocumentVersionReferenceID, RubricDigest: v.RubricDigest,
		EffectiveFrom: timestamppb.New(v.EffectiveFrom), EffectiveTo: timestamp(v.EffectiveTo), ApprovedByPrincipalId: v.ApprovedByPrincipalID, ApprovedAt: timestamppb.New(v.ApprovedAt)}
}
func frameworkResult(v *entity.TalentFramework, err error) (*pb.TalentFrameworkResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.TalentFrameworkResponse{Value: framework(v)}, nil
}
func frameworks(p operation.Page[entity.TalentFramework], err error) (*pb.TalentFrameworksResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.TalentFramework, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, framework(&p.Items[i]))
	}
	return &pb.TalentFrameworksResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateTalentFramework(ctx context.Context, r *pb.CreateTalentFrameworkRequest) (*pb.TalentFrameworkResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	from, err := response.Timestamp(in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := response.OptionalTimestamp(in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.Timestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return frameworkResult(s.service.CreateTalentFramework(ctx, service.CreateTalentFrameworkInput{
		FrameworkCode: in.GetFrameworkCode(), FrameworkVersion: in.GetFrameworkVersion(), FrameworkName: in.GetFrameworkName(),
		RubricDocumentVersionReferenceID: in.GetRubricDocumentVersionReferenceId(), RubricDigest: in.RubricDigest,
		EffectiveFrom: from, EffectiveTo: to, ApprovedByPrincipalID: in.GetApprovedByPrincipalId(), ApprovedAt: approved}))
}
func (s *server) UpdateTalentFramework(ctx context.Context, r *pb.UpdateTalentFrameworkRequest) (*pb.TalentFrameworkResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "framework_code", "framework_version", "framework_name", "rubric_document_version_reference_id", "rubric_digest", "effective_from", "effective_to", "approved_by_principal_id", "approved_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	from, err := timeField(f["effective_from"], in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := timeField(f["effective_to"], in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := timeField(f["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return frameworkResult(s.service.UpdateTalentFramework(ctx, r.Id, service.UpdateTalentFrameworkInput{
		FrameworkCode: operation.Field[string]{Set: f["framework_code"], Value: in.FrameworkCode}, FrameworkVersion: operation.Field[int32]{Set: f["framework_version"], Value: in.FrameworkVersion},
		FrameworkName:                    operation.Field[string]{Set: f["framework_name"], Value: in.FrameworkName},
		RubricDocumentVersionReferenceID: operation.Field[string]{Set: f["rubric_document_version_reference_id"], Value: in.RubricDocumentVersionReferenceId},
		RubricDigest:                     bytesField(f["rubric_digest"], in.RubricDigest), EffectiveFrom: from, EffectiveTo: to,
		ApprovedByPrincipalID: operation.Field[string]{Set: f["approved_by_principal_id"], Value: in.ApprovedByPrincipalId}, ApprovedAt: approved}))
}
func (s *server) DeleteTalentFramework(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteTalentFramework(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetTalentFrameworkByID(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentFrameworkResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return frameworkResult(s.service.GetTalentFrameworkByID(ctx, r.Id))
}
func (s *server) GetAllTalentFrameworks(ctx context.Context, r *commonv1.PageRequest) (*pb.TalentFrameworksResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return frameworks(s.service.GetAllTalentFrameworks(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentFrameworkByFrameworkCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentFrameworksResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return frameworks(s.service.FindTalentFrameworkByFrameworkCode(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentFrameworkByRubricDocumentVersionReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentFrameworksResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return frameworks(s.service.FindTalentFrameworkByRubricDocumentVersionReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentFrameworkByApprovedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentFrameworksResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return frameworks(s.service.FindTalentFrameworkByApprovedByPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
