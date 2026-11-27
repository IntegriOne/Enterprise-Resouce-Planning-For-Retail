package provisioning_test

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/services/human_resource/contract/provisioning"
)

func TestWorkerLinkEvidenceRejectsWeakKeyAndEmptyIdentity(t *testing.T) {
	const worker = "10000000-0000-4000-8000-000000000001"
	const party = "10000000-0000-4000-8000-000000000002"
	const principal = "10000000-0000-4000-8000-000000000003"
	const actor = "10000000-0000-4000-8000-000000000004"
	key := bytes.Repeat([]byte{0x61}, 32)
	if _, err := provisioning.WorkerLinkEvidence(key[:31], worker, party, principal, actor); err == nil {
		t.Fatal("accepted a short key")
	}
	if _, err := provisioning.WorkerLinkEvidence(key, worker, party, uuid.Nil.String(), actor); err == nil {
		t.Fatal("accepted an empty principal")
	}
	first, err := provisioning.WorkerLinkEvidence(key, worker, party, principal, actor)
	if err != nil || first == "" {
		t.Fatalf("sign worker link evidence: %v", err)
	}
	second, err := provisioning.WorkerLinkEvidence(key, worker, party, actor, principal)
	if err != nil || second == first {
		t.Fatalf("worker link evidence must distinguish the principal and actor: %v", err)
	}
}
