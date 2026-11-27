// Command enterprise-foundation menjalankan proses gRPC server untuk layanan Enterprise Foundation.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/config"
	transportgrpc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load Enterprise Foundation configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	deps, cleanup, err := BuildDependencies(signalCtx, cfg, logger)
	if err != nil {
		logger.Error("failed to build Enterprise Foundation dependencies", slog.String("error", err.Error()))
		os.Exit(1)
	}

	srv, err := transportgrpc.NewServer(transportgrpc.ServerConfig{
		ServiceName:               cfg.ServiceName,
		Address:                   cfg.GRPCAddress(),
		EnableReflection:          cfg.IsDevelopment(),
		InternalGatewayHMACSecret: cfg.InternalGatewayHMACSecret,
		ShutdownTimeout:           cfg.ShutdownTimeout,
		Logger:                    logger,
	}, deps.HandlerDeps)
	if err != nil {
		logger.Error("failed to initialize Enterprise Foundation gRPC server", slog.String("error", err.Error()))
		_ = cleanup(context.Background())
		os.Exit(1)
	}

	relayCtx, cancelRelay := context.WithCancel(signalCtx)
	defer cancelRelay()

	var relayWG sync.WaitGroup
	relayErrCh := make(chan error, 1)
	if deps.OutboxRelay != nil {
		relayWG.Add(1)
		go func() {
			defer relayWG.Done()
			if runErr := deps.OutboxRelay.Run(relayCtx); runErr != nil {
				relayErrCh <- runErr
			}
		}()
	}

	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- srv.ListenAndServe()
	}()

	waitOutboxStop := func(waitCtx context.Context) {
		cancelRelay()
		relayDone := make(chan struct{})
		go func() {
			relayWG.Wait()
			close(relayDone)
		}()
		select {
		case <-relayDone:
		case <-waitCtx.Done():
			logger.Warn("timed out waiting for outbox relay worker to stop cleanly")
		}
	}

	var exitErr error
	select {
	case err := <-serveErrCh:
		exitErr = err
		stopCtx, cancelStop := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		waitOutboxStop(stopCtx)
		cancelStop()
	case err := <-relayErrCh:
		logger.Error("outbox relay exited with unexpected error", slog.String("error", err.Error()))
		exitErr = err
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancelShutdown()
		if stopErr := srv.Stop(shutdownCtx); stopErr != nil {
			logger.Error("error during gRPC server shutdown", slog.String("error", stopErr.Error()))
		}
		<-serveErrCh
		waitOutboxStop(shutdownCtx)
	case <-signalCtx.Done():
		logger.Info("shutdown signal received, stopping Enterprise Foundation gRPC server")
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancelShutdown()

		if err := srv.Stop(shutdownCtx); err != nil {
			logger.Error("error during gRPC server shutdown", slog.String("error", err.Error()))
			exitErr = err
		}
		if err := <-serveErrCh; err != nil && exitErr == nil {
			exitErr = err
		}
		waitOutboxStop(shutdownCtx)
	}

	cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelCleanup()
	if err := cleanup(cleanupCtx); err != nil {
		logger.Error("error during dependency cleanup", slog.String("error", err.Error()))
		if exitErr == nil {
			exitErr = err
		}
	}

	if exitErr != nil {
		logger.Error("Enterprise Foundation server exited with error", slog.String("error", exitErr.Error()))
		os.Exit(1)
	}

	logger.Info("Enterprise Foundation server stopped cleanly")
}
