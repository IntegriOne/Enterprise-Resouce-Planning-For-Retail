//go:build integration

package leave_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type leaveIntegrationClock struct{ seconds atomic.Int64 }

func (c *leaveIntegrationClock) Now() time.Time {
	return time.Unix(c.seconds.Load(), 0).UTC()
}

type leaveIntegrationPolicy struct{}

func (leaveIntegrationPolicy) Authorize(ctx context.Context, _ operation.Access) (operation.Actor, error) {
	principalID := "00000000-0000-4000-8000-000000000001"
	if values, ok := metadata.FromIncomingContext(ctx); ok && len(values.Get("x-test-approver")) != 0 {
		principalID = "00000000-0000-4000-8000-000000000003"
	}
	return operation.Actor{PrincipalID: principalID, SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (leaveIntegrationPolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func TestLeaveLedgerRPCWithPostgreSQL(t *testing.T) {
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
	clock := &leaveIntegrationClock{}
	clock.seconds.Store(time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC).Unix())
	svc, err := service.NewService(service.Dependencies{
		Runtime: operation.Dependencies{Transactor: repos.Transactor, Outbox: repos.OutboxEvent,
			Clock: clock, Policy: leaveIntegrationPolicy{}, Integrity: security.NewIntegrity(validators)},
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
	start, end, leaveType, opening := "2025-01-01", "2025-12-31", "ANNUAL", "12"
	policyCode, policyName, jurisdiction, method, effectiveFrom := "ANNUAL-"+workerID, "Annual leave", "ID", "annual_grant", "2025-01-01"
	version := int32(1)
	invalidMethod := "statutory_engine"
	policyInput := &pb.LeavePolicyInput{PolicyCode: &policyCode, PolicyVersion: &version,
		PolicyName: &policyName, LeaveTypeCode: &leaveType, JurisdictionCode: &jurisdiction,
		AccrualMethod: &invalidMethod, AnnualEntitlement: &opening, MaxBalance: &opening, EffectiveFrom: &effectiveFrom}
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
	adjustmentKey, adjustmentReason, explanation := "adjust-"+id, "CORRECTION", "Correct the initial balance"
	draftRequest := &pb.CreateLeaveAdjustmentRequest{Input: &pb.LeaveAdjustmentInput{
		EntitlementId: &id, QuantityDelta: &debit, ReasonCode: &adjustmentReason,
		Reason: &explanation, IdempotencyKey: &adjustmentKey,
	}}
	adjustment, err := client.CreateLeaveAdjustment(ctx, draftRequest)
	if err != nil || adjustment.GetItem().GetId() == "" {
		t.Fatalf("create adjustment: %+v, %v", adjustment, err)
	}
	adjustmentID := adjustment.GetItem().GetId()
	if _, err := client.CreateLeaveAdjustment(ctx, draftRequest); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate adjustment draft = %v, want AlreadyExists", err)
	}
	recoveredDraft, err := client.FindLeaveAdjustmentByIdempotencyKey(ctx, &common.StringFinderRequest{Value: &adjustmentKey, Limit: 1})
	if err != nil || recoveredDraft.GetPage().GetTotal() != 1 || len(recoveredDraft.GetItems()) != 1 ||
		recoveredDraft.GetItems()[0].GetId() != adjustmentID || recoveredDraft.GetItems()[0].GetEntitlementId() != id ||
		recoveredDraft.GetItems()[0].GetApprovedAt() != nil || recoveredDraft.GetItems()[0].GetRejectedAt() != nil {
		t.Fatalf("recover adjustment draft by key: %+v, %v", recoveredDraft, err)
	}
	unknownAdjustmentKey := adjustmentKey + "-unknown"
	unknownDraft, err := client.FindLeaveAdjustmentByIdempotencyKey(ctx, &common.StringFinderRequest{Value: &unknownAdjustmentKey, Limit: 1})
	if err != nil || unknownDraft.GetPage().GetTotal() != 0 || len(unknownDraft.GetItems()) != 0 {
		t.Fatalf("unknown adjustment key: %+v, %v", unknownDraft, err)
	}

	if _, err := client.ApproveLeaveAdjustment(ctx, &common.IDRequest{Id: adjustmentID}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self approval status = %v, want PermissionDenied", status.Code(err))
	}
	approverCtx := metadata.AppendToOutgoingContext(ctx, "x-test-approver", "true")
	approvedAdjustment, err := client.ApproveLeaveAdjustment(approverCtx, &common.IDRequest{Id: adjustmentID})
	if err != nil || approvedAdjustment.GetItem().GetApprovedByPrincipalId() != "00000000-0000-4000-8000-000000000003" {
		t.Fatalf("independent approval: %+v, %v", approvedAdjustment, err)
	}
	if _, err := client.DeleteLeaveAdjustment(ctx, &common.IDRequest{Id: adjustmentID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete approved adjustment status = %v, want FailedPrecondition", status.Code(err))
	}
	adjustmentType := "adjustment"
	rejectedKey := "reject-" + id
	pendingRejection, err := client.CreateLeaveAdjustment(ctx, &pb.CreateLeaveAdjustmentRequest{Input: &pb.LeaveAdjustmentInput{
		EntitlementId: &id, QuantityDelta: &debit, ReasonCode: &adjustmentReason,
		Reason: &explanation, IdempotencyKey: &rejectedKey,
	}})
	if err != nil {
		t.Fatalf("create adjustment for rejection: %v", err)
	}
	rejectedID := pendingRejection.GetItem().GetId()
	rejectRequest := &pb.RejectLeaveAdjustmentRequest{Id: rejectedID, ReasonCode: "DUPLICATE_EVIDENCE"}
	if _, err := client.RejectLeaveAdjustment(ctx, rejectRequest); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("self rejection = %v", err)
	}
	rejected, err := client.RejectLeaveAdjustment(approverCtx, rejectRequest)
	if err != nil || rejected.GetItem().GetRejectedAt() == nil ||
		rejected.GetItem().GetRejectedByPrincipalId() != approvedAdjustment.GetItem().GetApprovedByPrincipalId() ||
		rejected.GetItem().GetRejectionReasonCode() != rejectRequest.ReasonCode {
		t.Fatalf("reject adjustment: %+v, %v", rejected, err)
	}
	if _, err := client.ApproveLeaveAdjustment(approverCtx, &common.IDRequest{Id: rejectedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve rejected adjustment = %v", err)
	}
	if _, err := client.DeleteLeaveAdjustment(ctx, &common.IDRequest{Id: rejectedID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete rejected adjustment = %v", err)
	}
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &adjustmentType, QuantityDelta: &debit, EffectiveDate: &start,
		AdjustmentId: &rejectedID, SourceManagement: &source, SourceEventId: &rejectedKey,
		SourceLineKey: &lineKey, SourceDigest: []byte{7, 8, 9}, ReasonCode: &adjustmentReason,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("post rejected adjustment = %v", err)
	}
	postingFinder := &common.StringFinderRequest{Value: &adjustmentID, Limit: 1}
	beforePosting, err := client.FindLeaveLedgerEntryByAdjustmentID(ctx, postingFinder)
	if err != nil || beforePosting.GetPage().GetTotal() != 0 || len(beforePosting.GetItems()) != 0 {
		t.Fatalf("approved adjustment must not appear in ledger before posting: %+v, %v", beforePosting, err)
	}
	postingLine := "adjustment-posting-v1"
	postingPayload, err := json.Marshal([]string{postingLine, id, adjustmentID, debit, start, adjustmentReason})
	if err != nil {
		t.Fatal(err)
	}
	postingDigest := sha256.Sum256(postingPayload)
	posting := &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &adjustmentType, QuantityDelta: &debit, EffectiveDate: &start,
		AdjustmentId: &adjustmentID, SourceManagement: &source, SourceEventId: &adjustmentID,
		SourceLineKey: &postingLine, SourceDigest: postingDigest[:], ReasonCode: &adjustmentReason,
	}}
	corrected, err := client.PostLedgerEntry(ctx, posting)
	if err != nil || corrected.GetItem().GetRevision() != 3 || corrected.GetItem().GetClosingBalanceSnapshot() != "11.0000" {
		t.Fatalf("post approved adjustment: %+v, %v", corrected, err)
	}
	if _, err := client.PostLedgerEntry(ctx, posting); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate gateway adjustment posting = %v, want AlreadyExists", err)
	}
	reloaded, err := client.GetLeaveEntitlementByID(ctx, &common.IDRequest{Id: id})
	if err != nil || reloaded.GetItem().GetRevision() != 3 || reloaded.GetItem().GetClosingBalanceSnapshot() != "11.0000" {
		t.Fatalf("balance changed after duplicate posting: %+v, %v", reloaded, err)
	}

	afterPosting, err := client.FindLeaveLedgerEntryByAdjustmentID(ctx, postingFinder)
	if err != nil || afterPosting.GetPage().GetTotal() != 1 || len(afterPosting.GetItems()) != 1 ||
		afterPosting.GetItems()[0].GetAdjustmentId() != adjustmentID || afterPosting.GetItems()[0].GetEntitlementId() != id ||
		afterPosting.GetItems()[0].GetEntryType() != adjustmentType {
		t.Fatalf("expected one ledger entry for the posted adjustment: %+v, %v", afterPosting, err)
	}

	excessKey, excessDelta, excessEvent := "excess-"+id, "2", "excess-"+id
	excessAdjustment, err := client.CreateLeaveAdjustment(ctx, &pb.CreateLeaveAdjustmentRequest{Input: &pb.LeaveAdjustmentInput{
		EntitlementId: &id, QuantityDelta: &excessDelta, ReasonCode: &adjustmentReason,
		Reason: &explanation, IdempotencyKey: &excessKey,
	}})
	if err != nil {
		t.Fatalf("create excess adjustment: %v", err)
	}
	excessID := excessAdjustment.GetItem().GetId()
	if _, err := client.ApproveLeaveAdjustment(approverCtx, &common.IDRequest{Id: excessID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve excess adjustment = %v, want FailedPrecondition", err)
	}
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &adjustmentType, QuantityDelta: &excessDelta, EffectiveDate: &start,
		AdjustmentId: &excessID, SourceManagement: &source, SourceEventId: &excessEvent,
		SourceLineKey: &lineKey, SourceDigest: []byte{10, 11, 12}, ReasonCode: &adjustmentReason,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("excess adjustment posting status = %v, want InvalidArgument", status.Code(err))
	}
	for _, scenario := range []struct{ name, delta string }{
		{"deduction", "-6"},
		{"credit", "0.75"},
	} {
		t.Run("concurrent_"+scenario.name, func(t *testing.T) {
			reservationIDs := make([]string, 0, 2)
			for _, suffix := range []string{"first", "second"} {
				key, delta := "reserve-"+id+"-"+scenario.name+"-"+suffix, scenario.delta
				correction, err := client.CreateLeaveAdjustment(ctx, &pb.CreateLeaveAdjustmentRequest{Input: &pb.LeaveAdjustmentInput{
					EntitlementId: &id, QuantityDelta: &delta, ReasonCode: &adjustmentReason,
					Reason: &explanation, IdempotencyKey: &key,
				}})
				if err != nil {
					t.Fatalf("create competing correction: %v", err)
				}
				reservationIDs = append(reservationIDs, correction.GetItem().GetId())
			}
			startApprovals := make(chan struct{})
			approvalResults := make(chan error, len(reservationIDs))
			for _, correctionID := range reservationIDs {
				go func() {
					<-startApprovals
					_, err := client.ApproveLeaveAdjustment(approverCtx, &common.IDRequest{Id: correctionID})
					approvalResults <- err
				}()
			}
			close(startApprovals)
			approvalStatuses := make(map[codes.Code]int)
			for range reservationIDs {
				approvalStatuses[status.Code(<-approvalResults)]++
			}
			if approvalStatuses[codes.OK] != 1 || approvalStatuses[codes.FailedPrecondition] != 1 {
				t.Fatalf("concurrent corrections returned %v, want one approval and one balance rejection", approvalStatuses)
			}
			var persistedApprovals int64
			if err := database.WithContext(ctx).Table("hr_leave_adjustments").Where("id IN ? AND approved_at IS NOT NULL", reservationIDs).Count(&persistedApprovals).Error; err != nil || persistedApprovals != 1 {
				t.Fatalf("persisted competing approvals = %d, want 1: %v", persistedApprovals, err)
			}
			// Hapus fixture koreksi sebelum melanjutkan skenario saldo awal.
			if err := database.WithContext(ctx).Exec("DELETE FROM hr_leave_adjustments WHERE id IN ?", reservationIDs).Error; err != nil {
				t.Fatalf("remove competing correction fixtures: %v", err)
			}
		})
	}
	underflowKey, underflowDelta := "underflow-"+id, "-12"
	underflow, err := client.CreateLeaveAdjustment(ctx, &pb.CreateLeaveAdjustmentRequest{Input: &pb.LeaveAdjustmentInput{
		EntitlementId: &id, QuantityDelta: &underflowDelta, ReasonCode: &adjustmentReason,
		Reason: &explanation, IdempotencyKey: &underflowKey,
	}})
	if err != nil {
		t.Fatalf("create adjustment below minimum balance: %v", err)
	}
	underflowID := underflow.GetItem().GetId()
	if _, err := client.ApproveLeaveAdjustment(approverCtx, &common.IDRequest{Id: underflowID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve adjustment below minimum balance = %v, want FailedPrecondition", err)
	}
	var underflowEvents int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", underflowID).Count(&underflowEvents).Error; err != nil || underflowEvents != 1 {
		t.Fatalf("adjustment events after failed approval = %d, error = %v", underflowEvents, err)
	}
	declined, err := client.RejectLeaveAdjustment(approverCtx, &pb.RejectLeaveAdjustmentRequest{
		Id: underflowID, ReasonCode: "INSUFFICIENT_BALANCE",
	})
	if err != nil || declined.GetItem().GetRejectedAt() == nil || declined.GetItem().GetApprovedAt() != nil {
		t.Fatalf("reject adjustment after failed approval: %+v, %v", declined, err)
	}
	expiryType, expiryEvent, expiryReason, remaining := "expiry", "expiry-"+id, "PERIOD_END", "-11"
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &expiryType, QuantityDelta: &remaining, EffectiveDate: &start,
		SourceManagement: &source, SourceEventId: &expiryEvent, SourceLineKey: &lineKey,
		SourceDigest: []byte{13, 14, 15}, ReasonCode: &expiryReason,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("early expiry status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &expiryType, QuantityDelta: &remaining, EffectiveDate: &end,
		SourceManagement: &source, SourceEventId: &expiryEvent, SourceLineKey: &lineKey,
		SourceDigest: []byte{13, 14, 15}, ReasonCode: &expiryReason,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expiry on last entitlement day = %v, want FailedPrecondition", err)
	}
	clock.seconds.Add(1)
	partialExpiry := "-10"
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &expiryType, QuantityDelta: &partialExpiry, EffectiveDate: &end,
		SourceManagement: &source, SourceEventId: &expiryEvent, SourceLineKey: &lineKey,
		SourceDigest: []byte{13, 14, 15}, ReasonCode: &expiryReason,
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("partial expiry status = %v, want InvalidArgument", status.Code(err))
	}
	expiryRequest := &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &expiryType, QuantityDelta: &remaining, EffectiveDate: &end,
		SourceManagement: &source, SourceEventId: &expiryEvent, SourceLineKey: &lineKey,
		SourceDigest: []byte{13, 14, 15}, ReasonCode: &expiryReason,
	}}
	blockingKey, blockingDelta := "expiry-block-"+id, "1"
	blocking, err := client.CreateLeaveAdjustment(ctx, &pb.CreateLeaveAdjustmentRequest{Input: &pb.LeaveAdjustmentInput{
		EntitlementId: &id, QuantityDelta: &blockingDelta, ReasonCode: &adjustmentReason,
		Reason: &explanation, IdempotencyKey: &blockingKey,
	}})
	if err != nil {
		t.Fatalf("create correction before expiry: %v", err)
	}
	blockingID := blocking.GetItem().GetId()
	if _, err := client.ApproveLeaveAdjustment(approverCtx, &common.IDRequest{Id: blockingID}); err != nil {
		t.Fatalf("approve correction before expiry: %v", err)
	}
	if _, err := client.PostLedgerEntry(ctx, expiryRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expire with unposted approved correction: %v", err)
	}
	// Hapus koreksi fixture agar skenario kedaluwarsa berikutnya tetap menguji saldo yang sama.
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_leave_adjustments WHERE id = ?", blockingID).Error; err != nil {
		t.Fatalf("remove blocking adjustment fixture: %v", err)
	}
	expired, err := client.PostLedgerEntry(ctx, expiryRequest)
	if err != nil || expired.GetItem().GetRevision() != 4 || expired.GetItem().GetClosingBalanceSnapshot() != "0.0000" {
		t.Fatalf("expire remaining balance: %+v, %v", expired, err)
	}
	secondExpiryEvent := "second-expiry-" + id
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &expiryType, QuantityDelta: &remaining, EffectiveDate: &end,
		SourceManagement: &source, SourceEventId: &secondExpiryEvent, SourceLineKey: &lineKey,
		SourceDigest: []byte{16, 17, 18}, ReasonCode: &expiryReason,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("second expiry status = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: id, Input: &pb.PostLedgerEntryInput{
		EntryType: &adjustmentType, QuantityDelta: &excessDelta, EffectiveDate: &end,
		AdjustmentId: &excessID, SourceManagement: &source, SourceEventId: &excessEvent,
		SourceLineKey: &lineKey, SourceDigest: []byte{10, 11, 12}, ReasonCode: &adjustmentReason,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("adjustment after expiry status = %v, want FailedPrecondition", status.Code(err))
	}
	afterExpiryKey := "late-adjustment-" + id
	if _, err := client.CreateLeaveAdjustment(ctx, &pb.CreateLeaveAdjustmentRequest{Input: &pb.LeaveAdjustmentInput{
		EntitlementId: &id, QuantityDelta: &excessDelta, ReasonCode: &adjustmentReason,
		Reason: &explanation, IdempotencyKey: &afterExpiryKey,
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("create adjustment after expiry status = %v, want FailedPrecondition", status.Code(err))
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
	if ledgerCount != 3 || outboxCount < 4 {
		t.Fatalf("ledger/outbox count = %d/%d, want 3/at least 4", ledgerCount, outboxCount)
	}
	secondWorkerID, err := ids.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if err := database.WithContext(ctx).Exec(
		"INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, hire_date, primary_legal_entity_reference_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		secondWorkerID, "LEAVE-"+secondWorkerID, "11111111-1111-4111-8111-111111111111", "employee", "active", "2025-01-01", "22222222-2222-4222-8222-222222222222",
	).Error; err != nil {
		t.Fatal(err)
	}
	second, err := client.CreateLeaveEntitlement(ctx, &pb.CreateLeaveEntitlementRequest{Input: &pb.LeaveEntitlementInput{
		WorkerId: &secondWorkerID, LeavePolicyId: &policyID, LeaveTypeCodeSnapshot: &leaveType,
		EntitlementPeriodStart: &start, EntitlementPeriodEnd: &end,
	}})
	if err != nil {
		t.Fatalf("create reversal fixture: %v", err)
	}
	secondID := second.GetItem().GetId()
	secondOpeningEvent := "opening-" + secondID
	_, err = client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: secondID, Input: &pb.PostLedgerEntryInput{
		EntryType: &entryType, QuantityDelta: &opening, EffectiveDate: &start,
		SourceManagement: &source, SourceEventId: &secondOpeningEvent, SourceLineKey: &lineKey,
		SourceDigest: []byte{21}, ReasonCode: &reason,
	}})
	if err != nil {
		t.Fatalf("post reversal fixture opening: %v", err)
	}
	openingPage, err := client.FindLeaveLedgerEntryBySourceEventID(ctx, &common.StringFinderRequest{
		Value: &secondOpeningEvent, Limit: 1,
	})
	if err != nil || len(openingPage.GetItems()) != 1 {
		t.Fatalf("find original opening: %+v, %v", openingPage, err)
	}
	originalID := openingPage.GetItems()[0].GetId()
	reversalType, reversalAmount, reversalReason := "reversal", "-12", "CORRECT_OPENING"
	reversalEvent := "reversal-" + secondID
	reversalInput := &pb.PostLedgerEntryInput{
		EntryType: &reversalType, QuantityDelta: &reversalAmount, EffectiveDate: &start,
		ReversesEntryId: &originalID, SourceManagement: &source,
		SourceEventId: &reversalEvent, SourceLineKey: &lineKey,
		SourceDigest: []byte{22}, ReasonCode: &reversalReason,
	}
	if _, err := client.PostLedgerEntry(ctx, &pb.PostLedgerEntryRequest{EntitlementId: secondID, Input: reversalInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("self reversal status = %v, want FailedPrecondition", status.Code(err))
	}
	partialReversal := proto.CloneOf(reversalInput)
	partialAmount := "-11"
	partialReversal.QuantityDelta = &partialAmount
	if _, err := client.PostLedgerEntry(approverCtx, &pb.PostLedgerEntryRequest{EntitlementId: secondID, Input: partialReversal}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("partial reversal status = %v, want InvalidArgument", status.Code(err))
	}
	foreignOpening, err := client.FindLeaveLedgerEntryBySourceEventID(ctx, &common.StringFinderRequest{Value: &eventID, Limit: 1})
	if err != nil || len(foreignOpening.GetItems()) != 1 {
		t.Fatalf("find foreign opening: %+v, %v", foreignOpening, err)
	}
	foreignReversal := proto.CloneOf(reversalInput)
	foreignID := foreignOpening.GetItems()[0].GetId()
	foreignReversal.ReversesEntryId = &foreignID
	if _, err := client.PostLedgerEntry(approverCtx, &pb.PostLedgerEntryRequest{EntitlementId: secondID, Input: foreignReversal}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("foreign entitlement reversal status = %v, want InvalidArgument", status.Code(err))
	}
	reversalBlockKey, reversalBlockDelta := "reserve-reversal-"+secondID, "-1"
	reversalBlock, err := client.CreateLeaveAdjustment(ctx, &pb.CreateLeaveAdjustmentRequest{Input: &pb.LeaveAdjustmentInput{
		EntitlementId: &secondID, QuantityDelta: &reversalBlockDelta, ReasonCode: &adjustmentReason,
		Reason: &explanation, IdempotencyKey: &reversalBlockKey,
	}})
	if err != nil {
		t.Fatalf("create correction before reversal: %v", err)
	}
	reversalBlockID := reversalBlock.GetItem().GetId()
	if _, err := client.ApproveLeaveAdjustment(approverCtx, &common.IDRequest{Id: reversalBlockID}); err != nil {
		t.Fatalf("approve correction before reversal: %v", err)
	}
	if _, err := client.PostLedgerEntry(approverCtx, &pb.PostLedgerEntryRequest{EntitlementId: secondID, Input: reversalInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("reverse opening while approved deduction reserves balance: %v", err)
	}
	// Hapus fixture koreksi agar pembalikan saldo awal dapat diuji.
	if err := database.WithContext(ctx).Exec("DELETE FROM hr_leave_adjustments WHERE id = ?", reversalBlockID).Error; err != nil {
		t.Fatalf("remove reversal correction fixture: %v", err)
	}
	reversed, err := client.PostLedgerEntry(approverCtx, &pb.PostLedgerEntryRequest{EntitlementId: secondID, Input: reversalInput})
	if err != nil || reversed.GetItem().GetRevision() != 3 || reversed.GetItem().GetClosingBalanceSnapshot() != "0.0000" {
		t.Fatalf("reverse opening: %+v, %v", reversed, err)
	}
	if _, err := client.PostLedgerEntry(approverCtx, &pb.PostLedgerEntryRequest{EntitlementId: secondID, Input: reversalInput}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate reversal source status = %v, want AlreadyExists", status.Code(err))
	}
	newReversalEvent := "another-reversal-" + secondID
	secondReversal := proto.CloneOf(reversalInput)
	secondReversal.SourceEventId = &newReversalEvent
	if _, err := client.PostLedgerEntry(approverCtx, &pb.PostLedgerEntryRequest{EntitlementId: secondID, Input: secondReversal}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("second reversal status = %v, want InvalidArgument", status.Code(err))
	}
	pendingKey := "pending-close-" + secondID
	pendingDelta := "1"
	pendingAtClose, err := client.CreateLeaveAdjustment(ctx, &pb.CreateLeaveAdjustmentRequest{Input: &pb.LeaveAdjustmentInput{
		EntitlementId: &secondID, QuantityDelta: &pendingDelta, ReasonCode: &adjustmentReason,
		Reason: &explanation, IdempotencyKey: &pendingKey,
	}})
	if err != nil || pendingAtClose.GetItem().GetId() == "" {
		t.Fatalf("create correction before closure: %+v, %v", pendingAtClose, err)
	}
	closeRequest := &common.IDRequest{Id: secondID}
	pendingAtCloseID := pendingAtClose.GetItem().GetId()
	if _, err := client.ApproveLeaveAdjustment(approverCtx, &common.IDRequest{Id: pendingAtCloseID}); err != nil {
		t.Fatalf("approve correction before closure: %v", err)
	}
	if _, err := client.CloseLeaveEntitlement(approverCtx, closeRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("closing with unposted approved adjustment: %v", err)
	}
	// Kembalikan fixture ke draf untuk menguji alur penolakan setelah penutupan.
	if err := database.WithContext(ctx).Exec("UPDATE hr_leave_adjustments SET approved_at = NULL, approved_by_principal_id = NULL WHERE id = ?", pendingAtCloseID).Error; err != nil {
		t.Fatalf("restore draft adjustment fixture: %v", err)
	}
	closed, err := client.CloseLeaveEntitlement(approverCtx, closeRequest)
	if err != nil || closed.GetItem().GetClosedAt() == nil ||
		closed.GetItem().GetClosedByPrincipalId() != approvedAdjustment.GetItem().GetApprovedByPrincipalId() ||
		closed.GetItem().GetRevision() != reversed.GetItem().GetRevision() {
		t.Fatalf("close zero-balance entitlement: %+v, %v", closed, err)
	}
	if _, err := client.CloseLeaveEntitlement(approverCtx, closeRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat entitlement close = %v", err)
	}
	if _, err := client.ApproveLeaveAdjustment(approverCtx, &common.IDRequest{Id: pendingAtCloseID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("approve pending correction after closure = %v", err)
	}
	rejectedAfterClose, err := client.RejectLeaveAdjustment(approverCtx, &pb.RejectLeaveAdjustmentRequest{
		Id: pendingAtCloseID, ReasonCode: "PERIOD_CLOSED",
	})
	if err != nil || rejectedAfterClose.GetItem().GetRejectedAt() == nil || rejectedAfterClose.GetItem().GetApprovedAt() != nil {
		t.Fatalf("reject pending correction after closure: %+v, %v", rejectedAfterClose, err)
	}
	if _, err := client.PostLedgerEntry(approverCtx, &pb.PostLedgerEntryRequest{EntitlementId: secondID, Input: reversalInput}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("posting after close = %v", err)
	}
	if _, err := client.DeleteLeaveEntitlement(ctx, closeRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete closed entitlement = %v", err)
	}
	if err := database.WithContext(ctx).Table("hr_leave_ledger_entries").Where("entitlement_id = ?", secondID).Count(&ledgerCount).Error; err != nil {
		t.Fatal(err)
	}
	if ledgerCount != 2 {
		t.Fatalf("reversed entitlement has %d entries, want 2", ledgerCount)
	}
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", secondID).Count(&outboxCount).Error; err != nil {
		t.Fatal(err)
	}
	if outboxCount != 4 {
		t.Fatalf("closed entitlement has %d events, want 4", outboxCount)
	}
}
