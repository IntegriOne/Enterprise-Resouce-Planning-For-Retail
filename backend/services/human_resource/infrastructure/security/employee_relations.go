package security

import (
	"context"
	"reflect"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type employeeRelationsCaseReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.EmployeeRelationsCase, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.EmployeeRelationsCase, error)
}

type employeeRelationsEventCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type collectiveAgreementReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.CollectiveAgreement, error)
}

type collectiveAgreementContractCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type employeeRelationsCaseValidator struct {
	workers    lockedWorkerReader
	cases      employeeRelationsCaseReader
	events     employeeRelationsEventCounter
	foundation FoundationVerifier
}

type employeeRelationsEventValidator struct {
	cases employeeRelationsCaseReader
}

type collectiveAgreementValidator struct {
	agreements collectiveAgreementReader
	contracts  collectiveAgreementContractCounter
	foundation FoundationVerifier
}

// EmployeeRelationsValidators mendaftarkan validator integritas kasus hubungan kerja dan perjanjian kerja bersama.
func EmployeeRelationsValidators(repos *repository.Registry, foundation FoundationVerifier) (map[string]MutationValidator, error) {
	if repos == nil ||
		operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.EmployeeRelationsCase) != nil ||
		operation.RequireDependency(repos.EmployeeRelationsEvent) != nil ||
		operation.RequireDependency(repos.CollectiveAgreement) != nil ||
		operation.RequireDependency(repos.EmploymentContract) != nil ||
		operation.RequireDependency(foundation) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_employee_relations_cases": &employeeRelationsCaseValidator{
			workers:    repos.Worker,
			cases:      repos.EmployeeRelationsCase,
			events:     repos.EmployeeRelationsEvent,
			foundation: foundation,
		},
		"hr_employee_relations_events": &employeeRelationsEventValidator{
			cases: repos.EmployeeRelationsCase,
		},
		"hr_collective_agreements": &collectiveAgreementValidator{
			agreements: repos.CollectiveAgreement,
			contracts:  repos.EmploymentContract,
			foundation: foundation,
		},
	}, nil
}

