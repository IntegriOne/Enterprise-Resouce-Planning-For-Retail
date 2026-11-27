package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/config"
)

func TestLoad(t *testing.T) {
	// Tidak menggunakan t.Parallel karena memakai t.Setenv.
	setCleanEnv := func(t *testing.T, overrides map[string]string) {
		t.Helper()
		keys := []string{
			"APP_ENV",
			"EF_SERVICE_NAME",
			"EF_GRPC_HOST",
			"EF_GRPC_PORT",
			"EF_SHUTDOWN_TIMEOUT_SECONDS",
			"EF_DATABASE_DSN",
			"EF_TEST_DATABASE_DSN",
			"EF_INTERNAL_GATEWAY_HMAC_SECRET",
			"EF_OUTBOX_MODE",
			"EF_OUTBOX_WEBHOOK_URL",
			"EF_OUTBOX_WEBHOOK_HMAC_SECRET",
			"EF_OUTBOX_WEBHOOK_REQUIRE_HTTPS",
			"EF_OUTBOX_WEBHOOK_TIMEOUT_SECONDS",
			"EF_OUTBOX_POLL_INTERVAL_MS",
			"EF_OUTBOX_BATCH_SIZE",
			"EF_OUTBOX_MAX_ATTEMPTS",
			"EF_OUTBOX_PUBLISH_TIMEOUT_SECONDS",
			"EF_OUTBOX_CLEANUP_INTERVAL_MINUTES",
			"EF_OUTBOX_RETENTION_HOURS",
		}
		for _, k := range keys {
			t.Setenv(k, "")
		}
		// Hapus APP_ENV / defaultable vars agar perilaku fallback teruji eksplisit bila tidak di-override.
		for _, k := range keys {
			if _, ok := overrides[k]; ok {
				t.Setenv(k, overrides[k])
			}
		}
	}

	t.Run("defaults when optional values omitted", func(t *testing.T) {
		setCleanEnv(t, map[string]string{
			"APP_ENV":         "development",
			"EF_SERVICE_NAME": "integrione-enterprise-foundation",
			"EF_GRPC_HOST":    "0.0.0.0",
		})

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("config.Load() unexpected error: %v", err)
		}
		if cfg.GRPCPort != 9002 {
			t.Fatalf("got GRPCPort=%d, expected 9002", cfg.GRPCPort)
		}
		if cfg.ShutdownTimeout != 10*time.Second {
			t.Fatalf("got ShutdownTimeout=%v, expected 10s", cfg.ShutdownTimeout)
		}
		if cfg.GRPCAddress() != "0.0.0.0:9002" {
			t.Fatalf("got GRPCAddress=%q, expected 0.0.0.0:9002", cfg.GRPCAddress())
		}
		if !cfg.IsDevelopment() || cfg.IsProduction() {
			t.Fatalf("unexpected environment flags: dev=%v prod=%v", cfg.IsDevelopment(), cfg.IsProduction())
		}
	})

	t.Run("production succeeds with valid 32+ byte gateway secret", func(t *testing.T) {
		secret := strings.Repeat("a", 32)
		setCleanEnv(t, map[string]string{
			"APP_ENV":                         "production",
			"EF_SERVICE_NAME":                 "integrione-ef-prod",
			"EF_GRPC_HOST":                    "127.0.0.1",
			"EF_GRPC_PORT":                    "9102",
			"EF_SHUTDOWN_TIMEOUT_SECONDS":     "15",
			"EF_INTERNAL_GATEWAY_HMAC_SECRET": secret,
		})

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("config.Load() unexpected error: %v", err)
		}
		if !cfg.IsProduction() {
			t.Fatal("expected IsProduction() == true")
		}
		if cfg.GRPCPort != 9102 || cfg.ShutdownTimeout != 15*time.Second {
			t.Fatalf("unexpected port/timeout: %d / %v", cfg.GRPCPort, cfg.ShutdownTimeout)
		}
	})

	t.Run("production fails closed without gateway HMAC secret", func(t *testing.T) {
		setCleanEnv(t, map[string]string{
			"APP_ENV":         "production",
			"EF_SERVICE_NAME": "integrione-enterprise-foundation",
			"EF_GRPC_HOST":    "0.0.0.0",
		})

		if _, err := config.Load(); err == nil {
			t.Fatal("expected production config without HMAC secret to fail")
		}
	})

	t.Run("short non-empty HMAC secret rejected even in development", func(t *testing.T) {
		setCleanEnv(t, map[string]string{
			"APP_ENV":                         "development",
			"EF_SERVICE_NAME":                 "integrione-enterprise-foundation",
			"EF_GRPC_HOST":                    "0.0.0.0",
			"EF_INTERNAL_GATEWAY_HMAC_SECRET": "too-short",
		})

		if _, err := config.Load(); err == nil {
			t.Fatal("expected short HMAC secret to fail validation")
		}
	})

	t.Run("invalid port rejected", func(t *testing.T) {
		for _, badPort := range []string{"0", "-1", "70000", "abc"} {
			setCleanEnv(t, map[string]string{
				"APP_ENV":         "development",
				"EF_SERVICE_NAME": "integrione-enterprise-foundation",
				"EF_GRPC_HOST":    "0.0.0.0",
				"EF_GRPC_PORT":    badPort,
			})
			if _, err := config.Load(); err == nil {
				t.Fatalf("expected invalid EF_GRPC_PORT=%q to fail", badPort)
			}
		}
	})

	t.Run("outbox defaults to standby and validates webhook or log policies", func(t *testing.T) {
		setCleanEnv(t, map[string]string{
			"APP_ENV":         "development",
			"EF_SERVICE_NAME": "integrione-enterprise-foundation",
			"EF_GRPC_HOST":    "0.0.0.0",
		})
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("default Load failed: %v", err)
		}
		if cfg.ResolvedOutboxMode() != "standby" {
			t.Fatalf("expected default outbox mode standby, got %q", cfg.ResolvedOutboxMode())
		}

		// Webhook URL otomatis memilih mode webhook saat EF_OUTBOX_MODE kosong.
		setCleanEnv(t, map[string]string{
			"APP_ENV":                         "production",
			"EF_SERVICE_NAME":                 "integrione-enterprise-foundation",
			"EF_GRPC_HOST":                    "0.0.0.0",
			"EF_INTERNAL_GATEWAY_HMAC_SECRET": strings.Repeat("b", 32),
			"EF_OUTBOX_WEBHOOK_URL":           "https://events.internal.local/foundation",
			"EF_OUTBOX_WEBHOOK_REQUIRE_HTTPS": "true",
			"EF_OUTBOX_WEBHOOK_HMAC_SECRET":   strings.Repeat("c", 32),
			"EF_OUTBOX_BATCH_SIZE":            "50",
		})
		cfg, err = config.Load()
		if err != nil {
			t.Fatalf("production webhook Load failed: %v", err)
		}
		if cfg.ResolvedOutboxMode() != "webhook" || cfg.Outbox.BatchSize != 50 {
			t.Fatalf("unexpected webhook config: mode=%s batch=%d", cfg.ResolvedOutboxMode(), cfg.Outbox.BatchSize)
		}

		// Mode log wajib ditolak pada production.
		setCleanEnv(t, map[string]string{
			"APP_ENV":                         "production",
			"EF_SERVICE_NAME":                 "integrione-enterprise-foundation",
			"EF_GRPC_HOST":                    "0.0.0.0",
			"EF_INTERNAL_GATEWAY_HMAC_SECRET": strings.Repeat("b", 32),
			"EF_OUTBOX_MODE":                  "log",
		})
		if _, err := config.Load(); err == nil {
			t.Fatal("expected EF_OUTBOX_MODE=log to fail in production")
		}

		// Skema HTTP wajib ditolak jika EF_OUTBOX_WEBHOOK_REQUIRE_HTTPS=true.
		setCleanEnv(t, map[string]string{
			"APP_ENV":                         "production",
			"EF_SERVICE_NAME":                 "integrione-enterprise-foundation",
			"EF_GRPC_HOST":                    "0.0.0.0",
			"EF_INTERNAL_GATEWAY_HMAC_SECRET": strings.Repeat("b", 32),
			"EF_OUTBOX_MODE":                  "webhook",
			"EF_OUTBOX_WEBHOOK_URL":           "http://insecure.local/hook",
			"EF_OUTBOX_WEBHOOK_REQUIRE_HTTPS": "true",
		})
		if _, err := config.Load(); err == nil {
			t.Fatal("expected non-https webhook URL to fail when HTTPS is required")
		}
	})
}
