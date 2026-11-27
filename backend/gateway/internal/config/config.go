// Package config memuat dan memvalidasi konfigurasi runtime fail-closed untuk Gateway GraphQL IntegriOne.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/shared/authsession"
)

const (
	// SessionCookieName adalah nama cookie opaque session kanonik sesuai keputusan arsitektur D01.
	SessionCookieName = "__Host-integrione_session"
	// DefaultMaxComplexity membatasi kompleksitas GraphQL maksimum sesuai standar keamanan Gateway.
	DefaultMaxComplexity = 250
)

var defaultAllowedRedirects = []string{
	"/",
	"/identity",
	"/human-resource",
	"/procurement",
	"/warehouse",
	"/finance",
	"/accounting",
	"/product",
	"/supplier",
	"/vendor",
	"/commerce",
	"/account",
	"/profile",
}

// Config menyimpan seluruh parameter runtime Gateway yang tervalidasi secara ketat.
type Config struct {
	AppEnv                    string
	ServiceName               string
	HTTPHost                  string
	HTTPPort                  int
	ShutdownTimeout           time.Duration
	MaxComplexity             int
	EnableIntrospection       bool
	TrustedProxies            []string
	AllowedOrigins            []string
	AllowedRedirectPaths      []string
	CookieName                string
	CookieSecure              bool
	PreAuthTTL                time.Duration
	SessionIdleTimeout        time.Duration
	SessionMaxAge             time.Duration
	SnapshotMaxAge            time.Duration
	SensitiveReauthWindow     time.Duration
	RateLimitLoginMax         int64
	RateLimitLoginWindow      time.Duration
	RateLimitMutationMax      int64
	RateLimitMutationWindow   time.Duration
	IAMGRPCTarget             string
	IAMGRPCTimeout            time.Duration
	EFGRPCEnabled             bool
	EFGRPCTarget              string
	EFGRPCTimeout             time.Duration
	HRMGRPCEnabled            bool
	HRMGRPCTarget             string
	HRMGRPCServerName         string
	HRMGRPCCAFile             string
	HRMGRPCCertFile           string
	HRMGRPCKeyFile            string
	HRMGRPCHMACSecretFile     string
	HRMGRPCTimeout            time.Duration
	HRMGRPCConfigError        error
	InternalGatewayHMACSecret string
	AvatarBaseURL             string
	RedisEnabled              bool
	RedisEnv                  string
	Redis                     authsession.RedisConfig
}

// HTTPAddress mengembalikan pasangan host:port untuk HTTP listener Fiber Gateway.
func (c *Config) HTTPAddress() string {
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}

// IsProduction melaporkan apakah Gateway berjalan di lingkungan non-pengembangan yang mewajibkan pengetatan penuh.
func (c *Config) IsProduction() bool {
	if c == nil {
		return true
	}
	env := strings.ToLower(strings.TrimSpace(c.AppEnv))
	return env != "" && env != "development" && env != "test"
}

