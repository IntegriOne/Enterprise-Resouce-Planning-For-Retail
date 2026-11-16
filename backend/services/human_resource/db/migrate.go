package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

//go:embed migrations/000001_hrm.up.sql
var baselineSQL string

// Kunci transaksi advisory khusus HRM; lock dilepas otomatis saat commit/rollback.
const migrationLockID int64 = 718291825

// RunMigrationsUp memasang baseline tunggal 000001 secara atomik dan idempoten.
// Checksum berbeda atau versi asing ditolak; tidak ada AutoMigrate maupun reset otomatis.
func RunMigrationsUp(ctx context.Context, pool *sql.DB) (resultErr error) {
	if ctx == nil {
		return hrerrors.ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if pool == nil {
		return hrerrors.ErrUnavailable
	}
	tx, err := pool.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return migrationError(ctx, "begin migration transaction", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			resultErr = errors.Join(resultErr, migrationError(ctx, "rollback migration transaction", err))
		}
	}()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", migrationLockID); err != nil {
		return migrationError(ctx, "acquire migration lock", err)
	}
	const tracker = `CREATE TABLE IF NOT EXISTS hrm_schema_migrations (
		version BIGINT PRIMARY KEY,
		checksum CHAR(64) NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`
	if _, err := tx.ExecContext(ctx, tracker); err != nil {
		return migrationError(ctx, "initialize migration tracker", err)
	}
	digest := sha256.Sum256([]byte(baselineSQL))
	checksum := hex.EncodeToString(digest[:])
	var unknownVersions int64
	if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM hrm_schema_migrations WHERE version <> 1").Scan(&unknownVersions); err != nil {
		return migrationError(ctx, "inspect migration versions", err)
	}
	if unknownVersions != 0 {
		return fmt.Errorf("%w: unsupported HRM migration version", hrerrors.ErrConflict)
	}
	var installedVersion int64
	var installedChecksum string
	err = tx.QueryRowContext(ctx, "SELECT version, checksum FROM hrm_schema_migrations WHERE version = 1").Scan(&installedVersion, &installedChecksum)
	if err == nil {
		if installedVersion != 1 || installedChecksum != checksum {
			return fmt.Errorf("%w: HRM migration version or checksum mismatch", hrerrors.ErrConflict)
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return migrationError(ctx, "read migration tracker", err)
	} else {
		var existingSchema bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = current_schema() AND left(table_name, 3) = 'hr_'
		)`).Scan(&existingSchema); err != nil {
			return migrationError(ctx, "inspect existing HRM schema", err)
		}
		if existingSchema {
			return fmt.Errorf("%w: existing HRM schema has no verified migration checksum", hrerrors.ErrConflict)
		}
		if _, err := tx.ExecContext(ctx, baselineSQL); err != nil {
			return migrationError(ctx, "apply baseline 000001", err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO hrm_schema_migrations (version, checksum) VALUES ($1, $2)", 1, checksum); err != nil {
			return migrationError(ctx, "record baseline checksum", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return migrationError(ctx, "commit migration transaction", err)
	}
	return nil
}

// migrationError menjaga context cancellation tanpa membocorkan SQL/credential dari driver.
func migrationError(ctx context.Context, stage string, _ error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return fmt.Errorf("%w: %s", hrerrors.ErrUnavailable, stage)
}
