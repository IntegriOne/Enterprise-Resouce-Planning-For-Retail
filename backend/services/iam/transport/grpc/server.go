package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/interceptor"
)

// Server membungkus instance grpc.Server beserta komponen siklus hidupnya.
type Server struct {
	cfg          *config.Config
	logger       *slog.Logger
	grpcServer   *grpc.Server
	healthServer *health.Server
}

// NewServer menginisialisasi gRPC server baru dengan interceptor standar, health check, dan reflection.
func NewServer(cfg *config.Config, logger *slog.Logger, extraOpts ...grpc.ServerOption) *Server {
	var hmacSecret []byte
	var requireProof bool
	if cfg != nil {
		hmacSecret = []byte(strings.TrimSpace(cfg.InternalGatewayHMACSecret))
		envNorm := strings.ToLower(strings.TrimSpace(cfg.AppEnv))
		requireProof = envNorm != "" && envNorm != "development" && envNorm != "test" && len(hmacSecret) > 0
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryUnaryInterceptor(logger),
			interceptor.InternalContextUnaryInterceptor(hmacSecret, requireProof),
			interceptor.LoggingUnaryInterceptor(logger),
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     5 * time.Minute,
			MaxConnectionAge:      30 * time.Minute,
			MaxConnectionAgeGrace: 10 * time.Second,
			Time:                  1 * time.Minute,
			Timeout:               20 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             20 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	opts = append(opts, extraOpts...)

	grpcSrv := grpc.NewServer(opts...)
	healthSrv := health.NewServer()

	// Daftarkan standar gRPC Health Checking Protocol.
	healthpb.RegisterHealthServer(grpcSrv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	if cfg != nil {
		healthSrv.SetServingStatus(cfg.ServiceName, healthpb.HealthCheckResponse_SERVING)
		// Aktifkan gRPC reflection pada environment development untuk mempermudah inspeksi kontrak.
		if strings.EqualFold(strings.TrimSpace(cfg.AppEnv), "development") {
			reflection.Register(grpcSrv)
		}
	}

	return &Server{
		cfg:          cfg,
		logger:       logger,
		grpcServer:   grpcSrv,
		healthServer: healthSrv,
	}
}

// GRPCServer mengembalikan instance *grpc.Server internal untuk registrasi handler domain.
func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}

// Start membuka TCP listener dan menjalankan gRPC server hingga berhenti.
func (s *Server) Start() error {
	addr := s.cfg.GRPCAddress()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.logger.Info("starting IAM gRPC server",
		slog.String("service", s.cfg.ServiceName),
		slog.String("address", addr),
		slog.String("env", s.cfg.AppEnv),
	)

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("grpc server failed: %w", err)
	}
	return nil
}

// Stop menghentikan gRPC server secara graceful dengan batas waktu dari context.
func (s *Server) Stop(ctx context.Context) {
	s.logger.Info("shutting down IAM gRPC server")

	// Tandai status layanan menjadi NOT_SERVING agar caller tidak mengirim request baru.
	s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	s.healthServer.SetServingStatus(s.cfg.ServiceName, healthpb.HealthCheckResponse_NOT_SERVING)

	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("IAM gRPC server stopped gracefully")
	case <-ctx.Done():
		s.logger.Warn("graceful shutdown timed out, forcing gRPC server stop")
		s.grpcServer.Stop()
	}
}
