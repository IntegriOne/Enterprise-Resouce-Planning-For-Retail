package user

import (
	"bytes"
	"strings"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/pgstore"
)

func credentialToDomain(m *userCredentialModel) (*entity.UserCredential, error) {
	if m == nil {
		return nil, iamerrors.ErrCredentialNotFound
	}
	return entity.RehydrateUserCredential(
		strings.ToLower(strings.TrimSpace(m.UserID)),
		m.PasswordHash,
		m.MustChangePassword,
		pgstore.NormalizeLocalTime(m.PasswordChangedAt),
		m.FailedLoginAttempts,
		pgstore.NormalizeLocalTimePtr(m.LockedUntil),
		pgstore.NormalizeLocalTime(m.CreatedAt),
		pgstore.NormalizeLocalTime(m.UpdatedAt),
	)
}

func credentialToModel(c *entity.UserCredential) (userCredentialModel, error) {
	if c == nil {
		return userCredentialModel{}, iamerrors.NewValidation("credential", iamerrors.Code("iam.credential.nil"), "user credential must not be nil")
	}
	userID, err := pgstore.ValidateUUIDParam("user_id", c.UserID)
	if err != nil {
		return userCredentialModel{}, err
	}
	if strings.TrimSpace(c.PasswordHash()) == "" {
		return userCredentialModel{}, iamerrors.NewValidation("password_hash", iamerrors.Code("iam.credential.empty_hash"), "password hash must not be empty")
	}
	return userCredentialModel{
		UserID:              userID,
		PasswordHash:        c.PasswordHash(),
		PasswordChangedAt:   pgstore.MustLocalNow(c.PasswordChangedAt),
		MustChangePassword:  c.MustChangePassword,
		FailedLoginAttempts: c.FailedLoginAttempts,
		LockedUntil:         pgstore.NormalizeLocalTimePtr(c.LockedUntil),
		CreatedAt:           pgstore.MustLocalNow(c.CreatedAt),
		UpdatedAt:           pgstore.MustLocalNow(c.UpdatedAt),
	}, nil
}

func totpToDomain(m *userTOTPFactorModel) (*entity.UserTOTPFactor, error) {
	if m == nil {
		return nil, iamerrors.ErrTOTPNotFound
	}
	return entity.RehydrateUserTOTPFactor(
		strings.ToLower(strings.TrimSpace(m.ID)),
		strings.ToLower(strings.TrimSpace(m.UserID)),
		m.EncryptedSecret,
		m.KeyVersion,
		m.IsEnabled,
		pgstore.NormalizeLocalTimePtr(m.ConfirmedAt),
		m.LastUsedStep,
		pgstore.NormalizeLocalTime(m.CreatedAt),
		pgstore.NormalizeLocalTime(m.UpdatedAt),
	)
}

