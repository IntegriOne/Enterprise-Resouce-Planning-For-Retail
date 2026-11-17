// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func setup(tb testing.TB, verifier ...service.RequisitionApprovalVerifier) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.Requisition, string]) {
	tb.Helper()
	h := mockdb.New()
	pRequisition := mockdb.NewStore[entity.Requisition, string](h, func(v *entity.Requisition) string { return v.ID })
	pCandidate := mockdb.NewStore[entity.Candidate, string](h, func(v *entity.Candidate) string { return v.ID })
	pApplication := mockdb.NewStore[entity.Application, string](h, func(v *entity.Application) string { return v.ID })
	pApplicationStageHistory := mockdb.NewStore[entity.ApplicationStageHistory, int64](h, func(v *entity.ApplicationStageHistory) int64 { return v.ID })
	pInterview := mockdb.NewStore[entity.Interview, string](h, func(v *entity.Interview) string { return v.ID })
	pInterviewPanelist := mockdb.NewStore[entity.InterviewPanelist, string](h, func(v *entity.InterviewPanelist) string { return v.ID })
	pInterviewFeedback := mockdb.NewStore[entity.InterviewFeedback, string](h, func(v *entity.InterviewFeedback) string { return v.ID })
	pOffer := mockdb.NewStore[entity.Offer, string](h, func(v *entity.Offer) string { return v.ID })
	pJobPosting := mockdb.NewStore[entity.JobPosting, string](h, func(v *entity.JobPosting) string { return v.ID })
	pApplicationAssessment := mockdb.NewStore[entity.ApplicationAssessment, string](h, func(v *entity.ApplicationAssessment) string { return v.ID })
	deps := service.Dependencies{Runtime: h.Runtime(), Requisition: pRequisition, Candidate: pCandidate, Application: pApplication, ApplicationStageHistory: pApplicationStageHistory, Interview: pInterview, InterviewPanelist: pInterviewPanelist, InterviewFeedback: pInterviewFeedback, Offer: pOffer, JobPosting: pJobPosting, ApplicationAssessment: pApplicationAssessment}
	if len(verifier) != 0 {
		deps.ApprovalVerifier = verifier[0]
	}
	svc, err := service.NewService(deps)
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pRequisition
}
func input() service.CreateRequisitionInput {
	return service.CreateRequisitionInput{RequisitionNumber: "TEST", JobID: mockdb.ID(20), OrganizationUnitReferenceID: mockdb.ID(20), HiringManagerPrincipalID: mockdb.ID(20), RequisitionType: "TEST", EmploymentCategory: "employee", IdempotencyKey: "TEST"}
}
