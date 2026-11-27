package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Init menginisialisasi basis data target, memastikan database fisik tersedia,
// menerapkan seluruh migrasi up, dan mengembalikan instance *gorm.DB terkonfigurasi.
// Jika isTesting bernilai true, nama database pada DSN wajib berakhiran '_test'.
func Init(ctx context.Context, dsn string, isTesting bool) (_ *gorm.DB, retErr error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("database DSN cannot be empty")
	}

	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, sanitizeDSNParseError("failed to parse database DSN", err)
	}

	if isTesting && !strings.HasSuffix(cfg.Database, "_test") {
		return nil, fmt.Errorf("test execution requires a database name ending with '_test', got %q", cfg.Database)
	}

	if err := EnsureDatabase(ctx, dsn); err != nil {
		return nil, fmt.Errorf("failed to ensure database existence: %w", err)
	}

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time {
			return time.Now()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open target database connection: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to obtain sql.DB handle: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	initialized := false
	defer func() {
		if !initialized {
			if closeErr := sqlDB.Close(); closeErr != nil {
				retErr = errors.Join(retErr, fmt.Errorf("failed to close database handle after init failure: %w", closeErr))
			}
		}
	}()

	if err := RunMigrationsUp(ctx, sqlDB); err != nil {
		return nil, fmt.Errorf("failed to apply database migrations: %w", err)
	}

	initialized = true
	return gormDB, nil
}

// InitDevAndTest menginisialisasi sekaligus basis data development (iam) dan basis data
// pengujian (iam_test) dengan skema migrasi yang identik.
// Jika testDSN kosong, nama database testing diturunkan otomatis menjadi '<devDB>_test'.
func InitDevAndTest(ctx context.Context, devDSN, testDSN string) (*gorm.DB, *gorm.DB, error) {
	devDB, err := Init(ctx, devDSN, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize development database: %w", err)
	}

	closeDev := func() error {
		sqlDB, dbErr := devDB.DB()
		if dbErr != nil {
			return fmt.Errorf("failed to acquire development sql.DB handle for cleanup: %w", dbErr)
		}
		if closeErr := sqlDB.Close(); closeErr != nil {
			return fmt.Errorf("failed to close development database after test init failure: %w", closeErr)
		}
		return nil
	}

	resolvedTestDSN := strings.TrimSpace(testDSN)
	if resolvedTestDSN == "" {
		resolvedTestDSN, err = DeriveTestDatabaseDSN(devDSN)
		if err != nil {
			return nil, nil, errors.Join(fmt.Errorf("failed to derive test database DSN: %w", err), closeDev())
		}
	}

	testDB, err := Init(ctx, resolvedTestDSN, true)
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("failed to initialize test database: %w", err), closeDev())
	}

	return devDB, testDB, nil
}
