// Package config memuat konfigurasi HRM tanpa mencetak credential atau DSN.
package config

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/shared/authsession"
)

var databaseNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
var namespacePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

// DatabaseConfig mengatur koneksi dan pool untuk database milik HRM saja.
type DatabaseConfig struct {
	Host, User, Password, Name, SSLMode, SSLRootCert string
	Port, MaxOpenConns, MaxIdleConns                 int
	ConnMaxLifetime                                  time.Duration
}

// TLSConfig menetapkan material mTLS untuk server dan probe internal.
type TLSConfig struct {
	CertFile, KeyFile, CAFile, ServerName string
}

// OutboxConfig mengatur pengiriman event; standby mempertahankan antrean tanpa mengklaim terkirim.
type OutboxConfig struct {
	Mode              string
	WebhookURL        string
	WebhookHMACSecret []byte
	PollInterval      time.Duration
	PublishTimeout    time.Duration
	MaxAttempts       int
}

// EnterpriseFoundationConfig mengatur koneksi gRPC internal ke pemilik data Enterprise Foundation.
type EnterpriseFoundationConfig struct {
	Enabled    bool
	Target     string
	Timeout    time.Duration
	HMACSecret []byte
}

// Config menyediakan dependency runtime yang sudah divalidasi sebelum listener dibuka.
type Config struct {
	AppEnv, ServiceName, GRPCHost string
	GRPCPort                      int
	Database                      DatabaseConfig
	TLS                           TLSConfig
	GatewayHMACSecret             []byte
	StartupTimeout                time.Duration
	ShutdownTimeout               time.Duration
	SecurityRedis                 authsession.RedisConfig
	SecurityNamespace             string
	Outbox                        OutboxConfig
	EnterpriseFoundation          EnterpriseFoundationConfig
	IAMMandateTarget              string
	IAMMandateTLS                 TLSConfig
	TerminationAccessAddress      string
	TerminationAccessTLS          TLSConfig
	PayrollEventsAddress          string
	PayrollEventsTLS              TLSConfig
}

