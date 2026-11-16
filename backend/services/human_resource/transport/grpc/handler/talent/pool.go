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

func pool(v *entity.TalentPool) *pb.TalentPool {
	if v == nil {
		return nil
	}
	return &pb.TalentPool{Id: v.ID, PoolCode: v.PoolCode, PoolVersion: v.PoolVersion, PoolName: v.PoolName,
		LegalEntityReferenceId: v.LegalEntityReferenceID, TargetJobId: v.TargetJobID, EligibilityDocumentVersionReferenceId: v.EligibilityDocumentVersionReferenceID,
		EligibilityDigest: v.EligibilityDigest, ValidFrom: timestamppb.New(v.ValidFrom), ValidTo: timestamp(v.ValidTo), OwnerPrincipalId: v.OwnerPrincipalID}
}
func poolResult(v *entity.TalentPool, err error) (*pb.TalentPoolResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.TalentPoolResponse{Value: pool(v)}, nil
}
func pools(p operation.Page[entity.TalentPool], err error) (*pb.TalentPoolsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.TalentPool, 0, len(p.Items))
	for i := range p.Items {
		items = append(items, pool(&p.Items[i]))
	}
	return &pb.TalentPoolsResponse{Items: items, Page: &commonv1.PageInfo{Total: p.Total, Limit: int32(p.Limit), Offset: int32(p.Offset)}}, nil
}
func (s *server) CreateTalentPool(ctx context.Context, r *pb.CreateTalentPoolRequest) (*pb.TalentPoolResponse, error) {
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
	return poolResult(s.service.CreateTalentPool(ctx, service.CreateTalentPoolInput{
		PoolCode: in.GetPoolCode(), PoolVersion: in.GetPoolVersion(), PoolName: in.GetPoolName(), LegalEntityReferenceID: in.GetLegalEntityReferenceId(),
		TargetJobID: in.TargetJobId, EligibilityDocumentVersionReferenceID: in.GetEligibilityDocumentVersionReferenceId(), EligibilityDigest: in.EligibilityDigest,
		ValidFrom: from, ValidTo: to, OwnerPrincipalID: in.GetOwnerPrincipalId()}))
}
func (s *server) UpdateTalentPool(ctx context.Context, r *pb.UpdateTalentPoolRequest) (*pb.TalentPoolResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	f, err := response.ValidateMask(r.UpdateMask, "pool_code", "pool_version", "pool_name", "legal_entity_reference_id", "target_job_id", "eligibility_document_version_reference_id", "eligibility_digest", "valid_from", "valid_to", "owner_principal_id")
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
	return poolResult(s.service.UpdateTalentPool(ctx, r.Id, service.UpdateTalentPoolInput{
		PoolCode: operation.Field[string]{Set: f["pool_code"], Value: in.PoolCode}, PoolVersion: operation.Field[int32]{Set: f["pool_version"], Value: in.PoolVersion},
		PoolName: operation.Field[string]{Set: f["pool_name"], Value: in.PoolName}, LegalEntityReferenceID: operation.Field[string]{Set: f["legal_entity_reference_id"], Value: in.LegalEntityReferenceId},
		TargetJobID:                           operation.Field[string]{Set: f["target_job_id"], Value: in.TargetJobId},
		EligibilityDocumentVersionReferenceID: operation.Field[string]{Set: f["eligibility_document_version_reference_id"], Value: in.EligibilityDocumentVersionReferenceId},
		EligibilityDigest:                     bytesField(f["eligibility_digest"], in.EligibilityDigest), ValidFrom: from, ValidTo: to,
		OwnerPrincipalID: operation.Field[string]{Set: f["owner_principal_id"], Value: in.OwnerPrincipalId}}))
}
func (s *server) DeleteTalentPool(ctx context.Context, r *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteTalentPool(ctx, r.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetTalentPoolByID(ctx context.Context, r *commonv1.IDRequest) (*pb.TalentPoolResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return poolResult(s.service.GetTalentPoolByID(ctx, r.Id))
}
func (s *server) GetAllTalentPools(ctx context.Context, r *commonv1.PageRequest) (*pb.TalentPoolsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return pools(s.service.GetAllTalentPools(ctx, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolByPoolCode(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return pools(s.service.FindTalentPoolByPoolCode(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolByLegalEntityReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return pools(s.service.FindTalentPoolByLegalEntityReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolByTargetJobID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return pools(s.service.FindTalentPoolByTargetJobID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolByEligibilityDocumentVersionReferenceID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return pools(s.service.FindTalentPoolByEligibilityDocumentVersionReferenceID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
func (s *server) FindTalentPoolByOwnerPrincipalID(ctx context.Context, r *commonv1.StringFinderRequest) (*pb.TalentPoolsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return pools(s.service.FindTalentPoolByOwnerPrincipalID(ctx, r.GetValue(), int(r.Limit), int(r.Offset)))
}
