package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
	"github.com/integrione/erp-retail/backend/services/iam/pkg/crypto"
	"github.com/integrione/erp-retail/backend/services/iam/services/session"
)

type failingOutboxRuntime struct {
	domainrepo.SecurityRuntimeRepository
	err error
}

type passthroughTransactor struct{}

func (passthroughTransactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestAuthenticateRejectsInactiveSessionOwner(t *testing.T) {
	clock := &mutableClock{now: time.Now()}
	idGen := crypto.NewUUIDGenerator()
	userID, err := idGen.NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	u, err := entity.NewUser(userID, entity.UserKindEmployee, entity.UserStatusActive, "Session owner", clock.Now())
	if err != nil {
		t.Fatal(err)
	}
	users := &mockUserRepo{users: map[string]*entity.User{userID: u}}
	svc, err := session.NewService(session.Dependencies{
		Sessions: newMockSessionRepo(), Users: users, Clock: clock, IDGen: idGen, Tokens: crypto.NewOpaqueTokenGenerator(),
	})
	if err != nil {
		t.Fatal(err)
	}
	issued, err := svc.Issue(context.Background(), session.IssueInput{UserID: userID})
	if err != nil {
		t.Fatal(err)
	}
	users.mu.Lock()
	u.Status = entity.UserStatusSuspended
	users.mu.Unlock()
	if _, err := svc.Authenticate(context.Background(), issued.SessionToken); !errors.Is(err, iamerrors.ErrUnauthorized) {
		t.Fatalf("Authenticate() error = %v, want unauthorized", err)
	}
	if _, err := svc.VerifyCSRF(context.Background(), issued.Session.ID, issued.CSRFToken); !errors.Is(err, iamerrors.ErrUnauthorized) {
		t.Fatalf("VerifyCSRF() error = %v, want unauthorized", err)
	}
	if _, err := svc.MarkMFAVerified(context.Background(), issued.Session.ID); !errors.Is(err, iamerrors.ErrUnauthorized) {
		t.Fatalf("MarkMFAVerified() error = %v, want unauthorized", err)
	}
}

func (r failingOutboxRuntime) Insert(context.Context, *entity.OutboxEvent) error {
	return r.err
}

func (r failingOutboxRuntime) GetRecoveryEpoch(context.Context) (int64, error) {
	return 0, r.err
}

func TestRefreshAccessSnapshotFailsWhenRecoveryEpochUnavailable(t *testing.T) {
	clock := &mutableClock{now: time.Now()}
	idGen := crypto.NewUUIDGenerator()
	userID, err := idGen.NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	u, err := entity.NewUser(userID, entity.UserKindEmployee, entity.UserStatusActive, "Snapshot owner", clock.Now())
	if err != nil {
		t.Fatal(err)
	}
	users := &mockUserRepo{
		users: map[string]*entity.User{userID: u},
		snapshots: map[string]*entity.AccessSnapshot{userID: {
			UserID: userID, UserStatus: entity.UserStatusActive,
		}},
	}
	epochErr := errors.New("recovery epoch unavailable")
	svc, err := session.NewService(session.Dependencies{
		Sessions: newMockSessionRepo(), Users: users, Clock: clock, IDGen: idGen,
		Tokens: crypto.NewOpaqueTokenGenerator(), SecurityRuntime: failingOutboxRuntime{err: epochErr},
		Redirect: domainservice.NewDefaultRedirectResolver(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RefreshAccessSnapshot(context.Background(), userID); !errors.Is(err, epochErr) {
		t.Fatalf("RefreshAccessSnapshot() error = %v, want recovery epoch error", err)
	}
}

func TestSessionMutationFailsWhenOutboxInsertFails(t *testing.T) {
	clock := &mutableClock{now: time.Now()}
	idGen := crypto.NewUUIDGenerator()
	tokens := crypto.NewOpaqueTokenGenerator()
	userID, err := idGen.NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	u, err := entity.NewUser(userID, entity.UserKindEmployee, entity.UserStatusActive, "Session owner", clock.Now())
	if err != nil {
		t.Fatal(err)
	}
	users := &mockUserRepo{users: map[string]*entity.User{userID: u}}
	sessions := newMockSessionRepo()
	base, err := session.NewService(session.Dependencies{Sessions: sessions, Users: users, Clock: clock, IDGen: idGen, Tokens: tokens})
	if err != nil {
		t.Fatal(err)
	}
	issued, err := base.Issue(context.Background(), session.IssueInput{UserID: userID})
	if err != nil {
		t.Fatal(err)
	}

	outboxErr := errors.New("outbox insert failed")
	coordinator := redisstore.NewCoordinator(failingOutboxRuntime{err: outboxErr}, nil, idGen, clock, nil)
	if _, err := session.NewService(session.Dependencies{
		Sessions: sessions, Users: users, Clock: clock, IDGen: idGen, Tokens: tokens, OutboxCoord: coordinator,
	}); err == nil || iamerrors.CodeOf(err) != iamerrors.Code("iam.service.nil_transactor") {
		t.Fatalf("NewService() error = %v, want missing transactor", err)
	}
	withOutbox, err := session.NewService(session.Dependencies{
		Sessions: sessions, Users: users, Clock: clock, IDGen: idGen, Tokens: tokens,
		Transactor: passthroughTransactor{}, OutboxCoord: coordinator,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := withOutbox.Issue(context.Background(), session.IssueInput{UserID: userID}); !errors.Is(err, outboxErr) {
		t.Fatalf("Issue() error = %v, want outbox error", err)
	}
	if err := withOutbox.Logout(context.Background(), issued.SessionToken); !errors.Is(err, outboxErr) {
		t.Fatalf("Logout() error = %v, want outbox error", err)
	}
}
