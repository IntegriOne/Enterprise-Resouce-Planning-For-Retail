package entity

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/netip"
	"strings"
	"time"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// Session merepresentasikan sesi autentikasi stateful pada tabel sessions dengan idle dan absolute timeout tetap.
type Session struct {
	ID                string
	UserID            string
	tokenDigest       []byte
	csrfTokenDigest   []byte
	Generation        int64
	CSRFKeyVersion    int
	MFAVerified       bool
	MFAVerifiedAt     *time.Time
	IPAddress         *netip.Addr
	UserAgent         string
	AuthenticatedAt   time.Time
	LastActiveAt      time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	EndedAt           *time.Time
	EndReason         *SessionEndReason
	CreatedAt         time.Time
}

// NewSession membentuk sesi baru dengan batas idle 30 menit dan batas mutlak 8 jam tanpa diperpanjang.
func NewSession(
	id, userID string,
	tokenDigest, csrfTokenDigest []byte,
	mfaVerified bool,
	ip *netip.Addr,
	userAgent string,
	now time.Time,
) (*Session, error) {
	ts := normalizeLocalTime(now)
	var mfaAt *time.Time
	if mfaVerified {
		v := ts
		mfaAt = &v
	}
	return RehydrateSession(
		id,
		userID,
		tokenDigest,
		csrfTokenDigest,
		mfaVerified,
		mfaAt,
		ip,
		userAgent,
		ts,
		ts,
		ts.Add(SessionIdleTimeout),
		ts.Add(SessionAbsoluteTimeout),
		nil,
		nil,
		ts,
	)
}

// RehydrateSession membangun kembali sesi dari lapisan persistensi sambil menegakkan ck_sessions_expiry_order dan ck_sessions_end_state.
func RehydrateSession(
	id, userID string,
	tokenDigest, csrfTokenDigest []byte,
	mfaVerified bool,
	mfaVerifiedAt *time.Time,
	ip *netip.Addr,
	userAgent string,
	authenticatedAt, lastActiveAt, idleExpiresAt, absoluteExpiresAt time.Time,
	endedAt *time.Time,
	endReason *SessionEndReason,
	createdAt time.Time,
) (*Session, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if err := ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	if len(tokenDigest) == 0 {
		return nil, iamerrors.NewValidation("token_digest", iamerrors.Code("iam.session.empty_token_digest"), "session token digest must not be empty")
	}
	if len(csrfTokenDigest) == 0 {
		return nil, iamerrors.NewValidation("csrf_token_digest", iamerrors.Code("iam.session.empty_csrf_digest"), "csrf token digest must not be empty")
	}
	ipCopy, err := cloneAndValidateIP(ip)
	if err != nil {
		return nil, err
	}

	authLocal := normalizeLocalTime(authenticatedAt)
	absLocal := normalizeLocalTime(absoluteExpiresAt)
	idleLocal := normalizeLocalTime(idleExpiresAt)
	if !absLocal.After(authLocal) || idleLocal.After(absLocal) {
		return nil, iamerrors.NewValidation("expiry", iamerrors.Code("iam.session.invalid_expiry_order"), "session expiry order violates idle <= absolute > authenticated")
	}
	if (endedAt == nil) != (endReason == nil) {
		return nil, iamerrors.NewValidation("end_state", iamerrors.Code("iam.session.invalid_end_state"), "ended_at and end_reason must both be nil or both non-nil")
	}
	var reasonCopy *SessionEndReason
	if endReason != nil {
		if !endReason.IsValid() {
			return nil, iamerrors.ErrInvalidSessionEndReason
		}
		r := *endReason
		reasonCopy = &r
	}

	return &Session{
		ID:                strings.TrimSpace(id),
		UserID:            strings.TrimSpace(userID),
		tokenDigest:       bytes.Clone(tokenDigest),
		csrfTokenDigest:   bytes.Clone(csrfTokenDigest),
		Generation:        1,
		CSRFKeyVersion:    1,
		MFAVerified:       mfaVerified,
		MFAVerifiedAt:     cloneTimePtr(mfaVerifiedAt),
		IPAddress:         ipCopy,
		UserAgent:         strings.TrimSpace(userAgent),
		AuthenticatedAt:   authLocal,
		LastActiveAt:      normalizeLocalTime(lastActiveAt),
		IdleExpiresAt:     idleLocal,
		AbsoluteExpiresAt: absLocal,
		EndedAt:           cloneTimePtr(endedAt),
		EndReason:         reasonCopy,
		CreatedAt:         normalizeLocalTime(createdAt),
	}, nil
}

// SetSecurityMetadata menetapkan generasi monotonik dan versi kunci CSRF secara aman.
func (s *Session) SetSecurityMetadata(generation int64, csrfKeyVersion int) {
	if s == nil {
		return
	}
	if generation >= 1 {
		s.Generation = generation
	} else {
		s.Generation = 1
	}
	if csrfKeyVersion >= 1 {
		s.CSRFKeyVersion = csrfKeyVersion
	} else {
		s.CSRFKeyVersion = 1
	}
}

// BumpGeneration menaikkan generasi sesi secara monotonik (misal saat step-up MFA merotasi CSRF).
func (s *Session) BumpGeneration() int64 {
	if s == nil {
		return 1
	}
	if s.Generation < 1 {
		s.Generation = 1
	}
	s.Generation++
	return s.Generation
}

// TokenDigest mengembalikan salinan defensif digest handle sesi.
func (s *Session) TokenDigest() []byte {
	if s == nil {
		return nil
	}
	return bytes.Clone(s.tokenDigest)
}

// CSRFTokenDigest mengembalikan salinan defensif digest token CSRF sesi.
func (s *Session) CSRFTokenDigest() []byte {
	if s == nil {
		return nil
	}
	return bytes.Clone(s.csrfTokenDigest)
}

