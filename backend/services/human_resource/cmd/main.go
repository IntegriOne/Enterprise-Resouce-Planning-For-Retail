package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/integrione/erp-retail/backend/services/human_resource/config"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	transportgrpc "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc"
	"github.com/integrione/erp-retail/backend/shared/authsession"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("HRM runtime stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load HRM configuration: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	startup, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancel()
	client, err := authsession.NewTLSClient(cfg.SecurityRedis)
	if err != nil {
		return errors.New("configure HRM session Redis TLS client failed")
	}
	defer func() {
		if err := client.Close(); err != nil {
			logger.Error("close HRM session Redis client failed")
		}
	}()
	if err := client.Ping(startup).Err(); err != nil {
		return errors.New("connect HRM session Redis failed")
	}
	store, err := authsession.NewStore(client, authsession.NewKeyspace(cfg.SecurityNamespace))
	if err != nil {
		return errors.New("configure HRM session store failed")
	}
	policy, err := security.NewPolicy(store, security.PermissionCatalog())
	if err != nil {
		return fmt.Errorf("configure HRM security policy: %w", err)
	}
	dependencies, err := BuildDependencies(startup, cfg, policy, security.NewIntegrity(nil))
	if err != nil {
		return fmt.Errorf("initialize HRM dependencies: %w", err)
	}
	defer func() {
		if err := dependencies.Close(); err != nil {
			logger.Error("close HRM database failed")
		}
	}()
	server, err := transportgrpc.NewServer(transportgrpc.Config{
		Address: cfg.GRPCAddress(), ServiceName: cfg.ServiceName, TLSCertFile: cfg.TLS.CertFile,
		TLSKeyFile: cfg.TLS.KeyFile, TLSCAFile: cfg.TLS.CAFile,
		GatewayHMACSecret: cfg.GatewayHMACSecret, ShutdownTimeout: cfg.ShutdownTimeout,
	}, dependencies.Services, logger)
	if err != nil {
		return fmt.Errorf("configure HRM gRPC: %w", err)
	}
	if err := startup.Err(); err != nil {
		return err
	}
	cancel()
	logger.Warn("HRM mutation validators are not activated; mutations remain unavailable")
	logger.Info("starting private HRM gRPC listener", "address", cfg.GRPCAddress())
	served := make(chan error, 1)
	go func() { served <- server.ListenAndServe() }()
	select {
	case err := <-served:
		if err == nil {
			return errors.New("HRM gRPC server stopped unexpectedly")
		}
		return err
	case <-ctx.Done():
		shutdown, finish := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer finish()
		stopErr := server.Stop(shutdown)
		serveErr := <-served
		return errors.Join(stopErr, serveErr)
	}
}
