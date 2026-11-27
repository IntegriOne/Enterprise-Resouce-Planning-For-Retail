//go:build integration

package db

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRunMigrationsUpRejectsInvalidHistory(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	adminDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal("open test database failed")
	}
	admin, err := adminDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := fmt.Sprintf("hrm_migration_test_%d", time.Now().UnixNano())
	if _, err := admin.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if _, err := admin.ExecContext(cleanupCtx, `DROP SCHEMA "`+schema+`" CASCADE`); err != nil {
			t.Errorf("drop migration test schema: %v", err)
		}
	}()
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	isolatedDB, err := gorm.Open(postgres.Open(parsed.String()), &gorm.Config{})
	if err != nil {
		t.Fatal("open isolated test schema failed")
	}
	pool, err := isolatedDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := RunMigrationsUp(ctx, pool); err != nil {
		t.Fatalf("install HRM schema: %v", err)
	}
	if err := RunMigrationsUp(ctx, pool); err != nil {
		t.Fatalf("repeat HRM schema: %v", err)
	}
	var installed int
	if err := pool.QueryRowContext(ctx, "SELECT count(*) FROM hrm_schema_migrations").Scan(&installed); err != nil || installed != 1 {
		t.Fatalf("installed migration count: got %d, error %v", installed, err)
	}
	var columns int
	if err := pool.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'hr_worker_assignments'
		AND column_name IN ('requested_by_principal_id', 'requested_at')`).Scan(&columns); err != nil || columns != 2 {
		t.Fatalf("assignment provenance columns: got %d, error %v", columns, err)
	}
	if _, err := pool.ExecContext(ctx, "INSERT INTO hrm_schema_migrations (version, checksum) VALUES (8, repeat('0', 64))"); err != nil {
		t.Fatal(err)
	}
	if err := RunMigrationsUp(ctx, pool); !errors.Is(err, hrerrors.ErrConflict) {
		t.Fatalf("unknown version: got %v, want conflict", err)
	}
	if _, err := pool.ExecContext(ctx, "DELETE FROM hrm_schema_migrations WHERE version = 8"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.ExecContext(ctx, "UPDATE hrm_schema_migrations SET checksum = repeat('0', 64) WHERE version = 1"); err != nil {
		t.Fatal(err)
	}
	if err := RunMigrationsUp(ctx, pool); !errors.Is(err, hrerrors.ErrConflict) {
		t.Fatalf("checksum mismatch: got %v, want conflict", err)
	}
	if _, err := pool.ExecContext(ctx, "DELETE FROM hrm_schema_migrations"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.ExecContext(ctx, "INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES ('558c401c-beda-458b-b6d6-33eacd335583', 'MIGRATION', 'Fixture', 'test')"); err != nil {
		t.Fatal(err)
	}
	for i, checksum := range legacyMigrationChecksums {
		if _, err := pool.ExecContext(ctx, "INSERT INTO hrm_schema_migrations (version, checksum) VALUES ($1, $2)", i+1, checksum); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		if err := RunMigrationsUp(ctx, pool); err != nil {
			t.Fatalf("accept equivalent legacy schema: %v", err)
		}
	}
	if err := pool.QueryRowContext(ctx, "SELECT count(*) FROM hrm_schema_migrations").Scan(&installed); err != nil || installed != len(legacyMigrationChecksums) {
		t.Fatalf("legacy migration count: got %d, error %v", installed, err)
	}
	var familyName string
	if err := pool.QueryRowContext(ctx, "SELECT family_name FROM hr_job_families WHERE family_code = 'MIGRATION'").Scan(&familyName); err != nil || familyName != "Fixture" {
		t.Fatalf("business data after verification: got %q, error %v", familyName, err)
	}
	if _, err := pool.ExecContext(ctx, "UPDATE hrm_schema_migrations SET checksum = repeat('0', 64) WHERE version = 3"); err != nil {
		t.Fatal(err)
	}
	if err := RunMigrationsUp(ctx, pool); !errors.Is(err, hrerrors.ErrConflict) {
		t.Fatalf("legacy checksum mismatch: got %v, want conflict", err)
	}

}
