package service_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/payroll"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

type freezeSnapshotStoreStub struct {
	rows       map[string]entity.PayrollInputSnapshot
	captureErr error
	captures   int
}

func (s *freezeSnapshotStoreStub) Capture(_ context.Context, id, periodID, principalID string, at time.Time) (entity.PayrollInputSnapshot, error) {
	s.captures++
	if s.captureErr != nil {
		return entity.PayrollInputSnapshot{}, s.captureErr
	}
	if existing, ok := s.rows[id]; ok {
		return existing, hrerrors.ErrWorkflowRequired
	}
	digest := sha256.Sum256([]byte("snapshot:" + periodID))
	row := entity.PayrollInputSnapshot{
		ID:               id,
		PeriodID:         periodID,
		GroupID:          "11111111-1111-4111-8111-111111111111",
		EmployerID:       "22222222-2222-4222-8222-222222222222",
		JurisdictionCode: "ID",
		CurrencyCode:     "IDR",
		Content:          []byte(`{"sensitive":"salaries"}`),
		Digest:           digest[:],
		CreatedBy:        principalID,
		CreatedAt:        at,
	}
	s.rows[id] = row
	return row, nil
}

func (s *freezeSnapshotStoreStub) SelectMetadata(_ context.Context, id string) (entity.PayrollInputSnapshot, error) {
	row, ok := s.rows[id]
	if !ok {
		return entity.PayrollInputSnapshot{}, hrerrors.ErrNotFound
	}
	row.Content = nil
	return row, nil
}

type freezePolicyStub struct {
	calls int
	first operation.Actor
	next  operation.Actor
	errAt int
	err   error
}

func (p *freezePolicyStub) Authorize(_ context.Context, access operation.Access) (operation.Actor, error) {
	p.calls++
	if access.Resource != "hr_payroll_runs" || access.Action != "create" {
		return operation.Actor{}, hrerrors.ErrForbidden
	}
	if p.errAt > 0 && p.calls == p.errAt {
		return operation.Actor{}, p.err
	}
	if p.calls > 1 && p.next.PrincipalID != "" {
		return p.next, nil
	}
	return p.first, nil
}

