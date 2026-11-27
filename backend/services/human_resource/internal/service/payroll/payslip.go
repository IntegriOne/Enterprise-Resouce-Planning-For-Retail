package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"slices"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/contract"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type PayslipWorkerReader interface {
	SelectByID(context.Context, string) (*entity.Worker, error)
	FindByIAMPrincipalReferenceID(context.Context, *string, int, int) ([]entity.Worker, error)
}

type PayslipCheck struct {
	ResultID, RunID, WorkerID, PrincipalID, DocumentVersionID, PublicationID string
}

type PublishedPayslip struct {
	PayrollPeriodID, PeriodCode                               string
	PeriodStart, PeriodEnd                                    time.Time
	ResultID, PaymentStatus, DocumentVersionID, PublicationID string
	PublishedAt, ScheduledPayDate                             time.Time
	PDF                                                       []byte
}

// PayslipVerifier memeriksa penerbitan eksplisit Finance dan hak akses dokumen privat DRM.
type PayslipVerifier interface {
	VerifyPublishedPayslip(context.Context, PayslipCheck) (PublishedPayslip, error)
}

type VerifiedPayslipPDF struct {
	DocumentVersionID string
	Content           []byte
	Digest            [sha256.Size]byte
}

// PayslipPDFReader membaca versi privat dari DRM dengan identitas dan referensi terverifikasi.
type PayslipPDFReader interface {
	ReadVerifiedPayslipPDF(context.Context, PayslipCheck, int) (VerifiedPayslipPDF, error)
}

// GetMyPayslip membaca metadata slip milik principal yang berasal dari sesi tepercaya.
func (s *Service) GetMyPayslip(ctx context.Context, resultID string) (PublishedPayslip, error) {
	return s.getMyPayslip(ctx, resultID, false)
}

// GetMyPayslipWithPDF membuka versi yang telah diverifikasi untuk pekerja pemiliknya.
func (s *Service) GetMyPayslipWithPDF(ctx context.Context, resultID string) (PublishedPayslip, error) {
	return s.getMyPayslip(ctx, resultID, true)
}

