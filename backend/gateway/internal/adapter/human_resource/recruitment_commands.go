package human_resource

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	recruitment "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
)

// StartApplicationScreening memulai tahap penyaringan awal lamaran kandidat.
func (a *Adapter) StartApplicationScreening(ctx context.Context, rawID string) (*model.Application, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_applications", "start_screening", "human_resource.applications.start_screening")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "application", "application")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.StartApplicationScreening(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "application")
	}
	return toModelApplication(resp.GetItem())
}

// ShortlistApplication memajukan lamaran ke daftar pendek.
func (a *Adapter) ShortlistApplication(ctx context.Context, rawID, reasonCode string) (*model.Application, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_applications", "advance", "human_resource.applications.advance")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "application", "application")
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(reasonCode)
	if reason == "" {
		return nil, middleware.ErrInvalidArgument("missing reason_code", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.ShortlistApplication(rpcCtx, &recruitment.ShortlistApplicationRequest{
		Id:         id,
		ReasonCode: reason,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "application")
	}
	return toModelApplication(resp.GetItem())
}

// RejectApplication menolak lamaran kandidat dengan kode alasan tertentu.
func (a *Adapter) RejectApplication(ctx context.Context, rawID, reasonCode string) (*model.Application, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_applications", "reject", "human_resource.applications.reject")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "application", "application")
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(reasonCode)
	if reason == "" {
		return nil, middleware.ErrInvalidArgument("missing reason_code", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.RejectApplication(rpcCtx, &recruitment.RejectApplicationRequest{
		Id:         id,
		ReasonCode: reason,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "application")
	}
	return toModelApplication(resp.GetItem())
}

// StartInterview memulai pelaksanaan wawancara kandidat.
func (a *Adapter) StartInterview(ctx context.Context, rawID string) (*model.Interview, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_interviews", "start", "human_resource.interviews.start")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "interview", "interview")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.StartInterview(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "interview")
	}
	return toModelInterview(resp.GetItem())
}

// CompleteInterview menandai wawancara kandidat telah selesai.
func (a *Adapter) CompleteInterview(ctx context.Context, rawID string) (*model.Interview, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_interviews", "complete", "human_resource.interviews.complete")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "interview", "interview")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.CompleteInterview(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "interview")
	}
	return toModelInterview(resp.GetItem())
}

// CancelInterview membatalkan jadwal wawancara kandidat.
func (a *Adapter) CancelInterview(ctx context.Context, rawID string) (*model.Interview, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_interviews", "cancel", "human_resource.interviews.cancel")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "interview", "interview")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.CancelInterview(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "interview")
	}
	return toModelInterview(resp.GetItem())
}

// ApproveOffer menyetujui penawaran kerja dengan dokumen kompensasi terkait.
func (a *Adapter) ApproveOffer(ctx context.Context, rawID, compDocVerID string) (*model.Offer, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_offers", "approve", "human_resource.offers.approve")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "offer", "offer")
	if err != nil {
		return nil, err
	}
	docVerID, err := normalizeRequiredUUID(compDocVerID, "document-version", "compensation_document_version_id")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.ApproveOffer(rpcCtx, &recruitment.ApproveOfferRequest{
		Id:                            id,
		CompensationDocumentVersionId: docVerID,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "offer")
	}
	return toModelOffer(resp.GetItem())
}

// IssueOffer menerbitkan surat penawaran resmi kepada kandidat.
func (a *Adapter) IssueOffer(ctx context.Context, rawID string, input model.IssueOfferInput) (*model.Offer, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_offers", "issue", "human_resource.offers.issue")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "offer", "offer")
	if err != nil {
		return nil, err
	}
	docID, err := normalizeRequiredUUID(input.DocumentID, "document", "document_id")
	if err != nil {
		return nil, err
	}
	docVerID, err := normalizeRequiredUUID(input.DocumentVersionID, "document-version", "document_version_id")
	if err != nil {
		return nil, err
	}
	digest, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input.ContentDigest))
	if err != nil || len(digest) == 0 {
		return nil, middleware.ErrInvalidArgument("invalid base64 for content_digest", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.IssueOffer(rpcCtx, &recruitment.IssueOfferRequest{
		Id:                id,
		DocumentId:        docID,
		DocumentVersionId: docVerID,
		ContentDigest:     digest,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "offer")
	}
	return toModelOffer(resp.GetItem())
}

// AcceptOffer mencatat penerimaan penawaran kerja oleh kandidat.
func (a *Adapter) AcceptOffer(ctx context.Context, rawID, responseEvidenceID string) (*model.Offer, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_offers", "accept", "human_resource.offers.accept")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "offer", "offer")
	if err != nil {
		return nil, err
	}
	evID, err := normalizeRequiredUUID(responseEvidenceID, "document-version", "response_evidence_id")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.AcceptOffer(rpcCtx, &recruitment.AcceptOfferRequest{
		Id:                 id,
		ResponseEvidenceId: evID,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "offer")
	}
	return toModelOffer(resp.GetItem())
}

// DeclineOffer mencatat penolakan penawaran kerja oleh kandidat.
func (a *Adapter) DeclineOffer(ctx context.Context, rawID, responseEvidenceID string) (*model.Offer, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_offers", "decline", "human_resource.offers.decline")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "offer", "offer")
	if err != nil {
		return nil, err
	}
	evID, err := normalizeRequiredUUID(responseEvidenceID, "document-version", "response_evidence_id")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.DeclineOffer(rpcCtx, &recruitment.DeclineOfferRequest{
		Id:                 id,
		ResponseEvidenceId: evID,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "offer")
	}
	return toModelOffer(resp.GetItem())
}

// WithdrawOffer menarik kembali penawaran kerja yang sudah diterbitkan.
func (a *Adapter) WithdrawOffer(ctx context.Context, rawID, reasonCode string) (*model.Offer, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_offers", "withdraw", "human_resource.offers.withdraw")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "offer", "offer")
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(reasonCode)
	if reason == "" {
		return nil, middleware.ErrInvalidArgument("missing reason_code", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.WithdrawOffer(rpcCtx, &recruitment.WithdrawOfferRequest{
		Id:         id,
		ReasonCode: reason,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "offer")
	}
	return toModelOffer(resp.GetItem())
}

// PublishJobPosting mempublikasikan lowongan pekerjaan ke kanal terkait.
func (a *Adapter) PublishJobPosting(ctx context.Context, rawID string) (*model.JobPosting, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_job_postings", "publish", "human_resource.job_postings.publish")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "job-posting", "job-posting")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.PublishJobPosting(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "job-posting")
	}
	return toModelJobPosting(resp.GetItem())
}

// CloseJobPosting menutup publikasi lowongan pekerjaan.
func (a *Adapter) CloseJobPosting(ctx context.Context, rawID string) (*model.JobPosting, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_job_postings", "close", "human_resource.job_postings.close")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "job-posting", "job-posting")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.CloseJobPosting(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapHRMWriteError(err, "job-posting")
	}
	return toModelJobPosting(resp.GetItem())
}

// WithdrawJobPosting menarik kembali publikasi lowongan pekerjaan.
func (a *Adapter) WithdrawJobPosting(ctx context.Context, rawID, reasonCode string) (*model.JobPosting, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_job_postings", "withdraw", "human_resource.job_postings.withdraw")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "job-posting", "job-posting")
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(reasonCode)
	if reason == "" {
		return nil, middleware.ErrInvalidArgument("missing reason_code", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.WithdrawJobPosting(rpcCtx, &recruitment.WithdrawJobPostingRequest{
		Id:         id,
		ReasonCode: reason,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "job-posting")
	}
	return toModelJobPosting(resp.GetItem())
}

// CompleteApplicationAssessment menyelesaikan asesmen lamaran kandidat.
func (a *Adapter) CompleteApplicationAssessment(ctx context.Context, rawID string, input *model.CompleteApplicationAssessmentInput) (*model.ApplicationAssessment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_application_assessments", "complete", "human_resource.application_assessments.complete")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "application-assessment", "application-assessment")
	if err != nil {
		return nil, err
	}
	req := &recruitment.CompleteApplicationAssessmentRequest{Id: id}
	if input != nil {
		req.NormalizedScore = input.NormalizedScore
		req.RecommendationCode = input.RecommendationCode
		docVerID, err := normalizeOptionalUUID(input.EvidenceDocumentVersionReferenceID, "document-version", "evidence_document_version_reference_id")
		if err != nil {
			return nil, err
		}
		req.EvidenceDocumentVersionReferenceId = docVerID
		digest, err := optionalBase64Bytes(input.EvidenceDigest, "evidence_digest")
		if err != nil {
			return nil, err
		}
		req.EvidenceDigest = digest
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.CompleteApplicationAssessment(rpcCtx, req)
	if err != nil {
		return nil, mapHRMWriteError(err, "application-assessment")
	}
	return toModelApplicationAssessment(resp.GetItem())
}
