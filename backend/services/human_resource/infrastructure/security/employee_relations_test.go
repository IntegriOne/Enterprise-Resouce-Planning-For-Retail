package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type erCaseStub struct {
	byID map[string]entity.EmployeeRelationsCase
	rows []entity.EmployeeRelationsCase
}

func (s *erCaseStub) FindByIDForUpdate(_ context.Context, id string) (*entity.EmployeeRelationsCase, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *erCaseStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.EmployeeRelationsCase, error) {
	var out []entity.EmployeeRelationsCase
	for _, row := range s.rows {
		match := true
		for _, f := range filters {
			switch f.Column {
			case "case_number":
				if row.CaseNumber != f.Values[0] {
					match = false
				}
			case "idempotency_key":
				if row.IdempotencyKey != f.Values[0] {
					match = false
				}
			}
		}
		if match {
			out = append(out, row)
		}
	}
	return out, nil
}

type erEventCounterStub struct {
	count int64
}

func (s *erEventCounterStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, nil
}

type erCollectiveAgreementStub struct {
	rows []entity.CollectiveAgreement
}

func (s *erCollectiveAgreementStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.CollectiveAgreement, error) {
	var out []entity.CollectiveAgreement
	for _, row := range s.rows {
		match := true
		for _, f := range filters {
			switch f.Column {
			case "legal_entity_reference_id":
				if row.LegalEntityReferenceID != f.Values[0] {
					match = false
				}
			case "agreement_code":
				if row.AgreementCode != f.Values[0] {
					match = false
				}
			case "agreement_version":
				if row.AgreementVersion != f.Values[0] {
					match = false
				}
			}
		}
		if match {
			out = append(out, row)
		}
	}
	return out, nil
}

type erFoundationStub struct {
	leErr    error
	partyErr error
}

func (s *erFoundationStub) VerifyPositionReferences(context.Context, PositionReferenceCheck) error {
	return nil
}
func (s *erFoundationStub) VerifyWorkerReferences(context.Context, WorkerReferenceCheck) error {
	return nil
}
func (s *erFoundationStub) VerifyLegalEntity(context.Context, string) error { return s.leErr }
func (s *erFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return nil
}
func (s *erFoundationStub) VerifyLocation(context.Context, string, string) error { return nil }
func (s *erFoundationStub) VerifyParty(context.Context, string, bool) error      { return s.partyErr }

func TestEmployeeRelationsValidators(t *testing.T) {
	t.Parallel()
	leID := "00000000-0000-4000-8000-000000000201"
	partyID := "00000000-0000-4000-8000-000000000202"
	actorID := "00000000-0000-4000-8000-000000000203"
	workerID := "00000000-0000-4000-8000-000000000204"
	foundation := &erFoundationStub{}

	casesRepo := &erCaseStub{byID: map[string]entity.EmployeeRelationsCase{}}
	eventsCounter := &erEventCounterStub{}
	caseVal := &employeeRelationsCaseValidator{
		workers:    &lockedWorkerStub{worker: &entity.Worker{ID: workerID, Status: entity.WorkerStatusActive}, order: &[]string{}},
		cases:      casesRepo,
		events:     eventsCounter,
		foundation: foundation,
	}

	erCase := &entity.EmployeeRelationsCase{
		ID:                           "case-1",
		CaseNumber:                   "ER-001",
		CaseType:                     "grievance",
		Status:                       entity.CaseStatusOpen,
		SubjectWorkerID:              &workerID,
		ReportingPartyReferenceID:    &partyID,
		OpenedAt:                     time.Now().UTC(),
		RestrictedSummaryReferenceID: "sum-1",
		ConfidentialityClass:         "highly_restricted",
		IdempotencyKey:               "idem-er-1",
	}
	if err := caseVal.ValidateMutation(t.Context(), "hr_employee_relations_cases", nil, erCase); err != nil {
		t.Fatalf("create valid ER case: %v", err)
	}
	badClass := *erCase
	badClass.ConfidentialityClass = "public"
	if err := caseVal.ValidateMutation(t.Context(), "hr_employee_relations_cases", nil, &badClass); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("lower confidentiality must fail: %v", err)
	}
	eventsCounter.count = 1
	if err := caseVal.ValidateMutation(t.Context(), "hr_employee_relations_cases", erCase, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete referenced case must fail: %v", err)
	}

	// Kejadian kasus harus append-only dan terikat pada kasus aktif.
	casesRepo.byID["case-1"] = *erCase
	evVal := &employeeRelationsEventValidator{cases: casesRepo}
	ctx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: actorID, Authenticated: true})
	event := &entity.EmployeeRelationsEvent{
		ID: 1, CaseID: "case-1", EventType: "intake", OccurredAt: time.Now().UTC(), ActorPrincipalID: &actorID, EventDigest: []byte("d"),
	}
	if err := evVal.ValidateMutation(ctx, "hr_employee_relations_events", nil, event); err != nil {
		t.Fatalf("create valid ER event: %v", err)
	}
	earlyEvent := *event
	earlyEvent.OccurredAt = erCase.OpenedAt.Add(-time.Second)
	if err := evVal.ValidateMutation(ctx, "hr_employee_relations_events", nil, &earlyEvent); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("record event before case opening: %v", err)
	}
	if err := evVal.ValidateMutation(ctx, "hr_employee_relations_events", event, event); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("update ER event must fail: %v", err)
	}

	// Perjanjian kerja bersama memverifikasi EF & blokir hapus bila dipakai kontrak.
	agRepo := &erCollectiveAgreementStub{}
	contractCounter := &erEventCounterStub{}
	agVal := &collectiveAgreementValidator{agreements: agRepo, contracts: contractCounter, foundation: foundation}
	ag := &entity.CollectiveAgreement{
		ID: "ag-1", AgreementCode: "CBA-27", AgreementVersion: 1, LegalEntityReferenceID: leID, RepresentativePartyReferenceID: partyID,
	}
	if err := agVal.ValidateMutation(t.Context(), "hr_collective_agreements", nil, ag); err != nil {
		t.Fatalf("create valid collective agreement: %v", err)
	}
	contractCounter.count = 1
	if err := agVal.ValidateMutation(t.Context(), "hr_collective_agreements", ag, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete referenced collective agreement must fail: %v", err)
	}
}