func (p *freezePolicyStub) ReadScope(_ context.Context, _ operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestFreezePayrollInput(t *testing.T) {
	const (
		principalID = "00000000-0000-4000-8000-000000000010"
		otherID     = "00000000-0000-4000-8000-000000000011"
		periodID    = "00000000-0000-4000-8000-000000000020"
		idemKey     = "00000000-0000-4000-8000-000000000030"
	)

	t.Run("invalid uuid", func(t *testing.T) {
		svc := mustNewPayrollService(t, &freezePolicyStub{first: operation.Actor{Kind: "human", PrincipalID: principalID}}, &freezeSnapshotStoreStub{rows: map[string]entity.PayrollInputSnapshot{}})
		if _, err := svc.FreezePayrollInput(t.Context(), "bad", idemKey); !errors.Is(err, hrerrors.ErrInvalidInput) {
			t.Fatalf("bad period = %v, want ErrInvalidInput", err)
		}
		if _, err := svc.FreezePayrollInput(t.Context(), periodID, "bad"); !errors.Is(err, hrerrors.ErrInvalidInput) {
			t.Fatalf("bad idempotency key = %v, want ErrInvalidInput", err)
		}
	})

	t.Run("non-human or unauthorized", func(t *testing.T) {
		svc := mustNewPayrollService(t, &freezePolicyStub{first: operation.Actor{Kind: "service", PrincipalID: principalID}}, &freezeSnapshotStoreStub{rows: map[string]entity.PayrollInputSnapshot{}})
		if _, err := svc.FreezePayrollInput(t.Context(), periodID, idemKey); !errors.Is(err, hrerrors.ErrForbidden) {
			t.Fatalf("service actor = %v, want ErrForbidden", err)
		}
	})

	t.Run("capture and idempotent replay omits content", func(t *testing.T) {
		store := &freezeSnapshotStoreStub{rows: map[string]entity.PayrollInputSnapshot{}}
		svc := mustNewPayrollService(t, &freezePolicyStub{first: operation.Actor{Kind: "human", PrincipalID: principalID}}, store)

		first, err := svc.FreezePayrollInput(t.Context(), periodID, idemKey)
		if err != nil {
			t.Fatalf("first freeze: %v", err)
		}
		if first.Content != nil || len(first.Digest) != sha256.Size || store.captures != 1 {
			t.Fatalf("unexpected first freeze result: %+v captures=%d", first, store.captures)
		}
		second, err := svc.FreezePayrollInput(t.Context(), periodID, idemKey)
		if err != nil {
			t.Fatalf("replay freeze: %v", err)
		}
		if second.ID != first.ID || !bytes.Equal(second.Digest, first.Digest) || store.captures != 1 {
			t.Fatalf("unexpected replay result: %+v vs %+v captures=%d", second, first, store.captures)
		}
	})

	t.Run("reauthorization failure inside transaction aborts", func(t *testing.T) {
		store := &freezeSnapshotStoreStub{rows: map[string]entity.PayrollInputSnapshot{}}
		policy := &freezePolicyStub{
			first: operation.Actor{Kind: "human", PrincipalID: principalID},
			next:  operation.Actor{Kind: "human", PrincipalID: otherID},
		}
		svc := mustNewPayrollService(t, policy, store)
		if _, err := svc.FreezePayrollInput(t.Context(), periodID, idemKey); !errors.Is(err, hrerrors.ErrForbidden) {
			t.Fatalf("actor mismatch on recheck = %v, want ErrForbidden", err)
		}
	})

	t.Run("closed or empty period returns workflow required", func(t *testing.T) {
		store := &freezeSnapshotStoreStub{rows: map[string]entity.PayrollInputSnapshot{}, captureErr: hrerrors.ErrWorkflowRequired}
		svc := mustNewPayrollService(t, &freezePolicyStub{first: operation.Actor{Kind: "human", PrincipalID: principalID}}, store)
		if _, err := svc.FreezePayrollInput(t.Context(), periodID, idemKey); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
			t.Fatalf("closed/empty period = %v, want ErrWorkflowRequired", err)
		}
	})
}

func mustNewPayrollService(t *testing.T, policy operation.Policy, store service.InputSnapshotStore) *service.Service {
	t.Helper()
	h := mockdb.New()
	rt := h.Runtime()
	rt.Policy = policy
	svc, err := service.NewService(service.Dependencies{
		Runtime:           rt,
		PayrollProfile:    mockdb.NewStore[entity.PayrollProfile, string](h, func(v *entity.PayrollProfile) string { return v.ID }),
		PayComponent:      mockdb.NewStore[entity.PayComponent, string](h, func(v *entity.PayComponent) string { return v.ID }),
		PayrollPeriod:     mockdb.NewStore[entity.PayrollPeriod, string](h, func(v *entity.PayrollPeriod) string { return v.ID }),
		PayrollRun:        mockdb.NewStore[entity.PayrollRun, string](h, func(v *entity.PayrollRun) string { return v.ID }),
		PayrollResult:     mockdb.NewStore[entity.PayrollResult, string](h, func(v *entity.PayrollResult) string { return v.ID }),
		PayrollResultLine: mockdb.NewStore[entity.PayrollResultLine, string](h, func(v *entity.PayrollResultLine) string { return v.ID }),
		PayrollAdjustment: mockdb.NewStore[entity.PayrollAdjustment, string](h, func(v *entity.PayrollAdjustment) string { return v.ID }),
		PayrollGroup:      mockdb.NewStore[entity.PayrollGroup, string](h, func(v *entity.PayrollGroup) string { return v.ID }),
		PayrollRuleSet:    mockdb.NewStore[entity.PayrollRuleSet, string](h, func(v *entity.PayrollRuleSet) string { return v.ID }),
		InputSnapshots:    store,
	})
	if err != nil {
		t.Fatalf("new payroll service: %v", err)
	}
	return svc
}
