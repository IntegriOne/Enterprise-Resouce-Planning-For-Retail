package human_resource

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	employeerelations "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/employeerelations/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelEmployeeRelationsCase(row *employeerelations.EmployeeRelationsCase) (*model.EmployeeRelationsCase, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource employee-relations-case response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource employee-relations-case ID", nil)
	}
	out := &model.EmployeeRelationsCase{ID: nodeID("employee-relations-case", row.GetId()), RawID: row.GetId()}
	out.CaseNumber = row.GetCaseNumber()
	out.CaseType = row.GetCaseType()
	out.Status = row.GetStatus()
	out.SubjectWorkerID = row.SubjectWorkerId
	out.ReportingPartyReferenceID = row.ReportingPartyReferenceId
	out.AssignedInvestigatorPrincipalID = row.AssignedInvestigatorPrincipalId
	out.OpenedAt = pbTimePtr(row.GetOpenedAt())
	out.DueAt = pbTimePtr(row.GetDueAt())
	out.ResolvedAt = pbTimePtr(row.GetResolvedAt())
	out.OutcomeCode = row.OutcomeCode
	out.RestrictedSummaryReferenceID = row.GetRestrictedSummaryReferenceId()
	out.ConfidentialityClass = row.GetConfidentialityClass()
	out.LegalHoldReferenceID = row.LegalHoldReferenceId
	out.IdempotencyKey = row.GetIdempotencyKey()
	return out, nil
}

// EmployeeRelationsCase membaca detail employee-relations-case berdasarkan ID.
func (a *Adapter) EmployeeRelationsCase(ctx context.Context, rawID string) (*model.EmployeeRelationsCase, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_employee_relations_cases", "human_resource.employee_relations_cases.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "employee-relations-case", "employee-relations-case")
	if err != nil {
		return nil, err
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.employeeRelations.GetEmployeeRelationsCaseByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "employee-relations-case")
	}
	return toModelEmployeeRelationsCase(resp.GetValue())
}

// EmployeeRelationsCases mendaftar dan memfilter employee-relations-case dengan paginasi.
func (a *Adapter) EmployeeRelationsCases(ctx context.Context, limit, offset int, f_case_number *string, f_case_type *string, f_status *string, f_subject_worker_id *string, f_reporting_party_reference_id *string, f_assigned_investigator_principal_id *string, f_outcome_code *string, f_restricted_summary_reference_id *string, f_legal_hold_reference_id *string, f_idempotency_key *string) (*Page[model.EmployeeRelationsCase], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_employee_relations_cases", "human_resource.employee_relations_cases.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid employee-relations-case pagination", nil)
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_case_number != nil, f_case_type != nil, f_status != nil, f_subject_worker_id != nil, f_reporting_party_reference_id != nil, f_assigned_investigator_principal_id != nil, f_outcome_code != nil, f_restricted_summary_reference_id != nil, f_legal_hold_reference_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *employeerelations.EmployeeRelationsCasesResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_case_number, "case_number")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsCaseByCaseNumber(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_case_type, "case_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsCaseByCaseType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_status, "status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsCaseByStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_subject_worker_id, "subject_worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsCaseBySubjectWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_reporting_party_reference_id, "reporting_party_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsCaseByReportingPartyReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_assigned_investigator_principal_id, "assigned_investigator_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsCaseByAssignedInvestigatorPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_outcome_code, "outcome_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsCaseByOutcomeCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 7:
		v, fErr := validateFilterString(f_restricted_summary_reference_id, "restricted_summary_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsCaseByRestrictedSummaryReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 8:
		v, fErr := validateFilterString(f_legal_hold_reference_id, "legal_hold_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsCaseByLegalHoldReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 9:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsCaseByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.employeeRelations.GetAllEmployeeRelationsCases(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "employee-relations-case")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "employee-relations-case", toModelEmployeeRelationsCase)
}

// CreateEmployeeRelationsCase membuat employee-relations-case baru melalui HRM.
func (a *Adapter) CreateEmployeeRelationsCase(ctx context.Context, input model.CreateEmployeeRelationsCaseInput) (*model.EmployeeRelationsCase, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_employee_relations_cases", "create", "human_resource.employee_relations_cases.create")
	if err != nil {
		return nil, err
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &employeerelations.EmployeeRelationsCaseInput{}
	patch.CaseNumber = input.CaseNumber
	patch.CaseType = input.CaseType
	patch.Status = input.Status
	patch.SubjectWorkerId = input.SubjectWorkerID
	patch.ReportingPartyReferenceId = input.ReportingPartyReferenceID
	patch.AssignedInvestigatorPrincipalId = input.AssignedInvestigatorPrincipalID
	patch.DueAt = optionalTimestamp(input.DueAt)
	patch.ResolvedAt = optionalTimestamp(input.ResolvedAt)
	patch.OutcomeCode = input.OutcomeCode
	patch.RestrictedSummaryReferenceId = input.RestrictedSummaryReferenceID
	patch.ConfidentialityClass = input.ConfidentialityClass
	patch.LegalHoldReferenceId = input.LegalHoldReferenceID
	patch.IdempotencyKey = input.IdempotencyKey
	resp, err := a.employeeRelations.CreateEmployeeRelationsCase(rpcCtx, &employeerelations.CreateEmployeeRelationsCaseRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "employee-relations-case")
	}
	return toModelEmployeeRelationsCase(resp.GetValue())
}

// UpdateEmployeeRelationsCase memperbarui employee-relations-case berdasarkan FieldMask.
func (a *Adapter) UpdateEmployeeRelationsCase(ctx context.Context, rawID string, input model.UpdateEmployeeRelationsCaseInput) (*model.EmployeeRelationsCase, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_employee_relations_cases", "update", "human_resource.employee_relations_cases.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "employee-relations-case", "employee-relations-case")
	if err != nil {
		return nil, err
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &employeerelations.EmployeeRelationsCaseInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "case_number", input.CaseNumber, &patch.CaseNumber, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "case_type", input.CaseType, &patch.CaseType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "status", input.Status, &patch.Status, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "subject_worker_id", input.SubjectWorkerID, &patch.SubjectWorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "reporting_party_reference_id", input.ReportingPartyReferenceID, &patch.ReportingPartyReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "assigned_investigator_principal_id", input.AssignedInvestigatorPrincipalID, &patch.AssignedInvestigatorPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "due_at", input.DueAt, &patch.DueAt, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "resolved_at", input.ResolvedAt, &patch.ResolvedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "outcome_code", input.OutcomeCode, &patch.OutcomeCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "restricted_summary_reference_id", input.RestrictedSummaryReferenceID, &patch.RestrictedSummaryReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "confidentiality_class", input.ConfidentialityClass, &patch.ConfidentialityClass, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "legal_hold_reference_id", input.LegalHoldReferenceID, &patch.LegalHoldReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "idempotency_key", input.IdempotencyKey, &patch.IdempotencyKey, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("employee-relations-case update is empty", nil)
	}
	resp, err := a.employeeRelations.UpdateEmployeeRelationsCase(rpcCtx, &employeerelations.UpdateEmployeeRelationsCaseRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "employee-relations-case")
	}
	return toModelEmployeeRelationsCase(resp.GetValue())
}

