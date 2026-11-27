//go:build integration

package engagement_test

import (
	"bytes"
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/engagement/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/engagement"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/engagement"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type engagementPolicy struct{ principalID string }

func (p engagementPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.principalID, SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (engagementPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

type engagementFoundationStub struct{}

func (engagementFoundationStub) VerifyLegalEntity(context.Context, string) error { return nil }
func (engagementFoundationStub) VerifyPositionReferences(context.Context, security.PositionReferenceCheck) error {
	return nil
}
func (engagementFoundationStub) VerifyWorkerReferences(context.Context, security.WorkerReferenceCheck) error {
	return nil
}
func (engagementFoundationStub) VerifyOrganizationUnit(context.Context, string, string) error {
	return nil
}
func (engagementFoundationStub) VerifyLocation(context.Context, string, string) error { return nil }
func (engagementFoundationStub) VerifyParty(context.Context, string, bool) error      { return nil }

func text(v string) *string { return &v }

func TestEngagementSurveyAndReadResultsWithPostgreSQL(t *testing.T) {
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
	validators, err := security.EngagementValidators(repos, engagementFoundationStub{})
	if err != nil {
		t.Fatal(err)
	}

	actorID := "00000000-0000-4000-8000-000000000091"
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: engagementPolicy{principalID: actorID}, Integrity: security.NewIntegrity(validators),
		},
		EngagementSurvey:       repos.EngagementSurvey,
		EngagementSurveyResult: repos.EngagementSurveyResult,
	})
	if err != nil {
		t.Fatal(err)
	}

	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		return next(authcontext.WithCaller(ctx, authcontext.CallerContext{PrincipalID: actorID, Authenticated: true}), req)
	}))
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-engagement-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewEngagementServiceClient(conn)

	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	unitID, _ := ids.NewID()
	docID, _ := ids.NewID()
	resultID, _ := ids.NewID()
	var surveyID, tempSurveyID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, stmt := range []struct {
			q string
			v string
		}{
			{"DELETE FROM hr_engagement_survey_results WHERE id = ?", resultID},
			{"DELETE FROM hr_engagement_surveys WHERE id = ?", surveyID},
			{"DELETE FROM hr_engagement_surveys WHERE id = ?", tempSurveyID},
		} {
			if stmt.v != "" {
				if err := database.WithContext(cleanupCtx).Exec(stmt.q, stmt.v).Error; err != nil {
					t.Errorf("cleanup %s: %v", stmt.q, err)
				}
			}
		}
		for _, agg := range []string{surveyID, tempSurveyID} {
			if agg != "" {
				if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_outbox_events WHERE aggregate_id = ?", agg).Error; err != nil {
					t.Errorf("cleanup outbox %s: %v", agg, err)
				}
			}
		}
	})

	ver := int32(1)
	nowUTC := time.Now().UTC()
	tooSmall := int32(4)
	if _, err := client.CreateEngagementSurvey(ctx, &pb.CreateEngagementSurveyRequest{Input: &pb.EngagementSurveyInput{
		SurveyCode: text("ENG-SMALL-" + legalID[:8]), SurveyVersion: &ver,
		LegalEntityReferenceId: &legalID, TopicCode: text("engagement"),
		OpensAt: timestamppb.New(nowUTC), ClosesAt: timestamppb.New(nowUTC.Add(72 * time.Hour)),
		QuestionnaireDocumentVersionReferenceId: &docID, QuestionnaireDigest: bytes.Repeat([]byte{0x41}, 32),
		MinimumReportingPopulation: &tooSmall,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("population < 5 = %v, want InvalidArgument", status.Code(err))
	}

	minPop := int32(6)
	svInput := &pb.EngagementSurveyInput{
		SurveyCode: text("ENG-" + legalID[:8]), SurveyVersion: &ver,
		LegalEntityReferenceId: &legalID, TopicCode: text("engagement"),
		OpensAt: timestamppb.New(nowUTC), ClosesAt: timestamppb.New(nowUTC.Add(7 * 24 * time.Hour)),
		QuestionnaireDocumentVersionReferenceId: &docID, QuestionnaireDigest: bytes.Repeat([]byte{0x42}, 32),
		MinimumReportingPopulation: &minPop,
	}
	svRes, err := client.CreateEngagementSurvey(ctx, &pb.CreateEngagementSurveyRequest{Input: svInput})
	if err != nil {
		t.Fatalf("create survey: %v", err)
	}
	surveyID = svRes.GetValue().GetId()
	if _, err := client.CreateEngagementSurvey(ctx, &pb.CreateEngagementSurveyRequest{Input: svInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate survey = %v, want AlreadyExists", status.Code(err))
	}

	// Hasil agregasi survei read-only dibaca dari DB; penulisan langsung ditolak oleh integrity.
	if _, err := client.CreateEngagementSurveyResult(ctx, &pb.CreateEngagementSurveyResultRequest{Input: &pb.EngagementSurveyResultInput{
		SurveyId: &surveyID, QuestionCode: text("Q1"), ResultVersion: &ver, RespondentCount: &minPop,
		ScoreValue: text("4.200000"), ScaleMinimum: text("1.000000"), ScaleMaximum: text("5.000000"),
		SourceDigest: bytes.Repeat([]byte{0x43}, 32),
	}}); status.Code(err) != codes.Unavailable {
		t.Fatalf("direct result write without validator = %v, want Unavailable", status.Code(err))
	}

	if err := database.WithContext(ctx).Exec(
		`INSERT INTO hr_engagement_survey_results (id, survey_id, organization_unit_reference_id, question_code,
		 result_version, respondent_count, score_value, scale_minimum, scale_maximum, source_digest)
		 VALUES (?, ?, ?, 'Q1', 1, 6, 4.200000, 1.000000, 5.000000, ?)`,
		resultID, surveyID, unitID, bytes.Repeat([]byte{0x43}, 32),
	).Error; err != nil {
		t.Fatal(err)
	}

	gotRes, err := client.FindEngagementSurveyResultBySurveyID(ctx, &common.StringFinderRequest{Value: &surveyID, Limit: 10})
	if err != nil || gotRes.GetPage().GetTotal() != 1 || len(gotRes.GetItems()) != 1 {
		t.Fatalf("find survey results: res=%+v err=%v", gotRes, err)
	}

	if err := database.WithContext(ctx).Table("hr_engagement_survey_results").
		Where("id = ?", resultID).Update("respondent_count", 5).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetEngagementSurveyResultByID(ctx, &common.IDRequest{Id: resultID}); status.Code(err) != codes.NotFound {
		t.Fatalf("get suppressed result = %v, want NotFound", status.Code(err))
	}
	suppressed, err := client.FindEngagementSurveyResultBySurveyID(ctx, &common.StringFinderRequest{Value: &surveyID, Limit: 1})
	if err != nil || suppressed.GetPage().GetTotal() != 0 || len(suppressed.GetItems()) != 0 {
		t.Fatalf("suppressed results: res=%+v err=%v", suppressed, err)
	}

	newPop := int32(7)
	if _, err := client.UpdateEngagementSurvey(ctx, &pb.UpdateEngagementSurveyRequest{
		Id: surveyID, Input: &pb.EngagementSurveyInput{MinimumReportingPopulation: &newPop},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"minimum_reporting_population"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("update survey with results = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteEngagementSurvey(ctx, &common.IDRequest{Id: surveyID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete survey with results = %v, want FailedPrecondition", status.Code(err))
	}

	ver2 := int32(2)
	tmpInput := proto.CloneOf(svInput)
	tmpInput.SurveyVersion = &ver2
	tmpRes, err := client.CreateEngagementSurvey(ctx, &pb.CreateEngagementSurveyRequest{Input: tmpInput})
	if err != nil {
		t.Fatalf("create temp survey: %v", err)
	}
	tempSurveyID = tmpRes.GetValue().GetId()
	if _, err := client.DeleteEngagementSurvey(ctx, &common.IDRequest{Id: tempSurveyID}); err != nil {
		t.Fatalf("delete unreferenced survey: %v", err)
	}
	tempSurveyID = ""
}