// ValidateMutation menjaga draf, kerahasiaan, dan keunikan kasus hubungan industrial.
func (v *employeeRelationsCaseValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_employee_relations_cases" ||
		operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.cases) != nil ||
		operation.RequireDependency(v.events) != nil ||
		operation.RequireDependency(v.foundation) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.EmployeeRelationsCase
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.EmployeeRelationsCase)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status == entity.CaseStatusResolved ||
			previous.Status == entity.CaseStatusClosed ||
			previous.Status == entity.CaseStatusWithdrawn {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.Status != entity.CaseStatusOpen {
			return hrerrors.ErrWorkflowRequired
		}
		return v.requireNoEvents(ctx, previous.ID)
	}
	item, ok := after.(*entity.EmployeeRelationsCase)
	if !ok || item == nil || (previous != nil && previous.ID != item.ID) {
		return hrerrors.ErrInvalidInput
	}
	if !supportedERCaseType(item.CaseType) || !supportedERConfidentiality(item.ConfidentialityClass) {
		return hrerrors.ErrInvalidInput
	}
	if previous == nil {
		if item.Status != entity.CaseStatusOpen || item.ResolvedAt != nil || item.OutcomeCode != nil {
			return hrerrors.ErrWorkflowRequired
		}
	} else {
		if previous.Status != item.Status || !previous.OpenedAt.Equal(item.OpenedAt) ||
			item.ResolvedAt != nil || item.OutcomeCode != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if previous != nil && (!reflect.DeepEqual(previous.SubjectWorkerID, item.SubjectWorkerID) ||
		!reflect.DeepEqual(previous.ReportingPartyReferenceID, item.ReportingPartyReferenceID)) {
		if err := v.requireNoEvents(ctx, item.ID); err != nil {
			return err
		}
	}
	if item.DueAt != nil && item.DueAt.Before(item.OpenedAt) &&
		(previous == nil || previous.DueAt == nil || !item.DueAt.Equal(*previous.DueAt)) {
		return hrerrors.Invalid("due_at", "must not precede case opening")
	}
	if item.SubjectWorkerID != nil {
		worker, err := v.workers.FindByIDForUpdate(ctx, *item.SubjectWorkerID)
		if err != nil {
			return reviewReferenceError(err)
		}
		if worker == nil {
			return hrerrors.ErrReferenceConflict
		}
	}
	if item.ReportingPartyReferenceID != nil {
		if err := verifyFoundationParty(ctx, v.foundation, *item.ReportingPartyReferenceID, false); err != nil {
			return err
		}
	}
	byNumber, err := v.cases.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "case_number", Values: []any{item.CaseNumber}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range byNumber {
		if existing.ID != item.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	byIdem, err := v.cases.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "idempotency_key", Values: []any{item.IdempotencyKey}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range byIdem {
		if existing.ID != item.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

// requireNoEvents memeriksa riwayat ketika baris kasus sudah dikunci.
func (v *employeeRelationsCaseValidator) requireNoEvents(ctx context.Context, caseID string) error {
	count, err := v.events.CountWhere(ctx, []domainrepo.Filter{
		{Column: "case_id", Values: []any{caseID}},
	})
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

// ValidateMutation menjaga riwayat kejadian kasus tetap append-only dan terikat pada kasus aktif.
func (v *employeeRelationsEventValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_employee_relations_events" || operation.RequireDependency(v.cases) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if before != nil || after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	event, ok := after.(*entity.EmployeeRelationsEvent)
	if !ok || event == nil {
		return hrerrors.ErrInvalidInput
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return hrerrors.ErrUnauthenticated
	}
	if event.ActorPrincipalID != nil && !strings.EqualFold(caller.PrincipalID, *event.ActorPrincipalID) {
		return hrerrors.ErrForbidden
	}
	// Perubahan status kasus hanya dicatat lewat workflow yang berwenang.
	switch event.EventType {
	case "intake", "evidence_logged":
	default:
		return hrerrors.ErrWorkflowRequired
	}
	item, err := v.cases.FindByIDForUpdate(ctx, event.CaseID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if item == nil {
		return hrerrors.ErrReferenceConflict
	}
	if item.Status == entity.CaseStatusClosed || item.Status == entity.CaseStatusWithdrawn {
		return hrerrors.ErrWorkflowRequired
	}
	if event.OccurredAt.Before(item.OpenedAt) {
		return hrerrors.Invalid("occurred_at", "must not precede case opening")
	}
	return nil
}

// ValidateMutation memverifikasi entitas legal, pihak perwakilan, dan keunikan versi perjanjian kerja bersama.
func (v *collectiveAgreementValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_collective_agreements" ||
		operation.RequireDependency(v.agreements) != nil ||
		operation.RequireDependency(v.contracts) != nil ||
		operation.RequireDependency(v.foundation) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.CollectiveAgreement
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.CollectiveAgreement)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return v.rejectIfReferenced(ctx, previous.ID)
	}
	ag, ok := after.(*entity.CollectiveAgreement)
	if !ok || ag == nil || (previous != nil && previous.ID != ag.ID) {
		return hrerrors.ErrInvalidInput
	}
	if err := verifyFoundationLegalEntity(ctx, v.foundation, ag.LegalEntityReferenceID); err != nil {
		return err
	}
	if err := verifyFoundationParty(ctx, v.foundation, ag.RepresentativePartyReferenceID, false); err != nil {
		return err
	}
	matches, err := v.agreements.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "legal_entity_reference_id", Values: []any{ag.LegalEntityReferenceID}},
		{Column: "agreement_code", Values: []any{ag.AgreementCode}},
		{Column: "agreement_version", Values: []any{ag.AgreementVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != ag.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous != nil {
		return v.rejectIfReferenced(ctx, ag.ID)
	}
	return nil
}

func (v *collectiveAgreementValidator) rejectIfReferenced(ctx context.Context, id string) error {
	count, err := v.contracts.CountWhere(ctx, []domainrepo.Filter{
		{Column: "collective_agreement_id", Values: []any{id}},
	})
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func supportedERCaseType(value string) bool {
	switch value {
	case "grievance", "conduct", "investigation", "accommodation", "employee_support", "ethics_referral":
		return true
	default:
		return false
	}
}

func supportedERConfidentiality(value string) bool {
	switch value {
	case "restricted", "highly_restricted":
		return true
	default:
		return false
	}
}
