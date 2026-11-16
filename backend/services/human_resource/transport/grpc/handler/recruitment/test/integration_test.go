//go:build integration

package recruitment_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/recruitment/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/recruitment"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/recruitment"
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

type candidateIntegrationPolicy struct{}

func (candidateIntegrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: "00000000-0000-4000-8000-000000000001", SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (candidateIntegrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestCandidateDraftRPCWithPostgreSQL(t *testing.T) {
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
	validators, err := security.RecruitmentCandidateValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: candidateIntegrationPolicy{}, Integrity: security.NewIntegrity(validators)},
		Requisition: repos.Requisition, Candidate: repos.Candidate, Application: repos.Application,
		ApplicationStageHistory: repos.ApplicationStageHistory, Interview: repos.Interview,
		InterviewPanelist: repos.InterviewPanelist, InterviewFeedback: repos.InterviewFeedback,
		Offer: repos.Offer, JobPosting: repos.JobPosting, ApplicationAssessment: repos.ApplicationAssessment,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	connection, err := grpc.NewClient("passthrough:///hrm-candidate-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	client := pb.NewRecruitmentServiceClient(connection)
	ack := timestamppb.New(time.Now().Add(-time.Minute))
	retention := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
	name := "Candidate A"
	create := func(number string) (*pb.CandidateResponse, error) {
		return client.CreateCandidate(ctx, &pb.CreateCandidateRequest{Input: &pb.CandidateInput{
			CandidateNumber: &number, PreferredName: &name, PrivacyNoticeVersion: text("v1"),
			NoticeAcknowledgedAt: ack, DataRetentionUntil: &retention,
		}})
	}
	created, err := create("CAND-DRAFT-1")
	if err != nil || created.GetItem().GetId() == "" {
		t.Fatalf("create candidate: %v", err)
	}
	id := created.GetItem().GetId()
	if created.GetItem().GetCreatedByPrincipalId() != "00000000-0000-4000-8000-000000000001" {
		t.Fatal("candidate creator did not come from server actor")
	}
	if _, err := create("CAND-DRAFT-1"); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate candidate number: %v", err)
	}
	badNumber := "CAND-DRAFT-BAD"
	_, err = client.CreateCandidate(ctx, &pb.CreateCandidateRequest{Input: &pb.CandidateInput{
		CandidateNumber: &badNumber, PrivacyNoticeVersion: text("v1"), DataRetentionUntil: &retention,
	}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("candidate without notice acknowledgement: %v", err)
	}
	_, err = client.CreateCandidate(ctx, &pb.CreateCandidateRequest{Input: &pb.CandidateInput{
		CandidateNumber: &badNumber, PrivacyNoticeVersion: text("v1"),
		NoticeAcknowledgedAt: ack, DataRetentionUntil: &retention,
		ContactReferenceId: text("00000000-0000-4000-8000-000000000003"),
	}})
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("unverified Foundation contact reference: %v", err)
	}
	page, err := client.GetAllCandidates(ctx, &common.PageRequest{Limit: 1})
	if err != nil || page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 1 {
		t.Fatalf("candidate pagination: %v", err)
	}
	corrected := "Corrected candidate name"
	updated, err := client.UpdateCandidate(ctx, &pb.UpdateCandidateRequest{Id: id,
		Input:      &pb.CandidateInput{PreferredName: &corrected},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"preferred_name"}}})
	if err != nil || updated.GetItem().GetPreferredName() != corrected {
		t.Fatalf("candidate name correction: %v", err)
	}
	_, err = client.UpdateCandidate(ctx, &pb.UpdateCandidateRequest{Id: id,
		Input:      &pb.CandidateInput{PrivacyNoticeVersion: text("v2")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"privacy_notice_version"}}})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("notice version changed through CRUD: %v", err)
	}
	var outboxCount int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", id).Count(&outboxCount).Error; err != nil || outboxCount != 2 {
		t.Fatalf("candidate outbox count = %d, err = %v", outboxCount, err)
	}
	if _, err := client.DeleteCandidate(ctx, &common.IDRequest{Id: id}); err != nil {
		t.Fatalf("delete unlinked candidate: %v", err)
	}
}
