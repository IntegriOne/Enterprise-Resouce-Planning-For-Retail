package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type emergencyContactStub struct {
	matches []entity.EmergencyContact
}

func (s *emergencyContactStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.EmergencyContact, error) {
	return s.matches, nil
}

func TestPersonnelRecordValidators(t *testing.T) {
	t.Parallel()
	workerID := "00000000-0000-4000-8000-000000000201"
	personPartyID := "00000000-0000-4000-8000-000000000202"
	contactPartyID := "00000000-0000-4000-8000-000000000203"
	docID := "00000000-0000-4000-8000-000000000204"
	refID := "00000000-0000-4000-8000-000000000205"

	worker := &entity.Worker{
		ID: workerID, PersonPartyReferenceID: personPartyID,
		Status: entity.WorkerStatusActive,
	}
	workers := &onboardingWorkerStub{value: worker}
	foundation := &foundationVerifierStub{}

	dv := &workerDocumentValidator{workers: workers}
	doc := &entity.WorkerDocument{
		ID: docID, WorkerID: workerID, DocumentType: "id_card",
		DocumentReferenceID: refID, DocumentVersionReferenceID: refID,
		ContentDigest: []byte{1}, VerificationStatus: "pending",
		RetentionClass: "standard", AccessClass: "restricted",
		CreatedAt: time.Now(),
	}
	if err := dv.ValidateMutation(t.Context(), "hr_worker_documents", nil, doc); err != nil {
		t.Fatalf("pending worker document rejected: %v", err)
	}
	verified := *doc
	verified.VerificationStatus = "verified"
	if err := dv.ValidateMutation(t.Context(), "hr_worker_documents", doc, &verified); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("direct verification accepted via CRUD: %v", err)
	}

	cv := &emergencyContactValidator{
		workers: workers, contacts: &emergencyContactStub{}, foundation: foundation,
	}
	contact := &entity.EmergencyContact{
		ID: docID, WorkerID: workerID, ContactPartyReferenceID: contactPartyID,
		RelationshipCode: "spouse", PriorityOrder: 1, IsActive: true,
	}
	if err := cv.ValidateMutation(t.Context(), "hr_emergency_contacts", nil, contact); err != nil {
		t.Fatalf("valid emergency contact rejected: %v", err)
	}
	selfContact := *contact
	selfContact.ContactPartyReferenceID = personPartyID
	if err := cv.ValidateMutation(t.Context(), "hr_emergency_contacts", nil, &selfContact); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("self emergency contact accepted: %v", err)
	}
	foundation.partyErr = hrerrors.ErrReferenceConflict
	if err := cv.ValidateMutation(t.Context(), "hr_emergency_contacts", nil, contact); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("invalid party emergency contact accepted: %v", err)
	}
}
