package human_resource

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	learning "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/learning/v1"
	performance "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/performance/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	policy "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/policy/v1"
	succession "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/succession/v1"
	workforce "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc"
)

type finderPersonnelFixture struct {
	personnel.PersonnelServiceClient
	byStatus string
}

func (f *finderPersonnelFixture) FindWorkerByStatus(_ context.Context, req *common.StringFinderRequest, _ ...grpc.CallOption) (*personnel.WorkersResponse, error) {
	f.byStatus = req.GetValue()
	return &personnel.WorkersResponse{
		Items: []*personnel.Worker{{Id: "10000000-0000-4000-8000-000000000011", Status: req.GetValue()}},
		Page:  &common.PageInfo{Total: 1, Limit: req.GetLimit(), Offset: req.GetOffset()},
	}, nil
}

type learningCompositeFixture struct {
	learning.LearningServiceClient
	gotKey     *learning.LearningCourseCompetencyKey
	updatedReq *learning.UpdateLearningCourseCompetencyRequest
}

func (f *learningCompositeFixture) GetLearningCourseCompetencyByID(_ context.Context, key *learning.LearningCourseCompetencyKey, _ ...grpc.CallOption) (*learning.LearningCourseCompetencyResponse, error) {
	f.gotKey = key
	return &learning.LearningCourseCompetencyResponse{
		Item: &learning.LearningCourseCompetency{
			CourseId:     key.GetCourseId(),
			CompetencyId: key.GetCompetencyId(),
			TargetLevel:  3,
		},
	}, nil
}

func (f *learningCompositeFixture) UpdateLearningCourseCompetency(_ context.Context, req *learning.UpdateLearningCourseCompetencyRequest, _ ...grpc.CallOption) (*learning.LearningCourseCompetencyResponse, error) {
	f.updatedReq = req
	return &learning.LearningCourseCompetencyResponse{
		Item: &learning.LearningCourseCompetency{
			CourseId:     req.GetKey().GetCourseId(),
			CompetencyId: req.GetKey().GetCompetencyId(),
			TargetLevel:  req.GetTargetLevel(),
		},
	}, nil
}

type workforceMaskFixture struct {
	workforce.WorkforcePlanningServiceClient
	updatedReq *workforce.UpdateJobFamilyRequest
}

func (f *workforceMaskFixture) UpdateJobFamily(_ context.Context, req *workforce.UpdateJobFamilyRequest, _ ...grpc.CallOption) (*workforce.JobFamilyResponse, error) {
	f.updatedReq = req
	return &workforce.JobFamilyResponse{
		Item: &workforce.JobFamily{
			Id:         req.GetId(),
			FamilyCode: req.GetInput().GetFamilyCode(),
			FamilyName: "Retail Ops",
			JobDomain:  "operations",
			IsActive:   true,
		},
	}, nil
}

type performanceCmdFixture struct {
	performance.PerformanceServiceClient
	launchedID string
}

func (f *performanceCmdFixture) LaunchPerformanceCycle(_ context.Context, req *common.IDRequest, _ ...grpc.CallOption) (*performance.PerformanceCycleResponse, error) {
	f.launchedID = req.GetId()
	return &performance.PerformanceCycleResponse{
		Item: &performance.PerformanceCycle{Id: req.GetId(), CycleCode: "2026-H1", Status: "open"},
	}, nil
}

type successionCmdFixture struct {
	succession.SuccessionServiceClient
	approvedID string
}

func (f *successionCmdFixture) ApproveSuccessionPlan(_ context.Context, req *common.IDRequest, _ ...grpc.CallOption) (*succession.SuccessionPlanResponse, error) {
	f.approvedID = req.GetId()
	return &succession.SuccessionPlanResponse{
		Value: &succession.SuccessionPlan{Id: req.GetId(), PlanCode: "SP-01"},
	}, nil
}

type policyCmdFixture struct {
	policy.PolicyServiceClient
	ackReq *policy.AcknowledgePolicyRequest
}

func (f *policyCmdFixture) AcknowledgePolicy(_ context.Context, req *policy.AcknowledgePolicyRequest, _ ...grpc.CallOption) (*policy.PolicyAcknowledgementResponse, error) {
	f.ackReq = req
	return &policy.PolicyAcknowledgementResponse{
		Value: &policy.PolicyAcknowledgement{
			Id:       "10000000-0000-4000-8000-000000000099",
			PolicyId: req.GetId(),
			WorkerId: "10000000-0000-4000-8000-000000000088",
		},
	}, nil
}

