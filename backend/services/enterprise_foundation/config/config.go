// Package config memuat, memvalidasi, dan menyediakan konfigurasi runtime layanan Enterprise Foundation.
package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAppEnv                      = "development"
	defaultServiceName                 = "integrione-enterprise-foundation"
	defaultGRPCHost                    = "0.0.0.0"
	defaultGRPCPort                    = 9002
	defaultShutdownTimeoutSeconds      = 10
	minProductionHMACSecretBytes       = 32
	defaultOutboxPollIntervalMS        = 5000
	defaultOutboxBatchSize             = 25
	defaultOutboxMaxAttempts           = 20
	defaultOutboxPublishTimeoutSeconds = 5
	defaultOutboxWebhookTimeoutSeconds = 5
	defaultOutboxCleanupIntervalMins   = 60
	defaultOutboxRetentionHours        = 168
	maxOutboxBatchSize                 = 100
)

// OutboxConfig mengatur mode penerbitan, target webhook, interval polling, dan kebijakan retensi outbox relay.
type OutboxConfig struct {
	Mode                string
	WebhookURL          string
	WebhookHMACSecret   []byte
	WebhookRequireHTTPS bool
	WebhookTimeout      time.Duration
	PollInterval        time.Duration
	BatchSize           int
	MaxAttempts         int
	PublishTimeout      time.Duration
	CleanupInterval     time.Duration
	RetentionPeriod     time.Duration
}

// Config menyimpan seluruh parameter konfigurasi untuk proses gRPC server Enterprise Foundation.
type Config struct {
	AppEnv                    string
	ServiceName               string
	GRPCHost                  string
	GRPCPort                  int
	ShutdownTimeout           time.Duration
	DatabaseDSN               string
	TestDatabaseDSN           string
	InternalGatewayHMACSecret []byte
	Outbox                    OutboxConfig
}

