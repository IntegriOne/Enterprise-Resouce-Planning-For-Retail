package service

import (
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// SessionPolicy menegakkan aturan timeout sesi, sliding activity dengan clamping mutlak, re-autentikasi, dan kesegaran snapshot.
type SessionPolicy struct {
	idleTimeout           time.Duration
	absoluteTimeout       time.Duration
	snapshotMaxFreshness  time.Duration
	sensitiveReauthWindow time.Duration
}

// NewDefaultSessionPolicy membuat SessionPolicy sesuai standar D01-D12 (idle 30m, absolut 8h, snapshot 30s, reauth 10m).
func NewDefaultSessionPolicy() SessionPolicy {
	return SessionPolicy{
		idleTimeout:           entity.SessionIdleTimeout,
		absoluteTimeout:       entity.SessionAbsoluteTimeout,
		snapshotMaxFreshness:  entity.SnapshotMaxFreshness,
		sensitiveReauthWindow: entity.DefaultSensitiveReauthWindow,
	}
}

// NewSessionPolicy membuat SessionPolicy terkustomisasi dengan nilai bawaan aman jika argumen non-positif.
func NewSessionPolicy(idleTimeout, absoluteTimeout, snapshotMaxFreshness, sensitiveReauthWindow time.Duration) SessionPolicy {
	if idleTimeout <= 0 {
		idleTimeout = entity.SessionIdleTimeout
	}
	if absoluteTimeout <= 0 {
		absoluteTimeout = entity.SessionAbsoluteTimeout
	}
	if snapshotMaxFreshness <= 0 {
		snapshotMaxFreshness = entity.SnapshotMaxFreshness
	}
	if sensitiveReauthWindow <= 0 {
		sensitiveReauthWindow = entity.DefaultSensitiveReauthWindow
	}
	return SessionPolicy{
		idleTimeout:           idleTimeout,
		absoluteTimeout:       absoluteTimeout,
		snapshotMaxFreshness:  snapshotMaxFreshness,
		sensitiveReauthWindow: sensitiveReauthWindow,
	}
}

// IdleTimeout mengembalikan durasi batas idle sesi yang dikonfigurasi.
func (p SessionPolicy) IdleTimeout() time.Duration {
	return p.idleTimeout
}

// AbsoluteTimeout mengembalikan durasi batas mutlak sesi tanpa perpanjangan.
func (p SessionPolicy) AbsoluteTimeout() time.Duration {
	return p.absoluteTimeout
}

// ValidateAndTouch memverifikasi keaktifan sesi secara fail-closed lalu menggeser idle_expires_at tanpa melewati batas mutlak.
func (p SessionPolicy) ValidateAndTouch(session *entity.Session, now time.Time) error {
	if session == nil {
		return iamerrors.ErrSessionNotFound
	}
	return session.Touch(p.idleTimeout, now)
}

// ValidateSensitiveReauth memastikan sesi masih aktif, MFA terverifikasi bila diminta, serta autentikasinya masih segar.
func (p SessionPolicy) ValidateSensitiveReauth(session *entity.Session, requireMFA bool, now time.Time) error {
	if err := session.CheckStatus(now); err != nil {
		return err
	}
	if requireMFA && !session.MFAVerified {
		return iamerrors.ErrSessionReauthRequired
	}
	if !session.IsRecentAuth(p.sensitiveReauthWindow, now) {
		return iamerrors.ErrSessionReauthRequired
	}
	return nil
}

// ValidateSnapshotFreshness menolak secara fail-closed bila snapshot otorisasi kosong atau melewati batas kesegaran 30 detik.
func (p SessionPolicy) ValidateSnapshotFreshness(snapshot *entity.AccessSnapshot, now time.Time) error {
	if snapshot == nil || !snapshot.IsFresh(p.snapshotMaxFreshness, now) {
		return iamerrors.ErrSnapshotStale
	}
	return nil
}
