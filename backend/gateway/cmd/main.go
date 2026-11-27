// Binary main menjalankan GraphQL Gateway tunggal IntegriOne pada endpoint `/api/graphql`.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/graph/generated"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/graph/resolvers"
	efadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/enterprise_foundation"
	hrmadapter "github.com/integrione/erp-retail/backend/gateway/internal/adapter/human_resource"
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

	if err := run(logger); err != nil {
		logger.Error("gateway terminated with failure", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load strict gateway config: %w", err)
	}

	redisClient, err := authsession.NewTLSClient(cfg.Redis)
	if err != nil {
		return fmt.Errorf("create gateway TLS redis client: %w", err)
	}
	defer func() {
		if cErr := redisClient.Close(); cErr != nil {
			logger.Warn("failed closing redis client", slog.Any("error", cErr))
		}
	}()

	// Kegagalan koneksi Redis saat startup tidak menghentikan proses Gateway; seluruh operasi tetap fail-closed UNAVAILABLE hingga Redis pulih otomatis.
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := authsession.Ping(pingCtx, redisClient); err != nil {
		logger.Warn("gateway redis unreachable on startup; running in fail-closed degraded mode until redis recovers", slog.Any("error", err))
	}
	pingCancel()

	sessionStore, err := authsession.NewStore(redisClient, authsession.NewKeyspace(cfg.RedisEnv))
	if err != nil {
		return fmt.Errorf("create session store: %w", err)
	}

	registry, err := source_connections.NewRegistry(source_connections.Config{
		IAMTarget:      cfg.IAMGRPCTarget,
		IAMTimeout:     cfg.IAMGRPCTimeout,
		EFEnabled:      cfg.EFGRPCEnabled,
		EFTarget:       cfg.EFGRPCTarget,
		EFTimeout:      cfg.EFGRPCTimeout,
		HRMEnabled:     cfg.HRMGRPCEnabled,
		HRMConfigError: cfg.HRMGRPCConfigError,
		HRM: source_connections.HumanResourceConnectionConfig{
			Target:         cfg.HRMGRPCTarget,
			ServerName:     cfg.HRMGRPCServerName,
			CAFile:         cfg.HRMGRPCCAFile,
			CertFile:       cfg.HRMGRPCCertFile,
			KeyFile:        cfg.HRMGRPCKeyFile,
			HMACSecretFile: cfg.HRMGRPCHMACSecretFile,
			Timeout:        cfg.HRMGRPCTimeout,
		},
		HMACSecret: []byte(cfg.InternalGatewayHMACSecret),
	})
	if err != nil {
		return fmt.Errorf("initialize source connections registry: %w", err)
	}
	if registry.IAMInitError != nil {
		// Kegagalan inisialisasi koneksi IAM tidak boleh mematikan proses Gateway atau EF; trafik IAM tetap fail-closed UNAVAILABLE.
		logger.Warn("IAM connection degraded on startup; running fail-closed until IAM recovers", slog.Any("error", registry.IAMInitError))
	}
	if registry.EFInitError != nil {
		// Kegagalan inisialisasi koneksi Enterprise Foundation tidak boleh menghentikan startup Gateway maupun IAM.
		logger.Warn("enterprise foundation connection degraded on startup; continuing without blocking IAM", slog.Any("error", registry.EFInitError))
	}
	if registry.HRMInitError != nil {
		logger.Warn("human resource connection degraded on startup", slog.Any("error", registry.HRMInitError))
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
		return fmt.Errorf("initialize auth manager: %w", err)
	}

	authorizer := middleware.NewAuthorizer(authManager, iamBridge, cfg.SensitiveReauthWindow)

	originValidator, err := browserauth.NewOriginValidator(cfg.AllowedOrigins)
	if err != nil {
		return fmt.Errorf("initialize origin validator: %w", err)
	}
	redirectValidator := browserauth.NewRedirectValidator(cfg.AllowedRedirectPaths)

	efAdp, efAdpErr := efadapter.NewAdapter(registry.EnterpriseFoundation, authorizer, cfg.EFGRPCTimeout)
	if efAdpErr != nil {
		// Tetap izinkan Gateway berjalan bila adapter EF terdegradasi agar layanan IAM tidak terdampak.
		logger.Warn("enterprise foundation adapter degraded on startup; EF operations unavailable", slog.Any("error", efAdpErr))
	}

	hrmAdp := hrmadapter.NewAdapter(authManager, authorizer, registry.HumanResource, hrmadapter.WithEmployeeRequestDelivery(registry.IAM))
	var iamOpts []iamadapter.Option
	if registry.HumanResource != nil {
		key, keyErr := source_connections.HumanResourceSigningKey(cfg.HRMGRPCHMACSecretFile, []byte(cfg.InternalGatewayHMACSecret))
		if keyErr != nil {
			logger.Warn("employee provisioning unavailable", slog.Any("error", keyErr))
		} else {
			iamOpts = append(iamOpts, iamadapter.WithWorkerPrincipalProvisioner(hrmadapter.NewWorkerProvisioner(hrmAdp, key)))
		}
	}
	if efAdp != nil {
		iamOpts = append(iamOpts, iamadapter.WithPartyPortalProvisioner(efPortalProvisionerBridge{ef: efAdp}))
	}

	iamAdp, err := iamadapter.NewAdapter(registry.IAM, authManager, authorizer, redirectValidator, iamadapter.Config{
		RPCTimeout:      cfg.IAMGRPCTimeout,
		LoginRateMax:    cfg.RateLimitLoginMax,
		LoginRateWindow: cfg.RateLimitLoginWindow,
		SessionMaxAge:   cfg.SessionMaxAge,
		AvatarBaseURL:   cfg.AvatarBaseURL,
	}, iamOpts...)
	if err != nil {
		return fmt.Errorf("initialize IAM adapter: %w", err)
	}

	execSchema := generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolvers.Resolver{
			IAMAdapter: iamAdp,
			EFAdapter:  efAdp,
			HRMAdapter: hrmAdp,
		},
		Complexity: resolvers.NewComplexityRoot(cfg.MaxComplexity),
	})

	srv, err := server.New(server.Dependencies{
		Config:          cfg,
		Logger:          logger,
		Schema:          execSchema,
		AuthManager:     authManager,
		OriginValidator: originValidator,
	})
	if err != nil {
		return fmt.Errorf("initialize gateway HTTP server: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Listen()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		logger.Info("received shutdown signal", slog.String("signal", sig.String()))
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()
		if sErr := srv.Shutdown(shutdownCtx); sErr != nil && !errors.Is(sErr, context.Canceled) {
			return fmt.Errorf("gateway shutdown: %w", sErr)
		}
		if listenErr := <-errCh; listenErr != nil && !errors.Is(listenErr, context.Canceled) {
			return fmt.Errorf("gateway server stop: %w", listenErr)
		}
		return nil
	case listenErr := <-errCh:
		if listenErr != nil {
			return fmt.Errorf("gateway server listen: %w", listenErr)
		}
		return nil
	}
}

type efPortalProvisionerBridge struct {
	ef *efadapter.Adapter
}

func (b efPortalProvisionerBridge) PreparePortalUserProvisioning(
	ctx context.Context,
	partyID string,
	contactID *string,
	requiredRole string,
) (string, *string, error) {
	return b.ef.PreparePortalUserProvisioning(ctx, partyID, contactID, requiredRole)
}

func (b efPortalProvisionerBridge) LinkPortalUser(
	ctx context.Context,
	req iamadapter.PartyUserLinkRequest,
	iamUserID string,
) (*model.PartyContact, error) {
	return b.ef.CompletePortalUserProvisioning(ctx, efadapter.PortalUserProvisioningInput{
		PartyID:          req.PartyID,
		ContactID:        req.ContactID,
		RequiredRole:     req.RequiredRole,
		FullName:         req.FullName,
		Email:            req.Email,
		ContactRole:      req.ContactRole,
		JobTitle:         req.JobTitle,
		Phone:            req.Phone,
		IsPrimaryContact: req.IsPrimaryContact,
		IAMUserID:        iamUserID,
	})
}
