package human_resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	performance "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/performance/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelPerformanceCycle(row *performance.PerformanceCycle) (*model.PerformanceCycle, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource performance-cycle response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource performance-cycle ID", nil)
	}
	out := &model.PerformanceCycle{ID: nodeID("performance-cycle", row.GetId()), RawID: row.GetId()}
	out.CycleCode = row.GetCycleCode()
	out.CycleName = row.GetCycleName()
	out.CycleType = row.GetCycleType()
	out.Status = row.GetStatus()
	out.PeriodStart = row.GetPeriodStart()
	out.PeriodEnd = row.GetPeriodEnd()
	out.SelfReviewDueAt = pbTimePtr(row.GetSelfReviewDueAt())
	out.ManagerReviewDueAt = pbTimePtr(row.GetManagerReviewDueAt())
	out.CalibrationDueAt = pbTimePtr(row.GetCalibrationDueAt())
	out.LaunchedAt = pbTimePtr(row.GetLaunchedAt())
	out.LaunchedByPrincipalID = row.LaunchedByPrincipalId
	if doc := row.GetConfiguration(); doc != nil {
		out.Configuration = pbJSONString(doc.GetUtf8())
	}
	return out, nil
}

// PerformanceCycle membaca detail performance-cycle berdasarkan ID.
func (a *Adapter) PerformanceCycle(ctx context.Context, rawID string) (*model.PerformanceCycle, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_performance_cycles", "human_resource.performance_cycles.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-cycle", "performance-cycle")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.GetPerformanceCycleByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "performance-cycle")
	}
	return toModelPerformanceCycle(resp.GetItem())
}

// PerformanceCycles mendaftar dan memfilter performance-cycle dengan paginasi.
func (a *Adapter) PerformanceCycles(ctx context.Context, limit, offset int, f_cycle_code *string, f_cycle_type *string, f_status *string, f_launched_by_principal_id *string) (*Page[model.PerformanceCycle], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_performance_cycles", "human_resource.performance_cycles.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid performance-cycle pagination", nil)
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_cycle_code != nil, f_cycle_type != nil, f_status != nil, f_launched_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *performance.PerformanceCyclesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_cycle_code, "cycle_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceCycleByCycleCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_cycle_type, "cycle_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceCycleByCycleType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceCycleByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_launched_by_principal_id, "launched_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceCycleByLaunchedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.performance.GetAllPerformanceCycles(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "performance-cycle")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "performance-cycle", toModelPerformanceCycle)
}

// CreatePerformanceCycle membuat performance-cycle baru melalui HRM.
func (a *Adapter) CreatePerformanceCycle(ctx context.Context, input model.CreatePerformanceCycleInput) (*model.PerformanceCycle, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_cycles", "create", "human_resource.performance_cycles.create")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &performance.PerformanceCycleInput{}
	patch.CycleCode = input.CycleCode
	patch.CycleName = input.CycleName
	patch.CycleType = input.CycleType
	patch.Status = input.Status
	patch.PeriodStart = input.PeriodStart
	patch.PeriodEnd = input.PeriodEnd
	patch.SelfReviewDueAt = optionalTimestamp(input.SelfReviewDueAt)
	patch.ManagerReviewDueAt = optionalTimestamp(input.ManagerReviewDueAt)
	patch.CalibrationDueAt = optionalTimestamp(input.CalibrationDueAt)
	patch.LaunchedAt = optionalTimestamp(input.LaunchedAt)
	patch.LaunchedByPrincipalId = input.LaunchedByPrincipalID
	if b, jErr := optionalJSONBytes(input.Configuration, "configuration"); jErr != nil {
		return nil, jErr
	} else if b != nil {
		patch.Configuration = &common.JsonDocument{Utf8: b}
	}
	resp, err := a.performance.CreatePerformanceCycle(rpcCtx, &performance.CreatePerformanceCycleRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-cycle")
	}
	return toModelPerformanceCycle(resp.GetItem())
}

