package service

import (
	"context"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	talentpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/talent"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// RecordTalentCalibrationDecisionInput membawa skor hasil kalibrasi serta bukti dokumennya.
type RecordTalentCalibrationDecisionInput struct {
	SessionID                           string
	AssessmentID                        string
	DimensionID                         string
	CalibratedScore                     int16
	RationaleDocumentVersionReferenceID string
	RationaleDigest                     []byte
}

// LaunchTalentReviewCycle membuka siklus review talenta dari status draft ke open.
func (s *Service) LaunchTalentReviewCycle(ctx context.Context, cycleID string) (*entity.TalentReviewCycle, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.Command(ctx, cycleID, "launch", func(_ context.Context, cycle *entity.TalentReviewCycle, actor operation.Actor, _ time.Time) (*entity.TalentReviewCycle, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if cycle.Status != "draft" || cycle.ClosedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		cycle.Status = "open"
		return cycle, nil
	})
}

// StartTalentReviewCycleCalibration memindahkan siklus review talenta dari open ke tahap kalibrasi.
func (s *Service) StartTalentReviewCycleCalibration(ctx context.Context, cycleID string) (*entity.TalentReviewCycle, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.Command(ctx, cycleID, "start_calibration", func(_ context.Context, cycle *entity.TalentReviewCycle, actor operation.Actor, _ time.Time) (*entity.TalentReviewCycle, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if cycle.Status != "open" || cycle.ClosedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		cycle.Status = "calibration"
		return cycle, nil
	})
}

// CloseTalentReviewCycle menutup siklus talenta setelah seluruh asesmen dan sesi kalibrasi selesai.
func (s *Service) CloseTalentReviewCycle(ctx context.Context, cycleID string) (*entity.TalentReviewCycle, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentReviewCycle.Command(ctx, cycleID, "close", func(txCtx context.Context, cycle *entity.TalentReviewCycle, actor operation.Actor, now time.Time) (*entity.TalentReviewCycle, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if (cycle.Status != "open" && cycle.Status != "calibration") || cycle.ClosedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		pendingAssessments, err := s.dependencies.TalentAssessment.CountWhere(txCtx, []domainrepo.Filter{
			{Column: "talent_review_cycle_id", Values: []any{cycle.ID}},
			{Column: "status", Values: []any{"draft", "submitted"}},
		})
		if err != nil {
			return nil, err
		}
		if pendingAssessments != 0 {
			return nil, hrerrors.ErrWorkflowRequired
		}
		openSessions, err := s.dependencies.TalentCalibrationSession.CountWhere(txCtx, []domainrepo.Filter{
			{Column: "talent_review_cycle_id", Values: []any{cycle.ID}},
			{Column: "status", Values: []any{"planned", "open"}},
		})
		if err != nil {
			return nil, err
		}
		if openSessions != 0 {
			return nil, hrerrors.ErrWorkflowRequired
		}
		cycle.Status = "closed"
		cycle.ClosedAt = &now
		return cycle, nil
	})
}

// OpenTalentCalibrationSession membuka sesi kalibrasi oleh pimpinan panel saat siklus berada pada tahap calibration.
func (s *Service) OpenTalentCalibrationSession(ctx context.Context, sessionID string) (*entity.TalentCalibrationSession, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentCalibrationSession.Command(ctx, sessionID, "open", func(txCtx context.Context, session *entity.TalentCalibrationSession, actor operation.Actor, _ time.Time) (*entity.TalentCalibrationSession, error) {
		if actor.Kind != "human" || !strings.EqualFold(session.ChairPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if session.Status != "planned" || session.ClosedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		cycle, err := s.talentReviewCycle.GetForUpdate(txCtx, session.TalentReviewCycleID)
		if err != nil {
			return nil, err
		}
		if cycle.Status != "calibration" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		session.Status = "open"
		return session, nil
	})
}

// RecordTalentCalibrationDecision mencatat keputusan panel atas skor terfinalisasi tanpa menimpa skor awal.
func (s *Service) RecordTalentCalibrationDecision(ctx context.Context, in RecordTalentCalibrationDecisionInput) (*entity.TalentCalibrationDecision, error) {
	if s == nil || s.talentCalibrationSession == nil || s.talentCalibrationDecision == nil ||
		operation.RequireDependency(s.dependencies.TalentCalibrationDecision) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := entity.ValidateUUID("session_id", in.SessionID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("assessment_id", in.AssessmentID); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("dimension_id", in.DimensionID); err != nil {
		return nil, err
	}
	var recorded *entity.TalentCalibrationDecision
	_, err := s.talentCalibrationSession.Command(ctx, in.SessionID, "approve", func(txCtx context.Context, session *entity.TalentCalibrationSession, actor operation.Actor, now time.Time) (*entity.TalentCalibrationSession, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if session.Status != "open" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		assessment, err := s.talentAssessment.GetForUpdate(txCtx, in.AssessmentID)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(assessment.AssessedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		scoreMatches, err := s.dependencies.TalentAssessmentScore.SelectWhere(txCtx, []domainrepo.Filter{
			{Column: "assessment_id", Values: []any{assessment.ID}},
			{Column: "dimension_id", Values: []any{in.DimensionID}},
		}, 2, 0)
		if err != nil {
			return nil, err
		}
		if len(scoreMatches) != 1 {
			return nil, hrerrors.ErrReferenceConflict
		}
		decisionID, err := s.talentCalibrationDecision.NewID()
		if err != nil {
			return nil, err
		}
		decision := entity.TalentCalibrationDecision{
			ID:                                  decisionID,
			SessionID:                           session.ID,
			TalentReviewCycleID:                 session.TalentReviewCycleID,
			AssessmentID:                        assessment.ID,
			DimensionID:                         in.DimensionID,
			OriginalScore:                       scoreMatches[0].Score,
			CalibratedScore:                     in.CalibratedScore,
			RationaleDocumentVersionReferenceID: in.RationaleDocumentVersionReferenceID,
			RationaleDigest:                     append([]byte(nil), in.RationaleDigest...),
			DecidedByPrincipalID:                actor.PrincipalID,
			DecidedAt:                           now,
		}
		if err := talentpolicy.ValidateCalibration(session, assessment, &scoreMatches[0], &decision); err != nil {
			return nil, err
		}
		if err := s.talentCalibrationDecision.CheckIntegrity(txCtx, nil, &decision); err != nil {
			return nil, err
		}
		if err := s.dependencies.TalentCalibrationDecision.Insert(txCtx, &decision); err != nil {
			return nil, err
		}
		copyDecision := decision
		recorded = &copyDecision
		return session, nil
	})
	if err != nil {
		return nil, err
	}
	return recorded, nil
}

// ApproveTalentPoolMembership mengaktifkan nominasi talent pool dengan pemisahan tugas dari pengusul.
func (s *Service) ApproveTalentPoolMembership(ctx context.Context, membershipID string) (*entity.TalentPoolMembership, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.talentPoolMembership.Command(ctx, membershipID, "approve", func(_ context.Context, membership *entity.TalentPoolMembership, actor operation.Actor, now time.Time) (*entity.TalentPoolMembership, error) {
		if actor.Kind != "human" || strings.EqualFold(membership.NominatedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if membership.Status != "nominated" || membership.ApprovedByPrincipalID != nil || membership.ApprovedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		membership.Status = "active"
		membership.ApprovedByPrincipalID = &actor.PrincipalID
		membership.ApprovedAt = &now
		membership.ExitReasonCode = nil
		return membership, nil
	})
}

// RejectTalentPoolMembership menolak nominasi talent pool dan menyimpan kode alasan.
func (s *Service) RejectTalentPoolMembership(ctx context.Context, membershipID, reasonCode string) (*entity.TalentPoolMembership, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	reason, err := cleanTalentReasonCode(reasonCode)
	if err != nil {
		return nil, err
	}
	return s.talentPoolMembership.Command(ctx, membershipID, "reject", func(_ context.Context, membership *entity.TalentPoolMembership, actor operation.Actor, _ time.Time) (*entity.TalentPoolMembership, error) {
		if actor.Kind != "human" || strings.EqualFold(membership.NominatedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if membership.Status != "nominated" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		membership.Status = "rejected"
		membership.ExitReasonCode = &reason
		return membership, nil
	})
}

// ExitTalentPoolMembership mengakhiri keanggotaan talent pool aktif tanpa menghapus riwayat.
func (s *Service) ExitTalentPoolMembership(ctx context.Context, membershipID, reasonCode string) (*entity.TalentPoolMembership, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	reason, err := cleanTalentReasonCode(reasonCode)
	if err != nil {
		return nil, err
	}
	return s.talentPoolMembership.Command(ctx, membershipID, "exit", func(_ context.Context, membership *entity.TalentPoolMembership, actor operation.Actor, now time.Time) (*entity.TalentPoolMembership, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if membership.Status != "active" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		exitDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if entity.CompareDates(exitDate, membership.ValidFrom) < 0 {
			return nil, hrerrors.ErrWorkflowRequired
		}
		membership.Status = "exited"
		membership.ValidTo = &exitDate
		membership.ExitReasonCode = &reason
		return membership, nil
	})
}

func cleanTalentReasonCode(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || len(trimmed) > 80 {
		return "", hrerrors.Invalid("reason_code", "has invalid value")
	}
	return trimmed, nil
}
