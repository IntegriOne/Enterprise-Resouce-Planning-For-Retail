//go:build integration

package test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/workforceplanning/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	appservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/workforceplanning"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/workforceplanning"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type wfPlanFoundationStub struct {
	security.PositionReferenceVerifier
	leErr          error
	ouErr          error
	rejectPosition atomic.Bool
}

func (s *wfPlanFoundationStub) VerifyPositionReferences(context.Context, security.PositionReferenceCheck) error {
	if s.rejectPosition.Load() {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}
func (s *wfPlanFoundationStub) VerifyWorkerReferences(context.Context, security.WorkerReferenceCheck) error {
	return nil
}
func (s *wfPlanFoundationStub) VerifyLegalEntity(context.Context, string) error { return s.leErr }
func (s *wfPlanFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return s.ouErr
}
func (s *wfPlanFoundationStub) VerifyLocation(context.Context, string, string) error { return nil }
func (s *wfPlanFoundationStub) VerifyParty(context.Context, string, bool) error      { return nil }

type workforceApprovalPolicy struct{ positionPolicy }

func (p workforceApprovalPolicy) Authorize(ctx context.Context, access operation.Access) (operation.Actor, error) {
	actor, err := p.positionPolicy.Authorize(ctx, access)
	if md, ok := metadata.FromIncomingContext(ctx); ok && len(md.Get("x-test-approver")) == 1 {
		actor.PrincipalID = "00000000-0000-4000-8000-000000000003"
	}
	return actor, err
}

const approvalWorkLocationID = "9c243a6f-bcd7-4055-a73e-b02db4683ba6"

type workforceBudgetStub struct{ reject atomic.Bool }

func (s *workforceBudgetStub) VerifyWorkforceBudget(_ context.Context, request service.WorkforceBudgetCheck) (string, error) {
	if s.reject.Load() || request.PlanVersion != 2 || len(request.Lines) != 1 ||
		request.LegalEntityID == "" || request.Lines[0].JobID == "" {
		return "", hrerrors.ErrReferenceConflict
	}
	line := request.Lines[0]
	if line.WorkerCategory != "employee" || line.ActionCode != "recruit" ||
		line.WorkLocationID == nil || *line.WorkLocationID != approvalWorkLocationID {
		return "", hrerrors.ErrReferenceConflict
	}
	return request.PlanID, nil
}

func TestWorkforcePlanAndLineRPCWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("integration test requires a valid disposable HRM URL")
	}
	dbName := strings.TrimPrefix(parsed.Path, "/")
	if parsed.Scheme != "postgres" ||
		!strings.HasPrefix(dbName, "hrm_disposable_") || !strings.HasSuffix(dbName, "_test") ||
		os.Getenv("HRM_TEST_DATABASE_DISPOSABLE") != "1" {
		t.Fatal("integration test requires a disposable HRM test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
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
	foundation := &wfPlanFoundationStub{}
	validators, err := security.WorkforcePlanValidators(repos, foundation)
	if err != nil {
		t.Fatal(err)
	}
	budget := &workforceBudgetStub{}
	registry, err := appservice.NewRegistry(repos, operation.Dependencies{
		Policy: workforceApprovalPolicy{}, Integrity: security.NewIntegrity(validators),
	}, appservice.WithWorkforceBudgetVerifier(budget))
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := handler.Register(server, registry.WorkforcePlanning); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-wfplan-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewWorkforcePlanningServiceClient(conn)

	ids := operation.UUIDGenerator{}
	familyID, _ := ids.NewID()
	jobID, _ := ids.NewID()
	legalID, _ := ids.NewID()
	unitID, _ := ids.NewID()
	var planID, lineID, approvalPlanID, approvalLineID, replacementID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, stmt := range []struct {
			q string
			v string
		}{
			{"DELETE FROM hr_workforce_plan_lines WHERE id = ?", approvalLineID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", approvalLineID},
			{"DELETE FROM hr_workforce_plans WHERE id = ?", approvalPlanID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", approvalPlanID},
			{"DELETE FROM hr_workforce_plans WHERE id = ?", replacementID},
			{"DELETE FROM hr_workforce_plan_lines WHERE id = ?", lineID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", lineID},
			{"DELETE FROM hr_workforce_plans WHERE id = ?", planID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", planID},
			{"DELETE FROM hr_jobs WHERE id = ?", jobID},
			{"DELETE FROM hr_job_families WHERE id = ?", familyID},
		} {
			if stmt.v == "" {
				continue
			}
			_ = database.WithContext(cleanupCtx).Exec(stmt.q, stmt.v).Error
		}
	})

	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, ?, ?)",
		familyID, "WF-FAM-"+familyID, "Operations", "operations",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, valid_from) VALUES (?, ?, ?, ?, ?)",
		jobID, familyID, "WF-JOB-"+jobID, "Planner", "2026-01-01",
	).Error; err != nil {
		t.Fatal(err)
	}

	planCode := "WFP-" + familyID
	planVersion := int32(1)
	periodStart, periodEnd := "2027-01-01", "2027-12-31"
	scenario, draftStatus := "baseline", "draft"
	planRes, err := client.CreateWorkforcePlan(ctx, &pb.CreateWorkforcePlanRequest{Input: &pb.WorkforcePlanInput{
		PlanCode: &planCode, PlanVersion: &planVersion, LegalEntityReferenceId: &legalID,
		PeriodStart: &periodStart, PeriodEnd: &periodEnd, ScenarioCode: &scenario, Status: &draftStatus,
	}})
	if err != nil {
		t.Fatalf("create workforce plan: %v", err)
	}
	planID = planRes.GetItem().GetId()

	// Duplikat kode + versi rencana pada entitas legal sama harus ditolak.
	if _, err := client.CreateWorkforcePlan(ctx, &pb.CreateWorkforcePlanRequest{Input: &pb.WorkforcePlanInput{
		PlanCode: &planCode, PlanVersion: &planVersion, LegalEntityReferenceId: &legalID,
		PeriodStart: &periodStart, PeriodEnd: &periodEnd, ScenarioCode: &scenario, Status: &draftStatus,
	}}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate workforce plan status = %v, want AlreadyExists", status.Code(err))
	}

	if _, err := client.SubmitWorkforcePlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("submit empty workforce plan = %v", err)
	}
	lineNum := int32(1)
	category, currentFTE, reqFTE := "employee", "2.000", "4.000"
	actionCode, rationale := "recruit", "Seasonal expansion"
	lineRes, err := client.CreateWorkforcePlanLine(ctx, &pb.CreateWorkforcePlanLineRequest{Input: &pb.WorkforcePlanLineInput{
		WorkforcePlanId: &planID, LineNumber: &lineNum, OrganizationUnitReferenceId: &unitID,
		JobId: &jobID, WorkerCategory: &category, CurrentFte: &currentFTE, RequiredFte: &reqFTE,
		NeededFrom: &periodStart, ActionCode: &actionCode, Rationale: &rationale,
	}})
	if err != nil {
		t.Fatalf("create workforce plan line: %v", err)
	}
	lineID = lineRes.GetItem().GetId()

	// Rencana yang sudah memiliki baris tidak dapat dihapus sebelum barisnya dihapus.
	if _, err := client.DeleteWorkforcePlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete referenced workforce plan = %v, want FailedPrecondition", status.Code(err))
	}
	submitRequest := &common.IDRequest{Id: planID}
	foundation.rejectPosition.Store(true)
	if _, err := client.SubmitWorkforcePlan(ctx, submitRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("submit with invalidated organization references = %v, want FailedPrecondition", err)
	}
	foundation.rejectPosition.Store(false)
	unchanged, err := client.GetWorkforcePlanByID(ctx, submitRequest)
	if err != nil || unchanged.GetItem().GetStatus() != "draft" {
		t.Fatalf("plan after rejected submission = %+v, %v", unchanged, err)
	}
	if err := database.WithContext(ctx).Table("hr_jobs").Where("id = ?", jobID).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.SubmitWorkforcePlan(ctx, submitRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("submit with inactive job = %v, want FailedPrecondition", err)
	}
	if err := database.WithContext(ctx).Table("hr_jobs").Where("id = ?", jobID).Update("is_active", true).Error; err != nil {
		t.Fatal(err)
	}
	submitted, err := client.SubmitWorkforcePlan(ctx, submitRequest)
	if err != nil || submitted.GetItem().GetStatus() != "review" ||
		submitted.GetItem().GetFinanceBudgetReferenceId() != "" || submitted.GetItem().GetApprovedAt() != nil {
		t.Fatalf("submit workforce plan: %+v, %v", submitted, err)
	}
	if _, err := client.SubmitWorkforcePlan(ctx, submitRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat workforce submission = %v", err)
	}
	if _, err := client.DeleteWorkforcePlanLine(ctx, &common.IDRequest{Id: lineID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete line during review = %v", err)
	}
	cancelled, err := client.CancelWorkforcePlan(ctx, submitRequest)
	if err != nil || cancelled.GetItem().GetStatus() != "cancelled" ||
		cancelled.GetItem().GetFinanceBudgetReferenceId() != "" || cancelled.GetItem().GetApprovedAt() != nil {
		t.Fatalf("cancel workforce plan: %+v, %v", cancelled, err)
	}
	if _, err := client.CancelWorkforcePlan(ctx, submitRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat workforce cancellation = %v", err)
	}
	if _, err := client.DeleteWorkforcePlanLine(ctx, &common.IDRequest{Id: lineID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete line after cancellation = %v", err)
	}
	var planEvents int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", planID).Count(&planEvents).Error; err != nil || planEvents != 3 {
		t.Fatalf("workforce plan events = %d, error = %v", planEvents, err)
	}

	verifyWorkforcePlanReadScope(t, ctx, database, repos, validators, planID, lineID, legalID, planCode)
	secondVersion := int32(2)
	approvalDraft, err := client.CreateWorkforcePlan(ctx, &pb.CreateWorkforcePlanRequest{Input: &pb.WorkforcePlanInput{
		PlanCode: &planCode, PlanVersion: &secondVersion, LegalEntityReferenceId: &legalID,
		PeriodStart: &periodStart, PeriodEnd: &periodEnd, ScenarioCode: &scenario,
	}})
	if err != nil {
		t.Fatalf("create draft for approval: %v", err)
	}
	approvalPlanID = approvalDraft.GetItem().GetId()
	approvalLine, err := client.CreateWorkforcePlanLine(ctx, &pb.CreateWorkforcePlanLineRequest{Input: &pb.WorkforcePlanLineInput{
		WorkforcePlanId: &approvalPlanID, LineNumber: &lineNum, OrganizationUnitReferenceId: &unitID,
		JobId: &jobID, WorkLocationReferenceId: new(approvalWorkLocationID), WorkerCategory: &category, CurrentFte: &currentFTE, RequiredFte: &reqFTE,
		NeededFrom: &periodStart, ActionCode: &actionCode, Rationale: &rationale,
	}})
	if err != nil {
		t.Fatalf("create approval line: %v", err)
	}
	approvalLineID = approvalLine.GetItem().GetId()
	approvalRequest := &common.IDRequest{Id: approvalPlanID}
	if _, err := client.SubmitWorkforcePlan(ctx, approvalRequest); err != nil {
		t.Fatalf("submit draft for approval: %v", err)
	}
	if _, err := client.ApproveWorkforcePlan(ctx, approvalRequest); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self approval = %v", err)
	}
	approverCtx := metadata.AppendToOutgoingContext(ctx, "x-test-approver", "true")
	budget.reject.Store(true)
	if _, err := client.ApproveWorkforcePlan(approverCtx, approvalRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approval with invalid budget = %v", err)
	}
	budget.reject.Store(false)
	approved, err := client.ApproveWorkforcePlan(approverCtx, approvalRequest)
	if err != nil || approved.GetItem().GetStatus() != "approved" ||
		approved.GetItem().GetFinanceBudgetReferenceId() != approvalPlanID || approved.GetItem().GetApprovedAt() == nil {
		t.Fatalf("approve workforce plan: %+v, %v", approved, err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", approvalPlanID).Count(&planEvents).Error; err != nil || planEvents != 3 {
		t.Fatalf("approved workforce plan events = %d, error = %v", planEvents, err)
	}

	replacementID, err = ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(`
		INSERT INTO hr_workforce_plans (id, plan_code, plan_version, legal_entity_reference_id, period_start,
		   period_end, scenario_code, status, created_by_principal_id, approved_at, approved_by_principal_id, finance_budget_reference_id)
		SELECT ?, plan_code, 3, legal_entity_reference_id, period_start, period_end, scenario_code,
		   'approved', created_by_principal_id, approved_at, approved_by_principal_id, ?
		FROM hr_workforce_plans WHERE id = ?`, replacementID, replacementID, approvalPlanID).Error; err != nil {
		t.Fatal(err)
	}
	supersede := &pb.SupersedeWorkforcePlanRequest{Id: approvalPlanID, ReplacementPlanId: planID, ReasonCode: "NEW_VERSION"}
	if _, err := client.SupersedeWorkforcePlan(approverCtx, supersede); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("supersede using cancelled version = %v", err)
	}
	supersede.ReplacementPlanId = replacementID
	superseded, err := client.SupersedeWorkforcePlan(approverCtx, supersede)
	if err != nil || superseded.GetItem().GetStatus() != "superseded" ||
		superseded.GetItem().GetSupersededByPlanId() != replacementID ||
		superseded.GetItem().GetSupersessionReasonCode() != supersede.ReasonCode ||
		superseded.GetItem().GetFinanceBudgetReferenceId() != approvalPlanID {
		t.Fatalf("supersede workforce plan: %+v, %v", superseded, err)
	}
	if _, err := client.SupersedeWorkforcePlan(approverCtx, supersede); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat workforce supersession = %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", approvalPlanID).Count(&planEvents).Error; err != nil || planEvents != 4 {
		t.Fatalf("superseded workforce plan events = %d, error = %v", planEvents, err)
	}

}

