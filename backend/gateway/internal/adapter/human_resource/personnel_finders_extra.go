package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
)

// WorkersFiltered mendaftar pekerja dengan filter konsolidasi.
func (a *Adapter) WorkersFiltered(ctx context.Context, limit, offset int, workerNumber, legalEntityID, personPartyRefID, iamPrincipalRefID, status, primaryAssignmentID, terminationReasonCode, createdByPrincipalID *string) (*Page[model.Worker], error) {
	if personPartyRefID == nil && iamPrincipalRefID == nil && status == nil && primaryAssignmentID == nil && terminationReasonCode == nil && createdByPrincipalID == nil {
		return a.Workers(ctx, limit, offset, workerNumber, legalEntityID)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_workers", "human_resource.workers.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid worker pagination", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		workerNumber != nil,
		legalEntityID != nil,
		personPartyRefID != nil,
		iamPrincipalRefID != nil,
		status != nil,
		primaryAssignmentID != nil,
		terminationReasonCode != nil,
		createdByPrincipalID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *personnel.WorkersResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(workerNumber, "worker_number")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerByWorkerNumber(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(legalEntityID, "legal-entity", "legal_entity_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerByPrimaryLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(personPartyRefID, "person-party", "person_party_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerByPersonPartyReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateUUIDFilter(iamPrincipalRefID, "iam-principal", "iam_principal_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerByIAMPrincipalReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateUUIDFilter(primaryAssignmentID, "worker-assignment", "primary_assignment_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerByPrimaryAssignmentID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(terminationReasonCode, "termination_reason_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerByTerminationReasonCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateUUIDFilter(createdByPrincipalID, "principal", "created_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerByCreatedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.personnel.GetAllWorkers(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "worker")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "worker", func(row *personnel.Worker) (*model.Worker, error) {
		return toModelWorker(row), nil
	})
}

// WorkerAssignmentsFiltered mendaftar penugasan pekerja dengan filter konsolidasi.
func (a *Adapter) WorkerAssignmentsFiltered(ctx context.Context, limit, offset int, workerID, positionID, managerWorkerID, assignmentType, requestedByPrincipalID, approvedByPrincipalID *string) (*Page[model.WorkerAssignment], error) {
	if managerWorkerID == nil && assignmentType == nil && requestedByPrincipalID == nil && approvedByPrincipalID == nil {
		return a.WorkerAssignments(ctx, workerID, positionID, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_assignments", "human_resource.worker_assignments.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid worker-assignment pagination", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		workerID != nil,
		positionID != nil,
		managerWorkerID != nil,
		assignmentType != nil,
		requestedByPrincipalID != nil,
		approvedByPrincipalID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *personnel.WorkerAssignmentsResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(workerID, "worker", "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerAssignmentByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateUUIDFilter(positionID, "position", "position_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerAssignmentByPositionID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(managerWorkerID, "worker", "manager_worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerAssignmentByManagerWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(assignmentType, "assignment_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerAssignmentByAssignmentType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateUUIDFilter(requestedByPrincipalID, "principal", "requested_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerAssignmentByRequestedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateUUIDFilter(approvedByPrincipalID, "principal", "approved_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerAssignmentByApprovedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.personnel.GetAllWorkerAssignments(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "worker-assignment")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "worker-assignment", toModelWorkerAssignment)
}

// EmploymentContractsFiltered mendaftar kontrak kerja dengan filter konsolidasi.
func (a *Adapter) EmploymentContractsFiltered(ctx context.Context, limit, offset int, workerID, contractNumber, employingLegalEntityRefID, contractType, contractStatus, governingJurisdictionCode, workCountryCode, collectiveAgreementID, signedDocRefID, signedDocVerRefID *string) (*Page[model.EmploymentContract], error) {
	if workerID != nil && contractNumber == nil && employingLegalEntityRefID == nil && contractType == nil && contractStatus == nil && governingJurisdictionCode == nil && workCountryCode == nil && collectiveAgreementID == nil && signedDocRefID == nil && signedDocVerRefID == nil {
		return a.EmploymentContracts(ctx, *workerID, limit, offset)
	}
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_employment_contracts", "human_resource.employment_contracts.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid employment-contract pagination", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(
		workerID != nil,
		contractNumber != nil,
		employingLegalEntityRefID != nil,
		contractType != nil,
		contractStatus != nil,
		governingJurisdictionCode != nil,
		workCountryCode != nil,
		collectiveAgreementID != nil,
		signedDocRefID != nil,
		signedDocVerRefID != nil,
	)
	if err != nil {
		return nil, err
	}
	var resp *personnel.EmploymentContractsResponse
	switch idx {
	case 0:
		v, fErr := validateUUIDFilter(workerID, "worker", "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmploymentContractByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(contractNumber, "contract_number")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmploymentContractByContractNumber(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateUUIDFilter(employingLegalEntityRefID, "legal-entity", "employing_legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmploymentContractByEmployingLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(contractType, "contract_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmploymentContractByContractType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(contractStatus, "contract_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmploymentContractByContractStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(governingJurisdictionCode, "governing_jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmploymentContractByGoverningJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(workCountryCode, "work_country_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmploymentContractByWorkCountryCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateUUIDFilter(collectiveAgreementID, "collective-agreement", "collective_agreement_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmploymentContractByCollectiveAgreementID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateUUIDFilter(signedDocRefID, "document", "signed_document_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmploymentContractBySignedDocumentReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 9:
		v, fErr := validateUUIDFilter(signedDocVerRefID, "document-version", "signed_document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmploymentContractBySignedDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.personnel.GetAllEmploymentContracts(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "employment-contract")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "employment-contract", toModelEmploymentContract)
}
