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

type migration struct {
	version int64
	sql     string
}

var migrations = []migration{
	{version: 1, sql: baselineSQL},
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
	if len(installed) == len(legacyMigrationChecksums) && matchesLegacyBaseline(installed) {
		if err := tx.Commit(); err != nil {
			return migrationError(ctx, "commit HRM legacy baseline verification", err)
		}
		return nil
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

// Konsolidasi hanya menerima tujuh checksum historis yang lengkap dan baseline yang sudah terbukti setara.
const consolidatedBaselineChecksum = "39aaa9acaf545cc9e8d8044e946dad21fcb0e1ca3555889741c192e53857486b"

var legacyMigrationChecksums = [...]string{
	"fb66aa622a64ad1005e24e874dcc41b49f30489a3a5f033812d3f4e6fdced42e",
	"04a8bdfb0a48f1124097a6d250c293147674d69ce6aeb20a4dd4947d43339789",
	"e7da65112aff0c3a18fb8fad4858b83a22cd398bdbf9881403ee9e5a422a48e7",
	"d42c71d208e226c054b045bbc5e5f686f2f7ef7fc631183c4849a31815653bc4",
	"4442e1c00aa87c7eaa54522ed5539e1d361ff5f0d545a49457c462c9c59d2808",
	"54f517cddd8cfd2679f27424e319fd4641d0d2cd2bd16d479a08dcba29035a6c",
	"404b9b6634a239bf69883c767b93636e7f5ddff73716bcd8357e03e58f033955",
}

func matchesLegacyBaseline(installed map[int64]string) bool {
	digest := sha256.Sum256([]byte(baselineSQL))
	if hex.EncodeToString(digest[:]) != consolidatedBaselineChecksum {
		return false
	}
	for index, checksum := range legacyMigrationChecksums {
		if installed[int64(index+1)] != checksum {
			return false
		}
	}
	return true
}
