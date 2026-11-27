package human_resource

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	recruitment "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelApplication(row *recruitment.Application) (*model.Application, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource application response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource application ID", nil)
	}
	out := &model.Application{ID: nodeID("application", row.GetId()), RawID: row.GetId()}
	out.CandidateID = row.GetCandidateId()
	out.RequisitionID = row.GetRequisitionId()
	out.ApplicationNumber = row.GetApplicationNumber()
	out.Status = row.GetStatus()
	out.AppliedAt = pbTimePtr(row.GetAppliedAt())
	out.CurrentStageCode = row.CurrentStageCode
	out.AssignedRecruiterPrincipalID = row.AssignedRecruiterPrincipalId
	out.ScreeningScore = row.ScreeningScore
	out.ScorecardVersion = row.ScorecardVersion
	out.DispositionReasonCode = row.DispositionReasonCode
	out.DecisionAt = pbTimePtr(row.GetDecisionAt())
	out.DecisionByPrincipalID = row.DecisionByPrincipalId
	out.HiredWorkerID = row.HiredWorkerId
	out.SourcePostingID = row.SourcePostingId
	out.IdempotencyKey = row.GetIdempotencyKey()
	return out, nil
}

// Application membaca detail application berdasarkan ID.
func (a *Adapter) Application(ctx context.Context, rawID string) (*model.Application, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_applications", "human_resource.applications.read")
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
	resp, err := a.recruitment.GetApplicationByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "application")
	}
	return toModelApplication(resp.GetItem())
}

// Applications mendaftar dan memfilter application dengan paginasi.
func (a *Adapter) Applications(ctx context.Context, limit, offset int, f_candidate_id *string, f_requisition_id *string, f_application_number *string, f_status *string, f_current_stage_code *string, f_assigned_recruiter_principal_id *string, f_disposition_reason_code *string, f_decision_by_principal_id *string, f_hired_worker_id *string, f_source_posting_id *string, f_idempotency_key *string) (*Page[model.Application], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_applications", "human_resource.applications.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid application pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_candidate_id != nil, f_requisition_id != nil, f_application_number != nil, f_status != nil, f_current_stage_code != nil, f_assigned_recruiter_principal_id != nil, f_disposition_reason_code != nil, f_decision_by_principal_id != nil, f_hired_worker_id != nil, f_source_posting_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *recruitment.ApplicationsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_candidate_id, "candidate_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationByCandidateID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_requisition_id, "requisition_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationByRequisitionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_application_number, "application_number")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationByApplicationNumber(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_current_stage_code, "current_stage_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationByCurrentStageCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_assigned_recruiter_principal_id, "assigned_recruiter_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationByAssignedRecruiterPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_disposition_reason_code, "disposition_reason_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationByDispositionReasonCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_decision_by_principal_id, "decision_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationByDecisionByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(f_hired_worker_id, "hired_worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationByHiredWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 9:
		v, fErr := validateFilterString(f_source_posting_id, "source_posting_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationBySourcePostingID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 10:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.recruitment.GetAllApplications(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "application")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "application", toModelApplication)
}

// CreateApplication membuat application baru melalui HRM.
func (a *Adapter) CreateApplication(ctx context.Context, input model.CreateApplicationInput) (*model.Application, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_applications", "create", "human_resource.applications.create")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.ApplicationInput{}
	patch.CandidateId = input.CandidateID
	patch.RequisitionId = input.RequisitionID
	patch.ApplicationNumber = input.ApplicationNumber
	patch.Status = input.Status
	patch.CurrentStageCode = input.CurrentStageCode
	patch.AssignedRecruiterPrincipalId = input.AssignedRecruiterPrincipalID
	patch.ScreeningScore = input.ScreeningScore
	patch.ScorecardVersion = input.ScorecardVersion
	patch.DispositionReasonCode = input.DispositionReasonCode
	patch.DecisionAt = optionalTimestamp(input.DecisionAt)
	patch.DecisionByPrincipalId = input.DecisionByPrincipalID
	patch.HiredWorkerId = input.HiredWorkerID
	patch.SourcePostingId = input.SourcePostingID
	patch.IdempotencyKey = input.IdempotencyKey
	resp, err := a.recruitment.CreateApplication(rpcCtx, &recruitment.CreateApplicationRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "application")
	}
	return toModelApplication(resp.GetItem())
}

// UpdateApplication memperbarui application berdasarkan FieldMask.
func (a *Adapter) UpdateApplication(ctx context.Context, rawID string, input model.UpdateApplicationInput) (*model.Application, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_applications", "update", "human_resource.applications.update")
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
	patch := &recruitment.ApplicationInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "candidate_id", input.CandidateID, &patch.CandidateId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "requisition_id", input.RequisitionID, &patch.RequisitionId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "application_number", input.ApplicationNumber, &patch.ApplicationNumber, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "current_stage_code", input.CurrentStageCode, &patch.CurrentStageCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assigned_recruiter_principal_id", input.AssignedRecruiterPrincipalID, &patch.AssignedRecruiterPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "screening_score", input.ScreeningScore, &patch.ScreeningScore, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "scorecard_version", input.ScorecardVersion, &patch.ScorecardVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "disposition_reason_code", input.DispositionReasonCode, &patch.DispositionReasonCode, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "decision_at", input.DecisionAt, &patch.DecisionAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "decision_by_principal_id", input.DecisionByPrincipalID, &patch.DecisionByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "hired_worker_id", input.HiredWorkerID, &patch.HiredWorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "source_posting_id", input.SourcePostingID, &patch.SourcePostingId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("application update is empty", nil)
	}
	resp, err := a.recruitment.UpdateApplication(rpcCtx, &recruitment.UpdateApplicationRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "application")
	}
	return toModelApplication(resp.GetItem())
}

