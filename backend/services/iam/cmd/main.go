package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
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

	if err := run(logger); err != nil {
		logger.Error("iam service terminated with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// run mengeksekusi siklus hidup penuh daemon IAM sehingga seluruh defer pembersihan koneksi dan antrean tetap berjalan saat gagal.
func run(logger *slog.Logger) error {
	// Muat konfigurasi runtime IAM.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load iam configuration: %w", err)
	}

	// Inisialisasi database IAM (membuat database & menerapkan migrasi baseline 000001 secara otomatis).
	dbInitCtx, dbCancel := context.WithTimeout(context.Background(), 30*time.Second)
	gormDB, err := db.Init(dbInitCtx, cfg.DatabaseDSN, false)
	dbCancel()
	if err != nil {
		return fmt.Errorf("initialize iam database: %w", err)
	}

	// Pastikan koneksi database ditutup dengan bersih saat proses berhenti.
	sqlDB, dbErr := gormDB.DB()
	if dbErr != nil {
		logger.Warn("failed to acquire sql.DB handle for cleanup", slog.String("error", dbErr.Error()))
	} else {
		defer func() {
			if closeErr := sqlDB.Close(); closeErr != nil {
				logger.Warn("failed to close database connection", slog.String("error", closeErr.Error()))
			}
		}()
	}

	// Rangkai repositori GORM, client Redis TLS+ACL, dan seluruh layanan aplikasi IAM secara fail-closed.
	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	svcRegistry, repos, redisClient, efScopeCloser, err := wireRuntimeStack(initCtx, gormDB, cfg, logger)
	initCancel()
	if err != nil {
		return fmt.Errorf("wire iam runtime stack: %w", err)
	}
	if efScopeCloser != nil {
		defer func() {
			if closeErr := efScopeCloser.Close(); closeErr != nil {
				logger.Warn("failed to close enterprise foundation scope verifier connection", slog.String("error", closeErr.Error()))
			}
		}()
	}
	if redisClient != nil {
		defer func() {
			if closeErr := redisClient.Close(); closeErr != nil {
				logger.Warn("failed to close redis connection", slog.String("error", closeErr.Error()))
			}
		}()
	}
	if svcRegistry.EmailDispatcher != nil {
		defer func() {
			drainTimeout := cfg.ShutdownTimeout
			if drainTimeout <= 0 {
				drainTimeout = 10 * time.Second
			}
			closeCtx, closeCancel := context.WithTimeout(context.Background(), drainTimeout)
			defer closeCancel()
			if closeErr := svcRegistry.EmailDispatcher.Close(closeCtx); closeErr != nil {
				logger.Warn("failed to drain email delivery queue on shutdown", slog.String("error", closeErr.Error()))
			}
		}()
	}

	// Jalankan bootstrap idempoten katalog RBAC, akun layanan otomasi per management, serta akun human superadmin perdana.
	bootCtx, bootCancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = services.BootstrapPrincipalsAndRBAC(bootCtx, repos, svcRegistry, cfg.Bootstrap, logger)
	bootCancel()
	if err != nil {
		return fmt.Errorf("bootstrap rbac and initial principals: %w", err)
	}

	// Pulihkan antrean outbox pending & sesi aktif sebelum membuka security gate Redis.
	if svcRegistry.Reconciler != nil && cfg.RedisEnabled {
		recCtx, recCancel := context.WithTimeout(context.Background(), 30*time.Second)
		recErr := svcRegistry.Reconciler.BootstrapGateAndDrain(recCtx)
		recCancel()
		if recErr != nil {
			return fmt.Errorf("reconcile redis security gate on startup: %w", recErr)
		}
	}

	// Buat instance gRPC server dan daftarkan seluruh layanan kontrak iam.v1 secara fail-closed.
	srv, err := wireGRPCServer(cfg, svcRegistry, logger)
	if err != nil {
		return fmt.Errorf("wire iam grpc server: %w", err)
	}

	errCh := make(chan error, 2)
	stopWorkloads, err := startHRMWorkloads(cfg, svcRegistry.Users, svcRegistry.EmployeeProvisioning, logger, errCh)
	if err != nil {
		return err
	}
	defer stopWorkloads()

	// Tangkap sinyal OS (SIGINT, SIGTERM) untuk graceful shutdown.
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var recWG sync.WaitGroup
	if svcRegistry.Reconciler != nil && cfg.RedisEnabled {
		recWG.Add(1)
		go func() {
			defer recWG.Done()
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("iam reconciler goroutine panic recovered",
						slog.Any("panic", rec),
						slog.String("stack", string(debug.Stack())),
					)
				}
			}()
			svcRegistry.Reconciler.Run(sigCtx)
		}()
	}
	if svcRegistry.EmailDispatcher != nil {
		svcRegistry.EmailDispatcher.Start(sigCtx)
	}

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("iam gRPC server goroutine panic recovered",
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				)
				errCh <- fmt.Errorf("grpc server panicked: %v", rec)
			}
		}()
		if serveErr := srv.Start(); serveErr != nil {
			errCh <- serveErr
		}
	}()

	var serveErr error
	select {
	case serveErr = <-errCh:
		stop()
	case <-sigCtx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	srv.Stop(shutdownCtx)
	recWG.Wait()
	if serveErr != nil {
		return fmt.Errorf("grpc server exited with error: %w", serveErr)
	}
	return nil
}
