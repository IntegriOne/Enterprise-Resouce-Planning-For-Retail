package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

type TerminationAccessDecision struct {
	PrincipalID   string
	EffectiveDate string
	MayDisable    bool
}

type TerminationAccessVerifier interface {
	Verify(context.Context, string, string, string) (TerminationAccessDecision, error)
}

type TerminationOutcome string

const (
	TerminationAccountDisabled TerminationOutcome = "account_disabled"
	TerminationAccountRetained TerminationOutcome = "account_retained"
	TerminationNoLinkedAccount TerminationOutcome = "no_linked_account"
)

var errTerminationAlreadyApplied = errors.New("termination already applied")

// ApplyEmployeeTermination memakai bukti HRM dan mekanisme deaktivasi IAM untuk memproses terminasi.
func (s *Service) ApplyEmployeeTermination(ctx context.Context, verifier TerminationAccessVerifier, workerID, planID, effectiveDate string) (TerminationOutcome, error) {
	if err := s.ensureReady(); err != nil {
		return "", err
	}
	if verifier == nil {
		return "", iamerrors.ErrUnavailable
	}
	if !entity.IsValidUUID(workerID) || !entity.IsValidUUID(planID) {
		return "", iamerrors.ErrInvalidInput
	}
	decision, err := verifier.Verify(ctx, workerID, planID, effectiveDate)
	if err != nil {
		return "", err
	}
	date, err := time.Parse(time.DateOnly, decision.EffectiveDate)
	if err != nil || date.IsZero() || (effectiveDate != "" && effectiveDate != decision.EffectiveDate) ||
		((decision.PrincipalID != "" || decision.MayDisable) && !entity.IsValidUUID(decision.PrincipalID)) {
		return "", iamerrors.ErrConflict
	}
	if !decision.MayDisable {
		if decision.PrincipalID == "" {
			return TerminationNoLinkedAccount, nil
		}
		return TerminationAccountRetained, nil
	}
	_, err = s.transitionUserStatus(ctx, decision.PrincipalID, true, func(account *entity.User) error {
		if account == nil || account.Kind != entity.UserKindEmployee || !strings.EqualFold(account.ID, decision.PrincipalID) {
			return iamerrors.ErrConflict
		}
		current, err := verifier.Verify(ctx, workerID, planID, decision.EffectiveDate)
		if err != nil {
			return err
		}
		if !current.MayDisable || !strings.EqualFold(current.PrincipalID, decision.PrincipalID) || current.EffectiveDate != decision.EffectiveDate {
			return iamerrors.ErrConflict
		}
		if account.Status == entity.UserStatusDisabled {
			return errTerminationAlreadyApplied
		}
		return account.Disable(s.clock.Now())
	})
	if err != nil && !errors.Is(err, errTerminationAlreadyApplied) {
		return "", err
	}
	return TerminationAccountDisabled, nil
}

// VerifyEmployeeTerminationClearance memeriksa bukti hubungan kerja dan status akun persisten tanpa mengubah akun.
func (s *Service) VerifyEmployeeTerminationClearance(ctx context.Context, verifier TerminationAccessVerifier, workerID, planID, effectiveDate, expectedPrincipal string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if verifier == nil {
		return iamerrors.ErrUnavailable
	}
	date, err := time.Parse(time.DateOnly, effectiveDate)
	if err != nil || date.IsZero() || !entity.IsValidUUID(workerID) || !entity.IsValidUUID(planID) ||
		(expectedPrincipal != "" && !entity.IsValidUUID(expectedPrincipal)) {
		return iamerrors.ErrInvalidInput
	}
	decision, err := verifier.Verify(ctx, workerID, planID, effectiveDate)
	if err != nil {
		return err
	}
	if decision.EffectiveDate != effectiveDate || !strings.EqualFold(decision.PrincipalID, expectedPrincipal) {
		return iamerrors.ErrConflict
	}
	if expectedPrincipal == "" {
		if decision.MayDisable {
			return iamerrors.ErrConflict
		}
		return nil
	}
	account, err := s.users.SelectByID(ctx, expectedPrincipal)
	if err != nil {
		return err
	}
	if account == nil || account.Kind != entity.UserKindEmployee || !strings.EqualFold(account.ID, expectedPrincipal) {
		return iamerrors.ErrConflict
	}
	if decision.MayDisable {
		if account.Status != entity.UserStatusDisabled {
			return iamerrors.ErrConflict
		}
		if s.projector != nil && s.projector.IsEnabled() {
			reader, ok := s.projector.(interface {
				VerifyDisabledUserProjection(context.Context, *entity.User) error
			})
			if !ok {
				return iamerrors.ErrUnavailable
			}
			return reader.VerifyDisabledUserProjection(ctx, account)
		}
	}
	return nil
}
