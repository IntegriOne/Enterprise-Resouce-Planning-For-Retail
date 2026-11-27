//go:build integration

package db_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	payrollrepo "github.com/integrione/erp-retail/backend/services/human_resource/repository/payroll"
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
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatalf("install HRM schema: %v", err)
	}
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatalf("repeat HRM schema: %v", err)
	}
	if _, err := pool.ExecContext(ctx, `DROP TABLE hr_payroll_input_snapshots;
		UPDATE hrm_schema_migrations SET checksum = 'c0ad8ee18cb67fb2fd29be226bb326168c7107c821181df43be4dd71f9c305d4' WHERE version = 1;`); err != nil {
		t.Fatal(err)
	}
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatalf("reconcile payroll snapshot storage: %v", err)
	}
	verifyPayrollInputSnapshots(t, ctx, isolatedDB)

	var installed int
	if err := pool.QueryRowContext(ctx, "SELECT count(*) FROM hrm_schema_migrations").Scan(&installed); err != nil || installed != 1 {
		t.Fatalf("installed migration count: got %d, error %v", installed, err)
	}
	if _, err := pool.ExecContext(ctx, "DELETE FROM hrm_schema_migrations WHERE version = 1"); err != nil {
		t.Fatal(err)
	}
	downSQL, err := os.ReadFile("migrations/000001_hrm.down.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.ExecContext(ctx, string(downSQL)); err != nil {
		t.Fatal(err)
	}
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatalf("reinstall HRM baseline after down: %v", err)
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
	if err := hrdb.RunMigrationsUp(ctx, pool); !errors.Is(err, hrerrors.ErrConflict) {
		t.Fatalf("unknown version: got %v, want conflict", err)
	}
	if _, err := pool.ExecContext(ctx, "DELETE FROM hrm_schema_migrations WHERE version = 8"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.ExecContext(ctx, "UPDATE hrm_schema_migrations SET checksum = repeat('0', 64) WHERE version = 1"); err != nil {
		t.Fatal(err)
	}
	if err := hrdb.RunMigrationsUp(ctx, pool); !errors.Is(err, hrerrors.ErrConflict) {
		t.Fatalf("checksum mismatch: got %v, want conflict", err)
	}
}

func verifyPayrollInputSnapshots(t *testing.T, ctx context.Context, db *gorm.DB) {
	t.Helper()
	const period = "10000000-0000-4000-8000-000000000001"
	const group = "10000000-0000-4000-8000-000000000002"
	const employer = "10000000-0000-4000-8000-000000000003"
	const snapshotID = "10000000-0000-4000-8000-000000000004"
	const creator = "10000000-0000-4000-8000-000000000005"
	if err := db.WithContext(ctx).Exec(`INSERT INTO hr_payroll_groups
		(id, group_code, group_name, employing_legal_entity_reference_id, jurisdiction_code, currency_code, currency_scale, pay_frequency, timezone)
		VALUES (?, 'TEST', 'Fixture', ?, 'ID', 'IDR', 0, 'monthly', 'Asia/Jakarta')`, group, employer).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(`INSERT INTO hr_payroll_periods
		(id, payroll_group_id, employing_legal_entity_reference_id, currency_code, jurisdiction_code, period_code, period_start, period_end, pay_date)
		VALUES (?, ?, ?, 'IDR', 'ID', 'TEST', CURRENT_DATE, CURRENT_DATE, CURRENT_DATE)`, period, group, employer).Error; err != nil {
		t.Fatal(err)
	}
	repo, err := payrollrepo.NewInputSnapshotRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("frozen fixture")
	digest := sha256.Sum256(content)
	snapshot := entity.PayrollInputSnapshot{
		ID: snapshotID, PeriodID: period, GroupID: group, EmployerID: employer, JurisdictionCode: "ID",
		CurrencyCode: "IDR", Content: content, Digest: digest[:], CreatedBy: creator, CreatedAt: time.Now(),
	}
	invalid := snapshot
	invalid.Digest = make([]byte, sha256.Size)
	if err := repo.Insert(ctx, &invalid); !errors.Is(err, hrerrors.ErrConstraintViolation) {
		t.Fatalf("bad digest: %v", err)
	}
	invalid = snapshot
	invalid.GroupID = employer
	if err := repo.Insert(ctx, &invalid); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("wrong group: %v", err)
	}
	if err := repo.Insert(ctx, &snapshot); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ReadCalculationSnapshot(ctx, snapshotID)
	if err != nil || got.PeriodID != period || got.GroupID != group || got.EmployerID != employer || !bytes.Equal(got.Content, content) {
		t.Fatalf("read immutable snapshot: %v", err)
	}
	for _, query := range []string{"UPDATE hr_payroll_input_snapshots SET content = content WHERE id = ?", "DELETE FROM hr_payroll_input_snapshots WHERE id = ?"} {
		if err := db.WithContext(ctx).Exec(query, snapshotID).Error; err == nil {
			t.Fatalf("immutable snapshot allowed mutation: %q", query)
		}
	}
	const worker = "10000000-0000-4000-8000-000000000006"
	const profile = "10000000-0000-4000-8000-000000000007"
	const captureID = "10000000-0000-4000-8000-000000000008"
	if _, err := repo.Capture(ctx, captureID, period, creator, time.Now()); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("empty payroll scope: %v", err)
	}
	if err := db.WithContext(ctx).Exec(`INSERT INTO hr_workers
		(id, worker_number, person_party_reference_id, worker_category, status, hire_date, primary_legal_entity_reference_id)
		VALUES (?, 'SNAPSHOT-WORKER', ?, 'employee', 'active', CURRENT_DATE, ?)`, worker, worker, employer).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(`INSERT INTO hr_payroll_profiles
		(id, worker_id, payroll_country_code, payroll_jurisdiction_code, payroll_group_id, employing_legal_entity_reference_id,
		 currency_code, bank_account_token_reference_id, effective_from)
		VALUES (?, ?, 'ID', 'ID', ?, ?, 'IDR', ?, CURRENT_DATE)`, profile, worker, group, employer, creator).Error; err != nil {
		t.Fatal(err)
	}
	captured, err := repo.Capture(ctx, captureID, period, creator, time.Now())
	if err != nil || captured.ID != captureID || len(captured.Digest) != sha256.Size || len(captured.Content) != 0 {
		t.Fatalf("capture payroll input metadata: %v", err)
	}
	got, err = repo.ReadCalculationSnapshot(ctx, captureID)
	var frozen struct {
		Profiles []struct{ ID string }
	}
	if err != nil || json.Unmarshal(got.Content, &frozen) != nil || len(frozen.Profiles) != 1 || frozen.Profiles[0].ID != profile ||
		bytes.Contains(got.Content, []byte(creator)) {
		t.Fatalf("captured content: profiles=%d error=%v", len(frozen.Profiles), err)
	}

}
