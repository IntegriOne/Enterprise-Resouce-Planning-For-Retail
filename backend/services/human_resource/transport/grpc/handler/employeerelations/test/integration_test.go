//go:build integration

package employeerelations_test

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
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/employeerelations/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/employeerelations"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/employeerelations"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type erTestPolicy struct {
	mu          sync.Mutex
	principalID string
}

func (p *erTestPolicy) currentPrincipal() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.principalID
}

func (p *erTestPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.currentPrincipal(), SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*erTestPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

type erFoundationStub struct{}

func (erFoundationStub) VerifyLegalEntity(context.Context, string) error { return nil }
func (erFoundationStub) VerifyPositionReferences(context.Context, security.PositionReferenceCheck) error {
	return nil
}
func (erFoundationStub) VerifyWorkerReferences(context.Context, security.WorkerReferenceCheck) error {
	return nil
}
func (erFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error { return nil }
func (erFoundationStub) VerifyLocation(context.Context, string, string) error         { return nil }
func (erFoundationStub) VerifyParty(context.Context, string, bool) error              { return nil }

func text(v string) *string { return &v }

func TestEmployeeRelationsRPCWithPostgreSQL(t *testing.T) {
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
	validators, err := security.EmployeeRelationsValidators(repos, erFoundationStub{})
	if err != nil {
		t.Fatal(err)
	}

	officerID := "00000000-0000-4000-8000-000000000061"
	policy := &erTestPolicy{principalID: officerID}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(validators),
		},
		EmployeeRelationsCase:  repos.EmployeeRelationsCase,
		EmployeeRelationsEvent: repos.EmployeeRelationsEvent,
		CollectiveAgreement:    repos.CollectiveAgreement,
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
	conn, err := grpc.NewClient("passthrough:///hrm-er-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewEmployeeRelationsServiceClient(conn)

	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	personID, _ := ids.NewID()
	workerID, _ := ids.NewID()
	unionPartyID, _ := ids.NewID()
	docID, _ := ids.NewID()
	summaryDocID, _ := ids.NewID()
	var agreementID, caseID, draftCaseID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if caseID != "" {
			if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_employee_relations_events WHERE case_id = ?", caseID).Error; err != nil {
				t.Errorf("cleanup er events: %v", err)
			}
		}
		for _, stmt := range []struct {
			q string
			v string
		}{
			{"DELETE FROM hr_employee_relations_cases WHERE id = ?", caseID},
			{"DELETE FROM hr_employee_relations_cases WHERE id = ?", draftCaseID},
			{"DELETE FROM hr_collective_agreements WHERE id = ?", agreementID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
		} {
			if stmt.v != "" {
				if err := database.WithContext(cleanupCtx).Exec(stmt.q, stmt.v).Error; err != nil {
					t.Errorf("cleanup %s: %v", stmt.q, err)
				}
			}
		}
		for _, agg := range []string{caseID, draftCaseID, agreementID} {
			if agg != "" {
				if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", agg).Error; err != nil {
					t.Errorf("cleanup outbox %s: %v", agg, err)
				}
			}
		}
	})

	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
		 primary_legal_entity_reference_id, hire_date, service_start_date)
		 VALUES (?, ?, ?, 'employee', 'active', ?, '2026-01-01', '2026-01-01')`,
		workerID, "ER-W-"+workerID, personID, legalID,
	).Error; err != nil {
		t.Fatal(err)
	}

	agVer := int32(1)
	nowUTC := time.Now().UTC()
	agInput := &pb.CollectiveAgreementInput{
		AgreementCode: text("PKB-" + legalID[:8]), AgreementVersion: &agVer,
		LegalEntityReferenceId: &legalID, RepresentativePartyReferenceId: &unionPartyID,
		JurisdictionCode: text("ID"), EffectiveFrom: timestamppb.New(nowUTC.AddDate(0, -1, 0)),
		EffectiveTo: timestamppb.New(nowUTC.AddDate(1, 0, 0)), SignedAt: timestamppb.New(nowUTC.AddDate(0, -1, 0)),
		DocumentVersionReferenceId: &docID, ContentDigest: bytes.Repeat([]byte{0x11}, 32),
	}
	agRes, err := client.CreateCollectiveAgreement(ctx, &pb.CreateCollectiveAgreementRequest{Input: agInput})
	if err != nil {
		t.Fatalf("create collective agreement: %v", err)
	}
	agreementID = agRes.GetValue().GetId()
	if _, err := client.CreateCollectiveAgreement(ctx, &pb.CreateCollectiveAgreementRequest{Input: agInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate collective agreement = %v, want AlreadyExists", status.Code(err))
	}

	if _, err := client.CreateEmployeeRelationsCase(ctx, &pb.CreateEmployeeRelationsCaseRequest{Input: &pb.EmployeeRelationsCaseInput{
		CaseNumber: text("ER-BAD-" + workerID[:8]), CaseType: text("unsupported"), Status: text("open"),
		RestrictedSummaryReferenceId: &summaryDocID, ConfidentialityClass: text("restricted"),
		IdempotencyKey: text("idem-er-bad-" + workerID),
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid case type = %v, want InvalidArgument", status.Code(err))
	}

	caseRes, err := client.CreateEmployeeRelationsCase(ctx, &pb.CreateEmployeeRelationsCaseRequest{Input: &pb.EmployeeRelationsCaseInput{
		CaseNumber: text("ER-CASE-" + workerID[:8]), CaseType: text("grievance"), Status: text("open"),
		SubjectWorkerId: &workerID, ReportingPartyReferenceId: &unionPartyID,
		RestrictedSummaryReferenceId: &summaryDocID, ConfidentialityClass: text("restricted"),
		IdempotencyKey: text("idem-er-1-" + workerID),
	}})
	if err != nil {
		t.Fatalf("create ER case: %v", err)
	}
	caseID = caseRes.GetValue().GetId()

	if _, err := client.UpdateEmployeeRelationsCase(ctx, &pb.UpdateEmployeeRelationsCaseRequest{
		Id: caseID, Input: &pb.EmployeeRelationsCaseInput{DueAt: timestamppb.New(nowUTC.Add(-time.Hour))},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"due_at"}},
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("case deadline before opening = %v, want InvalidArgument", err)
	}

	if _, err := client.UpdateEmployeeRelationsCase(ctx, &pb.UpdateEmployeeRelationsCaseRequest{
		Id: caseID, Input: &pb.EmployeeRelationsCaseInput{DueAt: timestamppb.New(nowUTC.Add(72 * time.Hour))},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"due_at"}},
	}); err != nil {
		t.Fatalf("update open ER case: %v", err)
	}

	if _, err := client.CreateEmployeeRelationsEvent(ctx, &pb.CreateEmployeeRelationsEventRequest{Input: &pb.EmployeeRelationsEventInput{
		CaseId: &caseID, EventType: text("resolved"), EventDigest: bytes.Repeat([]byte{0x22}, 32),
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("record case resolution through generic event RPC = %v", err)
	}
	evRes, err := client.CreateEmployeeRelationsEvent(ctx, &pb.CreateEmployeeRelationsEventRequest{Input: &pb.EmployeeRelationsEventInput{
		CaseId: &caseID, EventType: text("evidence_logged"),
		EventDigest: bytes.Repeat([]byte{0x22}, 32), EventReferenceId: &docID,
	}})
	if err != nil || evRes.GetValue().GetActorPrincipalId() != officerID {
		t.Fatalf("create ER event: res=%+v err=%v", evRes, err)
	}

	if _, err := client.UpdateEmployeeRelationsCase(ctx, &pb.UpdateEmployeeRelationsCaseRequest{
		Id: caseID, Input: &pb.EmployeeRelationsCaseInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"subject_worker_id"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("remove subject from case with history = %v, want FailedPrecondition", err)
	}

	if _, err := client.DeleteEmployeeRelationsEvent(ctx, &common.IdentityIDRequest{Id: evRes.GetValue().GetId()}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete append-only ER event = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteEmployeeRelationsCase(ctx, &common.IDRequest{Id: caseID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete ER case with event = %v, want FailedPrecondition", status.Code(err))
	}

	// Kasus terbuka tanpa event masih boleh dihapus.
	emptyCase, err := client.CreateEmployeeRelationsCase(ctx, &pb.CreateEmployeeRelationsCaseRequest{Input: &pb.EmployeeRelationsCaseInput{
		CaseNumber: text("ER-TMP-" + workerID[:8]), CaseType: text("accommodation"), Status: text("open"),
		SubjectWorkerId: &workerID, RestrictedSummaryReferenceId: &summaryDocID,
		ConfidentialityClass: text("highly_restricted"), IdempotencyKey: text("idem-er-tmp-" + workerID),
	}})
	if err != nil {
		t.Fatalf("create empty ER case: %v", err)
	}
	draftCaseID = emptyCase.GetValue().GetId()
	if _, err := client.DeleteEmployeeRelationsCase(ctx, &common.IDRequest{Id: draftCaseID}); err != nil {
		t.Fatalf("delete empty ER case: %v", err)
	}
	draftCaseID = ""
}
