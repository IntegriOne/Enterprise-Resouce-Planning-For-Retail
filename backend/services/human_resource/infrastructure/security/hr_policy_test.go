package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type polStub struct {
	byID map[string]entity.Policy
	rows []entity.Policy
}

func (s *polStub) FindByIDForUpdate(_ context.Context, id string) (*entity.Policy, error) {
	v, ok := s.byID[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copyVal := v
	return &copyVal, nil
}

func (s *polStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.Policy, error) {
	var out []entity.Policy
	for _, row := range s.rows {
		match := true
		for _, f := range filters {
			switch f.Column {
			case "legal_entity_reference_id":
				if row.LegalEntityReferenceID != f.Values[0] {
					match = false
				}
			case "policy_code":
				if row.PolicyCode != f.Values[0] {
					match = false
				}
			case "policy_version":
				if row.PolicyVersion != f.Values[0] {
					match = false
				}
			}
		}
		if match {
			out = append(out, row)
		}
	}
	return out, nil
}

type polAckStub struct {
	rows []entity.PolicyAcknowledgement
}

func (s *polAckStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.PolicyAcknowledgement, error) {
	var out []entity.PolicyAcknowledgement
	for _, row := range s.rows {
		match := true
		for _, f := range filters {
			switch f.Column {
			case "policy_id":
				if row.PolicyID != f.Values[0] {
					match = false
				}
			case "worker_id":
				if row.WorkerID != f.Values[0] {
					match = false
				}
			}
		}
		if match {
			out = append(out, row)
		}
	}
	return out, nil
}

func TestPolicyAndAcknowledgementValidators(t *testing.T) {
	t.Parallel()
	admin := "00000000-0000-4000-8000-000000000301"
	workerPrincipal := "00000000-0000-4000-8000-000000000302"
	otherPrincipal := "00000000-0000-4000-8000-000000000303"
	leID := "00000000-0000-4000-8000-000000000304"

	polRepo := &polStub{byID: map[string]entity.Policy{}}
	ackRepo := &polAckStub{}
	foundation := &erFoundationStub{}

	polVal := &policyValidator{policies: polRepo, foundation: foundation}
	ackVal := &policyAcknowledgementValidator{
		policies: polRepo,
		workers: &lockedWorkerStub{
			worker: &entity.Worker{
				ID:                            "worker-1",
				Status:                        entity.WorkerStatusActive,
				PrimaryLegalEntityReferenceID: leID,
				IAMPrincipalReferenceID:       &workerPrincipal,
			},
			order: &[]string{},
		},
		acks: ackRepo,
	}

	adminCtx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: admin, Authenticated: true})
	now := time.Date(2027, 1, 10, 9, 0, 0, 0, time.UTC)
	pol := &entity.Policy{
		ID: "pol-1", PolicyCode: "POL-01", PolicyVersion: 1, CategoryCode: "conduct",
		LegalEntityReferenceID: leID, EffectiveFrom: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		ApprovedByPrincipalID: admin, ApprovedAt: now,
	}
	if err := polVal.ValidateMutation(adminCtx, "hr_policies", nil, pol); err != nil {
		t.Fatalf("create valid policy: %v", err)
	}
	if err := polVal.ValidateMutation(adminCtx, "hr_policies", pol, pol); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("published policy update must be frozen: %v", err)
	}

	polRepo.byID["pol-1"] = *pol
	ack := &entity.PolicyAcknowledgement{ID: "ack-1", PolicyID: "pol-1", WorkerID: "worker-1", AssignedAt: now}
	if err := ackVal.ValidateMutation(adminCtx, "hr_policy_acknowledgements", nil, ack); err != nil {
		t.Fatalf("assign policy acknowledgement: %v", err)
	}

	earlyAssignment := *ack
	earlyAssignment.AssignedAt = pol.ApprovedAt.Add(-time.Second)
	if err := ackVal.ValidateMutation(adminCtx, "hr_policy_acknowledgements", nil, &earlyAssignment); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("assign policy before approval: %v", err)
	}

	// Hanya worker bersangkutan yang boleh menyelesaikan acknowledge.
	acked := *ack
	ackTime := now.Add(time.Hour)
	acked.AcknowledgedAt = &ackTime
	acked.AcknowledgedByPrincipalID = &otherPrincipal
	otherCtx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: otherPrincipal, Authenticated: true})
	if err := ackVal.ValidateMutation(otherCtx, "hr_policy_acknowledgements", ack, &acked); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("non-owner acknowledge must be forbidden: %v", err)
	}
	acked.AcknowledgedByPrincipalID = &workerPrincipal
	workerCtx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: workerPrincipal, Authenticated: true})
	if err := ackVal.ValidateMutation(workerCtx, "hr_policy_acknowledgements", ack, &acked); err != nil {
		t.Fatalf("worker owner acknowledge failed: %v", err)
	}
	if err := ackVal.ValidateMutation(workerCtx, "hr_policy_acknowledgements", &acked, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete acknowledged row must fail: %v", err)
	}
}
