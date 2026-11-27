package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	personnel "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
	"github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/terminationaccess"
	stdgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// StartTerminationAccess membuka listener mTLS khusus verifier IAM saat dikonfigurasi.
func StartTerminationAccess(address, certFile, keyFile, caFile string, service *personnel.Service, logger *slog.Logger, failed chan<- error) (func(), error) {
	if address == "" {
		return func() {}, nil
	}
	tlsConfig, err := loadServerTLS(certFile, keyFile, caFile)
	if err != nil {
		return nil, err
	}
	server := stdgrpc.NewServer(
		stdgrpc.Creds(credentials.NewTLS(tlsConfig)),
		stdgrpc.UnaryInterceptor(func(ctx context.Context, req any, _ *stdgrpc.UnaryServerInfo, next stdgrpc.UnaryHandler) (any, error) {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return next(ctx, req)
		}),
	)
	if err := terminationaccess.Register(server, service); err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen HRM termination access: %w", err)
	}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, stdgrpc.ErrServerStopped) {
			failed <- fmt.Errorf("serve HRM termination access: %w", err)
		}
	}()
	logger.Info("starting HRM termination access listener", "address", listener.Addr().String())
	return server.Stop, nil
}
