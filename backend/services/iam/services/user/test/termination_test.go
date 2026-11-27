package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
)

type terminationVerifier struct {
	decisions []user.TerminationAccessDecision
}

func (v *terminationVerifier) Verify(context.Context, string, string, string) (user.TerminationAccessDecision, error) {
	decision := v.decisions[0]
	v.decisions = v.decisions[1:]
	return decision, nil
}

func TestApplyEmployeeTerminationRechecksHRMAndRevokesSession(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	h := newTestHarness(t)
	const workerID = "00000000-0000-4000-8000-000000000001"
	const planID = "00000000-0000-4000-8000-000000000002"
	const date = "2026-11-27"
	const email = "termination.fixture@integrione.test"
	const password = "V7$K9q#R2w!X8mZp"
	account, err := h.userSvc.CreateUser(ctx, user.CreateUserInput{
		Kind: entity.UserKindEmployee, Status: entity.UserStatusActive,
		DisplayName: "Termination Fixture", Email: email, Password: password, MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	active, err := h.userSvc.AuthenticatePassword(ctx, user.AuthenticatePasswordInput{Identifier: email, Password: password})
	if err != nil {
		t.Fatal(err)
	}
	allowed := user.TerminationAccessDecision{PrincipalID: account.ID, EffectiveDate: date, MayDisable: true}
	retained := allowed
	retained.MayDisable = false
	apply := func(decisions ...user.TerminationAccessDecision) (user.TerminationOutcome, error) {
		return h.userSvc.ApplyEmployeeTermination(ctx, &terminationVerifier{decisions: decisions}, workerID, planID, date)
	}
	for _, decision := range []user.TerminationAccessDecision{
		{PrincipalID: account.ID, EffectiveDate: "2026-11-26", MayDisable: true},
		{EffectiveDate: "2026-11-26"},
		{PrincipalID: "invalid", EffectiveDate: date},
		{EffectiveDate: date, MayDisable: true},
	} {
		if outcome, err := apply(decision); outcome != "" || !errors.Is(err, iamerrors.ErrConflict) {
			t.Fatalf("invalid verifier decision was acknowledged: %s, %v", outcome, err)
		}
	}
	if outcome, err := apply(retained); err != nil || outcome != user.TerminationAccountRetained {
		t.Fatalf("active employment should retain account: %s, %v", outcome, err)
	}
	if _, err := apply(allowed, retained); !errors.Is(err, iamerrors.ErrConflict) {
		t.Fatalf("changed HRM decision must stop account closure: %v", err)
	}
	if _, err := h.sessSvc.Authenticate(ctx, active.IssuedSession.SessionToken); err != nil {
		t.Fatalf("session was revoked after HRM denied closure: %v", err)
	}
	verify := func(decision user.TerminationAccessDecision, expected string) error {
		return h.userSvc.VerifyEmployeeTerminationClearance(ctx, &terminationVerifier{decisions: []user.TerminationAccessDecision{decision}}, workerID, planID, date, expected)
	}
	if err := verify(allowed, account.ID); !errors.Is(err, iamerrors.ErrConflict) {
		t.Fatalf("active account requiring deactivation was accepted: %v", err)
	}
	if err := verify(retained, account.ID); err != nil {
		t.Fatalf("active account with retained employment was denied: %v", err)
	}
	if err := verify(retained, ""); !errors.Is(err, iamerrors.ErrConflict) {
		t.Fatalf("missing expected account accepted: %v", err)
	}
	if err := verify(user.TerminationAccessDecision{EffectiveDate: date}, ""); err != nil {
		t.Fatalf("verified absence of linked account was denied: %v", err)
	}
	for i := 0; i < 2; i++ {
		if outcome, err := apply(allowed, allowed); err != nil || outcome != user.TerminationAccountDisabled {
			t.Fatalf("apply termination: %s, %v", outcome, err)
		}
	}
	if err := verify(allowed, account.ID); err != nil {
		t.Fatalf("disabled account clearance failed: %v", err)
	}
	if _, err := h.sessSvc.Authenticate(ctx, active.IssuedSession.SessionToken); !errors.Is(err, iamerrors.ErrSessionRevoked) {
		t.Fatalf("session was not revoked: %v", err)
	}
}