// Load membaca environment; nilai secret dapat dipasang sebagai berkas Docker secret.
func Load() (Config, error) {
	cfg := Config{
		AppEnv: env("APP_ENV", "development"), ServiceName: env("HRM_SERVICE_NAME", "integrione-human-resource"),
		GRPCHost:          env("HRM_GRPC_HOST", "127.0.0.1"),
		SecurityNamespace: env("HRM_SECURITY_NAMESPACE", "development"),
		SecurityRedis: authsession.RedisConfig{
			Addr: env("HRM_REDIS_ADDR", "127.0.0.1:6379"), Username: env("HRM_REDIS_USERNAME", "hrm_reader"),
			CACertPath: env("HRM_REDIS_CA_CERT_PATH", ""), ServerName: env("HRM_REDIS_SERVER_NAME", "redis"),
		},
		Database: DatabaseConfig{
			Host: env("HRM_DB_HOST", "127.0.0.1"), User: env("HRM_DB_USER", "hrm"),
			Name: env("HRM_DB_NAME", "human_resource"), SSLMode: env("HRM_DB_SSLMODE", "verify-full"),
			SSLRootCert: env("HRM_DB_SSLROOTCERT", ""),
		},
		TLS: TLSConfig{
			CertFile: env("HRM_GRPC_TLS_CERT_FILE", ""), KeyFile: env("HRM_GRPC_TLS_KEY_FILE", ""),
			CAFile: env("HRM_GRPC_TLS_CA_FILE", ""), ServerName: env("HRM_GRPC_TLS_SERVER_NAME", "human-resource"),
		},
		TerminationAccessAddress: env("HRM_TERMINATION_ACCESS_ADDR", ""),
		TerminationAccessTLS: TLSConfig{
			CertFile: env("HRM_TERMINATION_ACCESS_CERT_FILE", ""), KeyFile: env("HRM_TERMINATION_ACCESS_KEY_FILE", ""),
			CAFile: env("HRM_TERMINATION_ACCESS_CLIENT_CA_FILE", ""),
		},
		PayrollEventsAddress: env("HRM_PAYROLL_EVENTS_ADDR", ""),
		PayrollEventsTLS: TLSConfig{
			CertFile: env("HRM_PAYROLL_EVENTS_CERT_FILE", ""), KeyFile: env("HRM_PAYROLL_EVENTS_KEY_FILE", ""),
			CAFile: env("HRM_PAYROLL_EVENTS_CLIENT_CA_FILE", ""),
		},
		IAMMandateTarget: env("HRM_IAM_MANDATE_TARGET", ""),
		IAMMandateTLS: TLSConfig{
			CertFile: env("HRM_IAM_MANDATE_CERT_FILE", ""), KeyFile: env("HRM_IAM_MANDATE_KEY_FILE", ""),
			CAFile: env("HRM_IAM_MANDATE_CA_FILE", ""), ServerName: env("HRM_IAM_MANDATE_SERVER_NAME", ""),
		},
		Outbox: OutboxConfig{
			Mode:       env("HRM_OUTBOX_MODE", "standby"),
			WebhookURL: env("HRM_OUTBOX_WEBHOOK_URL", ""),
		},
		EnterpriseFoundation: EnterpriseFoundationConfig{
			Enabled: strings.EqualFold(env("HRM_EF_ENABLED", "false"), "true"),
			Target:  env("HRM_EF_GRPC_TARGET", ""),
		},
	}
	integers := []struct {
		key      string
		fallback int
		target   *int
	}{
		{"HRM_GRPC_PORT", 9003, &cfg.GRPCPort}, {"HRM_DB_PORT", 6543, &cfg.Database.Port},
		{"HRM_DB_MAX_OPEN_CONNS", 20, &cfg.Database.MaxOpenConns},
		{"HRM_DB_MAX_IDLE_CONNS", 5, &cfg.Database.MaxIdleConns},
		{"HRM_REDIS_POOL_SIZE", 16, &cfg.SecurityRedis.PoolSize},
		{"HRM_REDIS_MIN_IDLE_CONNS", 0, &cfg.SecurityRedis.MinIdleConns},
		{"HRM_OUTBOX_MAX_ATTEMPTS", 20, &cfg.Outbox.MaxAttempts},
	}
	for _, setting := range integers {
		value, err := integer(setting.key, setting.fallback)
		if err != nil {
			return Config{}, err
		}
		*setting.target = value
	}
	durations := []struct {
		key      string
		fallback time.Duration
		target   *time.Duration
	}{
		{"HRM_STARTUP_TIMEOUT", 60 * time.Second, &cfg.StartupTimeout},
		{"HRM_SHUTDOWN_TIMEOUT", 10 * time.Second, &cfg.ShutdownTimeout},
		{"HRM_DB_CONN_MAX_LIFETIME", 30 * time.Minute, &cfg.Database.ConnMaxLifetime},
		{"HRM_REDIS_DIAL_TIMEOUT", 2 * time.Second, &cfg.SecurityRedis.DialTimeout},
		{"HRM_REDIS_READ_TIMEOUT", time.Second, &cfg.SecurityRedis.ReadTimeout},
		{"HRM_REDIS_WRITE_TIMEOUT", time.Second, &cfg.SecurityRedis.WriteTimeout},
		{"HRM_OUTBOX_POLL_INTERVAL", 5 * time.Second, &cfg.Outbox.PollInterval},
		{"HRM_OUTBOX_PUBLISH_TIMEOUT", 5 * time.Second, &cfg.Outbox.PublishTimeout},
		{"HRM_EF_GRPC_TIMEOUT", 5 * time.Second, &cfg.EnterpriseFoundation.Timeout},
	}
	for _, setting := range durations {
		value, err := time.ParseDuration(env(setting.key, setting.fallback.String()))
		if err != nil {
			return Config{}, fmt.Errorf("%s must be a valid duration", setting.key)
		}
		*setting.target = value
	}
	password, err := secret("HRM_DB_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	cfg.Database.Password = string(password)
	redisPassword, err := secret("HRM_REDIS_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	cfg.SecurityRedis.Password = string(redisPassword)
	cfg.GatewayHMACSecret, err = secret("HRM_GATEWAY_HMAC_SECRET")
	if err != nil {
		return Config{}, err
	}
	if cfg.Outbox.Mode == "webhook" {
		cfg.Outbox.WebhookHMACSecret, err = secret("HRM_OUTBOX_WEBHOOK_HMAC_SECRET")
		if err != nil {
			return Config{}, err
		}
	}
	if cfg.EnterpriseFoundation.Enabled {
		cfg.EnterpriseFoundation.HMACSecret, err = secret("HRM_EF_HMAC_SECRET")
		if err != nil {
			return Config{}, err
		}
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate menolak konfigurasi yang tidak lengkap atau mengendurkan mTLS.
func (c Config) Validate() error {
	if c.AppEnv != "development" && c.AppEnv != "test" && c.AppEnv != "production" {
		return errors.New("APP_ENV must be development, test, or production")
	}
	if strings.TrimSpace(c.ServiceName) == "" || strings.ContainsAny(c.ServiceName, "\r\n\t ") {
		return errors.New("HRM_SERVICE_NAME must be a nonempty token")
	}
	if strings.TrimSpace(c.GRPCHost) == "" || strings.ContainsAny(c.GRPCHost, "\r\n\t /") {
		return errors.New("HRM_GRPC_HOST must be a valid host")
	}
	if c.GRPCPort < 1 || c.GRPCPort > 65535 {
		return errors.New("HRM_GRPC_PORT must be between 1 and 65535")
	}
	if c.StartupTimeout <= 0 || c.StartupTimeout > 10*time.Minute || c.ShutdownTimeout <= 0 || c.ShutdownTimeout > 2*time.Minute {
		return errors.New("HRM startup and shutdown timeouts are outside allowed bounds")
	}
	if c.TLS.CertFile == "" || c.TLS.KeyFile == "" || c.TLS.CAFile == "" || strings.TrimSpace(c.TLS.ServerName) == "" {
		return errors.New("HRM gRPC mTLS certificate, key, CA, and server name are required")
	}
	if c.TerminationAccessAddress != "" || c.TerminationAccessTLS != (TLSConfig{}) {
		_, port, err := net.SplitHostPort(c.TerminationAccessAddress)
		portNumber, portErr := strconv.Atoi(port)
		tls := c.TerminationAccessTLS
		if err != nil || portErr != nil || portNumber < 1 || portNumber > 65535 ||
			portNumber == c.GRPCPort || tls.CertFile == "" || tls.KeyFile == "" || tls.CAFile == "" {
			return errors.New("HRM termination access requires a separate port, certificate, key, and client CA")
		}
		// Relay memegang transaksi outbox saat IAM memanggil balik verifier HRM.
		if c.IAMMandateTarget != "" && c.Database.MaxOpenConns < 2 {
			return errors.New("HRM termination delivery requires HRM_DB_MAX_OPEN_CONNS to be at least 2")
		}
	}

	if c.PayrollEventsAddress != "" || c.PayrollEventsTLS != (TLSConfig{}) {
		_, port, err := net.SplitHostPort(c.PayrollEventsAddress)
		payrollPort, portErr := strconv.Atoi(port)
		_, terminationPort, _ := net.SplitHostPort(c.TerminationAccessAddress)
		terminationPortNumber, _ := strconv.Atoi(terminationPort)
		tls := c.PayrollEventsTLS
		if err != nil || portErr != nil || payrollPort < 1 || payrollPort > 65535 ||
			payrollPort == c.GRPCPort || payrollPort == terminationPortNumber ||
			tls.CertFile == "" || tls.KeyFile == "" || tls.CAFile == "" {
			return errors.New("HRM payroll events requires a separate port, certificate, key, and client CA")
		}
	}

	if len(c.GatewayHMACSecret) < 32 {
		return errors.New("HRM_GATEWAY_HMAC_SECRET must contain at least 32 bytes")
	}
	rc := c.SecurityRedis
	redisHost, redisPort, redisAddressErr := net.SplitHostPort(rc.Addr)
	redisPortNumber, redisPortErr := strconv.Atoi(redisPort)
	if redisAddressErr != nil || redisHost == "" || redisPortErr != nil || redisPortNumber < 1 || redisPortNumber > 65535 {
		return errors.New("HRM_REDIS_ADDR must contain a host and port")
	}
	if rc.Username == "" || rc.Password == "" || rc.CACertPath == "" || rc.ServerName == "" || rc.TLSInsecureSkipVerify {
		return errors.New("HRM Redis requires ACL credentials and verified TLS")
	}
	if rc.PoolSize < 1 || rc.PoolSize > 100 || rc.MinIdleConns < 0 || rc.MinIdleConns > rc.PoolSize ||
		rc.DialTimeout <= 0 || rc.DialTimeout > 30*time.Second || rc.ReadTimeout <= 0 || rc.ReadTimeout > 30*time.Second ||
		rc.WriteTimeout <= 0 || rc.WriteTimeout > 30*time.Second {
		return errors.New("HRM Redis pool or timeout configuration is invalid")
	}
	if !namespacePattern.MatchString(c.SecurityNamespace) {
		return errors.New("HRM_SECURITY_NAMESPACE must be a lowercase namespace token")
	}
	db := c.Database
	if db.Host == "" || strings.ContainsAny(db.Host, "\r\n\t /") || db.User == "" || db.Password == "" {
		return errors.New("HRM database host, user, and password are required")
	}
	if !databaseNamePattern.MatchString(db.Name) ||
		(db.Name != "human_resource" && !strings.HasPrefix(db.Name, "hrm_") && !strings.HasPrefix(db.Name, "human_resource_")) {
		return errors.New("HRM_DB_NAME must identify an HRM database")
	}
	if db.Port < 1 || db.Port > 65535 || db.MaxOpenConns < 1 || db.MaxOpenConns > 100 ||
		db.MaxIdleConns < 0 || db.MaxIdleConns > db.MaxOpenConns || db.ConnMaxLifetime <= 0 {
		return errors.New("HRM database pool or port configuration is invalid")
	}
	if db.SSLMode != "verify-full" && db.SSLMode != "verify-ca" && db.SSLMode != "require" && db.SSLMode != "disable" {
		return errors.New("HRM_DB_SSLMODE must be verify-full, verify-ca, require, or disable")
	}
	if c.AppEnv == "production" && db.SSLMode != "verify-full" {
		return errors.New("production HRM database requires verify-full TLS")
	}
	mode := strings.TrimSpace(c.Outbox.Mode)
	if mode != "" && mode != "standby" && mode != "log" && mode != "webhook" {
		return errors.New("HRM_OUTBOX_MODE must be standby, log, or webhook")
	}
	if c.AppEnv == "production" && mode == "log" {
		return errors.New("production HRM outbox cannot use log mode")
	}
	if mode == "webhook" {
		endpoint, err := url.Parse(c.Outbox.WebhookURL)
		if err != nil || endpoint.Host == "" || endpoint.User != nil || endpoint.Fragment != "" ||
			(endpoint.Scheme != "https" && (c.AppEnv == "production" || endpoint.Scheme != "http")) ||
			len(c.Outbox.WebhookHMACSecret) < 32 {
			return errors.New("HRM outbox webhook requires a valid URL and HMAC secret")
		}
	}
	if c.Outbox.PollInterval <= 0 || c.Outbox.PollInterval > time.Hour ||
		c.Outbox.PublishTimeout <= 0 || c.Outbox.PublishTimeout > time.Minute ||
		c.Outbox.MaxAttempts < 1 || c.Outbox.MaxAttempts > 1000 {
		return errors.New("HRM outbox retry settings are invalid")
	}
	if (db.SSLMode == "verify-full" || db.SSLMode == "verify-ca") && db.SSLRootCert == "" {
		return errors.New("HRM_DB_SSLROOTCERT is required for database certificate verification")
	}
	if c.EnterpriseFoundation.Enabled {
		efHost, efPort, efAddrErr := net.SplitHostPort(strings.TrimSpace(c.EnterpriseFoundation.Target))
		efPortNumber, efPortErr := strconv.Atoi(efPort)
		if efAddrErr != nil || efHost == "" || strings.ContainsAny(efHost, "\r\n\t /") ||
			efPortErr != nil || efPortNumber < 1 || efPortNumber > 65535 {
			return errors.New("HRM_EF_GRPC_TARGET must contain a host and port")
		}
		if c.EnterpriseFoundation.Timeout <= 0 || c.EnterpriseFoundation.Timeout > time.Minute {
			return errors.New("HRM_EF_GRPC_TIMEOUT is outside allowed bounds")
		}
		if len(c.EnterpriseFoundation.HMACSecret) < 32 {
			return errors.New("HRM_EF_HMAC_SECRET must contain at least 32 bytes")
		}
	}
	return nil
}

// GRPCAddress membentuk alamat listener termasuk dukungan host IPv6.
func (c Config) GRPCAddress() string {
	return net.JoinHostPort(c.GRPCHost, strconv.Itoa(c.GRPCPort))
}

// DSN membentuk URL ter-escape; hasilnya tidak boleh ditulis ke log atau respons API.
func (c DatabaseConfig) DSN() string {
	value := &url.URL{
		Scheme: "postgres", User: url.UserPassword(c.User, c.Password),
		Host: net.JoinHostPort(c.Host, strconv.Itoa(c.Port)), Path: "/" + c.Name,
	}
	query := url.Values{"sslmode": {c.SSLMode}, "connect_timeout": {"10"}, "timezone": {"UTC"}}
	if c.SSLRootCert != "" {
		query.Set("sslrootcert", c.SSLRootCert)
	}
	value.RawQuery = query.Encode()
	return value.String()
}

func env(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return fallback
}

func integer(key string, fallback int) (int, error) {
	value, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return value, nil
}

func secret(key string) ([]byte, error) {
	value, file := os.Getenv(key), env(key+"_FILE", "")
	if value != "" && file != "" {
		return nil, fmt.Errorf("%s and its FILE setting are mutually exclusive", key)
	}
	if file == "" {
		return []byte(value), nil
	}
	handle, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s secret file", key)
	}
	defer handle.Close()
	data, err := io.ReadAll(io.LimitReader(handle, 65537))
	if err != nil || len(data) > 65536 {
		return nil, fmt.Errorf("invalid %s secret file", key)
	}
	return []byte(strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r")), nil
}
