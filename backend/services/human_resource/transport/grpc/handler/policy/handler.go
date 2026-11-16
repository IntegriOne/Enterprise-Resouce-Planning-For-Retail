// Package policy memetakan RPC kebijakan HRM dan bukti penerimaan kebijakan.
package policy

import (
	"context"
	"time"

	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/policy/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/policy"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedPolicyServiceServer
	service *service.Service
}

// Register memasang kontrak bertipe pada application service kebijakan.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterPolicyServiceServer(registrar, &server{service: svc})
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
func bytesField(selected bool, value []byte) operation.Field[[]byte] {
	if value == nil {
		return operation.Field[[]byte]{Set: selected}
	}
	return operation.Field[[]byte]{Set: selected, Value: &value}
}
func record(value *entity.Policy) *pb.Policy {
	if value == nil {
		return nil
	}
	return &pb.Policy{
		Id: value.ID, PolicyCode: value.PolicyCode, PolicyVersion: value.PolicyVersion, PolicyName: value.PolicyName,
		CategoryCode: value.CategoryCode, LegalEntityReferenceId: value.LegalEntityReferenceID,
		EffectiveFrom: timestamppb.New(value.EffectiveFrom), EffectiveTo: timestamp(value.EffectiveTo),
		DocumentVersionReferenceId: value.DocumentVersionReferenceID, ContentDigest: value.ContentDigest,
		ApprovedByPrincipalId: value.ApprovedByPrincipalID, ApprovedAt: timestamppb.New(value.ApprovedAt),
	}
}
func acknowledgement(value *entity.PolicyAcknowledgement) *pb.PolicyAcknowledgement {
	if value == nil {
		return nil
	}
	return &pb.PolicyAcknowledgement{
		Id: value.ID, PolicyId: value.PolicyID, WorkerId: value.WorkerID, AssignedAt: timestamppb.New(value.AssignedAt),
		DueAt: timestamp(value.DueAt), AcknowledgedAt: timestamp(value.AcknowledgedAt),
		AcknowledgedByPrincipalId: value.AcknowledgedByPrincipalID, EvidenceDocumentVersionReferenceId: value.EvidenceDocumentVersionReferenceID,
	}
}
func records(page operation.Page[entity.Policy]) *pb.PoliciesResponse {
	items := make([]*pb.Policy, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, record(&page.Items[index]))
	}
	return &pb.PoliciesResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}
}
func acknowledgements(page operation.Page[entity.PolicyAcknowledgement]) *pb.PolicyAcknowledgementsResponse {
	items := make([]*pb.PolicyAcknowledgement, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, acknowledgement(&page.Items[index]))
	}
	return &pb.PolicyAcknowledgementsResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}
}

