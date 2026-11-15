package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"github.com/integrione/erp-retail/backend/services/iam/db"
	"github.com/integrione/erp-retail/backend/services/iam/services"
)

func main() {
	// Inisialisasi logger terstruktur JSON.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Muat konfigurasi runtime IAM.
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Inisialisasi database IAM (membuat database & menerapkan migrasi baseline 000001 secara otomatis).
	dbInitCtx, dbCancel := context.WithTimeout(context.Background(), 30*time.Second)
	gormDB, err := db.Init(dbInitCtx, cfg.DatabaseDSN, false)
	dbCancel()
	if err != nil {
		logger.Error("failed to initialize IAM database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Pastikan koneksi database ditutup dengan bersih saat proses berhenti.
	if sqlDB, dbErr := gormDB.DB(); dbErr == nil {
		defer func() {
			if closeErr := sqlDB.Close(); closeErr != nil {
				logger.Warn("failed to close database connection", slog.String("error", closeErr.Error()))
			}
		}()
	}

	// Rangkai repositori GORM, client Redis TLS+ACL, dan seluruh layanan aplikasi IAM secara fail-closed.
	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	svcRegistry, repos, redisClient, err := wireRuntimeStack(initCtx, gormDB, cfg, logger)
	initCancel()
	if err != nil {
		logger.Error("failed to wire IAM runtime stack", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if redisClient != nil {
		defer func() {
			if closeErr := redisClient.Close(); closeErr != nil {
				logger.Warn("failed to close redis connection", slog.String("error", closeErr.Error()))
			}
		}()
	}

	// Jalankan bootstrap idempoten katalog RBAC & akun systemiam bila dikonfigurasi.
	bootCtx, bootCancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = services.BootstrapSystemPrincipalAndRBAC(bootCtx, repos, svcRegistry, cfg.SystemBootstrap, logger)
	bootCancel()
	if err != nil {
		logger.Error("failed to bootstrap system RBAC and principal", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Pulihkan antrean outbox pending & sesi aktif sebelum membuka security gate Redis.
	if svcRegistry.Reconciler != nil && cfg.RedisEnabled {
		recCtx, recCancel := context.WithTimeout(context.Background(), 30*time.Second)
		recErr := svcRegistry.Reconciler.BootstrapGateAndDrain(recCtx)
		recCancel()
		if recErr != nil {
			logger.Error("failed to reconcile Redis security gate on startup", slog.String("error", recErr.Error()))
			os.Exit(1)
		}
	}

	// Buat instance gRPC server dan daftarkan seluruh layanan kontrak iam.v1 secara fail-closed.
	srv, err := wireGRPCServer(cfg, svcRegistry, logger)
	if err != nil {
		logger.Error("failed to wire IAM gRPC server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Tangkap sinyal OS (SIGINT, SIGTERM) untuk graceful shutdown.
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var recWG sync.WaitGroup
	if svcRegistry.Reconciler != nil && cfg.RedisEnabled {
		recWG.Add(1)
		go func() {
			defer recWG.Done()
			svcRegistry.Reconciler.Run(sigCtx)
		}()
	}
	if svcRegistry.EmailDispatcher != nil {
		svcRegistry.EmailDispatcher.Start(sigCtx)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		logger.Error("gRPC server exited with error", slog.String("error", err.Error()))
		stop()
		recWG.Wait()
		os.Exit(1)
	case <-sigCtx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	srv.Stop(shutdownCtx)
	recWG.Wait()
	if svcRegistry.EmailDispatcher != nil {
		if closeErr := svcRegistry.EmailDispatcher.Close(shutdownCtx); closeErr != nil {
			logger.Warn("failed to drain email delivery queue on shutdown", slog.String("error", closeErr.Error()))
		}
	}
}