// DeleteApplication menghapus application berdasarkan ID.
func (a *Adapter) DeleteApplication(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_applications", "human_resource.applications.delete", "application", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.recruitment.DeleteApplication(c, req)
	})
}

func toModelApplicationStageHistory(row *recruitment.ApplicationStageHistory) (*model.ApplicationStageHistory, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource application-stage-history response", nil)
	}
	if row.GetId() <= 0 {
		return nil, middleware.ErrUnavailable("invalid human resource application-stage-history ID", nil)
	}
	rawID := strconv.FormatInt(row.GetId(), 10)
	out := &model.ApplicationStageHistory{ID: int64NodeID("application-stage-history", row.GetId()), RawID: rawID}
	out.ApplicationID = row.GetApplicationId()
	out.PreviousStatus = row.PreviousStatus
	out.NewStatus = row.GetNewStatus()
	out.StageCode = row.StageCode
	out.ChangedAt = pbTimePtr(row.GetChangedAt())
	out.ChangedByPrincipalID = row.ChangedByPrincipalId
	out.ReasonCode = row.ReasonCode
	return out, nil
}

// ApplicationStageHistory membaca detail application-stage-history berdasarkan ID.
func (a *Adapter) ApplicationStageHistory(ctx context.Context, rawID string) (*model.ApplicationStageHistory, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_application_stage_history", "human_resource.application_stage_history.read")
	if err != nil {
		return nil, err
	}
	id, err := parsePositiveInt64NodeID(rawID, "application-stage-history")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.GetApplicationStageHistoryByID(rpcCtx, &common.IdentityIDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "application-stage-history")
	}
	return toModelApplicationStageHistory(resp.GetItem())
}

// ApplicationStageHistories mendaftar dan memfilter application-stage-history dengan paginasi.
func (a *Adapter) ApplicationStageHistories(ctx context.Context, limit, offset int, f_application_id *string, f_previous_status *string, f_new_status *string, f_stage_code *string, f_changed_by_principal_id *string, f_reason_code *string) (*Page[model.ApplicationStageHistory], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_application_stage_history", "human_resource.application_stage_history.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid application-stage-history pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_application_id != nil, f_previous_status != nil, f_new_status != nil, f_stage_code != nil, f_changed_by_principal_id != nil, f_reason_code != nil)
	if err != nil {
		return nil, err
	}
	var resp *recruitment.ApplicationStageHistoriesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_application_id, "application_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationStageHistoryByApplicationID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_previous_status, "previous_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationStageHistoryByPreviousStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_new_status, "new_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationStageHistoryByNewStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_stage_code, "stage_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationStageHistoryByStageCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_changed_by_principal_id, "changed_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationStageHistoryByChangedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_reason_code, "reason_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationStageHistoryByReasonCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.recruitment.GetAllApplicationStageHistories(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "application-stage-history")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "application-stage-history", toModelApplicationStageHistory)
}

func toModelInterview(row *recruitment.Interview) (*model.Interview, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource interview response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource interview ID", nil)
	}
	out := &model.Interview{ID: nodeID("interview", row.GetId()), RawID: row.GetId()}
	out.ApplicationID = row.GetApplicationId()
	out.InterviewRound = int(row.GetInterviewRound())
	out.InterviewType = row.GetInterviewType()
	out.ScheduledStart = pbTimePtr(row.GetScheduledStart())
	out.ScheduledEnd = pbTimePtr(row.GetScheduledEnd())
	out.LocationReferenceID = row.LocationReferenceId
	out.MeetingReferenceID = row.MeetingReferenceId
	out.Status = row.GetStatus()
	out.CandidateTimezone = row.CandidateTimezone
	out.CreatedByPrincipalID = row.CreatedByPrincipalId
	return out, nil
}

// Interview membaca detail interview berdasarkan ID.
func (a *Adapter) Interview(ctx context.Context, rawID string) (*model.Interview, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_interviews", "human_resource.interviews.read")
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
	resp, err := a.recruitment.GetInterviewByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "interview")
	}
	return toModelInterview(resp.GetItem())
}

// Interviews mendaftar dan memfilter interview dengan paginasi.
func (a *Adapter) Interviews(ctx context.Context, limit, offset int, f_application_id *string, f_interview_type *string, f_location_reference_id *string, f_meeting_reference_id *string, f_status *string, f_created_by_principal_id *string) (*Page[model.Interview], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_interviews", "human_resource.interviews.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid interview pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_application_id != nil, f_interview_type != nil, f_location_reference_id != nil, f_meeting_reference_id != nil, f_status != nil, f_created_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *recruitment.InterviewsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_application_id, "application_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewByApplicationID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_interview_type, "interview_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewByInterviewType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_location_reference_id, "location_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewByLocationReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_meeting_reference_id, "meeting_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewByMeetingReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_created_by_principal_id, "created_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewByCreatedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.recruitment.GetAllInterviews(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "interview")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "interview", toModelInterview)
}

