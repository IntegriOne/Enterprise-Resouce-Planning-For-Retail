// Package db membuka koneksi PostgreSQL HRM dan menerapkan baseline berversi.
package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/integrione/erp-retail/backend/services/human_resource/config"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// Init membuka database yang sudah diprovisikan; tidak membuat database management lain.
// Gagal inisialisasi selalu menutup pool dan tidak mengekspos DSN pada error.
func Init(ctx context.Context, cfg config.Config) (*gorm.DB, error) {
	if ctx == nil {
		return nil, hrerrors.ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	handle, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), DisableAutomaticPing: true,
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("%w: cannot initialize HRM database driver", hrerrors.ErrUnavailable)
	}
	pool, err := handle.DB()
	if err != nil {
		return nil, fmt.Errorf("%w: cannot obtain HRM database pool", hrerrors.ErrUnavailable)
	}
	pool.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	pool.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	pool.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	pool.SetConnMaxIdleTime(5 * time.Minute)
	initialized := false
	defer func() {
		if !initialized {
			_ = pool.Close()
		}
	}()
	if err := pool.PingContext(ctx); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("%w: cannot connect to HRM database", hrerrors.ErrUnavailable)
	}
	if err := RunMigrationsUp(ctx, pool); err != nil {
		return nil, err
	}
	initialized = true
	return handle, nil
}