func (s *server) CreatePolicy(ctx context.Context, request *pb.CreatePolicyRequest) (*pb.PolicyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
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
	value, err := s.service.CreatePolicy(ctx, service.CreatePolicyInput{
		PolicyCode: in.GetPolicyCode(), PolicyVersion: in.GetPolicyVersion(), PolicyName: in.GetPolicyName(),
		CategoryCode: in.GetCategoryCode(), LegalEntityReferenceID: in.GetLegalEntityReferenceId(),
		EffectiveFrom: from, EffectiveTo: to, DocumentVersionReferenceID: in.GetDocumentVersionReferenceId(),
		ContentDigest: in.ContentDigest, ApprovedByPrincipalID: in.GetApprovedByPrincipalId(), ApprovedAt: approved,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PolicyResponse{Value: record(value)}, nil
}
func (s *server) UpdatePolicy(ctx context.Context, request *pb.UpdatePolicyRequest) (*pb.PolicyResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"policy_code", "policy_version", "policy_name", "category_code", "legal_entity_reference_id",
		"effective_from", "effective_to", "document_version_reference_id", "content_digest", "approved_by_principal_id", "approved_at",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	from, err := timeField(selected["effective_from"], in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := timeField(selected["effective_to"], in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := timeField(selected["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.UpdatePolicy(ctx, request.Id, service.UpdatePolicyInput{
		PolicyCode:             operation.Field[string]{Set: selected["policy_code"], Value: in.PolicyCode},
		PolicyVersion:          operation.Field[int32]{Set: selected["policy_version"], Value: in.PolicyVersion},
		PolicyName:             operation.Field[string]{Set: selected["policy_name"], Value: in.PolicyName},
		CategoryCode:           operation.Field[string]{Set: selected["category_code"], Value: in.CategoryCode},
		LegalEntityReferenceID: operation.Field[string]{Set: selected["legal_entity_reference_id"], Value: in.LegalEntityReferenceId},
		EffectiveFrom:          from, EffectiveTo: to,
		DocumentVersionReferenceID: operation.Field[string]{Set: selected["document_version_reference_id"], Value: in.DocumentVersionReferenceId},
		ContentDigest:              bytesField(selected["content_digest"], in.ContentDigest),
		ApprovedByPrincipalID:      operation.Field[string]{Set: selected["approved_by_principal_id"], Value: in.ApprovedByPrincipalId}, ApprovedAt: approved,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PolicyResponse{Value: record(value)}, nil
}
func (s *server) DeletePolicy(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeletePolicy(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetPolicyByID(ctx context.Context, request *commonv1.IDRequest) (*pb.PolicyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	value, err := s.service.GetPolicyByID(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PolicyResponse{Value: record(value)}, nil
}
func (s *server) GetAllPolicies(ctx context.Context, request *commonv1.PageRequest) (*pb.PoliciesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.GetAllPolicies(ctx, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return records(page), nil
}
func (s *server) FindPolicyByPolicyCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.PoliciesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindPolicyByPolicyCode(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return records(page), nil
}
func (s *server) FindPolicyByCategoryCode(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.PoliciesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindPolicyByCategoryCode(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return records(page), nil
}
func (s *server) FindPolicyByLegalEntityReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.PoliciesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindPolicyByLegalEntityReferenceID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return records(page), nil
}
func (s *server) FindPolicyByDocumentVersionReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.PoliciesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindPolicyByDocumentVersionReferenceID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return records(page), nil
}
func (s *server) FindPolicyByApprovedByPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.PoliciesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindPolicyByApprovedByPrincipalID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return records(page), nil
}

func (s *server) CreatePolicyAcknowledgement(ctx context.Context, request *pb.CreatePolicyAcknowledgementRequest) (*pb.PolicyAcknowledgementResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	due, err := response.OptionalTimestamp(in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	acknowledged, err := response.OptionalTimestamp(in.AcknowledgedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.CreatePolicyAcknowledgement(ctx, service.CreatePolicyAcknowledgementInput{
		PolicyID: in.GetPolicyId(), WorkerID: in.GetWorkerId(), DueAt: due, AcknowledgedAt: acknowledged,
		AcknowledgedByPrincipalID: in.AcknowledgedByPrincipalId, EvidenceDocumentVersionReferenceID: in.EvidenceDocumentVersionReferenceId,
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PolicyAcknowledgementResponse{Value: acknowledgement(value)}, nil
}
func (s *server) UpdatePolicyAcknowledgement(ctx context.Context, request *pb.UpdatePolicyAcknowledgementRequest) (*pb.PolicyAcknowledgementResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"policy_id", "worker_id", "due_at", "acknowledged_at", "acknowledged_by_principal_id", "evidence_document_version_reference_id",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	due, err := timeField(selected["due_at"], in.DueAt)
	if err != nil {
		return nil, response.Error(err)
	}
	acknowledged, err := timeField(selected["acknowledged_at"], in.AcknowledgedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.UpdatePolicyAcknowledgement(ctx, request.Id, service.UpdatePolicyAcknowledgementInput{
		PolicyID: operation.Field[string]{Set: selected["policy_id"], Value: in.PolicyId},
		WorkerID: operation.Field[string]{Set: selected["worker_id"], Value: in.WorkerId}, DueAt: due, AcknowledgedAt: acknowledged,
		AcknowledgedByPrincipalID:          operation.Field[string]{Set: selected["acknowledged_by_principal_id"], Value: in.AcknowledgedByPrincipalId},
		EvidenceDocumentVersionReferenceID: operation.Field[string]{Set: selected["evidence_document_version_reference_id"], Value: in.EvidenceDocumentVersionReferenceId},
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PolicyAcknowledgementResponse{Value: acknowledgement(value)}, nil
}
func (s *server) DeletePolicyAcknowledgement(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeletePolicyAcknowledgement(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetPolicyAcknowledgementByID(ctx context.Context, request *commonv1.IDRequest) (*pb.PolicyAcknowledgementResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	value, err := s.service.GetPolicyAcknowledgementByID(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.PolicyAcknowledgementResponse{Value: acknowledgement(value)}, nil
}
func (s *server) GetAllPolicyAcknowledgements(ctx context.Context, request *commonv1.PageRequest) (*pb.PolicyAcknowledgementsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.GetAllPolicyAcknowledgements(ctx, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return acknowledgements(page), nil
}
func (s *server) FindPolicyAcknowledgementByPolicyID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.PolicyAcknowledgementsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindPolicyAcknowledgementByPolicyID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return acknowledgements(page), nil
}
func (s *server) FindPolicyAcknowledgementByWorkerID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.PolicyAcknowledgementsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindPolicyAcknowledgementByWorkerID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return acknowledgements(page), nil
}
func (s *server) FindPolicyAcknowledgementByAcknowledgedByPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.PolicyAcknowledgementsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindPolicyAcknowledgementByAcknowledgedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return acknowledgements(page), nil
}
func (s *server) FindPolicyAcknowledgementByEvidenceDocumentVersionReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.PolicyAcknowledgementsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindPolicyAcknowledgementByEvidenceDocumentVersionReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return acknowledgements(page), nil
}
