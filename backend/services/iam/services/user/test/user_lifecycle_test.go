package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
	"github.com/integrione/erp-retail/backend/services/iam/services/user"
)

func TestUserService_CreateUserPreventsDuplicateEmailAndUsername(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)

	created, err := h.userSvc.CreateUser(context.Background(), user.CreateUserInput{
		Kind:                   entity.UserKindEmployee,
		Status:                 entity.UserStatusActive,
		DisplayName:            "Budi Santoso",
		Email:                  "Budi.Santoso@IntegriOne.local",
		Username:               "Budi.Staff",
		Password:               "Valid#Password!2026",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if created.Status != entity.UserStatusActive {
		t.Fatalf("expected active user, got %q", created.Status)
	}

	// Duplikasi email meskipun huruf kapital/spasi berbeda harus ditolak dengan field "email".
	_, dupEmailErr := h.userSvc.CreateUser(context.Background(), user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusActive,
		DisplayName: "Other User",
		Email:       " budi.santoso@integrione.local ",
		Username:    "other.user",
		Password:    "Kx9#AlphaVault!2026",
	})
	if !errors.Is(dupEmailErr, iamerrors.ErrIdentityAlreadyExists) {
		t.Fatalf("expected ErrIdentityAlreadyExists for duplicate email, got %v", dupEmailErr)
	}
	if domErr, ok := errors.AsType[*iamerrors.DomainError](dupEmailErr); !ok || domErr.Field != "email" {
		t.Fatalf("expected field \"email\" on duplicate email error, got %+v", dupEmailErr)
	}

	// Duplikasi username dengan email berbeda juga harus ditolak dengan field "username" dan pesan "username has been taken".
	_, dupUnameErr := h.userSvc.CreateUser(context.Background(), user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusActive,
		DisplayName: "Second Account",
		Email:       "budi2@integrione.local",
		Username:    "budi.staff",
		Password:    "Mz4#BetaVault!2026",
	})
	if !errors.Is(dupUnameErr, iamerrors.ErrIdentityAlreadyExists) {
		t.Fatalf("expected ErrIdentityAlreadyExists for duplicate username, got %v", dupUnameErr)
	}
	if domErr, ok := errors.AsType[*iamerrors.DomainError](dupUnameErr); !ok || domErr.Field != "username" || domErr.Message != "username has been taken" {
		t.Fatalf("expected field \"username\" with \"username has been taken\", got %+v", dupUnameErr)
	}

	// Bila email DAN username sama-sama duplikat, username wajib diprioritaskan agar username terpakai tidak menjadi oracle eksistensi email.
	_, bothDupErr := h.userSvc.CreateUser(context.Background(), user.CreateUserInput{
		Kind:        entity.UserKindCustomer,
		Status:      entity.UserStatusPending,
		DisplayName: "Both Conflict",
		Email:       "budi.santoso@integrione.local",
		Username:    "budi.staff",
		Password:    "Mz4#BetaVault!2026",
	})
	if domErr, ok := errors.AsType[*iamerrors.DomainError](bothDupErr); !ok || domErr.Field != "username" || domErr.Message != "username has been taken" {
		t.Fatalf("expected username priority over email conflict, got %+v", bothDupErr)
	}

	// Kata sandi yang mengandung username/email harus ditolak PasswordPolicy.
	if _, err := h.userSvc.CreateUser(context.Background(), user.CreateUserInput{
		Kind:        entity.UserKindEmployee,
		Status:      entity.UserStatusActive,
		DisplayName: "Test Policy",
		Email:       "policy@integrione.local",
		Username:    "policyuser",
		Password:    "Strong#policyuser!9",
	}); !errors.Is(err, iamerrors.ErrPasswordContainsIdentity) {
		t.Fatalf("expected ErrPasswordContainsIdentity, got %v", err)
	}
}

