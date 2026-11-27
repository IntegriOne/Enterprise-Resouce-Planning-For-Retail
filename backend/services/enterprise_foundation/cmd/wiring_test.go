package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/config"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/infrastructure/outbox"
	transportgrpc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc"
)

func TestBuildDependencies_ProductionRequiresDSN(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{
		AppEnv:                    "production",
		ServiceName:               "integrione-enterprise-foundation",
		GRPCHost:                  "127.0.0.1",
		GRPCPort:                  9002,
		ShutdownTimeout:           5 * time.Second,
		InternalGatewayHMACSecret: []byte("0123456789abcdef0123456789abcdef"),
	}

	if _, _, err := BuildDependencies(context.Background(), cfg, logger); err == nil {
		t.Fatal("expected production BuildDependencies without DSN to fail")
	}
}

func TestBuildDependencies_DevelopmentWithoutDSN(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{
		AppEnv:                    "development",
		ServiceName:               "integrione-enterprise-foundation",
		GRPCHost:                  "127.0.0.1",
		GRPCPort:                  9002,
		ShutdownTimeout:           5 * time.Second,
		InternalGatewayHMACSecret: []byte("0123456789abcdef0123456789abcdef"),
	}

	deps, cleanup, err := BuildDependencies(context.Background(), cfg, logger)
	if err != nil {
		t.Fatalf("expected development BuildDependencies without DSN to succeed, got %v", err)
	}
	if deps == nil || deps.DB != nil || deps.Repos != nil || deps.Services != nil || deps.HandlerDeps.Services != nil {
		t.Fatalf("expected empty database/service wiring when DSN is empty, got %+v", deps)
	}
	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}
	if err := cleanup(context.Background()); err != nil {
		t.Fatalf("cleanup returned error: %v", err)
	}
}

func TestBuildDependencies_WithLiveDatabase(t *testing.T) {
	testDSN := strings.TrimSpace(os.Getenv("EF_TEST_DATABASE_DSN"))
	if testDSN == "" {
		testDSN = strings.TrimSpace(os.Getenv("EF_DATABASE_DSN"))
	}
	if testDSN == "" {
		t.Skip("skipping live PostgreSQL wiring test: set EF_TEST_DATABASE_DSN or EF_DATABASE_DSN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{
		AppEnv:                    "development",
		ServiceName:               "integrione-enterprise-foundation",
		GRPCHost:                  "127.0.0.1",
		GRPCPort:                  0,
		ShutdownTimeout:           5 * time.Second,
		DatabaseDSN:               testDSN,
		InternalGatewayHMACSecret: []byte("0123456789abcdef0123456789abcdef"),
	}

	deps, cleanup, err := BuildDependencies(ctx, cfg, logger)
	if err != nil {
		t.Fatalf("BuildDependencies with live DSN failed: %v", err)
	}
	t.Cleanup(func() {
		if cleanup != nil {
			_ = cleanup(context.Background())
		}
	})

	if deps.DB == nil || deps.Repos == nil || deps.Services == nil || deps.OutboxRelay == nil || deps.HandlerDeps.Services == nil {
		t.Fatalf("expected complete wired runtime dependencies, got %+v", deps)
	}
	if deps.OutboxRelay.Mode() != outbox.ModeStandby {
		t.Fatalf("expected default outbox relay mode standby, got %q", deps.OutboxRelay.Mode())
	}

	srv, err := transportgrpc.NewServer(transportgrpc.ServerConfig{
		ServiceName:               cfg.ServiceName,
		Address:                   "127.0.0.1:0",
		EnableReflection:          true,
		InternalGatewayHMACSecret: cfg.InternalGatewayHMACSecret,
		ShutdownTimeout:           cfg.ShutdownTimeout,
		Logger:                    logger,
	}, deps.HandlerDeps)
	if err != nil {
		t.Fatalf("NewServer with wired deps failed: %v", err)
	}
	if srv == nil {
		t.Fatal("expected non-nil gRPC server")
	}

	t.Run("webhook mode builds active outbox relay", func(t *testing.T) {
		webhookCfg := cfg
		webhookCfg.Outbox = config.OutboxConfig{
			Mode:       "webhook",
			WebhookURL: "http://127.0.0.1:18080/events",
		}
		whDeps, whCleanup, err := BuildDependencies(ctx, webhookCfg, logger)
		if err != nil {
			t.Fatalf("BuildDependencies webhook mode failed: %v", err)
		}
		defer func() {
			if whCleanup != nil {
				_ = whCleanup(context.Background())
			}
		}()
		if whDeps.OutboxRelay == nil || whDeps.OutboxRelay.Mode() != outbox.ModeActive {
			t.Fatalf("expected active outbox relay, got %+v", whDeps.OutboxRelay)
		}
	})
}
