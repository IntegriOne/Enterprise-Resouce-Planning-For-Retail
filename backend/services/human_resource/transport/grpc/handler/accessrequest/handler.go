// Package accessrequest memetakan RPC permintaan akses; keputusan grant tetap milik IAM.
package accessrequest

import (
	"context"
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/accessrequest/v1"
	commonv1 "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/accessrequest"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedAccessRequestServiceServer
	service *service.Service
}

// Register mengikat handler pada application service yang telah disusun pemilik server.
func Register(registrar grpc.ServiceRegistrar, svc *service.Service) error {
	if err := operation.RequireDependency(registrar); err != nil {
		return err
	}
	if svc == nil {
		return hrerrors.ErrUnavailable
	}
	pb.RegisterAccessRequestServiceServer(registrar, &server{service: svc})
	return nil
}

func record(value *entity.AccessRequest) *pb.AccessRequest {
	if value == nil {
		return nil
	}
	var reconciled *timestamppb.Timestamp
	if value.LastReconciledAt != nil {
		reconciled = timestamppb.New(*value.LastReconciledAt)
	}
	return &pb.AccessRequest{
		Id: value.ID, WorkerId: value.WorkerID, LifecycleEventId: value.LifecycleEventID,
		RequestKind: value.RequestKind, RequestedScopeType: value.RequestedScopeType,
		RequestedScopeReferenceId: value.RequestedScopeReferenceID, ApprovedCatalogReferenceId: value.ApprovedCatalogReferenceID,
		IamRequestReferenceId: value.IAMRequestReferenceID, IamDecisionEventReferenceId: value.IAMDecisionEventReferenceID,
		Status: value.Status, RequestedByPrincipalId: value.RequestedByPrincipalID,
		RequestedAt: timestamppb.New(value.RequestedAt), LastReconciledAt: reconciled, IdempotencyKey: value.IdempotencyKey,
	}
}

func pageResult(page operation.Page[entity.AccessRequest]) *pb.AccessRequestsResponse {
	items := make([]*pb.AccessRequest, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, record(&page.Items[index]))
	}
	return &pb.AccessRequestsResponse{Items: items, Page: &commonv1.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}
}

func (s *server) CreateAccessRequest(ctx context.Context, request *pb.CreateAccessRequestRequest) (*pb.AccessRequestResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	reconciled, err := response.OptionalTimestamp(in.LastReconciledAt)
	if err != nil {
		return nil, response.Error(err)
	}
	value, err := s.service.CreateAccessRequest(ctx, service.CreateAccessRequestInput{
		WorkerID: in.GetWorkerId(), LifecycleEventID: in.GetLifecycleEventId(), RequestKind: in.GetRequestKind(),
		RequestedScopeType: in.GetRequestedScopeType(), RequestedScopeReferenceID: in.GetRequestedScopeReferenceId(),
		ApprovedCatalogReferenceID: in.ApprovedCatalogReferenceId, IAMRequestReferenceID: in.IamRequestReferenceId,
		IAMDecisionEventReferenceID: in.IamDecisionEventReferenceId, Status: in.Status,
		LastReconciledAt: reconciled, IdempotencyKey: in.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.AccessRequestResponse{Value: record(value)}, nil
}

func (s *server) UpdateAccessRequest(ctx context.Context, request *pb.UpdateAccessRequestRequest) (*pb.AccessRequestResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	selected, err := response.ValidateMask(request.UpdateMask,
		"worker_id", "lifecycle_event_id", "request_kind", "requested_scope_type", "requested_scope_reference_id",
		"approved_catalog_reference_id", "iam_request_reference_id", "iam_decision_event_reference_id", "status",
		"last_reconciled_at", "idempotency_key",
	)
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	var reconciled *time.Time
	if selected["last_reconciled_at"] {
		reconciled, err = response.OptionalTimestamp(in.LastReconciledAt)
		if err != nil {
			return nil, response.Error(err)
		}
	}
	value, err := s.service.UpdateAccessRequest(ctx, request.Id, service.UpdateAccessRequestInput{
		WorkerID:                    operation.Field[string]{Set: selected["worker_id"], Value: in.WorkerId},
		LifecycleEventID:            operation.Field[int64]{Set: selected["lifecycle_event_id"], Value: in.LifecycleEventId},
		RequestKind:                 operation.Field[string]{Set: selected["request_kind"], Value: in.RequestKind},
		RequestedScopeType:          operation.Field[string]{Set: selected["requested_scope_type"], Value: in.RequestedScopeType},
		RequestedScopeReferenceID:   operation.Field[string]{Set: selected["requested_scope_reference_id"], Value: in.RequestedScopeReferenceId},
		ApprovedCatalogReferenceID:  operation.Field[string]{Set: selected["approved_catalog_reference_id"], Value: in.ApprovedCatalogReferenceId},
		IAMRequestReferenceID:       operation.Field[string]{Set: selected["iam_request_reference_id"], Value: in.IamRequestReferenceId},
		IAMDecisionEventReferenceID: operation.Field[string]{Set: selected["iam_decision_event_reference_id"], Value: in.IamDecisionEventReferenceId},
		Status:                      operation.Field[string]{Set: selected["status"], Value: in.Status},
		LastReconciledAt:            operation.Field[time.Time]{Set: selected["last_reconciled_at"], Value: reconciled},
		IdempotencyKey:              operation.Field[string]{Set: selected["idempotency_key"], Value: in.IdempotencyKey},
	})
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.AccessRequestResponse{Value: record(value)}, nil
}

func (s *server) DeleteAccessRequest(ctx context.Context, request *commonv1.IDRequest) (*commonv1.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := s.service.DeleteAccessRequest(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &commonv1.EmptyResponse{}, nil
}
func (s *server) GetAccessRequestByID(ctx context.Context, request *commonv1.IDRequest) (*pb.AccessRequestResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	value, err := s.service.GetAccessRequestByID(ctx, request.Id)
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.AccessRequestResponse{Value: record(value)}, nil
}
func (s *server) GetAllAccessRequests(ctx context.Context, request *commonv1.PageRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.GetAllAccessRequests(ctx, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
func (s *server) FindAccessRequestByWorkerID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindAccessRequestByWorkerID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
func (s *server) FindAccessRequestByLifecycleEventID(ctx context.Context, request *commonv1.Int64FinderRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindAccessRequestByLifecycleEventID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
func (s *server) FindAccessRequestByRequestedScopeType(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindAccessRequestByRequestedScopeType(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
func (s *server) FindAccessRequestByRequestedScopeReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindAccessRequestByRequestedScopeReferenceID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
func (s *server) FindAccessRequestByApprovedCatalogReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindAccessRequestByApprovedCatalogReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
func (s *server) FindAccessRequestByIAMRequestReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindAccessRequestByIAMRequestReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
func (s *server) FindAccessRequestByIAMDecisionEventReferenceID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindAccessRequestByIAMDecisionEventReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
func (s *server) FindAccessRequestByStatus(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindAccessRequestByStatus(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
func (s *server) FindAccessRequestByRequestedByPrincipalID(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindAccessRequestByRequestedByPrincipalID(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
func (s *server) FindAccessRequestByIdempotencyKey(ctx context.Context, request *commonv1.StringFinderRequest) (*pb.AccessRequestsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	page, err := s.service.FindAccessRequestByIdempotencyKey(ctx, request.GetValue(), int(request.Limit), int(request.Offset))
	if err != nil {
		return nil, response.Error(err)
	}
	return pageResult(page), nil
}
