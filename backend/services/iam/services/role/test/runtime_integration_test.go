//go:build integration

package role_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/integrione/erp-retail/backend/services/iam/db"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services/group"
	"github.com/integrione/erp-retail/backend/services/iam/services/role"
)

type failingOutboxRuntime struct {
	domainrepo.SecurityRuntimeRepository
	err error
}

func (r failingOutboxRuntime) Insert(context.Context, *entity.OutboxEvent) error {
	return r.err
}

func TestRoleRevocationRollsBackWhenOutboxFails(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("IAM_TEST_DATABASE_DSN"))
	if os.Getenv("IAM_TEST_DATABASE_DISPOSABLE") != "1" || dsn == "" {
		t.Skip("disposable IAM test database is required")
	}
	parsed, err := pgx.ParseConfig(dsn)
	if err != nil || !strings.HasPrefix(parsed.Database, "iam_disposable_") || !strings.HasSuffix(parsed.Database, "_test") {
		t.Fatal("IAM_TEST_DATABASE_DSN must target a disposable IAM database")
	}
	ctx := context.Background()
	conn, err := db.Init(ctx, dsn, true)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := conn.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = sqlDB.Close()
		adminConfig := parsed.Copy()
		adminConfig.Database = "postgres"
		admin, connectErr := pgx.ConnectConfig(context.Background(), adminConfig)
		if connectErr != nil {
			t.Errorf("connect for disposable database cleanup: %v", connectErr)
			return
		}
		defer func() { _ = admin.Close(context.Background()) }()
		_, dropErr := admin.Exec(context.Background(), "DROP DATABASE IF EXISTS "+pgx.Identifier{parsed.Database}.Sanitize()+" WITH (FORCE)")
		if dropErr != nil {
			t.Errorf("drop disposable database: %v", dropErr)
		}
	}()

	repos, err := repository.New(conn)
	if err != nil {
		t.Fatal(err)
	}
	const (
		userID        = "11111111-1111-4111-8111-111111111111"
		roleID        = "22222222-2222-4222-8222-222222222222"
		groupID       = "33333333-3333-4333-8333-333333333333"
		userRoleID    = "44444444-4444-4444-8444-444444444444"
		groupRoleID   = "55555555-5555-4555-8555-555555555555"
		groupMemberID = "66666666-6666-4666-8666-666666666666"
	)
	statements := []struct {
		query string
		args  []any
	}{
		{"INSERT INTO users (id, display_name, status) VALUES (?, ?, 'active')", []any{userID, "RBAC audit"}},
		{"INSERT INTO roles (id, code, name) VALUES (?, ?, ?)", []any{roleID, "rbac_audit", "RBAC Audit"}},
		{"INSERT INTO groups (id, code, name) VALUES (?, ?, ?)", []any{groupID, "rbac_audit", "RBAC Audit"}},
		{"INSERT INTO user_roles (id, user_id, role_id) VALUES (?, ?, ?)", []any{userRoleID, userID, roleID}},
		{"INSERT INTO group_members (id, group_id, user_id) VALUES (?, ?, ?)", []any{groupMemberID, groupID, userID}},
		{"INSERT INTO group_roles (id, group_id, role_id) VALUES (?, ?, ?)", []any{groupRoleID, groupID, roleID}},
	}
	for _, statement := range statements {
		if err := conn.Exec(statement.query, statement.args...).Error; err != nil {
			t.Fatal(err)
		}
	}

	outboxErr := errors.New("injected outbox failure")
	runtime := failingOutboxRuntime{SecurityRuntimeRepository: repos.SecurityRuntime, err: outboxErr}
	idGen := crypto.NewUUIDGenerator()
	clock := fixedClock{now: time.Now()}
	coord := redisstore.NewCoordinator(runtime, nil, idGen, clock, nil)
	roleSvc, err := role.NewService(role.Dependencies{
		Transactor: repos.Transactor, Roles: repos.Roles, RolePermissions: repos.RolePermissions,
		Permissions: repos.Permissions, Users: repos.Users, UserRoles: repos.UserRoles,
		SecurityRuntime: runtime, IDGen: idGen, Clock: clock, OutboxCoord: coord,
	})
	if err != nil {
		t.Fatal(err)
	}
	groupSvc, err := group.NewService(group.Dependencies{
		Transactor: repos.Transactor, Groups: repos.Groups, GroupMembers: repos.GroupMembers,
		GroupRoles: repos.GroupRoles, Roles: repos.Roles, Users: repos.Users,
		SecurityRuntime: runtime, IDGen: idGen, Clock: clock, OutboxCoord: coord,
	})
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		name       string
		revoke     func(context.Context) error
		table      string
		assignment string
	}{
		{"direct role", func(ctx context.Context) error { return roleSvc.RevokeUserRole(ctx, userRoleID) }, "user_roles", userRoleID},
		{"group role", func(ctx context.Context) error { return groupSvc.RevokeRole(ctx, groupRoleID) }, "group_roles", groupRoleID},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.revoke(ctx); !errors.Is(err, outboxErr) {
				t.Fatalf("revoke error = %v, want injected outbox failure", err)
			}
			var assignmentCount int64
			if err := conn.Table(check.table).Where("id = ?", check.assignment).Count(&assignmentCount).Error; err != nil {
				t.Fatal(err)
			}
			if assignmentCount != 1 {
				t.Fatalf("assignment count = %d, want 1 after rollback", assignmentCount)
			}
			var revision int64
			if err := conn.Table("users").Select("perm_revision").Where("id = ?", userID).Scan(&revision).Error; err != nil {
				t.Fatal(err)
			}
			if revision != 1 {
				t.Fatalf("permission revision = %d, want 1 after rollback", revision)
			}
		})
	}

	workingCoord := redisstore.NewCoordinator(repos.SecurityRuntime, nil, idGen, clock, nil)
	workingRoleSvc, err := role.NewService(role.Dependencies{
		Transactor: repos.Transactor, Roles: repos.Roles, RolePermissions: repos.RolePermissions,
		Permissions: repos.Permissions, Users: repos.Users, UserRoles: repos.UserRoles,
		SecurityRuntime: repos.SecurityRuntime, IDGen: idGen, Clock: clock, OutboxCoord: workingCoord,
	})
	if err != nil {
		t.Fatal(err)
	}
	workingGroupSvc, err := group.NewService(group.Dependencies{
		Transactor: repos.Transactor, Groups: repos.Groups, GroupMembers: repos.GroupMembers,
		GroupRoles: repos.GroupRoles, Roles: repos.Roles, Users: repos.Users,
		SecurityRuntime: repos.SecurityRuntime, IDGen: idGen, Clock: clock, OutboxCoord: workingCoord,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := workingRoleSvc.RevokeUserRole(ctx, userRoleID); err != nil {
		t.Fatalf("revoke direct role: %v", err)
	}
	if err := workingGroupSvc.RevokeRole(ctx, groupRoleID); err != nil {
		t.Fatalf("revoke group role: %v", err)
	}
	var revision int64
	if err := conn.Table("users").Select("perm_revision").Where("id = ?", userID).Scan(&revision).Error; err != nil {
		t.Fatal(err)
	}
	if revision != 3 {
		t.Fatalf("permission revision = %d, want 3 after both revocations", revision)
	}
	var remaining, events int64
	if err := conn.Table("user_roles").Where("id = ?", userRoleID).Count(&remaining).Error; err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("direct role count = %d, want 0", remaining)
	}
	if err := conn.Table("group_roles").Where("id = ?", groupRoleID).Count(&remaining).Error; err != nil {
		t.Fatal(err)
	}
	if remaining != 0 {
		t.Fatalf("group role count = %d, want 0", remaining)
	}
	if err := conn.Table("iam_outbox_events").Count(&events).Error; err != nil {
		t.Fatal(err)
	}
	if events != 2 {
		t.Fatalf("outbox events = %d, want 2", events)
	}
}
