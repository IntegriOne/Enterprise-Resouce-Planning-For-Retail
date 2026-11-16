package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	if err := check(); err != nil {
		os.Exit(1)
	}
}

func check() error {
	cert, err := tls.LoadX509KeyPair(os.Getenv("HRM_GRPC_TLS_CERT_FILE"), os.Getenv("HRM_GRPC_TLS_KEY_FILE"))
	if err != nil {
		return errors.New("load HRM healthcheck client certificate failed")
	}
	ca, err := os.ReadFile(os.Getenv("HRM_GRPC_TLS_CA_FILE"))
	if err != nil {
		return errors.New("load HRM healthcheck CA failed")
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(ca) {
		return errors.New("invalid HRM healthcheck CA")
	}
	serverName := os.Getenv("HRM_GRPC_TLS_SERVER_NAME")
	if serverName == "" {
		serverName = "human-resource"
	}
	port := os.Getenv("HRM_GRPC_PORT")
	if port == "" {
		port = "9003"
	}
	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort < 1 || parsedPort > 65535 {
		return errors.New("invalid HRM healthcheck port")
	}
	client, err := grpc.NewClient(net.JoinHostPort("127.0.0.1", port), grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS13, RootCAs: roots, Certificates: []tls.Certificate{cert}, ServerName: serverName,
	})))
	if err != nil {
		return errors.New("create HRM healthcheck client failed")
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := healthpb.NewHealthClient(client).Check(ctx, &healthpb.HealthCheckRequest{})
	if err != nil || result.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		return errors.New("HRM gRPC healthcheck failed")
	}
	return nil
}
