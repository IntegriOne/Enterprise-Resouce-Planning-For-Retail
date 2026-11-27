package source_connections

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const humanResourceSessionHashKey = "x-ig-session-token-hash"

// HumanResourceConnectionConfig memuat target dan identitas mTLS koneksi privat HRM.
type HumanResourceConnectionConfig struct {
	Target, ServerName        string
	CAFile, CertFile, KeyFile string
	HMACSecretFile            string
	Timeout                   time.Duration
}

// HumanResourceSigningKey membaca kunci yang dipakai koneksi privat Gateway ke HRM.
func HumanResourceSigningKey(path string, hmacSecret []byte) ([]byte, error) {
	if path != "" {
		secret, err := os.ReadFile(path)
		if err != nil {
			return nil, errors.New("read human resource caller signing key failed")
		}
		return bytes.TrimSpace(secret), nil
	}
	return bytes.Clone(hmacSecret), nil
}

// NewHumanResourceConnection menyiapkan koneksi HRM tanpa menunggu upstream hidup.
func NewHumanResourceConnection(cfg HumanResourceConnectionConfig, hmacSecret []byte) (*grpc.ClientConn, error) {
	hmacSecret, err := HumanResourceSigningKey(cfg.HMACSecretFile, hmacSecret)
	if err != nil {
		return nil, err
	}
	target, serverName := strings.TrimSpace(cfg.Target), strings.TrimSpace(cfg.ServerName)
	if target == "" || serverName == "" || len(hmacSecret) < 32 {
		return nil, errors.New("invalid human resource target, TLS server name, or caller signing key")
	}
	certificate, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, errors.New("load human resource client certificate failed")
	}
	caPEM, err := os.ReadFile(cfg.CAFile)
	if err != nil {
		return nil, errors.New("read human resource CA failed")
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		return nil, errors.New("parse human resource CA failed")
	}
	transport := credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS13, ServerName: serverName,
		RootCAs: roots, Certificates: []tls.Certificate{certificate},
	})
	breaker := newCircuitBreaker(defaultCircuitFailureThreshold, defaultCircuitOpenDuration, errors.New("human resource circuit breaker is open"), time.Now)
	return grpc.NewClient(target,
		grpc.WithTransportCredentials(transport),
		grpc.WithChainUnaryInterceptor(humanResourceSessionInterceptor(cfg.Timeout), humanResourceCircuitInterceptor(breaker), internalContextClientInterceptor(bytes.Clone(hmacSecret))),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(false)),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.Config{BaseDelay: 250 * time.Millisecond, Multiplier: 1.6, Jitter: 0.2, MaxDelay: 2 * time.Second},
			MinConnectTimeout: 2 * time.Second,
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time: 60 * time.Second, Timeout: 10 * time.Second, PermitWithoutStream: true,
		}),
	)
}

// humanResourceSessionInterceptor meneruskan hash dari sesi terverifikasi Gateway.
func humanResourceSessionInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		session, ok := middleware.ResolvedSessionFromContext(ctx)
		if !ok || !session.Caller.Authenticated || session.Caller.UserKind != "employee" ||
			session.Caller.PrincipalID == "" || session.Caller.SessionID == "" || len(session.TokenHash) != 64 {
			return status.Error(codes.Unauthenticated, "verified employee session is required")
		}
		if _, err := hex.DecodeString(session.TokenHash); err != nil {
			return status.Error(codes.Unauthenticated, "invalid verified session hash")
		}
		md, _ := metadata.FromOutgoingContext(ctx)
		if md == nil {
			md = metadata.MD{}
		}
		md.Set(humanResourceSessionHashKey, session.TokenHash)
		ctx = metadata.NewOutgoingContext(authcontext.WithCaller(ctx, session.Caller), md)
		ctx, cancel := withBoundedTimeout(ctx, timeout)
		defer cancel()
		return invoke(ctx, method, req, reply, conn, opts...)
	}
}

// humanResourceCircuitInterceptor mengisolasi kegagalan upstream HRM dari service lain.
func humanResourceCircuitInterceptor(breaker *circuitBreaker) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, conn *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		probe, err := breaker.beforeCall()
		if err != nil {
			return status.Error(codes.Unavailable, "human resource service unavailable")
		}
		err = invoke(ctx, method, req, reply, conn, opts...)
		breaker.afterCall(probe, isUpstreamUnavailableFailure(ctx, err), isLocalCallerCancellation(ctx, err))
		return err
	}
}
