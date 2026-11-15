package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
)

func TestRefreshAccessSnapshot_IdentityDashboardAccessExcludesSelfServiceOnly(t *testing.T) {
	t.Parallel()

	clock := &mutableClock{now: time.Date(2026, 11, 15, 11, 0, 0, 0, time.Local)}
	idGen := crypto.NewUUIDGenerator()
	tokGen := crypto.NewOpaqueTokenGenerator()
	sessRepo := newMockSessionRepo()
	userRepo := &mockUserRepo{
		users:     make(map[string]*entity.User),
		snapshots: make(map[string]*entity.AccessSnapshot),
		clock:     clock,
	}

	svc, err := session.NewService(session.Dependencies{
		Sessions:        sessRepo,
		Users:           userRepo,
		Clock:           clock,
		IDGen:           idGen,
		Tokens:          tokGen,
		SessionPolicy:   domainservice.NewDefaultSessionPolicy(),
		AccessEvaluator: domainservice.NewAccessEvaluator(domainservice.NewDefaultSessionPolicy()),
		Redirect:        domainservice.NewDefaultRedirectResolver(),
	})
	if err != nil {
		t.Fatalf("session.NewService() error = %v", err)
	}

	empSelfOnlyID, _ := idGen.NewUUID()
	empSelfOnly, _ := entity.NewUser(empSelfOnlyID, entity.UserKindEmployee, entity.UserStatusActive, "Employee Self", clock.Now())
	userRepo.users[empSelfOnlyID] = empSelfOnly
	selfScope, err := entity.NewScope(entity.ScopeTypeSelf, &empSelfOnlyID)
	if err != nil {
		t.Fatalf("NewScope(self) error = %v", err)
	}
	userRepo.snapshots[empSelfOnlyID] = &entity.AccessSnapshot{
		UserID:     empSelfOnlyID,
		UserKind:   entity.UserKindEmployee,
		UserStatus: entity.UserStatusActive,
		Roles: []entity.Role{
			{
				ID:                  "11111111-1111-4111-8111-111111111111",
				Code:                "iam_selfservice",
				Name:                "IAM Account Self-Service",
				ManagementDomain:    "iam",
				DefaultRedirectPath: "",
				RedirectPriority:    999,
				IsSystem:            true,
				IsActive:            true,
			},
			{
				ID:                  "22222222-2222-4222-8222-222222222222",
				Code:                "wms_operator",
				Name:                "WMS Operator",
				ManagementDomain:    "warehouse",
				DefaultRedirectPath: "/warehouse",
				RedirectPriority:    20,
				IsActive:            true,
			},
		},
		Permissions: []entity.ScopedPermission{
			{PermissionCode: "iam.users.readself", Scope: selfScope},
			{PermissionCode: "iam.users.updateself", Scope: selfScope},
			{PermissionCode: "iam.users.deleteself", Scope: selfScope},
			{PermissionCode: "iam.sessions.readself", Scope: selfScope},
			{PermissionCode: "iam.sessions.deleteself", IsSensitive: true, Scope: selfScope},
		},
		CheckedAt: clock.Now(),
	}

	resSelf, err := svc.RefreshAccessSnapshot(context.Background(), empSelfOnlyID)
	if err != nil {
		t.Fatalf("RefreshAccessSnapshot(empSelfOnly) error = %v", err)
	}
	if resSelf.CanAccessIdentity {
		t.Fatal("user with only *self IAM permissions must have CanAccessIdentity = false")
	}
	if resSelf.DefaultRedirectPath != "/warehouse" {
		t.Fatalf("DefaultRedirectPath = %q, want /warehouse", resSelf.DefaultRedirectPath)
	}

	empIAMAdminID, _ := idGen.NewUUID()
	empIAMAdmin, _ := entity.NewUser(empIAMAdminID, entity.UserKindEmployee, entity.UserStatusActive, "Employee IAM", clock.Now())
	userRepo.users[empIAMAdminID] = empIAMAdmin
	userRepo.snapshots[empIAMAdminID] = &entity.AccessSnapshot{
		UserID:     empIAMAdminID,
		UserKind:   entity.UserKindEmployee,
		UserStatus: entity.UserStatusActive,
		Permissions: []entity.ScopedPermission{
			{PermissionCode: "iam.users.read", Scope: entity.GlobalScope()},
		},
		CheckedAt: clock.Now(),
	}

	resAdmin, err := svc.RefreshAccessSnapshot(context.Background(), empIAMAdminID)
	if err != nil {
		t.Fatalf("RefreshAccessSnapshot(empIAMAdmin) error = %v", err)
	}
	if !resAdmin.CanAccessIdentity {
		t.Fatal("user holding internal iam.users.read permission must have CanAccessIdentity = true")
	}
}
