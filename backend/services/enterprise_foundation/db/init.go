package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	defaultMaxIdleConns    = 10
	defaultMaxOpenConns    = 50
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
)

// Init memastikan basis data target tersedia, membuka koneksi GORM terkonfigurasi pool,
// serta menjalankan seluruh migrasi skema up secara otomatis dan idempoten.
// Apabila isTesting bernilai true, nama database pada DSN wajib berakhiran '_test'.
func Init(ctx context.Context, dsn string, isTesting bool) (*gorm.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("database DSN must not be empty")
	}

	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database DSN: %w", err)
	}

	if isTesting && !strings.HasSuffix(cfg.Database, "_test") {
		return nil, fmt.Errorf("test execution requires a database name ending with '_test', got %q", cfg.Database)
	}

	if err := EnsureDatabase(ctx, dsn); err != nil {
		return nil, fmt.Errorf("ensure database existence: %w", err)
	}

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time {
			return time.Now()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open target database connection: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("obtain sql.DB handle: %w", err)
	}

	sqlDB.SetMaxIdleConns(defaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(defaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(defaultConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(defaultConnMaxIdleTime)

	initialized := false
	defer func() {
		if !initialized {
			_ = sqlDB.Close()
		}
	}()

	if err := RunMigrationsUp(ctx, sqlDB); err != nil {
		return nil, fmt.Errorf("apply database migrations: %w", err)
	}

	initialized = true
	return gormDB, nil
}

// InitDevAndTest menginisialisasi sekaligus basis data development (`enterprise_foundation`)
// dan basis data pengujian (`enterprise_foundation_test`) dengan skema migrasi identik.
// Jika testDSN kosong, nama basis data pengujian diturunkan otomatis menjadi '<devDB>_test'.
func InitDevAndTest(ctx context.Context, devDSN, testDSN string) (*gorm.DB, *gorm.DB, error) {
	devDB, err := Init(ctx, devDSN, false)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize development database: %w", err)
	}

	resolvedTestDSN := strings.TrimSpace(testDSN)
	if resolvedTestDSN == "" {
		resolvedTestDSN, err = DeriveTestDatabaseDSN(devDSN)
		if err != nil {
			closeGORMQuietly(devDB)
			return nil, nil, fmt.Errorf("derive test database DSN: %w", err)
		}
	}

	testDB, err := Init(ctx, resolvedTestDSN, true)
	if err != nil {
		closeGORMQuietly(devDB)
		return nil, nil, fmt.Errorf("initialize test database: %w", err)
	}

	return devDB, testDB, nil
}

func closeGORMQuietly(gormDB *gorm.DB) {
	if gormDB == nil {
		return
	}
	if sqlDB, err := gormDB.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
