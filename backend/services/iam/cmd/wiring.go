package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/mailer"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/scopeverifier"
	"github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services"
	iamgrpc "github.com/integrione/erp-retail/backend/services/iam/transport/grpc"
	"github.com/integrione/erp-retail/backend/services/iam/transport/grpc/handler"
	"github.com/integrione/erp-retail/backend/shared/authsession"
)

// wireIAMServices merangkai repository.Registry berbasis GORM ke seluruh layanan domain IAM secara fail-closed.
func wireIAMServices(gormDB *gorm.DB, cfg *config.Config, logger *slog.Logger) (*services.Registry, error) {
	repos, err := repository.New(gormDB)
	if err != nil {
		return nil, fmt.Errorf("wire iam repositories: %w", err)
	}

	appEnv := ""
	if cfg != nil {
		appEnv = cfg.AppEnv
	}
	secCfg, err := config.LoadSecurity(appEnv)
	if err != nil {
		return nil, fmt.Errorf("load iam security config: %w", err)
	}

	reg, err := services.NewRegistryFromConfig(repos, secCfg)
	if err != nil {
		return nil, fmt.Errorf("build iam service registry: %w", err)
	}
	if err := reg.Validate(); err != nil {
		return nil, fmt.Errorf("validate iam service registry: %w", err)
	}

	if logger != nil {
		logger.Info("iam application services wired",
			slog.Int("totp_active_key_version", int(secCfg.TOTPActiveKeyVersion)),
			slog.Bool("totp_ephemeral_dev_key", secCfg.AllowEphemeralDevKey),
			slog.Int("max_failed_login_attempts", secCfg.MaxFailedAttempts),
		)
	}
	return reg, nil
}

