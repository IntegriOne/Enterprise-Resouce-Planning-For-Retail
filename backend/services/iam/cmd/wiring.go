package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/iam/config"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/mailer"
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
		return nil, err
	}

	appEnv := ""
	if cfg != nil {
		appEnv = cfg.AppEnv
	}
	secCfg, err := config.LoadSecurity(appEnv)
	if err != nil {
		return nil, err
	}

	reg, err := services.NewRegistryFromConfig(repos, secCfg)
	if err != nil {
		return nil, err
	}
	if err := reg.Validate(); err != nil {
		return nil, err
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
) (*services.Registry, *repository.Registry, *redis.Client, error) {
	if cfg == nil {
		return nil, nil, nil, errors.New("iam config is required")
	}

	secCfg, err := config.LoadSecurity(cfg.AppEnv)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := cfg.ValidateRuntimeStrict(secCfg); err != nil {
		return nil, nil, nil, err
	}

	repos, err := repository.New(gormDB)
	if err != nil {
		return nil, nil, nil, err
	}

	var redisClient *redis.Client
	if cfg.RedisEnabled {
		client, err := authsession.NewTLSClient(cfg.Redis)
		if err != nil {
			return nil, nil, nil, err
		}
		if pingErr := client.Ping(ctx).Err(); pingErr != nil {
			_ = client.Close()
			return nil, nil, nil, pingErr
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
		if redisClient != nil {
			_ = redisClient.Close()
		}
		return nil, nil, nil, err
	}

	secOpts := services.BuildSecurityOptionsFromConfig(secCfg, redisClient, cfg.RedisEnv, logger)
	secOpts.EmailSender = emailSender
	secOpts.PublicBaseURL = cfg.Email.PublicBaseURL

	reg, err := services.NewRegistry(repos, secOpts)
	if err != nil {
		_ = emailSender.Close(ctx)
		if redisClient != nil {
			_ = redisClient.Close()
		}
		return nil, nil, nil, err
	}
	if err := reg.Validate(); err != nil {
		_ = emailSender.Close(ctx)
		if redisClient != nil {
			_ = redisClient.Close()
		}
		return nil, nil, nil, err
	}

	if logger != nil {
		logger.Info("iam runtime stack wired",
			slog.Int("totp_active_key_version", int(secCfg.TOTPActiveKeyVersion)),
			slog.Int("csrf_active_key_version", secCfg.CSRFActiveKeyVersion),
			slog.Bool("csrf_keyring_configured", secCfg.CSRFKeyringConfigured),
			slog.Bool("redis_enabled", cfg.RedisEnabled),
			slog.String("email_provider", string(cfg.Email.Provider)),
			slog.String("email_delivery_mode", string(cfg.Email.DeliveryMode)),
		)
	}
	return reg, repos, redisClient, nil
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
		return nil, err
	}
	return srv, nil
}