// UpdatePerformanceCycle memperbarui performance-cycle berdasarkan FieldMask.
func (a *Adapter) UpdatePerformanceCycle(ctx context.Context, rawID string, input model.UpdatePerformanceCycleInput) (*model.PerformanceCycle, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_cycles", "update", "human_resource.performance_cycles.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-cycle", "performance-cycle")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &performance.PerformanceCycleInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "cycle_code", input.CycleCode, &patch.CycleCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "cycle_name", input.CycleName, &patch.CycleName, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "cycle_type", input.CycleType, &patch.CycleType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "period_start", input.PeriodStart, &patch.PeriodStart, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "period_end", input.PeriodEnd, &patch.PeriodEnd, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "self_review_due_at", input.SelfReviewDueAt, &patch.SelfReviewDueAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "manager_review_due_at", input.ManagerReviewDueAt, &patch.ManagerReviewDueAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "calibration_due_at", input.CalibrationDueAt, &patch.CalibrationDueAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "launched_at", input.LaunchedAt, &patch.LaunchedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "launched_by_principal_id", input.LaunchedByPrincipalID, &patch.LaunchedByPrincipalId, true); err != nil {
		return nil, err
	}
	if val, set := input.Configuration.ValueOK(); set {
		if val == nil {
			patch.Configuration = nil
		} else if b, jErr := requiredJSONBytes(*val, "configuration"); jErr != nil {
			return nil, jErr
		} else {
			patch.Configuration = &common.JsonDocument{Utf8: b}
		}
		mask.Paths = append(mask.Paths, "configuration")
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("performance-cycle update is empty", nil)
	}
	resp, err := a.performance.UpdatePerformanceCycle(rpcCtx, &performance.UpdatePerformanceCycleRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-cycle")
	}
	return toModelPerformanceCycle(resp.GetItem())
}

// DeletePerformanceCycle menghapus performance-cycle berdasarkan ID.
func (a *Adapter) DeletePerformanceCycle(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_performance_cycles", "human_resource.performance_cycles.delete", "performance-cycle", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.performance.DeletePerformanceCycle(c, req)
	})
}

func toModelPerformanceGoal(row *performance.PerformanceGoal) (*model.PerformanceGoal, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource performance-goal response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource performance-goal ID", nil)
	}
	out := &model.PerformanceGoal{ID: nodeID("performance-goal", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.PerformanceCycleID = row.GetPerformanceCycleId()
	out.ParentGoalID = row.ParentGoalId
	out.GoalCode = row.GoalCode
	out.GoalTitle = row.GetGoalTitle()
	out.GoalDescription = row.GoalDescription
	out.GoalType = row.GetGoalType()
	if doc := row.GetMeasureDefinition(); doc != nil {
		out.MeasureDefinition = pbJSONString(doc.GetUtf8())
	}
	out.TargetValue = row.TargetValue
	out.ActualValue = row.ActualValue
	out.UnitCode = row.UnitCode
	out.WeightPercent = row.WeightPercent
	out.DueDate = row.DueDate
	out.GoalStatus = row.GetGoalStatus()
	out.ApprovedByPrincipalID = row.ApprovedByPrincipalId
	out.ApprovedAt = pbTimePtr(row.GetApprovedAt())
	return out, nil
}

// PerformanceGoal membaca detail performance-goal berdasarkan ID.
func (a *Adapter) PerformanceGoal(ctx context.Context, rawID string) (*model.PerformanceGoal, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_performance_goals", "human_resource.performance_goals.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-goal", "performance-goal")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.GetPerformanceGoalByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "performance-goal")
	}
	return toModelPerformanceGoal(resp.GetItem())
}

// PerformanceGoals mendaftar dan memfilter performance-goal dengan paginasi.
func (a *Adapter) PerformanceGoals(ctx context.Context, limit, offset int, f_worker_id *string, f_performance_cycle_id *string, f_parent_goal_id *string, f_goal_code *string, f_goal_type *string, f_unit_code *string, f_goal_status *string, f_approved_by_principal_id *string) (*Page[model.PerformanceGoal], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_performance_goals", "human_resource.performance_goals.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid performance-goal pagination", nil)
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_performance_cycle_id != nil, f_parent_goal_id != nil, f_goal_code != nil, f_goal_type != nil, f_unit_code != nil, f_goal_status != nil, f_approved_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *performance.PerformanceGoalsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_performance_cycle_id, "performance_cycle_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalByPerformanceCycleID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_parent_goal_id, "parent_goal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalByParentGoalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_goal_code, "goal_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalByGoalCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_goal_type, "goal_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalByGoalType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_unit_code, "unit_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalByUnitCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_goal_status, "goal_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalByGoalStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_approved_by_principal_id, "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.performance.GetAllPerformanceGoals(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "performance-goal")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "performance-goal", toModelPerformanceGoal)
}