// CreateInterview membuat interview baru melalui HRM.
func (a *Adapter) CreateInterview(ctx context.Context, input model.CreateInterviewInput) (*model.Interview, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_interviews", "create", "human_resource.interviews.create")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.InterviewInput{}
	patch.ApplicationId = input.ApplicationID
	patch.InterviewRound = optionalInt32Ptr(input.InterviewRound)
	patch.InterviewType = input.InterviewType
	patch.ScheduledStart = optionalTimestamp(input.ScheduledStart)
	patch.ScheduledEnd = optionalTimestamp(input.ScheduledEnd)
	patch.LocationReferenceId = input.LocationReferenceID
	patch.MeetingReferenceId = input.MeetingReferenceID
	patch.Status = input.Status
	patch.CandidateTimezone = input.CandidateTimezone
	resp, err := a.recruitment.CreateInterview(rpcCtx, &recruitment.CreateInterviewRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "interview")
	}
	return toModelInterview(resp.GetItem())
}

// UpdateInterview memperbarui interview berdasarkan FieldMask.
func (a *Adapter) UpdateInterview(ctx context.Context, rawID string, input model.UpdateInterviewInput) (*model.Interview, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_interviews", "update", "human_resource.interviews.update")
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
	patch := &recruitment.InterviewInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "application_id", input.ApplicationID, &patch.ApplicationId, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "interview_round", input.InterviewRound, &patch.InterviewRound, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "interview_type", input.InterviewType, &patch.InterviewType, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "scheduled_start", input.ScheduledStart, &patch.ScheduledStart, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "scheduled_end", input.ScheduledEnd, &patch.ScheduledEnd, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "location_reference_id", input.LocationReferenceID, &patch.LocationReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "meeting_reference_id", input.MeetingReferenceID, &patch.MeetingReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "candidate_timezone", input.CandidateTimezone, &patch.CandidateTimezone, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("interview update is empty", nil)
	}
	resp, err := a.recruitment.UpdateInterview(rpcCtx, &recruitment.UpdateInterviewRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "interview")
	}
	return toModelInterview(resp.GetItem())
}

// DeleteInterview menghapus interview berdasarkan ID.
func (a *Adapter) DeleteInterview(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_interviews", "human_resource.interviews.delete", "interview", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.recruitment.DeleteInterview(c, req)
	})
}

func toModelInterviewPanelist(row *recruitment.InterviewPanelist) (*model.InterviewPanelist, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource interview-panelist response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource interview-panelist ID", nil)
	}
	out := &model.InterviewPanelist{ID: nodeID("interview-panelist", row.GetId()), RawID: row.GetId()}
	out.InterviewID = row.GetInterviewId()
	out.InterviewerPrincipalID = row.GetInterviewerPrincipalId()
	out.InterviewerRole = row.InterviewerRole
	out.ScorecardTemplateReferenceID = row.ScorecardTemplateReferenceId
	out.FeedbackDueAt = pbTimePtr(row.GetFeedbackDueAt())
	return out, nil
}

// InterviewPanelist membaca detail interview-panelist berdasarkan ID.
func (a *Adapter) InterviewPanelist(ctx context.Context, rawID string) (*model.InterviewPanelist, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_interview_panelists", "human_resource.interview_panelists.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "interview-panelist", "interview-panelist")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.GetInterviewPanelistByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "interview-panelist")
	}
	return toModelInterviewPanelist(resp.GetItem())
}

// InterviewPanelists mendaftar dan memfilter interview-panelist dengan paginasi.
func (a *Adapter) InterviewPanelists(ctx context.Context, limit, offset int, f_interview_id *string, f_interviewer_principal_id *string, f_scorecard_template_reference_id *string) (*Page[model.InterviewPanelist], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_interview_panelists", "human_resource.interview_panelists.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid interview-panelist pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_interview_id != nil, f_interviewer_principal_id != nil, f_scorecard_template_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *recruitment.InterviewPanelistsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_interview_id, "interview_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewPanelistByInterviewID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_interviewer_principal_id, "interviewer_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewPanelistByInterviewerPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_scorecard_template_reference_id, "scorecard_template_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewPanelistByScorecardTemplateReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.recruitment.GetAllInterviewPanelists(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "interview-panelist")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "interview-panelist", toModelInterviewPanelist)
}

// CreateInterviewPanelist membuat interview-panelist baru melalui HRM.
func (a *Adapter) CreateInterviewPanelist(ctx context.Context, input model.CreateInterviewPanelistInput) (*model.InterviewPanelist, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_interview_panelists", "create", "human_resource.interview_panelists.create")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.InterviewPanelistInput{}
	patch.InterviewId = input.InterviewID
	patch.InterviewerPrincipalId = input.InterviewerPrincipalID
	patch.InterviewerRole = input.InterviewerRole
	patch.ScorecardTemplateReferenceId = input.ScorecardTemplateReferenceID
	patch.FeedbackDueAt = optionalTimestamp(input.FeedbackDueAt)
	resp, err := a.recruitment.CreateInterviewPanelist(rpcCtx, &recruitment.CreateInterviewPanelistRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "interview-panelist")
	}
	return toModelInterviewPanelist(resp.GetItem())
}

