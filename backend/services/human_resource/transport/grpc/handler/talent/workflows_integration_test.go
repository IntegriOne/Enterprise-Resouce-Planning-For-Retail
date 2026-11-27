//go:build integration

package talent_test

import (
	"bytes"
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/talent/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/talent"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/talent"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type talentWorkflowsPolicy struct {
	mu          sync.Mutex
	principalID string
}

func (p *talentWorkflowsPolicy) setPrincipal(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.principalID = id
}

func (p *talentWorkflowsPolicy) currentPrincipal() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.principalID
}

func (p *talentWorkflowsPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.currentPrincipal(), SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*talentWorkflowsPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

type talentWorkflowsFoundation struct{}

func (talentWorkflowsFoundation) VerifyLegalEntity(context.Context, string) error { return nil }
func (talentWorkflowsFoundation) VerifyPositionReferences(context.Context, security.PositionReferenceCheck) error {
	return nil
}
func (talentWorkflowsFoundation) VerifyWorkerReferences(context.Context, security.WorkerReferenceCheck) error {
	return nil
}
func (talentWorkflowsFoundation) VerifyOrganizationUnit(context.Context, string, string) error {
	return nil
}
func (talentWorkflowsFoundation) VerifyLocation(context.Context, string, string) error { return nil }
func (talentWorkflowsFoundation) VerifyParty(context.Context, string, bool) error      { return nil }

type talentReviewReaderStub struct{}

func (talentReviewReaderStub) GetPerformanceReviewByID(context.Context, string) (*entity.PerformanceReview, error) {
	return nil, hrerrors.ErrNotFound
}

func TestTalentWorkflowsWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open HRM test database failed")
	}
	pool, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(database)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.TalentDimensionValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	talentVals, err := security.TalentValidators(repos, talentWorkflowsFoundation{})
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range talentVals {
		validators[k] = v
	}

	ownerID := "00000000-0000-4000-8000-000000000031"
	assessorID := "00000000-0000-4000-8000-000000000032"
	calibratorID := "00000000-0000-4000-8000-000000000033"
	subjectPrincipalID := "00000000-0000-4000-8000-000000000034"
	policy := &talentWorkflowsPolicy{principalID: ownerID}

	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(validators),
		},
		PerformanceReviewReader:   talentReviewReaderStub{},
		TalentFramework:           repos.TalentFramework,
		TalentDimension:           repos.TalentDimension,
		TalentReviewCycle:         repos.TalentReviewCycle,
		TalentAssessment:          repos.TalentAssessment,
		TalentAssessmentScore:     repos.TalentAssessmentScore,
		TalentCalibrationSession:  repos.TalentCalibrationSession,
		TalentCalibrationDecision: repos.TalentCalibrationDecision,
		TalentPool:                repos.TalentPool,
		TalentPoolMembership:      repos.TalentPoolMembership,
	})
	if err != nil {
		t.Fatal(err)
	}

	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		return next(authcontext.WithCaller(ctx, authcontext.CallerContext{PrincipalID: policy.currentPrincipal(), Authenticated: true}), req)
	}))
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-talent-workflows-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewTalentServiceClient(conn)

	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	frameworkID, _ := ids.NewID()
	docID, _ := ids.NewID()
	ownerWorkerID, _ := ids.NewID()
	assessorWorkerID, _ := ids.NewID()
	calibratorWorkerID, _ := ids.NewID()
	subjectWorkerID, _ := ids.NewID()
	var dimID, cycleID, closeCycleID, assessID, withdrawAssessID, scoreID, sessionID, decisionID, poolID, memID, memRejectedID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		err := database.WithContext(cleanupCtx).Transaction(func(tx *gorm.DB) error {
			for _, ddl := range []string{
				"ALTER TABLE hr_talent_calibration_decisions DISABLE TRIGGER tr_hr_talent_calibration_immutable",
				"ALTER TABLE hr_talent_assessment_scores DISABLE TRIGGER tr_hr_talent_score_guard",
				"ALTER TABLE hr_talent_assessments DISABLE TRIGGER tr_hr_talent_assessment_guard",
				"ALTER TABLE hr_talent_dimensions DISABLE TRIGGER tr_hr_talent_dimension_guard",
				"ALTER TABLE hr_talent_frameworks DISABLE TRIGGER tr_hr_talent_framework_immutable",
			} {
				if err := tx.Exec(ddl).Error; err != nil {
					return err
				}
			}
			for _, stmt := range []struct {
				q string
				v string
			}{
				{"DELETE FROM hr_talent_pool_memberships WHERE id = ?", memID},
				{"DELETE FROM hr_talent_pool_memberships WHERE id = ?", memRejectedID},
				{"DELETE FROM hr_talent_pools WHERE id = ?", poolID},
				{"DELETE FROM hr_talent_calibration_decisions WHERE id = ?", decisionID},
				{"DELETE FROM hr_talent_calibration_sessions WHERE id = ?", sessionID},
				{"DELETE FROM hr_talent_assessment_scores WHERE id = ?", scoreID},
				{"DELETE FROM hr_talent_assessments WHERE id = ?", assessID},
				{"DELETE FROM hr_talent_assessments WHERE id = ?", withdrawAssessID},
				{"DELETE FROM hr_talent_review_cycles WHERE id = ?", cycleID},
				{"DELETE FROM hr_talent_review_cycles WHERE id = ?", closeCycleID},
				{"DELETE FROM hr_talent_dimensions WHERE id = ?", dimID},
				{"DELETE FROM hr_talent_frameworks WHERE id = ?", frameworkID},
				{"DELETE FROM hr_workers WHERE id = ?", ownerWorkerID},
				{"DELETE FROM hr_workers WHERE id = ?", assessorWorkerID},
				{"DELETE FROM hr_workers WHERE id = ?", calibratorWorkerID},
				{"DELETE FROM hr_workers WHERE id = ?", subjectWorkerID},
			} {
				if stmt.v != "" {
					if err := tx.Exec(stmt.q, stmt.v).Error; err != nil {
						return err
					}
				}
			}
			for _, agg := range []string{memID, memRejectedID, poolID, decisionID, sessionID, scoreID, assessID, withdrawAssessID, cycleID, closeCycleID, dimID} {
				if agg != "" {
					if err := tx.Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", agg).Error; err != nil {
						return err
					}
				}
			}
			for _, ddl := range []string{
				"ALTER TABLE hr_talent_calibration_decisions ENABLE TRIGGER tr_hr_talent_calibration_immutable",
				"ALTER TABLE hr_talent_assessment_scores ENABLE TRIGGER tr_hr_talent_score_guard",
				"ALTER TABLE hr_talent_assessments ENABLE TRIGGER tr_hr_talent_assessment_guard",
				"ALTER TABLE hr_talent_dimensions ENABLE TRIGGER tr_hr_talent_dimension_guard",
				"ALTER TABLE hr_talent_frameworks ENABLE TRIGGER tr_hr_talent_framework_immutable",
			} {
				if err := tx.Exec(ddl).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			t.Errorf("talent workflows cleanup failed: %v", err)
		}
	})

	for _, w := range []struct {
		id        string
		principal string
		suffix    string
	}{
		{ownerWorkerID, ownerID, "OWN"},
		{assessorWorkerID, assessorID, "ASS"},
		{calibratorWorkerID, calibratorID, "CAL"},
		{subjectWorkerID, subjectPrincipalID, "SUB"},
	} {
		personID, _ := ids.NewID()
		if err := database.WithContext(ctx).Exec(
			`INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
			 primary_legal_entity_reference_id, iam_principal_reference_id, hire_date, service_start_date)
			 VALUES (?, ?, ?, 'employee', 'active', ?, ?, '2025-01-01', '2025-01-01')`,
			w.id, "TL-"+w.suffix+"-"+w.id, personID, legalID, w.principal,
		).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_talent_frameworks (id, framework_code, framework_version, framework_name,
		 rubric_document_version_reference_id, rubric_digest, effective_from, approved_by_principal_id, approved_at)
		 VALUES (?, ?, 1, 'Talent Rubric', ?, ?, '2025-01-01', ?, NOW())`,
		frameworkID, "FW-WF-"+frameworkID, docID, bytes.Repeat([]byte{0x12}, 32), ownerID,
	).Error; err != nil {
		t.Fatal(err)
	}

	dimReq := true
	dimCode, dimName, dimDef, dimWeight := "IMPACT", "Impact", "Measured impact", "100.000"
	dimRes, err := client.CreateTalentDimension(ctx, &pb.CreateTalentDimensionRequest{Input: &pb.TalentDimensionInput{
		FrameworkId: &frameworkID, DimensionCode: &dimCode, DimensionName: &dimName,
		Definition: &dimDef, WeightPercent: &dimWeight, IsRequired: &dimReq,
		RatingAnchors: &common.JsonDocument{Utf8: []byte(`{"1":"basic","2":"developing","3":"capable","4":"strong","5":"expert"}`)},
	}})
	if err != nil {
		t.Fatalf("create dimension: %v", err)
	}
	dimID = dimRes.GetValue().GetId()

	cycleStart := timestamppb.New(time.Now().UTC().AddDate(0, -1, 0))
	cycleEnd := timestamppb.New(time.Now().UTC().AddDate(0, 2, 0))
	cycleCode := "CYC-WF-" + frameworkID
	cycleRes, err := client.CreateTalentReviewCycle(ctx, &pb.CreateTalentReviewCycleRequest{Input: &pb.TalentReviewCycleInput{
		CycleCode: &cycleCode, FrameworkId: &frameworkID, LegalEntityReferenceId: &legalID,
		PeriodStart: cycleStart, PeriodEnd: cycleEnd, OwnerPrincipalId: &ownerID,
	}})
	if err != nil {
		t.Fatalf("create cycle: %v", err)
	}
	cycleID = cycleRes.GetValue().GetId()
	if launched, err := client.LaunchTalentReviewCycle(ctx, &common.IDRequest{Id: cycleID}); err != nil || launched.GetValue().GetStatus() != "open" {
		t.Fatalf("launch cycle: res=%+v err=%v", launched, err)
	}

	// Buat asesmen draft lalu verifikasi finalisasi tanpa skor ditolak.
	policy.setPrincipal(assessorID)
	v1 := int32(1)
	assessRes, err := client.CreateTalentAssessment(ctx, &pb.CreateTalentAssessmentRequest{Input: &pb.TalentAssessmentInput{
		WorkerId: &subjectWorkerID, TalentReviewCycleId: &cycleID, FrameworkId: &frameworkID,
		AssessmentVersion: &v1, EvidenceDocumentVersionReferenceId: &docID,
		EvidenceDigest: bytes.Repeat([]byte{0x23}, 32),
	}})
	if err != nil {
		t.Fatalf("create assessment: %v", err)
	}
	assessID = assessRes.GetValue().GetId()
	if _, err := client.FinalizeTalentAssessment(ctx, &common.IDRequest{Id: assessID}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("finalize unscored assessment = %v, want InvalidArgument", status.Code(err))
	}

	rawScore := int32(3)
	scoreRes, err := client.CreateTalentAssessmentScore(ctx, &pb.CreateTalentAssessmentScoreRequest{Input: &pb.TalentAssessmentScoreInput{
		AssessmentId: &assessID, FrameworkId: &frameworkID, DimensionId: &dimID,
		Score: &rawScore, EvidenceDocumentVersionReferenceId: &docID, EvidenceDigest: bytes.Repeat([]byte{0x34}, 32),
	}})
	if err != nil {
		t.Fatalf("create score: %v", err)
	}
	scoreID = scoreRes.GetValue().GetId()

	if sub, err := client.SubmitTalentAssessment(ctx, &common.IDRequest{Id: assessID}); err != nil || sub.GetValue().GetStatus() != "submitted" {
		t.Fatalf("submit assessment: res=%+v err=%v", sub, err)
	}
	fin, err := client.FinalizeTalentAssessment(ctx, &common.IDRequest{Id: assessID})
	if err != nil || fin.GetValue().GetStatus() != "finalized" {
		t.Fatalf("finalize assessment: res=%+v err=%v", fin, err)
	}

	// Buat asesmen versi 2 dan tarik kembali (withdraw).
	v2 := int32(2)
	assess2Res, err := client.CreateTalentAssessment(ctx, &pb.CreateTalentAssessmentRequest{Input: &pb.TalentAssessmentInput{
		WorkerId: &subjectWorkerID, TalentReviewCycleId: &cycleID, FrameworkId: &frameworkID,
		AssessmentVersion: &v2, EvidenceDocumentVersionReferenceId: &docID,
		EvidenceDigest: bytes.Repeat([]byte{0x35}, 32),
	}})
	if err != nil {
		t.Fatalf("create assessment v2: %v", err)
	}
	withdrawAssessID = assess2Res.GetValue().GetId()
	if withdrawn, err := client.WithdrawTalentAssessment(ctx, &common.IDRequest{Id: withdrawAssessID}); err != nil || withdrawn.GetValue().GetStatus() != "withdrawn" {
		t.Fatalf("withdraw assessment: res=%+v err=%v", withdrawn, err)
	}

	// Sesi kalibrasi, SoD penilai/subjek, dan keputusan kalibrasi.
	policy.setPrincipal(calibratorID)
	sessCode := "CAL-" + cycleID[:8]
	sessRes, err := client.CreateTalentCalibrationSession(ctx, &pb.CreateTalentCalibrationSessionRequest{Input: &pb.TalentCalibrationSessionInput{
		TalentReviewCycleId: &cycleID, SessionCode: &sessCode, ChairPrincipalId: &calibratorID,
		ScheduledAt: timestamppb.New(time.Now().UTC().Add(24 * time.Hour)),
	}})
	if err != nil {
		t.Fatalf("create calibration session: %v", err)
	}
	sessionID = sessRes.GetValue().GetId()

	policy.setPrincipal(ownerID)
	if calCycle, err := client.StartTalentReviewCycleCalibration(ctx, &common.IDRequest{Id: cycleID}); err != nil || calCycle.GetValue().GetStatus() != "calibration" {
		t.Fatalf("start calibration: res=%+v err=%v", calCycle, err)
	}
	policy.setPrincipal(calibratorID)
	if openedSess, err := client.OpenTalentCalibrationSession(ctx, &common.IDRequest{Id: sessionID}); err != nil || openedSess.GetValue().GetStatus() != "open" {
		t.Fatalf("open calibration session: res=%+v err=%v", openedSess, err)
	}

	policy.setPrincipal(assessorID)
	if _, err := client.RecordTalentCalibrationDecision(ctx, &pb.RecordTalentCalibrationDecisionRequest{
		SessionId: sessionID, AssessmentId: assessID, DimensionId: dimID,
		CalibratedScore: 4, RationaleDocumentVersionReferenceId: docID, RationaleDigest: bytes.Repeat([]byte{0x45}, 32),
	}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("assessor calibration SoD = %v, want PermissionDenied", status.Code(err))
	}

	policy.setPrincipal(subjectPrincipalID)
	if _, err := client.RecordTalentCalibrationDecision(ctx, &pb.RecordTalentCalibrationDecisionRequest{
		SessionId: sessionID, AssessmentId: assessID, DimensionId: dimID,
		CalibratedScore: 4, RationaleDocumentVersionReferenceId: docID, RationaleDigest: bytes.Repeat([]byte{0x45}, 32),
	}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("subject calibration SoD = %v, want PermissionDenied", status.Code(err))
	}

	policy.setPrincipal(calibratorID)
	decRes, err := client.RecordTalentCalibrationDecision(ctx, &pb.RecordTalentCalibrationDecisionRequest{
		SessionId: sessionID, AssessmentId: assessID, DimensionId: dimID,
		CalibratedScore: 4, RationaleDocumentVersionReferenceId: docID, RationaleDigest: bytes.Repeat([]byte{0x45}, 32),
	})
	if err != nil || decRes.GetValue().GetCalibratedScore() != 4 {
		t.Fatalf("record calibration decision: res=%+v err=%v", decRes, err)
	}
	decisionID = decRes.GetValue().GetId()

	// Menutup siklus saat masih ada sesi terbuka ditolak.
	policy.setPrincipal(ownerID)
	if _, err := client.CloseTalentReviewCycle(ctx, &common.IDRequest{Id: cycleID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("close cycle with open session = %v, want FailedPrecondition", status.Code(err))
	}

	// Siklus kedua tanpa sesi terbuka berhasil ditutup.
	closeCode := "CYC-CLS-" + frameworkID
	closeRes, err := client.CreateTalentReviewCycle(ctx, &pb.CreateTalentReviewCycleRequest{Input: &pb.TalentReviewCycleInput{
		CycleCode: &closeCode, FrameworkId: &frameworkID, LegalEntityReferenceId: &legalID,
		PeriodStart: cycleStart, PeriodEnd: cycleEnd, OwnerPrincipalId: &ownerID,
	}})
	if err != nil {
		t.Fatalf("create closeable cycle: %v", err)
	}
	closeCycleID = closeRes.GetValue().GetId()
	if _, err := client.LaunchTalentReviewCycle(ctx, &common.IDRequest{Id: closeCycleID}); err != nil {
		t.Fatalf("launch closeable cycle: %v", err)
	}
	if closed, err := client.CloseTalentReviewCycle(ctx, &common.IDRequest{Id: closeCycleID}); err != nil || closed.GetValue().GetStatus() != "closed" {
		t.Fatalf("close cycle: res=%+v err=%v", closed, err)
	}

	// Talent pool & lifecycle keanggotaan (nominasi -> approve -> GIST/validator reject overlap -> exit -> reject).
	poolVer := int32(1)
	poolCode := "POOL-LEAD-" + legalID[:8]
	poolName := "Leadership Pool"
	poolRes, err := client.CreateTalentPool(ctx, &pb.CreateTalentPoolRequest{Input: &pb.TalentPoolInput{
		PoolCode: &poolCode, PoolVersion: &poolVer, PoolName: &poolName,
		LegalEntityReferenceId: &legalID, EligibilityDocumentVersionReferenceId: &docID,
		EligibilityDigest: bytes.Repeat([]byte{0x56}, 32),
		ValidFrom:         timestamppb.New(time.Now().UTC().AddDate(0, -1, 0)),
		ValidTo:           timestamppb.New(time.Now().UTC().AddDate(1, 0, 0)),
		OwnerPrincipalId:  &ownerID,
	}})
	if err != nil {
		t.Fatalf("create talent pool: %v", err)
	}
	poolID = poolRes.GetValue().GetId()

	nowDate := time.Now().UTC()
	todayDate := time.Date(nowDate.Year(), nowDate.Month(), nowDate.Day(), 12, 0, 0, 0, time.UTC)
	validFrom := timestamppb.New(todayDate)
	validTo := timestamppb.New(todayDate.Add(30 * 24 * time.Hour))
	policy.setPrincipal(assessorID)
	memRes, err := client.CreateTalentPoolMembership(ctx, &pb.CreateTalentPoolMembershipRequest{Input: &pb.TalentPoolMembershipInput{
		TalentPoolId: &poolID, WorkerId: &subjectWorkerID, AssessmentId: &assessID,
		NominatedByPrincipalId: &assessorID, ValidFrom: validFrom, ValidTo: validTo,
	}})
	if err != nil {
		t.Fatalf("create membership: %v", err)
	}
	memID = memRes.GetValue().GetId()

	if _, err := client.ApproveTalentPoolMembership(ctx, &common.IDRequest{Id: memID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("nominator self approval = %v, want PermissionDenied", status.Code(err))
	}
	policy.setPrincipal(subjectPrincipalID)
	if _, err := client.ApproveTalentPoolMembership(ctx, &common.IDRequest{Id: memID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("subject self membership approval = %v, want PermissionDenied", status.Code(err))
	}

	policy.setPrincipal(ownerID)
	if appMem, err := client.ApproveTalentPoolMembership(ctx, &common.IDRequest{Id: memID}); err != nil || appMem.GetValue().GetStatus() != "active" {
		t.Fatalf("approve membership: res=%+v err=%v", appMem, err)
	}
	if _, err := client.ExitTalentPoolMembership(ctx, &pb.TalentReasonActionRequest{Id: memID, ReasonCode: "GRADUATED"}); err != nil {
		t.Fatalf("exit membership: %v", err)
	}

	// Nominasi kedua periode berbeda lalu ditolak.
	laterFrom := timestamppb.New(todayDate.Add(45 * 24 * time.Hour))
	laterTo := timestamppb.New(todayDate.Add(60 * 24 * time.Hour))
	policy.setPrincipal(assessorID)
	mem2Res, err := client.CreateTalentPoolMembership(ctx, &pb.CreateTalentPoolMembershipRequest{Input: &pb.TalentPoolMembershipInput{
		TalentPoolId: &poolID, WorkerId: &subjectWorkerID, AssessmentId: &assessID,
		NominatedByPrincipalId: &assessorID, ValidFrom: laterFrom, ValidTo: laterTo,
	}})
	if err != nil {
		t.Fatalf("create second membership: %v", err)
	}
	memRejectedID = mem2Res.GetValue().GetId()
	policy.setPrincipal(ownerID)
	if rejMem, err := client.RejectTalentPoolMembership(ctx, &pb.TalentReasonActionRequest{Id: memRejectedID, ReasonCode: "DEFERRED"}); err != nil || rejMem.GetValue().GetStatus() != "rejected" {
		t.Fatalf("reject membership: res=%+v err=%v", rejMem, err)
	}
}
