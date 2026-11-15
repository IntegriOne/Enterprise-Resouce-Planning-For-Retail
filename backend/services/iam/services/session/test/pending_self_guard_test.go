package session_test

import (
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

func TestSessionService_IssueRejectsPendingAndNonActiveUsers(t *testing.T) {
	t.Parallel()

	clock := &mutableClock{now: time.Now()}
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
		t.Fatalf("NewService() error = %v", err)
	}

	// Uji penolakan Issue pada PENDING, SUSPENDED, DISABLED, dan soft-deleted.
	cases := []struct {
		name    string
		setup   func(u *entity.User)
		wantErr error
	}{
		{
			name:    "pending",
			setup:   func(*entity.User) {},
			wantErr: iamerrors.ErrUserInactive,
		},
		{
			name: "suspended",
			setup: func(u *entity.User) {
				_ = u.Activate(clock.Now())
				_ = u.Suspend(clock.Now())
			},
			wantErr: iamerrors.ErrUserSuspended,
		},
		{
			name: "disabled",
			setup: func(u *entity.User) {
				_ = u.Disable(clock.Now())
			},
			wantErr: iamerrors.ErrUserDisabled,
		},
		{
			name: "deleted",
			setup: func(u *entity.User) {
				_ = u.Activate(clock.Now())
				_ = u.SoftDelete(clock.Now())
			},
			wantErr: iamerrors.ErrUserDeleted,
		},
	}

	for _, tc := range cases {
		uid, _ := idGen.NewUUID()
		u, err := entity.NewUser(uid, entity.UserKindCustomer, entity.UserStatusPending, "Guard "+tc.name, clock.Now())
		if err != nil {
			t.Fatalf("NewUser(%s) error = %v", tc.name, err)
		}
		tc.setup(u)
		userRepo.users[uid] = u

		if _, err := svc.Issue(context.Background(), session.IssueInput{UserID: uid}); !errors.Is(err, tc.wantErr) {
			t.Fatalf("%s Issue() error = %v, want %v", tc.name, err, tc.wantErr)
		}
	}

	// Akun ACTIVE diizinkan memperoleh sesi baru.
	activeID, _ := idGen.NewUUID()
	activeUser, _ := entity.NewUser(activeID, entity.UserKindCustomer, entity.UserStatusActive, "Active Customer", clock.Now())
	userRepo.users[activeID] = activeUser
	if issued, err := svc.Issue(context.Background(), session.IssueInput{UserID: activeID}); err != nil || issued == nil {
		t.Fatalf("active Issue() = (%+v, %v), want valid session", issued, err)
	}
}
