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

//go:embed migrations/000002_hrm_requisition_provenance.up.sql
var requisitionProvenanceSQL string

//go:embed migrations/000003_hrm_job_posting_withdrawal.up.sql
var jobPostingWithdrawalSQL string

//go:embed migrations/000004_hrm_application_rejection.up.sql
var applicationRejectionSQL string

//go:embed migrations/000005_hrm_application_shortlisting.up.sql
var applicationShortlistingSQL string

//go:embed migrations/000006_hrm_requisition_rejection.up.sql
var requisitionRejectionSQL string

// Kunci transaksi advisory khusus HRM; lock dilepas otomatis saat commit/rollback.
const migrationLockID int64 = 718291825

type migration struct {
	version int64
	sql     string
}

var migrations = []migration{
	{version: 1, sql: baselineSQL},
	{version: 2, sql: requisitionProvenanceSQL},
	{version: 3, sql: jobPostingWithdrawalSQL},
	{version: 4, sql: applicationRejectionSQL},
	{version: 5, sql: applicationShortlistingSQL},
	{version: 6, sql: requisitionRejectionSQL},
}

// RunMigrationsUp memasang migrasi berurutan secara atomik dan idempoten.
// Checksum berbeda, versi asing, dan riwayat yang berlubang ditolak.
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
	rows, err := tx.QueryContext(ctx, "SELECT version, checksum FROM hrm_schema_migrations ORDER BY version")
	if err != nil {
		return migrationError(ctx, "inspect migration versions", err)
	}
	installed := make(map[int64]string)
	for rows.Next() {
		var version int64
		var checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			_ = rows.Close()
			return migrationError(ctx, "read migration tracker", err)
		}
		installed[version] = checksum
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return migrationError(ctx, "read migration tracker", err)
	}
	if err := rows.Close(); err != nil {
		return migrationError(ctx, "close migration tracker", err)
	}
	if len(installed) > len(migrations) {
		return fmt.Errorf("%w: unsupported HRM migration version", hrerrors.ErrConflict)
	}
	for index, item := range migrations {
		if item.version != int64(index+1) {
			return fmt.Errorf("%w: invalid HRM migration order", hrerrors.ErrConflict)
		}
		checksum, exists := installed[item.version]
		if !exists {
			if index < len(installed) {
				return fmt.Errorf("%w: missing HRM migration version", hrerrors.ErrConflict)
			}
			continue
		}
		digest := sha256.Sum256([]byte(item.sql))
		if checksum != hex.EncodeToString(digest[:]) {
			return fmt.Errorf("%w: HRM migration version or checksum mismatch", hrerrors.ErrConflict)
		}
	}
	if _, exists := installed[1]; !exists {
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
	}
	for _, item := range migrations {
		if _, exists := installed[item.version]; exists {
			continue
		}
		if _, err := tx.ExecContext(ctx, item.sql); err != nil {
			return migrationError(ctx, fmt.Sprintf("apply HRM migration %06d", item.version), err)
		}
		digest := sha256.Sum256([]byte(item.sql))
		if _, err := tx.ExecContext(ctx, "INSERT INTO hrm_schema_migrations (version, checksum) VALUES ($1, $2)", item.version, hex.EncodeToString(digest[:])); err != nil {
			return migrationError(ctx, fmt.Sprintf("record HRM migration %06d checksum", item.version), err)
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
