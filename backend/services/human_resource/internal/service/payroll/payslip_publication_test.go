package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	repo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type movedPayslipResult struct {
	repo.RecordRepository[entity.PayrollResult, string]
	listed  []entity.PayrollResult
	current entity.PayrollResult
}

func (r movedPayslipResult) CountWhere(context.Context, []repo.Filter) (int64, error) {
	return int64(len(r.listed)), nil
}

func (r movedPayslipResult) SelectWhere(context.Context, []repo.Filter, int, int) ([]entity.PayrollResult, error) {
	return r.listed, nil
}

func (r movedPayslipResult) SelectByID(context.Context, string) (*entity.PayrollResult, error) {
	return &r.current, nil
}

type payslipPublicationTransaction struct{}

func (payslipPublicationTransaction) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type untouchedPayslipProcessor struct{ PayslipPublicationProcessor }

func TestPublishPayslipsKeepsEventRunAfterListing(t *testing.T) {
	const publication, expectedRun, otherRun, resultID = "00000000-0000-4000-8000-000000000001",
		"00000000-0000-4000-8000-000000000002", "00000000-0000-4000-8000-000000000003",
		"00000000-0000-4000-8000-000000000004"
	s := &Service{dependencies: Dependencies{
		Runtime:            operation.Dependencies{Transactor: payslipPublicationTransaction{}},
		PayslipPublication: untouchedPayslipProcessor{},
		PayrollResult: movedPayslipResult{
			listed:  []entity.PayrollResult{{ID: resultID, PayrollRunID: expectedRun}},
			current: entity.PayrollResult{ID: resultID, PayrollRunID: otherRun},
		},
	}}
	if err := s.PublishPayslipsForRun(t.Context(), publication, expectedRun); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("publish after result no longer belongs to event run: %v", err)
	}
}

func TestPublishPayslipsReportsEachResultFailure(t *testing.T) {
	const publication, eventRun, otherRun, anomalous, moved = "00000000-0000-4000-8000-000000000001",
		"00000000-0000-4000-8000-000000000002", "00000000-0000-4000-8000-000000000003",
		"00000000-0000-4000-8000-000000000004", "00000000-0000-4000-8000-000000000005"
	s := &Service{dependencies: Dependencies{
		Runtime:            operation.Dependencies{Transactor: payslipPublicationTransaction{}},
		PayslipPublication: untouchedPayslipProcessor{},
		PayrollResult: movedPayslipResult{
			listed: []entity.PayrollResult{
				{ID: anomalous, PayrollRunID: otherRun},
				{ID: moved, PayrollRunID: eventRun},
			},
			current: entity.PayrollResult{ID: moved, PayrollRunID: otherRun},
		},
	}}
	err := s.PublishPayslipsForRun(t.Context(), publication, eventRun)
	if !errors.Is(err, hrerrors.ErrReferenceConflict) ||
		!strings.Contains(err.Error(), anomalous) || !strings.Contains(err.Error(), moved) {
		t.Fatalf("expected both payroll result failures: %v", err)
	}
}
