package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/contract/provisioning"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
)

func TestWorkerLinkVerifierBindsProvisioningReferences(t *testing.T) {
	verifier := workerLinkVerifier{key: []byte(strings.Repeat("a", 32))}
	check := personnel.PrincipalLinkCheck{
		WorkerID: "10000000-0000-4000-8000-000000000001", PersonPartyID: "10000000-0000-4000-8000-000000000002",
		PrincipalID: "10000000-0000-4000-8000-000000000003", RequestedByPrincipalID: "10000000-0000-4000-8000-000000000004",
	}
	var err error
	check.EvidenceID, err = provisioning.WorkerLinkEvidence(verifier.key, check.WorkerID, check.PersonPartyID, check.PrincipalID, check.RequestedByPrincipalID)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifier.VerifyPrincipalLink(t.Context(), check); err != nil {
		t.Fatal(err)
	}
	check.PrincipalID = check.WorkerID
	if err := verifier.VerifyPrincipalLink(t.Context(), check); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("evidence for another account must fail, got %v", err)
	}
}
