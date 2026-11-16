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

func application(v *entity.Application) *pb.Application {
	return &pb.Application{
		Id: v.ID, CandidateId: v.CandidateID, RequisitionId: v.RequisitionID, ApplicationNumber: v.ApplicationNumber, Status: string(v.Status), AppliedAt: timestamppb.New(v.AppliedAt),
		CurrentStageCode: v.CurrentStageCode, AssignedRecruiterPrincipalId: v.AssignedRecruiterPrincipalID, ScreeningScore: stringPointer(v.ScreeningScore),
		ScorecardVersion: v.ScorecardVersion, DispositionReasonCode: v.DispositionReasonCode, DecisionAt: timestamp(v.DecisionAt), DecisionByPrincipalId: v.DecisionByPrincipalID,
		HiredWorkerId: v.HiredWorkerID, SourcePostingId: v.SourcePostingID, IdempotencyKey: v.IdempotencyKey,
	}
}
func applicationResponse(v *entity.Application, err error) (*pb.ApplicationResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	return &pb.ApplicationResponse{Item: application(v)}, nil
}
func applications(page operation.Page[entity.Application], err error) (*pb.ApplicationsResponse, error) {
	if err != nil {
		return nil, response.Error(err)
	}
	items := make([]*pb.Application, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, application(&page.Items[i]))
	}
	return &pb.ApplicationsResponse{Items: items, Page: pageInfo(page)}, nil
}
func (h *Handler) CreateApplication(ctx context.Context, r *pb.CreateApplicationRequest) (*pb.ApplicationResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	in := r.Input
	decision, err := response.OptionalTimestamp(in.DecisionAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return applicationResponse(h.service.CreateApplication(ctx, service.CreateApplicationInput{
		CandidateID: in.GetCandidateId(), RequisitionID: in.GetRequisitionId(), ApplicationNumber: in.GetApplicationNumber(), Status: enumPointer[entity.ApplicationStatus](in.Status),
		CurrentStageCode: in.CurrentStageCode, AssignedRecruiterPrincipalID: in.AssignedRecruiterPrincipalId, ScreeningScore: decimalPointer(in.ScreeningScore), ScorecardVersion: in.ScorecardVersion,
		DispositionReasonCode: in.DispositionReasonCode, DecisionAt: decision, DecisionByPrincipalID: in.DecisionByPrincipalId, HiredWorkerID: in.HiredWorkerId, SourcePostingID: in.SourcePostingId, IdempotencyKey: in.GetIdempotencyKey(),
	}))
}
func (h *Handler) UpdateApplication(ctx context.Context, r *pb.UpdateApplicationRequest) (*pb.ApplicationResponse, error) {
	if r == nil || r.Input == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	m, err := response.ValidateMask(r.UpdateMask, "candidate_id", "requisition_id", "application_number", "status", "current_stage_code", "assigned_recruiter_principal_id", "screening_score", "scorecard_version", "disposition_reason_code", "decision_at", "decision_by_principal_id", "hired_worker_id", "source_posting_id", "idempotency_key")
	if err != nil {
		return nil, response.Error(err)
	}
	in := r.Input
	decision, err := timeField(m["decision_at"], in.DecisionAt)
	if err != nil {
		return nil, response.Error(err)
	}
	return applicationResponse(h.service.UpdateApplication(ctx, r.Id, service.UpdateApplicationInput{
		CandidateID: field(m["candidate_id"], in.CandidateId), RequisitionID: field(m["requisition_id"], in.RequisitionId), ApplicationNumber: field(m["application_number"], in.ApplicationNumber),
		Status: field(m["status"], enumPointer[entity.ApplicationStatus](in.Status)), CurrentStageCode: field(m["current_stage_code"], in.CurrentStageCode), AssignedRecruiterPrincipalID: field(m["assigned_recruiter_principal_id"], in.AssignedRecruiterPrincipalId),
		ScreeningScore: field(m["screening_score"], decimalPointer(in.ScreeningScore)), ScorecardVersion: field(m["scorecard_version"], in.ScorecardVersion), DispositionReasonCode: field(m["disposition_reason_code"], in.DispositionReasonCode),
		DecisionAt: decision, DecisionByPrincipalID: field(m["decision_by_principal_id"], in.DecisionByPrincipalId), HiredWorkerID: field(m["hired_worker_id"], in.HiredWorkerId), SourcePostingID: field(m["source_posting_id"], in.SourcePostingId), IdempotencyKey: field(m["idempotency_key"], in.IdempotencyKey),
	}))
}
func (h *Handler) DeleteApplication(ctx context.Context, r *common.IDRequest) (*common.EmptyResponse, error) {
	return deleteRecord(ctx, r, h.service.DeleteApplication)
}
func (h *Handler) StartApplicationScreening(ctx context.Context, r *common.IDRequest) (*pb.ApplicationResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationResponse(h.service.StartApplicationScreening(ctx, r.Id))
}
func (h *Handler) GetApplicationByID(ctx context.Context, r *common.IDRequest) (*pb.ApplicationResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applicationResponse(h.service.GetApplicationByID(ctx, r.Id))
}
func (h *Handler) GetAllApplications(ctx context.Context, r *common.PageRequest) (*pb.ApplicationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.GetAllApplications(ctx, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationByCandidateID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationByCandidateID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationByRequisitionID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationByRequisitionID(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationByApplicationNumber(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationByApplicationNumber(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationByStatus(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationByStatus(ctx, entity.ApplicationStatus(*r.Value), int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationByCurrentStageCode(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationByCurrentStageCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationByAssignedRecruiterPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationByAssignedRecruiterPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationByDispositionReasonCode(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationByDispositionReasonCode(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationByDecisionByPrincipalID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationByDecisionByPrincipalID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationByHiredWorkerID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationByHiredWorkerID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationBySourcePostingID(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationBySourcePostingID(ctx, r.Value, int(r.Limit), int(r.Offset)))
}
func (h *Handler) FindApplicationByIdempotencyKey(ctx context.Context, r *common.StringFinderRequest) (*pb.ApplicationsResponse, error) {
	if r == nil || r.Value == nil {
		return nil, response.Error(hrerrors.ErrInvalidInput)
	}
	return applications(h.service.FindApplicationByIdempotencyKey(ctx, *r.Value, int(r.Limit), int(r.Offset)))
}
