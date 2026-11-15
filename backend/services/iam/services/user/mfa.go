package user

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

// EnrollTOTP membuat secret TOTP terenkripsi AES-GCM baru dan mengembalikan URI otpauth:// sekali tampil.
func (s *Service) EnrollTOTP(ctx context.Context, userID, accountLabel string) (*EnrollTOTPResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}

	// Pastikan hanya pengguna berstatus ACTIVE yang dapat mendaftarkan faktor TOTP secara mandiri.
	u, err := s.loadAuthenticatableUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	label := strings.TrimSpace(accountLabel)
	if label == "" {
		label = u.DisplayName
	}

	secret, uri, err := s.totpProvider.GenerateSecret(label, s.cfg.TOTPIssuer)
	if err != nil {
		return nil, err
	}
	defer clear(secret)

	ciphertext, keyVersion, err := s.secretCipher.Encrypt(secret)
	if err != nil {
		return nil, err
	}
	factorID, err := s.idGen.NewUUID()
	if err != nil {
		return nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.totp.id_generation_failed"), "failed to generate totp factor id")
	}

	factor, err := entity.NewUserTOTPFactor(factorID, userID, ciphertext, keyVersion, s.clock.Now())
	if err != nil {
		return nil, err
	}

	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		existing, err := s.totpFactors.SelectByUserID(txCtx, userID)
		if err == nil && existing != nil {
			if existing.IsEnabled {
				return iamerrors.ErrTOTPAlreadyExists
			}
			if delErr := s.totpFactors.DeleteByUserID(txCtx, userID); delErr != nil {
				return delErr
			}
		} else if err != nil && !errors.Is(err, iamerrors.ErrTOTPNotFound) {
			return err
		}

		return s.totpFactors.Insert(txCtx, factor)
	})
	if err != nil {
		return nil, err
	}

	return &EnrollTOTPResult{
		FactorID:   factor.ID,
		OTPAuthURI: uri,
	}, nil
}

// ConfirmTOTP memvalidasi kode OTP pertama, mengaktifkan 2FA dengan step anti-replay, dan menerbitkan kode pemulihan.
func (s *Service) ConfirmTOTP(ctx context.Context, userID, code string) (*ConfirmTOTPResult, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}

	var confirmedFactor *entity.UserTOTPFactor
	var plainCodes []string
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Pastikan hanya pengguna berstatus ACTIVE yang dapat mengonfirmasi faktor TOTP secara mandiri.
		if _, err := s.loadAuthenticatableUser(txCtx, userID); err != nil {
			return err
		}
		factor, err := s.totpFactors.SelectByUserID(txCtx, userID)
		if err != nil {
			return err
		}
		if factor.IsEnabled {
			return iamerrors.ErrTOTPAlreadyConfirmed
		}

		secret, err := s.secretCipher.Decrypt(factor.EncryptedSecret(), factor.KeyVersion)
		if err != nil {
			return err
		}
		defer clear(secret)

		now := s.clock.Now()
		matched, step, err := s.totpProvider.VerifyCode(secret, code, now)
		if err != nil {
			return err
		}
		if !matched {
			return iamerrors.ErrInvalidTOTPCode
		}

		if err := factor.Confirm(now); err != nil {
			return err
		}
		if err := factor.MarkStepUsed(step, now); err != nil {
			return err
		}

		codes, entities, err := s.buildRecoveryCodes(userID, now)
		if err != nil {
			return err
		}
		if err := s.totpFactors.Update(txCtx, factor); err != nil {
			return err
		}
		if err := s.recoveryCodes.DeleteByUserID(txCtx, userID); err != nil {
			return err
		}
		if err := s.recoveryCodes.InsertBatch(txCtx, entities); err != nil {
			return err
		}
		if err := s.recordUserSecurityChangedInTx(txCtx, userID, false, true); err != nil {
			return err
		}
		confirmedFactor = factor
		plainCodes = codes
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ConfirmTOTPResult{
		Factor:        confirmedFactor,
		RecoveryCodes: plainCodes,
	}, nil
}

// DisableTOTP menghapus faktor TOTP dan seluruh kode pemulihan milik pengguna dalam satu transaksi.
func (s *Service) DisableTOTP(ctx context.Context, userID string) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return err
	}

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Pastikan hanya pengguna berstatus ACTIVE yang dapat menonaktifkan faktor TOTP secara mandiri.
		if _, err := s.loadAuthenticatableUser(txCtx, userID); err != nil {
			return err
		}
		if _, err := s.totpFactors.SelectByUserID(txCtx, userID); err != nil {
			return err
		}
		if err := s.totpFactors.DeleteByUserID(txCtx, userID); err != nil {
			return err
		}
		if err := s.recoveryCodes.DeleteByUserID(txCtx, userID); err != nil {
			return err
		}
		return s.recordUserSecurityChangedInTx(txCtx, userID, false, true)
	})
}

// RegenerateRecoveryCodes mengganti seluruh kode cadangan pemulihan dengan kumpulan kode acak baru bagi pengguna yang 2FA-nya aktif.
func (s *Service) RegenerateRecoveryCodes(ctx context.Context, userID string) ([]string, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return nil, err
	}

	// Pastikan hanya pengguna berstatus ACTIVE yang dapat meregenerasi kode pemulihan secara mandiri.
	if _, err := s.loadAuthenticatableUser(ctx, userID); err != nil {
		return nil, err
	}

	factor, err := s.totpFactors.SelectByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !factor.IsEnabled || factor.ConfirmedAt == nil {
		return nil, iamerrors.ErrTOTPNotConfirmed
	}

	now := s.clock.Now()
	plainCodes, entities, err := s.buildRecoveryCodes(userID, now)
	if err != nil {
		return nil, err
	}

	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.recoveryCodes.DeleteByUserID(txCtx, userID); err != nil {
			return err
		}
		if err := s.recoveryCodes.InsertBatch(txCtx, entities); err != nil {
			return err
		}
		return s.recordUserSecurityChangedInTx(txCtx, userID, false, true)
	})
	if err != nil {
		return nil, err
	}
	return plainCodes, nil
}

// CountUnusedRecoveryCodes menghitung sisa kode cadangan pemulihan yang belum dipakai pengguna.
func (s *Service) CountUnusedRecoveryCodes(ctx context.Context, userID string) (int64, error) {
	if err := s.ensureReady(); err != nil {
		return 0, err
	}
	if err := entity.ValidateUUID("user_id", userID); err != nil {
		return 0, err
	}
	return s.recoveryCodes.CountUnusedByUserID(ctx, userID)
}

func (s *Service) buildRecoveryCodes(userID string, now time.Time) ([]string, []*entity.UserRecoveryCode, error) {
	plainCodes, err := s.tokenGen.GenerateRecoveryCodes(s.cfg.RecoveryCodeCount)
	if err != nil {
		return nil, nil, err
	}

	entities := make([]*entity.UserRecoveryCode, 0, len(plainCodes))
	for _, code := range plainCodes {
		rcID, err := s.idGen.NewUUID()
		if err != nil {
			return nil, nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.recovery_code.id_generation_failed"), "failed to generate recovery code id")
		}
		digest := s.tokenGen.Digest(strings.ToUpper(code))
		rc, err := entity.NewUserRecoveryCode(rcID, userID, hex.EncodeToString(digest), now)
		if err != nil {
			return nil, nil, err
		}
		entities = append(entities, rc)
	}
	return plainCodes, entities, nil
}
