package main

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestCheckHealth(t *testing.T) {
	t.Parallel()

	t.Run("rejects empty target address", func(t *testing.T) {
		t.Parallel()
		if err := checkHealth(context.Background(), "   ", "", time.Second); err == nil {
			t.Fatal("expected error for empty target address")
		}
	})

	t.Run("verifies SERVING and rejects NOT_SERVING", func(t *testing.T) {
		t.Parallel()

		lis, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen: %v", err)
		}
		defer func() { _ = lis.Close() }()

		srv := grpc.NewServer()
		hs := health.NewServer()
		hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
		hs.SetServingStatus("integrione-enterprise-foundation", healthpb.HealthCheckResponse_SERVING)
		hs.SetServingStatus("degraded-svc", healthpb.HealthCheckResponse_NOT_SERVING)
		healthpb.RegisterHealthServer(srv, hs)

		go func() {
			_ = srv.Serve(lis)
		}()
		t.Cleanup(srv.Stop)

		addr := lis.Addr().String()
		if err := checkHealth(context.Background(), addr, "integrione-enterprise-foundation", time.Second); err != nil {
			t.Fatalf("expected SERVING check to pass, got %v", err)
		}
		if err := checkHealth(context.Background(), addr, "", 0); err != nil {
			t.Fatalf("expected default timeout check to pass, got %v", err)
		}
		if err := checkHealth(context.Background(), addr, "degraded-svc", time.Second); err == nil {
			t.Fatal("expected NOT_SERVING service to fail checkHealth")
		}
		if err := checkHealth(context.Background(), addr, "unknown-svc", time.Second); err == nil {
			t.Fatal("expected unknown service to fail checkHealth")
		}
	})
}
