package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

type contractWorkerStub struct {
	worker *entity.Worker
	err    error
}

func (s *contractWorkerStub) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	return s.worker, s.err
}

type contractNumberStub struct {
	matches []entity.EmploymentContract
	err     error
}

func (s *contractNumberStub) FindByContractNumber(context.Context, string, int, int) ([]entity.EmploymentContract, error) {
	return s.matches, s.err
}

func TestEmploymentContractValidatorDraft(t *testing.T) {
	t.Parallel()
	worker := &contractWorkerStub{worker: &entity.Worker{
		ID: "worker-1", Status: entity.WorkerStatusPreboarding,
		PrimaryLegalEntityReferenceID: "legal-1",
	}}
	contracts := &contractNumberStub{}
	validator := &employmentContractValidator{workers: worker, contracts: contracts}
	draft := &entity.EmploymentContract{
		ID: "contract-1", WorkerID: "worker-1", ContractNumber: "CON-1",
		ContractType: "fixed_term", ContractStatus: "draft",
		EmployingLegalEntityReferenceID: "legal-1", CreatedAt: time.Now(),
	}
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", nil, draft); err != nil {
		t.Fatalf("create draft: %v", err)
	}
	updated := *draft
	updated.ContractType = "permanent"
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", draft, &updated); err != nil {
		t.Fatalf("update draft: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", draft, nil); err != nil {
		t.Fatalf("delete draft: %v", err)
	}
	worker.worker.Status = entity.WorkerStatusTerminated
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("create for terminated worker = %v", err)
	}
	worker.worker.Status = entity.WorkerStatusActive
	worker.worker.PrimaryLegalEntityReferenceID = "legal-2"
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("wrong employer = %v", err)
	}
	worker.worker.PrimaryLegalEntityReferenceID = "legal-1"
	terminationDate := time.Now()
	worker.worker.TerminationDate = &terminationDate
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("create with active status but recorded termination: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", draft, &updated); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("update with active status but recorded termination: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", draft, nil); err != nil {
		t.Fatalf("remove unused draft after termination: %v", err)
	}
	worker.worker.TerminationDate = nil
	contracts.matches = []entity.EmploymentContract{{ID: "another-contract"}}
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", nil, draft); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate number = %v", err)
	}
	contracts.matches = nil
	worker.err = hrerrors.ErrNotFound
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("missing worker = %v", err)
	}
}

func TestEmploymentContractValidatorRejectsWorkflowFields(t *testing.T) {
	t.Parallel()
	validator := &employmentContractValidator{
		workers:   &contractWorkerStub{worker: &entity.Worker{Status: entity.WorkerStatusActive, PrimaryLegalEntityReferenceID: "legal-1"}},
		contracts: &contractNumberStub{},
	}
	draft := &entity.EmploymentContract{
		ID: "contract-1", WorkerID: "worker-1", ContractNumber: "CON-1",
		ContractType: "permanent", ContractStatus: "draft",
		EmployingLegalEntityReferenceID: "legal-1", CreatedAt: time.Now(),
	}
	for _, tc := range []struct {
		name   string
		mutate func(*entity.EmploymentContract)
		want   error
	}{
		{"signed status", func(c *entity.EmploymentContract) { c.ContractStatus = "signed" }, hrerrors.ErrWorkflowRequired},
		{"signed timestamp", func(c *entity.EmploymentContract) { now := time.Now(); c.SignedAt = &now }, hrerrors.ErrWorkflowRequired},
		{"document reference", func(c *entity.EmploymentContract) { id := "document"; c.SignedDocumentReferenceID = &id }, hrerrors.ErrWorkflowRequired},
		{"content digest", func(c *entity.EmploymentContract) { c.SignedContentDigest = []byte("digest") }, hrerrors.ErrWorkflowRequired},
		{"empty content digest", func(c *entity.EmploymentContract) { c.SignedContentDigest = []byte{} }, hrerrors.ErrWorkflowRequired},
		{"unsupported type", func(c *entity.EmploymentContract) { c.ContractType = "unknown" }, hrerrors.ErrWorkflowRequired},
		{"unverified agreement", func(c *entity.EmploymentContract) { id := "agreement"; c.CollectiveAgreementID = &id }, hrerrors.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := *draft
			tc.mutate(&candidate)
			if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", nil, &candidate); !errors.Is(err, tc.want) {
				t.Fatalf("create = %v, want %v", err, tc.want)
			}
		})
	}
	changedWorker := *draft
	changedWorker.WorkerID = "worker-2"
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", draft, &changedWorker); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("change worker = %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", &changedWorker, draft); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("change worker back = %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := validator.ValidateMutation(ctx, "hr_employment_contracts", nil, draft); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context = %v", err)
	}
}

type contractAgreementStub struct {
	agreement *entity.CollectiveAgreement
	err       error
}

func (s *contractAgreementStub) FindByIDForUpdate(context.Context, string) (*entity.CollectiveAgreement, error) {
	return s.agreement, s.err
}

func TestEmploymentContractValidatorCollectiveAgreement(t *testing.T) {
	t.Parallel()
	agreementID := "agreement-1"
	from := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)
	agreements := &contractAgreementStub{
		agreement: &entity.CollectiveAgreement{
			ID:                     agreementID,
			LegalEntityReferenceID: "legal-1",
			EffectiveFrom:          from,
			EffectiveTo:            &end,
		},
	}
	validator := &employmentContractValidator{
		workers:    &contractWorkerStub{worker: &entity.Worker{Status: entity.WorkerStatusActive, PrimaryLegalEntityReferenceID: "legal-1"}},
		contracts:  &contractNumberStub{},
		agreements: agreements,
	}
	draft := &entity.EmploymentContract{
		ID: "contract-1", WorkerID: "worker-1", ContractNumber: "CON-CBA-1",
		ContractType: "collective_agreement", ContractStatus: "draft",
		EmployingLegalEntityReferenceID: "legal-1",
		EffectiveFrom:                   time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
		CollectiveAgreementID:           &agreementID,
		CreatedAt:                       time.Now(),
	}
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", nil, draft); err != nil {
		t.Fatalf("valid local collective agreement rejected: %v", err)
	}
	agreements.agreement.LegalEntityReferenceID = "legal-2"
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("wrong legal entity agreement = %v", err)
	}
	agreements.agreement.LegalEntityReferenceID = "legal-1"
	draft.EffectiveFrom = time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := validator.ValidateMutation(t.Context(), "hr_employment_contracts", nil, draft); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("expired collective agreement = %v", err)
	}
}