// MatchesCSRFDigest membandingkan kandidat digest CSRF secara constant-time terhadap token sesi aktif.
func (s *Session) MatchesCSRFDigest(candidate []byte) bool {
	if s == nil || len(s.csrfTokenDigest) == 0 || len(candidate) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare(s.csrfTokenDigest, candidate) == 1
}

// CheckStatus memverifikasi apakah sesi masih aktif dan belum berakhir karena pencabutan, idle, atau batas mutlak.
func (s *Session) CheckStatus(now time.Time) error {
	if s == nil {
		return iamerrors.ErrSessionNotFound
	}
	if s.EndedAt != nil || s.EndReason != nil {
		if s.EndReason != nil {
			switch *s.EndReason {
			case SessionEndIdleTimeout:
				return iamerrors.ErrSessionIdleExpired
			case SessionEndAbsoluteTimeout:
				return iamerrors.ErrSessionAbsoluteExpired
			}
		}
		return iamerrors.ErrSessionRevoked
	}
	nowLocal := normalizeLocalTime(now)
	if !nowLocal.Before(normalizeLocalTime(s.AbsoluteExpiresAt)) {
		return iamerrors.ErrSessionAbsoluteExpired
	}
	if !nowLocal.Before(normalizeLocalTime(s.IdleExpiresAt)) {
		return iamerrors.ErrSessionIdleExpired
	}
	return nil
}

// IsActive melaporkan apakah sesi valid untuk digunakan pada waktu now.
func (s *Session) IsActive(now time.Time) bool {
	return s.CheckStatus(now) == nil
}

// Touch memperbarui last_active_at dan menggeser idle_expires_at tanpa melampaui absolute_expires_at.
func (s *Session) Touch(idleTimeout time.Duration, now time.Time) error {
	if err := s.CheckStatus(now); err != nil {
		return err
	}
	if idleTimeout <= 0 {
		idleTimeout = SessionIdleTimeout
	}
	nowLocal := normalizeLocalTime(now)
	absLocal := normalizeLocalTime(s.AbsoluteExpiresAt)
	nextIdle := nowLocal.Add(idleTimeout)
	if nextIdle.After(absLocal) {
		nextIdle = absLocal
	}
	s.LastActiveAt = nowLocal
	s.IdleExpiresAt = nextIdle
	return nil
}

// MarkMFAVerified mencatat keberhasilan tantangan MFA pada sesi aktif.
func (s *Session) MarkMFAVerified(now time.Time) error {
	if err := s.CheckStatus(now); err != nil {
		return err
	}
	ts := normalizeLocalTime(now)
	s.MFAVerified = true
	s.MFAVerifiedAt = &ts
	return nil
}

// RotateCSRF mengganti digest token CSRF dengan salinan defensif baru.
func (s *Session) RotateCSRF(newDigest []byte) error {
	if s == nil {
		return iamerrors.ErrSessionNotFound
	}
	if len(newDigest) == 0 {
		return iamerrors.NewValidation("csrf_token_digest", iamerrors.Code("iam.session.empty_csrf_digest"), "csrf token digest must not be empty")
	}
	s.csrfTokenDigest = bytes.Clone(newDigest)
	return nil
}

// End mengakhiri sesi secara permanen dengan alasan standar skema session_end_reason.
func (s *Session) End(reason SessionEndReason, now time.Time) error {
	if s == nil {
		return iamerrors.ErrSessionNotFound
	}
	if s.EndedAt != nil {
		return iamerrors.ErrSessionAlreadyEnded
	}
	if !reason.IsValid() {
		return iamerrors.ErrInvalidSessionEndReason
	}
	ts := normalizeLocalTime(now)
	r := reason
	s.EndedAt = &ts
	s.EndReason = &r
	return nil
}

// IsRecentAuth memeriksa apakah sesi masih memenuhi batas kesegaran re-autentikasi untuk aksi sensitif.
func (s *Session) IsRecentAuth(maxAge time.Duration, now time.Time) bool {
	if err := s.CheckStatus(now); err != nil || maxAge <= 0 {
		return false
	}
	authTime := s.AuthenticatedAt
	if s.MFAVerified && s.MFAVerifiedAt != nil {
		authTime = *s.MFAVerifiedAt
	}
	nowLocal := normalizeLocalTime(now)
	authLocal := normalizeLocalTime(authTime)
	if nowLocal.Before(authLocal) {
		return false
	}
	return nowLocal.Sub(authLocal) <= maxAge
}

// String menyembunyikan digest token sesi dan CSRF dari keluaran string.
func (s Session) String() string {
	return fmt.Sprintf("Session{ID:%s, UserID:%s, MFAVerified:%t, TokenDigest:%s, CSRFDigest:%s}",
		s.ID, s.UserID, s.MFAVerified, redactedSecret, redactedSecret)
}

// LogValue memancarkan atribut sesi aman untuk slog tanpa token mentah atau digest.
func (s Session) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", s.ID),
		slog.String("user_id", s.UserID),
		slog.Bool("mfa_verified", s.MFAVerified),
		slog.Time("idle_expires_at", s.IdleExpiresAt),
		slog.Time("absolute_expires_at", s.AbsoluteExpiresAt),
		slog.String("token_digest", redactedSecret),
		slog.String("csrf_token_digest", redactedSecret),
	)
}

func cloneAndValidateIP(ip *netip.Addr) (*netip.Addr, error) {
	if ip == nil {
		return nil, nil
	}
	if !ip.IsValid() {
		return nil, iamerrors.ErrInvalidIPAddress
	}
	v := ip.Unmap()
	return &v, nil
}
