package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type compensationBandStub struct {
	rows []entity.CompensationBand
	err  error
}

func (s *compensationBandStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.CompensationBand, error) {
	return s.rows, s.err
}

type compensationRecordStub struct {
	count int64
	err   error
}

func (s *compensationRecordStub) CountWhere(context.Context, []domainrepo.Filter) (int64, error) {
	return s.count, s.err
}

func TestCompensationBandValidator(t *testing.T) {
	t.Parallel()
	bands := &compensationBandStub{}
	records := &compensationRecordStub{}
	validator := &compensationBandValidator{bands: bands, records: records}
	band := &entity.CompensationBand{ID: "band-1", BandCode: "STORE-L2", BandVersion: 1, JurisdictionCode: "ID"}
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_bands", nil, band); err != nil {
		t.Fatalf("create band: %v", err)
	}
	approvalID := "approval-1"
	approved := *band
	approved.ApprovalReferenceID = &approvalID
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_bands", nil, &approved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified approval: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_bands", &approved, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete approved band: %v", err)
	}
	bands.rows = []entity.CompensationBand{{ID: "band-2"}}
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_bands", nil, band); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate version: %v", err)
	}
	bands.rows = nil
	records.count = 1
	changed := *band
	changed.JobLevelCode = "L3"
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_bands", band, &changed); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("change referenced band: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_bands", band, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("delete referenced band: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_bands", band, band); err != nil {
		t.Fatalf("no-op update: %v", err)
	}
	records.err = hrerrors.ErrUnavailable
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_bands", band, nil); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("counter error: %v", err)
	}
	records.err = nil
	bands.err = hrerrors.ErrUnavailable
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_bands", nil, band); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("lookup error: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := validator.ValidateMutation(ctx, "hr_compensation_bands", nil, band); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context: %v", err)
	}
}

type compensationWorkerStub struct{ value *entity.Worker }

func (s compensationWorkerStub) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	return s.value, nil
}

type compensationAssignmentStub struct{ value *entity.WorkerAssignment }

func (s compensationAssignmentStub) FindByIDForUpdate(context.Context, string) (*entity.WorkerAssignment, error) {
	return s.value, nil
}

type compensationPositionStub struct{ value *entity.Position }

func (s compensationPositionStub) FindByIDForUpdate(context.Context, string) (*entity.Position, error) {
	return s.value, nil
}

type compensationJobStub struct{ value *entity.Job }

func (s compensationJobStub) FindByIDForUpdate(context.Context, string) (*entity.Job, error) {
	return s.value, nil
}

type compensationComponentStub struct{ value *entity.PayComponent }

func (s compensationComponentStub) FindByIDForUpdate(context.Context, string) (*entity.PayComponent, error) {
	return s.value, nil
}

type compensationBandLockerStub struct{ value *entity.CompensationBand }

func (s compensationBandLockerStub) FindByIDForUpdate(context.Context, string) (*entity.CompensationBand, error) {
	return s.value, nil
}

func TestCompensationRecordValidatorRequiresMatchingActiveParents(t *testing.T) {
	t.Parallel()
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	level, bandID, currency := "L2", "band-1", "IDR"
	worker := &entity.Worker{ID: "worker-1", Status: entity.WorkerStatusActive,
		HireDate: &from, PrimaryLegalEntityReferenceID: "legal-1"}
	assignment := &entity.WorkerAssignment{ID: "assignment-1", WorkerID: worker.ID,
		PositionID: "position-1", ValidFrom: from, ValidTo: &until}
	position := &entity.Position{ID: assignment.PositionID, JobID: "job-1",
		LegalEntityReferenceID: worker.PrimaryLegalEntityReferenceID, ValidFrom: from, ValidTo: &until}
	job := &entity.Job{ID: position.JobID, JobLevelCode: &level, ValidFrom: from, ValidTo: &until, IsActive: true}
	component := &entity.PayComponent{ID: "component-1", ComponentCode: "BASE", ComponentType: "earning",
		CalculationMethod: "fixed", CurrencyCode: &currency, JurisdictionCode: "ID", EffectiveFrom: from, EffectiveTo: &until, IsActive: true}
	band := &entity.CompensationBand{ID: bandID, JobLevelCode: level, JurisdictionCode: "ID",
		CurrencyCode: currency, PayFrequency: "monthly", MinimumAmount: "5000000", MaximumAmount: "7000000",
		ValidFrom: from, ValidTo: &until}
	validator := &compensationRecordValidator{workers: compensationWorkerStub{worker},
		assignments: compensationAssignmentStub{assignment}, positions: compensationPositionStub{position},
		jobs: compensationJobStub{job}, components: compensationComponentStub{component}, bands: compensationBandLockerStub{band}}
	record := &entity.CompensationRecord{ID: "record-1", WorkerID: worker.ID, AssignmentID: assignment.ID,
		ComponentID: component.ID, ComponentCodeSnapshot: component.ComponentCode, CompensationBandID: &bandID,
		CompensationType: "base_salary", Amount: "6000000", CurrencyCode: currency, PayFrequency: "monthly",
		EffectiveFrom: from, EffectiveTo: &until, Status: entity.CompensationStatusProposed, AccessClass: "restricted"}
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", nil, record); err != nil {
		t.Fatalf("valid proposal: %v", err)
	}
	changed := *record
	changed.Amount = "6500000"
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", record, &changed); err != nil {
		t.Fatalf("update proposal amount: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", record, nil); err != nil {
		t.Fatalf("delete proposal: %v", err)
	}
	approved := *record
	approved.Status = entity.CompensationStatusApproved
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", nil, &approved); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("approval through CRUD: %v", err)
	}
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", &approved, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete approved record: %v", err)
	}
	sourceID := "change-1"
	unsourced := *record
	unsourced.SourceChangeReferenceID = &sourceID
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", nil, &unsourced); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("unverified source: %v", err)
	}
	moved := *record
	moved.AssignmentID = "assignment-2"
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", record, &moved); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("move proposal assignment: %v", err)
	}
	worker.Status = entity.WorkerStatusTerminated
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", nil, record); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive worker: %v", err)
	}
	worker.Status = entity.WorkerStatusActive
	assignment.WorkerID = "worker-2"
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", nil, record); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("assignment owner mismatch: %v", err)
	}
	assignment.WorkerID = worker.ID
	component.ComponentCode = "OTHER"
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", nil, record); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("component snapshot mismatch: %v", err)
	}
	component.ComponentCode = record.ComponentCodeSnapshot
	tooHigh := *record
	tooHigh.Amount = "7000001"
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", nil, &tooHigh); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("amount above band: %v", err)
	}
	outside := *record
	outside.EffectiveFrom = time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	if err := validator.ValidateMutation(t.Context(), "hr_compensation_records", nil, &outside); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("period outside assignment: %v", err)
	}
}

func TestCompensationReferenceError(t *testing.T) {
	t.Parallel()
	if err := compensationReferenceError(hrerrors.ErrNotFound); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("missing parent error: %v", err)
	}
	if err := compensationReferenceError(hrerrors.ErrUnavailable); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("dependency error: %v", err)
	}
}
