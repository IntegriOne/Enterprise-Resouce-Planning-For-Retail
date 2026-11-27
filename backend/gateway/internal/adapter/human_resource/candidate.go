package human_resource

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	recruitment "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateCandidate mencatat kandidat dan meneruskan referensi ke pemeriksaan HRM.
func (a *Adapter) CreateCandidate(ctx context.Context, input model.CreateCandidateInput) (*model.Candidate, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.candidates.create")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}

	response, err := a.recruitment.CreateCandidate(rpcCtx, &recruitment.CreateCandidateRequest{
		Input: &recruitment.CandidateInput{
			CandidateNumber: &input.CandidateNumber, PreferredName: input.PreferredName,
			PersonPartyReferenceId: input.PersonPartyReferenceID, ContactReferenceId: input.ContactReferenceID,
			PrivacyNoticeVersion: &input.PrivacyNoticeVersion, NoticeAcknowledgedAt: timestamppb.New(input.NoticeAcknowledgedAt),
			DataRetentionUntil: &input.DataRetentionUntil,
		},
	})
	if err := candidateMutationError(err); err != nil {
		return nil, err
	}

	item := response.GetItem()
	if item.GetCandidateStatus() != "active" {
		return nil, middleware.ErrUnavailable("invalid human resource candidate response", nil)
	}
	return toModelCandidate(item)
}

// UpdateCandidate mengoreksi nama pilihan kandidat melalui field mask.
func (a *Adapter) UpdateCandidate(ctx context.Context, rawID string, input model.UpdateCandidateInput) (*model.Candidate, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, "human_resource.candidates.update")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "candidate")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid candidate ID", nil)
	}
	preferredName, set := input.PreferredName.ValueOK()
	if !set {
		return nil, middleware.ErrInvalidArgument("candidate update is empty", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.recruitment.UpdateCandidate(rpcCtx, &recruitment.UpdateCandidateRequest{
		Id:         id.String(),
		Input:      &recruitment.CandidateInput{PreferredName: preferredName},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"preferred_name"}},
	})
	if err := candidateMutationError(err); err != nil {
		return nil, err
	}
	item := response.GetItem()
	if item.GetId() != id.String() || item.GetCandidateStatus() != "active" {
		return nil, middleware.ErrUnavailable("invalid human resource candidate response", nil)
	}
	return toModelCandidate(item)
}

func candidateMutationError(err error) error {
	switch status.Code(err) {
	case codes.AlreadyExists, codes.FailedPrecondition:
		return middleware.ErrConflict("candidate conflicts with existing state or references", err)
	case codes.OK:
		return nil
	default:
		return mapReadError(err, "candidate")
	}

}

func toModelCandidate(item *recruitment.Candidate) (*model.Candidate, error) {
	id, err := uuid.Parse(item.GetId())
	if err != nil || id == uuid.Nil || item.GetCandidateNumber() == "" || item.GetCandidateStatus() == "" {
		return nil, middleware.ErrUnavailable("invalid human resource candidate response", nil)
	}
	return &model.Candidate{
		ID: nodeID("candidate", id.String()), RawID: id.String(), CandidateNumber: item.GetCandidateNumber(), PreferredName: item.PreferredName,
		CandidateStatus: item.GetCandidateStatus(),
	}, nil
}

const candidateReadPermission = "human_resource.candidates.read"

// Candidate membaca detail kandidat berdasarkan ID.
func (a *Adapter) Candidate(ctx context.Context, rawID string) (*model.Candidate, error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, candidateReadPermission)
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "candidate")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid candidate ID", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.recruitment.GetCandidateByID(rpcCtx, &common.IDRequest{Id: id.String()})
	if err != nil {
		return nil, mapReadError(err, "candidate")
	}
	item := response.GetItem()
	if item.GetId() != id.String() {
		return nil, middleware.ErrUnavailable("invalid human resource candidate response", nil)
	}
	return toModelCandidate(item)
}

// Candidates membaca satu halaman kandidat dengan filter nomor opsional.
func (a *Adapter) Candidates(ctx context.Context, number *string, limit, offset int) (*Page[model.Candidate], error) {
	rpcCtx, err := a.authorizeGlobalEmployee(ctx, candidateReadPermission)
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid candidate pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	var response *recruitment.CandidatesResponse
	if number == nil {
		response, err = a.recruitment.GetAllCandidates(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	} else {
		value := *number
		if strings.TrimSpace(value) == "" || !utf8.ValidString(value) || strings.ContainsRune(value, 0) || utf8.RuneCountInString(value) > 120 {
			return nil, middleware.ErrInvalidArgument("invalid candidate number", nil)
		}
		response, err = a.recruitment.FindCandidateByCandidateNumber(rpcCtx, &common.StringFinderRequest{
			Value: &value, Limit: int32(limit), Offset: int32(offset),
		})
	}
	if err != nil {
		return nil, mapReadError(err, "candidate")
	}
	page, items := response.GetPage(), response.GetItems()
	if !validPage(page, limit, offset, len(items)) {
		return nil, middleware.ErrUnavailable("invalid human resource candidate page", nil)
	}
	result := &Page[model.Candidate]{
		Items: make([]*model.Candidate, 0, len(items)), TotalCount: int(page.GetTotal()), Limit: limit, Offset: offset,
	}
	for _, row := range items {
		item, err := toModelCandidate(row)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
}

// DeleteCandidate meminta penghapusan kandidat tanpa lamaran.
func (a *Adapter) DeleteCandidate(ctx context.Context, rawID string) (*bool, error) {
	rpcCtx, err := a.authorizeSensitiveEmployee(ctx, "human_resource.candidates.delete")
	if err != nil {
		return nil, err
	}
	id, err := parseNodeID(rawID, "candidate")
	if err != nil || id == uuid.Nil {
		return nil, middleware.ErrInvalidArgument("invalid candidate ID", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	response, err := a.recruitment.DeleteCandidate(rpcCtx, &common.IDRequest{Id: id.String()})
	if err := candidateMutationError(err); err != nil {
		return nil, err
	}
	if response == nil {
		return nil, middleware.ErrUnavailable("invalid human resource candidate deletion response", nil)
	}
	deleted := true
	return &deleted, nil
}
