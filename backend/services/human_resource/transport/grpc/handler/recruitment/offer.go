package recruitment

import (
	"context"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/response"
)

func offer(v *entity.Offer) *pb.Offer {
	return &pb.Offer{
		Id: v.ID, ApplicationId: v.ApplicationID, OfferNumber: v.OfferNumber, OfferVersion: v.OfferVersion, OfferStatus: v.OfferStatus,
		ProposedJobId: v.ProposedJobID, ProposedLegalEntityReferenceId: v.ProposedLegalEntityReferenceID, ProposedStartDate: dateString(v.ProposedStartDate),
		ProposedContractType: v.ProposedContractType, ProposedCompensationDocumentVersionReferenceId: v.ProposedCompensationDocumentVersionReferenceID,
		CurrencyCode: v.CurrencyCode, OfferDocumentReferenceId: v.OfferDocumentReferenceID, OfferDocumentVersionReferenceId: v.OfferDocumentVersionReferenceID,
		OfferContentDigest: append([]byte(nil), v.OfferContentDigest...), IssuedAt: timestamp(v.IssuedAt), ExpiresAt: timestamp(v.ExpiresAt),
		CreatedByPrincipalId: v.CreatedByPrincipalID, CreatedAt: timestamp(v.CreatedAt),
		WithdrawnAt: timestamp(v.WithdrawnAt), WithdrawnByPrincipalId: v.WithdrawnByPrincipalID, WithdrawalReasonCode: v.WithdrawalReasonCode,
		CandidateResponse: v.CandidateResponse, RespondedAt: timestamp(v.RespondedAt), ApprovedByPrincipalId: v.ApprovedByPrincipalID, ApprovedAt: timestamp(v.ApprovedAt),
	}
}
func offerResponse(v *entity.Offer, err error) (*pb.OfferResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.OfferResponse{Item: offer(v)}, nil
}
func offers(page operation.Page[entity.Offer], err error) (*pb.OffersResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.Offer, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, offer(&page.Items[i]))
	}
	return &pb.OffersResponse{Items: items, Page: pageInfo(page)}, nil
}

// ApproveOffer meneruskan persetujuan penawaran ke service.
func (h *Handler) ApproveOffer(ctx context.Context, request *pb.ApproveOfferRequest) (*pb.OfferResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offerResponse(h.service.ApproveOffer(ctx, request.Id, request.CompensationDocumentVersionId))
}

// IssueOffer meneruskan penerbitan penawaran ke service.
func (h *Handler) IssueOffer(ctx context.Context, request *pb.IssueOfferRequest) (*pb.OfferResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offerResponse(h.service.IssueOffer(ctx, request.Id, request.DocumentId, request.DocumentVersionId, request.ContentDigest))
}

// AcceptOffer mencatat penerimaan penawaran yang telah dikonfirmasi kandidat.
func (h *Handler) AcceptOffer(ctx context.Context, request *pb.AcceptOfferRequest) (*pb.OfferResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offerResponse(h.service.AcceptOffer(ctx, request.Id, request.ResponseEvidenceId))
}

// DeclineOffer mencatat penolakan penawaran yang telah dikonfirmasi kandidat.
func (h *Handler) DeclineOffer(ctx context.Context, request *pb.DeclineOfferRequest) (*pb.OfferResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offerResponse(h.service.DeclineOffer(ctx, request.Id, request.ResponseEvidenceId))
}

// WithdrawOffer meneruskan penarikan penawaran ke service.
func (h *Handler) WithdrawOffer(ctx context.Context, request *pb.WithdrawOfferRequest) (*pb.OfferResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offerResponse(h.service.WithdrawOffer(ctx, request.Id, request.ReasonCode))
}