// UpdateInterviewPanelist memperbarui interview-panelist berdasarkan FieldMask.
func (a *Adapter) UpdateInterviewPanelist(ctx context.Context, rawID string, input model.UpdateInterviewPanelistInput) (*model.InterviewPanelist, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_interview_panelists", "update", "human_resource.interview_panelists.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "interview-panelist", "interview-panelist")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.InterviewPanelistInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "interview_id", input.InterviewID, &patch.InterviewId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "interviewer_principal_id", input.InterviewerPrincipalID, &patch.InterviewerPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "interviewer_role", input.InterviewerRole, &patch.InterviewerRole, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "scorecard_template_reference_id", input.ScorecardTemplateReferenceID, &patch.ScorecardTemplateReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "feedback_due_at", input.FeedbackDueAt, &patch.FeedbackDueAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("interview-panelist update is empty", nil)
	}
	resp, err := a.recruitment.UpdateInterviewPanelist(rpcCtx, &recruitment.UpdateInterviewPanelistRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "interview-panelist")
	}
	return toModelInterviewPanelist(resp.GetItem())
}

// DeleteInterviewPanelist menghapus interview-panelist berdasarkan ID.
func (a *Adapter) DeleteInterviewPanelist(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_interview_panelists", "human_resource.interview_panelists.delete", "interview-panelist", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.recruitment.DeleteInterviewPanelist(c, req)
	})
}

func toModelInterviewFeedback(row *recruitment.InterviewFeedback) (*model.InterviewFeedback, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource interview-feedback response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource interview-feedback ID", nil)
	}
	out := &model.InterviewFeedback{ID: nodeID("interview-feedback", row.GetId()), RawID: row.GetId()}
	out.InterviewID = row.GetInterviewId()
	out.InterviewerPrincipalID = row.GetInterviewerPrincipalId()
	out.FeedbackVersion = int(row.GetFeedbackVersion())
	out.SubmittedAt = pbTimePtr(row.GetSubmittedAt())
	out.OverallRecommendation = row.GetOverallRecommendation()
	if doc := row.GetCompetencyScores(); doc != nil {
		out.CompetencyScores = pbJSONString(doc.GetUtf8())
	}
	out.EvidenceNotes = row.EvidenceNotes
	out.ContentDigest = pbBase64StringValue(row.GetContentDigest())
	out.VisibilityClass = row.GetVisibilityClass()
	return out, nil
}

// InterviewFeedback membaca detail interview-feedback berdasarkan ID.
func (a *Adapter) InterviewFeedback(ctx context.Context, rawID string) (*model.InterviewFeedback, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_interview_feedback", "human_resource.interview_feedback.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "interview-feedback", "interview-feedback")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.GetInterviewFeedbackByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "interview-feedback")
	}
	return toModelInterviewFeedback(resp.GetItem())
}

// InterviewFeedbacks mendaftar dan memfilter interview-feedback dengan paginasi.
func (a *Adapter) InterviewFeedbacks(ctx context.Context, limit, offset int, f_interview_id *string, f_interviewer_principal_id *string) (*Page[model.InterviewFeedback], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_interview_feedback", "human_resource.interview_feedback.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid interview-feedback pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_interview_id != nil, f_interviewer_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *recruitment.InterviewFeedbacksResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_interview_id, "interview_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewFeedbackByInterviewID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_interviewer_principal_id, "interviewer_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindInterviewFeedbackByInterviewerPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.recruitment.GetAllInterviewFeedbacks(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "interview-feedback")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "interview-feedback", toModelInterviewFeedback)
}

// CreateInterviewFeedback membuat interview-feedback baru melalui HRM.
func (a *Adapter) CreateInterviewFeedback(ctx context.Context, input model.CreateInterviewFeedbackInput) (*model.InterviewFeedback, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_interview_feedback", "create", "human_resource.interview_feedback.create")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.InterviewFeedbackInput{}
	patch.InterviewId = input.InterviewID
	patch.InterviewerPrincipalId = input.InterviewerPrincipalID
	patch.FeedbackVersion = optionalInt32Ptr(input.FeedbackVersion)
	patch.OverallRecommendation = input.OverallRecommendation
	if b, jErr := optionalJSONBytes(input.CompetencyScores, "competency_scores"); jErr != nil {
		return nil, jErr
	} else if b != nil {
		patch.CompetencyScores = &common.JsonDocument{Utf8: b}
	}
	patch.EvidenceNotes = input.EvidenceNotes
	if b, bErr := optionalBase64Bytes(input.ContentDigest, "content_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.ContentDigest = b
	}
	patch.VisibilityClass = input.VisibilityClass
	resp, err := a.recruitment.CreateInterviewFeedback(rpcCtx, &recruitment.CreateInterviewFeedbackRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "interview-feedback")
	}
	return toModelInterviewFeedback(resp.GetItem())
}

