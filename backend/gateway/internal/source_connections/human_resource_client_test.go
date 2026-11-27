package source_connections

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestHumanResourceSessionUsesVerifiedCallerAndHash(t *testing.T) {
	t.Parallel()
	hash := strings.Repeat("a", 64)
	caller := authcontext.CallerContext{Authenticated: true, PrincipalID: "employee", SessionID: "session", UserKind: "employee"}
	ctx := metadata.AppendToOutgoingContext(t.Context(), humanResourceSessionHashKey, "untrusted")
	ctx = middleware.WithResolvedSession(ctx, &middleware.ResolvedSession{TokenHash: hash, Caller: caller})
	invoked := false
	err := humanResourceSessionInterceptor(time.Second)(ctx, "/test/Worker", nil, nil, nil,
		func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			invoked = true
			md, _ := metadata.FromOutgoingContext(ctx)
			if hashes := md.Get(humanResourceSessionHashKey); len(hashes) != 1 || hashes[0] != hash {
				t.Error("worker call did not receive only the verified session hash")
			}
			if got, ok := authcontext.FromContext(ctx); !ok || got != caller {
				t.Error("worker call did not receive the verified caller")
			}
			if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > time.Second {
				t.Error("worker call did not receive a bounded deadline")
			}
			return nil
		})
	if err != nil || !invoked {
		t.Fatalf("call failed: invoked=%t, error=%v", invoked, err)
	}
}

func TestHumanResourceSessionRejectsUnverifiedRequests(t *testing.T) {
	t.Parallel()
	ctx := metadata.AppendToOutgoingContext(t.Context(), humanResourceSessionHashKey, strings.Repeat("a", 64))
	err := humanResourceSessionInterceptor(time.Second)(ctx, "/test/Worker", nil, nil, nil,
		func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
			t.Fatal("unverified session reached HRM")
			return nil
		})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("error = %v, want Unauthenticated", err)
	}
}

func TestHumanResourceConnectionFailureDoesNotBlockRegistry(t *testing.T) {
	t.Parallel()
	registry, err := NewRegistry(Config{
		IAMTarget:  "127.0.0.1:9001",
		HRMEnabled: true,
		HRM: HumanResourceConnectionConfig{
			Target: "127.0.0.1:9003", ServerName: "human-resource", CertFile: "missing-client-certificate",
		},
		HMACSecret: []byte(strings.Repeat("a", 32)),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := registry.Close(); err != nil {
			t.Error(err)
		}
	})
	if registry.HumanResource != nil || registry.HRMInitError == nil || registry.IAM.Users == nil || registry.IAMInitError != nil {
		t.Fatal("invalid HRM TLS configuration should degrade only HRM")
	}
}

func TestHumanResourceConnectionMutualTLSAndCallerProof(t *testing.T) {
	t.Parallel()
	certPath, keyPath, pool := writeHumanResourceTLSFixture(t)
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}
	secret := []byte(strings.Repeat("a", 32))
	hash := strings.Repeat("b", 64)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{cert},
			ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: pool,
		})),
		grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			caller, found, err := authcontext.ExtractFromIncomingGRPCContext(ctx, secret, time.Now())
			md, _ := metadata.FromIncomingContext(ctx)
			hashes := md.Get(humanResourceSessionHashKey)
			if err != nil || !found || caller.PrincipalID != "employee" || caller.SessionID != "session" ||
				len(hashes) != 1 || hashes[0] != hash {
				return nil, status.Error(codes.Unauthenticated, "invalid caller proof")
			}
			return handler(ctx, req)
		}),
	)
	healthpb.RegisterHealthServer(server, health.NewServer())
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	session := &middleware.ResolvedSession{
		TokenHash: hash,
		Caller: authcontext.CallerContext{
			Authenticated: true, PrincipalID: "employee", SessionID: "session", UserKind: "employee",
		},
	}
	ctx := middleware.WithResolvedSession(t.Context(), session)
	cfg := HumanResourceConnectionConfig{
		Target: listener.Addr().String(), ServerName: "human-resource", Timeout: time.Second,
		CAFile: certPath, CertFile: certPath, KeyFile: keyPath,
	}
	cfg.HMACSecretFile = filepath.Join(t.TempDir(), "hmac.key")
	if err := os.WriteFile(cfg.HMACSecretFile, []byte(string(secret)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if conn, err := NewHumanResourceConnection(HumanResourceConnectionConfig{HMACSecretFile: cfg.HMACSecretFile + ".missing"}, secret); err == nil {
		_ = conn.Close()
		t.Fatal("configured signing key must not fall back when its file is missing")
	}
	for _, tc := range []struct {
		name       string
		serverName string
		want       codes.Code
	}{
		{"trusted server", "human-resource", codes.OK},
		{"wrong server name", "another-service", codes.Unavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg.ServerName = tc.serverName
			conn, err := NewHumanResourceConnection(cfg, []byte(strings.Repeat("c", 32)))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = conn.Close() })
			result, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{})
			if status.Code(err) != tc.want || (tc.want == codes.OK && result.GetStatus() != healthpb.HealthCheckResponse_SERVING) {
				t.Fatalf("health = %v, error = %v, want %s", result, err, tc.want)
			}
		})
	}
}

func writeHumanResourceTLSFixture(t *testing.T) (string, string, *x509.CertPool) {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), NotBefore: now.Add(-time.Minute), NotAfter: now.Add(time.Hour),
		DNSNames: []string{"human-resource"}, IsCA: true, BasicConstraintsValid: true,
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certPath, keyPath := filepath.Join(dir, "certificate.pem"), filepath.Join(dir, "key.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0600); err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certPEM) {
		t.Fatal("invalid test certificate")
	}
	return certPath, keyPath, pool
}
