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

func careerPath(v *entity.CareerPath) *pb.CareerPath {
	if v == nil {
		return nil
	}
	return &pb.CareerPath{Id: v.ID, PathCode: v.PathCode, PathVersion: v.PathVersion,
		PathName: v.PathName, JobFamilyId: v.JobFamilyID, ValidFrom: timestamppb.New(v.ValidFrom), ValidTo: timestamp(v.ValidTo),
		ApprovedByPrincipalId: v.ApprovedByPrincipalID, ApprovedAt: timestamp(v.ApprovedAt),
		Status: v.Status, CreatedByPrincipalId: v.CreatedByPrincipalID, CreatedAt: timestamppb.New(v.CreatedAt)}
}
func pathResult(v *entity.CareerPath, err error) (*pb.CareerPathResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.CareerPathResponse{Value: careerPath(v)}, nil
}
func paths(p operation.Page[entity.CareerPath], err error) (*pb.CareerPathsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.CareerPath, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, careerPath(&p.Items[i]))
	}
	return &pb.CareerPathsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateCareerPath(ctx context.Context, r *pb.CreateCareerPathRequest) (*pb.CareerPathResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	if in.ApprovedByPrincipalId != nil || in.ApprovedAt != nil {
		return nil, response.Error(hrerrors.ErrWorkflowRequired)
	}
	from, err := response.Timestamp(in.ValidFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := response.OptionalTimestamp(in.ValidTo)
	if err != nil {
		return nil, response.Error(err)
	}
	return pathResult(s.service.CreateCareerPath(ctx, service.CreateCareerPathInput{
		PathCode: in.GetPathCode(), PathVersion: in.GetPathVersion(), PathName: in.GetPathName(), JobFamilyID: in.GetJobFamilyId(),
		ValidFrom: from, ValidTo: to}))
}
func (s *server) UpdateCareerPath(ctx context.Context, r *pb.UpdateCareerPathRequest) (*pb.CareerPathResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if r.Input.ApprovedByPrincipalId != nil || r.Input.ApprovedAt != nil {
		return nil, response.Error(hrerrors.ErrWorkflowRequired)
	}
	f, err := response.ValidateMask(r.UpdateMask, "path_code", "path_version", "path_name", "job_family_id", "valid_from", "valid_to")
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
	return pathResult(s.service.UpdateCareerPath(ctx, r.Id, service.UpdateCareerPathInput{
		PathCode: operation.Field[string]{Set: f["path_code"], Value: in.PathCode}, PathVersion: operation.Field[int32]{Set: f["path_version"], Value: in.PathVersion},
		PathName: operation.Field[string]{Set: f["path_name"], Value: in.PathName}, JobFamilyID: operation.Field[string]{Set: f["job_family_id"], Value: in.JobFamilyId},
		ValidFrom: from, ValidTo: to}))
}

func (s *server) ApproveCareerPath(ctx context.Context, r *commonv1.IDRequest) (*pb.CareerPathResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return pathResult(s.service.ApproveCareerPath(ctx, r.Id))
}
func (s *server) DeleteCareerPath(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteCareerPath(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetCareerPathByID(ctx context.Context, r *commonv1.IDRequest) (*pb.CareerPathResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return pathResult(s.service.GetCareerPathByID(ctx, r.Id))
}
func (s *server) GetAllCareerPaths(ctx context.Context, r *commonv1.PageRequest) (*pb.CareerPathsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return paths(s.service.GetAllCareerPaths(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindCareerPathByPathCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CareerPathsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return paths(s.service.FindCareerPathByPathCode(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCareerPathByJobFamilyID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CareerPathsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return paths(s.service.FindCareerPathByJobFamilyID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCareerPathByStatus(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CareerPathsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return paths(s.service.FindCareerPathByStatus(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCareerPathByCreatedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CareerPathsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return paths(s.service.FindCareerPathByCreatedByPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindCareerPathByApprovedByPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.CareerPathsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return paths(s.service.FindCareerPathByApprovedByPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