// Load membaca konfigurasi dari environment variable beserta nilai bawaan yang aman dan memvalidasinya.
func Load() (Config, error) {
	appEnv := strings.ToLower(envOrDefault("APP_ENV", defaultAppEnv))
	serviceName := envOrDefault("EF_SERVICE_NAME", defaultServiceName)
	grpcHost := envOrDefault("EF_GRPC_HOST", defaultGRPCHost)

	grpcPort, err := envIntOrDefault("EF_GRPC_PORT", defaultGRPCPort)
	if err != nil {
		return Config{}, fmt.Errorf("parse EF_GRPC_PORT: %w", err)
	}

	shutdownSeconds, err := envIntOrDefault("EF_SHUTDOWN_TIMEOUT_SECONDS", defaultShutdownTimeoutSeconds)
	if err != nil {
		return Config{}, fmt.Errorf("parse EF_SHUTDOWN_TIMEOUT_SECONDS: %w", err)
	}

	outboxCfg, err := loadOutboxConfig()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv:                    appEnv,
		ServiceName:               serviceName,
		GRPCHost:                  grpcHost,
		GRPCPort:                  grpcPort,
		ShutdownTimeout:           time.Duration(shutdownSeconds) * time.Second,
		DatabaseDSN:               strings.TrimSpace(os.Getenv("EF_DATABASE_DSN")),
		TestDatabaseDSN:           strings.TrimSpace(os.Getenv("EF_TEST_DATABASE_DSN")),
		InternalGatewayHMACSecret: []byte(strings.TrimSpace(os.Getenv("EF_INTERNAL_GATEWAY_HMAC_SECRET"))),
		Outbox:                    outboxCfg,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadOutboxConfig() (OutboxConfig, error) {
	pollMS, err := envIntOrDefault("EF_OUTBOX_POLL_INTERVAL_MS", defaultOutboxPollIntervalMS)
	if err != nil {
		return OutboxConfig{}, fmt.Errorf("parse EF_OUTBOX_POLL_INTERVAL_MS: %w", err)
	}
	batchSize, err := envIntOrDefault("EF_OUTBOX_BATCH_SIZE", defaultOutboxBatchSize)
	if err != nil {
		return OutboxConfig{}, fmt.Errorf("parse EF_OUTBOX_BATCH_SIZE: %w", err)
	}
	maxAttempts, err := envIntOrDefault("EF_OUTBOX_MAX_ATTEMPTS", defaultOutboxMaxAttempts)
	if err != nil {
		return OutboxConfig{}, fmt.Errorf("parse EF_OUTBOX_MAX_ATTEMPTS: %w", err)
	}
	publishSeconds, err := envIntOrDefault("EF_OUTBOX_PUBLISH_TIMEOUT_SECONDS", defaultOutboxPublishTimeoutSeconds)
	if err != nil {
		return OutboxConfig{}, fmt.Errorf("parse EF_OUTBOX_PUBLISH_TIMEOUT_SECONDS: %w", err)
	}
	webhookTimeoutSeconds, err := envIntOrDefault("EF_OUTBOX_WEBHOOK_TIMEOUT_SECONDS", defaultOutboxWebhookTimeoutSeconds)
	if err != nil {
		return OutboxConfig{}, fmt.Errorf("parse EF_OUTBOX_WEBHOOK_TIMEOUT_SECONDS: %w", err)
	}
	cleanupMins, err := envIntOrDefault("EF_OUTBOX_CLEANUP_INTERVAL_MINUTES", defaultOutboxCleanupIntervalMins)
	if err != nil {
		return OutboxConfig{}, fmt.Errorf("parse EF_OUTBOX_CLEANUP_INTERVAL_MINUTES: %w", err)
	}
	retentionHours, err := envIntOrDefault("EF_OUTBOX_RETENTION_HOURS", defaultOutboxRetentionHours)
	if err != nil {
		return OutboxConfig{}, fmt.Errorf("parse EF_OUTBOX_RETENTION_HOURS: %w", err)
	}
	requireHTTPS, err := envBoolOrDefault("EF_OUTBOX_WEBHOOK_REQUIRE_HTTPS", false)
	if err != nil {
		return OutboxConfig{}, fmt.Errorf("parse EF_OUTBOX_WEBHOOK_REQUIRE_HTTPS: %w", err)
	}

	return OutboxConfig{
		Mode:                strings.ToLower(strings.TrimSpace(os.Getenv("EF_OUTBOX_MODE"))),
		WebhookURL:          strings.TrimSpace(os.Getenv("EF_OUTBOX_WEBHOOK_URL")),
		WebhookHMACSecret:   []byte(strings.TrimSpace(os.Getenv("EF_OUTBOX_WEBHOOK_HMAC_SECRET"))),
		WebhookRequireHTTPS: requireHTTPS,
		WebhookTimeout:      time.Duration(webhookTimeoutSeconds) * time.Second,
		PollInterval:        time.Duration(pollMS) * time.Millisecond,
		BatchSize:           batchSize,
		MaxAttempts:         maxAttempts,
		PublishTimeout:      time.Duration(publishSeconds) * time.Second,
		CleanupInterval:     time.Duration(cleanupMins) * time.Minute,
		RetentionPeriod:     time.Duration(retentionHours) * time.Hour,
	}, nil
}

// Validate memastikan seluruh field konfigurasi runtime memenuhi batas aman dan konsisten.
func (c Config) Validate() error {
	if strings.TrimSpace(c.ServiceName) == "" {
		return fmt.Errorf("EF_SERVICE_NAME must not be empty")
	}
	if strings.TrimSpace(c.GRPCHost) == "" {
		return fmt.Errorf("EF_GRPC_HOST must not be empty")
	}
	if c.GRPCPort < 1 || c.GRPCPort > 65535 {
		return fmt.Errorf("EF_GRPC_PORT must be between 1 and 65535")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("EF_SHUTDOWN_TIMEOUT_SECONDS must be greater than zero")
	}

	// Pada environment production, bukti HMAC internal Gateway wajib tersedia agar trust boundary gagal tertutup.
	if c.IsProduction() && len(c.InternalGatewayHMACSecret) < minProductionHMACSecretBytes {
		return fmt.Errorf("EF_INTERNAL_GATEWAY_HMAC_SECRET must be at least %d bytes in production", minProductionHMACSecretBytes)
	}
	if !c.IsProduction() && len(c.InternalGatewayHMACSecret) > 0 && len(c.InternalGatewayHMACSecret) < minProductionHMACSecretBytes {
		return fmt.Errorf("EF_INTERNAL_GATEWAY_HMAC_SECRET must be empty or at least %d bytes", minProductionHMACSecretBytes)
	}

	if err := c.ValidateOutbox(); err != nil {
		return err
	}

	return nil
}

// ResolvedOutboxMode mengembalikan mode outbox efektif setelah resolusi default otomatis dari konfigurasi webhook.
func (c Config) ResolvedOutboxMode() string {
	mode := strings.ToLower(strings.TrimSpace(c.Outbox.Mode))
	if mode != "" {
		return mode
	}
	if strings.TrimSpace(c.Outbox.WebhookURL) != "" {
		return "webhook"
	}
	return "standby"
}

// ValidateOutbox memvalidasi khusus kebijakan mode, endpoint webhook, HMAC, dan parameter relay outbox.
func (c Config) ValidateOutbox() error {
	mode := c.ResolvedOutboxMode()
	switch mode {
	case "standby", "webhook", "log":
	default:
		return fmt.Errorf("EF_OUTBOX_MODE must be one of standby, webhook, or log")
	}

	// Di environment production, mode log dilarang karena hanya mensimulasikan pengiriman tanpa konsumen nyata.
	if c.IsProduction() && mode == "log" {
		return fmt.Errorf("EF_OUTBOX_MODE=log is not allowed in production")
	}

	if mode == "webhook" {
		rawURL := strings.TrimSpace(c.Outbox.WebhookURL)
		if rawURL == "" || strings.ContainsAny(rawURL, "\r\n\t ") {
			return fmt.Errorf("EF_OUTBOX_WEBHOOK_URL must be set to a valid URL when outbox mode is webhook")
		}
		parsed, err := url.Parse(rawURL)
		if err != nil || parsed == nil || parsed.Host == "" || parsed.User != nil {
			return fmt.Errorf("EF_OUTBOX_WEBHOOK_URL is invalid")
		}
		scheme := strings.ToLower(parsed.Scheme)
		if c.Outbox.WebhookRequireHTTPS && scheme != "https" {
			return fmt.Errorf("EF_OUTBOX_WEBHOOK_URL must use https scheme when EF_OUTBOX_WEBHOOK_REQUIRE_HTTPS is enabled")
		}
		if scheme != "http" && scheme != "https" {
			return fmt.Errorf("EF_OUTBOX_WEBHOOK_URL must use http or https scheme")
		}
	}

	if len(c.Outbox.WebhookHMACSecret) > 0 && len(c.Outbox.WebhookHMACSecret) < minProductionHMACSecretBytes {
		return fmt.Errorf("EF_OUTBOX_WEBHOOK_HMAC_SECRET must be empty or at least %d bytes", minProductionHMACSecretBytes)
	}
	if c.Outbox.BatchSize < 0 || c.Outbox.BatchSize > maxOutboxBatchSize {
		return fmt.Errorf("EF_OUTBOX_BATCH_SIZE must be between 0 and %d", maxOutboxBatchSize)
	}
	if c.Outbox.MaxAttempts < 0 {
		return fmt.Errorf("EF_OUTBOX_MAX_ATTEMPTS must not be negative")
	}
	if c.Outbox.PollInterval < 0 {
		return fmt.Errorf("EF_OUTBOX_POLL_INTERVAL_MS must not be negative")
	}
	if c.Outbox.PublishTimeout < 0 || c.Outbox.WebhookTimeout < 0 {
		return fmt.Errorf("outbox timeout settings must not be negative")
	}
	if c.Outbox.CleanupInterval < 0 || c.Outbox.RetentionPeriod < 0 {
		return fmt.Errorf("outbox cleanup and retention settings must not be negative")
	}
	return nil
}

// GRPCAddress mengembalikan alamat jaringan host:port siap pakai untuk listener TCP gRPC.
func (c Config) GRPCAddress() string {
	return net.JoinHostPort(c.GRPCHost, strconv.Itoa(c.GRPCPort))
}

// IsDevelopment menandakan layanan berjalan pada mode pengembangan lokal.
func (c Config) IsDevelopment() bool {
	return c.AppEnv == "development" || c.AppEnv == "dev" || c.AppEnv == "local"
}

// IsProduction menandakan layanan berjalan pada mode produksi dengan kebijakan fail-closed ketat.
func (c Config) IsProduction() bool {
	return c.AppEnv == "production" || c.AppEnv == "prod"
}

func envOrDefault(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return strings.TrimSpace(val)
}

func envIntOrDefault(key string, fallback int) (int, error) {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil {
		return 0, fmt.Errorf("invalid integer value")
	}
	return parsed, nil
}

func envBoolOrDefault(key string, fallback bool) (bool, error) {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(val))
	if err != nil {
		return false, fmt.Errorf("invalid boolean value")
	}
	return parsed, nil
}
