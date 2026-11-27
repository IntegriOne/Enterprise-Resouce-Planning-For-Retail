package service

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ApprovalCheck mengikat mandat panel pada rencana, kandidat, dan pelaku keputusan.
type ApprovalCheck struct {
	PlanID, PositionID, PrincipalID string
	PlanVersion                     int32
	ReviewDate, DecidedAt           time.Time
	Nominations                     []ApprovalNomination
}

// ApprovalNomination memuat identitas dan bukti kesiapan nominasi untuk verifikasi panel.
type ApprovalNomination struct {
	ID, WorkerID, ReadinessCode           string
	ReadinessAssessedAt                   time.Time
	TalentAssessmentID, DevelopmentPlanID *string
}

// ApprovalVerifier membuktikan mandat panel yang berlaku dan ketiadaan konflik kepentingan.
type ApprovalVerifier interface {
	VerifySuccessionApproval(context.Context, ApprovalCheck) (string, error)
}

// ApproveSuccessionPlan mengesahkan rencana tanpa mengubah posisi pekerja yang dinominasikan.
func (s *Service) ApproveSuccessionPlan(ctx context.Context, id string) (*entity.SuccessionPlan, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.successionPlan.Command(ctx, id, "approve", func(txCtx context.Context, plan *entity.SuccessionPlan, actor operation.Actor, now time.Time) (*entity.SuccessionPlan, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if plan.ApprovedAt != nil || plan.ApprovedByPrincipalID != nil || plan.ApprovalMandateReferenceID != nil || plan.SuccessionRisk != nil ||
			plan.ConfidentialityClass != "highly_restricted" || entity.CompareDates(now, plan.ReviewDate) < 0 {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if plan.CreatedByPrincipalID == nil || plan.CreatedAt == nil || now.Before(*plan.CreatedAt) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if strings.EqualFold(*plan.CreatedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if operation.RequireDependency(s.approvalVerifier) != nil || operation.RequireDependency(s.approvalWorkers) != nil {
			return nil, hrerrors.ErrUnavailable
		}
		filters := []domainrepo.Filter{{Column: "succession_plan_id", Values: []any{plan.ID}}}
		count, err := s.candidateRecords.CountWhere(txCtx, filters)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, hrerrors.ErrReferenceConflict
		}
		candidates, err := operation.CollectComplete(txCtx, count, func(limit, offset int) (operation.Page[entity.SuccessionCandidate], error) {
			items, err := s.candidateRecords.SelectWhere(txCtx, filters, limit, offset)
			return operation.Page[entity.SuccessionCandidate]{Items: items}, err
		})
		if err != nil {
			return nil, err
		}
		// Rencana berbeda dapat mengunci pekerja yang sama dalam satu transaksi.
		slices.SortFunc(candidates, func(a, b entity.SuccessionCandidate) int {
			return strings.Compare(a.CandidateWorkerID, b.CandidateWorkerID)
		})
		nominations := make([]ApprovalNomination, 0, len(candidates))
		for _, candidate := range candidates {
			if err := s.successionCandidate.CheckIntegrity(txCtx, &candidate, &candidate); err != nil {
				return nil, err
			}
			worker, err := s.approvalWorkers.FindByIDForUpdate(txCtx, candidate.CandidateWorkerID)
			if errors.Is(err, hrerrors.ErrNotFound) || (err == nil && worker == nil) {
				return nil, hrerrors.ErrReferenceConflict
			}
			if err != nil {
				return nil, err
			}
			if worker.IAMPrincipalReferenceID != nil && strings.EqualFold(*worker.IAMPrincipalReferenceID, actor.PrincipalID) {
				return nil, hrerrors.ErrForbidden
			}
			nominations = append(nominations, ApprovalNomination{
				ID: candidate.ID, WorkerID: candidate.CandidateWorkerID, ReadinessCode: candidate.ReadinessCode,
				ReadinessAssessedAt: candidate.ReadinessAssessedAt,
				TalentAssessmentID:  candidate.TalentAssessmentID, DevelopmentPlanID: candidate.DevelopmentPlanID,
			})
		}
		mandate, err := s.approvalVerifier.VerifySuccessionApproval(txCtx, ApprovalCheck{
			PlanID: plan.ID, PositionID: plan.PositionID, PrincipalID: actor.PrincipalID, PlanVersion: plan.PlanVersion,
			ReviewDate: plan.ReviewDate, DecidedAt: now, Nominations: nominations,
		})
		if err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if err := entity.ValidateUUID("approval_mandate_reference_id", mandate); err != nil {
			return nil, hrerrors.ErrReferenceConflict
		}
		plan.ApprovedAt, plan.ApprovedByPrincipalID, plan.ApprovalMandateReferenceID = &now, &actor.PrincipalID, &mandate
		return plan, nil
	})
}