// UpdateInterviewFeedback memperbarui interview-feedback berdasarkan FieldMask.
func (a *Adapter) UpdateInterviewFeedback(ctx context.Context, rawID string, input model.UpdateInterviewFeedbackInput) (*model.InterviewFeedback, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_interview_feedback", "update", "human_resource.interview_feedback.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "interview-feedback", "interview-feedback")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.InterviewFeedbackInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "interview_id", input.InterviewID, &patch.InterviewId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "interviewer_principal_id", input.InterviewerPrincipalID, &patch.InterviewerPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "feedback_version", input.FeedbackVersion, &patch.FeedbackVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "overall_recommendation", input.OverallRecommendation, &patch.OverallRecommendation, true); err != nil {
		return nil, err
	}
	if val, set := input.CompetencyScores.ValueOK(); set {
		if val == nil {
			patch.CompetencyScores = nil
		} else if b, jErr := requiredJSONBytes(*val, "competency_scores"); jErr != nil {
			return nil, jErr
		} else {
			patch.CompetencyScores = &common.JsonDocument{Utf8: b}
		}
		mask.Paths = append(mask.Paths, "competency_scores")
	}
	if err := applyStringMask(mask, "evidence_notes", input.EvidenceNotes, &patch.EvidenceNotes, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "content_digest", input.ContentDigest, &patch.ContentDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "visibility_class", input.VisibilityClass, &patch.VisibilityClass, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("interview-feedback update is empty", nil)
	}
	resp, err := a.recruitment.UpdateInterviewFeedback(rpcCtx, &recruitment.UpdateInterviewFeedbackRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "interview-feedback")
	}
	return toModelInterviewFeedback(resp.GetItem())
}

// DeleteInterviewFeedback menghapus interview-feedback berdasarkan ID.
func (a *Adapter) DeleteInterviewFeedback(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_interview_feedback", "human_resource.interview_feedback.delete", "interview-feedback", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.recruitment.DeleteInterviewFeedback(c, req)
	})
}

func toModelOffer(row *recruitment.Offer) (*model.Offer, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource offer response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource offer ID", nil)
	}
	out := &model.Offer{ID: nodeID("offer", row.GetId()), RawID: row.GetId()}
	out.ApplicationID = row.GetApplicationId()
	out.OfferNumber = row.GetOfferNumber()
	out.OfferVersion = int(row.GetOfferVersion())
	out.OfferStatus = row.GetOfferStatus()
	out.ProposedJobID = row.GetProposedJobId()
	out.ProposedLegalEntityReferenceID = row.GetProposedLegalEntityReferenceId()
	out.ProposedStartDate = row.ProposedStartDate
	out.ProposedContractType = row.GetProposedContractType()
	out.ProposedCompensationDocumentVersionReferenceID = row.ProposedCompensationDocumentVersionReferenceId
	out.CurrencyCode = row.CurrencyCode
	out.OfferDocumentReferenceID = row.OfferDocumentReferenceId
	out.OfferDocumentVersionReferenceID = row.OfferDocumentVersionReferenceId
	out.OfferContentDigest = pbBase64StringValue(row.GetOfferContentDigest())
	out.IssuedAt = pbTimePtr(row.GetIssuedAt())
	out.ExpiresAt = pbTimePtr(row.GetExpiresAt())
	out.CandidateResponse = row.CandidateResponse
	out.RespondedAt = pbTimePtr(row.GetRespondedAt())
	out.ApprovedByPrincipalID = row.ApprovedByPrincipalId
	out.ApprovedAt = pbTimePtr(row.GetApprovedAt())
	out.CreatedByPrincipalID = row.CreatedByPrincipalId
	out.CreatedAt = pbTimePtr(row.GetCreatedAt())
	out.WithdrawnAt = pbTimePtr(row.GetWithdrawnAt())
	out.WithdrawnByPrincipalID = row.WithdrawnByPrincipalId
	out.WithdrawalReasonCode = row.WithdrawalReasonCode
	return out, nil
}

// Offer membaca detail offer berdasarkan ID.
func (a *Adapter) Offer(ctx context.Context, rawID string) (*model.Offer, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_offers", "human_resource.offers.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "offer", "offer")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.GetOfferByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "offer")
	}
	return toModelOffer(resp.GetItem())
}

