package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
)

// JobFamiliesFiltered mendaftar rumpun jabatan dengan filter opsional family_code.
func (a *Adapter) JobFamiliesFiltered(ctx context.Context, limit, offset int, familyCode *string) (*Page[model.JobFamily], error) {
	if familyCode == nil {
		return a.JobFamilies(ctx, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_job_families", "human_resource.job_families.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid job-family pagination", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	var resp *workforce.JobFamiliesResponse
	if familyCode != nil {
		v, fErr := validateFilterString(familyCode, "family_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindJobFamilyByFamilyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	} else {
		resp, err = a.workforce.GetAllJobFamilies(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "job-family")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "job-family", toModelJobFamily)
}

// JobsFiltered mendaftar jabatan dengan filter konsolidasi.
func (a *Adapter) JobsFiltered(ctx context.Context, limit, offset int, jobFamilyID, jobCode, jobLevelCode *string) (*Page[model.Job], error) {
	if jobCode == nil && jobLevelCode == nil {
		return a.Jobs(ctx, limit, offset, jobFamilyID)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_jobs", "human_resource.jobs.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid job pagination", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(jobFamilyID != nil, jobCode != nil, jobLevelCode != nil)
	if err != nil {
		return nil, err
	}
	var resp *workforce.JobsResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(jobFamilyID, "job-family", "job_family_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindJobByJobFamilyID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(jobCode, "job_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindJobByJobCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(jobLevelCode, "job_level_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindJobByJobLevelCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.workforce.GetAllJobs(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "job")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "job", toModelJob)
}

// PositionsFiltered mendaftar posisi jabatan dengan filter konsolidasi.
func (a *Adapter) PositionsFiltered(ctx context.Context, limit, offset int, legalEntityID, positionCode, jobID, orgUnitRefID, reportsToPosID, positionStatus, workLocRefID, costCenterRefID *string) (*Page[model.Position], error) {
	if positionCode == nil && jobID == nil && orgUnitRefID == nil && reportsToPosID == nil && positionStatus == nil && workLocRefID == nil && costCenterRefID == nil {
		return a.Positions(ctx, limit, offset, legalEntityID)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_positions", "human_resource.positions.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid position pagination", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		legalEntityID != nil,
		positionCode != nil,
		jobID != nil,
		orgUnitRefID != nil,
		reportsToPosID != nil,
		positionStatus != nil,
		workLocRefID != nil,
		costCenterRefID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *workforce.PositionsResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(legalEntityID, "legal-entity", "legal_entity_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindPositionByLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(positionCode, "position_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindPositionByPositionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(jobID, "job", "job_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindPositionByJobID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateUUIDFilter(orgUnitRefID, "organization-unit", "organization_unit_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindPositionByOrganizationUnitReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateUUIDFilter(reportsToPosID, "position", "reports_to_position_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindPositionByReportsToPositionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(positionStatus, "position_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindPositionByPositionStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateUUIDFilter(workLocRefID, "work-location", "work_location_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindPositionByWorkLocationReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateUUIDFilter(costCenterRefID, "cost-center", "cost_center_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindPositionByCostCenterReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.workforce.GetAllPositions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "position")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "position", toModelPosition)
}

// WorkforcePlansFiltered mendaftar rencana tenaga kerja dengan filter konsolidasi.
func (a *Adapter) WorkforcePlansFiltered(ctx context.Context, limit, offset int, planCode, legalEntityRefID, scenarioCode, status, financeBudgetRefID, createdByPrincipalID, approvedByPrincipalID *string) (*Page[model.WorkforcePlan], error) {
	if planCode == nil && legalEntityRefID == nil && scenarioCode == nil && status == nil && financeBudgetRefID == nil && createdByPrincipalID == nil && approvedByPrincipalID == nil {
		return a.WorkforcePlans(ctx, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_workforce_plans", "human_resource.workforce_plans.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid workforce-plan pagination", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		planCode != nil,
		legalEntityRefID != nil,
		scenarioCode != nil,
		status != nil,
		financeBudgetRefID != nil,
		createdByPrincipalID != nil,
		approvedByPrincipalID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *workforce.WorkforcePlansResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(planCode, "plan_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanByPlanCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(legalEntityRefID, "legal-entity", "legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanByLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(scenarioCode, "scenario_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanByScenarioCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateUUIDFilter(financeBudgetRefID, "finance-budget", "finance_budget_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanByFinanceBudgetReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateUUIDFilter(createdByPrincipalID, "principal", "created_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanByCreatedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateUUIDFilter(approvedByPrincipalID, "principal", "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.workforce.GetAllWorkforcePlans(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "workforce-plan")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "workforce-plan", func(row *workforce.WorkforcePlan) (*model.WorkforcePlan, error) {
		return toModelWorkforcePlan(row), nil
	})
}

// WorkforcePlanLinesFiltered mendaftar baris rencana tenaga kerja dengan filter konsolidasi.
func (a *Adapter) WorkforcePlanLinesFiltered(ctx context.Context, limit, offset int, workforcePlanID *string, lineNumber *int, orgUnitRefID, jobID, workLocRefID, currencyCode, actionCode *string) (*Page[model.WorkforcePlanLine], error) {
	if lineNumber == nil && orgUnitRefID == nil && jobID == nil && workLocRefID == nil && currencyCode == nil && actionCode == nil {
		return a.WorkforcePlanLines(ctx, workforcePlanID, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_workforce_plan_lines", "human_resource.workforce_plan_lines.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid workforce-plan-line pagination", nil)
	}
	if a.workforce == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		workforcePlanID != nil,
		lineNumber != nil,
		orgUnitRefID != nil,
		jobID != nil,
		workLocRefID != nil,
		currencyCode != nil,
		actionCode != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *workforce.WorkforcePlanLinesResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(workforcePlanID, "workforce-plan", "workforce_plan_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanLineByWorkforcePlanID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v := int32(*lineNumber)
		resp, err = a.workforce.FindWorkforcePlanLineByLineNumber(rpcCtx, &common.Int32FinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(orgUnitRefID, "organization-unit", "organization_unit_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanLineByOrganizationUnitReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateUUIDFilter(jobID, "job", "job_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanLineByJobID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateUUIDFilter(workLocRefID, "work-location", "work_location_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanLineByWorkLocationReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(currencyCode, "currency_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanLineByCurrencyCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(actionCode, "action_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.workforce.FindWorkforcePlanLineByActionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.workforce.GetAllWorkforcePlanLines(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "workforce-plan-line")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "workforce-plan-line", func(row *workforce.WorkforcePlanLine) (*model.WorkforcePlanLine, error) {
		return toModelWorkforcePlanLine(row), nil
	})
}