func newTestEmployeeSession(userKind string, mfa bool, perms ...string) *middleware.ResolvedSession {
	grants := make([]authz.PermissionGrant, 0, len(perms))
	for _, p := range perms {
		grants = append(grants, authz.PermissionGrant{
			Code:  p,
			Scope: authz.ScopeGrant{Type: authz.ScopeGlobal},
		})
	}
	var issuedAt int64
	if mfa {
		issuedAt = time.Now().Unix()
	}
	return &middleware.ResolvedSession{
		Token: &authsession.SessionTokenRecord{
			MFAAuthenticated: mfa,
			IssuedAtUnix:     issuedAt,
		},
		Snapshot: &authz.AccessSnapshot{
			Status:      authz.StatusActive,
			Permissions: grants,
		},
		Caller: authcontext.CallerContext{
			Authenticated: true,
			UserKind:      userKind,
		},
	}
}

func newTestAdapter(t *testing.T) *Adapter {
	t.Helper()
	auth, err := middleware.NewAuthManager(&sessionStoreFixture{}, middleware.NewSnapshotLoader(nil, 0, 0), middleware.SessionConfig{})
	if err != nil {
		t.Fatal(err)
	}
	return NewAdapter(auth, middleware.NewAuthorizer(auth, nil, 0), nil)
}

func TestConsolidatedFinderSingleFilterAndMultiFilterRejection(t *testing.T) {
	t.Parallel()
	a := newTestAdapter(t)
	client := &finderPersonnelFixture{}
	a.personnel = client

	ctx := middleware.WithResolvedSession(t.Context(), newTestEmployeeSession("employee", true, "human_resource.workers.read"))
	statusActive := "active"
	workerNum := "W-001"

	// Menolak lebih dari satu filter aktif.
	if _, err := a.WorkersFiltered(ctx, 10, 0, &workerNum, nil, nil, nil, &statusActive, nil, nil, nil); err == nil {
		t.Fatal("expected error when supplying more than one filter")
	}

	// Meneruskan tepat satu filter ke RPC FindWorkerByStatus.
	page, err := a.WorkersFiltered(ctx, 10, 0, nil, nil, nil, nil, &statusActive, nil, nil, nil)
	if err != nil {
		t.Fatalf("WorkersFiltered failed: %v", err)
	}
	if client.byStatus != "active" || len(page.Items) != 1 {
		t.Fatalf("unexpected finder result: byStatus=%q len=%d", client.byStatus, len(page.Items))
	}
}

func TestIDORAndSensitivityGuardsOnMutationsAndReads(t *testing.T) {
	t.Parallel()
	a := newTestAdapter(t)
	a.performance = &performanceCmdFixture{}
	a.succession = &successionCmdFixture{}

	const sampleID = "10000000-0000-4000-8000-000000000022"

	// Customer tidak boleh memanggil operasi internal HRM meski punya token.
	custCtx := middleware.WithResolvedSession(t.Context(), newTestEmployeeSession("customer", true, "human_resource.performance_cycles.launch"))
	if _, err := a.LaunchPerformanceCycle(custCtx, sampleID); err == nil {
		t.Fatal("expected customer session to be rejected")
	}

	// Operasi sensitif (hr_workers.read) menolak employee tanpa MFA.
	noMFACtx := middleware.WithResolvedSession(t.Context(), newTestEmployeeSession("employee", false, "human_resource.workers.read"))
	if _, err := a.WorkersFiltered(noMFACtx, 10, 0, nil, nil, nil, nil, nil, nil, nil, nil); err == nil {
		t.Fatal("expected missing MFA to be rejected on sensitive hr_workers read")
	}

	// Operasi dengan grant + employee lolos ke gRPC.
	empCtx := middleware.WithResolvedSession(t.Context(), newTestEmployeeSession("employee", true,
		"human_resource.performance_cycles.launch",
		"human_resource.succession_plans.approve",
	))
	if _, err := a.LaunchPerformanceCycle(empCtx, sampleID); err != nil {
		t.Fatalf("LaunchPerformanceCycle failed: %v", err)
	}
	if _, err := a.ApproveSuccessionPlan(empCtx, sampleID); err != nil {
		t.Fatalf("ApproveSuccessionPlan failed: %v", err)
	}
}

