//go:build integration

package policy_test

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
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/policy/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/policy"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/policy"
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

type policyTestAuth struct {
	mu          sync.Mutex
	principalID string
}

func (p *policyTestAuth) setPrincipal(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.principalID = id
}

func (p *policyTestAuth) currentPrincipal() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.principalID
}

func (p *policyTestAuth) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.currentPrincipal(), SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*policyTestAuth) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

type policyFoundationStub struct{}

func (policyFoundationStub) VerifyLegalEntity(context.Context, string) error { return nil }
func (policyFoundationStub) VerifyPositionReferences(context.Context, security.PositionReferenceCheck) error {
	return nil
}
func (policyFoundationStub) VerifyWorkerReferences(context.Context, security.WorkerReferenceCheck) error {
	return nil
}
func (policyFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return nil
}
func (policyFoundationStub) VerifyLocation(context.Context, string, string) error { return nil }
func (policyFoundationStub) VerifyParty(context.Context, string, bool) error      { return nil }

func text(v string) *string { return &v }

func TestPolicyAndAcknowledgementRPCWithPostgreSQL(t *testing.T) {
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
	validators, err := security.PolicyValidators(repos, policyFoundationStub{})
	if err != nil {
		t.Fatal(err)
	}

	adminID := "00000000-0000-4000-8000-000000000071"
	workerPrincipalID := "00000000-0000-4000-8000-000000000072"
	otherID := "00000000-0000-4000-8000-000000000073"
	policyAuth := &policyTestAuth{principalID: adminID}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policyAuth, Integrity: security.NewIntegrity(validators),
		},
		Policy:                repos.Policy,
		PolicyAcknowledgement: repos.PolicyAcknowledgement,
		Worker:                repos.Worker,
	})
	if err != nil {
		t.Fatal(err)
	}

	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		return next(authcontext.WithCaller(ctx, authcontext.CallerContext{PrincipalID: policyAuth.currentPrincipal(), Authenticated: true}), req)
	}))
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-policy-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewPolicyServiceClient(conn)

	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	personID, _ := ids.NewID()
	workerID, _ := ids.NewID()
	docID, _ := ids.NewID()
	evidenceDocID, _ := ids.NewID()
	var policyID, ackID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, stmt := range []struct {
			q string
			v string
		}{
			{"DELETE FROM hr_policy_acknowledgements WHERE id = ?", ackID},
			{"DELETE FROM hr_policies WHERE id = ?", policyID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
		} {
			if stmt.v != "" {
				if err := database.WithContext(cleanupCtx).Exec(stmt.q, stmt.v).Error; err != nil {
					t.Errorf("cleanup %s: %v", stmt.q, err)
				}
			}
		}
		for _, agg := range []string{ackID, policyID} {
			if agg != "" {
				if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", agg).Error; err != nil {
					t.Errorf("cleanup outbox %s: %v", agg, err)
				}
			}
		}
	})

	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status,
		 primary_legal_entity_reference_id, iam_principal_reference_id, hire_date, service_start_date)
		 VALUES (?, ?, ?, 'employee', 'active', ?, ?, '2026-01-01', '2026-01-01')`,
		workerID, "POL-W-"+workerID, personID, legalID, workerPrincipalID,
	).Error; err != nil {
		t.Fatal(err)
	}

	ver := int32(1)
	nowUTC := time.Now().UTC()
	polInput := &pb.PolicyInput{
		PolicyCode: text("POL-SEC-" + legalID[:8]), PolicyVersion: &ver, PolicyName: text("Security Policy"),
		CategoryCode: text("conduct"), LegalEntityReferenceId: &legalID,
		EffectiveFrom:              timestamppb.New(nowUTC.AddDate(0, -1, 0)),
		DocumentVersionReferenceId: &docID, ContentDigest: bytes.Repeat([]byte{0x31}, 32),
	}
	polRes, err := client.CreatePolicy(ctx, &pb.CreatePolicyRequest{Input: polInput})
	if err != nil || polRes.GetValue().GetApprovedByPrincipalId() != adminID {
		t.Fatalf("create policy: res=%+v err=%v", polRes, err)
	}
	policyID = polRes.GetValue().GetId()

	if _, err := client.CreatePolicy(ctx, &pb.CreatePolicyRequest{Input: polInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate policy version = %v, want AlreadyExists", status.Code(err))
	}
	if _, err := client.DeletePolicy(ctx, &common.IDRequest{Id: policyID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete published policy = %v, want FailedPrecondition", status.Code(err))
	}

	ackRes, err := client.CreatePolicyAcknowledgement(ctx, &pb.CreatePolicyAcknowledgementRequest{Input: &pb.PolicyAcknowledgementInput{
		PolicyId: &policyID, WorkerId: &workerID,
	}})
	if err != nil {
		t.Fatalf("create acknowledgement: %v", err)
	}
	ackID = ackRes.GetValue().GetId()

	// Pihak lain dilarang mengakui atas nama pekerja pemilik.
	policyAuth.setPrincipal(otherID)
	if _, err := client.AcknowledgePolicy(ctx, &pb.AcknowledgePolicyRequest{
		Id: ackID, EvidenceDocumentVersionReferenceId: &evidenceDocID,
	}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("non-owner acknowledge = %v, want PermissionDenied", status.Code(err))
	}

	// Pekerja pemilik berhasil mengakui kebijakan; pengulangan kedua ditolak.
	policyAuth.setPrincipal(workerPrincipalID)
	if _, err := client.AcknowledgePolicy(ctx, &pb.AcknowledgePolicyRequest{
		Id: ackID, EvidenceDocumentVersionReferenceId: &evidenceDocID,
	}); status.Code(err) != codes.Unavailable {
		t.Fatalf("acknowledge with unverified evidence = %v, want Unavailable", status.Code(err))
	}
	acked, err := client.AcknowledgePolicy(ctx, &pb.AcknowledgePolicyRequest{
		Id: ackID,
	})
	if err != nil || acked.GetValue().GetAcknowledgedByPrincipalId() != workerPrincipalID || acked.GetValue().GetAcknowledgedAt() == nil {
		t.Fatalf("acknowledge policy: res=%+v err=%v", acked, err)
	}
	if _, err := client.AcknowledgePolicy(ctx, &pb.AcknowledgePolicyRequest{
		Id: ackID,
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat acknowledge = %v, want FailedPrecondition", status.Code(err))
	}
}
