package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func loadServerTLS(certFile, keyFile, caFile string) (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load HRM server certificate: %w", err)
	}
	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("load HRM client CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("invalid HRM client CA")
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}, nil
}
