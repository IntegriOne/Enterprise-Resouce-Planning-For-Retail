package main

import (
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestCheckHealth(t *testing.T) {
	t.Parallel()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	srv := grpc.NewServer()
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	hs.SetServingStatus("integrione-iam", healthpb.HealthCheckResponse_NOT_SERVING)
	healthpb.RegisterHealthServer(srv, hs)

	go func() {
		_ = srv.Serve(lis)
	}()
	t.Cleanup(srv.Stop)

	addr := lis.Addr().String()
	if err := checkHealth(t.Context(), addr, "", 2*time.Second); err != nil {
		t.Fatalf("expected SERVING for default service, got error = %v", err)
	}

	if err := checkHealth(t.Context(), addr, "integrione-iam", 2*time.Second); err == nil {
		t.Fatal("expected error when service status is NOT_SERVING")
	}

	if err := checkHealth(t.Context(), "127.0.0.1:1", "", 300*time.Millisecond); err == nil {
		t.Fatal("expected error when dialing unreachable port")
	}
}
