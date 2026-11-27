package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// WithdrawOffer menarik penawaran terbit tanpa mengubah dokumen atau keputusan kandidat.
func (s *Service) WithdrawOffer(ctx context.Context, id, reasonCode string) (*entity.Offer, error) {
	if s == nil || operation.RequireDependency(s.applicationRecords) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	reasonCode = strings.TrimSpace(reasonCode)
	if reasonCode == "" || utf8.RuneCountInString(reasonCode) > 80 {
		return nil, hrerrors.Invalid("reason_code", "has invalid value")
	}
	return s.offer.Command(ctx, id, "withdraw", func(txCtx context.Context, offer *entity.Offer, actor operation.Actor, now time.Time) (*entity.Offer, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if offer.OfferStatus != "issued" || offer.IssuedAt == nil || now.Before(*offer.IssuedAt) ||
			offer.CandidateResponse != nil || offer.RespondedAt != nil {
			return nil, hrerrors.ErrWorkflowRequired
		}
		application, err := s.applicationRecords.FindByIDForUpdate(txCtx, offer.ApplicationID)
		if err != nil {
			return nil, err
		}
		if application == nil {
			return nil, hrerrors.ErrReferenceConflict
		}
		if application.AssignedRecruiterPrincipalID == nil || !strings.EqualFold(*application.AssignedRecruiterPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if application.Status == entity.ApplicationStatusHired {
			return nil, hrerrors.ErrWorkflowRequired
		}
		offer.OfferStatus = "withdrawn"
		offer.WithdrawnAt, offer.WithdrawnByPrincipalID, offer.WithdrawalReasonCode = &now, &actor.PrincipalID, &reasonCode
		return offer, nil
	})
}
