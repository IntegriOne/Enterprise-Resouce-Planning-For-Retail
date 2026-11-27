// Program grpc-healthcheck melakukan pemeriksaan kesehatan gRPC standar pada localhost untuk Docker HEALTHCHECK.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	defaultPort := strings.TrimSpace(os.Getenv("EF_GRPC_PORT"))
	if defaultPort == "" {
		defaultPort = "9002"
	}

	addr := flag.String("addr", "127.0.0.1:"+defaultPort, "gRPC server address to probe")
	service := flag.String("service", strings.TrimSpace(os.Getenv("EF_SERVICE_NAME")), "gRPC service name to probe")
	timeout := flag.Duration("timeout", 2*time.Second, "probe RPC timeout")
	flag.Parse()

	if err := checkHealth(context.Background(), *addr, *service, *timeout); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// checkHealth menghubungi endpoint grpc.health.v1.Health/Check dan memastikan status layanan SERVING.
func checkHealth(ctx context.Context, target, service string, timeout time.Duration) error {
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("healthcheck target address is required")
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("create gRPC health client: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	resp, err := healthpb.NewHealthClient(conn).Check(probeCtx, &healthpb.HealthCheckRequest{
		Service: service,
	})
	if err != nil {
		return fmt.Errorf("gRPC health check RPC failed: %w", err)
	}
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("gRPC health status is %s, want SERVING", resp.GetStatus().String())
	}
	return nil
}
