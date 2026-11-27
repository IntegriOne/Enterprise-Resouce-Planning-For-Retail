package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type mandateFixture struct{}

func (mandateFixture) HasRequisitionApprovalMandate(context.Context, string, string) (bool, error) {
	return true, nil
}

func issueMandateCertificate(t *testing.T, template, issuer *x509.Certificate, signer ed25519.PrivateKey) (*x509.Certificate, ed25519.PrivateKey) {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if issuer == nil {
		issuer, signer = template, key
	}
	der, err := x509.CreateCertificate(rand.Reader, template, issuer, pub, signer)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return cert, key
}

func TestRequisitionMandateListenerEnforcesMTLS(t *testing.T) {
	t.Parallel()
	now := time.Now()
	template := func(serial int64) *x509.Certificate {
		return &x509.Certificate{SerialNumber: big.NewInt(serial), NotBefore: now.Add(-time.Minute), NotAfter: now.Add(time.Hour)}
	}
	caTemplate := template(1)
	caTemplate.IsCA, caTemplate.BasicConstraintsValid, caTemplate.KeyUsage = true, true, x509.KeyUsageCertSign
	ca, caKey := issueMandateCertificate(t, caTemplate, nil, nil)
	serverTemplate := template(2)
	serverTemplate.DNSNames = []string{"iam-mandate.test"}
	serverTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	serverCert, serverKey := issueMandateCertificate(t, serverTemplate, ca, caKey)
	directory := t.TempDir()
	writePEM := func(name, kind string, data []byte) string {
		t.Helper()
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: kind, Bytes: data}), 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(serverKey)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.RequisitionMandateConfig{
		Address: "127.0.0.1:0", CertFile: writePEM("server.crt", "CERTIFICATE", serverCert.Raw),
		KeyFile: writePEM("server.key", "PRIVATE KEY", keyDER), ClientCAFile: writePEM("ca.crt", "CERTIFICATE", ca.Raw),
	}
	tlsConfig, err := requisitionMandateTLS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	handler.RegisterRequisitionMandate(server, mandateFixture{}, handler.NewErrorMapper(nil))
	t.Cleanup(server.Stop)
	go func() { _ = server.Serve(listener) }()
	roots := x509.NewCertPool()
	roots.AddCert(ca)

	for _, tc := range []struct {
		name string
		uri  string
		code codes.Code
	}{
		{name: "no client certificate", code: codes.Unavailable},
		{name: "other workload", uri: "spiffe://integrione/gateway", code: codes.PermissionDenied},
		{name: "HRM workload", uri: "spiffe://integrione/human-resource", code: codes.OK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clientTLS := &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: roots, ServerName: "iam-mandate.test"}
			if tc.uri != "" {
				workload, err := url.Parse(tc.uri)
				if err != nil {
					t.Fatal(err)
				}
				clientTemplate := template(3)
				clientTemplate.URIs = []*url.URL{workload}
				clientTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
				clientCert, clientKey := issueMandateCertificate(t, clientTemplate, ca, caKey)
				clientTLS.Certificates = []tls.Certificate{{Certificate: [][]byte{clientCert.Raw}, PrivateKey: clientKey}}
			}
			conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(credentials.NewTLS(clientTLS)))
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			response, err := iamv1.NewRequisitionMandateServiceClient(conn).VerifyRequisitionMandate(ctx,
				&iamv1.VerifyRequisitionMandateRequest{PrincipalId: "reviewer", OrganizationUnitId: "unit"})
			if status.Code(err) != tc.code || (tc.code == codes.OK && !response.GetAllowed()) {
				t.Fatalf("decision = %v, %v; want code %v", response.GetAllowed(), err, tc.code)
			}
		})
	}
}
