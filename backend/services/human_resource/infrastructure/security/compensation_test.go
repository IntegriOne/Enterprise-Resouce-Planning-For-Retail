package security

import (
	"context"
	"errors"
	"testing"

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
