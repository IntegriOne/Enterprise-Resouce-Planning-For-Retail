package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
)

// SecurityConfig menyimpan konfigurasi kriptografi TOTP/CSRF, pembuktian internal Gateway, pembatasan login, dan allowlist rute redirect IAM.
type SecurityConfig struct {
	TOTPIssuer                string
	TOTPActiveKeyVersion      int16
	TOTPKeyring               map[int16][]byte
	AllowEphemeralDevKey      bool
	CSRFActiveKeyVersion      int
	CSRFKeyring               map[int][]byte
	AllowEphemeralCSRFKey     bool
	CSRFKeyringConfigured     bool
	InternalGatewayHMACSecret []byte
	MaxFailedAttempts         int
	LockDuration              time.Duration
	AllowedRedirectPaths      []string
}

// LoadSecurity membaca konfigurasi keamanan IAM dan gagal tertutup pada lingkungan non-pengembangan tanpa IAM_TOTP_KEYRING.
func LoadSecurity(appEnv string) (SecurityConfig, error) {
	issuer := strings.TrimSpace(getEnv("IAM_TOTP_ISSUER", "IntegriOne ERP"))
	if issuer == "" {
		issuer = "IntegriOne ERP"
	}

	activeVerInt, err := getEnvAsInt("IAM_TOTP_ACTIVE_KEY_VERSION", 1)
	if err != nil || activeVerInt <= 0 || activeVerInt > 32767 {
		return SecurityConfig{}, fmt.Errorf("invalid IAM_TOTP_ACTIVE_KEY_VERSION: must be positive int16")
	}

	csrfActiveVer, err := getEnvAsInt("IAM_CSRF_ACTIVE_KEY_VERSION", 1)
	if err != nil || csrfActiveVer <= 0 || csrfActiveVer > 32767 {
		return SecurityConfig{}, fmt.Errorf("invalid IAM_CSRF_ACTIVE_KEY_VERSION: must be positive int16")
	}

	maxFailed, err := getEnvAsInt("IAM_MAX_FAILED_LOGIN_ATTEMPTS", entity.DefaultMaxFailedLoginAttempts)
	if err != nil || maxFailed <= 0 {
		return SecurityConfig{}, fmt.Errorf("invalid IAM_MAX_FAILED_LOGIN_ATTEMPTS: must be positive")
	}

	lockSec, err := getEnvAsInt("IAM_LOGIN_LOCK_DURATION_SECONDS", int(entity.DefaultLoginLockDuration.Seconds()))
	if err != nil || lockSec <= 0 {
		return SecurityConfig{}, fmt.Errorf("invalid IAM_LOGIN_LOCK_DURATION_SECONDS: must be positive")
	}

	var allowedPaths []string
	if rawPaths := strings.TrimSpace(getEnv("IAM_ALLOWED_REDIRECT_PATHS", "")); rawPaths != "" {
		for _, p := range strings.Split(rawPaths, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				allowedPaths = append(allowedPaths, trimmed)
			}
		}
	}

	envNormalized := strings.ToLower(strings.TrimSpace(appEnv))
	isDevOrTest := envNormalized == "development" || envNormalized == "test" || envNormalized == ""

	rawKeyring := strings.TrimSpace(getEnv("IAM_TOTP_KEYRING", ""))
	var totpKeys map[int16][]byte
	allowEphemeralTOTP := false
	if rawKeyring == "" {
		if !isDevOrTest {
			return SecurityConfig{}, fmt.Errorf("IAM_TOTP_KEYRING is required in %q environment", appEnv)
		}
		allowEphemeralTOTP = true
	} else {
		parsed, err := crypto.ParseKeyringSpec(rawKeyring)
		if err != nil {
			return SecurityConfig{}, fmt.Errorf("invalid IAM_TOTP_KEYRING: %w", err)
		}
		activeVer := int16(activeVerInt)
		if _, ok := parsed[activeVer]; !ok {
			return SecurityConfig{}, fmt.Errorf("active IAM_TOTP_ACTIVE_KEY_VERSION=%d is missing from IAM_TOTP_KEYRING", activeVer)
		}
		totpKeys = parsed
	}

	rawCSRFKeyring := strings.TrimSpace(getEnv("IAM_CSRF_HMAC_KEYRING", ""))
	var csrfKeys map[int][]byte
	csrfConfigured := false
	allowEphemeralCSRF := isDevOrTest
	if rawCSRFKeyring != "" {
		parsedCSRF, err := crypto.ParseKeyringSpec(rawCSRFKeyring)
		if err != nil {
			return SecurityConfig{}, fmt.Errorf("invalid IAM_CSRF_HMAC_KEYRING: %w", err)
		}
		csrfKeys = make(map[int][]byte, len(parsedCSRF))
		for k, v := range parsedCSRF {
			csrfKeys[int(k)] = v
		}
		if _, ok := csrfKeys[csrfActiveVer]; !ok {
			return SecurityConfig{}, fmt.Errorf("active IAM_CSRF_ACTIVE_KEY_VERSION=%d is missing from IAM_CSRF_HMAC_KEYRING", csrfActiveVer)
		}
		csrfConfigured = true
		allowEphemeralCSRF = false
	}

	internalHMAC := []byte(strings.TrimSpace(getEnv("IAM_INTERNAL_GATEWAY_HMAC_SECRET", "")))

	return SecurityConfig{
		TOTPIssuer:                issuer,
		TOTPActiveKeyVersion:      int16(activeVerInt),
		TOTPKeyring:               totpKeys,
		AllowEphemeralDevKey:      allowEphemeralTOTP,
		CSRFActiveKeyVersion:      csrfActiveVer,
		CSRFKeyring:               csrfKeys,
		AllowEphemeralCSRFKey:     allowEphemeralCSRF,
		CSRFKeyringConfigured:     csrfConfigured,
		InternalGatewayHMACSecret: internalHMAC,
		MaxFailedAttempts:         maxFailed,
		LockDuration:              time.Duration(lockSec) * time.Second,
		AllowedRedirectPaths:      allowedPaths,
	}, nil
}
