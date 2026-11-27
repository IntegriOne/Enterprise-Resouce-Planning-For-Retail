// Package grpc mengelola transport privat HRM dan penghentian request yang tertib.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	appservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/interceptor"
	stdgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

// Config berisi konfigurasi transport dan bukti konteks pemanggil internal.
type Config struct {
	Address           string
	ServiceName       string
	TLSCertFile       string
	TLSKeyFile        string
	TLSCAFile         string
	GatewayHMACSecret []byte
	ShutdownTimeout   time.Duration
}

// Server menyimpan lifecycle gRPC yang hanya menerima koneksi mTLS.
type Server struct {
	cfg    Config
	server *stdgrpc.Server
	health *health.Server
}

// NewServer menolak konfigurasi keamanan kosong sebelum mendaftarkan domain.
func NewServer(cfg Config, services *appservice.Registry, logger *slog.Logger) (*Server, error) {
	if services == nil || cfg.Address == "" || cfg.ServiceName == "" || len(cfg.GatewayHMACSecret) < 32 {
		return nil, fmt.Errorf("services, address and gateway proof key are required")
	}
	tlsConfig, err := loadServerTLS(cfg.TLSCertFile, cfg.TLSKeyFile, cfg.TLSCAFile)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	key := append([]byte{}, cfg.GatewayHMACSecret...)
	server := stdgrpc.NewServer(
		stdgrpc.Creds(credentials.NewTLS(tlsConfig)),
		stdgrpc.MaxRecvMsgSize(4<<20),
		stdgrpc.MaxSendMsgSize(4<<20),
		stdgrpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              2 * time.Minute,
			Timeout:           20 * time.Second,
		}),
		stdgrpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             20 * time.Second,
			PermitWithoutStream: true,
		}),
		stdgrpc.ChainUnaryInterceptor(interceptor.Unary(key, logger)),
	)
	if err := handler.RegisterAll(server, services); err != nil {
		server.Stop()
		return nil, fmt.Errorf("register HRM handlers: %w", err)
	}
	h := health.NewServer()
	h.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	h.SetServingStatus(cfg.ServiceName, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, h)
	return &Server{cfg: cfg, server: server, health: h}, nil
}

// ListenAndServe membuka listener privat setelah wiring selesai.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return fmt.Errorf("listen HRM gRPC: %w", err)
	}
	if err := s.server.Serve(listener); err != nil && !errors.Is(err, stdgrpc.ErrServerStopped) {
		return fmt.Errorf("serve HRM gRPC: %w", err)
	}
	return nil
}

// Stop membatasi waktu drain sebelum menghentikan koneksi yang belum selesai.
func (s *Server) Stop(ctx context.Context) error {
	s.health.Shutdown()
	timeout := s.cfg.ShutdownTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return DrainServer(ctx, s.server)
}

// DrainServer memberi RPC aktif waktu selesai hingga batas context pemanggil.
func DrainServer(ctx context.Context, server *stdgrpc.Server) error {
	done := make(chan struct{})
	go func() { server.GracefulStop(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		server.Stop()
		// GracefulStop masih dapat menunggu handler yang mengabaikan pembatalan.
		return fmt.Errorf("force HRM gRPC shutdown: %w", ctx.Err())
	}
}
