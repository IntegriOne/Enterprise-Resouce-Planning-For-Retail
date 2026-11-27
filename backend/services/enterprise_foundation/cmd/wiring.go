package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/config"
	efdb "github.com/integrione/erp-retail/backend/services/enterprise_foundation/db"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/infrastructure/outbox"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc/handler"
)

// RuntimeDependencies menyimpan dependensi yang diinisialisasi saat startup proses Enterprise Foundation.
type RuntimeDependencies struct {
	DB          *gorm.DB
	Repos       *repository.Registry
	Services    *services.Registry
	OutboxRelay *outbox.Relay
	HandlerDeps handler.Dependencies
}

// BuildDependencies menyiapkan seluruh dependensi aplikasi secara berurutan (termasuk koneksi
// PostgreSQL dan migrasi otomatis saat DSN tersedia) dan mengembalikan fungsi cleanup.
func BuildDependencies(ctx context.Context, cfg config.Config, logger *slog.Logger) (*RuntimeDependencies, func(context.Context) error, error) {
	deps := &RuntimeDependencies{
		HandlerDeps: handler.Dependencies{
			Logger: logger,
		},
	}

	cleanup := func(context.Context) error {
		if deps.DB == nil {
			return nil
		}
		sqlDB, err := deps.DB.DB()
		if err != nil {
			return fmt.Errorf("obtain sql.DB handle during cleanup: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("close database connection pool: %w", err)
		}
		return nil
	}

	dsn := strings.TrimSpace(cfg.DatabaseDSN)
	if cfg.IsProduction() && dsn == "" {
		return nil, nil, fmt.Errorf("EF_DATABASE_DSN must be set in production")
	}

	if dsn != "" {
		gormDB, err := efdb.Init(ctx, dsn, false)
		if err != nil {
			return nil, nil, fmt.Errorf("initialize Enterprise Foundation database: %w", err)
		}
		deps.DB = gormDB

		repos, err := repository.New(gormDB)
		if err != nil {
			_ = cleanup(ctx)
			return nil, nil, fmt.Errorf("initialize Enterprise Foundation repositories: %w", err)
		}

		relay, err := buildOutboxRelay(repos, cfg, logger)
		if err != nil {
			_ = cleanup(ctx)
			return nil, nil, fmt.Errorf("initialize Enterprise Foundation outbox relay: %w", err)
		}

		// Bungkus OutboxRepository khusus lapisan service agar transaksi yang commit langsung membangunkan relay.
		serviceRepos := *repos
		serviceRepos.Outbox = outbox.WrapNotifyingRepository(repos.Outbox, relay, pgstore.RegisterAfterCommit)
		svcReg, err := services.NewRegistry(&serviceRepos, services.Options{})
		if err != nil {
			_ = cleanup(ctx)
			return nil, nil, fmt.Errorf("initialize Enterprise Foundation services: %w", err)
		}
		deps.Repos = repos
		deps.Services = svcReg
		deps.OutboxRelay = relay
		deps.HandlerDeps.Services = svcReg

		if logger != nil {
			logger.Info(
				"Enterprise Foundation PostgreSQL database, repositories, services, and outbox relay initialized",
				slog.String("outbox_mode", string(relay.Mode())),
			)
		}
	} else if logger != nil {
		logger.Warn("starting Enterprise Foundation server without database connection because EF_DATABASE_DSN is empty")
	}

	return deps, cleanup, nil
}

func buildOutboxRelay(repos *repository.Registry, cfg config.Config, logger *slog.Logger) (*outbox.Relay, error) {
	if err := cfg.ValidateOutbox(); err != nil {
		return nil, err
	}

	defaults := outbox.DefaultConfig()
	relayCfg := outbox.Config{
		PollInterval:    defaults.PollInterval,
		BatchSize:       defaults.BatchSize,
		MaxAttempts:     defaults.MaxAttempts,
		PublishTimeout:  defaults.PublishTimeout,
		CleanupInterval: defaults.CleanupInterval,
		RetentionPeriod: defaults.RetentionPeriod,
		Backoff:         defaults.Backoff,
	}
	if cfg.Outbox.PollInterval > 0 {
		relayCfg.PollInterval = cfg.Outbox.PollInterval
	}
	if cfg.Outbox.BatchSize > 0 {
		relayCfg.BatchSize = cfg.Outbox.BatchSize
	}
	if cfg.Outbox.MaxAttempts > 0 {
		relayCfg.MaxAttempts = cfg.Outbox.MaxAttempts
	}
	if cfg.Outbox.PublishTimeout > 0 {
		relayCfg.PublishTimeout = cfg.Outbox.PublishTimeout
	}
	hasLoadedOutbox := cfg.Outbox.PollInterval > 0 ||
		cfg.Outbox.BatchSize > 0 ||
		cfg.Outbox.MaxAttempts > 0 ||
		strings.TrimSpace(cfg.Outbox.Mode) != ""
	if cfg.Outbox.CleanupInterval > 0 || hasLoadedOutbox {
		relayCfg.CleanupInterval = cfg.Outbox.CleanupInterval
	}
	if cfg.Outbox.RetentionPeriod > 0 || hasLoadedOutbox {
		relayCfg.RetentionPeriod = cfg.Outbox.RetentionPeriod
	}

	var publisher outbox.Publisher
	switch cfg.ResolvedOutboxMode() {
	case "standby":
		relayCfg.Standby = true
	case "log":
		publisher = outbox.NewLogPublisher(logger, nil)
	case "webhook":
		pub, err := outbox.NewWebhookPublisher(outbox.WebhookConfig{
			Endpoint:     cfg.Outbox.WebhookURL,
			Timeout:      cfg.Outbox.WebhookTimeout,
			HMACSecret:   cfg.Outbox.WebhookHMACSecret,
			RequireHTTPS: cfg.Outbox.WebhookRequireHTTPS,
		})
		if err != nil {
			return nil, err
		}
		publisher = pub
	default:
		return nil, fmt.Errorf("unsupported outbox mode: %s", cfg.ResolvedOutboxMode())
	}

	return outbox.NewRelay(outbox.Dependencies{
		Transactor: repos.Transactor,
		Repo:       repos.Outbox,
		Publisher:  publisher,
		Logger:     logger,
	}, relayCfg)
}
