package config

import (
	"testing"
	"time"
)

func setMinimalStrictGatewayEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GATEWAY_REDIS_ENABLED", "true")
	t.Setenv("GATEWAY_REDIS_ADDR", "127.0.0.1:6379")
	t.Setenv("GATEWAY_REDIS_PASSWORD", "strong-redis-password-12345")
	t.Setenv("GATEWAY_REDIS_CA_CERT_PATH", "../../deploy/certs/ca.crt")
	t.Setenv("GATEWAY_INTERNAL_HMAC_SECRET", "12345678901234567890123456789012")
}

// TestLoad_EnterpriseFoundationConfig memverifikasi pemuatan konfigurasi gRPC Enterprise Foundation yang aman dan terisolasi.
func TestLoad_EnterpriseFoundationConfig(t *testing.T) {
	t.Run("default values when env unset", func(t *testing.T) {
		setMinimalStrictGatewayEnv(t)
		t.Setenv("EF_GRPC_ENABLED", "")
		t.Setenv("EF_GRPC_TARGET", "")
		t.Setenv("EF_GRPC_TIMEOUT_SECONDS", "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() unexpected error: %v", err)
		}
		if !cfg.EFGRPCEnabled {
			t.Fatalf("expected default EFGRPCEnabled=true, got false")
		}
		if cfg.EFGRPCTarget != "127.0.0.1:9002" {
			t.Fatalf("expected default EFGRPCTarget=127.0.0.1:9002, got %q", cfg.EFGRPCTarget)
		}
		if cfg.EFGRPCTimeout != 5*time.Second {
			t.Fatalf("expected default EFGRPCTimeout=5s, got %s", cfg.EFGRPCTimeout)
		}
	})

	t.Run("disabled enterprise foundation does not block ValidateRuntimeStrict", func(t *testing.T) {
		setMinimalStrictGatewayEnv(t)
		t.Setenv("APP_ENV", "production")
		t.Setenv("GATEWAY_ENABLE_INTROSPECTION", "false")
		t.Setenv("GATEWAY_ALLOWED_ORIGINS", "https://erp.integrione.id")
		t.Setenv("GATEWAY_COOKIE_SECURE", "true")
		t.Setenv("GATEWAY_REDIS_TLS_INSECURE_SKIP_VERIFY", "false")
		t.Setenv("EF_GRPC_ENABLED", "false")
		t.Setenv("EF_GRPC_TARGET", "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() unexpected error: %v", err)
		}
		if cfg.EFGRPCEnabled {
			t.Fatalf("expected EFGRPCEnabled=false")
		}
		if err := cfg.ValidateRuntimeStrict(); err != nil {
			t.Fatalf("ValidateRuntimeStrict() unexpected error when EF disabled: %v", err)
		}
	})

	t.Run("invalid timeout returns error", func(t *testing.T) {
		setMinimalStrictGatewayEnv(t)
		t.Setenv("EF_GRPC_TIMEOUT_SECONDS", "0")
		if _, err := Load(); err == nil {
			t.Fatalf("expected error for non-positive EF_GRPC_TIMEOUT_SECONDS")
		}
	})
}
