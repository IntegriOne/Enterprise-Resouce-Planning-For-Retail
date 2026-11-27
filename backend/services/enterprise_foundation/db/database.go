// Package db menyediakan inisialisasi basis data PostgreSQL, pengaturan pool,
// serta eksekusi migrasi skema untuk layanan Enterprise Foundation.
package db

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

var (
	validDBName  = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	keywordDBKey = regexp.MustCompile(`dbname=[^\s]+`)
)

// EnsureDatabase memeriksa keberadaan basis data target di server PostgreSQL.
// Bila belum tersedia, fungsi membuka koneksi ke basis data maintenance 'postgres'
// dan membuat basis data baru berencoding UTF8 secara aman.
func EnsureDatabase(ctx context.Context, dsn string) error {
	if strings.TrimSpace(dsn) == "" {
		return fmt.Errorf("database DSN must not be empty")
	}

	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse database DSN: %w", err)
	}

	targetDB := cfg.Database
	if targetDB == "" {
		return fmt.Errorf("target database name in DSN must not be empty")
	}
	if !validDBName.MatchString(targetDB) {
		return fmt.Errorf("target database name %q contains invalid characters", targetDB)
	}

	// Jangan pernah melakukan auto-create pada basis data sistem bawaan PostgreSQL.
	if targetDB == "postgres" || targetDB == "template0" || targetDB == "template1" {
		return nil
	}

	maintCfg := cfg.Copy()
	maintCfg.Database = "postgres"

	maintDB := stdlib.OpenDB(*maintCfg)
	defer func() { _ = maintDB.Close() }()

	if err := maintDB.PingContext(ctx); err != nil {
		return fmt.Errorf("connect to maintenance database postgres: %w", err)
	}

	var exists bool
	const checkQuery = `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1);`
	if err := maintDB.QueryRowContext(ctx, checkQuery, targetDB).Scan(&exists); err != nil {
		return fmt.Errorf("check database existence %q: %w", targetDB, err)
	}

	if !exists {
		slog.Info("creating Enterprise Foundation target database", slog.String("database", targetDB))
		createQuery := fmt.Sprintf(`CREATE DATABASE "%s" ENCODING 'UTF8';`, targetDB)
		if _, err := maintDB.ExecContext(ctx, createQuery); err != nil {
			return fmt.Errorf("create database %q: %w", targetDB, err)
		}
		slog.Info("created Enterprise Foundation target database", slog.String("database", targetDB))
	}

	return nil
}

// DeriveTestDatabaseDSN menurunkan DSN basis data pengujian berakhiran '_test'
// dari DSN basis data development bila DSN pengujian belum disediakan terpisah.
func DeriveTestDatabaseDSN(devDSN string) (string, error) {
	if strings.TrimSpace(devDSN) == "" {
		return "", fmt.Errorf("development DSN must not be empty")
	}

	cfg, err := pgx.ParseConfig(devDSN)
	if err != nil {
		return "", fmt.Errorf("parse development DSN: %w", err)
	}

	baseDB := cfg.Database
	if baseDB == "" {
		return "", fmt.Errorf("database name in development DSN must not be empty")
	}

	testDB := baseDB
	if !strings.HasSuffix(testDB, "_test") {
		testDB = baseDB + "_test"
	}

	if !validDBName.MatchString(testDB) {
		return "", fmt.Errorf("derived test database name %q contains invalid characters", testDB)
	}

	if strings.HasPrefix(devDSN, "postgres://") || strings.HasPrefix(devDSN, "postgresql://") {
		parsedURL, parseErr := url.Parse(devDSN)
		if parseErr != nil {
			return "", fmt.Errorf("parse DSN URL: %w", parseErr)
		}
		parsedURL.Path = "/" + testDB
		return parsedURL.String(), nil
	}

	if strings.Contains(devDSN, "dbname=") {
		return keywordDBKey.ReplaceAllString(devDSN, "dbname="+testDB), nil
	}

	return devDSN + " dbname=" + testDB, nil
}