func (s *Service) getMyPayslip(ctx context.Context, resultID string, includePDF bool) (PublishedPayslip, error) {
	if s == nil || operation.RequireDependency(s.dependencies.Workers) != nil {
		return PublishedPayslip{}, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("payroll_result_id", resultID); err != nil {
		return PublishedPayslip{}, err
	}
	deps := s.dependencies
	actor, workerIDs, err := s.payslipOwner(ctx)
	if err != nil {
		return PublishedPayslip{}, err
	}
	results, err := s.selectOwnedPayslipResults(ctx, []any{resultID}, workerIDs)
	if err != nil {
		return PublishedPayslip{}, err
	}
	if len(results) != 1 || !strings.EqualFold(results[0].ID, resultID) ||
		!slices.Contains(workerIDs, strings.ToLower(results[0].WorkerID)) || results[0].PayslipDocumentReferenceID == nil {
		return PublishedPayslip{}, hrerrors.ErrNotFound
	}
	result := results[0]
	verifier := deps.PayslipVerifier
	if err := operation.RequireDependency(verifier); err != nil {
		return PublishedPayslip{}, err
	}
	check := PayslipCheck{
		ResultID: result.ID, RunID: result.PayrollRunID, WorkerID: result.WorkerID,
		PrincipalID: actor.PrincipalID, DocumentVersionID: *result.PayslipDocumentReferenceID,
		PublicationID: payslipPublicationID(result),
	}
	slip, err := verifier.VerifyPublishedPayslip(ctx, check)
	if err != nil {
		return PublishedPayslip{}, operation.OwnerVerificationError(ctx, err)
	}
	if !validPublishedPayslip(slip, result.ID, deps.Runtime.Clock.Now()) || !matchesPayslipDocument(slip, result) {
		return PublishedPayslip{}, hrerrors.ErrReferenceConflict
	}
	if err := s.setPayslipPeriod(ctx, &slip, result, ""); err != nil {
		return PublishedPayslip{}, err
	}
	slip.PDF = nil
	if includePDF {
		reader, ok := verifier.(PayslipPDFReader)
		if !ok || operation.RequireDependency(reader) != nil {
			return PublishedPayslip{}, hrerrors.ErrUnavailable
		}
		check.PublicationID = slip.PublicationID
		document, err := reader.ReadVerifiedPayslipPDF(ctx, check, contract.MaxPayslipPDFBytes)
		if err != nil {
			return PublishedPayslip{}, operation.OwnerVerificationError(ctx, err)
		}
		if !strings.EqualFold(document.DocumentVersionID, check.DocumentVersionID) ||
			len(document.Content) > contract.MaxPayslipPDFBytes ||
			!bytes.HasPrefix(document.Content, []byte("%PDF-")) ||
			document.Digest != sha256.Sum256(document.Content) {
			return PublishedPayslip{}, hrerrors.ErrReferenceConflict
		}
		slip.PDF = bytes.Clone(document.Content)
	}
	checked, currentWorkerIDs, err := s.payslipOwner(ctx)
	if err != nil {
		return PublishedPayslip{}, err
	}
	if checked != actor || !slices.Contains(currentWorkerIDs, strings.ToLower(result.WorkerID)) {
		return PublishedPayslip{}, hrerrors.ErrForbidden
	}
	return slip, nil
}

func payslipReadAccess() operation.Access {
	return operation.Access{Domain: "payroll", Resource: "hr_payslips", Action: "read_self"}
}

func (s *Service) payslipOwner(ctx context.Context) (operation.Actor, []string, error) {
	deps := s.dependencies
	actor, err := deps.Runtime.Policy.Authorize(ctx, payslipReadAccess())
	if err != nil {
		return operation.Actor{}, nil, err
	}
	if actor.Kind != "human" || entity.ValidateUUID("principal_id", actor.PrincipalID) != nil {
		return operation.Actor{}, nil, hrerrors.ErrForbidden
	}
	var workerIDs []string
	var personPartyID string
	hasCurrentEmployment := false
	for offset := 0; ; offset += domainrepo.MaxLimit {
		workers, err := deps.Workers.FindByIAMPrincipalReferenceID(ctx, &actor.PrincipalID, domainrepo.MaxLimit, offset)
		if err != nil {
			return operation.Actor{}, nil, err
		}
		for _, worker := range workers {
			if worker.IAMPrincipalReferenceID == nil || !strings.EqualFold(*worker.IAMPrincipalReferenceID, actor.PrincipalID) ||
				entity.ValidateUUID("person_party_reference_id", worker.PersonPartyReferenceID) != nil {
				return operation.Actor{}, nil, hrerrors.ErrForbidden
			}
			if personPartyID == "" {
				personPartyID = worker.PersonPartyReferenceID
			}
			if !strings.EqualFold(worker.PersonPartyReferenceID, personPartyID) {
				return operation.Actor{}, nil, hrerrors.ErrForbidden
			}
			switch worker.Status {
			case entity.WorkerStatusActive, entity.WorkerStatusLeaveOfAbsence, entity.WorkerStatusNoticePeriod, entity.WorkerStatusSuspended:
				hasCurrentEmployment = true
			}
			workerIDs = append(workerIDs, strings.ToLower(worker.ID))
		}
		if len(workers) < domainrepo.MaxLimit {
			break
		}
	}
	if !hasCurrentEmployment {
		return operation.Actor{}, nil, hrerrors.ErrForbidden
	}
	slices.Sort(workerIDs)
	return actor, workerIDs, nil
}

func (s *Service) selectOwnedPayslipResults(ctx context.Context, resultIDs []any, workerIDs []string) ([]entity.PayrollResult, error) {
	var results []entity.PayrollResult
	for group := range slices.Chunk(workerIDs, domainrepo.MaxLimit) {
		values := make([]any, len(group))
		for i, id := range group {
			values[i] = id
		}
		rows, err := s.dependencies.PayrollResult.SelectWhere(ctx, []domainrepo.Filter{
			{Column: "id", Values: resultIDs},
			{Column: "worker_id", Values: values},
		}, len(resultIDs), 0)
		if err != nil {
			return nil, err
		}
		results = append(results, rows...)
	}
	return results, nil
}

func validPublishedPayslip(slip PublishedPayslip, resultID string, now time.Time) bool {
	if !strings.EqualFold(slip.ResultID, resultID) || entity.ValidateUUID("publication_id", slip.PublicationID) != nil ||
		slip.PublishedAt.IsZero() || slip.PublishedAt.After(now) || slip.ScheduledPayDate.IsZero() {
		return false
	}
	switch slip.PaymentStatus {
	case "scheduled", "processing", "paid", "failed":
		return true
	default:
		return false
	}
}

func payslipPublicationID(result entity.PayrollResult) string {
	if result.PayslipPublicationReferenceID == nil {
		return ""
	}
	return *result.PayslipPublicationReferenceID
}

func matchesPayslipDocument(slip PublishedPayslip, result entity.PayrollResult) bool {
	publicationID := payslipPublicationID(result)
	return result.PayslipDocumentReferenceID != nil &&
		strings.EqualFold(*result.PayslipDocumentReferenceID, slip.DocumentVersionID) &&
		(publicationID == "" || strings.EqualFold(publicationID, slip.PublicationID))
}

func (s *Service) setPayslipPeriod(ctx context.Context, slip *PublishedPayslip, result entity.PayrollResult, periodID string) error {
	run, err := s.dependencies.PayrollRun.SelectByID(ctx, result.PayrollRunID)
	if err != nil {
		return err
	}
	if !matchesPayslipRun(*slip, run, periodID) {
		return hrerrors.ErrReferenceConflict
	}
	period, err := s.dependencies.PayrollPeriod.SelectByID(ctx, run.PayrollPeriodID)
	if err != nil {
		return err
	}
	if period == nil || !applyPayslipPeriod(slip, *period) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func matchesPayslipRun(slip PublishedPayslip, run *entity.PayrollRun, periodID string) bool {
	return run != nil && (run.Status == entity.PayrollRunStatusPosted || run.Status == entity.PayrollRunStatusPaid) &&
		run.ApprovedAt != nil && !slip.PublishedAt.Before(*run.ApprovedAt) &&
		(periodID == "" || strings.EqualFold(run.PayrollPeriodID, periodID))
}

func applyPayslipPeriod(slip *PublishedPayslip, period entity.PayrollPeriod) bool {
	if period.PeriodStart.IsZero() || period.PeriodEnd.IsZero() ||
		entity.CompareDates(period.PeriodEnd, period.PeriodStart) < 0 ||
		(slip.PayrollPeriodID != "" && !strings.EqualFold(slip.PayrollPeriodID, period.ID)) {
		return false
	}
	slip.PayrollPeriodID, slip.PeriodCode = period.ID, period.PeriodCode
	slip.PeriodStart, slip.PeriodEnd = period.PeriodStart, period.PeriodEnd
	return true
}
