// Binary main menjalankan GraphQL Gateway tunggal IntegriOne pada endpoint `/api/graphql`.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/generated"
	"github.com/integrione/erp-retail/backend/gateway/graph/resolvers"
	iamadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/iam"
	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/config"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/server"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	"github.com/integrione/erp-retail/backend/shared/authsession"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load strict gateway config", slog.Any("error", err))
		os.Exit(1)
	}

	redisClient, err := authsession.NewTLSClient(cfg.Redis)
	if err != nil {
		logger.Error("failed to create gateway TLS redis client", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if cErr := redisClient.Close(); cErr != nil {
			logger.Warn("failed closing redis client", slog.Any("error", cErr))
		}
	}()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := authsession.Ping(pingCtx, redisClient); err != nil {
		pingCancel()
		logger.Error("gateway redis health check failed", slog.Any("error", err))
		os.Exit(1)
	}
	pingCancel()

	sessionStore, err := authsession.NewStore(redisClient, authsession.NewKeyspace(cfg.RedisEnv))
	if err != nil {
		logger.Error("failed to create session store", slog.Any("error", err))
		os.Exit(1)
	}

	registry, err := source_connections.NewRegistry(source_connections.Config{
		IAMTarget:  cfg.IAMGRPCTarget,
		IAMTimeout: cfg.IAMGRPCTimeout,
		HMACSecret: []byte(cfg.InternalGatewayHMACSecret),
	})
	if err != nil {
		logger.Error("failed to initialize source connections registry", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if cErr := registry.Close(); cErr != nil {
			logger.Warn("failed closing source connections registry", slog.Any("error", cErr))
		}
	}()

	iamBridge := source_connections.NewIAMRuntimeBridge(registry.IAM, cfg.IAMGRPCTimeout)
	snapshotLoader := middleware.NewSnapshotLoader(iamBridge, cfg.IAMGRPCTimeout, cfg.SnapshotMaxAge)

	authManager, err := middleware.NewAuthManager(sessionStore, snapshotLoader, middleware.SessionConfig{
		CookieName:         cfg.CookieName,
		CookieSecure:       cfg.CookieSecure,
		PreAuthTTL:         cfg.PreAuthTTL,
		SessionIdleTimeout: cfg.SessionIdleTimeout,
		SessionMaxAge:      cfg.SessionMaxAge,
	})
	if err != nil {
		logger.Error("failed to initialize auth manager", slog.Any("error", err))
		os.Exit(1)
	}

	authorizer := middleware.NewAuthorizer(authManager, iamBridge, cfg.SensitiveReauthWindow)

	originValidator, err := browserauth.NewOriginValidator(cfg.AllowedOrigins)
	if err != nil {
		logger.Error("failed to initialize origin validator", slog.Any("error", err))
		os.Exit(1)
	}
	redirectValidator := browserauth.NewRedirectValidator(cfg.AllowedRedirectPaths)

	iamAdp, err := iamadapter.NewAdapter(registry.IAM, authManager, authorizer, redirectValidator, iamadapter.Config{
		RPCTimeout:      cfg.IAMGRPCTimeout,
		LoginRateMax:    cfg.RateLimitLoginMax,
		LoginRateWindow: cfg.RateLimitLoginWindow,
		SessionMaxAge:   cfg.SessionMaxAge,
		AvatarBaseURL:   cfg.AvatarBaseURL,
	})
	if err != nil {
		logger.Error("failed to initialize IAM adapter", slog.Any("error", err))
		os.Exit(1)
	}

	execSchema := generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolvers.Resolver{
			IAMAdapter: iamAdp,
		},
		Complexity: resolvers.NewComplexityRoot(),
	})

	srv, err := server.New(server.Dependencies{
		Config:          cfg,
		Logger:          logger,
		Schema:          execSchema,
		AuthManager:     authManager,
		OriginValidator: originValidator,
	})
	if err != nil {
		logger.Error("failed to initialize gateway HTTP server", slog.Any("error", err))
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Listen()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("received shutdown signal", slog.String("signal", sig.String()))
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()
		if sErr := srv.Shutdown(shutdownCtx); sErr != nil && !errors.Is(sErr, context.Canceled) {
			logger.Error("gateway shutdown failed", slog.Any("error", sErr))
			os.Exit(1)
		}
	case listenErr := <-errCh:
		if listenErr != nil {
			logger.Error("gateway server terminated with error", slog.Any("error", listenErr))
			os.Exit(1)
		}
	}
}