func (h *Handler) CreateOffer(ctx context.Context, r *pb.CreateOfferRequest) (*pb.OfferResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	start, err := optionalDate(in.ProposedStartDate)
	if err != nil {
		return nil, response.Error(err)
	}
	issued, err := response.OptionalTimestamp(in.IssuedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	expires, err := response.OptionalTimestamp(in.ExpiresAt)
	if err != nil {
		return nil, response.Error(err)
	}
	responded, err := response.OptionalTimestamp(in.RespondedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := response.OptionalTimestamp(in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return offerResponse(h.service.CreateOffer(ctx, service.CreateOfferInput{
		ApplicationID: in.GetApplicationId(), OfferNumber: in.GetOfferNumber(), OfferVersion: in.OfferVersion, OfferStatus: in.OfferStatus,
		ProposedJobID: in.GetProposedJobId(), ProposedLegalEntityReferenceID: in.GetProposedLegalEntityReferenceId(), ProposedStartDate: start,
		ProposedContractType: in.GetProposedContractType(), ProposedCompensationDocumentVersionReferenceID: in.ProposedCompensationDocumentVersionReferenceId,
		CurrencyCode: in.CurrencyCode, OfferDocumentReferenceID: in.OfferDocumentReferenceId, OfferDocumentVersionReferenceID: in.OfferDocumentVersionReferenceId,
		OfferContentDigest: append([]byte(nil), in.OfferContentDigest...), IssuedAt: issued, ExpiresAt: expires, CandidateResponse: in.CandidateResponse,
		RespondedAt: responded, ApprovedByPrincipalID: in.ApprovedByPrincipalId, ApprovedAt: approved,
	}))
}
func (h *Handler) UpdateOffer(ctx context.Context, r *pb.UpdateOfferRequest) (*pb.OfferResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "application_id", "offer_number", "offer_version", "offer_status", "proposed_job_id", "proposed_legal_entity_reference_id", "proposed_start_date", "proposed_contract_type", "proposed_compensation_document_version_reference_id", "currency_code", "offer_document_reference_id", "offer_document_version_reference_id", "offer_content_digest", "issued_at", "expires_at", "candidate_response", "responded_at", "approved_by_principal_id", "approved_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	start, err := dateField(m["proposed_start_date"], in.ProposedStartDate)
	if err != nil {
		return nil, response.Error(err)
	}
	issued, err := timeField(m["issued_at"], in.IssuedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	expires, err := timeField(m["expires_at"], in.ExpiresAt)
	if err != nil {
		return nil, response.Error(err)
	}
	responded, err := timeField(m["responded_at"], in.RespondedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	approved, err := timeField(m["approved_at"], in.ApprovedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return offerResponse(h.service.UpdateOffer(ctx, r.Id, service.UpdateOfferInput{
		ApplicationID: field(m["application_id"], in.ApplicationId), OfferNumber: field(m["offer_number"], in.OfferNumber), OfferVersion: field(m["offer_version"], in.OfferVersion), OfferStatus: field(m["offer_status"], in.OfferStatus),
		ProposedJobID: field(m["proposed_job_id"], in.ProposedJobId), ProposedLegalEntityReferenceID: field(m["proposed_legal_entity_reference_id"], in.ProposedLegalEntityReferenceId), ProposedStartDate: start,
		ProposedContractType: field(m["proposed_contract_type"], in.ProposedContractType), ProposedCompensationDocumentVersionReferenceID: field(m["proposed_compensation_document_version_reference_id"], in.ProposedCompensationDocumentVersionReferenceId),
		CurrencyCode: field(m["currency_code"], in.CurrencyCode), OfferDocumentReferenceID: field(m["offer_document_reference_id"], in.OfferDocumentReferenceId), OfferDocumentVersionReferenceID: field(m["offer_document_version_reference_id"], in.OfferDocumentVersionReferenceId),
		OfferContentDigest: bytesField(m["offer_content_digest"], in.OfferContentDigest), IssuedAt: issued, ExpiresAt: expires, CandidateResponse: field(m["candidate_response"], in.CandidateResponse),
		RespondedAt: responded, ApprovedByPrincipalID: field(m["approved_by_principal_id"], in.ApprovedByPrincipalId), ApprovedAt: approved,
	}))
}
func (h *Handler) DeleteOffer(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	return deleteRecord(ctx, r, h.service.DeleteOffer)
}
func (h *Handler) GetOfferByID(ctx context.Context, r *common.IDRequest) (*pb.OfferResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offerResponse(h.service.GetOfferByID(ctx, r.Id))
}
func (h *Handler) GetAllOffers(ctx context.Context, r *common.PageRequest) (*pb.OffersResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.GetAllOffers(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByApplicationID(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByApplicationID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByOfferNumber(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByOfferNumber(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByOfferStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByOfferStatus(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByProposedJobID(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByProposedJobID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByProposedLegalEntityReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByProposedLegalEntityReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByProposedContractType(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByProposedContractType(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByProposedCompensationDocumentVersionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByProposedCompensationDocumentVersionReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByCurrencyCode(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByCurrencyCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByOfferDocumentReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByOfferDocumentReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByOfferDocumentVersionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByOfferDocumentVersionReferenceID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindOfferByApprovedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.OffersResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return offers(h.service.FindOfferByApprovedByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