// Load membaca konfigurasi Gateway dari environment variables dengan default yang aman.
func Load() (*Config, error) {
	appEnv := strings.TrimSpace(getEnv("APP_ENV", "development"))
	if appEnv == "" {
		appEnv = "development"
	}

	httpPort, err := getEnvInt("GATEWAY_HTTP_PORT", 8080)
	if err != nil || httpPort <= 0 || httpPort > 65535 {
		return nil, fmt.Errorf("invalid GATEWAY_HTTP_PORT")
	}

	shutdownSec, err := getEnvInt("GATEWAY_SHUTDOWN_TIMEOUT_SECONDS", 10)
	if err != nil || shutdownSec <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_SHUTDOWN_TIMEOUT_SECONDS")
	}

	maxComplexity, err := getEnvInt("GATEWAY_MAX_COMPLEXITY", DefaultMaxComplexity)
	if err != nil || maxComplexity <= 0 || maxComplexity > DefaultMaxComplexity {
		return nil, fmt.Errorf("invalid GATEWAY_MAX_COMPLEXITY: must be between 1 and %d", DefaultMaxComplexity)
	}

	enableIntrospection, err := parseBoolEnv("GATEWAY_ENABLE_INTROSPECTION", strings.EqualFold(appEnv, "development"))
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_ENABLE_INTROSPECTION: %w", err)
	}

	cookieSecure, err := parseBoolEnv("GATEWAY_COOKIE_SECURE", true)
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_COOKIE_SECURE: %w", err)
	}

	preAuthSec, err := getEnvInt("GATEWAY_PREAUTH_TTL_SECONDS", 900)
	if err != nil || preAuthSec <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_PREAUTH_TTL_SECONDS")
	}

	idleSec, err := getEnvInt("GATEWAY_SESSION_IDLE_TIMEOUT_SECONDS", 1800)
	if err != nil || idleSec <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_SESSION_IDLE_TIMEOUT_SECONDS")
	}

	absSec, err := getEnvInt("GATEWAY_SESSION_MAX_AGE_SECONDS", 28800)
	if err != nil || absSec <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_SESSION_MAX_AGE_SECONDS")
	}

	snapSec, err := getEnvInt("GATEWAY_SNAPSHOT_MAX_AGE_SECONDS", 30)
	if err != nil || snapSec <= 0 || snapSec > 30 {
		return nil, fmt.Errorf("invalid GATEWAY_SNAPSHOT_MAX_AGE_SECONDS: must be between 1 and 30")
	}

	reauthSec, err := getEnvInt("GATEWAY_SENSITIVE_REAUTH_WINDOW_SECONDS", 600)
	if err != nil || reauthSec <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_SENSITIVE_REAUTH_WINDOW_SECONDS")
	}

	rlLoginMax, err := getEnvInt("GATEWAY_RATE_LIMIT_LOGIN_MAX", 20)
	if err != nil || rlLoginMax <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_RATE_LIMIT_LOGIN_MAX")
	}
	rlLoginWin, err := getEnvInt("GATEWAY_RATE_LIMIT_LOGIN_WINDOW_SECONDS", 60)
	if err != nil || rlLoginWin <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_RATE_LIMIT_LOGIN_WINDOW_SECONDS")
	}

	rlMutMax, err := getEnvInt("GATEWAY_RATE_LIMIT_MUTATION_MAX", 60)
	if err != nil || rlMutMax <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_RATE_LIMIT_MUTATION_MAX")
	}
	rlMutWin, err := getEnvInt("GATEWAY_RATE_LIMIT_MUTATION_WINDOW_SECONDS", 60)
	if err != nil || rlMutWin <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_RATE_LIMIT_MUTATION_WINDOW_SECONDS")
	}

	iamTimeoutSec, err := getEnvInt("IAM_GRPC_TIMEOUT_SECONDS", 5)
	if err != nil || iamTimeoutSec <= 0 {
		return nil, fmt.Errorf("invalid IAM_GRPC_TIMEOUT_SECONDS")
	}

	efEnabled, err := parseBoolEnv("EF_GRPC_ENABLED", true)
	if err != nil {
		return nil, fmt.Errorf("invalid EF_GRPC_ENABLED: %w", err)
	}
	efTimeoutSec, err := getEnvInt("EF_GRPC_TIMEOUT_SECONDS", 5)
	if err != nil || efTimeoutSec <= 0 {
		return nil, fmt.Errorf("invalid EF_GRPC_TIMEOUT_SECONDS")
	}

	hrmEnabled, hrmConfigErr := parseBoolEnv("HRM_GRPC_ENABLED", false)
	hrmTimeoutSec := 5
	if hrmEnabled {
		hrmTimeoutSec, err = getEnvInt("HRM_GRPC_TIMEOUT_SECONDS", 5)
		if err != nil || hrmTimeoutSec <= 0 {
			hrmConfigErr = errors.New("invalid HRM_GRPC_TIMEOUT_SECONDS")
		}
	}

	redisAddr := strings.TrimSpace(getEnv("GATEWAY_REDIS_ADDR", ""))
	redisEnabled, err := parseBoolEnv("GATEWAY_REDIS_ENABLED", redisAddr != "")
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_REDIS_ENABLED: %w", err)
	}
	redisTLSInsecure, err := parseBoolEnv("GATEWAY_REDIS_TLS_INSECURE_SKIP_VERIFY", false)
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_REDIS_TLS_INSECURE_SKIP_VERIFY: %w", err)
	}

	dialSec, err := getEnvInt("GATEWAY_REDIS_DIAL_TIMEOUT_SECONDS", 2)
	if err != nil || dialSec <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_REDIS_DIAL_TIMEOUT_SECONDS")
	}
	readSec, err := getEnvInt("GATEWAY_REDIS_READ_TIMEOUT_SECONDS", 1)
	if err != nil || readSec <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_REDIS_READ_TIMEOUT_SECONDS")
	}
	writeSec, err := getEnvInt("GATEWAY_REDIS_WRITE_TIMEOUT_SECONDS", 1)
	if err != nil || writeSec <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_REDIS_WRITE_TIMEOUT_SECONDS")
	}
	poolSize, err := getEnvInt("GATEWAY_REDIS_POOL_SIZE", 64)
	if err != nil || poolSize <= 0 {
		return nil, fmt.Errorf("invalid GATEWAY_REDIS_POOL_SIZE")
	}
	minIdle, err := getEnvInt("GATEWAY_REDIS_MIN_IDLE_CONNS", 8)
	if err != nil || minIdle < 0 {
		return nil, fmt.Errorf("invalid GATEWAY_REDIS_MIN_IDLE_CONNS")
	}

	allowedOrigins, err := parseOrigins(getEnv("GATEWAY_ALLOWED_ORIGINS", "https://erp.localhost:8443"))
	if err != nil {
		return nil, fmt.Errorf("invalid GATEWAY_ALLOWED_ORIGINS: %w", err)
	}

	allowedRedirects := parseCSV(getEnv("GATEWAY_ALLOWED_REDIRECT_PATHS", ""))
	if len(allowedRedirects) == 0 {
		allowedRedirects = append([]string(nil), defaultAllowedRedirects...)
	}

	hmacSecret := strings.TrimSpace(getEnv("GATEWAY_INTERNAL_HMAC_SECRET", ""))
	if hmacSecret == "" {
		hmacSecret = strings.TrimSpace(getEnv("IAM_INTERNAL_GATEWAY_HMAC_SECRET", ""))
	}

	redisEnv := strings.TrimSpace(getEnv("GATEWAY_REDIS_ENV", appEnv))
	if redisEnv == "" {
		redisEnv = appEnv
	}

	avatarBaseURL := strings.TrimRight(strings.TrimSpace(getEnv("GATEWAY_AVATAR_BASE_URL", "https://erp.localhost:8443/media/avatars")), "/")

	cfg := &Config{
		AppEnv:                    appEnv,
		ServiceName:               getEnv("GATEWAY_SERVICE_NAME", "integrione-gateway"),
		HTTPHost:                  getEnv("GATEWAY_HTTP_HOST", "0.0.0.0"),
		HTTPPort:                  httpPort,
		ShutdownTimeout:           time.Duration(shutdownSec) * time.Second,
		MaxComplexity:             maxComplexity,
		EnableIntrospection:       enableIntrospection,
		TrustedProxies:            parseCSV(getEnv("GATEWAY_TRUSTED_PROXIES", "")),
		AllowedOrigins:            allowedOrigins,
		AllowedRedirectPaths:      allowedRedirects,
		CookieName:                getEnv("GATEWAY_SESSION_COOKIE_NAME", SessionCookieName),
		CookieSecure:              cookieSecure,
		PreAuthTTL:                time.Duration(preAuthSec) * time.Second,
		SessionIdleTimeout:        time.Duration(idleSec) * time.Second,
		SessionMaxAge:             time.Duration(absSec) * time.Second,
		SnapshotMaxAge:            time.Duration(snapSec) * time.Second,
		SensitiveReauthWindow:     time.Duration(reauthSec) * time.Second,
		RateLimitLoginMax:         int64(rlLoginMax),
		RateLimitLoginWindow:      time.Duration(rlLoginWin) * time.Second,
		RateLimitMutationMax:      int64(rlMutMax),
		RateLimitMutationWindow:   time.Duration(rlMutWin) * time.Second,
		IAMGRPCTarget:             strings.TrimSpace(getEnv("IAM_GRPC_TARGET", "127.0.0.1:9001")),
		IAMGRPCTimeout:            time.Duration(iamTimeoutSec) * time.Second,
		EFGRPCEnabled:             efEnabled,
		EFGRPCTarget:              strings.TrimSpace(getEnv("EF_GRPC_TARGET", "127.0.0.1:9002")),
		EFGRPCTimeout:             time.Duration(efTimeoutSec) * time.Second,
		HRMGRPCEnabled:            hrmEnabled,
		HRMGRPCTarget:             strings.TrimSpace(getEnv("HRM_GRPC_TARGET", "127.0.0.1:9003")),
		HRMGRPCServerName:         strings.TrimSpace(getEnv("HRM_GRPC_SERVER_NAME", "")),
		HRMGRPCCAFile:             strings.TrimSpace(getEnv("HRM_GRPC_CA_FILE", "")),
		HRMGRPCCertFile:           strings.TrimSpace(getEnv("HRM_GRPC_CLIENT_CERT_FILE", "")),
		HRMGRPCHMACSecretFile:     strings.TrimSpace(getEnv("HRM_GRPC_HMAC_SECRET_FILE", "")),
		HRMGRPCKeyFile:            strings.TrimSpace(getEnv("HRM_GRPC_CLIENT_KEY_FILE", "")),
		HRMGRPCTimeout:            time.Duration(hrmTimeoutSec) * time.Second,
		HRMGRPCConfigError:        hrmConfigErr,
		InternalGatewayHMACSecret: hmacSecret,
		AvatarBaseURL:             avatarBaseURL,
		RedisEnabled:              redisEnabled,
		RedisEnv:                  redisEnv,
		Redis: authsession.RedisConfig{
			Addr:                  redisAddr,
			Username:              strings.TrimSpace(getEnv("GATEWAY_REDIS_USERNAME", "gw_runtime")),
			Password:              getEnv("GATEWAY_REDIS_PASSWORD", ""),
			CACertPath:            strings.TrimSpace(getEnv("GATEWAY_REDIS_CA_CERT_PATH", "")),
			ServerName:            strings.TrimSpace(getEnv("GATEWAY_REDIS_SERVER_NAME", "redis")),
			TLSInsecureSkipVerify: redisTLSInsecure,
			DialTimeout:           time.Duration(dialSec) * time.Second,
			ReadTimeout:           time.Duration(readSec) * time.Second,
			WriteTimeout:          time.Duration(writeSec) * time.Second,
			PoolSize:              poolSize,
			MinIdleConns:          minIdle,
		},
	}

	if err := cfg.ValidateRuntimeStrict(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ValidateRuntimeStrict memvalidasi seluruh prasyarat keamanan fail-closed sebelum server menerima trafik.
func (c *Config) ValidateRuntimeStrict() error {
	if c == nil {
		return errors.New("gateway config is required")
	}
	if len(c.AllowedOrigins) == 0 {
		return errors.New("at least one GATEWAY_ALLOWED_ORIGINS entry is required")
	}
	if strings.TrimSpace(c.IAMGRPCTarget) == "" {
		return errors.New("IAM_GRPC_TARGET is required")
	}
	if c.CookieName != SessionCookieName {
		return fmt.Errorf("cookie name must be %q", SessionCookieName)
	}
	if !c.CookieSecure {
		return errors.New("GATEWAY_COOKIE_SECURE must remain true for __Host- session cookie prefix")
	}
	if len(strings.TrimSpace(c.InternalGatewayHMACSecret)) < 32 {
		return errors.New("GATEWAY_INTERNAL_HMAC_SECRET must be at least 32 characters")
	}
	if !c.RedisEnabled || strings.TrimSpace(c.Redis.Addr) == "" {
		return errors.New("GATEWAY_REDIS_ENABLED and GATEWAY_REDIS_ADDR are required for fail-closed gateway")
	}
	if strings.TrimSpace(c.Redis.Password) == "" {
		return errors.New("GATEWAY_REDIS_PASSWORD is required")
	}
	if strings.TrimSpace(c.Redis.CACertPath) == "" || c.Redis.TLSInsecureSkipVerify {
		return errors.New("verified TLS CA certificate is required for gateway redis connection")
	}
	if c.IsProduction() && c.EnableIntrospection {
		return errors.New("GraphQL introspection must be disabled outside development")
	}
	return nil
}

func parseOrigins(raw string) ([]string, error) {
	items := parseCSV(raw)
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == "*" {
			return nil, errors.New("wildcard '*' is prohibited in GATEWAY_ALLOWED_ORIGINS")
		}
		u, err := url.Parse(item)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.Path != "" && u.Path != "/" || u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("invalid origin URL %q", item)
		}
		norm := strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
		if _, dup := seen[norm]; !dup {
			seen[norm] = struct{}{}
			out = append(out, norm)
		}
	}
	return out, nil
}

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	return strconv.Atoi(raw)
}

func parseBoolEnv(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	return strconv.ParseBool(raw)
}
