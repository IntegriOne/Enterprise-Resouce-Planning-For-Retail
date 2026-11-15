package authsession

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig menyimpan parameter koneksi Redis TLS dan ACL untuk IAM maupun Gateway.
type RedisConfig struct {
	Addr                  string
	Username              string
	Password              string
	CACertPath            string
	ServerName            string
	TLSInsecureSkipVerify bool
	DialTimeout           time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	PoolSize              int
	MinIdleConns          int
}

// NewTLSClient membangun klien Redis RESP3 yang memaksa TLS 1.2+, verifikasi CA, dan ContextTimeoutEnabled.
func NewTLSClient(cfg RedisConfig) (*redis.Client, error) {
	addr := strings.TrimSpace(cfg.Addr)
	if addr == "" {
		return nil, errors.New("redis address is required")
	}
	username := strings.TrimSpace(cfg.Username)
	if username == "" {
		return nil, errors.New("redis ACL username is required")
	}
	if cfg.Password == "" {
		return nil, errors.New("redis ACL password is required")
	}

	tlsCfg, err := buildTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	dialTimeout := cfg.DialTimeout
	if dialTimeout <= 0 {
		dialTimeout = 2 * time.Second
	}
	readTimeout := cfg.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = 500 * time.Millisecond
	}
	writeTimeout := cfg.WriteTimeout
	if writeTimeout <= 0 {
		writeTimeout = 500 * time.Millisecond
	}
	poolSize := cfg.PoolSize
	if poolSize <= 0 {
		poolSize = 32
	}
	minIdle := cfg.MinIdleConns
	if minIdle < 0 {
		minIdle = 0
	}

	client := redis.NewClient(&redis.Options{
		Addr:                  addr,
		Username:              username,
		Password:              cfg.Password,
		TLSConfig:             tlsCfg,
		Protocol:              3,
		ContextTimeoutEnabled: true,
		DialTimeout:           dialTimeout,
		ReadTimeout:           readTimeout,
		WriteTimeout:          writeTimeout,
		PoolSize:              poolSize,
		MinIdleConns:          minIdle,
		// Gunakan batas retry moderat agar koneksi pool usang (EOF/broken pipe) pasca-restart Redis terkuras otomatis tanpa memblokir lama saat mati.
		MaxRetries:      4,
		MinRetryBackoff: 10 * time.Millisecond,
		MaxRetryBackoff: 100 * time.Millisecond,
	})
	return client, nil
}

// Ping memverifikasi konektivitas, handshake TLS, dan autentikasi ACL ke server Redis.
func Ping(ctx context.Context, client *redis.Client) error {
	if client == nil {
		return ErrRedisUnavailable
	}
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}
	return nil
}

func buildTLSConfig(cfg RedisConfig) (*tls.Config, error) {
	serverName := strings.TrimSpace(cfg.ServerName)
	if serverName == "" {
		host, _, found := strings.Cut(strings.TrimSpace(cfg.Addr), ":")
		if found && host != "" {
			serverName = host
		} else {
			serverName = "redis"
		}
	}

	tlsCfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         serverName,
		InsecureSkipVerify: cfg.TLSInsecureSkipVerify,
	}
	caPath := strings.TrimSpace(cfg.CACertPath)
	if caPath == "" {
		if !cfg.TLSInsecureSkipVerify {
			return nil, errors.New("redis CA cert path is required for TLS verification")
		}
		return tlsCfg, nil
	}

	pemBytes, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("read redis CA cert %q: %w", caPath, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		return nil, fmt.Errorf("parse redis CA cert %q: no valid PEM certificates found", caPath)
	}
	tlsCfg.RootCAs = pool
	return tlsCfg, nil
}
