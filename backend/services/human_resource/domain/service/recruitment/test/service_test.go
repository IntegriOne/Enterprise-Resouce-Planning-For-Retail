package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/recruitment"
	f "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func TestHiringIdentity(t *testing.T) {
	t.Parallel()
	worker := f.Worker()
	at := f.Date()
	actor := f.ID(99)
	candidate := entity.Candidate{ID: f.ID(2), CandidateNumber: "C-1", PersonPartyReferenceID: &worker.PersonPartyReferenceID, PrivacyNoticeVersion: "v1", CandidateStatus: "active", CreatedAt: at}
	requisition := entity.Requisition{ID: f.ID(3), RequisitionNumber: "R-1", JobID: f.ID(4), OrganizationUnitReferenceID: f.ID(5), HiringManagerPrincipalID: actor, RequisitionType: "replacement", Status: "open", Headcount: 1, EmploymentCategory: entity.WorkerCategoryEmployee, MinimumQualifications: entity.JSON(`{}`), DesiredQualifications: entity.JSON(`{}`), IdempotencyKey: "R-1"}
	application := entity.Application{ID: f.ID(6), CandidateID: candidate.ID, RequisitionID: requisition.ID, ApplicationNumber: "A-1", Status: entity.ApplicationStatusHired, AppliedAt: at, DecisionAt: &at, DecisionByPrincipalID: &actor, HiredWorkerID: &worker.ID, IdempotencyKey: "A-1"}
	if err := service.ValidateHiring(&application, &candidate, &requisition, &worker); err != nil {
		t.Fatal(err)
	}
	wrong := f.ID(999)
	candidate.PersonPartyReferenceID = &wrong
	if err := service.ValidateHiring(&application, &candidate, &requisition, &worker); err == nil {
		t.Fatal("candidate person mismatch accepted")
	}
}
