package db

import (
	"context"
	"database/sql"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestDeriveTestDatabaseDSN(t *testing.T) {
	tests := []struct {
		name    string
		devDSN  string
		want    string
		wantErr bool
	}{
		{
			name:   "uri dsn appends _test suffix",
			devDSN: "postgres://postgres:secret@127.0.0.1:6543/iam?sslmode=disable",
			want:   "postgres://postgres:secret@127.0.0.1:6543/iam_test?sslmode=disable",
		},
		{
			name:   "uri dsn already ending with _test stays unchanged",
			devDSN: "postgres://postgres:secret@127.0.0.1:6543/iam_test?sslmode=disable",
			want:   "postgres://postgres:secret@127.0.0.1:6543/iam_test?sslmode=disable",
		},
		{
			name:   "keyword dsn appends _test suffix",
			devDSN: "host=127.0.0.1 port=6543 user=postgres password=secret dbname=iam sslmode=disable",
			want:   "host=127.0.0.1 port=6543 user=postgres password=secret dbname=iam_test sslmode=disable",
		},
		{
			name:    "empty dsn fails",
			devDSN:  "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DeriveTestDatabaseDSN(tc.devDSN)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestInitFailFastWhenTestDBMissingSuffix(t *testing.T) {
	ctx := context.Background()
	_, err := Init(ctx, "postgres://postgres:secret@127.0.0.1:6543/iam?sslmode=disable", true)
	if err == nil {
		t.Fatalf("expected error when isTesting=true and database does not end with '_test'")
	}
	if !strings.Contains(err.Error(), "_test") {
		t.Fatalf("expected error mentioning '_test', got: %v", err)
	}
}

func TestEnsureAndMigrateDevAndTestDatabases(t *testing.T) {
	devDSN := strings.TrimSpace(os.Getenv("IAM_DATABASE_DSN"))
	testDSN := strings.TrimSpace(os.Getenv("IAM_TEST_DATABASE_DSN"))
	if devDSN == "" && testDSN == "" {
		t.Skip("skipping live PostgreSQL test: set IAM_DATABASE_DSN or IAM_TEST_DATABASE_DSN to run")
	}

	ctx := context.Background()
	if devDSN == "" {
		testDB, err := Init(ctx, testDSN, true)
		if err != nil {
			t.Fatalf("failed to initialize test database: %v", err)
		}
		sqlDB, err := testDB.DB()
		if err != nil {
			t.Fatalf("failed to get sql.DB: %v", err)
		}
		defer func() { _ = sqlDB.Close() }()

		gotTables, err := listPublicTables(ctx, sqlDB)
		if err != nil {
			t.Fatalf("failed to list tables: %v", err)
		}
		assertExpectedTables(t, gotTables)

		gotEnums, err := listPublicEnums(ctx, sqlDB)
		if err != nil {
			t.Fatalf("failed to list enums: %v", err)
		}
		assertExpectedEnums(t, gotEnums)
		assertDisposableRollback(t, ctx, sqlDB)
		return
	}

	devDB, testDB, err := InitDevAndTest(ctx, devDSN, testDSN)
	if err != nil {
		t.Fatalf("failed to initialize dev and test databases: %v", err)
	}

	devSQL, err := devDB.DB()
	if err != nil {
		t.Fatalf("failed to get dev sql.DB: %v", err)
	}
	defer func() { _ = devSQL.Close() }()

	testSQL, err := testDB.DB()
	if err != nil {
		t.Fatalf("failed to get test sql.DB: %v", err)
	}
	defer func() { _ = testSQL.Close() }()

	// Verifikasi daftar tabel di iam dan iam_test.
	devTables, err := listPublicTables(ctx, devSQL)
	if err != nil {
		t.Fatalf("failed to list dev tables: %v", err)
	}
	testTables, err := listPublicTables(ctx, testSQL)
	if err != nil {
		t.Fatalf("failed to list test tables: %v", err)
	}

	assertExpectedTables(t, devTables)
	if !slices.Equal(devTables, testTables) {
		t.Fatalf("tables mismatch between dev and test databases: dev=%v test=%v", devTables, testTables)
	}

	// Verifikasi daftar enum di iam dan iam_test.
	devEnums, err := listPublicEnums(ctx, devSQL)
	if err != nil {
		t.Fatalf("failed to list dev enums: %v", err)
	}
	testEnums, err := listPublicEnums(ctx, testSQL)
	if err != nil {
		t.Fatalf("failed to list test enums: %v", err)
	}

	assertExpectedEnums(t, devEnums)
	if !slices.Equal(devEnums, testEnums) {
		t.Fatalf("enums mismatch between dev and test databases: dev=%v test=%v", devEnums, testEnums)
	}

	// Verifikasi seluruh definisi kolom tabel identik antara iam dan iam_test.
	devCols, err := listTableColumns(ctx, devSQL)
	if err != nil {
		t.Fatalf("failed to list dev columns: %v", err)
	}
	testCols, err := listTableColumns(ctx, testSQL)
	if err != nil {
		t.Fatalf("failed to list test columns: %v", err)
	}
	if !slices.Equal(devCols, testCols) {
		t.Fatalf("columns mismatch between dev and test databases")
	}

	assertDisposableRollback(t, ctx, testSQL)

	// Pastikan migrasi ulang bersifat idempoten.
	if err := RunMigrationsUp(ctx, testSQL); err != nil {
		t.Fatalf("idempotent re-run on test database failed: %v", err)
	}
}

func assertDisposableRollback(t *testing.T, ctx context.Context, testSQL *sql.DB) {
	t.Helper()
	// Rollback skema hanya aman pada database sementara yang memang boleh dihapus.
	if os.Getenv("IAM_TEST_DATABASE_DISPOSABLE") != "1" {
		return
	}
	var databaseName string
	if err := testSQL.QueryRowContext(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		t.Fatalf("identify disposable test database: %v", err)
	}
	if !strings.HasPrefix(databaseName, "iam_disposable_") || !strings.HasSuffix(databaseName, "_test") {
		t.Fatalf("refusing rollback of non-disposable database %q", databaseName)
	}
	if err := RunMigrationsDown(ctx, testSQL); err != nil {
		t.Fatalf("rollback on disposable test database failed: %v", err)
	}
	if err := RunMigrationsUp(ctx, testSQL); err != nil {
		t.Fatalf("re-applying migrations on disposable test database failed: %v", err)
	}
}

func listPublicTables(ctx context.Context, sqlDB *sql.DB) ([]string, error) {
	const query = `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		ORDER BY tablename;`

	rows, err := sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func listPublicEnums(ctx context.Context, sqlDB *sql.DB) ([]string, error) {
	const query = `
		SELECT t.typname
		FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE n.nspname = 'public' AND t.typtype = 'e'
		ORDER BY t.typname;`

	rows, err := sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var enums []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		enums = append(enums, name)
	}
	return enums, rows.Err()
}

func listTableColumns(ctx context.Context, sqlDB *sql.DB) ([]string, error) {
	const query = `
		SELECT table_name || '.' || column_name || ':' || udt_name || ':' || is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position;`

	rows, err := sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var cols []string
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		cols = append(cols, item)
	}
	return cols, rows.Err()
}

func assertExpectedTables(t *testing.T, got []string) {
	t.Helper()

	expected := []string{
		"group_members",
		"group_roles",
		"groups",
		"permissions",
		"role_permissions",
		"roles",
		"schema_migrations",
		"sessions",
		"user_credentials",
		"user_identities",
		"user_recovery_codes",
		"user_roles",
		"user_tokens",
		"user_totp_factors",
		"users",
	}
	sort.Strings(expected)

	for _, want := range expected {
		if !slices.Contains(got, want) {
			t.Fatalf("missing expected table %q; got tables: %v", want, got)
		}
	}
}

func assertExpectedEnums(t *testing.T, got []string) {
	t.Helper()

	expected := []string{
		"scope_type",
		"session_end_reason",
		"token_purpose",
		"user_identities_kind",
		"user_kind",
		"user_status",
	}
	sort.Strings(expected)

	for _, want := range expected {
		if !slices.Contains(got, want) {
			t.Fatalf("missing expected enum %q; got enums: %v", want, got)
		}
	}
}
