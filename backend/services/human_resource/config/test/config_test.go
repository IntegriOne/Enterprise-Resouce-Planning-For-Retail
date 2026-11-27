package config_test

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/config"
)

func validEnvironment(t *testing.T) {
	t.Helper()
	values := map[string]string{
		"APP_ENV": "test", "HRM_SERVICE_NAME": "integrione-human-resource",
		"HRM_GRPC_HOST": "127.0.0.1", "HRM_GRPC_PORT": "9003",
		"HRM_DB_HOST": "127.0.0.1", "HRM_DB_PORT": "6543", "HRM_DB_USER": "hrm",
		"HRM_DB_PASSWORD": "fixture:@/?&password", "HRM_DB_PASSWORD_FILE": "",
		"HRM_DB_NAME": "human_resource_test", "HRM_DB_SSLMODE": "disable",
		"HRM_DB_SSLROOTCERT": "", "HRM_DB_MAX_OPEN_CONNS": "20", "HRM_DB_MAX_IDLE_CONNS": "5",
		"HRM_DB_CONN_MAX_LIFETIME": "30m", "HRM_STARTUP_TIMEOUT": "60s", "HRM_SHUTDOWN_TIMEOUT": "10s",
		"HRM_GRPC_TLS_CERT_FILE": "/fixture/server.crt", "HRM_GRPC_TLS_KEY_FILE": "/fixture/server.key",
		"HRM_GRPC_TLS_CA_FILE": "/fixture/ca.crt", "HRM_GRPC_TLS_SERVER_NAME": "human-resource",
		"HRM_GATEWAY_HMAC_SECRET": strings.Repeat("f", 32), "HRM_GATEWAY_HMAC_SECRET_FILE": "",
		"HRM_REDIS_ADDR": "127.0.0.1:6379", "HRM_REDIS_USERNAME": "hrm_reader",
		"HRM_REDIS_PASSWORD": "redis-fixture", "HRM_REDIS_PASSWORD_FILE": "",
		"HRM_REDIS_CA_CERT_PATH": "/fixture/redis-ca.crt", "HRM_REDIS_SERVER_NAME": "redis",
		"HRM_REDIS_POOL_SIZE": "16", "HRM_REDIS_MIN_IDLE_CONNS": "0",
		"HRM_REDIS_DIAL_TIMEOUT": "2s", "HRM_REDIS_READ_TIMEOUT": "1s", "HRM_REDIS_WRITE_TIMEOUT": "1s",
		"HRM_SECURITY_NAMESPACE": "test",
		"HRM_OUTBOX_MODE":        "standby", "HRM_OUTBOX_WEBHOOK_URL": "",
		"HRM_OUTBOX_WEBHOOK_HMAC_SECRET": "", "HRM_OUTBOX_WEBHOOK_HMAC_SECRET_FILE": "",
		"HRM_OUTBOX_MAX_ATTEMPTS": "20", "HRM_OUTBOX_POLL_INTERVAL": "5s",
		"HRM_OUTBOX_PUBLISH_TIMEOUT": "5s",
	}
	for key, value := range values {
		t.Setenv(key, value)
	}
}

func TestOutboxConfigurationGuardsDeliveryMode(t *testing.T) {
	validEnvironment(t)
	t.Setenv("HRM_OUTBOX_MODE", "webhook")
	t.Setenv("HRM_OUTBOX_WEBHOOK_URL", "https://events.example.test/hrm")
	t.Setenv("HRM_OUTBOX_WEBHOOK_HMAC_SECRET", strings.Repeat("h", 32))
	if _, err := config.Load(); err != nil {
		t.Fatalf("valid webhook configuration rejected: %v", err)
	}
	t.Setenv("HRM_OUTBOX_WEBHOOK_URL", "http://events.example.test/hrm")
	t.Setenv("APP_ENV", "production")
	t.Setenv("HRM_DB_SSLMODE", "verify-full")
	t.Setenv("HRM_DB_SSLROOTCERT", "/fixture/postgres-ca.crt")
	if _, err := config.Load(); err == nil {
		t.Fatal("production accepted plaintext webhook")
	}
	t.Setenv("HRM_OUTBOX_MODE", "log")
	if _, err := config.Load(); err == nil {
		t.Fatal("production accepted log publisher")
	}
}

