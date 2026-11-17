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

func jobPosting(v *entity.JobPosting) *pb.JobPosting {
	return &pb.JobPosting{
		Id: v.ID, RequisitionId: v.RequisitionID, ChannelCode: v.ChannelCode, PostingVersion: v.PostingVersion, ExternalPostingReference: v.ExternalPostingReference,
		ContentDocumentVersionReferenceId: v.ContentDocumentVersionReferenceID, ContentDigest: append([]byte(nil), v.ContentDigest...),
		Status: v.Status, PublishedAt: timestamp(v.PublishedAt), ClosesAt: timestamp(v.ClosesAt), CreatedByPrincipalId: v.CreatedByPrincipalID,
	}
}
func jobPostingResponse(v *entity.JobPosting, err error) (*pb.JobPostingResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.JobPostingResponse{Item: jobPosting(v)}, nil
}
func jobPostings(page operation.Page[entity.JobPosting], err error) (*pb.JobPostingsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.JobPosting, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, jobPosting(&page.Items[i]))
	}
	return &pb.JobPostingsResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreateJobPosting(ctx context.Context, r *pb.CreateJobPostingRequest) (*pb.JobPostingResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	published, err := response.OptionalTimestamp(in.PublishedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	closes, err := response.OptionalTimestamp(in.ClosesAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return jobPostingResponse(h.service.CreateJobPosting(ctx, service.CreateJobPostingInput{
		RequisitionID: in.GetRequisitionId(), ChannelCode: in.GetChannelCode(), PostingVersion: in.GetPostingVersion(), ExternalPostingReference: in.ExternalPostingReference,
		ContentDocumentVersionReferenceID: in.GetContentDocumentVersionReferenceId(), ContentDigest: append([]byte(nil), in.ContentDigest...), Status: in.Status, PublishedAt: published, ClosesAt: closes,
	}))
}

func (h *Handler) PublishJobPosting(ctx context.Context, r *common.IDRequest) (*pb.JobPostingResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobPostingResponse(h.service.PublishJobPosting(ctx, r.Id))
}
func (h *Handler) UpdateJobPosting(ctx context.Context, r *pb.UpdateJobPostingRequest) (*pb.JobPostingResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "requisition_id", "channel_code", "posting_version", "external_posting_reference", "content_document_version_reference_id", "content_digest", "status", "published_at", "closes_at")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	published, err := timeField(m["published_at"], in.PublishedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	closes, err := timeField(m["closes_at"], in.ClosesAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return jobPostingResponse(h.service.UpdateJobPosting(ctx, r.Id, service.UpdateJobPostingInput{
		RequisitionID: field(m["requisition_id"], in.RequisitionId), ChannelCode: field(m["channel_code"], in.ChannelCode), PostingVersion: field(m["posting_version"], in.PostingVersion),
		ExternalPostingReference: field(m["external_posting_reference"], in.ExternalPostingReference), ContentDocumentVersionReferenceID: field(m["content_document_version_reference_id"], in.ContentDocumentVersionReferenceId),
		ContentDigest: bytesField(m["content_digest"], in.ContentDigest), Status: field(m["status"], in.Status), PublishedAt: published, ClosesAt: closes,
	}))
}
func (h *Handler) DeleteJobPosting(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	return deleteRecord(ctx, r, h.service.DeleteJobPosting)
}
func (h *Handler) GetJobPostingByID(ctx context.Context, r *common.IDRequest) (*pb.JobPostingResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobPostingResponse(h.service.GetJobPostingByID(ctx, r.Id))
}
func (h *Handler) GetAllJobPostings(ctx context.Context, r *common.PageRequest) (*pb.JobPostingsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobPostings(h.service.GetAllJobPostings(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindJobPostingByRequisitionID(ctx context.Context, r *common.StringFinderRequest) (*pb.JobPostingsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobPostings(h.service.FindJobPostingByRequisitionID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindJobPostingByChannelCode(ctx context.Context, r *common.StringFinderRequest) (*pb.JobPostingsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobPostings(h.service.FindJobPostingByChannelCode(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindJobPostingByContentDocumentVersionReferenceID(ctx context.Context, r *common.StringFinderRequest) (*pb.JobPostingsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobPostings(h.service.FindJobPostingByContentDocumentVersionReferenceID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindJobPostingByStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.JobPostingsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobPostings(h.service.FindJobPostingByStatus(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindJobPostingByCreatedByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.JobPostingsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return jobPostings(h.service.FindJobPostingByCreatedByPrincipalID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
