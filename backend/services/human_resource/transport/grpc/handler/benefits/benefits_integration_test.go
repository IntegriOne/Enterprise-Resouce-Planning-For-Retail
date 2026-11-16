//go:build integration

package benefits_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/benefits/v1"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/benefits"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/benefits"
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

type benefitsIntegrationPolicy struct{}

func (benefitsIntegrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{
		PrincipalID: "00000000-0000-4000-8000-000000000001",
		SessionID:   "00000000-0000-4000-8000-000000000002",
		Kind:        "human",
	}, nil
}

func (benefitsIntegrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestBenefitRequestRPCWithPostgreSQL(t *testing.T) {
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
	validators, err := security.BenefitsValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: benefitsIntegrationPolicy{}, Integrity: security.NewIntegrity(validators)},
		BenefitPlan: repos.BenefitPlan, BenefitEnrollment: repos.BenefitEnrollment,
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
	t.Cleanup(func() {
		server.Stop()
		if err := <-served; err != nil {
			t.Errorf("serve HRM benefits gRPC: %v", err)
		}
	})
	connection, err := grpc.NewClient("passthrough:///hrm-benefits-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	client := pb.NewBenefitsServiceClient(connection)
	ids := operation.UUIDGenerator{}
	workerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	var planID, enrollmentID string
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if enrollmentID != "" {
			if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_benefit_enrollments WHERE id = ?", enrollmentID).Error; err != nil {
				t.Errorf("cleanup HRM benefit enrollment: %v", err)
			}
		}
		if planID != "" {
			if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_benefit_plans WHERE id = ?", planID).Error; err != nil {
				t.Errorf("cleanup HRM benefit plan: %v", err)
			}
		}
		if err := database.WithContext(cleanupCtx).Exec("DELETE FROM hr_workers WHERE id = ?", workerID).Error; err != nil {
			t.Errorf("cleanup HRM benefit worker: %v", err)
		}
		if err := database.WithContext(cleanupCtx).Table("hr_outbox_events").Where("aggregate_id IN ?", []string{planID, enrollmentID}).Delete(nil).Error; err != nil {
			t.Errorf("cleanup HRM benefit outbox: %v", err)
		}
	})
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, hire_date, primary_legal_entity_reference_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		workerID, "BENEFIT-"+workerID, "11111111-1111-4111-8111-111111111111", "employee", "active", "2025-01-01", "22222222-2222-4222-8222-222222222222",
	).Error; err != nil {
		t.Fatal(err)
	}
	code, name, benefitType, jurisdiction, from, to := "MEAL-"+workerID, "Meal allowance", "meal", "ID", "2026-01-01", "2026-12-31"
	version := int32(1)
	provider := "33333333-3333-4333-8333-333333333333"
	input := &pb.BenefitPlanInput{PlanCode: &code, PlanVersion: &version, PlanName: &name,
		BenefitType: &benefitType, JurisdictionCode: &jurisdiction, ValidFrom: &from, ValidTo: &to,
		ProviderPartyReferenceId: &provider}
	if _, err := client.CreateBenefitPlan(ctx, &pb.CreateBenefitPlanRequest{Input: input}); status.Code(err) != codes.Unavailable {
		t.Fatalf("unverified external provider status = %v, want Unavailable", status.Code(err))
	}
	input.ProviderPartyReferenceId = nil
	created, err := client.CreateBenefitPlan(ctx, &pb.CreateBenefitPlanRequest{Input: input})
	if err != nil || created.GetItem().GetId() == "" {
		t.Fatalf("create local benefit plan: %+v, %v", created, err)
	}
	planID = created.GetItem().GetId()
	if _, err := client.CreateBenefitPlan(ctx, &pb.CreateBenefitPlanRequest{Input: input}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate benefit plan status = %v, want AlreadyExists", status.Code(err))
	}
	listedPlans, err := client.GetAllBenefitPlans(ctx, &common.PageRequest{Limit: 1, Offset: 0})
	if err != nil || listedPlans.GetPage().GetTotal() != 1 || len(listedPlans.GetItems()) != 1 {
		t.Fatalf("plan first page: %+v, %v", listedPlans, err)
	}
	listedPlans, err = client.GetAllBenefitPlans(ctx, &common.PageRequest{Limit: 1, Offset: 1})
	if err != nil || listedPlans.GetPage().GetTotal() != 1 || len(listedPlans.GetItems()) != 0 {
		t.Fatalf("plan offset page: %+v, %v", listedPlans, err)
	}
	page, err := client.FindBenefitPlanByPlanCode(ctx, &common.StringFinderRequest{Value: &code, Limit: 1})
	if err != nil || page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 1 {
		t.Fatalf("plan finder: %+v, %v", page, err)
	}
	approved := "approved"
	request := &pb.BenefitEnrollmentInput{WorkerId: &workerID, BenefitPlanId: &planID, EffectiveFrom: &from, EffectiveTo: &to,
		EnrollmentStatus: &approved}
	if _, err := client.CreateBenefitEnrollment(ctx, &pb.CreateBenefitEnrollmentRequest{Input: request}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("direct approved enrollment status = %v, want FailedPrecondition", status.Code(err))
	}
	request.EnrollmentStatus = nil
	amount := "100.000000"
	request.EmployeeContributionAmount = &amount
	if _, err := client.CreateBenefitEnrollment(ctx, &pb.CreateBenefitEnrollmentRequest{Input: request}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("unverified contribution status = %v, want FailedPrecondition", status.Code(err))
	}
	request.EmployeeContributionAmount = nil
	enrolled, err := client.CreateBenefitEnrollment(ctx, &pb.CreateBenefitEnrollmentRequest{Input: request})
	if err != nil || enrolled.GetItem().GetId() == "" || enrolled.GetItem().GetEnrollmentStatus() != "pending" {
		t.Fatalf("create pending enrollment: %+v, %v", enrolled, err)
	}
	enrollmentID = enrolled.GetItem().GetId()
	listedEnrollments, err := client.FindBenefitEnrollmentByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1})
	if err != nil || listedEnrollments.GetPage().GetTotal() != 1 || len(listedEnrollments.GetItems()) != 1 ||
		listedEnrollments.GetItems()[0].GetId() != enrollmentID {
		t.Fatalf("worker enrollment finder: %+v, %v", listedEnrollments, err)
	}
	listedEnrollments, err = client.GetAllBenefitEnrollments(ctx, &common.PageRequest{Limit: 1, Offset: 1})
	if err != nil || listedEnrollments.GetPage().GetTotal() != 1 || len(listedEnrollments.GetItems()) != 0 {
		t.Fatalf("enrollment offset page: %+v, %v", listedEnrollments, err)
	}
	shortEnd := "2026-11-30"
	updated, err := client.UpdateBenefitEnrollment(ctx, &pb.UpdateBenefitEnrollmentRequest{Id: enrollmentID,
		Input: &pb.BenefitEnrollmentInput{EffectiveTo: &shortEnd}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"effective_to"}}})
	if err != nil || updated.GetItem().GetEffectiveTo() != shortEnd {
		t.Fatalf("update pending enrollment period: %+v, %v", updated, err)
	}
	loaded, err := client.GetBenefitEnrollmentByID(ctx, &common.IDRequest{Id: enrollmentID})
	if err != nil || loaded.GetItem().GetEffectiveTo() != shortEnd {
		t.Fatalf("read updated enrollment: %+v, %v", loaded, err)
	}
	if _, err := client.UpdateBenefitEnrollment(ctx, &pb.UpdateBenefitEnrollmentRequest{Id: enrollmentID,
		Input:      &pb.BenefitEnrollmentInput{EnrollmentStatus: &approved},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"enrollment_status"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("direct enrollment approval status = %v, want FailedPrecondition", status.Code(err))
	}
	outsidePlan := "2027-01-01"
	if _, err := client.UpdateBenefitEnrollment(ctx, &pb.UpdateBenefitEnrollmentRequest{Id: enrollmentID,
		Input:      &pb.BenefitEnrollmentInput{EffectiveTo: &outsidePlan},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"effective_to"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("enrollment outside plan status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.CreateBenefitEnrollment(ctx, &pb.CreateBenefitEnrollmentRequest{Input: request}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate worker plan status = %v, want AlreadyExists", status.Code(err))
	}
	changedName := "Meal revised"
	if _, err := client.UpdateBenefitPlan(ctx, &pb.UpdateBenefitPlanRequest{Id: planID,
		Input: &pb.BenefitPlanInput{PlanName: &changedName}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"plan_name"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("used plan update status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteBenefitPlan(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("used plan delete status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteBenefitEnrollment(ctx, &common.IDRequest{Id: enrollmentID}); err != nil {
		t.Fatalf("delete pending enrollment: %v", err)
	}
	if _, err := client.UpdateBenefitPlan(ctx, &pb.UpdateBenefitPlanRequest{Id: planID,
		Input: &pb.BenefitPlanInput{PlanName: &changedName}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"plan_name"}}}); err != nil {
		t.Fatalf("update unused plan: %v", err)
	}
	if _, err := client.DeleteBenefitPlan(ctx, &common.IDRequest{Id: planID}); err != nil {
		t.Fatalf("delete unused plan: %v", err)
	}
	if _, err := client.GetBenefitPlanByID(ctx, &common.IDRequest{Id: planID}); status.Code(err) != codes.NotFound {
		t.Fatalf("deleted plan read status = %v, want NotFound", status.Code(err))
	}
	var eventCount int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id IN ?", []string{planID, enrollmentID}).Count(&eventCount).Error; err != nil {
		t.Fatal(err)
	}
	if eventCount < 6 {
		t.Fatalf("benefits outbox events = %d, want at least 6", eventCount)
	}
}
