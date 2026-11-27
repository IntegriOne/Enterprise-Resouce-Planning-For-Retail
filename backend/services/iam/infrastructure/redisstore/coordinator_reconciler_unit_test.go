package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/redis/go-redis/v9"
)

type stubClock struct{ now time.Time }

func (c stubClock) Now() time.Time { return c.now }

type stubSecurityRuntimeRepo struct {
	domainrepo.SecurityRuntimeRepository
	epochErr     error
	epoch        int64
	retryCalled  bool
	retryEventID string
}

func (s *stubSecurityRuntimeRepo) GetRecoveryEpoch(context.Context) (int64, error) {
	if s.epochErr != nil {
		return 0, s.epochErr
	}
	return s.epoch, nil
}

func (s *stubSecurityRuntimeRepo) MarkRetry(_ context.Context, id string, _ int, _ time.Time, _ string) error {
	s.retryCalled = true
	s.retryEventID = id
	return nil
}

func (s *stubSecurityRuntimeRepo) SelectDuePending(context.Context, time.Time, int) ([]*entity.OutboxEvent, error) {
	panic("simulated reconciler panic")
}

func (s *stubSecurityRuntimeRepo) IncrementRecoveryEpoch(context.Context) (int64, error) {
	return 1, nil
}

type stubUserRepo struct {
	domainrepo.UserRepository
	err error
}

func (u *stubUserRepo) SelectByID(context.Context, string) (*entity.User, error) {
	if u.err != nil {
		return nil, u.err
	}
	return nil, iamerrors.ErrUserNotFound
}

type stubSessionRepo struct {
	domainrepo.SessionRepository
	err error
}

func (s *stubSessionRepo) SelectByID(context.Context, string) (*entity.Session, error) {
	if s.err != nil {
		return nil, s.err
	}
	return nil, iamerrors.ErrSessionNotFound
}

func newDummyEnabledProjector() *Projector {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	store, _ := authsession.NewStore(client, authsession.NewKeyspace("unit_test"))
	return NewProjector(store, stubClock{now: time.Now()})
}

func TestCoordinatorCompleteAfterCommitFailsClosedOnRecoveryEpochError(t *testing.T) {
	t.Parallel()

	rt := &stubSecurityRuntimeRepo{epochErr: errors.New("pg timeout")}
	coord := NewCoordinator(rt, newDummyEnabledProjector(), nil, stubClock{now: time.Now()}, nil)

	projectCalled := false
	coord.CompleteAfterCommit(context.Background(), "ev-1", func(int64) error {
		projectCalled = true
		return nil
	})
	if projectCalled {
		t.Fatal("projectFn must not run when GetRecoveryEpoch fails")
	}
	if !rt.retryCalled || rt.retryEventID != "ev-1" {
		t.Fatalf("expected MarkRetry for ev-1, got called=%v id=%q", rt.retryCalled, rt.retryEventID)
	}
	select {
	case <-coord.NotifyChannel():
	default:
		t.Fatal("expected reconciler notify channel to be signaled")
	}
}

func TestReconcilerStepOnceRecoversPanicsAndApplyEventFailsClosed(t *testing.T) {
	t.Parallel()

	rt := &stubSecurityRuntimeRepo{epoch: 1}
	rec := NewReconciler(rt, &stubUserRepo{err: errors.New("transient pg failure")}, &stubSessionRepo{}, newDummyEnabledProjector(), nil, stubClock{now: time.Now()}, nil)

	// stepOnce harus pulih tanpa menjatuhkan goroutine worker saat terjadi panic di repository.
	rec.stepOnce(context.Background())

	if err := rec.applyEvent(context.Background(), &entity.OutboxEvent{
		EventType: entity.OutboxEventSessionRevoked,
		Payload:   json.RawMessage(`{invalid-json`),
	}, 1); err == nil {
		t.Fatal("expected corrupted outbox JSON payload to fail closed")
	}

	badHexPayload, _ := json.Marshal(OutboxPayload{TokenDigestsHex: []string{"zzzz"}})
	if err := rec.applyEvent(context.Background(), &entity.OutboxEvent{
		EventType: entity.OutboxEventSessionRevoked,
		Payload:   badHexPayload,
	}, 1); err == nil {
		t.Fatal("expected malformed hex token digest to fail closed")
	}

	userErrPayload, _ := json.Marshal(OutboxPayload{UserID: "11111111-1111-4111-8111-111111111111"})
	if err := rec.applyEvent(context.Background(), &entity.OutboxEvent{
		EventType: entity.OutboxEventSessionRevoked,
		Payload:   userErrPayload,
	}, 1); err == nil {
		t.Fatal("expected transient user DB error on SessionRevoked to fail closed")
	}

	if err := rec.applyEvent(context.Background(), &entity.OutboxEvent{
		EventType: "iam.unknown.event",
	}, 1); err == nil {
		t.Fatal("expected unknown outbox event type to fail closed")
	}
}
