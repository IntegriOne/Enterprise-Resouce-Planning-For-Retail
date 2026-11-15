package entity_test

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"log/slog"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const testSessionID = "66666666-6666-4666-8666-666666666666"

func TestNewSessionLifecycleClampingAndRedaction(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)
	tokenDigest := sha256Sum("opaque-session-handle")
	csrfDigest := sha256Sum("csrf-secret-value")
	ip := netip.MustParseAddr("192.168.10.25")

	sess, err := entity.NewSession(testSessionID, testUserID, tokenDigest, csrfDigest, false, &ip, "Mozilla/5.0", now)
	if err != nil {
		t.Fatalf("NewSession() unexpected error: %v", err)
	}

	// Mutasi input slice tidak boleh mengubah digest di dalam entitas.
	tokenDigest[0] ^= 0xff
	csrfDigest[0] ^= 0xff
	if !sess.MatchesCSRFDigest(sha256Sum("csrf-secret-value")) {
		t.Fatal("MatchesCSRFDigest() failed after caller mutated input slice")
	}
	if sess.MatchesCSRFDigest(sha256Sum("wrong-csrf")) {
		t.Fatal("MatchesCSRFDigest() matched wrong digest")
	}

	// String dan slog wajib menyembunyikan digest sesi dan CSRF.
	if strings.Contains(sess.String(), string(sess.TokenDigest())) {
		t.Fatal("Session.String() must redact TokenDigest")
	}
	var logBuf bytes.Buffer
	slog.New(slog.NewTextHandler(&logBuf, nil)).Info("session", "session", sess)
	if strings.Contains(logBuf.String(), string(sess.CSRFTokenDigest())) {
		t.Fatal("Session.LogValue() must redact CSRFTokenDigest")
	}

	if !sess.IsActive(now.Add(15 * time.Minute)) {
		t.Fatal("session should be active within idle and absolute windows")
	}

	// Simulasi aktivitas berkala tiap 20 menit hingga mendekati akhir batas mutlak 8 jam.
	for step := 20 * time.Minute; step <= 7*time.Hour+50*time.Minute; step += 20 * time.Minute {
		if err := sess.Touch(entity.SessionIdleTimeout, now.Add(step)); err != nil {
			t.Fatalf("Touch(%v) unexpected error: %v", step, err)
		}
	}
	if sess.IdleExpiresAt.After(sess.AbsoluteExpiresAt) {
		t.Fatalf("IdleExpiresAt %v must not exceed AbsoluteExpiresAt %v", sess.IdleExpiresAt, sess.AbsoluteExpiresAt)
	}
	if !sess.IdleExpiresAt.Equal(sess.AbsoluteExpiresAt) {
		t.Fatalf("IdleExpiresAt %v should clamp exactly to AbsoluteExpiresAt %v near 8h limit", sess.IdleExpiresAt, sess.AbsoluteExpiresAt)
	}

	// Idle timeout pada sesi baru tanpa Touch selama 30 menit wajib gagal ErrSessionIdleExpired.
	fresh, err := entity.NewSession(testSessionID, testUserID, sha256Sum("t2"), sha256Sum("c2"), true, nil, "", now)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}
	if !fresh.MFAVerified || fresh.MFAVerifiedAt == nil {
		t.Fatal("MFAVerifiedAt should be populated when session starts MFA-verified")
	}
	if err := fresh.CheckStatus(now.Add(30 * time.Minute)); !errors.Is(err, iamerrors.ErrSessionIdleExpired) {
		t.Fatalf("CheckStatus at idle boundary = %v, want ErrSessionIdleExpired", err)
	}

	// Melewati absolute timeout 8 jam wajib gagal ErrSessionAbsoluteExpired.
	if err := sess.CheckStatus(now.Add(8 * time.Hour)); !errors.Is(err, iamerrors.ErrSessionAbsoluteExpired) {
		t.Fatalf("CheckStatus at absolute boundary = %v, want ErrSessionAbsoluteExpired", err)
	}

	// Pengakhiran sesi bersifat satu kali dan fail-closed setelahnya.
	if err := fresh.End(entity.SessionEndLogout, now.Add(5*time.Minute)); err != nil {
		t.Fatalf("End() error = %v", err)
	}
	if err := fresh.End(entity.SessionEndLogout, now.Add(6*time.Minute)); !errors.Is(err, iamerrors.ErrSessionAlreadyEnded) {
		t.Fatalf("second End() = %v, want ErrSessionAlreadyEnded", err)
	}
	if err := fresh.CheckStatus(now.Add(6 * time.Minute)); !errors.Is(err, iamerrors.ErrSessionRevoked) {
		t.Fatalf("ended session CheckStatus() = %v, want ErrSessionRevoked", err)
	}
}

func TestSessionRehydrateEndStateAndReauth(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.Local)
	reason := entity.SessionEndAdminRevoked

	// ended_at nil tetapi end_reason non-nil harus ditolak seperti ck_sessions_end_state.
	if _, err := entity.RehydrateSession(
		testSessionID,
		testUserID,
		sha256Sum("t"),
		sha256Sum("c"),
		false,
		nil,
		nil,
		"ua",
		now,
		now,
		now.Add(30*time.Minute),
		now.Add(8*time.Hour),
		nil,
		&reason,
		now,
	); !errors.Is(err, iamerrors.ErrInvalidInput) {
		t.Fatalf("RehydrateSession inconsistent end_state = %v, want ErrInvalidInput", err)
	}

	sess, err := entity.NewSession(testSessionID, testUserID, sha256Sum("t"), sha256Sum("c"), false, nil, "ua", now)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}
	if err := sess.MarkMFAVerified(now.Add(2 * time.Minute)); err != nil {
		t.Fatalf("MarkMFAVerified() error = %v", err)
	}
	if !sess.MFAVerified || sess.MFAVerifiedAt == nil {
		t.Fatal("MarkMFAVerified() must mark MFAVerified and MFAVerifiedAt")
	}
	if !sess.IsRecentAuth(10*time.Minute, now.Add(7*time.Minute)) {
		t.Fatal("IsRecentAuth() should be true within 10m window of MFA verification")
	}
	if sess.IsRecentAuth(10*time.Minute, now.Add(13*time.Minute)) {
		t.Fatal("IsRecentAuth() should be false after 10m window")
	}

	newCSRF := sha256Sum("rotated-csrf")
	if err := sess.RotateCSRF(newCSRF); err != nil {
		t.Fatalf("RotateCSRF() error = %v", err)
	}
	if !sess.MatchesCSRFDigest(newCSRF) {
		t.Fatal("MatchesCSRFDigest() should match rotated CSRF digest")
	}
}

func sha256Sum(s string) []byte {
	sum := sha256.Sum256([]byte(s))
	return sum[:]
}
