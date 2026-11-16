package permission_test

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/permission"
)

func TestDeleteRejectsAssignedPermission(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, ddl := range []string{
		"CREATE TABLE permissions (id text PRIMARY KEY, code text NOT NULL, resource text NOT NULL, action text NOT NULL)",
		"CREATE TABLE role_permissions (permission_id text NOT NULL)",
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatal(err)
		}
	}
	permissionID := "11111111-1111-4111-8111-111111111111"
	if err := db.Exec("INSERT INTO permissions (id, code, resource, action) VALUES (?, ?, ?, ?)", permissionID, "iam.users.read", "iam.users", "read").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO role_permissions (permission_id) VALUES (?)", permissionID).Error; err != nil {
		t.Fatal(err)
	}
	repo, err := permission.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(context.Background(), permissionID); !errors.Is(err, iamerrors.ErrPermissionInUse) {
		t.Fatalf("Delete(assigned) error = %v, want permission in use", err)
	}
	var count int64
	if err := db.Table("permissions").Where("id = ?", permissionID).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("assigned permission count = %d, err = %v", count, err)
	}
	if err := db.Exec("DELETE FROM role_permissions WHERE permission_id = ?", permissionID).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(context.Background(), permissionID); err != nil {
		t.Fatalf("Delete(unassigned) error = %v", err)
	}
	if err := repo.Delete(context.Background(), permissionID); !errors.Is(err, iamerrors.ErrPermissionNotFound) {
		t.Fatalf("Delete(missing) error = %v, want permission not found", err)
	}
}