func TestUserService_LifecycleRevokesSessionsOnSuspendDisableDelete(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	u, err := h.userSvc.CreateUser(context.Background(), user.CreateUserInput{
		Kind:                   entity.UserKindEmployee,
		Status:                 entity.UserStatusActive,
		DisplayName:            "Sari Dewi",
		Email:                  "sari@integrione.local",
		Username:               "sari.dewi",
		Password:               "Pq7#GreenShield!2026",
		MarkIdentitiesVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	loginRes, err := h.userSvc.AuthenticatePassword(context.Background(), user.AuthenticatePasswordInput{
		Identifier: "sari@integrione.local",
		Password:   "Pq7#GreenShield!2026",
		IPAddress:  "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("AuthenticatePassword() error = %v", err)
	}

	// Tangguhkan user -> sesi aktif otomatis dicabut.
	if _, err := h.userSvc.Suspend(context.Background(), u.ID); err != nil {
		t.Fatalf("Suspend() error = %v", err)
	}
	if _, err := h.sessSvc.Authenticate(context.Background(), loginRes.IssuedSession.SessionToken); !errors.Is(err, iamerrors.ErrSessionRevoked) {
		t.Fatalf("expected ErrSessionRevoked after user suspend, got %v", err)
	}

	// Aktifkan kembali -> bisa memperbarui profil dan hapus lunak.
	if _, err := h.userSvc.Activate(context.Background(), u.ID); err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	updated, err := h.userSvc.UpdateProfile(context.Background(), user.UpdateProfileInput{
		UserID:      u.ID,
		DisplayName: "Sari Dewi Wijaya",
		Locale:      "id-ID",
		Timezone:    "Asia/Jakarta",
	})
	if err != nil || updated.DisplayName != "Sari Dewi Wijaya" {
		t.Fatalf("UpdateProfile() = (%+v, %v)", updated, err)
	}

	loginInput := user.AuthenticatePasswordInput{
		Identifier: "sari@integrione.local",
		Password:   "Pq7#GreenShield!2026",
		IPAddress:  "127.0.0.1",
	}
	active, err := h.userSvc.AuthenticatePassword(context.Background(), loginInput)
	if err != nil {
		t.Fatalf("login after activation: %v", err)
	}
	disabled, err := h.userSvc.Disable(context.Background(), u.ID)
	if err != nil || disabled == nil || disabled.Status != entity.UserStatusDisabled {
		t.Fatalf("disable employee: %+v, %v", disabled, err)
	}
	if _, err := h.sessSvc.Authenticate(context.Background(), active.IssuedSession.SessionToken); !errors.Is(err, iamerrors.ErrSessionRevoked) {
		t.Fatalf("expected ErrSessionRevoked after disabling employee, got %v", err)
	}
	if result, err := h.userSvc.AuthenticatePassword(context.Background(), loginInput); result != nil || !errors.Is(err, iamerrors.ErrInvalidCredentials) {
		t.Fatalf("expected new login to fail after disabling employee: result returned=%t, %v", result != nil, err)
	}

	if issued, err := h.sessSvc.Issue(context.Background(), session.IssueInput{UserID: u.ID}); issued != nil || !errors.Is(err, iamerrors.ErrUserDisabled) {
		t.Fatalf("expected session issuance to fail after disabling employee: session returned=%t, %v", issued != nil, err)
	}

	if err := h.userSvc.SoftDelete(context.Background(), u.ID); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}
	page, err := h.userSvc.List(context.Background(), domainrepo.UserFilter{})
	if err != nil || page.Total != 0 {
		t.Fatalf("expected soft-deleted user excluded from default List, got %+v err=%v", page, err)
	}
}

func TestUserService_CreateUserAutoAssignsBaselineSelfRoleForAllUserKinds(t *testing.T) {
	t.Parallel()

	h := newTestHarness(t)
	now := h.clock.Now()

	custRoleID, _ := h.idGen.NewUUID()
	selfRoleID, _ := h.idGen.NewUUID()
	custRole, err := entity.NewRole(custRoleID, "customer", "Web Commerce Customer", "Default customer", "commerce", "/", 100, true, now)
	if err != nil {
		t.Fatalf("NewRole(customer) error = %v", err)
	}
	selfRole, err := entity.NewRole(selfRoleID, "iam_selfservice", "IAM Account Self-Service", "Default self role", "iam", "", 999, true, now)
	if err != nil {
		t.Fatalf("NewRole(iam_selfservice) error = %v", err)
	}

	h.store.mu.Lock()
	h.store.roleDefs[custRole.ID] = custRole
	h.store.roleDefs[selfRole.ID] = selfRole
	h.store.mu.Unlock()

	cases := []struct {
		kind         entity.UserKind
		wantRoleCode string
		wantRoleID   string
		email        string
		username     string
	}{
		{kind: entity.UserKindCustomer, wantRoleCode: "customer", wantRoleID: custRole.ID, email: "cust.auto@integrione.local", username: "cust.auto"},
		{kind: entity.UserKindEmployee, wantRoleCode: "iam_selfservice", wantRoleID: selfRole.ID, email: "emp.auto@integrione.local", username: "emp.auto"},
		{kind: entity.UserKindVendor, wantRoleCode: "iam_selfservice", wantRoleID: selfRole.ID, email: "vnd.auto@integrione.local", username: "vnd.auto"},
		{kind: entity.UserKindSupplier, wantRoleCode: "iam_selfservice", wantRoleID: selfRole.ID, email: "sup.auto@integrione.local", username: "sup.auto"},
	}

	for _, tc := range cases {
		created, err := h.userSvc.CreateUser(context.Background(), user.CreateUserInput{
			Kind:        tc.kind,
			Status:      entity.UserStatusPending,
			DisplayName: "Basename " + tc.username,
			Email:       tc.email,
			Username:    tc.username,
			Password:    "V7$K9q#R2w!X8mZp",
		})
		if err != nil {
			t.Fatalf("CreateUser(%s) error = %v", tc.kind, err)
		}

		h.store.mu.Lock()
		assignments := append([]*entity.UserRole(nil), h.store.userRoles[created.ID]...)
		h.store.mu.Unlock()

		if len(assignments) != 1 {
			t.Fatalf("CreateUser(%s) userRoles len = %d, want 1", tc.kind, len(assignments))
		}
		if assignments[0].RoleID != tc.wantRoleID {
			t.Fatalf("CreateUser(%s) role_id = %s, want %s (%s)", tc.kind, assignments[0].RoleID, tc.wantRoleID, tc.wantRoleCode)
		}
		if assignments[0].Scope.Type != entity.ScopeTypeSelf || assignments[0].Scope.ID == nil || *assignments[0].Scope.ID != created.ID {
			t.Fatalf("CreateUser(%s) scope = %+v, want ScopeSelf(%s)", tc.kind, assignments[0].Scope, created.ID)
		}
	}
}