// DeleteEmployeeRelationsCase menghapus employee-relations-case berdasarkan ID.
func (a *Adapter) DeleteEmployeeRelationsCase(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_employee_relations_cases", "human_resource.employee_relations_cases.delete", "employee-relations-case", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.employeeRelations.DeleteEmployeeRelationsCase(c, req)
	})
}

func toModelEmployeeRelationsEvent(row *employeerelations.EmployeeRelationsEvent) (*model.EmployeeRelationsEvent, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource employee-relations-event response", nil)
	}
	if row.GetId() <= 0 {
		return nil, middleware.ErrUnavailable("invalid human resource employee-relations-event ID", nil)
	}
	rawID := strconv.FormatInt(row.GetId(), 10)
	out := &model.EmployeeRelationsEvent{ID: int64NodeID("employee-relations-event", row.GetId()), RawID: rawID}
	out.CaseID = row.GetCaseId()
	out.EventType = row.GetEventType()
	out.OccurredAt = pbTimePtr(row.GetOccurredAt())
	out.ActorPrincipalID = row.ActorPrincipalId
	out.EventDigest = pbBase64StringValue(row.GetEventDigest())
	out.EventReferenceID = row.EventReferenceId
	return out, nil
}

// EmployeeRelationsEvent membaca detail employee-relations-event berdasarkan ID.
func (a *Adapter) EmployeeRelationsEvent(ctx context.Context, rawID string) (*model.EmployeeRelationsEvent, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_employee_relations_events", "human_resource.employee_relations_events.read")
	if err != nil {
		return nil, err
	}
	id, err := parsePositiveInt64NodeID(rawID, "employee-relations-event")
	if err != nil {
		return nil, err
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.employeeRelations.GetEmployeeRelationsEventByID(rpcCtx, &common.IdentityIDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "employee-relations-event")
	}
	return toModelEmployeeRelationsEvent(resp.GetValue())
}

