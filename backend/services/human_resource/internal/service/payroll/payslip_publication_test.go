package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

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

type publicationReplayResult struct {
	movedPayslipResult
}

func (r publicationReplayResult) FindByIDForUpdate(context.Context, string) (*entity.PayrollResult, error) {
	return &r.current, nil
}

type publicationReplayRun struct {
	repo.RecordRepository[entity.PayrollRun, string]
	run entity.PayrollRun
}

func (r publicationReplayRun) FindByIDForUpdate(context.Context, string) (*entity.PayrollRun, error) {
	return &r.run, nil
}

type publicationReplayProcessor struct {
	PayslipPublicationProcessor
	publishedAt time.Time
}

func (p publicationReplayProcessor) VerifyPayrollPublication(context.Context, PayslipPublicationCheck) (time.Time, error) {
	return p.publishedAt, nil
}

func TestPublishPayslipReplayRequiresValidStoredVersion(t *testing.T) {
	const publication = "00000000-0000-4000-8000-000000000001"
	const runID = "00000000-0000-4000-8000-000000000002"
	const resultID = "00000000-0000-4000-8000-000000000003"
	const invalidVersion = "00000000-0000-0000-0000-000000000000"
	const validVersion = "00000000-0000-4000-8000-000000000004"
	approvedAt := time.Now().Add(-time.Hour)
	for _, version := range []string{invalidVersion, validVersion} {
		t.Run(version, func(t *testing.T) {
			s := Service{dependencies: Dependencies{
				Runtime: operation.Dependencies{Transactor: payslipPublicationTransaction{}, Clock: operation.SystemClock{}},
				PayrollRun: publicationReplayRun{run: entity.PayrollRun{
					ID: runID, Status: entity.PayrollRunStatusPosted, FinanceJournalReferenceID: new(runID), ApprovedAt: &approvedAt,
					InputDigest: make([]byte, sha256.Size), OutputDigest: make([]byte, sha256.Size), InputSnapshotReferenceID: runID,
				}},
				PayrollResult: publicationReplayResult{movedPayslipResult{current: entity.PayrollResult{
					ID: resultID, PayrollRunID: runID, ResultStatus: "posted", CalculationDigest: make([]byte, sha256.Size),
					PayslipDocumentReferenceID: &version, PayslipPublicationReferenceID: new(publication),
				}}},
				PayslipPublication: publicationReplayProcessor{publishedAt: approvedAt},
			}}
			got, err := s.PublishPayslipForResult(t.Context(), publication, resultID)
			if version == invalidVersion {
				if got != "" || !errors.Is(err, hrerrors.ErrReferenceConflict) {
					t.Fatalf("invalid version replay: version %q error %v", got, err)
				}
			} else if got != version || err != nil {
				t.Fatalf("valid version replay: version %q error %v", got, err)
			}
		})
	}
}