// Offers mendaftar dan memfilter offer dengan paginasi.
func (a *Adapter) Offers(ctx context.Context, limit, offset int, f_application_id *string, f_offer_number *string, f_offer_status *string, f_proposed_job_id *string, f_proposed_legal_entity_reference_id *string, f_proposed_contract_type *string, f_proposed_compensation_document_version_reference_id *string, f_currency_code *string, f_offer_document_reference_id *string, f_offer_document_version_reference_id *string, f_approved_by_principal_id *string) (*Page[model.Offer], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_offers", "human_resource.offers.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid offer pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_application_id != nil, f_offer_number != nil, f_offer_status != nil, f_proposed_job_id != nil, f_proposed_legal_entity_reference_id != nil, f_proposed_contract_type != nil, f_proposed_compensation_document_version_reference_id != nil, f_currency_code != nil, f_offer_document_reference_id != nil, f_offer_document_version_reference_id != nil, f_approved_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *recruitment.OffersResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_application_id, "application_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByApplicationID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_offer_number, "offer_number")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByOfferNumber(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_offer_status, "offer_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByOfferStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_proposed_job_id, "proposed_job_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByProposedJobID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_proposed_legal_entity_reference_id, "proposed_legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByProposedLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_proposed_contract_type, "proposed_contract_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByProposedContractType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_proposed_compensation_document_version_reference_id, "proposed_compensation_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByProposedCompensationDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_currency_code, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(f_offer_document_reference_id, "offer_document_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByOfferDocumentReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 9:
		v, fErr := validateFilterString(f_offer_document_version_reference_id, "offer_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByOfferDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 10:
		v, fErr := validateFilterString(f_approved_by_principal_id, "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindOfferByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.recruitment.GetAllOffers(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "offer")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "offer", toModelOffer)
}

// CreateOffer membuat offer baru melalui HRM.
func (a *Adapter) CreateOffer(ctx context.Context, input model.CreateOfferInput) (*model.Offer, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_offers", "create", "human_resource.offers.create")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.OfferInput{}
	patch.ApplicationId = input.ApplicationID
	patch.OfferNumber = input.OfferNumber
	patch.OfferVersion = optionalInt32Ptr(input.OfferVersion)
	patch.OfferStatus = input.OfferStatus
	patch.ProposedJobId = input.ProposedJobID
	patch.ProposedLegalEntityReferenceId = input.ProposedLegalEntityReferenceID
	patch.ProposedStartDate = input.ProposedStartDate
	patch.ProposedContractType = input.ProposedContractType
	patch.ProposedCompensationDocumentVersionReferenceId = input.ProposedCompensationDocumentVersionReferenceID
	patch.CurrencyCode = input.CurrencyCode
	patch.OfferDocumentReferenceId = input.OfferDocumentReferenceID
	patch.OfferDocumentVersionReferenceId = input.OfferDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.OfferContentDigest, "offer_content_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.OfferContentDigest = b
	}
	patch.IssuedAt = optionalTimestamp(input.IssuedAt)
	patch.ExpiresAt = optionalTimestamp(input.ExpiresAt)
	patch.CandidateResponse = input.CandidateResponse
	patch.RespondedAt = optionalTimestamp(input.RespondedAt)
	patch.ApprovedByPrincipalId = input.ApprovedByPrincipalID
	patch.ApprovedAt = optionalTimestamp(input.ApprovedAt)
	resp, err := a.recruitment.CreateOffer(rpcCtx, &recruitment.CreateOfferRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "offer")
	}
	return toModelOffer(resp.GetItem())
}

// UpdateOffer memperbarui offer berdasarkan FieldMask.
func (a *Adapter) UpdateOffer(ctx context.Context, rawID string, input model.UpdateOfferInput) (*model.Offer, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_offers", "update", "human_resource.offers.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "offer", "offer")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.OfferInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "application_id", input.ApplicationID, &patch.ApplicationId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "offer_number", input.OfferNumber, &patch.OfferNumber, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "offer_version", input.OfferVersion, &patch.OfferVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "offer_status", input.OfferStatus, &patch.OfferStatus, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "proposed_job_id", input.ProposedJobID, &patch.ProposedJobId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "proposed_legal_entity_reference_id", input.ProposedLegalEntityReferenceID, &patch.ProposedLegalEntityReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "proposed_start_date", input.ProposedStartDate, &patch.ProposedStartDate, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "proposed_contract_type", input.ProposedContractType, &patch.ProposedContractType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "proposed_compensation_document_version_reference_id", input.ProposedCompensationDocumentVersionReferenceID, &patch.ProposedCompensationDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "currency_code", input.CurrencyCode, &patch.CurrencyCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "offer_document_reference_id", input.OfferDocumentReferenceID, &patch.OfferDocumentReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "offer_document_version_reference_id", input.OfferDocumentVersionReferenceID, &patch.OfferDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "offer_content_digest", input.OfferContentDigest, &patch.OfferContentDigest, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "issued_at", input.IssuedAt, &patch.IssuedAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "expires_at", input.ExpiresAt, &patch.ExpiresAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "candidate_response", input.CandidateResponse, &patch.CandidateResponse, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "responded_at", input.RespondedAt, &patch.RespondedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("offer update is empty", nil)
	}
	resp, err := a.recruitment.UpdateOffer(rpcCtx, &recruitment.UpdateOfferRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "offer")
	}
	return toModelOffer(resp.GetItem())
}

// DeleteOffer menghapus offer berdasarkan ID.
func (a *Adapter) DeleteOffer(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_offers", "human_resource.offers.delete", "offer", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.recruitment.DeleteOffer(c, req)
	})
}

func toModelJobPosting(row *recruitment.JobPosting) (*model.JobPosting, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource job-posting response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource job-posting ID", nil)
	}
	out := &model.JobPosting{ID: nodeID("job-posting", row.GetId()), RawID: row.GetId()}
	out.RequisitionID = row.GetRequisitionId()
	out.ChannelCode = row.GetChannelCode()
	out.PostingVersion = int(row.GetPostingVersion())
	out.ExternalPostingReference = row.ExternalPostingReference
	out.ContentDocumentVersionReferenceID = row.GetContentDocumentVersionReferenceId()
	out.ContentDigest = pbBase64StringValue(row.GetContentDigest())
	out.Status = row.GetStatus()
	out.PublishedAt = pbTimePtr(row.GetPublishedAt())
	out.ClosesAt = pbTimePtr(row.GetClosesAt())
	out.CreatedByPrincipalID = row.GetCreatedByPrincipalId()
	out.WithdrawnAt = pbTimePtr(row.GetWithdrawnAt())
	out.WithdrawalReasonCode = row.WithdrawalReasonCode
	out.WithdrawnByPrincipalID = row.WithdrawnByPrincipalId
	return out, nil
}