// wireRuntimeStack merangkai repositori, client Redis TLS+ACL (bila diaktifkan), dan layanan IAM daemon secara fail-closed.
func wireRuntimeStack(
	ctx context.Context,
	gormDB *gorm.DB,
	cfg *config.Config,
	logger *slog.Logger,
) (*services.Registry, *repository.Registry, *redis.Client, io.Closer, error) {
	if cfg == nil {
		return nil, nil, nil, nil, errors.New("iam config is required")
	}

	secCfg, err := config.LoadSecurity(cfg.AppEnv)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("load iam security config: %w", err)
	}
	if err := cfg.ValidateRuntimeStrict(secCfg); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("validate runtime config: %w", err)
	}

	repos, err := repository.New(gormDB)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("init iam repositories: %w", err)
	}

	var redisClient *redis.Client
	if cfg.RedisEnabled {
		client, err := authsession.NewTLSClient(cfg.Redis)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("init iam redis tls client: %w", err)
		}
		if pingErr := client.Ping(ctx).Err(); pingErr != nil {
			return nil, nil, nil, nil, errors.Join(
				fmt.Errorf("ping iam redis: %w", pingErr),
				wrapOptionalClose("close iam redis client", client.Close()),
			)
		}
		redisClient = client
		if logger != nil {
			logger.Info("iam redis tls+acl client connected",
				slog.String("addr", cfg.Redis.Addr),
				slog.String("username", cfg.Redis.Username),
				slog.String("redis_env", cfg.RedisEnv),
			)
		}
	}

	emailSender, err := mailer.NewSender(ctx, cfg.Email, cfg.IsProduction(), logger)
	if err != nil {
		var closeRedisErr error
		if redisClient != nil {
			closeRedisErr = wrapOptionalClose("close iam redis client", redisClient.Close())
		}
		return nil, nil, nil, nil, errors.Join(fmt.Errorf("init iam email sender: %w", err), closeRedisErr)
	}

	var scopeVerifierCloser io.Closer
	var scopeVerifier *scopeverifier.EFVerifier
	if cfg.EFGRPCEnabled {
		scopeVerifier, err = scopeverifier.NewEFVerifier(scopeverifier.Config{
			Target:     cfg.EFGRPCTarget,
			Timeout:    cfg.EFGRPCTimeout,
			HMACSecret: []byte(strings.TrimSpace(cfg.InternalGatewayHMACSecret)),
		})
		if err != nil {
			var closeRedisErr error
			if redisClient != nil {
				closeRedisErr = wrapOptionalClose("close iam redis client", redisClient.Close())
			}
			return nil, nil, nil, nil, errors.Join(
				fmt.Errorf("init enterprise foundation scope verifier: %w", err),
				wrapOptionalClose("close iam email sender", emailSender.Close(ctx)),
				closeRedisErr,
			)
		}
		scopeVerifierCloser = scopeVerifier
	}

	secOpts := services.BuildSecurityOptionsFromConfig(secCfg, redisClient, cfg.RedisEnv, logger)
	secOpts.EmailSender = emailSender
	secOpts.PublicBaseURL = cfg.Email.PublicBaseURL
	if scopeVerifier != nil {
		secOpts.ScopeVerifier = scopeVerifier
	}

	reg, err := services.NewRegistry(repos, secOpts)
	if err != nil {
		var closeRedisErr error
		if redisClient != nil {
			closeRedisErr = wrapOptionalClose("close iam redis client", redisClient.Close())
		}
		var closeScopeErr error
		if scopeVerifierCloser != nil {
			closeScopeErr = wrapOptionalClose("close enterprise foundation scope verifier", scopeVerifierCloser.Close())
		}
		return nil, nil, nil, nil, errors.Join(
			fmt.Errorf("build iam service registry: %w", err),
			wrapOptionalClose("close iam email sender", emailSender.Close(ctx)),
			closeRedisErr,
			closeScopeErr,
		)
	}
	if err := reg.Validate(); err != nil {
		var closeRedisErr error
		if redisClient != nil {
			closeRedisErr = wrapOptionalClose("close iam redis client", redisClient.Close())
		}
		var closeScopeErr error
		if scopeVerifierCloser != nil {
			closeScopeErr = wrapOptionalClose("close enterprise foundation scope verifier", scopeVerifierCloser.Close())
		}
		return nil, nil, nil, nil, errors.Join(
			fmt.Errorf("validate iam service registry: %w", err),
			wrapOptionalClose("close iam email sender", emailSender.Close(ctx)),
			closeRedisErr,
			closeScopeErr,
		)
	}

	if logger != nil {
		logger.Info("iam runtime stack wired",
			slog.Int("totp_active_key_version", int(secCfg.TOTPActiveKeyVersion)),
			slog.Int("csrf_active_key_version", secCfg.CSRFActiveKeyVersion),
			slog.Bool("csrf_keyring_configured", secCfg.CSRFKeyringConfigured),
			slog.Bool("redis_enabled", cfg.RedisEnabled),
			slog.Bool("ef_scope_verifier_enabled", cfg.EFGRPCEnabled),
			slog.String("email_provider", string(cfg.Email.Provider)),
			slog.String("email_delivery_mode", string(cfg.Email.DeliveryMode)),
		)
	}
	return reg, repos, redisClient, scopeVerifierCloser, nil
}

// wireGRPCServer membangun gRPC server IAM sekaligus mendaftarkan seluruh handler domain iam.v1 secara fail-closed.
func wireGRPCServer(cfg *config.Config, reg *services.Registry, logger *slog.Logger) (*iamgrpc.Server, error) {
	if cfg == nil {
		return nil, errors.New("iam grpc config is required")
	}
	if reg == nil {
		return nil, errors.New("iam service registry is required")
	}

	srv := iamgrpc.NewServer(cfg, logger)
	if err := handler.RegisterAll(srv.GRPCServer(), reg, logger); err != nil {
		return nil, fmt.Errorf("register iam grpc handlers: %w", err)
	}
	return srv, nil
}

func wrapOptionalClose(action string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", action, err)
}