// CreatePerformanceGoal membuat performance-goal baru melalui HRM.
func (a *Adapter) CreatePerformanceGoal(ctx context.Context, input model.CreatePerformanceGoalInput) (*model.PerformanceGoal, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_goals", "create", "human_resource.performance_goals.create")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &performance.PerformanceGoalInput{}
	patch.WorkerId = input.WorkerID
	patch.PerformanceCycleId = input.PerformanceCycleID
	patch.ParentGoalId = input.ParentGoalID
	patch.GoalCode = input.GoalCode
	patch.GoalTitle = input.GoalTitle
	patch.GoalDescription = input.GoalDescription
	patch.GoalType = input.GoalType
	if b, jErr := optionalJSONBytes(input.MeasureDefinition, "measure_definition"); jErr != nil {
		return nil, jErr
	} else if b != nil {
		patch.MeasureDefinition = &common.JsonDocument{Utf8: b}
	}
	patch.TargetValue = input.TargetValue
	patch.ActualValue = input.ActualValue
	patch.UnitCode = input.UnitCode
	patch.WeightPercent = input.WeightPercent
	patch.DueDate = input.DueDate
	patch.GoalStatus = input.GoalStatus
	patch.ApprovedByPrincipalId = input.ApprovedByPrincipalID
	patch.ApprovedAt = optionalTimestamp(input.ApprovedAt)
	resp, err := a.performance.CreatePerformanceGoal(rpcCtx, &performance.CreatePerformanceGoalRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-goal")
	}
	return toModelPerformanceGoal(resp.GetItem())
}

// UpdatePerformanceGoal memperbarui performance-goal berdasarkan FieldMask.
func (a *Adapter) UpdatePerformanceGoal(ctx context.Context, rawID string, input model.UpdatePerformanceGoalInput) (*model.PerformanceGoal, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_goals", "update", "human_resource.performance_goals.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-goal", "performance-goal")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &performance.PerformanceGoalInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "performance_cycle_id", input.PerformanceCycleID, &patch.PerformanceCycleId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "parent_goal_id", input.ParentGoalID, &patch.ParentGoalId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "goal_code", input.GoalCode, &patch.GoalCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "goal_title", input.GoalTitle, &patch.GoalTitle, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "goal_description", input.GoalDescription, &patch.GoalDescription, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "goal_type", input.GoalType, &patch.GoalType, true); err != nil {
		return nil, err
	}
	if val, set := input.MeasureDefinition.ValueOK(); set {
		if val == nil {
			patch.MeasureDefinition = nil
		} else if b, jErr := requiredJSONBytes(*val, "measure_definition"); jErr != nil {
			return nil, jErr
		} else {
			patch.MeasureDefinition = &common.JsonDocument{Utf8: b}
		}
		mask.Paths = append(mask.Paths, "measure_definition")
	}
	if err := applyStringMask(mask, "target_value", input.TargetValue, &patch.TargetValue, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "actual_value", input.ActualValue, &patch.ActualValue, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "unit_code", input.UnitCode, &patch.UnitCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "weight_percent", input.WeightPercent, &patch.WeightPercent, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "due_date", input.DueDate, &patch.DueDate, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "goal_status", input.GoalStatus, &patch.GoalStatus, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "approved_by_principal_id", input.ApprovedByPrincipalID, &patch.ApprovedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "approved_at", input.ApprovedAt, &patch.ApprovedAt, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("performance-goal update is empty", nil)
	}
	resp, err := a.performance.UpdatePerformanceGoal(rpcCtx, &performance.UpdatePerformanceGoalRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-goal")
	}
	return toModelPerformanceGoal(resp.GetItem())
}

// DeletePerformanceGoal menghapus performance-goal berdasarkan ID.
func (a *Adapter) DeletePerformanceGoal(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_performance_goals", "human_resource.performance_goals.delete", "performance-goal", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.performance.DeletePerformanceGoal(c, req)
	})
}