// JobPosting membaca detail job-posting berdasarkan ID.
func (a *Adapter) JobPosting(ctx context.Context, rawID string) (*model.JobPosting, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_job_postings", "human_resource.job_postings.read")
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
	resp, err := a.recruitment.GetJobPostingByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "job-posting")
	}
	return toModelJobPosting(resp.GetItem())
}

// JobPostings mendaftar dan memfilter job-posting dengan paginasi.
func (a *Adapter) JobPostings(ctx context.Context, limit, offset int, f_requisition_id *string, f_channel_code *string, f_content_document_version_reference_id *string, f_status *string, f_created_by_principal_id *string, f_withdrawal_reason_code *string, f_withdrawn_by_principal_id *string) (*Page[model.JobPosting], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_job_postings", "human_resource.job_postings.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid job-posting pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_requisition_id != nil, f_channel_code != nil, f_content_document_version_reference_id != nil, f_status != nil, f_created_by_principal_id != nil, f_withdrawal_reason_code != nil, f_withdrawn_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *recruitment.JobPostingsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_requisition_id, "requisition_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindJobPostingByRequisitionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_channel_code, "channel_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindJobPostingByChannelCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_content_document_version_reference_id, "content_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindJobPostingByContentDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindJobPostingByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_created_by_principal_id, "created_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindJobPostingByCreatedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_withdrawal_reason_code, "withdrawal_reason_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindJobPostingByWithdrawalReasonCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_withdrawn_by_principal_id, "withdrawn_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindJobPostingByWithdrawnByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.recruitment.GetAllJobPostings(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "job-posting")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "job-posting", toModelJobPosting)
}

// CreateJobPosting membuat job-posting baru melalui HRM.
func (a *Adapter) CreateJobPosting(ctx context.Context, input model.CreateJobPostingInput) (*model.JobPosting, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_job_postings", "create", "human_resource.job_postings.create")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.JobPostingInput{}
	patch.RequisitionId = input.RequisitionID
	patch.ChannelCode = input.ChannelCode
	patch.PostingVersion = optionalInt32Ptr(input.PostingVersion)
	patch.ExternalPostingReference = input.ExternalPostingReference
	patch.ContentDocumentVersionReferenceId = input.ContentDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.ContentDigest, "content_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.ContentDigest = b
	}
	patch.Status = input.Status
	patch.PublishedAt = optionalTimestamp(input.PublishedAt)
	patch.ClosesAt = optionalTimestamp(input.ClosesAt)
	resp, err := a.recruitment.CreateJobPosting(rpcCtx, &recruitment.CreateJobPostingRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "job-posting")
	}
	return toModelJobPosting(resp.GetItem())
}

// UpdateJobPosting memperbarui job-posting berdasarkan FieldMask.
func (a *Adapter) UpdateJobPosting(ctx context.Context, rawID string, input model.UpdateJobPostingInput) (*model.JobPosting, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_job_postings", "update", "human_resource.job_postings.update")
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
	patch := &recruitment.JobPostingInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "requisition_id", input.RequisitionID, &patch.RequisitionId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "channel_code", input.ChannelCode, &patch.ChannelCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "posting_version", input.PostingVersion, &patch.PostingVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "external_posting_reference", input.ExternalPostingReference, &patch.ExternalPostingReference, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "content_document_version_reference_id", input.ContentDocumentVersionReferenceID, &patch.ContentDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "content_digest", input.ContentDigest, &patch.ContentDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "published_at", input.PublishedAt, &patch.PublishedAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "closes_at", input.ClosesAt, &patch.ClosesAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("job-posting update is empty", nil)
	}
	resp, err := a.recruitment.UpdateJobPosting(rpcCtx, &recruitment.UpdateJobPostingRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "job-posting")
	}
	return toModelJobPosting(resp.GetItem())
}

// DeleteJobPosting menghapus job-posting berdasarkan ID.
func (a *Adapter) DeleteJobPosting(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_job_postings", "human_resource.job_postings.delete", "job-posting", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.recruitment.DeleteJobPosting(c, req)
	})
}

func toModelApplicationAssessment(row *recruitment.ApplicationAssessment) (*model.ApplicationAssessment, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource application-assessment response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource application-assessment ID", nil)
	}
	out := &model.ApplicationAssessment{ID: nodeID("application-assessment", row.GetId()), RawID: row.GetId()}
	out.ApplicationID = row.GetApplicationId()
	out.AssessmentCode = row.GetAssessmentCode()
	out.AttemptNumber = int(row.GetAttemptNumber())
	out.RubricDocumentVersionReferenceID = row.GetRubricDocumentVersionReferenceId()
	out.RubricDigest = pbBase64StringValue(row.GetRubricDigest())
	out.Status = row.GetStatus()
	out.AssessorPrincipalID = row.AssessorPrincipalId
	out.ScheduledAt = pbTimePtr(row.GetScheduledAt())
	out.CompletedAt = pbTimePtr(row.GetCompletedAt())
	out.NormalizedScore = row.NormalizedScore
	out.RecommendationCode = row.RecommendationCode
	out.EvidenceDocumentVersionReferenceID = row.EvidenceDocumentVersionReferenceId
	out.EvidenceDigest = pbBase64StringValue(row.GetEvidenceDigest())
	return out, nil
}

