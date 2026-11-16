package competency

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/competency/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/competency"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func qualification(value *entity.WorkerQualification) *pb.WorkerQualification {
	return &pb.WorkerQualification{
		Id: value.ID, WorkerId: value.WorkerID, QualificationType: value.QualificationType,
		QualificationCode: value.QualificationCode, Title: value.Title,
		IssuerPartyReferenceId: value.IssuerPartyReferenceID, CredentialReferenceId: value.CredentialReferenceID,
		ObtainedOn: dateString(value.ObtainedOn), ExpiresOn: dateString(value.ExpiresOn),
		VerificationStatus: value.VerificationStatus, VerifiedByPrincipalId: value.VerifiedByPrincipalID,
		VerifiedAt: timestamp(value.VerifiedAt),
	}
}
func qualificationResponse(value *entity.WorkerQualification, err error) (*pb.WorkerQualificationResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.WorkerQualificationResponse{Item: qualification(value)}, nil
}
func qualifications(page operation.Page[entity.WorkerQualification], err error) (*pb.WorkerQualificationsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.WorkerQualification, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, qualification(&page.Items[index]))
	}
	return &pb.WorkerQualificationsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateWorkerQualification(ctx context.Context, request *pb.CreateWorkerQualificationRequest) (*pb.WorkerQualificationResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	obtained, err := optionalDate(in.ObtainedOn)
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
	return qualificationResponse(h.service.CreateWorkerQualification(ctx, service.CreateWorkerQualificationInput{
		WorkerID: in.GetWorkerId(), QualificationType: in.GetQualificationType(),
		QualificationCode: in.QualificationCode, Title: in.GetTitle(),
		IssuerPartyReferenceID: in.IssuerPartyReferenceId, CredentialReferenceID: in.CredentialReferenceId,
		ObtainedOn: obtained, ExpiresOn: expires, VerificationStatus: in.VerificationStatus,
		VerifiedByPrincipalID: in.VerifiedByPrincipalId, VerifiedAt: verified,
	}))
}
func (h *Handler) UpdateWorkerQualification(ctx context.Context, request *pb.UpdateWorkerQualificationRequest) (*pb.WorkerQualificationResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_id", "qualification_type", "qualification_code", "title", "issuer_party_reference_id", "credential_reference_id", "obtained_on", "expires_on", "verification_status", "verified_by_principal_id", "verified_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	obtained, err := dateField(mask["obtained_on"], in.ObtainedOn)
	if err != nil {
		return nil, response.Error(err)
	}
	expires, err := dateField(mask["expires_on"], in.ExpiresOn)
	if err != nil {
		return nil, response.Error(err)
	}
	verified, err := timestampField(mask["verified_at"], in.VerifiedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return qualificationResponse(h.service.UpdateWorkerQualification(ctx, request.Id, service.UpdateWorkerQualificationInput{
		WorkerID:               field(mask["worker_id"], in.WorkerId),
		QualificationType:      field(mask["qualification_type"], in.QualificationType),
		QualificationCode:      field(mask["qualification_code"], in.QualificationCode),
		Title:                  field(mask["title"], in.Title),
		IssuerPartyReferenceID: field(mask["issuer_party_reference_id"], in.IssuerPartyReferenceId),
		CredentialReferenceID:  field(mask["credential_reference_id"], in.CredentialReferenceId),
		ObtainedOn:             obtained, ExpiresOn: expires,
		VerificationStatus:    field(mask["verification_status"], in.VerificationStatus),
		VerifiedByPrincipalID: field(mask["verified_by_principal_id"], in.VerifiedByPrincipalId),
		VerifiedAt:            verified,
	}))
}
func (h *Handler) DeleteWorkerQualification(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteWorkerQualification(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetWorkerQualificationByID(ctx context.Context, request *common.IDRequest) (*pb.WorkerQualificationResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return qualificationResponse(h.service.GetWorkerQualificationByID(ctx, request.Id))
}
func (h *Handler) GetAllWorkerQualifications(ctx context.Context, request *common.PageRequest) (*pb.WorkerQualificationsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return qualifications(h.service.GetAllWorkerQualifications(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerQualificationByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerQualificationsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return qualifications(h.service.FindWorkerQualificationByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerQualificationByQualificationType(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerQualificationsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return qualifications(h.service.FindWorkerQualificationByQualificationType(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerQualificationByQualificationCode(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerQualificationsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return qualifications(h.service.FindWorkerQualificationByQualificationCode(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerQualificationByIssuerPartyReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerQualificationsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return qualifications(h.service.FindWorkerQualificationByIssuerPartyReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerQualificationByCredentialReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerQualificationsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return qualifications(h.service.FindWorkerQualificationByCredentialReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerQualificationByVerificationStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerQualificationsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return qualifications(h.service.FindWorkerQualificationByVerificationStatus(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindWorkerQualificationByVerifiedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.WorkerQualificationsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return qualifications(h.service.FindWorkerQualificationByVerifiedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
