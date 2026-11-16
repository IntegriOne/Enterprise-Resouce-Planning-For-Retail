package personnel

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func workerDocument(value *entity.WorkerDocument) *pb.WorkerDocument {
	item := &pb.WorkerDocument{Id: value.ID, WorkerId: value.WorkerID, DocumentType: value.DocumentType, DocumentReferenceId: value.DocumentReferenceID,
		DocumentVersionReferenceId: value.DocumentVersionReferenceID, ContentDigest: value.ContentDigest, DocumentNumberDigest: value.DocumentNumberDigest,
		IssuingCountryCode: value.IssuingCountryCode, IssuedOn: dateString(value.IssuedOn), ExpiresOn: dateString(value.ExpiresOn), VerificationStatus: value.VerificationStatus,
		VerifiedByPrincipalId: value.VerifiedByPrincipalID, RetentionClass: value.RetentionClass, AccessClass: value.AccessClass, CreatedAt: timestamppb.New(value.CreatedAt)}
	if value.VerifiedAt != nil {
		item.VerifiedAt = timestamppb.New(*value.VerifiedAt)
	}
	return item
}
func documentResponse(value *entity.WorkerDocument, err error) (*pb.WorkerDocumentResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.WorkerDocumentResponse{Item: workerDocument(value)}, nil
}
func documents(page operation.Page[entity.WorkerDocument], err error) (*pb.WorkerDocumentsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.WorkerDocument, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, workerDocument(&page.Items[index]))
	}
	return &pb.WorkerDocumentsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateWorkerDocument(ctx context.Context, request *pb.CreateWorkerDocumentRequest) (*pb.WorkerDocumentResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	issued, err := optionalDate(in.IssuedOn)
	if err != nil {
		return nil, response.Error(err)
	}
	expires, err := optionalDate(in.ExpiresOn)
	if err != nil {
		return nil, response.Error(err)
	}
	verified, err := response.OptionalTimestamp(in.VerifiedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return documentResponse(h.service.CreateWorkerDocument(ctx, service.CreateWorkerDocumentInput{
		WorkerID: in.GetWorkerId(), DocumentType: in.GetDocumentType(), DocumentReferenceID: in.GetDocumentReferenceId(), DocumentVersionReferenceID: in.GetDocumentVersionReferenceId(),
		ContentDigest: in.ContentDigest, DocumentNumberDigest: in.DocumentNumberDigest, IssuingCountryCode: in.IssuingCountryCode, IssuedOn: issued, ExpiresOn: expires,
		VerificationStatus: in.VerificationStatus, VerifiedByPrincipalID: in.VerifiedByPrincipalId, VerifiedAt: verified, RetentionClass: in.GetRetentionClass(), AccessClass: in.AccessClass,
	}))
}
func (h *Handler) UpdateWorkerDocument(ctx context.Context, request *pb.UpdateWorkerDocumentRequest) (*pb.WorkerDocumentResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_id", "document_type", "document_reference_id", "document_version_reference_id", "content_digest", "document_number_digest", "issuing_country_code", "issued_on", "expires_on", "verification_status", "verified_by_principal_id", "verified_at", "retention_class", "access_class")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	issued, err := dateField(mask["issued_on"], in.IssuedOn)
	if err != nil {
		return nil, response.Error(err)
	}
	expires, err := dateField(mask["expires_on"], in.ExpiresOn)
	if err != nil {
		return nil, response.Error(err)
	}
	patch := service.UpdateWorkerDocumentInput{WorkerID: field(mask["worker_id"], in.WorkerId), DocumentType: field(mask["document_type"], in.DocumentType),
		DocumentReferenceID: field(mask["document_reference_id"], in.DocumentReferenceId), DocumentVersionReferenceID: field(mask["document_version_reference_id"], in.DocumentVersionReferenceId),
		ContentDigest: bytesField(mask["content_digest"], in.ContentDigest), DocumentNumberDigest: bytesField(mask["document_number_digest"], in.DocumentNumberDigest),
		IssuingCountryCode: field(mask["issuing_country_code"], in.IssuingCountryCode), IssuedOn: issued, ExpiresOn: expires, VerificationStatus: field(mask["verification_status"], in.VerificationStatus),
		VerifiedByPrincipalID: field(mask["verified_by_principal_id"], in.VerifiedByPrincipalId), RetentionClass: field(mask["retention_class"], in.RetentionClass), AccessClass: field(mask["access_class"], in.AccessClass)}
	if mask["verified_at"] {
		value, err := response.OptionalTimestamp(in.VerifiedAt)
		if err != nil {
			return nil, response.Error(err)
		}
		patch.VerifiedAt = field(true, value)
	}
	return documentResponse(h.service.UpdateWorkerDocument(ctx, request.Id, patch))
}
func (h *Handler) DeleteWorkerDocument(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteWorkerDocument(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetWorkerDocumentByID(ctx context.Context, request *common.IDRequest) (*pb.WorkerDocumentResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return documentResponse(h.service.GetWorkerDocumentByID(ctx, request.Id))
}
func (h *Handler) GetAllWorkerDocuments(ctx context.Context, request *common.PageRequest) (*pb.WorkerDocumentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return documents(h.service.GetAllWorkerDocuments(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerDocumentByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerDocumentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return documents(h.service.FindWorkerDocumentByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerDocumentByDocumentType(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerDocumentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return documents(h.service.FindWorkerDocumentByDocumentType(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerDocumentByDocumentReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerDocumentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return documents(h.service.FindWorkerDocumentByDocumentReferenceID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerDocumentByDocumentVersionReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerDocumentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return documents(h.service.FindWorkerDocumentByDocumentVersionReferenceID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerDocumentByIssuingCountryCode(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerDocumentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return documents(h.service.FindWorkerDocumentByIssuingCountryCode(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerDocumentByVerificationStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerDocumentsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return documents(h.service.FindWorkerDocumentByVerificationStatus(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerDocumentByVerifiedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerDocumentsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return documents(h.service.FindWorkerDocumentByVerifiedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
