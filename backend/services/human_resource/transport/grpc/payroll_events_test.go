package grpc

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workload/payroll/v1"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/payrollevents"
	stdgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type testPayrollEventReceiver struct {
	payrollevents.Receiver
	entered chan struct{}
	release <-chan struct{}
}

func (r testPayrollEventReceiver) ConfirmPayrollPosting(ctx context.Context, runID, journalID string) error {
	if runID == "blocked-run" {
		close(r.entered)
		<-r.release
		return ctx.Err()
	}
	if runID != "run-fixture" || journalID != "journal-fixture" {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func TestPayrollEventServerAuthenticatesFinanceCertificate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	now := time.Now()
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1), IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign, NotBefore: now.Add(-time.Minute), NotAfter: now.Add(time.Hour),
	}
	caDER, caKey := issuePayrollTestCertificate(t, ca, nil, nil)
	ca, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(ca)
	write := func(name, kind string, der []byte) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: kind, Bytes: der}), 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	serverCert := &x509.Certificate{
		SerialNumber: big.NewInt(2), DNSNames: []string{"payroll.test"},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		NotBefore:   ca.NotBefore, NotAfter: ca.NotAfter,
	}
	serverDER, serverKey := issuePayrollTestCertificate(t, serverCert, ca, caKey)
	keyDER, err := x509.MarshalPKCS8PrivateKey(serverKey)
	if err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	server, err := NewPayrollEventServer(
		write("server.pem", "CERTIFICATE", serverDER), write("server-key.pem", "PRIVATE KEY", keyDER),
		write("ca.pem", "CERTIFICATE", caDER), testPayrollEventReceiver{entered: entered, release: release},
	)
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	t.Cleanup(server.Stop)
	go func() { _ = server.Serve(listener) }()

	for i, tc := range []struct {
		name, workload string
		want           codes.Code
	}{
		{name: "no client certificate", want: codes.Unavailable},
		{name: "IAM certificate", workload: "iam", want: codes.PermissionDenied},
		{name: "Finance certificate", workload: "finance-accounting", want: codes.OK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: roots, ServerName: "payroll.test"}
			if tc.workload != "" {
				leaf := &x509.Certificate{
					SerialNumber: big.NewInt(int64(i + 3)),
					URIs:         []*url.URL{{Scheme: "spiffe", Host: "integrione", Path: "/" + tc.workload}},
					ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
					NotBefore:    ca.NotBefore, NotAfter: ca.NotAfter,
				}
				der, key := issuePayrollTestCertificate(t, leaf, ca, caKey)
				cfg.Certificates = []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}
			}
			conn, err := stdgrpc.NewClient("passthrough:///payroll.test",
				stdgrpc.WithTransportCredentials(credentials.NewTLS(cfg)),
				stdgrpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
					return listener.DialContext(ctx)
				}),
			)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = conn.Close() })
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			_, err = pb.NewPayrollEventServiceClient(conn).DeliverPayrollEvent(ctx, &pb.DeliverPayrollEventRequest{
				PayrollRunId: "run-fixture", SchemaVersion: 1,
				Evidence: &pb.DeliverPayrollEventRequest_Posting{Posting: &pb.PostingEvidence{JournalId: "journal-fixture"}},
			})
			if status.Code(err) != tc.want {
				t.Fatalf("delivery code=%v, want %v, error=%v", status.Code(err), tc.want, err)
			}
			if tc.want != codes.OK {
				return
			}
			rpcDone := make(chan error, 1)
			go func() {
				_, callErr := pb.NewPayrollEventServiceClient(conn).DeliverPayrollEvent(ctx, &pb.DeliverPayrollEventRequest{
					PayrollRunId: "blocked-run", SchemaVersion: 1,
					Evidence: &pb.DeliverPayrollEventRequest_Posting{Posting: &pb.PostingEvidence{JournalId: "journal-fixture"}},
				})
				rpcDone <- callErr
			}()
			select {
			case <-entered:
			case <-ctx.Done():
				t.Fatal("blocking event did not reach the receiver")
			}
			shutdown, finish := context.WithTimeout(t.Context(), 30*time.Millisecond)
			defer finish()
			drained := make(chan error, 1)
			go func() { drained <- DrainServer(shutdown, server) }()
			select {
			case err := <-drained:
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("shutdown with active event: %v", err)
				}
			case <-ctx.Done():
				t.Fatal("shutdown waited for a receiver that ignores cancellation")
			}
			if err := <-rpcDone; err == nil {
				t.Fatal("forced shutdown acknowledged an incomplete event")
			}

		})
	}
}

func issuePayrollTestCertificate(t *testing.T, template, issuer *x509.Certificate, issuerKey ed25519.PrivateKey) ([]byte, ed25519.PrivateKey) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if issuer == nil {
		issuer, issuerKey = template, privateKey
	}
	der, err := x509.CreateCertificate(rand.Reader, template, issuer, publicKey, issuerKey)
	if err != nil {
		t.Fatal(err)
	}
	return der, privateKey
}