func toModelPerformanceReview(row *performance.PerformanceReview) (*model.PerformanceReview, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource performance-review response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource performance-review ID", nil)
	}
	out := &model.PerformanceReview{ID: nodeID("performance-review", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.PerformanceCycleID = row.GetPerformanceCycleId()
	out.ReviewerPrincipalID = row.GetReviewerPrincipalId()
	out.ReviewerRole = row.GetReviewerRole()
	out.ReviewVersion = int(row.GetReviewVersion())
	out.OverallRatingCode = row.OverallRatingCode
	if doc := row.GetCompetencyAssessment(); doc != nil {
		out.CompetencyAssessment = pbJSONString(doc.GetUtf8())
	}
	out.Narrative = row.Narrative
	out.DevelopmentPlanID = row.DevelopmentPlanId
	out.SubmittedAt = pbTimePtr(row.GetSubmittedAt())
	out.FinalizedAt = pbTimePtr(row.GetFinalizedAt())
	out.VisibilityClass = row.GetVisibilityClass()
	out.ContentDigest = pbBase64StringValue(row.GetContentDigest())
	return out, nil
}

// PerformanceReview membaca detail performance-review berdasarkan ID.
func (a *Adapter) PerformanceReview(ctx context.Context, rawID string) (*model.PerformanceReview, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_performance_reviews", "human_resource.performance_reviews.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-review", "performance-review")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.GetPerformanceReviewByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "performance-review")
	}
	return toModelPerformanceReview(resp.GetItem())
}

// PerformanceReviews mendaftar dan memfilter performance-review dengan paginasi.
func (a *Adapter) PerformanceReviews(ctx context.Context, limit, offset int, f_worker_id *string, f_performance_cycle_id *string, f_reviewer_principal_id *string, f_overall_rating_code *string, f_development_plan_id *string) (*Page[model.PerformanceReview], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_performance_reviews", "human_resource.performance_reviews.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid performance-review pagination", nil)
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_performance_cycle_id != nil, f_reviewer_principal_id != nil, f_overall_rating_code != nil, f_development_plan_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *performance.PerformanceReviewsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_performance_cycle_id, "performance_cycle_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewByPerformanceCycleID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_reviewer_principal_id, "reviewer_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewByReviewerPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_overall_rating_code, "overall_rating_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewByOverallRatingCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_development_plan_id, "development_plan_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewByDevelopmentPlanID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.performance.GetAllPerformanceReviews(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "performance-review")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "performance-review", toModelPerformanceReview)
}

// CreatePerformanceReview membuat performance-review baru melalui HRM.
func (a *Adapter) CreatePerformanceReview(ctx context.Context, input model.CreatePerformanceReviewInput) (*model.PerformanceReview, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_reviews", "create", "human_resource.performance_reviews.create")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &performance.PerformanceReviewInput{}
	patch.WorkerId = input.WorkerID
	patch.PerformanceCycleId = input.PerformanceCycleID
	patch.ReviewerPrincipalId = input.ReviewerPrincipalID
	patch.ReviewerRole = input.ReviewerRole
	patch.ReviewVersion = optionalInt32Ptr(input.ReviewVersion)
	patch.OverallRatingCode = input.OverallRatingCode
	if b, jErr := optionalJSONBytes(input.CompetencyAssessment, "competency_assessment"); jErr != nil {
		return nil, jErr
	} else if b != nil {
		patch.CompetencyAssessment = &common.JsonDocument{Utf8: b}
	}
	patch.Narrative = input.Narrative
	patch.DevelopmentPlanId = input.DevelopmentPlanID
	patch.SubmittedAt = optionalTimestamp(input.SubmittedAt)
	patch.FinalizedAt = optionalTimestamp(input.FinalizedAt)
	patch.VisibilityClass = input.VisibilityClass
	if b, bErr := optionalBase64Bytes(input.ContentDigest, "content_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.ContentDigest = b
	}
	resp, err := a.performance.CreatePerformanceReview(rpcCtx, &performance.CreatePerformanceReviewRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-review")
	}
	return toModelPerformanceReview(resp.GetItem())
}