// ApplicationAssessment membaca detail application-assessment berdasarkan ID.
func (a *Adapter) ApplicationAssessment(ctx context.Context, rawID string) (*model.ApplicationAssessment, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_application_assessments", "human_resource.application_assessments.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "application-assessment", "application-assessment")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.recruitment.GetApplicationAssessmentByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "application-assessment")
	}
	return toModelApplicationAssessment(resp.GetItem())
}

// ApplicationAssessments mendaftar dan memfilter application-assessment dengan paginasi.
func (a *Adapter) ApplicationAssessments(ctx context.Context, limit, offset int, f_application_id *string, f_assessment_code *string, f_attempt_number *int, f_rubric_document_version_reference_id *string, f_status *string, f_assessor_principal_id *string, f_recommendation_code *string, f_evidence_document_version_reference_id *string) (*Page[model.ApplicationAssessment], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_application_assessments", "human_resource.application_assessments.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid application-assessment pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_application_id != nil, f_assessment_code != nil, f_attempt_number != nil, f_rubric_document_version_reference_id != nil, f_status != nil, f_assessor_principal_id != nil, f_recommendation_code != nil, f_evidence_document_version_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *recruitment.ApplicationAssessmentsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_application_id, "application_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationAssessmentByApplicationID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_assessment_code, "assessment_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationAssessmentByAssessmentCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v := int32(*f_attempt_number)
		resp, err = a.recruitment.FindApplicationAssessmentByAttemptNumber(rpcCtx, &common.Int32FinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_rubric_document_version_reference_id, "rubric_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationAssessmentByRubricDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationAssessmentByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_assessor_principal_id, "assessor_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationAssessmentByAssessorPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_recommendation_code, "recommendation_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationAssessmentByRecommendationCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_evidence_document_version_reference_id, "evidence_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindApplicationAssessmentByEvidenceDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.recruitment.GetAllApplicationAssessments(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "application-assessment")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "application-assessment", toModelApplicationAssessment)
}

// CreateApplicationAssessment membuat application-assessment baru melalui HRM.
func (a *Adapter) CreateApplicationAssessment(ctx context.Context, input model.CreateApplicationAssessmentInput) (*model.ApplicationAssessment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_application_assessments", "create", "human_resource.application_assessments.create")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.ApplicationAssessmentInput{}
	patch.ApplicationId = input.ApplicationID
	patch.AssessmentCode = input.AssessmentCode
	patch.AttemptNumber = optionalInt32Ptr(input.AttemptNumber)
	patch.RubricDocumentVersionReferenceId = input.RubricDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.RubricDigest, "rubric_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.RubricDigest = b
	}
	patch.Status = input.Status
	patch.AssessorPrincipalId = input.AssessorPrincipalID
	patch.ScheduledAt = optionalTimestamp(input.ScheduledAt)
	patch.CompletedAt = optionalTimestamp(input.CompletedAt)
	patch.NormalizedScore = input.NormalizedScore
	patch.RecommendationCode = input.RecommendationCode
	patch.EvidenceDocumentVersionReferenceId = input.EvidenceDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.EvidenceDigest, "evidence_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.EvidenceDigest = b
	}
	resp, err := a.recruitment.CreateApplicationAssessment(rpcCtx, &recruitment.CreateApplicationAssessmentRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "application-assessment")
	}
	return toModelApplicationAssessment(resp.GetItem())
}

// UpdateApplicationAssessment memperbarui application-assessment berdasarkan FieldMask.
func (a *Adapter) UpdateApplicationAssessment(ctx context.Context, rawID string, input model.UpdateApplicationAssessmentInput) (*model.ApplicationAssessment, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_application_assessments", "update", "human_resource.application_assessments.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "application-assessment", "application-assessment")
	if err != nil {
		return nil, err
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &recruitment.ApplicationAssessmentInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "application_id", input.ApplicationID, &patch.ApplicationId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assessment_code", input.AssessmentCode, &patch.AssessmentCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "attempt_number", input.AttemptNumber, &patch.AttemptNumber, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "rubric_document_version_reference_id", input.RubricDocumentVersionReferenceID, &patch.RubricDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "rubric_digest", input.RubricDigest, &patch.RubricDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assessor_principal_id", input.AssessorPrincipalID, &patch.AssessorPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "scheduled_at", input.ScheduledAt, &patch.ScheduledAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "completed_at", input.CompletedAt, &patch.CompletedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "normalized_score", input.NormalizedScore, &patch.NormalizedScore, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "recommendation_code", input.RecommendationCode, &patch.RecommendationCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "evidence_document_version_reference_id", input.EvidenceDocumentVersionReferenceID, &patch.EvidenceDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "evidence_digest", input.EvidenceDigest, &patch.EvidenceDigest, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("application-assessment update is empty", nil)
	}
	resp, err := a.recruitment.UpdateApplicationAssessment(rpcCtx, &recruitment.UpdateApplicationAssessmentRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "application-assessment")
	}
	return toModelApplicationAssessment(resp.GetItem())
}

// DeleteApplicationAssessment menghapus application-assessment berdasarkan ID.
func (a *Adapter) DeleteApplicationAssessment(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_application_assessments", "human_resource.application_assessments.delete", "application-assessment", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.recruitment.DeleteApplicationAssessment(c, req)
	})
}