func TestLoadAndEscapedDatabaseURL(t *testing.T) {
	validEnvironment(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(cfg.Database.DSN())
	if err != nil {
		t.Fatal("invalid database URL")
	}
	password, _ := parsed.User.Password()
	if password != "fixture:@/?&password" || parsed.User.Username() != "hrm" || parsed.Path != "/human_resource_test" {
		t.Fatal("database credentials must remain exact after URL escaping")
	}
	if cfg.GRPCAddress() != "127.0.0.1:9003" || parsed.Query().Get("sslmode") != "disable" {
		t.Fatal("unexpected listener or database TLS configuration")
	}
}

func TestInvalidConfigurationFailsWithoutSecretDisclosure(t *testing.T) {
	for _, setting := range []struct{ key, value string }{
		{"HRM_GRPC_PORT", "0"}, {"HRM_DB_PORT", "65536"}, {"HRM_DB_MAX_IDLE_CONNS", "21"},
		{"HRM_STARTUP_TIMEOUT", "0s"}, {"HRM_SHUTDOWN_TIMEOUT", "not-duration"},
		{"HRM_GATEWAY_HMAC_SECRET", "fixture-private-short-secret"},
		{"HRM_GRPC_TLS_CA_FILE", ""}, {"HRM_DB_NAME", "iam"}, {"HRM_DB_NAME", "hrm_;DROP DATABASE iam"},
		{"HRM_DB_SSLMODE", "prefer"}, {"APP_ENV", "production"},
		{"HRM_REDIS_CA_CERT_PATH", ""}, {"HRM_REDIS_PASSWORD", ""}, {"HRM_REDIS_POOL_SIZE", "0"},
		{"HRM_REDIS_ADDR", "redis"}, {"HRM_SECURITY_NAMESPACE", "test:other"},
		{"HRM_REDIS_ADDR", ":6379"}, {"HRM_REDIS_ADDR", "redis:70000"},
		{"HRM_OUTBOX_MAX_ATTEMPTS", "0"}, {"HRM_OUTBOX_POLL_INTERVAL", "0s"},
	} {
		t.Run(setting.key+"_"+setting.value, func(t *testing.T) {
			validEnvironment(t)
			t.Setenv(setting.key, setting.value)
			_, err := config.Load()
			if err == nil {
				t.Fatal("invalid configuration accepted")
			}
			if strings.Contains(err.Error(), "fixture-private") || strings.Contains(err.Error(), "fixture:@/") {
				t.Fatal("configuration error exposed a secret")
			}
		})
	}
}

func TestSecretFileAndAmbiguousSecretSource(t *testing.T) {
	validEnvironment(t)
	file := filepath.Join(t.TempDir(), "db-password")
	if err := os.WriteFile(file, []byte("fixture file password\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HRM_DB_PASSWORD", "")
	t.Setenv("HRM_DB_PASSWORD_FILE", file)
	cfg, err := config.Load()
	if err != nil || cfg.Database.Password != "fixture file password" {
		t.Fatal("secret file was not loaded correctly")
	}
	t.Setenv("HRM_DB_PASSWORD", "ambiguous-fixture")
	if _, err := config.Load(); err == nil {
		t.Fatal("ambiguous secret sources accepted")
	}
}

func TestTerminationRelayRequiresConnectionForIAMCallback(t *testing.T) {
	validEnvironment(t)
	t.Setenv("HRM_TERMINATION_ACCESS_ADDR", "127.0.0.1:9012")
	t.Setenv("HRM_TERMINATION_ACCESS_CERT_FILE", "/fixture/termination.crt")
	t.Setenv("HRM_TERMINATION_ACCESS_KEY_FILE", "/fixture/termination.key")
	t.Setenv("HRM_TERMINATION_ACCESS_CLIENT_CA_FILE", "/fixture/termination-ca.crt")
	t.Setenv("HRM_IAM_MANDATE_TARGET", "iam:9011")
	t.Setenv("HRM_DB_MAX_IDLE_CONNS", "1")
	t.Setenv("HRM_DB_MAX_OPEN_CONNS", "1")
	if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), "termination delivery requires HRM_DB_MAX_OPEN_CONNS") {
		t.Fatalf("termination delivery did not reject a single-connection pool: %v", err)
	}
	t.Setenv("HRM_DB_MAX_OPEN_CONNS", "2")
	if _, err := config.Load(); err != nil {
		t.Fatalf("termination delivery rejected a two-connection pool: %v", err)
	}
	t.Setenv("HRM_IAM_MANDATE_TARGET", "")
	t.Setenv("HRM_DB_MAX_OPEN_CONNS", "1")
	if _, err := config.Load(); err != nil {
		t.Fatalf("standalone verifier rejected a single-connection pool: %v", err)
	}
}