type workforceScopeSession struct {
	legalID string
}

func (s workforceScopeSession) ReadSessionContext(_ context.Context, _ string, now time.Time) (*authsession.SessionValidationContext, error) {
	return &authsession.SessionValidationContext{
		Gate:     &authsession.SecurityGateRecord{Open: true, RecoveryEpoch: 1},
		Token:    &authsession.SessionTokenRecord{UserID: "00000000-0000-4000-8000-000000000001", SessionID: "00000000-0000-4000-8000-000000000002", UserKind: "employee"},
		Security: &authsession.UserSecurityRecord{UserID: "00000000-0000-4000-8000-000000000001", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1},
		Snapshot: &authz.AccessSnapshot{
			PrincipalID: "00000000-0000-4000-8000-000000000001", UserKind: "employee", Status: "active", AuthEpoch: 1, PermRevision: 2, RecoveryEpoch: 1,
			SyncedAtUnix: now.Unix(), FreshUntilUnix: now.Add(30 * time.Second).Unix(),
			Permissions: []authz.PermissionGrant{
				{Code: "human_resource.workforce_plans.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: s.legalID}},
				{Code: "human_resource.workforce_plan_lines.read", Scope: authz.ScopeGrant{Type: "legal_entity", ID: s.legalID}},
			},
		},
	}, nil
}

func verifyWorkforcePlanReadScope(t *testing.T, ctx context.Context, db *gorm.DB, repos *repository.Registry, validators map[string]security.MutationValidator, planID, lineID, legalID, planCode string) {
	t.Helper()
	otherID, err := (operation.UUIDGenerator{}).NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(`
		INSERT INTO hr_workforce_plans (id, plan_code, plan_version, legal_entity_reference_id, period_start,
		 period_end, scenario_code, status, created_by_principal_id)
		VALUES (?, ?, 1, ?, '2027-01-01', '2027-12-31', 'baseline', 'draft', ?)`,
		otherID, planCode, otherID, otherID).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.WithContext(cleanup).Exec("DELETE FROM hr_workforce_plan_lines WHERE id = ?", otherID).Error; err != nil {
			t.Errorf("cleanup out-of-scope workforce line: %v", err)
		}
		if err := db.WithContext(cleanup).Exec("DELETE FROM hr_workforce_plans WHERE id = ?", otherID).Error; err != nil {
			t.Errorf("cleanup out-of-scope workforce plan: %v", err)
		}
	})

	if err := db.WithContext(ctx).Exec(`
		INSERT INTO hr_workforce_plan_lines (id, workforce_plan_id, line_number, organization_unit_reference_id,
		  job_id, worker_category, current_fte, required_fte, needed_from, action_code, rationale)
		SELECT ?, ?, line_number, organization_unit_reference_id, job_id, worker_category,
		  current_fte, required_fte, needed_from, action_code, rationale
		FROM hr_workforce_plan_lines WHERE id = ?`, otherID, otherID, lineID).Error; err != nil {
		t.Fatal(err)
	}

	policy, err := security.NewPolicy(workforceScopeSession{legalID: legalID}, security.PermissionCatalog())
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(validators),
		},
		JobFamily: repos.JobFamily, Job: repos.Job, Position: repos.Position,
		WorkforcePlan: repos.WorkforcePlan, WorkforcePlanLine: repos.WorkforcePlanLine,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		ctx = authcontext.WithCaller(ctx, authcontext.CallerContext{
			Authenticated: true, PrincipalID: "00000000-0000-4000-8000-000000000001", SessionID: "00000000-0000-4000-8000-000000000002", UserKind: "employee", AuthEpoch: 1, PermRevision: 2,
		})
		return next(ctx, req)
	}))
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-wfplan-scope",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewWorkforcePlanningServiceClient(conn)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-ig-session-token-hash", strings.Repeat("a", 64))
	item, err := client.GetWorkforcePlanByID(ctx, &common.IDRequest{Id: planID})
	if err != nil || item.GetItem().GetId() != planID {
		t.Fatalf("read workforce plan within grant: %+v, %v", item, err)
	}
	if _, err := client.GetWorkforcePlanByID(ctx, &common.IDRequest{Id: otherID}); status.Code(err) != codes.NotFound {
		t.Fatalf("read workforce plan outside grant = %v", err)
	}
	list, err := client.GetAllWorkforcePlans(ctx, &common.PageRequest{Limit: 1})
	if err != nil || len(list.GetItems()) != 1 || list.GetItems()[0].GetId() != planID || list.GetPage().GetTotal() != 1 {
		t.Fatalf("scoped workforce list: %+v, %v", list, err)
	}
	found, err := client.FindWorkforcePlanByPlanCode(ctx, &common.StringFinderRequest{Value: &planCode, Limit: 1})
	if err != nil || len(found.GetItems()) != 1 || found.GetItems()[0].GetId() != planID || found.GetPage().GetTotal() != 1 {
		t.Fatalf("scoped workforce finder: %+v, %v", found, err)
	}
	nextPage, err := client.GetAllWorkforcePlans(ctx, &common.PageRequest{Limit: 1, Offset: 1})
	if err != nil || len(nextPage.GetItems()) != 0 || nextPage.GetPage().GetTotal() != 1 {
		t.Fatalf("scoped workforce second page: %+v, %v", nextPage, err)
	}
	line, err := client.GetWorkforcePlanLineByID(ctx, &common.IDRequest{Id: lineID})
	if err != nil || line.GetItem().GetId() != lineID {
		t.Fatalf("read workforce line within grant: %+v, %v", line, err)
	}
	if _, err := client.GetWorkforcePlanLineByID(ctx, &common.IDRequest{Id: otherID}); status.Code(err) != codes.NotFound {
		t.Fatalf("read workforce line outside grant = %v", err)
	}
	lines, err := client.GetAllWorkforcePlanLines(ctx, &common.PageRequest{Limit: 1})
	if err != nil || len(lines.GetItems()) != 1 || lines.GetItems()[0].GetId() != lineID || lines.GetPage().GetTotal() != 1 {
		t.Fatalf("scoped workforce line list: %+v, %v", lines, err)
	}
	outsideLines, err := client.FindWorkforcePlanLineByWorkforcePlanID(ctx, &common.StringFinderRequest{Value: &otherID, Limit: 1})
	if err != nil || len(outsideLines.GetItems()) != 0 || outsideLines.GetPage().GetTotal() != 0 {
		t.Fatalf("scoped workforce line finder: %+v, %v", outsideLines, err)
	}

}