// EmployeeRelationsEvents mendaftar dan memfilter employee-relations-event dengan paginasi.
func (a *Adapter) EmployeeRelationsEvents(ctx context.Context, limit, offset int, f_case_id *string, f_event_type *string, f_actor_principal_id *string, f_event_reference_id *string) (*Page[model.EmployeeRelationsEvent], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_employee_relations_events", "human_resource.employee_relations_events.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid employee-relations-event pagination", nil)
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_case_id != nil, f_event_type != nil, f_actor_principal_id != nil, f_event_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *employeerelations.EmployeeRelationsEventsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_case_id, "case_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsEventByCaseID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_event_type, "event_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsEventByEventType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_actor_principal_id, "actor_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsEventByActorPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_event_reference_id, "event_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindEmployeeRelationsEventByEventReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.employeeRelations.GetAllEmployeeRelationsEvents(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "employee-relations-event")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "employee-relations-event", toModelEmployeeRelationsEvent)
}

// CreateEmployeeRelationsEvent membuat employee-relations-event baru melalui HRM.
func (a *Adapter) CreateEmployeeRelationsEvent(ctx context.Context, input model.CreateEmployeeRelationsEventInput) (*model.EmployeeRelationsEvent, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_employee_relations_events", "create", "human_resource.employee_relations_events.create")
	if err != nil {
		return nil, err
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &employeerelations.EmployeeRelationsEventInput{}
	patch.CaseId = input.CaseID
	patch.EventType = input.EventType
	if b, bErr := optionalBase64Bytes(input.EventDigest, "event_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.EventDigest = b
	}
	patch.EventReferenceId = input.EventReferenceID
	resp, err := a.employeeRelations.CreateEmployeeRelationsEvent(rpcCtx, &employeerelations.CreateEmployeeRelationsEventRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "employee-relations-event")
	}
	return toModelEmployeeRelationsEvent(resp.GetValue())
}

// UpdateEmployeeRelationsEvent memperbarui employee-relations-event berdasarkan FieldMask.
func (a *Adapter) UpdateEmployeeRelationsEvent(ctx context.Context, rawID string, input model.UpdateEmployeeRelationsEventInput) (*model.EmployeeRelationsEvent, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_employee_relations_events", "update", "human_resource.employee_relations_events.update")
	if err != nil {
		return nil, err
	}
	id, err := parsePositiveInt64NodeID(rawID, "employee-relations-event")
	if err != nil {
		return nil, err
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &employeerelations.EmployeeRelationsEventInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "case_id", input.CaseID, &patch.CaseId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "event_type", input.EventType, &patch.EventType, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "event_digest", input.EventDigest, &patch.EventDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "event_reference_id", input.EventReferenceID, &patch.EventReferenceId, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("employee-relations-event update is empty", nil)
	}
	resp, err := a.employeeRelations.UpdateEmployeeRelationsEvent(rpcCtx, &employeerelations.UpdateEmployeeRelationsEventRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "employee-relations-event")
	}
	return toModelEmployeeRelationsEvent(resp.GetValue())
}

// DeleteEmployeeRelationsEvent menghapus employee-relations-event berdasarkan ID.
func (a *Adapter) DeleteEmployeeRelationsEvent(ctx context.Context, rawID string) (*bool, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_employee_relations_events", "delete", "human_resource.employee_relations_events.delete")
	if err != nil {
		return nil, err
	}
	id, err := parsePositiveInt64NodeID(rawID, "employee-relations-event")
	if err != nil {
		return nil, err
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	if _, err := a.employeeRelations.DeleteEmployeeRelationsEvent(rpcCtx, &common.IdentityIDRequest{Id: id}); err != nil {
		return nil, mapHRMWriteError(err, "employee-relations-event")
	}
	ok := true
	return &ok, nil
}

func toModelCollectiveAgreement(row *employeerelations.CollectiveAgreement) (*model.CollectiveAgreement, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource collective-agreement response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource collective-agreement ID", nil)
	}
	out := &model.CollectiveAgreement{ID: nodeID("collective-agreement", row.GetId()), RawID: row.GetId()}
	out.AgreementCode = row.GetAgreementCode()
	out.AgreementVersion = int(row.GetAgreementVersion())
	out.LegalEntityReferenceID = row.GetLegalEntityReferenceId()
	out.RepresentativePartyReferenceID = row.GetRepresentativePartyReferenceId()
	out.JurisdictionCode = row.GetJurisdictionCode()
	out.EffectiveFrom = pbTimePtr(row.GetEffectiveFrom())
	out.EffectiveTo = pbTimePtr(row.GetEffectiveTo())
	out.SignedAt = pbTimePtr(row.GetSignedAt())
	out.DocumentVersionReferenceID = row.GetDocumentVersionReferenceId()
	out.ContentDigest = pbBase64StringValue(row.GetContentDigest())
	return out, nil
}

