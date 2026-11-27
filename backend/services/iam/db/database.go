// Package db menyediakan inisialisasi basis data PostgreSQL dan migrasi skema otomatis untuk layanan IAM.
package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

var validDBName = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// EnsureDatabase memeriksa keberadaan basis data target pada PostgreSQL.
// Jika belum ada, fungsi membuka koneksi sementara ke basis data maintenance 'postgres'
// dan membuat basis data baru dengan encoding UTF8 secara aman.
func EnsureDatabase(ctx context.Context, dsn string) (retErr error) {
	if strings.TrimSpace(dsn) == "" {
		return fmt.Errorf("database DSN cannot be empty")
	}

	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return sanitizeDSNParseError("failed to parse database DSN", err)
	}

	targetDB := cfg.Database
	if targetDB == "" {
		return fmt.Errorf("target database name in DSN cannot be empty")
	}

	if !validDBName.MatchString(targetDB) {
		return fmt.Errorf("target database name %q contains invalid characters", targetDB)
	}

	// Jangan melakukan auto-create pada basis data sistem bawaan PostgreSQL.
	if targetDB == "postgres" || targetDB == "template0" || targetDB == "template1" {
		return nil
	}

	maintCfg := cfg.Copy()
	maintCfg.Database = "postgres"

	maintDB := stdlib.OpenDB(*maintCfg)
	defer func() {
		if closeErr := maintDB.Close(); closeErr != nil {
			retErr = errors.Join(retErr, fmt.Errorf("failed to close maintenance database handle: %w", closeErr))
		}
	}()

	if err := maintDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to connect to maintenance database 'postgres': %w", err)
	}

	var exists bool
	const checkQuery = `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1);`
	if err := maintDB.QueryRowContext(ctx, checkQuery, targetDB).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check existence of database %q: %w", targetDB, err)
	}

	if !exists {
		slog.Info("creating target database", slog.String("database", targetDB))
		createQuery := fmt.Sprintf(`CREATE DATABASE "%s" ENCODING 'UTF8';`, targetDB)
		if _, err := maintDB.ExecContext(ctx, createQuery); err != nil {
			return fmt.Errorf("failed to create database %q: %w", targetDB, err)
		}
		slog.Info("target database created successfully", slog.String("database", targetDB))
	}

	return nil
}

// DeriveTestDatabaseDSN menurunkan DSN basis data pengujian berakhiran '_test'
// dari DSN basis data development apabila DSN pengujian belum disediakan secara eksplisit.
func DeriveTestDatabaseDSN(devDSN string) (string, error) {
	if strings.TrimSpace(devDSN) == "" {
		return "", fmt.Errorf("development DSN cannot be empty")
	}

	cfg, err := pgx.ParseConfig(devDSN)
	if err != nil {
		return "", sanitizeDSNParseError("failed to parse development DSN", err)
	}

	baseDB := cfg.Database
	if baseDB == "" {
		return "", fmt.Errorf("database name in development DSN cannot be empty")
	}

	testDB := baseDB
	if !strings.HasSuffix(testDB, "_test") {
		testDB = baseDB + "_test"
	}

	if !validDBName.MatchString(testDB) {
		return "", fmt.Errorf("derived test database name %q contains invalid characters", testDB)
	}

	// Ubah path database jika DSN memakai format URI (postgres:// atau postgresql://).
	if strings.HasPrefix(devDSN, "postgres://") || strings.HasPrefix(devDSN, "postgresql://") {
		parsedURL, parseErr := url.Parse(devDSN)
		if parseErr != nil {
			return "", sanitizeDSNParseError("failed to parse DSN URL", parseErr)
		}
		parsedURL.Path = "/" + testDB
		return parsedURL.String(), nil
	}

	// Fallback untuk DSN berformat keyword=value.
	if strings.Contains(devDSN, "dbname=") {
		re := regexp.MustCompile(`dbname=[^\s]+`)
		return re.ReplaceAllString(devDSN, "dbname="+testDB), nil
	}

	return devDSN + " dbname=" + testDB, nil
}

// sanitizeDSNParseError menghapus detail *url.Error atau string DSN mentah agar kata sandi database tidak bocor ke log.
func sanitizeDSNParseError(prefix string, err error) error {
	if urlErr, ok := errors.AsType[*url.Error](err); ok && urlErr.Err != nil {
		return fmt.Errorf("%s: %w", prefix, urlErr.Err)
	}
	return fmt.Errorf("%s: invalid connection string format", prefix)
}
