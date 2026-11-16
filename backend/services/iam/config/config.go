package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/mailer"
	"github.com/integrione/erp-retail/backend/shared/authsession"
)

// SystemBootstrapConfig menyimpan parameter akun system bootstrap IAM dari environment.
type SystemBootstrapConfig struct {
	UserUID     string
	Username    string
	Email       string
	Password    string
	DisplayName string
	UserKind    string
	RoleCode    string
}

// Config menyimpan seluruh konfigurasi runtime untuk layanan IAM.
type Config struct {
	AppEnv                    string
	ServiceName               string
	GRPCHost                  string
	GRPCPort                  int
	DatabaseDSN               string
	TestDatabaseDSN           string
	ShutdownTimeout           time.Duration
	RedisEnabled              bool
	RedisEnv                  string
	Redis                     authsession.RedisConfig
	InternalGatewayHMACSecret string
	SystemBootstrap           SystemBootstrapConfig
	Email                     mailer.Config
}

// Load membaca konfigurasi dari environment variables dengan nilai default pengembangan.
func Load() (*Config, error) {
	port, err := getEnvAsInt("IAM_GRPC_PORT", 9001)
	if err != nil {
		return nil, fmt.Errorf("invalid IAM_GRPC_PORT: %w", err)
	}

	shutdownSec, err := getEnvAsInt("IAM_SHUTDOWN_TIMEOUT_SECONDS", 10)
	if err != nil {
		return nil, fmt.Errorf("invalid IAM_SHUTDOWN_TIMEOUT_SECONDS: %w", err)
	}

	appEnv := strings.TrimSpace(getEnv("APP_ENV", "development"))
	if appEnv == "" {
		appEnv = "development"
	}

	redisAddr := strings.TrimSpace(getEnv("IAM_REDIS_ADDR", ""))
	redisEnabled, err := parseOptionalBoolEnv("IAM_REDIS_ENABLED", redisAddr != "")
	if err != nil {
		return nil, fmt.Errorf("invalid IAM_REDIS_ENABLED: %w", err)
	}

	redisTLSInsecure, err := parseOptionalBoolEnv("IAM_REDIS_TLS_INSECURE_SKIP_VERIFY", false)
	if err != nil {
		return nil, fmt.Errorf("invalid IAM_REDIS_TLS_INSECURE_SKIP_VERIFY: %w", err)
	}

	dialTimeoutSec, err := getEnvAsInt("IAM_REDIS_DIAL_TIMEOUT_SECONDS", 3)
	if err != nil || dialTimeoutSec <= 0 {
		return nil, fmt.Errorf("invalid IAM_REDIS_DIAL_TIMEOUT_SECONDS")
	}
	readTimeoutSec, err := getEnvAsInt("IAM_REDIS_READ_TIMEOUT_SECONDS", 2)
	if err != nil || readTimeoutSec <= 0 {
		return nil, fmt.Errorf("invalid IAM_REDIS_READ_TIMEOUT_SECONDS")
	}
	writeTimeoutSec, err := getEnvAsInt("IAM_REDIS_WRITE_TIMEOUT_SECONDS", 2)
	if err != nil || writeTimeoutSec <= 0 {
		return nil, fmt.Errorf("invalid IAM_REDIS_WRITE_TIMEOUT_SECONDS")
	}
	poolSize, err := getEnvAsInt("IAM_REDIS_POOL_SIZE", 32)
	if err != nil || poolSize <= 0 {
		return nil, fmt.Errorf("invalid IAM_REDIS_POOL_SIZE")
	}
	minIdle, err := getEnvAsInt("IAM_REDIS_MIN_IDLE_CONNS", 4)
	if err != nil || minIdle < 0 {
		return nil, fmt.Errorf("invalid IAM_REDIS_MIN_IDLE_CONNS")
	}

	redisEnv := strings.TrimSpace(getEnv("IAM_REDIS_ENV", appEnv))
	if redisEnv == "" {
		redisEnv = appEnv
	}

	emailCfg, err := loadEmailConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		AppEnv:          appEnv,
		ServiceName:     getEnv("IAM_SERVICE_NAME", "integrione-iam"),
		GRPCHost:        getEnv("IAM_GRPC_HOST", "0.0.0.0"),
		GRPCPort:        port,
		DatabaseDSN:     getEnv("IAM_DATABASE_DSN", "postgres://postgres:postgres@127.0.0.1:6543/iam?sslmode=disable"),
		TestDatabaseDSN: getEnv("IAM_TEST_DATABASE_DSN", ""),
		ShutdownTimeout: time.Duration(shutdownSec) * time.Second,
		RedisEnabled:    redisEnabled,
		RedisEnv:        redisEnv,
		Redis: authsession.RedisConfig{
			Addr:                  redisAddr,
			Username:              strings.TrimSpace(getEnv("IAM_REDIS_USERNAME", "iam_writer")),
			Password:              getEnv("IAM_REDIS_PASSWORD", ""),
			CACertPath:            strings.TrimSpace(getEnv("IAM_REDIS_CA_CERT_PATH", "")),
			ServerName:            strings.TrimSpace(getEnv("IAM_REDIS_SERVER_NAME", "redis")),
			TLSInsecureSkipVerify: redisTLSInsecure,
			DialTimeout:           time.Duration(dialTimeoutSec) * time.Second,
			ReadTimeout:           time.Duration(readTimeoutSec) * time.Second,
			WriteTimeout:          time.Duration(writeTimeoutSec) * time.Second,
			PoolSize:              poolSize,
			MinIdleConns:          minIdle,
		},
		InternalGatewayHMACSecret: strings.TrimSpace(getEnv("IAM_INTERNAL_GATEWAY_HMAC_SECRET", "")),
		SystemBootstrap: SystemBootstrapConfig{
			UserUID:     strings.TrimSpace(getEnv("IAM_SYSTEM_USER_UID", "")),
			Username:    strings.TrimSpace(getEnv("IAM_SYSTEM_USERNAME", "systemiam")),
			Email:       strings.TrimSpace(getEnv("IAM_SYSTEM_EMAIL", "systemiam@integrione.local")),
			Password:    getEnv("IAM_SYSTEM_PASSWORD", ""),
			DisplayName: strings.TrimSpace(getEnv("IAM_SYSTEM_DISPLAY_NAME", "IntegriOne System Administrator")),
			UserKind:    strings.TrimSpace(getEnv("IAM_SYSTEM_USER_KIND", "system")),
			RoleCode:    strings.TrimSpace(getEnv("IAM_SYSTEM_ROLE_CODE", "superadmin")),
		},
		Email: emailCfg,
	}, nil
}

