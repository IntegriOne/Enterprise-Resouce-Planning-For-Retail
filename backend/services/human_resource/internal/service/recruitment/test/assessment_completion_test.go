package service_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

type assessmentEvidenceVerifierStub struct {
	request service.AssessmentEvidenceRequest
	proof   service.AssessmentEvidenceProof
	err     error
}

func (v *assessmentEvidenceVerifierStub) VerifyAssessmentEvidence(_ context.Context, req service.AssessmentEvidenceRequest) (service.AssessmentEvidenceProof, error) {
	v.request = req
	return v.proof, v.err
}

func TestCompleteApplicationAssessment(t *testing.T) {
	t.Parallel()
	h := mockdb.New()
	h.Actor = operation.Actor{PrincipalID: "00000000-0000-4000-8000-000000000001", SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}
	h.Scope = operation.Scope{All: true}
	assessments := mockdb.NewStore[entity.ApplicationAssessment, string](h, func(v *entity.ApplicationAssessment) string { return v.ID })
	evidenceVerifier := &assessmentEvidenceVerifierStub{}
	svc, err := service.NewService(service.Dependencies{
		Runtime:                    operation.Dependencies{Transactor: h, Outbox: h, Policy: h, Integrity: h},
		AssessmentEvidenceVerifier: evidenceVerifier,
		Requisition:                mockdb.NewStore[entity.Requisition, string](h, func(v *entity.Requisition) string { return v.ID }),
		Candidate:                  mockdb.NewStore[entity.Candidate, string](h, func(v *entity.Candidate) string { return v.ID }),
		Application:                mockdb.NewStore[entity.Application, string](h, func(v *entity.Application) string { return v.ID }),
		ApplicationStageHistory:    mockdb.NewStore[entity.ApplicationStageHistory, int64](h, func(v *entity.ApplicationStageHistory) int64 { return v.ID }),
		Interview:                  mockdb.NewStore[entity.Interview, string](h, func(v *entity.Interview) string { return v.ID }),
		InterviewPanelist:          mockdb.NewStore[entity.InterviewPanelist, string](h, func(v *entity.InterviewPanelist) string { return v.ID }),
		InterviewFeedback:          mockdb.NewStore[entity.InterviewFeedback, string](h, func(v *entity.InterviewFeedback) string { return v.ID }),
		Offer:                      mockdb.NewStore[entity.Offer, string](h, func(v *entity.Offer) string { return v.ID }),
		JobPosting:                 mockdb.NewStore[entity.JobPosting, string](h, func(v *entity.JobPosting) string { return v.ID }),
		ApplicationAssessment:      assessments,
	})
	if err != nil {
		t.Fatal(err)
	}
	rubricID := "00000000-0000-4000-8000-000000000301"
	appID := "00000000-0000-4000-8000-000000000302"
	attempt := int32(1)
	created, err := svc.CreateApplicationAssessment(t.Context(), service.CreateApplicationAssessmentInput{
		ApplicationID:                    appID,
		AssessmentCode:                   "TECH-ASSESS-1",
		AttemptNumber:                    &attempt,
		RubricDocumentVersionReferenceID: rubricID,
		RubricDigest:                     bytes.Repeat([]byte{0x22}, 32),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Menolak penyelesaian oleh assessor yang berbeda.
	otherAssessor := "00000000-0000-4000-8000-000000000099"
	assigned := created.ID
	row := assessments.Rows[assigned]
	row.AssessorPrincipalID = &otherAssessor
	assessments.Rows[assigned] = row
	score := entity.Decimal("91.250")
	rec := "PASS"
	if _, err := svc.CompleteApplicationAssessment(t.Context(), created.ID, service.CompleteApplicationAssessmentInput{
		NormalizedScore: &score, RecommendationCode: &rec,
	}); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("different assessor completion = %v, want ErrForbidden", err)
	}
	row.AssessorPrincipalID = nil
	assessments.Rows[assigned] = row

	// Menolak bukti tidak lengkap.
	evidenceID := "00000000-0000-4000-8000-000000000303"
	if _, err := svc.CompleteApplicationAssessment(t.Context(), created.ID, service.CompleteApplicationAssessmentInput{
		EvidenceDocumentVersionReferenceID: &evidenceID,
	}); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("partial evidence = %v, want ErrInvalidInput", err)
	}

	// Berhasil menyelesaikan asesmen dengan bukti DRM terverifikasi.
	evidenceDigest := bytes.Repeat([]byte{0x33}, 32)
	evidenceVerifier.proof = service.AssessmentEvidenceProof{
		VersionID: evidenceID, Digest: evidenceDigest, Verified: true, Immutable: true,
		OwnerManagement: "human_resource", Purpose: "recruitment_assessment_evidence",
		AssessmentID: evidenceID, ApplicationID: created.ApplicationID,
	}
	input := service.CompleteApplicationAssessmentInput{
		NormalizedScore:                    &score,
		RecommendationCode:                 &rec,
		EvidenceDocumentVersionReferenceID: &evidenceID,
		EvidenceDigest:                     evidenceDigest,
	}
	if _, err := svc.CompleteApplicationAssessment(t.Context(), created.ID, input); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("evidence bound to a different assessment: %v", err)
	}
	evidenceVerifier.proof.AssessmentID = created.ID
	evidenceVerifier.proof.ApplicationID = evidenceID
	if _, err := svc.CompleteApplicationAssessment(t.Context(), created.ID, input); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("evidence bound to a different application: %v", err)
	}
	evidenceVerifier.proof.ApplicationID = created.ApplicationID
	completed, err := svc.CompleteApplicationAssessment(t.Context(), created.ID, input)
	if err != nil || completed.Status != "completed" || completed.CompletedAt == nil ||
		completed.AssessorPrincipalID == nil || *completed.AssessorPrincipalID != h.Actor.PrincipalID {
		t.Fatalf("complete application assessment = %+v, err = %v", completed, err)
	}
	if _, err := svc.CompleteApplicationAssessment(t.Context(), created.ID, service.CompleteApplicationAssessmentInput{
		NormalizedScore: &score,
	}); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat completion = %v, want ErrWorkflowRequired", err)
	}
}