func TestCompositeKeyAndFieldMaskValidation(t *testing.T) {
	t.Parallel()
	a := newTestAdapter(t)
	lc := &learningCompositeFixture{}
	wf := &workforceMaskFixture{}
	a.learning = lc
	a.workforce = wf

	ctx := middleware.WithResolvedSession(t.Context(), newTestEmployeeSession("employee", true,
		"human_resource.learning_course_competencies.read",
		"human_resource.learning_course_competencies.update",
		"human_resource.job_families.update",
	))

	const courseID = "10000000-0000-4000-8000-000000000031"
	const compID = "10000000-0000-4000-8000-000000000032"
	compKey := nodeID("learning-course-competency", courseID+":"+compID)

	got, err := a.LearningCourseCompetency(ctx, compKey)
	if err != nil {
		t.Fatalf("LearningCourseCompetency failed: %v", err)
	}
	if got.RawID != courseID+":"+compID || lc.gotKey.GetCourseId() != courseID {
		t.Fatalf("unexpected composite key result: %+v", got)
	}

	lvl := 4
	if _, err := a.UpdateLearningCourseCompetency(ctx, compKey, model.UpdateLearningCourseCompetencyInput{
		TargetLevel: graphql.OmittableOf(&lvl),
	}); err != nil {
		t.Fatalf("UpdateLearningCourseCompetency failed: %v", err)
	}
	if lc.updatedReq.GetTargetLevel() != 4 || len(lc.updatedReq.GetUpdateMask().GetPaths()) != 1 {
		t.Fatalf("unexpected composite update request: %+v", lc.updatedReq)
	}

	// UpdateJobFamily kosong harus ditolak sebelum memanggil RPC.
	if _, err := a.UpdateJobFamily(ctx, courseID, model.UpdateJobFamilyInput{}); err == nil || wf.updatedReq != nil {
		t.Fatal("expected empty UpdateJobFamily to be rejected")
	}

	code := "JF-OPS"
	if _, err := a.UpdateJobFamily(ctx, courseID, model.UpdateJobFamilyInput{
		FamilyCode: graphql.OmittableOf(&code),
	}); err != nil {
		t.Fatalf("UpdateJobFamily failed: %v", err)
	}
	if wf.updatedReq == nil || len(wf.updatedReq.GetUpdateMask().GetPaths()) != 1 || wf.updatedReq.GetUpdateMask().GetPaths()[0] != "family_code" {
		t.Fatalf("unexpected UpdateJobFamily mask: %+v", wf.updatedReq)
	}
}

func TestAcknowledgePolicyAndTalentCalibrationValidation(t *testing.T) {
	t.Parallel()
	a := newTestAdapter(t)
	pol := &policyCmdFixture{}
	a.policy = pol

	ctx := middleware.WithResolvedSession(t.Context(), newTestEmployeeSession("employee", true,
		"human_resource.policy_acknowledgements.acknowledge",
		"human_resource.talent_calibration_sessions.approve",
	))
	const polID = "10000000-0000-4000-8000-000000000041"
	const docID = "10000000-0000-4000-8000-000000000042"

	ack, err := a.AcknowledgePolicy(ctx, polID, strPtr(docID))
	if err != nil {
		t.Fatalf("AcknowledgePolicy failed: %v", err)
	}
	if ack == nil || pol.ackReq.GetId() != polID || pol.ackReq.GetEvidenceDocumentVersionReferenceId() != docID {
		t.Fatalf("unexpected AcknowledgePolicy request: %+v", pol.ackReq)
	}

	// Base64 tidak valid ditolak sebelum memanggil gRPC talent.
	_, err = a.RecordTalentCalibrationDecision(ctx, model.RecordTalentCalibrationDecisionInput{
		SessionID:                           polID,
		AssessmentID:                        polID,
		DimensionID:                         polID,
		CalibratedScore:                     4,
		RationaleDocumentVersionReferenceID: docID,
		RationaleDigest:                     "not-valid-base64!!!",
	})
	if err == nil || !strings.Contains(err.Error(), "base64") {
		t.Fatalf("expected invalid base64 error, got: %v", err)
	}

	// Pastikan helper base64 valid tidak panik dan mem-fail-closed saat client nil.
	validDigest := base64.StdEncoding.EncodeToString([]byte("rationale"))
	_, err = a.RecordTalentCalibrationDecision(ctx, model.RecordTalentCalibrationDecisionInput{
		SessionID:                           polID,
		AssessmentID:                        polID,
		DimensionID:                         polID,
		CalibratedScore:                     4,
		RationaleDocumentVersionReferenceID: docID,
		RationaleDigest:                     validDigest,
	})
	if err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("expected unavailable error when talent client nil, got: %v", err)
	}
}

func strPtr(s string) *string { return &s }
