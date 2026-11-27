package db_test

import (
	"context"
	"database/sql"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	efdb "github.com/integrione/erp-retail/backend/services/enterprise_foundation/db"
)

func TestDeriveTestDatabaseDSN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		devDSN  string
		want    string
		wantErr bool
	}{
		{
			name:   "uri postgres scheme appends _test suffix and keeps query parameters",
			devDSN: "postgres://postgres:secret@127.0.0.1:6543/enterprise_foundation?sslmode=disable&TimeZone=Asia/Jakarta",
			want:   "postgres://postgres:secret@127.0.0.1:6543/enterprise_foundation_test?sslmode=disable&TimeZone=Asia/Jakarta",
		},
		{
			name:   "uri postgresql scheme appends _test suffix",
			devDSN: "postgresql://postgres:secret@127.0.0.1:6543/enterprise_foundation?sslmode=disable",
			want:   "postgresql://postgres:secret@127.0.0.1:6543/enterprise_foundation_test?sslmode=disable",
		},
		{
			name:   "uri dsn already ending with _test stays unchanged",
			devDSN: "postgres://postgres:secret@127.0.0.1:6543/enterprise_foundation_test?sslmode=disable",
			want:   "postgres://postgres:secret@127.0.0.1:6543/enterprise_foundation_test?sslmode=disable",
		},
		{
			name:   "keyword dsn appends _test suffix",
			devDSN: "host=127.0.0.1 port=6543 user=postgres password=secret dbname=enterprise_foundation sslmode=disable",
			want:   "host=127.0.0.1 port=6543 user=postgres password=secret dbname=enterprise_foundation_test sslmode=disable",
		},
		{
			name:    "empty dsn fails",
			devDSN:  "",
			wantErr: true,
		},
		{
			name:    "uri dsn without database name fails",
			devDSN:  "postgres://postgres:secret@127.0.0.1:6543/?sslmode=disable",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := efdb.DeriveTestDatabaseDSN(tc.devDSN)
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

func TestValidationAndGuards(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Init fails fast when isTesting true without _test suffix", func(t *testing.T) {
		t.Parallel()
		_, err := efdb.Init(ctx, "postgres://postgres:secret@127.0.0.1:6543/enterprise_foundation?sslmode=disable", true)
		if err == nil || !strings.Contains(err.Error(), "_test") {
			t.Fatalf("expected '_test' validation error, got: %v", err)
		}
	})

	t.Run("Init fails fast on empty DSN", func(t *testing.T) {
		t.Parallel()
		if _, err := efdb.Init(ctx, "   ", false); err == nil {
			t.Fatal("expected error for empty DSN")
		}
	})

	t.Run("EnsureDatabase rejects empty or unsafe database names", func(t *testing.T) {
		t.Parallel()
		for _, badDSN := range []string{
			"",
			"postgres://postgres:secret@127.0.0.1:6543/?sslmode=disable",
			"host=127.0.0.1 port=6543 user=postgres dbname=bad-name;drop sslmode=disable",
		} {
			if err := efdb.EnsureDatabase(ctx, badDSN); err == nil {
				t.Fatalf("expected EnsureDatabase(%q) to fail", badDSN)
			}
		}
	})

	t.Run("RunMigrationsUp and RunMigrationsDown fail fast on nil handle", func(t *testing.T) {
		t.Parallel()
		if err := efdb.RunMigrationsUp(ctx, nil); err == nil {
			t.Fatal("expected RunMigrationsUp(nil) to fail")
		}
		if err := efdb.RunMigrationsDown(ctx, nil); err == nil {
			t.Fatal("expected RunMigrationsDown(nil) to fail")
		}
	})

	t.Run("embedded MigrationFiles includes baseline v000001 up and down SQL", func(t *testing.T) {
		t.Parallel()
		for _, file := range []string{
			"migrations/000001_create_enterprise_foundation_schema.up.sql",
			"migrations/000001_create_enterprise_foundation_schema.down.sql",
		} {
			data, err := efdb.MigrationFiles.ReadFile(file)
			if err != nil {
				t.Fatalf("expected embedded file %s: %v", file, err)
			}
			if len(strings.TrimSpace(string(data))) == 0 {
				t.Fatalf("embedded file %s must not be empty", file)
			}
		}
	})
}

