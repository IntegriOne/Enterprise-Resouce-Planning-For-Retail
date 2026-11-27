// Package grpc mengelola siklus hidup gRPC server Enterprise Foundation beserta healthcheck dan reflection.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	stdgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc/handler"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc/interceptor"
)

// ServerConfig menyimpan parameter transport gRPC server Enterprise Foundation.
type ServerConfig struct {
	ServiceName               string
	Address                   string
	EnableReflection          bool
	InternalGatewayHMACSecret []byte
	ShutdownTimeout           time.Duration
	Logger                    *slog.Logger
}

// Server membungkus instance stdgrpc.Server, health server, dan konfigurasi penghentian aman.
type Server struct {
	cfg          ServerConfig
	grpcServer   *stdgrpc.Server
	healthServer *health.Server
	logger       *slog.Logger
}

// NewServer membangun instance gRPC Server Enterprise Foundation beserta interceptor standar,
// endpoint grpc.health.v1.Health, dan opsional gRPC reflection di mode development.
func NewServer(cfg ServerConfig, deps handler.Dependencies) (*Server, error) {
	if strings.TrimSpace(cfg.ServiceName) == "" {
		return nil, fmt.Errorf("service name must not be empty")
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	keepaliveParams := keepalive.ServerParameters{
		MaxConnectionIdle: 5 * time.Minute,
		Time:              2 * time.Minute,
		Timeout:           20 * time.Second,
	}
	keepalivePolicy := keepalive.EnforcementPolicy{
		MinTime:             30 * time.Second,
		PermitWithoutStream: true,
	}

	grpcServer := stdgrpc.NewServer(
		stdgrpc.KeepaliveParams(keepaliveParams),
		stdgrpc.KeepaliveEnforcementPolicy(keepalivePolicy),
		stdgrpc.ChainUnaryInterceptor(
			interceptor.RecoveryUnaryInterceptor(logger),
			interceptor.InternalContextUnaryInterceptor(cfg.InternalGatewayHMACSecret, logger),
			interceptor.LoggingUnaryInterceptor(logger),
		),
	)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(cfg.ServiceName, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	if deps.Logger == nil {
		deps.Logger = logger
	}
	if deps.Services != nil {
		if err := handler.RegisterAll(grpcServer, deps); err != nil {
			return nil, err
		}
	} else {
		logger.Warn("enterprise_foundation domain handlers were not registered because services registry is nil")
	}

	if cfg.EnableReflection {
		reflection.Register(grpcServer)
	}

	return &Server{
		cfg:          cfg,
		grpcServer:   grpcServer,
		healthServer: healthServer,
		logger:       logger,
	}, nil
}

// GRPCServer mengembalikan pointer *stdgrpc.Server internal untuk kebutuhan registrasi lanjutan atau pengujian.
func (s *Server) GRPCServer() *stdgrpc.Server {
	return s.grpcServer
}

// SetServingStatus memperbarui status kesehatan layanan secara keseluruhan dan per nama layanan.
func (s *Server) SetServingStatus(status healthpb.HealthCheckResponse_ServingStatus) {
	if s.healthServer == nil {
		return
	}
	s.healthServer.SetServingStatus("", status)
	s.healthServer.SetServingStatus(s.cfg.ServiceName, status)
}

// ListenAndServe membuka TCP listener pada cfg.Address dan menjalankan gRPC server hingga berhenti.
func (s *Server) ListenAndServe() error {
	if strings.TrimSpace(s.cfg.Address) == "" {
		return fmt.Errorf("gRPC listen address must not be empty")
	}

	lis, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return fmt.Errorf("listen gRPC %s: %w", s.cfg.Address, err)
	}

	return s.Serve(lis)
}

// Serve menjalankan gRPC server di atas listener jaringan yang diberikan.
func (s *Server) Serve(lis net.Listener) error {
	if lis == nil {
		return fmt.Errorf("listener must not be nil")
	}

	s.logger.Info("Enterprise Foundation gRPC server listening",
		slog.String("service", s.cfg.ServiceName),
		slog.String("address", lis.Addr().String()),
		slog.Bool("reflection", s.cfg.EnableReflection),
	)

	if err := s.grpcServer.Serve(lis); err != nil && !errors.Is(err, stdgrpc.ErrServerStopped) {
		return fmt.Errorf("serve gRPC: %w", err)
	}
	return nil
}

// Stop menghentikan gRPC server secara graceful dengan batas waktu sebelum memaksa berhenti.
func (s *Server) Stop(ctx context.Context) error {
	s.healthServer.Shutdown()

	timeout := s.cfg.ShutdownTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	stopCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-stopCtx.Done():
		s.grpcServer.Stop()
		<-done
		return fmt.Errorf("graceful shutdown timed out, forced stop: %w", stopCtx.Err())
	}
}
