//go:build integration

package promotion_test

import (
	"bytes"
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/promotion/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/promotion"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/promotion"
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

type promotionTestPolicy struct {
	mu          sync.Mutex
	principalID string
}

func (p *promotionTestPolicy) setPrincipal(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.principalID = id
}

func (p *promotionTestPolicy) currentPrincipal() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.principalID
}

func (p *promotionTestPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.currentPrincipal(), SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*promotionTestPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func strPtr(v string) *string { return &v }

type promotionPositionVerifier struct {
	reject atomic.Bool
}

func (v *promotionPositionVerifier) VerifyPositionReferences(_ context.Context, check security.PositionReferenceCheck) error {
	if v.reject.Load() || check.OrganizationUnitID == "" || check.LegalEntityID == "" {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

func TestPromotionWorkflowAndApplicationWithPostgreSQL(t *testing.T) {
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
	verifier := &promotionPositionVerifier{}
	validators, err := security.PromotionValidators(repos, verifier)
	if err != nil {
		t.Fatal(err)
	}

	recommenderID := "00000000-0000-4000-8000-000000000051"
	approverID := "00000000-0000-4000-8000-000000000052"
	subjectPrincipalID := "00000000-0000-4000-8000-000000000053"
	policy := &promotionTestPolicy{principalID: recommenderID}

	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(validators),
		},
		Worker:                  repos.Worker,
		WorkerAssignment:        repos.WorkerAssignment,
		WorkerLifecycleEvent:    repos.WorkerLifecycleEvent,
		Position:                repos.Position,
		TalentAssessment:        repos.TalentAssessment,
		PromotionRecommendation: repos.PromotionRecommendation,
		PromotionDecision:       repos.PromotionDecision,
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
	conn, err := grpc.NewClient("passthrough:///hrm-promotion-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewPromotionServiceClient(conn)

	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	unitID, _ := ids.NewID()
	familyID, _ := ids.NewID()
	jobID, _ := ids.NewID()
	sourcePosID, _ := ids.NewID()
	targetPosID, _ := ids.NewID()
	subjectWorkerID, _ := ids.NewID()
	recommenderWorkerID, _ := ids.NewID()
	approverWorkerID, _ := ids.NewID()
	sourceAssignID, _ := ids.NewID()
	frameworkID, _ := ids.NewID()
	dimID, _ := ids.NewID()
	cycleID, _ := ids.NewID()
	assessID, _ := ids.NewID()
	scoreID, _ := ids.NewID()
	docID, _ := ids.NewID()
	var recReturnedID, recAppliedID, recRejectedID, appliedAssignID, decReturnID, decApproveID, decRejectID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		err := database.WithContext(cleanupCtx).Transaction(func(tx *gorm.DB) error {
			for _, ddl := range []string{
				"ALTER TABLE hr_promotion_recommendations DISABLE TRIGGER tr_hr_promotion_recommendation_guard",
				"ALTER TABLE hr_promotion_decisions DISABLE TRIGGER tr_hr_promotion_decision_immutable",
				"ALTER TABLE hr_talent_assessment_scores DISABLE TRIGGER tr_hr_talent_score_guard",
				"ALTER TABLE hr_talent_assessments DISABLE TRIGGER tr_hr_talent_assessment_guard",
				"ALTER TABLE hr_talent_dimensions DISABLE TRIGGER tr_hr_talent_dimension_guard",
				"ALTER TABLE hr_talent_frameworks DISABLE TRIGGER tr_hr_talent_framework_immutable",
			} {
				if err := tx.Exec(ddl).Error; err != nil {
					return err
				}
			}
			if err := tx.Exec("UPDATE hr_workers SET primary_assignment_id = NULL WHERE id = ?", subjectWorkerID).Error; err != nil {
				return err
			}
			for _, recID := range []string{recRejectedID, recAppliedID, recReturnedID} {
				if recID != "" {
					if err := tx.Exec(
						`UPDATE hr_promotion_recommendations
						 SET status = 'draft', supersedes_recommendation_id = NULL, approval_decision_id = NULL,
						     applied_assignment_id = NULL, applied_lifecycle_event_id = NULL, applied_at = NULL
						 WHERE id = ?`, recID,
					).Error; err != nil {
						return err
					}
				}
			}
			for _, stmt := range []struct {
				q string
				v string
			}{
				{"DELETE FROM hr_promotion_decisions WHERE id = ?", decRejectID},
				{"DELETE FROM hr_promotion_decisions WHERE id = ?", decApproveID},
				{"DELETE FROM hr_promotion_decisions WHERE id = ?", decReturnID},
				{"DELETE FROM hr_promotion_recommendations WHERE id = ?", recRejectedID},
				{"DELETE FROM hr_promotion_recommendations WHERE id = ?", recAppliedID},
				{"DELETE FROM hr_promotion_recommendations WHERE id = ?", recReturnedID},
				{"DELETE FROM hr_worker_lifecycle_events WHERE worker_id = ?", subjectWorkerID},
				{"DELETE FROM hr_worker_assignments WHERE id = ?", appliedAssignID},
				{"DELETE FROM hr_worker_assignments WHERE id = ?", sourceAssignID},
				{"DELETE FROM hr_talent_assessment_scores WHERE id = ?", scoreID},
				{"DELETE FROM hr_talent_assessments WHERE id = ?", assessID},
				{"DELETE FROM hr_talent_review_cycles WHERE id = ?", cycleID},
				{"DELETE FROM hr_talent_dimensions WHERE id = ?", dimID},
				{"DELETE FROM hr_talent_frameworks WHERE id = ?", frameworkID},
				{"DELETE FROM hr_workers WHERE id = ?", subjectWorkerID},
				{"DELETE FROM hr_workers WHERE id = ?", recommenderWorkerID},
				{"DELETE FROM hr_workers WHERE id = ?", approverWorkerID},
				{"DELETE FROM hr_positions WHERE id = ?", sourcePosID},
				{"DELETE FROM hr_positions WHERE id = ?", targetPosID},
				{"DELETE FROM hr_jobs WHERE id = ?", jobID},
				{"DELETE FROM hr_job_families WHERE id = ?", familyID},
			} {
				if stmt.v != "" {
					if err := tx.Exec(stmt.q, stmt.v).Error; err != nil {
						return err
					}
				}
			}
			for _, agg := range []string{recRejectedID, recAppliedID, recReturnedID, subjectWorkerID} {
				if agg != "" {
					if err := tx.Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", agg).Error; err != nil {
						return err
					}
				}
			}
			for _, ddl := range []string{
				"ALTER TABLE hr_promotion_recommendations ENABLE TRIGGER tr_hr_promotion_recommendation_guard",
				"ALTER TABLE hr_promotion_decisions ENABLE TRIGGER tr_hr_promotion_decision_immutable",
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
			t.Errorf("promotion integration cleanup failed: %v", err)
		}
	})

	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, 'Engineering', 'engineering')",
		familyID, "PRM-FAM-"+familyID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, valid_from) VALUES (?, ?, ?, 'Engineer', '2025-01-01')",
		jobID, familyID, "PRM-JOB-"+jobID,
	).Error; err != nil {
		t.Fatal(err)
	}
	for _, pos := range []struct{ id, code string }{
		{sourcePosID, "SRC-POS-" + sourcePosID[:8]},
		{targetPosID, "TGT-POS-" + targetPosID[:8]},
	} {
		if err := database.WithContext(ctx).Exec(
			`INSERT INTO hr_positions (id, position_code, position_title, job_id, organization_unit_reference_id,
			 legal_entity_reference_id, position_status, headcount_authorized, valid_from)
			 VALUES (?, ?, 'Role', ?, ?, ?, 'active', 1.000, '2025-01-01')`,
			pos.id, pos.code, jobID, unitID, legalID,
		).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, w := range []struct{ id, principal, suffix string }{
		{subjectWorkerID, subjectPrincipalID, "SUB"},
		{recommenderWorkerID, recommenderID, "REC"},
		{approverWorkerID, approverID, "APP"},
	} {
		personID, _ := ids.NewID()
		if err := database.WithContext(ctx).Exec(
			`INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
			 primary_legal_entity_reference_id, iam_principal_reference_id, hire_date, service_start_date, created_at, updated_at)
			 VALUES (?, ?, ?, 'employee', 'active', ?, ?, '2025-01-01', '2025-01-01', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour')`,
			w.id, "PRM-"+w.suffix+"-"+w.id, personID, legalID, w.principal,
		).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_worker_assignments (id, worker_id, position_id, manager_worker_id, assignment_type, fte_percent, valid_from)
		 VALUES (?, ?, ?, ?, 'primary', 100.00, '2026-01-01')`,
		sourceAssignID, subjectWorkerID, sourcePosID, recommenderWorkerID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		"UPDATE hr_workers SET primary_assignment_id = ? WHERE id = ?",
		sourceAssignID, subjectWorkerID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_talent_frameworks (id, framework_code, framework_version, framework_name,
		 rubric_document_version_reference_id, rubric_digest, effective_from, approved_by_principal_id, approved_at)
		 VALUES (?, ?, 1, 'Promo FW', ?, ?, '2025-01-01', ?, NOW())`,
		frameworkID, "PRM-FW-"+frameworkID, docID, bytes.Repeat([]byte{0x71}, 32), recommenderID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_talent_dimensions (id, framework_id, dimension_code, dimension_name, definition,
		 rating_anchors, weight_percent, is_required)
		 VALUES (?, ?, 'DELIVERY', 'Delivery', 'Exec', '{"1":"a","2":"b","3":"c","4":"d","5":"e"}'::jsonb, 100.000, true)`,
		dimID, frameworkID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_talent_review_cycles (id, cycle_code, framework_id, legal_entity_reference_id,
		 period_start, period_end, status, owner_principal_id)
		 VALUES (?, ?, ?, ?, '2025-01-01', '2027-12-31', 'open', ?)`,
		cycleID, "PRM-CYC-"+cycleID, frameworkID, legalID, recommenderID,
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_talent_assessments (id, worker_id, talent_review_cycle_id, framework_id, assessment_version,
		 status, assessed_by_principal_id, assessed_at, evidence_document_version_reference_id, evidence_digest)
		 VALUES (?, ?, ?, ?, 1, 'draft', ?, NOW(), ?, ?)`,
		assessID, subjectWorkerID, cycleID, frameworkID, recommenderID, docID, bytes.Repeat([]byte{0x72}, 32),
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_talent_assessment_scores (id, assessment_id, framework_id, dimension_id, score,
		 evidence_document_version_reference_id, evidence_digest)
		 VALUES (?, ?, ?, ?, 5, ?, ?)`,
		scoreID, assessID, frameworkID, dimID, docID, bytes.Repeat([]byte{0x73}, 32),
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`UPDATE hr_talent_assessments SET status = 'finalized', finalized_by_principal_id = ?, finalized_at = NOW() WHERE id = ?`,
		approverID, assessID,
	).Error; err != nil {
		t.Fatal(err)
	}

	effDate := time.Now().UTC().Truncate(24 * time.Hour)
	effProto := timestamppb.New(effDate)

	// Rekomendasi pertama: submit -> return.
	r1, err := client.CreatePromotionRecommendation(ctx, &pb.CreatePromotionRecommendationRequest{Input: &pb.PromotionRecommendationInput{
		RecommendationNumber: strPtr("PROMO-1-" + subjectWorkerID[:8]), WorkerId: &subjectWorkerID,
		SourceAssignmentId: &sourceAssignID, TargetPositionId: &targetPosID, AssessmentId: &assessID,
		RecommendedEffectiveDate: effProto, RationaleDocumentVersionReferenceId: &docID,
		RationaleDigest: bytes.Repeat([]byte{0x74}, 32), IdempotencyKey: strPtr("idem-rec-1-" + subjectWorkerID),
	}})
	if err != nil {
		t.Fatalf("create rec 1: %v", err)
	}
	recReturnedID = r1.GetValue().GetId()
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET primary_assignment_id = NULL WHERE id = ?", subjectWorkerID).Error; err != nil {
		t.Fatalf("clear current primary assignment before submission: %v", err)
	}
	if _, err := client.SubmitPromotionRecommendation(ctx, &common.IDRequest{Id: recReturnedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("submit promotion with stale source assignment = %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET primary_assignment_id = ? WHERE id = ?", sourceAssignID, subjectWorkerID).Error; err != nil {
		t.Fatalf("restore current primary assignment before submission: %v", err)
	}
	if _, err := client.SubmitPromotionRecommendation(ctx, &common.IDRequest{Id: recReturnedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("submit promotion with unapproved source assignment = %v", err)
	}
	if err := database.WithContext(ctx).Exec(
		`UPDATE hr_worker_assignments SET requested_at = NOW() - INTERVAL '1 hour', requested_by_principal_id = ?,
		  approved_at = NOW(), approved_by_principal_id = ?, employment_status_snapshot = 'active' WHERE id = ?`,
		recommenderID, approverID, sourceAssignID,
	).Error; err != nil {
		t.Fatalf("record source assignment approval in fixture: %v", err)
	}
	if sub1, err := client.SubmitPromotionRecommendation(ctx, &common.IDRequest{Id: recReturnedID}); err != nil || sub1.GetValue().GetStatus() != "submitted" {
		t.Fatalf("submit rec 1: res=%+v err=%v", sub1, err)
	}
	if _, err := client.ApprovePromotionRecommendation(ctx, &pb.DecidePromotionRecommendationRequest{
		RecommendationId: recReturnedID, RationaleDocumentVersionReferenceId: docID,
		RationaleDigest: bytes.Repeat([]byte{0x75}, 32), IdempotencyKey: "idem-self-app-" + recReturnedID,
	}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self approve promo = %v, want PermissionDenied", status.Code(err))
	}

	policy.setPrincipal(approverID)
	ret1, err := client.ReturnPromotionRecommendation(ctx, &pb.DecidePromotionRecommendationRequest{
		RecommendationId: recReturnedID, RationaleDocumentVersionReferenceId: docID,
		RationaleDigest: bytes.Repeat([]byte{0x76}, 32), IdempotencyKey: "idem-ret-1-" + recReturnedID,
	})
	if err != nil || ret1.GetValue().GetStatus() != "returned" {
		t.Fatalf("return rec 1: res=%+v err=%v", ret1, err)
	}
	if err := database.WithContext(ctx).Table("hr_promotion_decisions").Where("recommendation_id = ?", recReturnedID).Select("id").Scan(&decReturnID).Error; err != nil || decReturnID == "" {
		t.Fatalf("load return decision id: %v", err)
	}

	// Revisi menggantikan rekomendasi dikembalikan: submit -> approve -> apply atomik.
	policy.setPrincipal(recommenderID)
	r2, err := client.CreatePromotionRecommendation(ctx, &pb.CreatePromotionRecommendationRequest{Input: &pb.PromotionRecommendationInput{
		RecommendationNumber: strPtr("PROMO-2-" + subjectWorkerID[:8]), WorkerId: &subjectWorkerID,
		SourceAssignmentId: &sourceAssignID, TargetPositionId: &targetPosID, AssessmentId: &assessID,
		SupersedesRecommendationId: &recReturnedID, RecommendedEffectiveDate: effProto,
		RationaleDocumentVersionReferenceId: &docID, RationaleDigest: bytes.Repeat([]byte{0x77}, 32),
		IdempotencyKey: strPtr("idem-rec-2-" + subjectWorkerID),
	}})
	if err != nil {
		t.Fatalf("create revision rec 2: %v", err)
	}
	recAppliedID = r2.GetValue().GetId()
	if _, err := client.SubmitPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID}); err != nil {
		t.Fatalf("submit rec 2: %v", err)
	}

	policy.setPrincipal(approverID)
	approval := &pb.DecidePromotionRecommendationRequest{
		RecommendationId: recAppliedID, RationaleDocumentVersionReferenceId: docID,
		RationaleDigest: bytes.Repeat([]byte{0x78}, 32), IdempotencyKey: "idem-app-2-" + recAppliedID,
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET primary_assignment_id = NULL WHERE id = ?", subjectWorkerID).Error; err != nil {
		t.Fatalf("clear current primary assignment before approval: %v", err)
	}
	if _, err := client.ApprovePromotionRecommendation(ctx, approval); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve promotion with stale source assignment = %v", err)
	}
	var decisionCount int64
	if err := database.WithContext(ctx).Table("hr_promotion_decisions").Where("recommendation_id = ?", recAppliedID).
		Count(&decisionCount).Error; err != nil || decisionCount != 0 {
		t.Fatalf("failed approval left %d decisions: %v", decisionCount, err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET primary_assignment_id = ? WHERE id = ?", sourceAssignID, subjectWorkerID).Error; err != nil {
		t.Fatalf("restore current primary assignment before approval: %v", err)
	}
	app2, err := client.ApprovePromotionRecommendation(ctx, approval)
	if err != nil || app2.GetValue().GetStatus() != "approved" || app2.GetValue().GetApprovalDecisionId() == "" {
		t.Fatalf("approve rec 2: res=%+v err=%v", app2, err)
	}
	decApproveID = app2.GetValue().GetApprovalDecisionId()

	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET termination_date = ? WHERE id = ?", effDate, recommenderWorkerID).Error; err != nil {
		t.Fatalf("set manager termination date: %v", err)
	}
	if _, err := client.ApplyPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply promotion with exiting manager = %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET termination_date = NULL WHERE id = ?", recommenderWorkerID).Error; err != nil {
		t.Fatalf("restore manager employment period: %v", err)
	}

	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET primary_assignment_id = NULL WHERE id = ?", subjectWorkerID).Error; err != nil {
		t.Fatalf("clear current primary assignment: %v", err)
	}
	if _, err := client.ApplyPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply promotion without current source assignment = %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_workers SET primary_assignment_id = ? WHERE id = ?", sourceAssignID, subjectWorkerID).Error; err != nil {
		t.Fatalf("restore current primary assignment: %v", err)
	}

	futureAssignmentID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_worker_assignments (id, worker_id, position_id, assignment_type, fte_percent, valid_from,
		 employment_status_snapshot, requested_by_principal_id, requested_at, approved_by_principal_id, approved_at)
		 VALUES (?, ?, ?, 'secondary', 100.00, ?, 'active', ?, NOW() - INTERVAL '1 hour', ?, NOW())`,
		futureAssignmentID, approverWorkerID, targetPosID, effDate.AddDate(0, 0, 1), recommenderID, approverID,
	).Error; err != nil {
		t.Fatalf("create future approved assignment: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_worker_assignments WHERE id = ?", futureAssignmentID).Error; err != nil {
			t.Errorf("clean future assignment: %v", err)
		}
	})
	if _, err := client.ApplyPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply promotion with future capacity conflict = %v", err)
	}
	if err := database.WithContext(ctx).Exec(
		"UPDATE hr_worker_assignments SET worker_id = ?, position_id = ? WHERE id = ?", subjectWorkerID, sourcePosID, futureAssignmentID,
	).Error; err != nil {
		t.Fatalf("move reservation to worker on a different position: %v", err)
	}
	if _, err := client.ApplyPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply promotion with future worker FTE conflict = %v", err)
	}
	if err := database.WithContext(ctx).Exec(
		"UPDATE hr_worker_assignments SET worker_id = ?, position_id = ? WHERE id = ?", approverWorkerID, targetPosID, futureAssignmentID,
	).Error; err != nil {
		t.Fatalf("restore target position reservation: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET valid_to = ? WHERE id = ?", effDate.AddDate(0, 0, -1), targetPosID).Error; err != nil {
		t.Fatalf("expire target position: %v", err)
	}
	if _, err := client.ApplyPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply promotion to expired position = %v", err)
	}
	targetEnd := effDate.AddDate(0, 0, 30)
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET valid_to = ? WHERE id = ?", targetEnd, targetPosID).Error; err != nil {
		t.Fatalf("set target position end: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_worker_assignments SET valid_from = ? WHERE id = ?", targetEnd, futureAssignmentID).Error; err != nil {
		t.Fatalf("move reservation to last day: %v", err)
	}
	if _, err := client.ApplyPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply promotion with reservation on inclusive end date = %v", err)
	}
	// Fixture meniru reservasi lama yang tersisa setelah periode posisi diperpendek.
	if err := database.WithContext(ctx).Exec("UPDATE hr_worker_assignments SET valid_from = ? WHERE id = ?", targetEnd.AddDate(0, 0, 1), futureAssignmentID).Error; err != nil {
		t.Fatalf("move reservation after target expiry: %v", err)
	}

	var outboxBeforeApply int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", recAppliedID).
		Count(&outboxBeforeApply).Error; err != nil {
		t.Fatalf("count promotion outbox before apply: %v", err)
	}
	verifier.reject.Store(true)
	if _, err := client.ApplyPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply promotion with target rejected by Foundation = %v", err)
	}
	var primaryBeforeRetry string
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", subjectWorkerID).
		Select("primary_assignment_id").Scan(&primaryBeforeRetry).Error; err != nil || primaryBeforeRetry != sourceAssignID {
		t.Fatalf("failed promotion changed primary assignment: %q, %v", primaryBeforeRetry, err)
	}
	for _, check := range []struct {
		table, query string
		args         []any
		want         int64
	}{
		{"hr_worker_assignments", "id = ? AND valid_to IS NULL", []any{sourceAssignID}, 1},
		{"hr_worker_assignments", "worker_id = ? AND position_id = ?", []any{subjectWorkerID, targetPosID}, 0},
		{"hr_worker_lifecycle_events", "worker_id = ? AND source_reference = ?", []any{subjectWorkerID, recAppliedID}, 0},
		{"hr_promotion_recommendations", "id = ? AND status = 'approved' AND applied_assignment_id IS NULL", []any{recAppliedID}, 1},
		{"hr_outbox_events", "aggregate_id = ?", []any{recAppliedID}, outboxBeforeApply},
	} {
		var count int64
		if err := database.WithContext(ctx).Table(check.table).Where(check.query, check.args...).Count(&count).Error; err != nil || count != check.want {
			t.Fatalf("failed promotion %s (%s) count = %d, want %d: %v", check.table, check.query, count, check.want, err)
		}
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET position_status = 'inactive' WHERE id = ?", targetPosID).Error; err != nil {
		t.Fatalf("deactivate promotion target: %v", err)
	}
	if _, err := client.ApplyPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("apply promotion to inactive target: %v", err)
	}
	if err := database.WithContext(ctx).Exec("UPDATE hr_positions SET position_status = 'active' WHERE id = ?", targetPosID).Error; err != nil {
		t.Fatalf("restore promotion target: %v", err)
	}
	verifier.reject.Store(false)
	policy.setPrincipal(subjectPrincipalID)
	if _, err := client.ApplyPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("worker applied own promotion: %v", err)
	}
	policy.setPrincipal(approverID)
	applied, err := client.ApplyPromotionRecommendation(ctx, &common.IDRequest{Id: recAppliedID})
	if err != nil || applied.GetValue().GetStatus() != "applied" ||
		applied.GetValue().GetAppliedAssignmentId() == "" || applied.GetValue().GetAppliedLifecycleEventId() <= 0 {
		t.Fatalf("apply rec 2: res=%+v err=%v", applied, err)
	}
	appliedAssignID = applied.GetValue().GetAppliedAssignmentId()
	var assignmentEnd time.Time
	if err := database.WithContext(ctx).Table("hr_worker_assignments").Where("id = ?", appliedAssignID).
		Select("valid_to").Scan(&assignmentEnd).Error; err != nil || assignmentEnd.Format("2006-01-02") != targetEnd.Format("2006-01-02") {
		t.Fatalf("promotion assignment end = %v, want %v, err=%v", assignmentEnd, targetEnd, err)
	}

	// Verifikasi bukti DB pasca-apply (source assignment ditutup h-1, assignment baru primary, lifecycle event).
	var primaryAssignID string
	if err := database.WithContext(ctx).Table("hr_workers").Where("id = ?", subjectWorkerID).Select("primary_assignment_id").Scan(&primaryAssignID).Error; err != nil || primaryAssignID != appliedAssignID {
		t.Fatalf("worker primary assignment = %q, want %q err=%v", primaryAssignID, appliedAssignID, err)
	}
	var sourceValidTo time.Time
	if err := database.WithContext(ctx).Table("hr_worker_assignments").Where("id = ?", sourceAssignID).Select("valid_to").Scan(&sourceValidTo).Error; err != nil ||
		sourceValidTo.Format("2006-01-02") != effDate.AddDate(0, 0, -1).Format("2006-01-02") {
		t.Fatalf("source valid_to = %v, want %s err=%v", sourceValidTo, effDate.AddDate(0, 0, -1).Format("2006-01-02"), err)
	}
	var eventType, sourceRef string
	if err := database.WithContext(ctx).Table("hr_worker_lifecycle_events").
		Where("id = ?", applied.GetValue().GetAppliedLifecycleEventId()).
		Select("event_type", "source_reference").Row().Scan(&eventType, &sourceRef); err != nil || eventType != "promotion" || sourceRef != recAppliedID {
		t.Fatalf("lifecycle event = (%s, %s), want (promotion, %s) err=%v", eventType, sourceRef, recAppliedID, err)
	}

	// Usulan ketiga untuk jalur RejectPromotionRecommendation (pakai appliedAssignID sbg source ke sourcePosID).
	policy.setPrincipal(recommenderID)
	r3, err := client.CreatePromotionRecommendation(ctx, &pb.CreatePromotionRecommendationRequest{Input: &pb.PromotionRecommendationInput{
		RecommendationNumber: strPtr("PROMO-3-" + subjectWorkerID[:8]), WorkerId: &subjectWorkerID,
		SourceAssignmentId: &appliedAssignID, TargetPositionId: &sourcePosID, AssessmentId: &assessID,
		RecommendedEffectiveDate:            timestamppb.New(effDate.AddDate(0, 0, 7)),
		RationaleDocumentVersionReferenceId: &docID, RationaleDigest: bytes.Repeat([]byte{0x79}, 32),
		IdempotencyKey: strPtr("idem-rec-3-" + subjectWorkerID),
	}})
	if err != nil {
		t.Fatalf("create rec 3: %v", err)
	}
	recRejectedID = r3.GetValue().GetId()
	if _, err := client.SubmitPromotionRecommendation(ctx, &common.IDRequest{Id: recRejectedID}); err != nil {
		t.Fatalf("submit rec 3: %v", err)
	}
	policy.setPrincipal(approverID)
	if rej3, err := client.RejectPromotionRecommendation(ctx, &pb.DecidePromotionRecommendationRequest{
		RecommendationId: recRejectedID, RationaleDocumentVersionReferenceId: docID,
		RationaleDigest: bytes.Repeat([]byte{0x7a}, 32), IdempotencyKey: "idem-rej-3-" + recRejectedID,
	}); err != nil || rej3.GetValue().GetStatus() != "rejected" {
		t.Fatalf("reject rec 3: res=%+v err=%v", rej3, err)
	}
	if err := database.WithContext(ctx).Table("hr_promotion_decisions").Where("recommendation_id = ?", recRejectedID).Select("id").Scan(&decRejectID).Error; err != nil || decRejectID == "" {
		t.Fatalf("load reject decision id: %v", err)
	}
}