// CollectiveAgreement membaca detail collective-agreement berdasarkan ID.
func (a *Adapter) CollectiveAgreement(ctx context.Context, rawID string) (*model.CollectiveAgreement, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_collective_agreements", "human_resource.collective_agreements.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "collective-agreement", "collective-agreement")
	if err != nil {
		return nil, err
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.employeeRelations.GetCollectiveAgreementByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "collective-agreement")
	}
	return toModelCollectiveAgreement(resp.GetValue())
}

// CollectiveAgreements mendaftar dan memfilter collective-agreement dengan paginasi.
func (a *Adapter) CollectiveAgreements(ctx context.Context, limit, offset int, f_agreement_code *string, f_legal_entity_reference_id *string, f_representative_party_reference_id *string, f_jurisdiction_code *string, f_document_version_reference_id *string) (*Page[model.CollectiveAgreement], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_collective_agreements", "human_resource.collective_agreements.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid collective-agreement pagination", nil)
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_agreement_code != nil, f_legal_entity_reference_id != nil, f_representative_party_reference_id != nil, f_jurisdiction_code != nil, f_document_version_reference_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *employeerelations.CollectiveAgreementsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_agreement_code, "agreement_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindCollectiveAgreementByAgreementCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_legal_entity_reference_id, "legal_entity_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindCollectiveAgreementByLegalEntityReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_representative_party_reference_id, "representative_party_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindCollectiveAgreementByRepresentativePartyReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_jurisdiction_code, "jurisdiction_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindCollectiveAgreementByJurisdictionCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_document_version_reference_id, "document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.employeeRelations.FindCollectiveAgreementByDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.employeeRelations.GetAllCollectiveAgreements(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "collective-agreement")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "collective-agreement", toModelCollectiveAgreement)
}

// CreateCollectiveAgreement membuat collective-agreement baru melalui HRM.
func (a *Adapter) CreateCollectiveAgreement(ctx context.Context, input model.CreateCollectiveAgreementInput) (*model.CollectiveAgreement, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_collective_agreements", "create", "human_resource.collective_agreements.create")
	if err != nil {
		return nil, err
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &employeerelations.CollectiveAgreementInput{}
	patch.AgreementCode = input.AgreementCode
	patch.AgreementVersion = optionalInt32Ptr(input.AgreementVersion)
	patch.LegalEntityReferenceId = input.LegalEntityReferenceID
	patch.RepresentativePartyReferenceId = input.RepresentativePartyReferenceID
	patch.JurisdictionCode = input.JurisdictionCode
	patch.EffectiveFrom = optionalTimestamp(input.EffectiveFrom)
	patch.EffectiveTo = optionalTimestamp(input.EffectiveTo)
	patch.SignedAt = optionalTimestamp(input.SignedAt)
	patch.DocumentVersionReferenceId = input.DocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.ContentDigest, "content_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.ContentDigest = b
	}
	resp, err := a.employeeRelations.CreateCollectiveAgreement(rpcCtx, &employeerelations.CreateCollectiveAgreementRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "collective-agreement")
	}
	return toModelCollectiveAgreement(resp.GetValue())
}

// UpdateCollectiveAgreement memperbarui collective-agreement berdasarkan FieldMask.
func (a *Adapter) UpdateCollectiveAgreement(ctx context.Context, rawID string, input model.UpdateCollectiveAgreementInput) (*model.CollectiveAgreement, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_collective_agreements", "update", "human_resource.collective_agreements.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "collective-agreement", "collective-agreement")
	if err != nil {
		return nil, err
	}
	if a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &employeerelations.CollectiveAgreementInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "agreement_code", input.AgreementCode, &patch.AgreementCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "agreement_version", input.AgreementVersion, &patch.AgreementVersion, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "legal_entity_reference_id", input.LegalEntityReferenceID, &patch.LegalEntityReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "representative_party_reference_id", input.RepresentativePartyReferenceID, &patch.RepresentativePartyReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "jurisdiction_code", input.JurisdictionCode, &patch.JurisdictionCode, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "effective_from", input.EffectiveFrom, &patch.EffectiveFrom, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "effective_to", input.EffectiveTo, &patch.EffectiveTo, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "signed_at", input.SignedAt, &patch.SignedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "document_version_reference_id", input.DocumentVersionReferenceID, &patch.DocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "content_digest", input.ContentDigest, &patch.ContentDigest, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("collective-agreement update is empty", nil)
	}
	resp, err := a.employeeRelations.UpdateCollectiveAgreement(rpcCtx, &employeerelations.UpdateCollectiveAgreementRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "collective-agreement")
	}
	return toModelCollectiveAgreement(resp.GetValue())
}

// DeleteCollectiveAgreement menghapus collective-agreement berdasarkan ID.
func (a *Adapter) DeleteCollectiveAgreement(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.employeeRelations == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_collective_agreements", "human_resource.collective_agreements.delete", "collective-agreement", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.employeeRelations.DeleteCollectiveAgreement(c, req)
	})
}