// UpdatePerformanceReview memperbarui performance-review berdasarkan FieldMask.
func (a *Adapter) UpdatePerformanceReview(ctx context.Context, rawID string, input model.UpdatePerformanceReviewInput) (*model.PerformanceReview, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_reviews", "update", "human_resource.performance_reviews.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-review", "performance-review")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &performance.PerformanceReviewInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "performance_cycle_id", input.PerformanceCycleID, &patch.PerformanceCycleId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "reviewer_principal_id", input.ReviewerPrincipalID, &patch.ReviewerPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "reviewer_role", input.ReviewerRole, &patch.ReviewerRole, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "review_version", input.ReviewVersion, &patch.ReviewVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "overall_rating_code", input.OverallRatingCode, &patch.OverallRatingCode, true); err != nil {
		return nil, err
	}
	if val, set := input.CompetencyAssessment.ValueOK(); set {
		if val == nil {
			patch.CompetencyAssessment = nil
		} else if b, jErr := requiredJSONBytes(*val, "competency_assessment"); jErr != nil {
			return nil, jErr
		} else {
			patch.CompetencyAssessment = &common.JsonDocument{Utf8: b}
		}
		mask.Paths = append(mask.Paths, "competency_assessment")
	}
	if err := applyStringMask(mask, "narrative", input.Narrative, &patch.Narrative, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "development_plan_id", input.DevelopmentPlanID, &patch.DevelopmentPlanId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "submitted_at", input.SubmittedAt, &patch.SubmittedAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "finalized_at", input.FinalizedAt, &patch.FinalizedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "visibility_class", input.VisibilityClass, &patch.VisibilityClass, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "content_digest", input.ContentDigest, &patch.ContentDigest, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("performance-review update is empty", nil)
	}
	resp, err := a.performance.UpdatePerformanceReview(rpcCtx, &performance.UpdatePerformanceReviewRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-review")
	}
	return toModelPerformanceReview(resp.GetItem())
}

// DeletePerformanceReview menghapus performance-review berdasarkan ID.
func (a *Adapter) DeletePerformanceReview(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_performance_reviews", "human_resource.performance_reviews.delete", "performance-review", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.performance.DeletePerformanceReview(c, req)
	})
}

func toModelPerformanceGoalUpdate(row *performance.PerformanceGoalUpdate) (*model.PerformanceGoalUpdate, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource performance-goal-update response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource performance-goal-update ID", nil)
	}
	out := &model.PerformanceGoalUpdate{ID: nodeID("performance-goal-update", row.GetId()), RawID: row.GetId()}
	out.GoalID = row.GetGoalId()
	out.ObservedAt = pbTimePtr(row.GetObservedAt())
	out.ActualValue = row.GetActualValue()
	out.EvidenceDocumentVersionReferenceID = row.GetEvidenceDocumentVersionReferenceId()
	out.EvidenceDigest = pbBase64StringValue(row.GetEvidenceDigest())
	out.RecordedByPrincipalID = row.GetRecordedByPrincipalId()
	out.RecordedAt = pbTimePtr(row.GetRecordedAt())
	out.SupersedesUpdateID = row.SupersedesUpdateId
	out.IdempotencyKey = row.GetIdempotencyKey()
	return out, nil
}

// PerformanceGoalUpdate membaca detail performance-goal-update berdasarkan ID.
func (a *Adapter) PerformanceGoalUpdate(ctx context.Context, rawID string) (*model.PerformanceGoalUpdate, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_performance_goal_updates", "human_resource.performance_goal_updates.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-goal-update", "performance-goal-update")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.GetPerformanceGoalUpdateByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "performance-goal-update")
	}
	return toModelPerformanceGoalUpdate(resp.GetItem())
}

// PerformanceGoalUpdates mendaftar dan memfilter performance-goal-update dengan paginasi.
func (a *Adapter) PerformanceGoalUpdates(ctx context.Context, limit, offset int, f_goal_id *string, f_evidence_document_version_reference_id *string, f_recorded_by_principal_id *string, f_supersedes_update_id *string, f_idempotency_key *string) (*Page[model.PerformanceGoalUpdate], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_performance_goal_updates", "human_resource.performance_goal_updates.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid performance-goal-update pagination", nil)
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_goal_id != nil, f_evidence_document_version_reference_id != nil, f_recorded_by_principal_id != nil, f_supersedes_update_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *performance.PerformanceGoalUpdatesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_goal_id, "goal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalUpdateByGoalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_evidence_document_version_reference_id, "evidence_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalUpdateByEvidenceDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_recorded_by_principal_id, "recorded_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalUpdateByRecordedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_supersedes_update_id, "supersedes_update_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalUpdateBySupersedesUpdateID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceGoalUpdateByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.performance.GetAllPerformanceGoalUpdates(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "performance-goal-update")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "performance-goal-update", toModelPerformanceGoalUpdate)
}

