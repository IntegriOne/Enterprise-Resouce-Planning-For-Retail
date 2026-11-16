//go:build integration

package leave_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/leave/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/leave"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/leave"
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

type leaveIntegrationPolicy struct{}

func (leaveIntegrationPolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: "00000000-0000-4000-8000-000000000001", SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (leaveIntegrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestLeaveOpeningRPCWithPostgreSQL(t *testing.T) {
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
	validators, err := security.LeaveValidators(repos)
	if err != nil {
		t.Fatal(err)
	}
	ids := operation.UUIDGenerator{}
	workerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, hire_date, primary_legal_entity_reference_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		workerID, "LEAVE-"+workerID, "11111111-1111-4111-8111-111111111111", "employee", "active", "2025-01-01", "22222222-2222-4222-8222-222222222222",
	).Error; err != nil {
		t.Fatal(err)
	}
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Policy: leaveIntegrationPolicy{}, Integrity: security.NewIntegrity(validators)},
		LeavePolicy: repos.LeavePolicy, LeaveEntitlement: repos.LeaveEntitlement,
		LeaveAdjustment: repos.LeaveAdjustment, LeaveLedgerEntry: repos.LeaveLedgerEntry,
	})
	if err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	if err := handler.Register(server, svc); err != nil {
		t.Fatal(err)
	}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient("passthrough:///hrm-leave-ledger-integration", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewLeaveServiceClient(conn)
	start, end, leaveType, opening := "2026-01-01", "2026-12-31", "ANNUAL", "12"
	policyCode, policyName, jurisdiction, method, effectiveFrom := "ANNUAL-"+workerID, "Annual leave", "ID", "annual_grant", "2025-01-01"
	version := int32(1)
	invalidMethod := "statutory_engine"
	policyInput := &pb.LeavePolicyInput{PolicyCode: &policyCode, PolicyVersion: &version,
		PolicyName: &policyName, LeaveTypeCode: &leaveType, JurisdictionCode: &jurisdiction,
		AccrualMethod: &invalidMethod, AnnualEntitlement: &opening, EffectiveFrom: &effectiveFrom}
	if _, err := client.CreateLeavePolicy(ctx, &pb.CreateLeavePolicyRequest{Input: policyInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("unsupported policy method status = %v, want FailedPrecondition", status.Code(err))
	}
	policyInput.AccrualMethod = &method
	policy, err := client.CreateLeavePolicy(ctx, &pb.CreateLeavePolicyRequest{Input: policyInput})
	if err != nil || policy.GetItem().GetId() == "" {
		t.Fatalf("create annual policy: %+v, %v", policy, err)
	}
	policyID := policy.GetItem().GetId()
	if _, err := client.CreateLeavePolicy(ctx, &pb.CreateLeavePolicyRequest{Input: policyInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate policy version status = %v, want AlreadyExists", status.Code(err))
	}
	bad, err := client.CreateLeaveEntitlement(ctx, &pb.CreateLeaveEntitlementRequest{Input: &pb.LeaveEntitlementInput{
		WorkerId: &workerID, LeavePolicyId: &policyID, LeaveTypeCodeSnapshot: &leaveType,
		EntitlementPeriodStart: &start, EntitlementPeriodEnd: &end, OpeningBalance: &opening, ClosingBalanceSnapshot: &opening,
	}})
	if status.Code(err) != codes.FailedPrecondition || bad != nil {
		t.Fatalf("direct balance create status = %v, want FailedPrecondition", status.Code(err))
	}
	wrongType := "SICK"
	if _, err := client.CreateLeaveEntitlement(ctx, &pb.CreateLeaveEntitlementRequest{Input: &pb.LeaveEntitlementInput{
		WorkerId: &workerID, LeavePolicyId: &policyID, LeaveTypeCodeSnapshot: &wrongType,
		EntitlementPeriodStart: &start, EntitlementPeriodEnd: &end,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("mismatched leave type status = %v, want FailedPrecondition", status.Code(err))
	}
	created, err := client.CreateLeaveEntitlement(ctx, &pb.CreateLeaveEntitlementRequest{Input: &pb.LeaveEntitlementInput{
		WorkerId: &workerID, LeavePolicyId: &policyID, LeaveTypeCodeSnapshot: &leaveType,
		EntitlementPeriodStart: &start, EntitlementPeriodEnd: &end,
	}})
	if err != nil || created.GetItem().GetId() == "" {
		t.Fatalf("create empty entitlement: %+v, %v", created, err)
	}
	id := created.GetItem().GetId()
	page, err := client.FindLeaveEntitlementByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1, Offset: 0})
	if err != nil || page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 1 || page.GetItems()[0].GetId() != id {
		t.Fatalf("worker entitlement finder: %+v, %v", page, err)
	}
	page, err = client.FindLeaveEntitlementByWorkerID(ctx, &common.StringFinderRequest{Value: &workerID, Limit: 1, Offset: 1})
	if err != nil || page.GetPage().GetTotal() != 1 || len(page.GetItems()) != 0 {
		t.Fatalf("worker entitlement offset: %+v, %v", page, err)
	}
	if _, err := client.UpdateLeaveEntitlement(ctx, &pb.UpdateLeaveEntitlementRequest{Id: id,
		Input:      &pb.LeaveEntitlementInput{OpeningBalance: &opening, ClosingBalanceSnapshot: &opening},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"opening_balance", "closing_balance_snapshot"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("direct balance update status = %v, want FailedPrecondition", status.Code(err))
	}
	entryType, source, eventID, lineKey, reason := "opening", "hrm", "opening-"+id, "initial", "INITIAL"
	excess := "13"
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &entryType, QuantityDelta: &excess, EffectiveDate: &start,
		SourceManagement: &source, SourceEventId: &eventID, SourceLineKey: &lineKey,
		SourceDigest: []byte{1, 2, 3}, ReasonCode: &reason,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("excessive opening status = %v, want InvalidArgument", status.Code(err))
	}
	posted, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &entryType, QuantityDelta: &opening, EffectiveDate: &start,
		SourceManagement: &source, SourceEventId: &eventID, SourceLineKey: &lineKey,
		SourceDigest: []byte{1, 2, 3}, ReasonCode: &reason,
	}})
	if err != nil || posted.GetItem().GetRevision() != 2 || posted.GetItem().GetClosingBalanceSnapshot() != "12.0000" {
		t.Fatalf("post opening: %+v, %v", posted, err)
	}
	loaded, err := client.GetLeaveEntitlementByID(ctx, &common.IDRequest{Id: id})
	if err != nil || loaded.GetItem().GetRevision() != 2 {
		t.Fatalf("read posted entitlement: %+v, %v", loaded, err)
	}
	changedName := "Changed annual leave"
	if _, err := client.UpdateLeavePolicy(ctx, &pb.UpdateLeavePolicyRequest{Id: policyID,
		Input: &pb.LeavePolicyInput{PolicyName: &changedName}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"policy_name"}}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("used policy update status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteLeavePolicy(ctx, &common.IDRequest{Id: policyID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("used policy delete status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &entryType, QuantityDelta: &opening, EffectiveDate: &start,
		SourceManagement: &source, SourceEventId: &eventID, SourceLineKey: &lineKey,
		SourceDigest: []byte{1, 2, 3}, ReasonCode: &reason,
	}}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate source status = %v, want AlreadyExists", status.Code(err))
	}
	usage, debit, nextEvent := "usage", "-1", "usage-"+id
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &usage, QuantityDelta: &debit, EffectiveDate: &start,
		SourceManagement: &source, SourceEventId: &nextEvent, SourceLineKey: &lineKey,
		SourceDigest: []byte{4, 5, 6}, ReasonCode: &reason,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("unsupported usage status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteLeaveEntitlement(ctx, &common.IDRequest{Id: id}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete posted entitlement status = %v, want FailedPrecondition", status.Code(err))
	}
	var ledgerCount, outboxCount int64
	if err := database.WithContext(ctx).Table("hr_leave_ledger_entries").Where("entitlement_id = ?", id).Count(&ledgerCount).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", id).Count(&outboxCount).Error; err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 1 || outboxCount < 2 {
		t.Fatalf("ledger/outbox count = %d/%d, want 1/at least 2", ledgerCount, outboxCount)
	}
}
