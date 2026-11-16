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
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/competency/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	competencyservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/competency"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	competencyhandler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/competency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type integrationPolicy struct{ denyDelete atomic.Bool }

func (p *integrationPolicy) Authorize(_ context.Context, access operation.Access) (operation.Actor, error) {
	if p.denyDelete.Load() && access.Action == "delete" {
		return operation.Actor{}, hrerrors.ErrForbidden
	}
	return operation.Actor{
		PrincipalID: "00000000-0000-4000-8000-000000000001",
		SessionID:   "00000000-0000-4000-8000-000000000002",
		Kind:        "human",
	}, nil
}

func (*integrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestCompetencyRPCWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open HRM test database failed")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := pool.Close(); err != nil {
			t.Errorf("close HRM test database pool: %v", err)
		}
	})
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(db)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.CompetencyValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	policy := &integrationPolicy{}
	svc, err := competencyservice.NewService(competencyservice.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(validators),
		},
		WorkerQualification: repos.WorkerQualification, Competency: repos.Competency,
		JobCompetency: repos.JobCompetency, WorkerCompetencyAssessment: repos.WorkerCompetencyAssessment,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := competencyhandler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		if err := <-served; err != nil {
			t.Errorf("serve HRM test gRPC: %v", err)
		}
	})
	connection, err := grpc.NewClient("passthrough:///hrm-competency-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Errorf("close HRM test gRPC connection: %v", err)
		}
	})
	client := pb.NewCompetencyServiceClient(connection)
	ids := operation.UUIDGenerator{}
	familyID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	jobID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	requirementID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	var competencyID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, statement := range []struct {
			query string
			value string
		}{
			{"DELETE FROM hr_job_competencies WHERE id = ?", requirementID},
			{"DELETE FROM hr_jobs WHERE id = ?", jobID},
			{"DELETE FROM hr_job_families WHERE id = ?", familyID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", competencyID},
			{"DELETE FROM hr_competencies WHERE id = ?", competencyID},
		} {
			if err := db.WithContext(cleanupCtx).Exec(statement.query, statement.value).Error; err != nil {
				t.Errorf("clean HRM test data: %v", err)
			}
		}
	})
	code := "TEST-" + familyID
	name := "Leadership"
	category := "LEADERSHIP"
	version := int32(1)
	input := &pb.CompetencyInput{
		CompetencyCode: &code, CompetencyVersion: &version,
		CompetencyName: &name, CategoryCode: &category,
		ProficiencyScale: &common.JsonDocument{Utf8: []byte(`{"1":"basic"}`)},
	}
	created, err := client.CreateCompetency(ctx, &pb.CreateCompetencyRequest{Input: input})
	if err != nil || created.GetItem() == nil {
		t.Fatalf("create competency: %v", err)
	}
	competencyID = created.Item.Id
	if competencyID == "" || !created.Item.IsActive {
		t.Fatal("created competency has no identity or active default")
	}
	read, err := client.GetCompetencyByID(ctx, &common.IDRequest{Id: competencyID})
	if err != nil || read.GetItem().GetCompetencyCode() != code {
		t.Fatalf("read competency: %v", err)
	}
	found, err := client.FindCompetencyByCompetencyCode(ctx, &common.StringFinderRequest{Value: &code, Limit: 1})
	if err != nil || len(found.GetItems()) != 1 || found.GetPage().GetTotal() != 1 {
		t.Fatalf("find competency by code: %v", err)
	}
	nextPage, err := client.FindCompetencyByCompetencyCode(ctx, &common.StringFinderRequest{Value: &code, Limit: 1, Offset: 1})
	if err != nil || len(nextPage.GetItems()) != 0 || nextPage.GetPage().GetTotal() != 1 {
		t.Fatalf("find competency beyond first page: %v", err)
	}
	listed, err := client.GetAllCompetencies(ctx, &common.PageRequest{Limit: 1})
	if err != nil || len(listed.GetItems()) != 1 || listed.GetPage().GetLimit() != 1 || listed.GetPage().GetTotal() < 1 {
		t.Fatalf("list competencies with bounded page: %v", err)
	}
	if _, err := client.CreateCompetency(ctx, &pb.CreateCompetencyRequest{Input: input}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate code and version: %v", err)
	}
	updatedName := "Leadership and coaching"
	updated, err := client.UpdateCompetency(ctx, &pb.UpdateCompetencyRequest{
		Id: competencyID, Input: &pb.CompetencyInput{CompetencyName: &updatedName},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"competency_name"}},
	})
	if err != nil || updated.GetItem().GetCompetencyName() != updatedName {
		t.Fatalf("update competency metadata: %v", err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, ?, ?)",
		familyID, code, "Test family", "professional",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, valid_from) VALUES (?, ?, ?, ?, ?)",
		jobID, familyID, code, "Test job", "2026-01-01",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_job_competencies (id, job_id, competency_id, required_level, effective_from) VALUES (?, ?, ?, ?, ?)",
		requirementID, jobID, competencyID, 2, "2026-01-01",
	).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateCompetency(ctx, &pb.UpdateCompetencyRequest{
		Id: competencyID,
		Input: &pb.CompetencyInput{ProficiencyScale: &common.JsonDocument{Utf8: []byte(`{"1":"changed"}`)}},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"proficiency_scale"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("referenced rubric change: %v", err)
	}
	if _, err := client.DeleteCompetency(ctx, &common.IDRequest{Id: competencyID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("referenced competency deletion: %v", err)
	}
	policy.denyDelete.Store(true)
	if _, err := client.DeleteCompetency(ctx, &common.IDRequest{Id: competencyID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("permission denial: %v", err)
	}
	read, err = client.GetCompetencyByID(ctx, &common.IDRequest{Id: competencyID})
	if err != nil || read.GetItem().GetCompetencyName() != updatedName {
		t.Fatalf("denied operations changed competency: %v", err)
	}
	var events int64
	if err := db.WithContext(ctx).Raw("SELECT count(*) FROM hr_outbox_events WHERE aggregate_id = ?", competencyID).Scan(&events).Error; err != nil {
		t.Fatal(err)
	}
	if events != 2 {
		t.Fatalf("committed event count = %d, want 2", events)
	}
}
