package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ApproveBenefitEnrollment mengesahkan pengajuan lokal setelah pemeriksaan kelayakan HRM.
func (s *Service) ApproveBenefitEnrollment(ctx context.Context, id string) (*entity.BenefitEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.benefitEnrollment.Command(ctx, id, "approve", func(_ context.Context, enrollment *entity.BenefitEnrollment, actor operation.Actor, now time.Time) (*entity.BenefitEnrollment, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if enrollment.EnrollmentStatus != "pending" || enrollment.VerifiedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		enrollment.EnrollmentStatus = "approved"
		enrollment.VerifiedAt = &now
		return enrollment, nil
	})
}

// EndBenefitEnrollment mengakhiri enrollment lokal dan mempertahankan bukti persetujuan.
func (s *Service) EndBenefitEnrollment(ctx context.Context, id string, effectiveTo time.Time) (*entity.BenefitEnrollment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if effectiveTo.IsZero() {
		return nil, hrerrors.Invalid("effective_to", "is required")
	}
	return s.benefitEnrollment.Command(ctx, id, "end", func(_ context.Context, enrollment *entity.BenefitEnrollment, actor operation.Actor, now time.Time) (*entity.BenefitEnrollment, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if enrollment.EnrollmentStatus != "approved" || enrollment.VerifiedAt == nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if entity.CompareDates(effectiveTo, now.UTC()) > 0 || entity.CompareDates(effectiveTo, enrollment.EffectiveFrom) < 0 ||
			(enrollment.EffectiveTo != nil && entity.CompareDates(effectiveTo, *enrollment.EffectiveTo) > 0) {
			return nil, hrerrors.ErrReferenceConflict
		}
		enrollment.EnrollmentStatus = "ended"
		enrollment.EffectiveTo = &effectiveTo
		return enrollment, nil
	})
}