// CreatePerformanceGoalUpdate membuat performance-goal-update baru melalui HRM.
func (a *Adapter) CreatePerformanceGoalUpdate(ctx context.Context, input model.CreatePerformanceGoalUpdateInput) (*model.PerformanceGoalUpdate, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_goal_updates", "create", "human_resource.performance_goal_updates.create")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &performance.PerformanceGoalUpdateInput{}
	patch.GoalId = input.GoalID
	patch.ObservedAt = optionalTimestamp(input.ObservedAt)
	patch.ActualValue = input.ActualValue
	patch.EvidenceDocumentVersionReferenceId = input.EvidenceDocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.EvidenceDigest, "evidence_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.EvidenceDigest = b
	}
	patch.SupersedesUpdateId = input.SupersedesUpdateID
	patch.IdempotencyKey = input.IdempotencyKey
	resp, err := a.performance.CreatePerformanceGoalUpdate(rpcCtx, &performance.CreatePerformanceGoalUpdateRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-goal-update")
	}
	return toModelPerformanceGoalUpdate(resp.GetItem())
}

// UpdatePerformanceGoalUpdate memperbarui performance-goal-update berdasarkan FieldMask.
func (a *Adapter) UpdatePerformanceGoalUpdate(ctx context.Context, rawID string, input model.UpdatePerformanceGoalUpdateInput) (*model.PerformanceGoalUpdate, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_goal_updates", "update", "human_resource.performance_goal_updates.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-goal-update", "performance-goal-update")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &performance.PerformanceGoalUpdateInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "goal_id", input.GoalID, &patch.GoalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "observed_at", input.ObservedAt, &patch.ObservedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "actual_value", input.ActualValue, &patch.ActualValue, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "evidence_document_version_reference_id", input.EvidenceDocumentVersionReferenceID, &patch.EvidenceDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "evidence_digest", input.EvidenceDigest, &patch.EvidenceDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "supersedes_update_id", input.SupersedesUpdateID, &patch.SupersedesUpdateId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("performance-goal-update update is empty", nil)
	}
	resp, err := a.performance.UpdatePerformanceGoalUpdate(rpcCtx, &performance.UpdatePerformanceGoalUpdateRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-goal-update")
	}
	return toModelPerformanceGoalUpdate(resp.GetItem())
}

// DeletePerformanceGoalUpdate menghapus performance-goal-update berdasarkan ID.
func (a *Adapter) DeletePerformanceGoalUpdate(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_performance_goal_updates", "human_resource.performance_goal_updates.delete", "performance-goal-update", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.performance.DeletePerformanceGoalUpdate(c, req)
	})
}

func toModelPerformanceReviewGoalScore(row *performance.PerformanceReviewGoalScore) (*model.PerformanceReviewGoalScore, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource performance-review-goal-score response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource performance-review-goal-score ID", nil)
	}
	out := &model.PerformanceReviewGoalScore{ID: nodeID("performance-review-goal-score", row.GetId()), RawID: row.GetId()}
	out.PerformanceReviewID = row.GetPerformanceReviewId()
	out.GoalID = row.GetGoalId()
	out.WorkerID = row.GetWorkerId()
	out.PerformanceCycleID = row.GetPerformanceCycleId()
	out.Rating = int(row.GetRating())
	out.WeightPercentSnapshot = row.GetWeightPercentSnapshot()
	out.RubricDocumentVersionReferenceID = row.GetRubricDocumentVersionReferenceId()
	out.EvidenceDocumentVersionReferenceID = row.GetEvidenceDocumentVersionReferenceId()
	return out, nil
}

// PerformanceReviewGoalScore membaca detail performance-review-goal-score berdasarkan ID.
func (a *Adapter) PerformanceReviewGoalScore(ctx context.Context, rawID string) (*model.PerformanceReviewGoalScore, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_performance_review_goal_scores", "human_resource.performance_review_goal_scores.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-review-goal-score", "performance-review-goal-score")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.performance.GetPerformanceReviewGoalScoreByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "performance-review-goal-score")
	}
	return toModelPerformanceReviewGoalScore(resp.GetItem())
}