// IsProduction melaporkan apakah layanan IAM dijalankan pada mode non-development/test.
func (c *Config) IsProduction() bool {
	if c == nil {
		return true
	}
	envNormalized := strings.ToLower(strings.TrimSpace(c.AppEnv))
	return envNormalized != "" && envNormalized != "development" && envNormalized != "test"
}

// GRPCAddress mengembalikan alamat lengkap host:port untuk listener gRPC.
func (c *Config) GRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.GRPCHost, c.GRPCPort)
}

// ValidateRuntimeStrict menegakkan kebijakan fail-closed saat daemon IAM dijalankan di lingkungan non-pengembangan.
func (c *Config) ValidateRuntimeStrict(sec SecurityConfig) error {
	if c == nil {
		return errors.New("iam config is required")
	}
	if err := c.Email.Validate(c.IsProduction()); err != nil {
		return fmt.Errorf("invalid IAM email configuration in %q environment: %w", c.AppEnv, err)
	}
	if !strings.EqualFold(strings.TrimSpace(c.AppEnv), "test") && len(strings.TrimSpace(c.InternalGatewayHMACSecret)) < 32 {
		return fmt.Errorf("IAM_INTERNAL_GATEWAY_HMAC_SECRET (>= 32 bytes) is required in %q environment", c.AppEnv)
	}
	if !c.IsProduction() {
		return nil
	}
	if !c.RedisEnabled || strings.TrimSpace(c.Redis.Addr) == "" {
		return fmt.Errorf("IAM_REDIS_ENABLED and IAM_REDIS_ADDR are required in %q environment", c.AppEnv)
	}
	if strings.TrimSpace(c.Redis.Password) == "" {
		return fmt.Errorf("IAM_REDIS_PASSWORD is required in %q environment", c.AppEnv)
	}
	if strings.TrimSpace(c.Redis.CACertPath) == "" || c.Redis.TLSInsecureSkipVerify {
		return fmt.Errorf("verified TLS CA certificate is required for Redis in %q environment", c.AppEnv)
	}
	if !sec.CSRFKeyringConfigured || len(sec.CSRFKeyring) == 0 {
		return fmt.Errorf("IAM_CSRF_HMAC_KEYRING is required in %q environment", c.AppEnv)
	}
	return nil
}

