package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/policy"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func TestAcknowledgePolicyEnforcesWorkerOwnershipAndSingleAck(t *testing.T) {
	h := mockdb.New()
	adminPrincipal := mockdb.ID(10)
	workerPrincipal := mockdb.ID(11)
	otherPrincipal := mockdb.ID(12)
	h.Actor = operation.Actor{PrincipalID: adminPrincipal, SessionID: mockdb.ID(99), Kind: "human"}
	now := mockdb.Now()
	hireDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	policies := mockdb.NewStore[entity.Policy, string](h, func(v *entity.Policy) string { return v.ID })
	acks := mockdb.NewStore[entity.PolicyAcknowledgement, string](h, func(v *entity.PolicyAcknowledgement) string { return v.ID })
	workers := mockdb.NewStore[entity.Worker, string](h, func(v *entity.Worker) string { return v.ID })

	svc, err := service.NewService(service.Dependencies{
		Runtime:               h.Runtime(),
		Policy:                policies,
		PolicyAcknowledgement: acks,
		Worker:                workers,
	})
	if err != nil {
		t.Fatal(err)
	}

	workerID := mockdb.ID(1)
	workers.Rows[workerID] = entity.Worker{
		ID:                            workerID,
		WorkerNumber:                  "W-POL",
		PersonPartyReferenceID:        mockdb.ID(20),
		WorkerCategory:                entity.WorkerCategoryEmployee,
		Status:                        entity.WorkerStatusActive,
		HireDate:                      &hireDate,
		PrimaryLegalEntityReferenceID: mockdb.ID(21),
		IAMPrincipalReferenceID:       &workerPrincipal,
		CreatedAt:                     hireDate,
		UpdatedAt:                     hireDate,
	}

	ctx := context.Background()
	pol, err := svc.CreatePolicy(ctx, service.CreatePolicyInput{
		PolicyCode:                 "POL-SEC-01",
		PolicyVersion:              1,
		PolicyName:                 "Security Handbook",
		CategoryCode:               "conduct",
		LegalEntityReferenceID:     mockdb.ID(21),
		EffectiveFrom:              hireDate,
		DocumentVersionReferenceID: mockdb.ID(30),
		ContentDigest:              []byte("digest"),
	})
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}
	if pol.ApprovedByPrincipalID != adminPrincipal {
		t.Fatalf("approved_by_principal_id = %s, want %s", pol.ApprovedByPrincipalID, adminPrincipal)
	}

	ack, err := svc.CreatePolicyAcknowledgement(ctx, service.CreatePolicyAcknowledgementInput{
		PolicyID: pol.ID,
		WorkerID: workerID,
	})
	if err != nil {
		t.Fatalf("create acknowledgement: %v", err)
	}

	// Pihak lain tidak boleh mengakui atas nama pekerja bersangkutan.
	h.Actor.PrincipalID = otherPrincipal
	evidenceID := mockdb.ID(31)
	if _, err := svc.AcknowledgePolicy(ctx, ack.ID, &evidenceID); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("other principal acknowledge must be forbidden: %v", err)
	}

	// Pekerja yang bersangkutan mengakui kebijakan.
	h.Actor.PrincipalID = workerPrincipal
	acked, err := svc.AcknowledgePolicy(ctx, ack.ID, &evidenceID)
	if err != nil {
		t.Fatalf("worker acknowledge: %v", err)
	}
	if acked.AcknowledgedAt == nil || !acked.AcknowledgedAt.Equal(now) ||
		acked.AcknowledgedByPrincipalID == nil || *acked.AcknowledgedByPrincipalID != workerPrincipal ||
		acked.EvidenceDocumentVersionReferenceID == nil || *acked.EvidenceDocumentVersionReferenceID != evidenceID {
		t.Fatalf("unexpected acknowledged row: %+v", acked)
	}

	// Pengakuan kedua ditolak.
	if _, err := svc.AcknowledgePolicy(ctx, ack.ID, nil); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("second acknowledge must fail: %v", err)
	}
}
