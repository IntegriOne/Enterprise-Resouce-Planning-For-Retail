package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/contract"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ErrPayslipContentMissing menandakan DRM mengonfirmasi blob versi masih tercatat namun hilang.
var ErrPayslipContentMissing = errors.New("payslip version content missing")

// PayslipReissueVerifier menemukan penerbitan Finance dan membuka versi DRM untuk petugas berwenang.
type PayslipReissueVerifier interface {
	FindPublicationForReissue(context.Context, PayslipPublicationCheck) (string, error)
	ReadPayslipForReissue(context.Context, PayslipCheck, string, int) (VerifiedPayslipPDF, error)
}

// ReissueEmployeePayslip membuka PDF privat dan membuatnya bila hasil yang terbit belum memiliki versi.
func (s *Service) ReissueEmployeePayslip(ctx context.Context, workerID, resultID string) (VerifiedPayslipPDF, error) {
	if s == nil {
		return VerifiedPayslipPDF{}, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("worker_id", workerID); err != nil {
		return VerifiedPayslipPDF{}, err
	}
	if err := entity.ValidateUUID("payroll_result_id", resultID); err != nil {
		return VerifiedPayslipPDF{}, err
	}
	deps := s.dependencies
	if operation.RequireDependency(deps.PayslipReissue) != nil || operation.RequireDependency(deps.Workers) != nil ||
		operation.RequireDependency(deps.PayslipPublication) != nil {
		return VerifiedPayslipPDF{}, hrerrors.ErrUnavailable
	}
	access := operation.Access{Domain: "payroll", Resource: "hr_payslips", Action: "reissue"}
	actor, err := deps.Runtime.Policy.Authorize(ctx, access)
	if err != nil {
		return VerifiedPayslipPDF{}, err
	}
	if actor.Kind != "human" {
		return VerifiedPayslipPDF{}, hrerrors.ErrForbidden
	}
	result, err := deps.PayrollResult.SelectByID(ctx, resultID)
	if err != nil {
		return VerifiedPayslipPDF{}, err
	}
	if result == nil || !strings.EqualFold(result.ID, resultID) || !strings.EqualFold(result.WorkerID, workerID) {
		return VerifiedPayslipPDF{}, hrerrors.ErrNotFound
	}
	run, err := deps.PayrollRun.SelectByID(ctx, result.PayrollRunID)
	if err != nil {
		return VerifiedPayslipPDF{}, err
	}
	if run == nil || !strings.EqualFold(run.ID, result.PayrollRunID) {
		return VerifiedPayslipPDF{}, hrerrors.ErrReferenceConflict
	}
	access.Attributes = map[string]string{"employing_legal_entity_reference_id": run.EmployingLegalEntityReferenceID}
	recheck := func(ctx context.Context) error {
		current, err := deps.Runtime.Policy.Authorize(ctx, access)
		if err != nil {
			return err
		}
		if current != actor {
			return hrerrors.ErrForbidden
		}
		return ctx.Err()
	}
	if err := recheck(ctx); err != nil {
		return VerifiedPayslipPDF{}, err
	}
	if (run.Status != entity.PayrollRunStatusPosted && run.Status != entity.PayrollRunStatusPaid) ||
		run.FinanceJournalReferenceID == nil {
		return VerifiedPayslipPDF{}, hrerrors.ErrWorkflowRequired
	}
	publicationID := payslipPublicationID(*result)
	if publicationID == "" {
		publicationID, err = deps.PayslipReissue.FindPublicationForReissue(ctx, PayslipPublicationCheck{
			RunID: run.ID, JournalID: *run.FinanceJournalReferenceID, LegalEntityID: run.EmployingLegalEntityReferenceID,
			InputSnapshotID: run.InputSnapshotReferenceID, InputDigest: bytes.Clone(run.InputDigest), OutputDigest: bytes.Clone(run.OutputDigest),
		})
		if err != nil {
			return VerifiedPayslipPDF{}, operation.OwnerVerificationError(ctx, err)
		}
	}
	if err := recheck(ctx); err != nil {
		return VerifiedPayslipPDF{}, err
	}
	documentID, outcome, err := s.publishPayslipForResult(ctx, publicationID, result.ID, run.ID, "", recheck)
	if err != nil {
		return VerifiedPayslipPDF{}, err
	}
	if err := recheck(ctx); err != nil {
		return VerifiedPayslipPDF{}, err
	}
	check := PayslipCheck{
		ResultID: result.ID, RunID: run.ID, WorkerID: result.WorkerID, PrincipalID: actor.PrincipalID,
		DocumentVersionID: documentID, PublicationID: publicationID,
	}
	pdf, err := deps.PayslipReissue.ReadPayslipForReissue(ctx, check, run.EmployingLegalEntityReferenceID, contract.MaxPayslipPDFBytes)
	if errors.Is(err, ErrPayslipContentMissing) {
		if err := recheck(ctx); err != nil {
			return VerifiedPayslipPDF{}, err
		}
		_, outcome, err = s.publishPayslipForResult(ctx, publicationID, result.ID, run.ID, documentID, recheck)
		if err != nil {
			return VerifiedPayslipPDF{}, err
		}
		if err := recheck(ctx); err != nil {
			return VerifiedPayslipPDF{}, err
		}
		pdf, err = deps.PayslipReissue.ReadPayslipForReissue(ctx, check, run.EmployingLegalEntityReferenceID, contract.MaxPayslipPDFBytes)
	}
	if err != nil {
		return VerifiedPayslipPDF{}, operation.OwnerVerificationError(ctx, err)
	}
	if !strings.EqualFold(pdf.DocumentVersionID, documentID) || len(pdf.Content) > contract.MaxPayslipPDFBytes ||
		!bytes.HasPrefix(pdf.Content, []byte("%PDF-")) || pdf.Digest != sha256.Sum256(pdf.Content) {
		return VerifiedPayslipPDF{}, hrerrors.ErrReferenceConflict
	}
	if err := recheck(ctx); err != nil {
		return VerifiedPayslipPDF{}, err
	}
	if err := s.recordPayslipReissue(ctx, actor, result, documentID, outcome); err != nil {
		return VerifiedPayslipPDF{}, err
	}
	pdf.Content = bytes.Clone(pdf.Content)
	return pdf, nil
}

func (s *Service) recordPayslipReissue(ctx context.Context, actor operation.Actor, result *entity.PayrollResult, documentID, outcome string) error {
	deps := s.dependencies.Runtime
	id, err := deps.IDs.NewID()
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]string{
		"actor_principal_id": actor.PrincipalID, "worker_id": result.WorkerID,
		"payroll_result_id": result.ID, "document_version_id": documentID, "outcome": outcome,
	})
	if err != nil {
		return err
	}
	digest := sha256.Sum256(payload)
	now := deps.Clock.Now()
	event, err := entity.NewOutboxEvent(entity.OutboxEvent{
		ID: id, AggregateID: id, AggregateType: "hr_payslip_reissues", AggregateVersion: 1,
		EventType: "hr_payslip_reissues.recorded", SchemaVersion: 1, Payload: entity.JSON(payload),
		PayloadDigest: digest[:], OccurredAt: now, AvailableAt: now,
	})
	if err != nil {
		return err
	}
	return deps.Outbox.Insert(ctx, event)
}