func loadEmailConfig() (mailer.Config, error) {
	def := mailer.DefaultConfig()
	provider := mailer.Provider(strings.ToLower(strings.TrimSpace(getEnv("IAM_EMAIL_PROVIDER", string(def.Provider)))))
	mode := mailer.DeliveryMode(strings.ToLower(strings.TrimSpace(getEnv("IAM_EMAIL_DELIVERY_MODE", string(def.DeliveryMode)))))

	timeoutSec, err := getEnvAsInt("IAM_EMAIL_TIMEOUT_SECONDS", int(def.Timeout.Seconds()))
	if err != nil || timeoutSec <= 0 {
		return mailer.Config{}, fmt.Errorf("invalid IAM_EMAIL_TIMEOUT_SECONDS")
	}
	queueSize, err := getEnvAsInt("IAM_EMAIL_QUEUE_SIZE", def.QueueSize)
	if err != nil || queueSize <= 0 {
		return mailer.Config{}, fmt.Errorf("invalid IAM_EMAIL_QUEUE_SIZE")
	}
	workers, err := getEnvAsInt("IAM_EMAIL_WORKERS", def.Workers)
	if err != nil || workers <= 0 {
		return mailer.Config{}, fmt.Errorf("invalid IAM_EMAIL_WORKERS")
	}
	maxAttempts, err := getEnvAsInt("IAM_EMAIL_MAX_ATTEMPTS", def.MaxAttempts)
	if err != nil || maxAttempts <= 0 {
		return mailer.Config{}, fmt.Errorf("invalid IAM_EMAIL_MAX_ATTEMPTS")
	}
	smtpPort, err := getEnvAsInt("IAM_SMTP_PORT", def.SMTPPort)
	if err != nil || smtpPort <= 0 || smtpPort > 65535 {
		return mailer.Config{}, fmt.Errorf("invalid IAM_SMTP_PORT")
	}
	allowPlaintext, err := parseOptionalBoolEnv("IAM_SMTP_ALLOW_PLAINTEXT", false)
	if err != nil {
		return mailer.Config{}, fmt.Errorf("invalid IAM_SMTP_ALLOW_PLAINTEXT: %w", err)
	}

	sesRegion := firstNonEmptyEnv("IAM_AWS_REGION", "AWS_REGION", "AWS_DEFAULT_REGION")
	sesKey := firstNonEmptyEnv("IAM_AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID")
	sesSecret := firstNonEmptyEnv("IAM_AWS_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY")
	sesSession := firstNonEmptyEnv("IAM_AWS_SESSION_TOKEN", "AWS_SESSION_TOKEN")
	apiKey := resolveEmailAPIKey(provider)

	return mailer.Config{
		Provider:            provider,
		DeliveryMode:        mode,
		From:                getEnv("IAM_EMAIL_FROM", def.From),
		ReplyTo:             getEnv("IAM_EMAIL_REPLY_TO", ""),
		PublicBaseURL:       getEnv("IAM_PUBLIC_BASE_URL", def.PublicBaseURL),
		Timeout:             time.Duration(timeoutSec) * time.Second,
		QueueSize:           queueSize,
		Workers:             workers,
		MaxAttempts:         maxAttempts,
		InitialRetryBackoff: def.InitialRetryBackoff,
		SESRegion:           sesRegion,
		SESAccessKeyID:      sesKey,
		SESSecretAccessKey:  sesSecret,
		SESSessionToken:     sesSession,
		SESConfigurationSet: strings.TrimSpace(getEnv("IAM_AWS_SES_CONFIGURATION_SET", "")),
		SESEndpointURL:      strings.TrimSpace(getEnv("IAM_AWS_SES_ENDPOINT", "")),
		SMTPHost:            strings.TrimSpace(getEnv("IAM_SMTP_HOST", "")),
		SMTPPort:            smtpPort,
		SMTPUsername:        strings.TrimSpace(getEnv("IAM_SMTP_USERNAME", "")),
		SMTPPassword:        getEnv("IAM_SMTP_PASSWORD", ""),
		SMTPTLSMode:         mailer.SMTPTLSMode(strings.ToLower(strings.TrimSpace(getEnv("IAM_SMTP_TLS_MODE", string(def.SMTPTLSMode))))),
		SMTPServerName:      strings.TrimSpace(getEnv("IAM_SMTP_SERVER_NAME", "")),
		AllowPlaintext:      allowPlaintext,
		APIKey:              apiKey,
		APIBaseURL:          strings.TrimSpace(getEnv("IAM_EMAIL_API_BASE_URL", "")),
	}.Normalize(), nil
}

func firstNonEmptyEnv(keys ...string) string {
	for _, k := range keys {
		if val := strings.TrimSpace(getEnv(k, "")); val != "" {
			return val
		}
	}
	return ""
}

func resolveEmailAPIKey(provider mailer.Provider) string {
	switch provider {
	case mailer.ProviderSendGrid:
		return firstNonEmptyEnv("IAM_EMAIL_API_KEY", "IAM_SENDGRID_API_KEY")
	case mailer.ProviderResend:
		return firstNonEmptyEnv("IAM_EMAIL_API_KEY", "IAM_RESEND_API_KEY")
	case mailer.ProviderPostmark:
		return firstNonEmptyEnv("IAM_EMAIL_API_KEY", "IAM_POSTMARK_SERVER_TOKEN")
	default:
		return strings.TrimSpace(getEnv("IAM_EMAIL_API_KEY", ""))
	}
}

// getEnv mengambil nilai environment variable atau mengembalikan fallback jika kosong.
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

// getEnvAsInt mengambil nilai environment variable sebagai integer atau mengembalikan fallback.
func getEnvAsInt(key string, fallback int) (int, error) {
	valStr := getEnv(key, "")
	if valStr == "" {
		return fallback, nil
	}
	return strconv.Atoi(valStr)
}

// parseOptionalBoolEnv mem-parsing boolean dari environment variable atau mengembalikan defaultValue.
func parseOptionalBoolEnv(key string, defaultValue bool) (bool, error) {
	valStr := strings.TrimSpace(getEnv(key, ""))
	if valStr == "" {
		return defaultValue, nil
	}
	return strconv.ParseBool(valStr)
}
