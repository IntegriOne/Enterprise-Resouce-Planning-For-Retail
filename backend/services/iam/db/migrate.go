package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"
)

const iamMigrationLockID int64 = 718291823

//go:embed migrations/*.up.sql migrations/*.down.sql
var MigrationFiles embed.FS

type migrationItem struct {
	version  int64
	filename string
}

// RunMigrationsUp menerapkan seluruh berkas migrasi '*.up.sql' yang belum terpasang
// ke dalam basis data target secara berurutan dan transaksional.
func RunMigrationsUp(ctx context.Context, db *sql.DB) (resultErr error) {
	if db == nil {
		return fmt.Errorf("migration database handle is nil")
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire migration connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Ambil PostgreSQL advisory lock agar eksekusi migrasi paralel tetap aman dan bebas race condition.
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1);", iamMigrationLockID); err != nil {
		return fmt.Errorf("failed to acquire migration advisory lock: %w", err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var unlocked bool
		unlockErr := conn.QueryRowContext(unlockCtx, "SELECT pg_advisory_unlock($1);", iamMigrationLockID).Scan(&unlocked)
		if unlockErr == nil && !unlocked {
			unlockErr = fmt.Errorf("migration advisory lock was not held")
		}
		if unlockErr != nil {
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			resultErr = errors.Join(resultErr, fmt.Errorf("failed to release migration advisory lock: %w", unlockErr))
		}
	}()

	const createTrackerQuery = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`
	if _, err := conn.ExecContext(ctx, createTrackerQuery); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	applied, err := readAppliedMigrations(ctx, conn)
	if err != nil {
		return err
	}

	items, err := collectMigrationFiles(".up.sql", true)
	if err != nil {
		return err
	}

	for _, item := range items {
		if applied[item.version] {
			continue
		}

		slog.Info("applying database migration", slog.Int64("version", item.version), slog.String("file", item.filename))
		content, err := MigrationFiles.ReadFile("migrations/" + item.filename)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", item.filename, err)
		}

		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start migration transaction %d: %w", item.version, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			_, _ = conn.ExecContext(ctx,
				`INSERT INTO schema_migrations (version, dirty, applied_at) VALUES ($1, true, NOW()) ON CONFLICT (version) DO UPDATE SET dirty = true;`,
				item.version,
			)
			return fmt.Errorf("failed to execute migration %s: %w", item.filename, err)
		}

		const recordQuery = `INSERT INTO schema_migrations (version, dirty, applied_at) VALUES ($1, false, NOW()) ON CONFLICT (version) DO UPDATE SET dirty = false, applied_at = NOW();`
		if _, err := tx.ExecContext(ctx, recordQuery, item.version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration version %d: %w", item.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration version %d: %w", item.version, err)
		}

		slog.Info("database migration applied", slog.Int64("version", item.version))
	}

	return nil
}

// RunMigrationsDown membatalkan migrasi yang telah terpasang secara terbalik.
// Fungsi ini hanya ditujukan untuk kebutuhan reset database development atau pengujian.
func RunMigrationsDown(ctx context.Context, db *sql.DB) (resultErr error) {
	if db == nil {
		return fmt.Errorf("migration database handle is nil")
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire migration connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1);", iamMigrationLockID); err != nil {
		return fmt.Errorf("failed to acquire migration advisory lock: %w", err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var unlocked bool
		unlockErr := conn.QueryRowContext(unlockCtx, "SELECT pg_advisory_unlock($1);", iamMigrationLockID).Scan(&unlocked)
		if unlockErr == nil && !unlocked {
			unlockErr = fmt.Errorf("migration advisory lock was not held")
		}
		if unlockErr != nil {
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			resultErr = errors.Join(resultErr, fmt.Errorf("failed to release migration advisory lock: %w", unlockErr))
		}
	}()

	var exists bool
	if err := conn.QueryRowContext(ctx, `SELECT to_regclass('schema_migrations') IS NOT NULL;`).Scan(&exists); err != nil {
		return fmt.Errorf("failed to inspect schema_migrations: %w", err)
	}
	if !exists {
		return nil
	}

	applied, err := readAppliedMigrations(ctx, conn)
	if err != nil {
		return err
	}

	items, err := collectMigrationFiles(".down.sql", false)
	if err != nil {
		return err
	}

	for _, item := range items {
		if !applied[item.version] {
			continue
		}

		slog.Info("rolling back database migration", slog.Int64("version", item.version), slog.String("file", item.filename))
		content, err := MigrationFiles.ReadFile("migrations/" + item.filename)
		if err != nil {
			return fmt.Errorf("failed to read down migration %s: %w", item.filename, err)
		}

		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to start rollback transaction %d: %w", item.version, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute down migration %s: %w", item.filename, err)
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = $1;`, item.version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to remove migration record %d: %w", item.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit rollback %d: %w", item.version, err)
		}

		slog.Info("database migration rolled back", slog.Int64("version", item.version))
	}

	return nil
}

func readAppliedMigrations(ctx context.Context, conn *sql.Conn) (map[int64]bool, error) {
	rows, err := conn.QueryContext(ctx, `SELECT version, dirty FROM schema_migrations;`)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[int64]bool)
	for rows.Next() {
		var ver int64
		var dirty bool
		if err := rows.Scan(&ver, &dirty); err != nil {
			return nil, fmt.Errorf("failed to scan schema_migrations row: %w", err)
		}
		if dirty {
			return nil, fmt.Errorf("database schema is dirty at version %d; manual intervention required", ver)
		}
		applied[ver] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schema_migrations rows: %w", err)
	}
	return applied, nil
}

func collectMigrationFiles(suffix string, ascending bool) ([]migrationItem, error) {
	entries, err := MigrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded migrations directory: %w", err)
	}

	var items []migrationItem
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}
		ver, parseErr := strconv.ParseInt(parts[0], 10, 64)
		if parseErr != nil {
			continue
		}
		items = append(items, migrationItem{
			version:  ver,
			filename: entry.Name(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if ascending {
			return items[i].version < items[j].version
		}
		return items[i].version > items[j].version
	})

	return items, nil
}
