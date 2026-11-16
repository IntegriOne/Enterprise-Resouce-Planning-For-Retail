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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func candidate(value *entity.Candidate) *pb.Candidate {
	return &pb.Candidate{Id: value.ID, CandidateNumber: value.CandidateNumber, PersonPartyReferenceId: value.PersonPartyReferenceID,
		PreferredName: value.PreferredName, ContactReferenceId: value.ContactReferenceID, SourceType: value.SourceType,
		SourceReference: value.SourceReference, PrivacyNoticeVersion: value.PrivacyNoticeVersion,
		NoticeAcknowledgedAt: timestamp(value.NoticeAcknowledgedAt), DataRetentionUntil: dateString(value.DataRetentionUntil),
		CandidateStatus: value.CandidateStatus, CreatedAt: timestamppb.New(value.CreatedAt), CreatedByPrincipalId: value.CreatedByPrincipalID}
}
func candidateResponse(value *entity.Candidate, err error) (*pb.CandidateResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.CandidateResponse{Item: candidate(value)}, nil
}
func candidates(page operation.Page[entity.Candidate], err error) (*pb.CandidatesResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.Candidate, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, candidate(&page.Items[i]))
	}
	return &pb.CandidatesResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreateCandidate(ctx context.Context, request *pb.CreateCandidateRequest) (*pb.CandidateResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := request.Input
	notice, err := response.OptionalTimestamp(in.NoticeAcknowledgedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	retention, err := optionalDate(in.DataRetentionUntil)
	if err != nil {
		return nil, response.Error(err)
	}
	return candidateResponse(h.service.CreateCandidate(ctx, service.CreateCandidateInput{
		CandidateNumber: in.GetCandidateNumber(), PersonPartyReferenceID: in.PersonPartyReferenceId, PreferredName: in.PreferredName,
		ContactReferenceID: in.ContactReferenceId, SourceType: in.SourceType, SourceReference: in.SourceReference,
		PrivacyNoticeVersion: in.GetPrivacyNoticeVersion(), NoticeAcknowledgedAt: notice, DataRetentionUntil: retention, CandidateStatus: in.CandidateStatus,
	}))
}
func (h *Handler) UpdateCandidate(ctx context.Context, request *pb.UpdateCandidateRequest) (*pb.CandidateResponse, error) {
	if request == nil || request.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(request.UpdateMask, "candidate_number", "person_party_reference_id", "preferred_name", "contact_reference_id", "source_type", "source_reference", "privacy_notice_version", "notice_acknowledged_at", "data_retention_until", "candidate_status")
	if err != nil {
		return nil, response.Error(err)
	}
	in := request.Input
	notice, err := timeField(m["notice_acknowledged_at"], in.NoticeAcknowledgedAt)
	if err != nil {
		return nil, response.Error(err)
	}
	retention, err := dateField(m["data_retention_until"], in.DataRetentionUntil)
	if err != nil {
		return nil, response.Error(err)
	}
	return candidateResponse(h.service.UpdateCandidate(ctx, request.Id, service.UpdateCandidateInput{
		CandidateNumber: field(m["candidate_number"], in.CandidateNumber), PersonPartyReferenceID: field(m["person_party_reference_id"], in.PersonPartyReferenceId),
		PreferredName: field(m["preferred_name"], in.PreferredName), ContactReferenceID: field(m["contact_reference_id"], in.ContactReferenceId),
		SourceType: field(m["source_type"], in.SourceType), SourceReference: field(m["source_reference"], in.SourceReference),
		PrivacyNoticeVersion: field(m["privacy_notice_version"], in.PrivacyNoticeVersion), NoticeAcknowledgedAt: notice, DataRetentionUntil: retention,
		CandidateStatus: field(m["candidate_status"], in.CandidateStatus),
	}))
}
func (h *Handler) DeleteCandidate(ctx context.Context, request *common.IDRequest) (*common.EmptyResponse, error) {
	return deleteRecord(ctx, request, h.service.DeleteCandidate)
}
func (h *Handler) GetCandidateByID(ctx context.Context, request *common.IDRequest) (*pb.CandidateResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return candidateResponse(h.service.GetCandidateByID(ctx, request.Id))
}
func (h *Handler) GetAllCandidates(ctx context.Context, request *common.PageRequest) (*pb.CandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return candidates(h.service.GetAllCandidates(ctx, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCandidateByCandidateNumber(ctx context.Context, request *common.StringFinderRequest) (*pb.CandidatesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return candidates(h.service.FindCandidateByCandidateNumber(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCandidateByPersonPartyReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.CandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return candidates(h.service.FindCandidateByPersonPartyReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCandidateByContactReferenceID(ctx context.Context, request *common.StringFinderRequest) (*pb.CandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return candidates(h.service.FindCandidateByContactReferenceID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCandidateBySourceType(ctx context.Context, request *common.StringFinderRequest) (*pb.CandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return candidates(h.service.FindCandidateBySourceType(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCandidateByCandidateStatus(ctx context.Context, request *common.StringFinderRequest) (*pb.CandidatesResponse, error) {
	if request == nil || request.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return candidates(h.service.FindCandidateByCandidateStatus(ctx, *request.Value, int(request.Limit), int(request.Offset)))
}
func (h *Handler) FindCandidateByCreatedByPrincipalID(ctx context.Context, request *common.StringFinderRequest) (*pb.CandidatesResponse, error) {
	if request == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return candidates(h.service.FindCandidateByCreatedByPrincipalID(ctx, request.Value, int(request.Limit), int(request.Offset)))
}
