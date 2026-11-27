//go:build integration

package test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/career/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/career"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/career"
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

func TestCareerPathTransitionRPCWithPostgreSQL(t *testing.T) {
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
	t.Cleanup(func() { _ = pool.Close() })
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.New(db)
	if err != nil {
		t.Fatal(err)
	}
	validators, err := security.CareerValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	policy := &integrationPolicy{}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{
			Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: policy, Integrity: security.NewIntegrity(validators),
		},
		DevelopmentPlan: repos.DevelopmentPlan, DevelopmentAction: repos.DevelopmentAction,
		CareerPath: repos.CareerPath, CareerPathTransition: repos.CareerPathTransition,
		WorkerCareerAspiration: repos.WorkerCareerAspiration,
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
	conn, err := grpc.NewClient("passthrough:///hrm-career-transition-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewCareerServiceClient(conn)

	ids := operation.UUIDGenerator{}
	familyID, _ := ids.NewID()
	jobFromID, _ := ids.NewID()
	jobToID, _ := ids.NewID()
	reqDocVerID, _ := ids.NewID()
	var pathID, transID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, stmt := range []struct {
			q string
			v string
		}{
			{"DELETE FROM hr_career_path_transitions WHERE id = ?", transID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", transID},
			{"DELETE FROM hr_career_paths WHERE id = ?", pathID},
			{"DELETE FROM hr_outbox_events WHERE aggregate_id = ?", pathID},
			{"DELETE FROM hr_jobs WHERE id IN (?, ?)", jobFromID},
			{"DELETE FROM hr_job_families WHERE id = ?", familyID},
		} {
			if stmt.v == "" {
				continue
			}
			if stmt.q == "DELETE FROM hr_jobs WHERE id IN (?, ?)" {
				_ = db.WithContext(cleanupCtx).Exec(stmt.q, jobFromID, jobToID).Error
				continue
			}
			_ = db.WithContext(cleanupCtx).Exec(stmt.q, stmt.v).Error
		}
	})

	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_job_families (id, family_code, family_name, job_domain) VALUES (?, ?, ?, ?)",
		familyID, "CAR-FAM-"+familyID, "Engineering", "engineering",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Exec(
		"INSERT INTO hr_jobs (id, job_family_id, job_code, job_title, valid_from) VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)",
		jobFromID, familyID, "J1-"+jobFromID, "Engineer", "2026-01-01",
		jobToID, familyID, "J2-"+jobToID, "Senior Engineer", "2026-01-01",
	).Error; err != nil {
		t.Fatal(err)
	}

	pathCode, pathName := "CAR-PATH-"+familyID, "Engineering Track"
	pathVer := int32(1)
	pathRes, err := client.CreateCareerPath(ctx, &pb.CreateCareerPathRequest{Input: &pb.CareerPathInput{
		PathCode: &pathCode, PathVersion: &pathVer, PathName: &pathName, JobFamilyId: &familyID,
		ValidFrom: timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
	}})
	if err != nil {
		t.Fatalf("create career path: %v", err)
	}
	pathID = pathRes.GetValue().GetId()

	moveType := "promotion"
	tenure := int32(12)
	transRes, err := client.CreateCareerPathTransition(ctx, &pb.CreateCareerPathTransitionRequest{Input: &pb.CareerPathTransitionInput{
		CareerPathId: &pathID, FromJobId: &jobFromID, ToJobId: &jobToID, MovementType: &moveType,
		MinimumTenureMonths: &tenure, RequirementsDocumentVersionReferenceId: &reqDocVerID, RequirementsDigest: []byte("req"),
	}})
	if err != nil {
		t.Fatalf("create career path transition: %v", err)
	}
	transID = transRes.GetValue().GetId()

	// Transisi membalik atau swa-transisi ditolak.
	if _, err := client.CreateCareerPathTransition(ctx, &pb.CreateCareerPathTransitionRequest{Input: &pb.CareerPathTransitionInput{
		CareerPathId: &pathID, FromJobId: &jobFromID, ToJobId: &jobFromID, MovementType: &moveType,
		MinimumTenureMonths: &tenure, RequirementsDocumentVersionReferenceId: &reqDocVerID, RequirementsDigest: []byte("req"),
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("self transition status = %v, want InvalidArgument", status.Code(err))
	}

	// Jalur karir yang sudah memiliki transisi dapat disetujui oleh actor berbeda (SoD).
	policy.principalID = "00000000-0000-4000-8000-000000000009"
	approved, err := client.ApproveCareerPath(ctx, &common.IDRequest{Id: pathID})
	if err != nil || approved.GetValue().GetStatus() != "approved" {
		t.Fatalf("approve career path: res=%+v err=%v", approved, err)
	}
	// Transisi pada jalur yang sudah disetujui tidak boleh dihapus.
	if _, err := client.DeleteCareerPathTransition(ctx, &common.IDRequest{Id: transID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete transition on approved career path = %v, want FailedPrecondition", status.Code(err))
	}
}