func TestEnsureAndMigrateDevAndTestDatabases(t *testing.T) {
	devDSN := strings.TrimSpace(os.Getenv("EF_DATABASE_DSN"))
	testDSN := strings.TrimSpace(os.Getenv("EF_TEST_DATABASE_DSN"))
	if devDSN == "" && testDSN == "" {
		t.Skip("skipping live PostgreSQL test: set EF_DATABASE_DSN or EF_TEST_DATABASE_DSN to run")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var devSQL, testSQL *sql.DB
	if devDSN == "" {
		testDB, err := efdb.Init(ctx, testDSN, true)
		if err != nil {
			t.Fatalf("failed to initialize test database: %v", err)
		}
		testSQL, err = testDB.DB()
		if err != nil {
			t.Fatalf("failed to obtain test sql.DB: %v", err)
		}
		defer func() { _ = testSQL.Close() }()
	} else {
		devDB, testDB, err := efdb.InitDevAndTest(ctx, devDSN, testDSN)
		if err != nil {
			t.Fatalf("failed to initialize dev and test databases: %v", err)
		}
		var dbErr error
		devSQL, dbErr = devDB.DB()
		if dbErr != nil {
			t.Fatalf("failed to obtain dev sql.DB: %v", dbErr)
		}
		defer func() { _ = devSQL.Close() }()

		testSQL, dbErr = testDB.DB()
		if dbErr != nil {
			t.Fatalf("failed to obtain test sql.DB: %v", dbErr)
		}
		defer func() { _ = testSQL.Close() }()
	}

	assertFullSchemaObjects(t, ctx, testSQL)

	if devSQL != nil {
		assertFullSchemaObjects(t, ctx, devSQL)

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
	}

	// Uji perilaku constraint, partial unique index, FK chain, dan trigger pada DB testing.
	assertConstraintsAndTriggers(t, ctx, testSQL)

	// Uji fail-closed saat status schema_migrations kotor (dirty = true).
	assertDirtyMigrationFailClosed(t, ctx, testSQL)

	// Verifikasi rollback bersih menghapus seluruh tabel, trigger, & enum domain pada enterprise_foundation_test.
	if err := efdb.RunMigrationsDown(ctx, testSQL); err != nil {
		t.Fatalf("rollback on test database failed: %v", err)
	}
	afterDownTables, err := listPublicTables(ctx, testSQL)
	if err != nil {
		t.Fatalf("failed to list tables after down: %v", err)
	}
	for _, unexpected := range expectedBusinessTables() {
		if slices.Contains(afterDownTables, unexpected) {
			t.Fatalf("expected table %q to be dropped after down migration", unexpected)
		}
	}
	afterDownEnums, err := listPublicEnums(ctx, testSQL)
	if err != nil {
		t.Fatalf("failed to list enums after down: %v", err)
	}
	if len(afterDownEnums) != 0 {
		t.Fatalf("expected all enums dropped after down migration, still got: %v", afterDownEnums)
	}

	// Terapkan ulang migrasi up dan pastikan idempoten saat dijalankan dua kali.
	if err := efdb.RunMigrationsUp(ctx, testSQL); err != nil {
		t.Fatalf("re-applying migrations on test database failed: %v", err)
	}
	if err := efdb.RunMigrationsUp(ctx, testSQL); err != nil {
		t.Fatalf("idempotent re-run on test database failed: %v", err)
	}

	assertFullSchemaObjects(t, ctx, testSQL)
}

func assertFullSchemaObjects(t *testing.T, ctx context.Context, sqlDB *sql.DB) {
	t.Helper()

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

	gotIndexes, err := listPublicIndexes(ctx, sqlDB)
	if err != nil {
		t.Fatalf("failed to list indexes: %v", err)
	}
	assertExpectedIndexes(t, gotIndexes)

	gotTriggers, err := listPublicUpdateTriggers(ctx, sqlDB)
	if err != nil {
		t.Fatalf("failed to list triggers: %v", err)
	}
	assertExpectedTriggers(t, gotTriggers)
}

func assertDirtyMigrationFailClosed(t *testing.T, ctx context.Context, sqlDB *sql.DB) {
	t.Helper()

	if _, err := sqlDB.ExecContext(ctx, `
		INSERT INTO schema_migrations (version, dirty, applied_at)
		VALUES (999999, true, NOW());
	`); err != nil {
		t.Fatalf("insert synthetic dirty row: %v", err)
	}
	defer func() {
		_, _ = sqlDB.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = 999999;`)
	}()

	if err := efdb.RunMigrationsUp(ctx, sqlDB); err == nil || !strings.Contains(err.Error(), "dirty") {
		t.Fatalf("expected RunMigrationsUp to fail closed on dirty schema, got: %v", err)
	}
	if err := efdb.RunMigrationsDown(ctx, sqlDB); err == nil || !strings.Contains(err.Error(), "dirty") {
		t.Fatalf("expected RunMigrationsDown to fail closed on dirty schema, got: %v", err)
	}
}

func assertConstraintsAndTriggers(t *testing.T, ctx context.Context, sqlDB *sql.DB) {
	t.Helper()

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin verification tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Insert master alamat, legal entity, location, dan organization_unit (happy-path FK chain).
	var addressID, shippingAddrID, legalEntityID, locationID, orgUnitID string
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO addresses (label, address_line_1, city, province, postal_code)
		VALUES ('Kantor Utama', 'Jl. Sudirman No. 10', 'Jakarta Selatan', 'DKI Jakarta', '12190')
		RETURNING id::text;
	`).Scan(&addressID); err != nil {
		t.Fatalf("insert primary address: %v", err)
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO addresses (label, address_line_1, city)
		VALUES ('Gudang Mitra', 'Jl. Industri Raya No. 8', 'Bandung')
		RETURNING id::text;
	`).Scan(&shippingAddrID); err != nil {
		t.Fatalf("insert shipping address: %v", err)
	}

	var initialUpdatedAt time.Time
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO legal_entities (entity_code, legal_name, is_primary_operating_entity, registered_address_id, updated_at)
		VALUES ('PT-TEST-01', 'PT IntegriOne Uji', true, $1::uuid, NOW() - INTERVAL '1 hour')
		RETURNING id::text, updated_at;
	`, addressID).Scan(&legalEntityID, &initialUpdatedAt); err != nil {
		t.Fatalf("insert test legal entity: %v", err)
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO locations (legal_entity_id, address_id, location_code, name, location_kind)
		VALUES ($1::uuid, $2::uuid, 'WH-MAIN', 'Gudang Utama', 'warehouse')
		RETURNING id::text;
	`, legalEntityID, addressID).Scan(&locationID); err != nil {
		t.Fatalf("insert test location: %v", err)
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO organization_units (legal_entity_id, primary_location_id, unit_code, name, unit_kind)
		VALUES ($1::uuid, $2::uuid, 'WH-OPS', 'Tim Operasional Gudang', 'department')
		RETURNING id::text;
	`, legalEntityID, locationID).Scan(&orgUnitID); err != nil {
		t.Fatalf("insert test organization_unit: %v", err)
	}

	// 2. Verifikasi partial unique index active duplicate gagal, namun soft-deleted memperbolehkan pengganti.
	if _, err := tx.ExecContext(ctx, `SAVEPOINT sp_dup_active;`); err != nil {
		t.Fatalf("savepoint sp_dup_active: %v", err)
	}
	if _, dupErr := tx.ExecContext(ctx, `
		INSERT INTO legal_entities (entity_code, legal_name)
		VALUES ('PT-TEST-01', 'PT Duplikat Aktif');
	`); dupErr == nil {
		t.Fatal("expected partial unique index uq_legal_entities_code to reject active duplicate entity_code")
	}
	if _, err := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT sp_dup_active;`); err != nil {
		t.Fatalf("rollback to sp_dup_active: %v", err)
	}

	var bumpedUpdatedAt time.Time
	if err := tx.QueryRowContext(ctx, `
		UPDATE legal_entities
		SET deleted_at = NOW(), trading_name = 'IntegriOne Arsip'
		WHERE id = $1::uuid
		RETURNING updated_at;
	`, legalEntityID).Scan(&bumpedUpdatedAt); err != nil {
		t.Fatalf("soft-delete legal entity: %v", err)
	}
	if !bumpedUpdatedAt.After(initialUpdatedAt) {
		t.Fatalf("expected trg_legal_entities_updated_at to advance updated_at: before=%v after=%v", initialUpdatedAt, bumpedUpdatedAt)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO legal_entities (entity_code, legal_name, is_primary_operating_entity)
		VALUES ('PT-TEST-01', 'PT IntegriOne Pengganti', true);
	`); err != nil {
		t.Fatalf("expected soft-deleted entity_code and primary flag to allow active replacement, got: %v", err)
	}

	// 3. Verifikasi self-parent check constraint pada organization_units ditolak.
	if _, err := tx.ExecContext(ctx, `SAVEPOINT sp_self_parent;`); err != nil {
		t.Fatalf("savepoint sp_self_parent: %v", err)
	}
	if _, selfParentErr := tx.ExecContext(ctx, `
		UPDATE organization_units SET parent_unit_id = $1::uuid WHERE id = $1::uuid;
	`, orgUnitID); selfParentErr == nil {
		t.Fatal("expected ck_organization_units_not_self_parent to reject self-parent")
	}
	if _, err := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT sp_self_parent;`); err != nil {
		t.Fatalf("rollback to sp_self_parent: %v", err)
	}

	// 4. Verifikasi party, party_roles (supplier & vendor), party_addresses primary invariant, dan party_contacts.
	var partyID string
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO parties (partner_code, party_kind, display_name, status)
		VALUES ('MITRA-TEST-01', 'organization', 'Konveksi Maju Jaya', 'active')
		RETURNING id::text;
	`).Scan(&partyID); err != nil {
		t.Fatalf("insert test party: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO party_roles (party_id, role, is_active)
		VALUES ($1::uuid, 'supplier', true), ($1::uuid, 'vendor', true);
	`, partyID); err != nil {
		t.Fatalf("insert dual party_roles supplier+vendor: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO party_addresses (party_id, address_id, usage, is_primary)
		VALUES ($1::uuid, $2::uuid, 'shipping', true);
	`, partyID, shippingAddrID); err != nil {
		t.Fatalf("insert primary shipping address: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `SAVEPOINT sp_dup_primary_addr;`); err != nil {
		t.Fatalf("savepoint sp_dup_primary_addr: %v", err)
	}
	if _, dupPrimaryErr := tx.ExecContext(ctx, `
		INSERT INTO party_addresses (party_id, address_id, usage, is_primary)
		VALUES ($1::uuid, $2::uuid, 'shipping', true);
	`, partyID, addressID); dupPrimaryErr == nil {
		t.Fatal("expected uq_party_addresses_primary_per_usage to reject second primary shipping address")
	}
	if _, err := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT sp_dup_primary_addr;`); err != nil {
		t.Fatalf("rollback to sp_dup_primary_addr: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `SAVEPOINT sp_bad_portal_status;`); err != nil {
		t.Fatalf("savepoint sp_bad_portal_status: %v", err)
	}
	if _, checkErr := tx.ExecContext(ctx, `
		INSERT INTO party_contacts (party_id, full_name, email, portal_access_status)
		VALUES ($1::uuid, 'Budi PIC', 'budi@vendor.test', 'active');
	`, partyID); checkErr == nil {
		t.Fatal("expected ck_party_contacts_portal_link_invariant to reject portal_access_status='active' when iam_user_id is NULL")
	}
	if _, err := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT sp_bad_portal_status;`); err != nil {
		t.Fatalf("rollback to sp_bad_portal_status: %v", err)
	}

	// 5. Verifikasi outbox insert & attempt_count invariant.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO foundation_outbox_events (event_type, aggregate_type, aggregate_id, payload)
		VALUES ('party.status_changed', 'party', $1::uuid, '{"status":"active"}'::jsonb);
	`, partyID); err != nil {
		t.Fatalf("insert foundation_outbox_events: %v", err)
	}
}

func expectedBusinessTables() []string {
	return []string{
		"addresses",
		"foundation_outbox_events",
		"legal_entities",
		"locations",
		"organization_units",
		"parties",
		"party_addresses",
		"party_contacts",
		"party_roles",
	}
}

func assertExpectedTables(t *testing.T, got []string) {
	t.Helper()

	expected := append(expectedBusinessTables(), "schema_migrations")
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
		"address_usage_kind",
		"legal_entity_status",
		"location_kind",
		"org_unit_kind",
		"outbox_status",
		"party_kind",
		"party_role_kind",
		"party_status",
		"portal_access_status",
	}
	sort.Strings(expected)

	for _, want := range expected {
		if !slices.Contains(got, want) {
			t.Fatalf("missing expected enum %q; got enums: %v", want, got)
		}
	}
}

func assertExpectedIndexes(t *testing.T, got []string) {
	t.Helper()

	expected := []string{
		"ix_addresses_city_province",
		"ix_addresses_country_postal",
		"ix_foundation_outbox_aggregate",
		"ix_foundation_outbox_pending",
		"ix_legal_entities_deleted_at",
		"ix_legal_entities_parent",
		"ix_legal_entities_status",
		"ix_locations_address",
		"ix_locations_deleted_at",
		"ix_locations_entity_kind_active",
		"ix_locations_parent",
		"ix_organization_units_deleted_at",
		"ix_organization_units_entity_active",
		"ix_organization_units_location",
		"ix_organization_units_parent",
		"ix_parties_deleted_at",
		"ix_parties_display_name",
		"ix_parties_display_name_trgm",
		"ix_parties_kind_status",
		"ix_party_addresses_address",
		"ix_party_contacts_deleted_at",
		"ix_party_contacts_iam_user",
		"ix_party_contacts_party_lookup",
		"ix_party_roles_role_active",
		"uq_legal_entities_code",
		"uq_legal_entities_primary_operating",
		"uq_legal_entities_tax_id",
		"uq_locations_code",
		"uq_organization_units_entity_code",
		"uq_parties_partner_code",
		"uq_parties_tax_id_number",
		"uq_party_addresses_party_addr_usage",
		"uq_party_addresses_primary_per_usage",
		"uq_party_contacts_party_iam_user",
		"uq_party_roles_party_role",
	}

	for _, want := range expected {
		if !slices.Contains(got, want) {
			t.Fatalf("missing expected index %q; got indexes: %v", want, got)
		}
	}
}

func assertExpectedTriggers(t *testing.T, got []string) {
	t.Helper()

	expected := []string{
		"trg_addresses_updated_at",
		"trg_foundation_outbox_events_updated_at",
		"trg_legal_entities_updated_at",
		"trg_locations_updated_at",
		"trg_organization_units_updated_at",
		"trg_parties_updated_at",
		"trg_party_addresses_updated_at",
		"trg_party_contacts_updated_at",
		"trg_party_roles_updated_at",
	}

	for _, want := range expected {
		if !slices.Contains(got, want) {
			t.Fatalf("missing expected trigger %q; got triggers: %v", want, got)
		}
	}
}

func listPublicTables(ctx context.Context, sqlDB *sql.DB) ([]string, error) {
	const query = `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		ORDER BY tablename;`

	return queryStrings(ctx, sqlDB, query)
}

func listPublicEnums(ctx context.Context, sqlDB *sql.DB) ([]string, error) {
	const query = `
		SELECT t.typname
		FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE n.nspname = 'public' AND t.typtype = 'e'
		ORDER BY t.typname;`

	return queryStrings(ctx, sqlDB, query)
}

func listPublicIndexes(ctx context.Context, sqlDB *sql.DB) ([]string, error) {
	const query = `
		SELECT indexname
		FROM pg_indexes
		WHERE schemaname = 'public'
		ORDER BY indexname;`

	return queryStrings(ctx, sqlDB, query)
}

func listPublicUpdateTriggers(ctx context.Context, sqlDB *sql.DB) ([]string, error) {
	const query = `
		SELECT DISTINCT trigger_name
		FROM information_schema.triggers
		WHERE trigger_schema = 'public'
		ORDER BY trigger_name;`

	return queryStrings(ctx, sqlDB, query)
}

func listTableColumns(ctx context.Context, sqlDB *sql.DB) ([]string, error) {
	const query = `
		SELECT table_name || '.' || column_name || ':' || udt_name || ':' || is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position;`

	return queryStrings(ctx, sqlDB, query)
}

func queryStrings(ctx context.Context, sqlDB *sql.DB, query string) ([]string, error) {
	rows, err := sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []string
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		items = append(items, val)
	}
	return items, rows.Err()
}
