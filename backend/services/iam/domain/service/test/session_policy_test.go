package service_test

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

func TestSessionPolicy(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)
	policy := service.NewDefaultSessionPolicy()
	tokDigest := sha256.Sum256([]byte("session-token"))
	csrfDigest := sha256.Sum256([]byte("csrf-token"))

	sess, err := entity.NewSession(
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222",
		tokDigest[:],
		csrfDigest[:],
		true,
		nil,
		"Mozilla/5.0",
		now,
	)
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}

	if err := policy.ValidateAndTouch(nil, now); !errors.Is(err, iamerrors.ErrSessionNotFound) {
		t.Fatalf("ValidateAndTouch(nil) = %v, want ErrSessionNotFound", err)
	}

	if err := policy.ValidateAndTouch(sess, now.Add(15*time.Minute)); err != nil {
		t.Fatalf("ValidateAndTouch(15m) unexpected error: %v", err)
	}

	if err := policy.ValidateSensitiveReauth(sess, true, now.Add(5*time.Minute)); err != nil {
		t.Fatalf("ValidateSensitiveReauth(5m) unexpected error: %v", err)
	}
	if err := policy.ValidateSensitiveReauth(sess, true, now.Add(11*time.Minute)); !errors.Is(err, iamerrors.ErrSessionReauthRequired) {
		t.Fatalf("ValidateSensitiveReauth(11m) = %v, want ErrSessionReauthRequired", err)
	}

	// Non-MFA session must fail when MFA is required for sensitive action.
	noMFASess, _ := entity.NewSession(
		"11111111-1111-4111-8111-111111111112",
		"22222222-2222-4222-8222-222222222222",
		tokDigest[:],
		csrfDigest[:],
		false,
		nil,
		"Mozilla/5.0",
		now,
	)
	if err := policy.ValidateSensitiveReauth(noMFASess, true, now.Add(time.Minute)); !errors.Is(err, iamerrors.ErrSessionReauthRequired) {
		t.Fatalf("ValidateSensitiveReauth(noMFA) = %v, want ErrSessionReauthRequired", err)
	}

	snap := &entity.AccessSnapshot{
		UserID:     "22222222-2222-4222-8222-222222222222",
		UserStatus: entity.UserStatusActive,
		CheckedAt:  now,
	}
	if err := policy.ValidateSnapshotFreshness(nil, now); !errors.Is(err, iamerrors.ErrSnapshotStale) {
		t.Fatalf("ValidateSnapshotFreshness(nil) = %v, want ErrSnapshotStale", err)
	}
	if err := policy.ValidateSnapshotFreshness(snap, now.Add(29*time.Second)); err != nil {
		t.Fatalf("ValidateSnapshotFreshness(29s) error: %v", err)
	}
	if err := policy.ValidateSnapshotFreshness(snap, now.Add(31*time.Second)); !errors.Is(err, iamerrors.ErrSnapshotStale) {
		t.Fatalf("ValidateSnapshotFreshness(31s) = %v, want ErrSnapshotStale", err)
	}

	custom := service.NewSessionPolicy(0, 0, 0, 0)
	if custom.IdleTimeout() != entity.SessionIdleTimeout || custom.AbsoluteTimeout() != entity.SessionAbsoluteTimeout {
		t.Fatalf("NewSessionPolicy zero defaults mismatch: idle=%v abs=%v", custom.IdleTimeout(), custom.AbsoluteTimeout())
	}
	clockNow := (service.SystemClock{}).Now()
	if clockNow.IsZero() {
		t.Fatal("SystemClock.Now() returned zero time")
	}
	if clockNow.Location() != time.Local {
		t.Fatalf("SystemClock.Now().Location() = %v, want %v", clockNow.Location(), time.Local)
	}
}