func totpToModel(f *entity.UserTOTPFactor) (userTOTPFactorModel, error) {
	if f == nil {
		return userTOTPFactorModel{}, iamerrors.NewValidation("totp_factor", iamerrors.Code("iam.totp.nil"), "totp factor must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", f.ID)
	if err != nil {
		return userTOTPFactorModel{}, err
	}
	userID, err := pgstore.ValidateUUIDParam("user_id", f.UserID)
	if err != nil {
		return userTOTPFactorModel{}, err
	}
	secret := f.EncryptedSecret()
	if len(secret) == 0 {
		return userTOTPFactorModel{}, iamerrors.NewValidation("encrypted_secret", iamerrors.Code("iam.totp.empty_secret"), "encrypted secret must not be empty")
	}
	return userTOTPFactorModel{
		ID:              id,
		UserID:          userID,
		EncryptedSecret: bytes.Clone(secret),
		KeyVersion:      f.KeyVersion,
		IsEnabled:       f.IsEnabled,
		ConfirmedAt:     pgstore.NormalizeLocalTimePtr(f.ConfirmedAt),
		LastUsedStep:    f.LastUsedStep,
		CreatedAt:       pgstore.MustLocalNow(f.CreatedAt),
		UpdatedAt:       pgstore.MustLocalNow(f.UpdatedAt),
	}, nil
}

func recoveryCodeToDomain(m *userRecoveryCodeModel) (*entity.UserRecoveryCode, error) {
	if m == nil {
		return nil, iamerrors.ErrRecoveryCodeNotFound
	}
	return entity.RehydrateUserRecoveryCode(
		strings.ToLower(strings.TrimSpace(m.ID)),
		strings.ToLower(strings.TrimSpace(m.UserID)),
		m.CodeHash,
		pgstore.NormalizeLocalTimePtr(m.UsedAt),
		pgstore.NormalizeLocalTime(m.CreatedAt),
	)
}

func recoveryCodeToModel(r *entity.UserRecoveryCode) (userRecoveryCodeModel, error) {
	if r == nil {
		return userRecoveryCodeModel{}, iamerrors.NewValidation("recovery_code", iamerrors.Code("iam.recovery_code.nil"), "recovery code must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", r.ID)
	if err != nil {
		return userRecoveryCodeModel{}, err
	}
	userID, err := pgstore.ValidateUUIDParam("user_id", r.UserID)
	if err != nil {
		return userRecoveryCodeModel{}, err
	}
	if strings.TrimSpace(r.CodeHash()) == "" {
		return userRecoveryCodeModel{}, iamerrors.NewValidation("code_hash", iamerrors.Code("iam.recovery_code.empty_hash"), "recovery code hash must not be empty")
	}
	return userRecoveryCodeModel{
		ID:        id,
		UserID:    userID,
		CodeHash:  r.CodeHash(),
		UsedAt:    pgstore.NormalizeLocalTimePtr(r.UsedAt),
		CreatedAt: pgstore.MustLocalNow(r.CreatedAt),
	}, nil
}

func tokenToDomain(m *userTokenModel) (*entity.UserToken, error) {
	if m == nil {
		return nil, iamerrors.ErrTokenNotFound
	}
	purpose, err := entity.ParseTokenPurpose(m.Purpose)
	if err != nil {
		return nil, err
	}
	status := entity.TokenStatus(m.Status)
	if strings.TrimSpace(m.Status) == "" {
		if m.ConsumedAt != nil {
			status = entity.TokenStatusConsumed
		} else {
			status = entity.TokenStatusActive
		}
	} else if !status.IsValid() {
		return nil, iamerrors.ErrInvalidTokenStatus
	}
	var identityID *string
	if m.IdentityID != nil && strings.TrimSpace(*m.IdentityID) != "" {
		normalized := strings.ToLower(strings.TrimSpace(*m.IdentityID))
		identityID = &normalized
	}
	return entity.RehydrateUserToken(
		strings.ToLower(strings.TrimSpace(m.ID)),
		strings.ToLower(strings.TrimSpace(m.UserID)),
		identityID,
		purpose,
		status,
		m.TokenDigest,
		pgstore.NormalizeLocalTime(m.ExpiresAt),
		pgstore.NormalizeLocalTimePtr(m.ConsumedAt),
		pgstore.NormalizeLocalTime(m.CreatedAt),
	)
}

func tokenToModel(t *entity.UserToken) (userTokenModel, error) {
	if t == nil {
		return userTokenModel{}, iamerrors.NewValidation("token", iamerrors.Code("iam.token.nil"), "user token must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", t.ID)
	if err != nil {
		return userTokenModel{}, err
	}
	userID, err := pgstore.ValidateUUIDParam("user_id", t.UserID)
	if err != nil {
		return userTokenModel{}, err
	}
	var identityID *string
	if t.IdentityID != nil && strings.TrimSpace(*t.IdentityID) != "" {
		validatedIdentityID, idErr := pgstore.ValidateUUIDParam("identity_id", *t.IdentityID)
		if idErr != nil {
			return userTokenModel{}, idErr
		}
		identityID = &validatedIdentityID
	}
	if !t.Purpose.IsValid() {
		return userTokenModel{}, iamerrors.ErrTokenInvalidPurpose
	}
	status := t.Status
	if status == "" {
		status = entity.TokenStatusActive
	}
	if !status.IsValid() {
		return userTokenModel{}, iamerrors.ErrInvalidTokenStatus
	}
	digest := t.TokenDigest()
	if len(digest) == 0 {
		return userTokenModel{}, iamerrors.ErrInvalidTokenDigest
	}
	return userTokenModel{
		ID:          id,
		UserID:      userID,
		IdentityID:  identityID,
		Purpose:     t.Purpose.String(),
		Status:      status.String(),
		TokenDigest: bytes.Clone(digest),
		ExpiresAt:   pgstore.NormalizeLocalTime(t.ExpiresAt),
		ConsumedAt:  pgstore.NormalizeLocalTimePtr(t.UsedAt),
		CreatedAt:   pgstore.MustLocalNow(t.CreatedAt),
	}, nil
}
