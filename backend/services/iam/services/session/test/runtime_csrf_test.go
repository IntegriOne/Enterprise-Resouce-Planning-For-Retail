package session_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
)

func TestSessionService_DeterministicMultiTabCSRFAndRefreshSnapshot(t *testing.T) {
	t.Parallel()

	clock := &mutableClock{now: time.Now()}
	idGen := crypto.NewUUIDGenerator()
	tokGen := crypto.NewOpaqueTokenGenerator()
	csrfDeriver, err := crypto.NewHMACCSRFDeriver(1, map[int][]byte{
		1: bytes.Repeat([]byte{0x44}, 32),
	}, false)
	if err != nil {
		t.Fatalf("NewHMACCSRFDeriver() error = %v", err)
	}

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
		CSRFDeriver:     csrfDeriver,
		SessionPolicy:   domainservice.NewDefaultSessionPolicy(),
		AccessEvaluator: domainservice.NewAccessEvaluator(domainservice.NewDefaultSessionPolicy()),
		Redirect:        domainservice.NewDefaultRedirectResolver(),
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	adminID, _ := idGen.NewUUID()
	adminUser, _ := entity.NewUser(adminID, entity.UserKindEmployee, entity.UserStatusActive, "Super Administrator", clock.Now())
	userRepo.users[adminID] = adminUser
	roleIdentity, _ := entity.NewRole("11111111-1111-4111-8111-111111111111", "superadmin", "Super Admin", "All", "iam", "/identity", 1, true, clock.Now())
	userRepo.snapshots[adminID] = &entity.AccessSnapshot{
		UserID:       adminID,
		UserKind:     entity.UserKindEmployee,
		UserStatus:   entity.UserStatusActive,
		AuthEpoch:    1,
		PermRevision: 1,
		Roles:        []entity.Role{*roleIdentity},
		Permissions: []entity.ScopedPermission{
			{PermissionCode: "iam.dashboard.view", Scope: entity.Scope{Type: entity.ScopeTypeGlobal}},
		},
		CheckedAt: clock.Now(),
	}

	issued, err := svc.Issue(context.Background(), session.IssueInput{
		UserID:      adminID,
		MFAVerified: false,
		IPAddress:   "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	// Tab kedua melakukan sinkronisasi CSRF pada generasi yang sama dan memperoleh token deterministik yang sama.
	tab2CSRF, err := svc.GetSessionCSRF(context.Background(), issued.Session.ID)
	if err != nil {
		t.Fatalf("GetSessionCSRF(tab2) error = %v", err)
	}
	if tab2CSRF.CSRFToken != issued.CSRFToken {
		t.Fatalf("expected deterministic multi-tab CSRF token match, got %q != %q", tab2CSRF.CSRFToken, issued.CSRFToken)
	}

	// MarkMFAVerified merotasi generation 1 -> 2 sehingga token CSRF lama tidak lagi berlaku.
	rotated, err := svc.MarkMFAVerified(context.Background(), issued.Session.ID)
	if err != nil {
		t.Fatalf("MarkMFAVerified() error = %v", err)
	}
	if rotated.Session.Generation != 2 || rotated.CSRFToken == issued.CSRFToken {
		t.Fatalf("expected generation=2 and new CSRF token after MFA step-up, got gen=%d", rotated.Session.Generation)
	}
	if _, err := svc.VerifyCSRF(context.Background(), issued.Session.ID, issued.CSRFToken); err == nil {
		t.Fatal("expected old generation-1 CSRF token to be rejected")
	}

	reSynced, err := svc.GetSessionCSRF(context.Background(), issued.Session.ID)
	if err != nil || reSynced.CSRFToken != rotated.CSRFToken {
		t.Fatalf("expected GetSessionCSRF to return rotated gen-2 token, got (%+v, %v)", reSynced, err)
	}

	// Evaluasi snapshot admin (/identity) vs customer (/).
	resAdmin, err := svc.RefreshAccessSnapshot(context.Background(), adminID)
	if err != nil {
		t.Fatalf("RefreshAccessSnapshot(admin) error = %v", err)
	}
	if !resAdmin.CanAccessIdentity || resAdmin.DefaultRedirectPath != "/identity" {
		t.Fatalf("unexpected admin identity access result: %+v", resAdmin)
	}

	custID, _ := idGen.NewUUID()
	custUser, _ := entity.NewUser(custID, entity.UserKindCustomer, entity.UserStatusActive, "Pelanggan", clock.Now())
	userRepo.users[custID] = custUser
	userRepo.snapshots[custID] = &entity.AccessSnapshot{
		UserID:     custID,
		UserKind:   entity.UserKindCustomer,
		UserStatus: entity.UserStatusActive,
		CheckedAt:  clock.Now(),
	}
	resCust, err := svc.RefreshAccessSnapshot(context.Background(), custID)
	if err != nil {
		t.Fatalf("RefreshAccessSnapshot(customer) error = %v", err)
	}
	if resCust.CanAccessIdentity || resCust.DefaultRedirectPath != "/" {
		t.Fatalf("customer must NOT have identity dashboard access: %+v", resCust)
	}

	// Akun otomasi system wajib ditolak saat penerbitan sesi interaktif dan tidak pernah mendapat akses /identity.
	sysID, _ := idGen.NewUUID()
	sysUser, _ := entity.NewUser(sysID, entity.UserKindSystem, entity.UserStatusActive, "Auto System IAM", clock.Now())
	userRepo.users[sysID] = sysUser
	userRepo.snapshots[sysID] = &entity.AccessSnapshot{
		UserID:     sysID,
		UserKind:   entity.UserKindSystem,
		UserStatus: entity.UserStatusActive,
		Roles:      []entity.Role{*roleIdentity},
		Permissions: []entity.ScopedPermission{
			{PermissionCode: "iam.dashboard.view", Scope: entity.Scope{Type: entity.ScopeTypeGlobal}},
		},
		CheckedAt: clock.Now(),
	}
	if _, err := svc.Issue(context.Background(), session.IssueInput{
		UserID:    sysID,
		IPAddress: "127.0.0.1",
	}); !errors.Is(err, iamerrors.ErrSystemCannotLogin) {
		t.Fatalf("Issue(system) error = %v, want ErrSystemCannotLogin", err)
	}
	resSys, err := svc.RefreshAccessSnapshot(context.Background(), sysID)
	if err != nil {
		t.Fatalf("RefreshAccessSnapshot(system) error = %v", err)
	}
	if resSys.CanAccessIdentity || resSys.DefaultRedirectPath != "/" {
		t.Fatalf("system account must NOT have identity dashboard access: %+v", resSys)
	}
}
