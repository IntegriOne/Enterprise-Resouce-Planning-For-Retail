package service

import (
	"strings"
	"unicode"
	"unicode/utf8"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
)

const (
	// DefaultMinPasswordLength adalah panjang minimum kata sandi standar 12 karakter.
	DefaultMinPasswordLength = 12
	// DefaultMaxPasswordLength adalah batas maksimum panjang kata sandi untuk mencegah DoS pada Argon2id.
	DefaultMaxPasswordLength = 128
	minForbiddenTokenLen     = 3
)

// PasswordPolicyConfig mendefinisikan parameter kebijakan kekuatan kata sandi.
type PasswordPolicyConfig struct {
	MinLength     int
	MaxLength     int
	RequireUpper  bool
	RequireLower  bool
	RequireDigit  bool
	RequireSymbol bool
}

// PasswordPolicy mengevaluasi kelayakan dan kekuatan kata sandi di lapisan domain.
type PasswordPolicy struct {
	cfg PasswordPolicyConfig
}

// NewDefaultPasswordPolicy mengembalikan kebijakan kata sandi standar IntegriOne (12..128 karakter, huruf besar/kecil, angka, dan simbol).
func NewDefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		cfg: PasswordPolicyConfig{
			MinLength:     DefaultMinPasswordLength,
			MaxLength:     DefaultMaxPasswordLength,
			RequireUpper:  true,
			RequireLower:  true,
			RequireDigit:  true,
			RequireSymbol: true,
		},
	}
}

// NewPasswordPolicy membentuk PasswordPolicy dengan batas bawaan aman jika parameter bernilai nol/tidak valid.
func NewPasswordPolicy(cfg PasswordPolicyConfig) PasswordPolicy {
	if cfg.MinLength <= 0 {
		cfg.MinLength = DefaultMinPasswordLength
	}
	if cfg.MaxLength < cfg.MinLength {
		cfg.MaxLength = DefaultMaxPasswordLength
	}
	return PasswordPolicy{cfg: cfg}
}

// Validate memeriksa panjang, kompleksitas karakter, ketiadaan byte NUL, serta keterkaitan kata sandi dengan identitas akun.
func (p PasswordPolicy) Validate(password string, forbiddenTerms ...string) error {
	if strings.ContainsRune(password, '\x00') {
		return iamerrors.ErrPasswordTooWeak
	}

	length := utf8.RuneCountInString(password)
	if length < p.cfg.MinLength {
		return iamerrors.ErrPasswordTooShort
	}
	if length > p.cfg.MaxLength {
		return iamerrors.ErrPasswordTooLong
	}

	if err := p.checkComplexity(password); err != nil {
		return err
	}

	return checkForbiddenIdentityTerms(password, forbiddenTerms)
}

func (p PasswordPolicy) checkComplexity(password string) error {
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	if (p.cfg.RequireUpper && !hasUpper) ||
		(p.cfg.RequireLower && !hasLower) ||
		(p.cfg.RequireDigit && !hasDigit) ||
		(p.cfg.RequireSymbol && !hasSymbol) {
		return iamerrors.ErrPasswordTooWeak
	}
	return nil
}

func checkForbiddenIdentityTerms(password string, forbiddenTerms []string) error {
	lowerPassword := strings.ToLower(password)
	for _, rawTerm := range forbiddenTerms {
		term := strings.ToLower(strings.TrimSpace(rawTerm))
		if utf8.RuneCountInString(term) >= minForbiddenTokenLen && strings.Contains(lowerPassword, term) {
			return iamerrors.ErrPasswordContainsIdentity
		}
		tokens := strings.FieldsFunc(term, func(r rune) bool {
			return r == '@' || r == '.' || r == '_' || r == '-' || unicode.IsSpace(r)
		})
		for _, tok := range tokens {
			if utf8.RuneCountInString(tok) >= minForbiddenTokenLen && strings.Contains(lowerPassword, tok) {
				return iamerrors.ErrPasswordContainsIdentity
			}
		}
	}
	return nil
}
