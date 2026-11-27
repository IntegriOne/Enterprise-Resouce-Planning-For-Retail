package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	recruitment "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
)

// RequisitionsFiltered mendaftar permintaan rekrutmen dengan filter konsolidasi.
func (a *Adapter) RequisitionsFiltered(ctx context.Context, limit, offset int, requisitionNumber, workforcePlanLineID, positionID, jobID, orgUnitRefID, hiringManagerPrincipalID, recruiterPrincipalID, requisitionType, status, budgetRefID, workLocRefID, approvedByPrincipalID, rejectionReasonCode, rejectedByPrincipalID, idempotencyKey *string) (*Page[model.Requisition], error) {
	if workforcePlanLineID == nil && positionID == nil && jobID == nil && orgUnitRefID == nil && hiringManagerPrincipalID == nil && recruiterPrincipalID == nil && requisitionType == nil && status == nil && budgetRefID == nil && workLocRefID == nil && approvedByPrincipalID == nil && rejectionReasonCode == nil && rejectedByPrincipalID == nil && idempotencyKey == nil {
		return a.Requisitions(ctx, requisitionNumber, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_requisitions", "human_resource.requisitions.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid requisition pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		requisitionNumber != nil,
		workforcePlanLineID != nil,
		positionID != nil,
		jobID != nil,
		orgUnitRefID != nil,
		hiringManagerPrincipalID != nil,
		recruiterPrincipalID != nil,
		requisitionType != nil,
		status != nil,
		budgetRefID != nil,
		workLocRefID != nil,
		approvedByPrincipalID != nil,
		rejectionReasonCode != nil,
		rejectedByPrincipalID != nil,
		idempotencyKey != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *recruitment.RequisitionsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(requisitionNumber, "requisition_number")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByRequisitionNumber(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(workforcePlanLineID, "workforce-plan-line", "workforce_plan_line_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByWorkforcePlanLineID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(positionID, "position", "position_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByPositionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateUUIDFilter(jobID, "job", "job_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByJobID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateUUIDFilter(orgUnitRefID, "organization-unit", "organization_unit_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByOrganizationUnitReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateUUIDFilter(hiringManagerPrincipalID, "principal", "hiring_manager_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByHiringManagerPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateUUIDFilter(recruiterPrincipalID, "principal", "recruiter_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByRecruiterPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(requisitionType, "requisition_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByRequisitionType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 9:
		v, fErr := validateUUIDFilter(budgetRefID, "budget", "budget_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByBudgetReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 10:
		v, fErr := validateUUIDFilter(workLocRefID, "work-location", "work_location_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByWorkLocationReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 11:
		v, fErr := validateUUIDFilter(approvedByPrincipalID, "principal", "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 12:
		v, fErr := validateFilterString(rejectionReasonCode, "rejection_reason_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByRejectionReasonCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 13:
		v, fErr := validateUUIDFilter(rejectedByPrincipalID, "principal", "rejected_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByRejectedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 14:
		v, fErr := validateFilterString(idempotencyKey, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindRequisitionByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.recruitment.GetAllRequisitions(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "requisition")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "requisition", toModelRequisition)
}

// CandidatesFiltered mendaftar kandidat rekrutmen dengan filter konsolidasi.
func (a *Adapter) CandidatesFiltered(ctx context.Context, limit, offset int, candidateNumber, personPartyRefID, contactRefID, sourceType, candidateStatus, createdByPrincipalID *string) (*Page[model.Candidate], error) {
	if personPartyRefID == nil && contactRefID == nil && sourceType == nil && candidateStatus == nil && createdByPrincipalID == nil {
		return a.Candidates(ctx, candidateNumber, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_candidates", "human_resource.candidates.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid candidate pagination", nil)
	}
	if a.recruitment == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		candidateNumber != nil,
		personPartyRefID != nil,
		contactRefID != nil,
		sourceType != nil,
		candidateStatus != nil,
		createdByPrincipalID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *recruitment.CandidatesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(candidateNumber, "candidate_number")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindCandidateByCandidateNumber(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(personPartyRefID, "person-party", "person_party_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindCandidateByPersonPartyReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(contactRefID, "contact", "contact_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindCandidateByContactReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(sourceType, "source_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindCandidateBySourceType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(candidateStatus, "candidate_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindCandidateByCandidateStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateUUIDFilter(createdByPrincipalID, "principal", "created_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.recruitment.FindCandidateByCreatedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.recruitment.GetAllCandidates(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "candidate")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "candidate", toModelCandidate)
}
