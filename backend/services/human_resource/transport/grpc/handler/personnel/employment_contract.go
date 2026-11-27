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

func employmentContract(value *entity.EmploymentContract) *pb.EmploymentContract {
	item := &pb.EmploymentContract{Id: value.ID, WorkerId: value.WorkerID, ContractNumber: value.ContractNumber, ContractType: value.ContractType,
		ContractStatus: value.ContractStatus, EmployingLegalEntityReferenceId: value.EmployingLegalEntityReferenceID, EffectiveFrom: value.EffectiveFrom.Format("2006-01-02"),
		EffectiveTo: dateString(value.EffectiveTo), SignedDocumentReferenceId: value.SignedDocumentReferenceID, SignedDocumentVersionReferenceId: value.SignedDocumentVersionReferenceID,
		SignedContentDigest: value.SignedContentDigest, ProbationEndDate: dateString(value.ProbationEndDate), NoticePeriodDays: value.NoticePeriodDays,
		WorkCountryCode: value.WorkCountryCode, GoverningJurisdictionCode: value.GoverningJurisdictionCode, CollectiveAgreementId: value.CollectiveAgreementID,
		CreatedAt: timestamppb.New(value.CreatedAt)}
	if value.SignedAt != nil {
		item.SignedAt = timestamppb.New(*value.SignedAt)
	}
	return item
}
func contractResponse(value *entity.EmploymentContract, err error) (*pb.EmploymentContractResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.EmploymentContractResponse{Item: employmentContract(value)}, nil
}
func contracts(page operation.Page[entity.EmploymentContract], err error) (*pb.EmploymentContractsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.EmploymentContract, 0, len(page.Items))
	for index := range page.Items {
		items = append(items, employmentContract(&page.Items[index]))
	}
	return &pb.EmploymentContractsResponse{Items: items, Page: &common.PageInfo{Total: page.Total, Limit: int32(page.Limit), Offset: int32(page.Offset)}}, nil
}
func (h *Handler) CreateEmploymentContract(ctx context.Context, request *pb.CreateEmploymentContractRequest) (*pb.EmploymentContractResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	from, err := response.Date(in.GetEffectiveFrom())
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := optionalDate(in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	probation, err := optionalDate(in.ProbationEndDate)
	if err != nil {
		return nil, response.Error(err)
	}
	signedAt, err := response.OptionalTimestamp(in.SignedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return contractResponse(h.service.CreateEmploymentContract(ctx, service.CreateEmploymentContractInput{
		WorkerID: in.GetWorkerId(), ContractNumber: in.GetContractNumber(), ContractType: in.GetContractType(), ContractStatus: in.ContractStatus,
		EmployingLegalEntityReferenceID: in.GetEmployingLegalEntityReferenceId(), EffectiveFrom: from, EffectiveTo: to, SignedAt: signedAt,
		SignedDocumentReferenceID: in.SignedDocumentReferenceId, SignedDocumentVersionReferenceID: in.SignedDocumentVersionReferenceId, SignedContentDigest: in.SignedContentDigest,
		ProbationEndDate: probation, NoticePeriodDays: in.NoticePeriodDays, WorkCountryCode: in.GetWorkCountryCode(), GoverningJurisdictionCode: in.GetGoverningJurisdictionCode(), CollectiveAgreementID: in.CollectiveAgreementId,
	}))
}
func (h *Handler) UpdateEmploymentContract(ctx context.Context, request *pb.UpdateEmploymentContractRequest) (*pb.EmploymentContractResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	mask, err := response.ValidateMask(request.UpdateMask, "worker_id", "contract_number", "contract_type", "contract_status", "employing_legal_entity_reference_id", "effective_from", "effective_to", "signed_at", "signed_document_reference_id", "signed_document_version_reference_id", "signed_content_digest", "probation_end_date", "notice_period_days", "work_country_code", "governing_jurisdiction_code", "collective_agreement_id")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	from, err := dateField(mask["effective_from"], in.EffectiveFrom)
	if err != nil {
		return nil, response.Error(err)
	}
	to, err := dateField(mask["effective_to"], in.EffectiveTo)
	if err != nil {
		return nil, response.Error(err)
	}
	probation, err := dateField(mask["probation_end_date"], in.ProbationEndDate)
	if err != nil {
		return nil, response.Error(err)
	}
	patch := service.UpdateEmploymentContractInput{WorkerID: field(mask["worker_id"], in.WorkerId), ContractNumber: field(mask["contract_number"], in.ContractNumber),
		ContractType: field(mask["contract_type"], in.ContractType), ContractStatus: field(mask["contract_status"], in.ContractStatus), EmployingLegalEntityReferenceID: field(mask["employing_legal_entity_reference_id"], in.EmployingLegalEntityReferenceId),
		EffectiveFrom: from, EffectiveTo: to, SignedDocumentReferenceID: field(mask["signed_document_reference_id"], in.SignedDocumentReferenceId), SignedDocumentVersionReferenceID: field(mask["signed_document_version_reference_id"], in.SignedDocumentVersionReferenceId),
		SignedContentDigest: bytesField(mask["signed_content_digest"], in.SignedContentDigest), ProbationEndDate: probation, NoticePeriodDays: field(mask["notice_period_days"], in.NoticePeriodDays),
		WorkCountryCode: field(mask["work_country_code"], in.WorkCountryCode), GoverningJurisdictionCode: field(mask["governing_jurisdiction_code"], in.GoverningJurisdictionCode), CollectiveAgreementID: field(mask["collective_agreement_id"], in.CollectiveAgreementId)}
	if mask["signed_at"] {
		value, err := response.OptionalTimestamp(in.SignedAt)
		if err != nil {
			return nil, response.Error(err)
		}
		patch.SignedAt = field(true, value)
	}
	return contractResponse(h.service.UpdateEmploymentContract(ctx, request.Id, patch))
}
func (h *Handler) DeleteEmploymentContract(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	if err := h.service.DeleteEmploymentContract(ctx, request.Id); err != nil {
		return nil, response.Error(err)
	}
	return &common.EmptyResponse{}, nil
}
func (h *Handler) GetEmploymentContractByID(ctx context.Context, request *common.IDRequest) (*pb.EmploymentContractResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contractResponse(h.service.GetEmploymentContractByID(ctx, request.Id))
}
func (h *Handler) GetAllEmploymentContracts(ctx context.Context, request *common.PageRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.GetAllEmploymentContracts(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmploymentContractByWorkerID(ctx context.Context, request *common.StringFinderRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.FindEmploymentContractByWorkerID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmploymentContractByContractNumber(ctx context.Context, request *common.StringFinderRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.FindEmploymentContractByContractNumber(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmploymentContractByContractType(ctx context.Context, request *common.StringFinderRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.FindEmploymentContractByContractType(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmploymentContractByContractStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.FindEmploymentContractByContractStatus(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmploymentContractByEmployingLegalEntityReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.FindEmploymentContractByEmployingLegalEntityReferenceID(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmploymentContractBySignedDocumentReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.FindEmploymentContractBySignedDocumentReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmploymentContractBySignedDocumentVersionReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.FindEmploymentContractBySignedDocumentVersionReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmploymentContractByWorkCountryCode(ctx context.Context, request *common.StringFinderRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.FindEmploymentContractByWorkCountryCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmploymentContractByGoverningJurisdictionCode(ctx context.Context, request *common.StringFinderRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.FindEmploymentContractByGoverningJurisdictionCode(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindEmploymentContractByCollectiveAgreementID(ctx context.Context, request *common.StringFinderRequest) (*pb.EmploymentContractsResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return contracts(h.service.FindEmploymentContractByCollectiveAgreementID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}

// RecordSignedEmploymentContract mencatat versi kontrak bertanda tangan yang telah diverifikasi.
func (h *Handler) RecordSignedEmploymentContract(ctx context.Context, request *pb.RecordSignedEmploymentContractRequest) (*pb.EmploymentContractResponse, error) {
	return contractResponse(h.service.RecordSignedEmploymentContract(ctx, request.GetContractId(), request.GetDocumentVersionId()))
}