// PerformanceReviewGoalScores mendaftar dan memfilter performance-review-goal-score dengan paginasi.
func (a *Adapter) PerformanceReviewGoalScores(ctx context.Context, limit, offset int, f_performance_review_id *string, f_goal_id *string, f_worker_id *string, f_performance_cycle_id *string, f_rubric_document_version_reference_id *string, f_evidence_document_version_reference_id *string) (*Page[model.PerformanceReviewGoalScore], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_performance_review_goal_scores", "human_resource.performance_review_goal_scores.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid performance-review-goal-score pagination", nil)
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_performance_review_id != nil, f_goal_id != nil, f_worker_id != nil, f_performance_cycle_id != nil, f_rubric_document_version_reference_id != nil, f_evidence_document_version_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *performance.PerformanceReviewGoalScoresResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_performance_review_id, "performance_review_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewGoalScoreByPerformanceReviewID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_goal_id, "goal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewGoalScoreByGoalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewGoalScoreByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_performance_cycle_id, "performance_cycle_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewGoalScoreByPerformanceCycleID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_rubric_document_version_reference_id, "rubric_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewGoalScoreByRubricDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_evidence_document_version_reference_id, "evidence_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.performance.FindPerformanceReviewGoalScoreByEvidenceDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.performance.GetAllPerformanceReviewGoalScores(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "performance-review-goal-score")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "performance-review-goal-score", toModelPerformanceReviewGoalScore)
}

// CreatePerformanceReviewGoalScore membuat performance-review-goal-score baru melalui HRM.
func (a *Adapter) CreatePerformanceReviewGoalScore(ctx context.Context, input model.CreatePerformanceReviewGoalScoreInput) (*model.PerformanceReviewGoalScore, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_review_goal_scores", "create", "human_resource.performance_review_goal_scores.create")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &performance.PerformanceReviewGoalScoreInput{}
	patch.PerformanceReviewId = input.PerformanceReviewID
	patch.GoalId = input.GoalID
	patch.WorkerId = input.WorkerID
	patch.PerformanceCycleId = input.PerformanceCycleID
	patch.Rating = optionalInt32Ptr(input.Rating)
	patch.WeightPercentSnapshot = input.WeightPercentSnapshot
	patch.RubricDocumentVersionReferenceId = input.RubricDocumentVersionReferenceID
	patch.EvidenceDocumentVersionReferenceId = input.EvidenceDocumentVersionReferenceID
	resp, err := a.performance.CreatePerformanceReviewGoalScore(rpcCtx, &performance.CreatePerformanceReviewGoalScoreRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-review-goal-score")
	}
	return toModelPerformanceReviewGoalScore(resp.GetItem())
}

// UpdatePerformanceReviewGoalScore memperbarui performance-review-goal-score berdasarkan FieldMask.
func (a *Adapter) UpdatePerformanceReviewGoalScore(ctx context.Context, rawID string, input model.UpdatePerformanceReviewGoalScoreInput) (*model.PerformanceReviewGoalScore, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_performance_review_goal_scores", "update", "human_resource.performance_review_goal_scores.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "performance-review-goal-score", "performance-review-goal-score")
	if err != nil {
		return nil, err
	}
	if a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &performance.PerformanceReviewGoalScoreInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "performance_review_id", input.PerformanceReviewID, &patch.PerformanceReviewId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "goal_id", input.GoalID, &patch.GoalId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "performance_cycle_id", input.PerformanceCycleID, &patch.PerformanceCycleId, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "rating", input.Rating, &patch.Rating, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "weight_percent_snapshot", input.WeightPercentSnapshot, &patch.WeightPercentSnapshot, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "rubric_document_version_reference_id", input.RubricDocumentVersionReferenceID, &patch.RubricDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "evidence_document_version_reference_id", input.EvidenceDocumentVersionReferenceID, &patch.EvidenceDocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("performance-review-goal-score update is empty", nil)
	}
	resp, err := a.performance.UpdatePerformanceReviewGoalScore(rpcCtx, &performance.UpdatePerformanceReviewGoalScoreRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "performance-review-goal-score")
	}
	return toModelPerformanceReviewGoalScore(resp.GetItem())
}

// DeletePerformanceReviewGoalScore menghapus performance-review-goal-score berdasarkan ID.
func (a *Adapter) DeletePerformanceReviewGoalScore(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.performance == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_performance_review_goal_scores", "human_resource.performance_review_goal_scores.delete", "performance-review-goal-score", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.performance.DeletePerformanceReviewGoalScore(c, req)
	})
}
