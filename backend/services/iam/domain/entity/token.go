package entity

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"time"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// UserToken merepresentasikan token sekali pakai bertujuan khusus pada tabel user_tokens.
type UserToken struct {
	ID          string
	UserID      string
	IdentityID  *string
	Purpose     TokenPurpose
	Status      TokenStatus
	tokenDigest []byte
	ExpiresAt   time.Time
	UsedAt      *time.Time
	CreatedAt   time.Time
}

// NewUserToken membentuk token sekali pakai baru sekaligus menegakkan constraint expires_at > created_at.
func NewUserToken(id, userID string, purpose TokenPurpose, tokenDigest []byte, now, expiresAt time.Time) (*UserToken, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if err := ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}
	if !purpose.IsValid() {
		return nil, iamerrors.ErrTokenInvalidPurpose
	}
	if len(tokenDigest) == 0 {
		return nil, iamerrors.ErrInvalidTokenDigest
	}
	createdLocal := normalizeLocalTime(now)
	expiresLocal := normalizeLocalTime(expiresAt)
	if !expiresLocal.After(createdLocal) {
		return nil, iamerrors.NewValidation("expires_at", iamerrors.Code("iam.token.invalid_expiry"), "expires_at must be strictly after created_at")
	}

	return &UserToken{
		ID:          strings.TrimSpace(id),
		UserID:      strings.TrimSpace(userID),
		Purpose:     purpose,
		Status:      TokenStatusActive,
		tokenDigest: bytes.Clone(tokenDigest),
		ExpiresAt:   expiresLocal,
		CreatedAt:   createdLocal,
	}, nil
}

// NewUserTokenForIdentity membentuk token sekali pakai yang tertaut ke baris user_identities tertentu.
func NewUserTokenForIdentity(id, userID, identityID string, purpose TokenPurpose, tokenDigest []byte, now, expiresAt time.Time) (*UserToken, error) {
	tok, err := NewUserToken(id, userID, purpose, tokenDigest, now, expiresAt)
	if err != nil {
		return nil, err
	}
	trimmedIdentityID := strings.TrimSpace(identityID)
	if trimmedIdentityID != "" {
		if err := ValidateUUID("identity_id", trimmedIdentityID); err != nil {
			return nil, err
		}
		tok.IdentityID = &trimmedIdentityID
	}
	return tok, nil
}

// RehydrateUserToken membangun kembali entitas token dari lapisan persistensi.
func RehydrateUserToken(
	id, userID string,
	identityID *string,
	purpose TokenPurpose,
	status TokenStatus,
	tokenDigest []byte,
	expiresAt time.Time,
	usedAt *time.Time,
	createdAt time.Time,
) (*UserToken, error) {
	tok, err := NewUserToken(id, userID, purpose, tokenDigest, createdAt, expiresAt)
	if err != nil {
		return nil, err
	}
	if !status.IsValid() {
		return nil, iamerrors.ErrInvalidTokenStatus
	}
	if identityID != nil {
		trimmedIdentityID := strings.TrimSpace(*identityID)
		if trimmedIdentityID != "" {
			if err := ValidateUUID("identity_id", trimmedIdentityID); err != nil {
				return nil, err
			}
			tok.IdentityID = &trimmedIdentityID
		}
	}
	if status == TokenStatusConsumed && usedAt == nil {
		return nil, iamerrors.NewValidation("consumed_at", iamerrors.Code("iam.token.missing_consumed_at"), "consumed token must have consumed_at")
	}
	if status != TokenStatusConsumed && usedAt != nil {
		return nil, iamerrors.NewValidation("status", iamerrors.Code("iam.token.inconsistent_status"), "non-consumed token must not have consumed_at")
	}
	tok.Status = status
	tok.UsedAt = cloneTimePtr(usedAt)
	return tok, nil
}

// TokenDigest mengembalikan salinan defensif digest SHA-256 token.
func (t *UserToken) TokenDigest() []byte {
	if t == nil {
		return nil
	}
	return bytes.Clone(t.tokenDigest)
}

// ValidateForUse memeriksa tujuan, status siklus hidup, pemakaian, dan masa berlaku token secara fail-closed.
func (t *UserToken) ValidateForUse(expectedPurpose TokenPurpose, now time.Time) error {
	if t == nil {
		return iamerrors.ErrTokenNotFound
	}
	if t.Purpose != expectedPurpose {
		return iamerrors.ErrTokenInvalidPurpose
	}
	if t.Status == TokenStatusConsumed || t.UsedAt != nil {
		return iamerrors.ErrTokenAlreadyUsed
	}
	if t.Status == TokenStatusRevoked {
		return iamerrors.ErrTokenRevoked
	}
	if t.Status == TokenStatusExpired {
		return iamerrors.ErrTokenExpired
	}
	if t.Status != TokenStatusActive {
		return iamerrors.ErrInvalidTokenStatus
	}
	if !normalizeLocalTime(now).Before(normalizeLocalTime(t.ExpiresAt)) {
		return iamerrors.ErrTokenExpired
	}
	return nil
}

// Consume memvalidasi dan menandai token telah dipakai tepat satu kali.
func (t *UserToken) Consume(expectedPurpose TokenPurpose, now time.Time) error {
	if err := t.ValidateForUse(expectedPurpose, now); err != nil {
		return err
	}
	ts := normalizeLocalTime(now)
	t.Status = TokenStatusConsumed
	t.UsedAt = &ts
	return nil
}

// Expire menandai token aktif sebagai kedaluwarsa karena digantikan oleh token baru.
func (t *UserToken) Expire() {
	if t == nil || t.Status != TokenStatusActive {
		return
	}
	t.Status = TokenStatusExpired
}

// Revoke membatalkan token aktif sebelum digunakan.
func (t *UserToken) Revoke() {
	if t == nil || t.Status != TokenStatusActive {
		return
	}
	t.Status = TokenStatusRevoked
}

// String menyembunyikan digest token pada keluaran string.
func (t UserToken) String() string {
	return fmt.Sprintf("UserToken{ID:%s, UserID:%s, Purpose:%s, Status:%s, Used:%t, TokenDigest:%s}",
		t.ID, t.UserID, t.Purpose, t.Status, t.UsedAt != nil, redactedSecret)
}

// LogValue memancarkan atribut aman token untuk slog tanpa digest mentah.
func (t UserToken) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", t.ID),
		slog.String("user_id", t.UserID),
		slog.String("purpose", t.Purpose.String()),
		slog.String("status", t.Status.String()),
		slog.Bool("used", t.UsedAt != nil),
		slog.String("token_digest", redactedSecret),
	)
}
