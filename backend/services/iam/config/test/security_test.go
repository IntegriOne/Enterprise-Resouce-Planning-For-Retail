package config_test

import (
	"bytes"
	"encoding/base64"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/config"
)

func TestLoadSecurity_EnvironmentAndKeyringPolicies(t *testing.T) {
	// Tidak menggunakan t.Parallel karena memakai t.Setenv.

	// Mode production wajib gagal tertutup bila IAM_TOTP_KEYRING tidak disetel.
	t.Setenv("IAM_TOTP_KEYRING", "")
	t.Setenv("IAM_TOTP_ACTIVE_KEY_VERSION", "")
	if _, err := config.LoadSecurity("production"); err == nil {
		t.Fatal("expected LoadSecurity(production) to fail when IAM_TOTP_KEYRING is empty")
	}

	// Mode development mengizinkan kunci ephemeral bila env kosong.
	devCfg, err := config.LoadSecurity("development")
	if err != nil {
		t.Fatalf("LoadSecurity(development) error = %v", err)
	}
	if !devCfg.AllowEphemeralDevKey {
		t.Fatal("expected AllowEphemeralDevKey=true in development without IAM_TOTP_KEYRING")
	}

	// Mode production dengan keyring terkonfigurasi berhasil dimuat.
	rawKey := bytes.Repeat([]byte{0x7a}, 32)
	spec := "2:" + base64.StdEncoding.EncodeToString(rawKey)
	t.Setenv("IAM_TOTP_KEYRING", spec)
	t.Setenv("IAM_TOTP_ACTIVE_KEY_VERSION", "2")
	t.Setenv("IAM_MAX_FAILED_LOGIN_ATTEMPTS", "7")
	t.Setenv("IAM_LOGIN_LOCK_DURATION_SECONDS", "1200")
	t.Setenv("IAM_TOTP_ISSUER", "IntegriOne Prod")

	prodCfg, err := config.LoadSecurity("production")
	if err != nil {
		t.Fatalf("LoadSecurity(production with keyring) error = %v", err)
	}
	if prodCfg.AllowEphemeralDevKey {
		t.Fatal("expected AllowEphemeralDevKey=false when explicit keyring configured")
	}
	if prodCfg.TOTPActiveKeyVersion != 2 || len(prodCfg.TOTPKeyring[2]) != 32 {
		t.Fatalf("unexpected loaded keyring: ver=%d len=%d", prodCfg.TOTPActiveKeyVersion, len(prodCfg.TOTPKeyring[2]))
	}
	if prodCfg.MaxFailedAttempts != 7 || prodCfg.LockDuration != 20*time.Minute {
		t.Fatalf("unexpected lockout config: %+v", prodCfg)
	}
	if prodCfg.TOTPIssuer != "IntegriOne Prod" {
		t.Fatalf("unexpected TOTP issuer: %q", prodCfg.TOTPIssuer)
	}
}

func TestLoadAndValidateRuntimeStrict_ProductionPolicies(t *testing.T) {
	rawKey := bytes.Repeat([]byte{0x7a}, 32)
	spec := "1:" + base64.StdEncoding.EncodeToString(rawKey)

	t.Setenv("APP_ENV", "production")
	t.Setenv("IAM_TOTP_KEYRING", spec)
	t.Setenv("IAM_TOTP_ACTIVE_KEY_VERSION", "1")
	t.Setenv("IAM_CSRF_HMAC_KEYRING", "")
	t.Setenv("IAM_REDIS_ENABLED", "false")
	t.Setenv("IAM_INTERNAL_GATEWAY_HMAC_SECRET", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	secCfg, err := config.LoadSecurity("production")
	if err != nil {
		t.Fatalf("LoadSecurity() error = %v", err)
	}
	if err := cfg.ValidateRuntimeStrict(secCfg); err == nil {
		t.Fatal("expected ValidateRuntimeStrict to fail in production without Redis/CSRF/HMAC")
	}

	csrfKey := bytes.Repeat([]byte{0x55}, 32)
	t.Setenv("IAM_CSRF_HMAC_KEYRING", "1:"+base64.StdEncoding.EncodeToString(csrfKey))
	t.Setenv("IAM_CSRF_ACTIVE_KEY_VERSION", "1")
	t.Setenv("IAM_REDIS_ENABLED", "true")
	t.Setenv("IAM_REDIS_ADDR", "redis:6379")
	t.Setenv("IAM_REDIS_PASSWORD", "secret-redis-password")
	t.Setenv("IAM_REDIS_CA_CERT_PATH", "/etc/certs/ca.crt")
	t.Setenv("IAM_INTERNAL_GATEWAY_HMAC_SECRET", "0123456789abcdef0123456789abcdef")

	cfgProd, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() prod error = %v", err)
	}
	secProd, err := config.LoadSecurity("production")
	if err != nil {
		t.Fatalf("LoadSecurity() prod error = %v", err)
	}
	if err := cfgProd.ValidateRuntimeStrict(secProd); err != nil {
		t.Fatalf("expected ValidateRuntimeStrict to succeed in configured production, got: %v", err)
	}
}

func TestValidateRuntimeStrict_DevelopmentRequiresInternalProofSecret(t *testing.T) {
	cfg := &config.Config{AppEnv: "development"}
	if err := cfg.ValidateRuntimeStrict(config.SecurityConfig{}); err == nil {
		t.Fatal("development runtime must reject missing internal HMAC secret")
	}
	cfg.InternalGatewayHMACSecret = "0123456789abcdef0123456789abcdef"
	if err := cfg.ValidateRuntimeStrict(config.SecurityConfig{}); err != nil {
		t.Fatalf("development runtime rejected configured internal HMAC secret: %v", err)
	}
}
