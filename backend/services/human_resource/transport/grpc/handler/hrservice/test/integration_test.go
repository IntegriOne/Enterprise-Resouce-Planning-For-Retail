//go:build integration

package hrservice_test

import (
	"context"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	pb "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/hrservice/v1"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	appservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/hrservice"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	handler "github.com/integrione/erp-retail/backend/services/human_resource/transport/grpc/handler/hrservice"
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

type hrServicePolicy struct {
	mu          sync.Mutex
	principalID string
}

func (p *hrServicePolicy) currentPrincipal() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.principalID
}

func (p *hrServicePolicy) Authorize(context.Context, operation.Access) (operation.Actor, error) {
	return operation.Actor{PrincipalID: p.currentPrincipal(), SessionID: "00000000-0000-4000-8000-000000000002", Kind: "human"}, nil
}

func (*hrServicePolicy) ReadScope(context.Context, operation.Access) (operation.Scope, error) {
	return operation.Scope{All: true}, nil
}

func text(v string) *string { return &v }

type resolutionStub struct {
	versionID string
	calls     atomic.Int32
}

func (s *resolutionStub) VerifyServiceRequestResolution(_ context.Context, check service.ResolutionCheck) error {
	s.calls.Add(1)
	if check.RequestID == "" || check.WorkerID == "" || check.PrincipalID == "" {
		return hrerrors.ErrReferenceConflict
	}
	if check.VersionID != s.versionID {
		return hrerrors.ErrNotFound
	}
	return nil
}

func TestHRServiceRequestAndEventsWithPostgreSQL(t *testing.T) {
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
	validators, err := security.HRServiceValidators(repos)
	if err != nil {
		t.Fatal(err)
	}

	requesterID := "00000000-0000-4000-8000-000000000081"
	policy := &hrServicePolicy{principalID: requesterID}
	const resolutionID = "00000000-0000-4000-8000-000000000082"
	resolutionVerifier := &resolutionStub{versionID: resolutionID}
	registry, err := appservice.NewRegistry(repos, operation.Dependencies{
		Policy: policy, Integrity: security.NewIntegrity(validators),
	}, appservice.WithHRServiceResolutionVerifier(resolutionVerifier))
	if err != nil {
		t.Fatal(err)
	}

	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		return next(authcontext.WithCaller(ctx, authcontext.CallerContext{PrincipalID: policy.currentPrincipal(), Authenticated: true}), req)
	}))
	if err := handler.Register(server, registry.HRService); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close(); <-served })
	conn, err := grpc.NewClient("passthrough:///hrm-hrservice-integration",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	client := pb.NewHRServiceClient(conn)

	ids := operation.UUIDGenerator{}
	legalID, _ := ids.NewID()
	personID, _ := ids.NewID()
	workerID, _ := ids.NewID()
	docID, _ := ids.NewID()
	var reqID, tempReqID, cancelledReqID, evID string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, stmt := range []struct {
			q string
			v string
		}{
			{"DELETE FROM hr_service_request_events WHERE id = ?", evID},
			{"DELETE FROM hr_service_requests WHERE id = ?", reqID},
			{"DELETE FROM hr_service_requests WHERE id = ?", tempReqID},
			{"DELETE FROM hr_service_requests WHERE id = ?", cancelledReqID},
			{"DELETE FROM hr_workers WHERE id = ?", workerID},
		} {
			if stmt.v != "" {
				if err := database.WithContext(cleanupCtx).Exec(stmt.q, stmt.v).Error; err != nil {
					t.Errorf("cleanup %s: %v", stmt.q, err)
				}
			}
		}
		for _, agg := range []string{evID, reqID, tempReqID, cancelledReqID} {
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
		workerID, "HRS-W-"+workerID, personID, legalID, requesterID,
	).Error; err != nil {
		t.Fatal(err)
	}

	if _, err := client.CreateServiceRequest(ctx, &pb.CreateServiceRequestRequest{Input: &pb.ServiceRequestInput{
		RequestNumber: text("SR-BAD-" + workerID[:8]), WorkerId: &workerID,
		RequestType: text("invalid_kind"), Status: text("open"),
		RestrictedDetailsDocumentVersionReferenceId: &docID,
		IdempotencyKey: text("idem-sr-bad-" + workerID),
	}}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid request type = %v, want InvalidArgument", status.Code(err))
	}

	srRes, err := client.CreateServiceRequest(ctx, &pb.CreateServiceRequestRequest{Input: &pb.ServiceRequestInput{
		RequestNumber: text("SR-" + workerID[:8]), WorkerId: &workerID,
		RequestType: text("employment_letter"), Status: text("open"),
		RestrictedDetailsDocumentVersionReferenceId: &docID,
		IdempotencyKey: text("idem-sr-1-" + workerID),
	}})
	if err != nil {
		t.Fatalf("create service request: %v", err)
	}
	reqID = srRes.GetValue().GetId()
	if srRes.GetValue().GetRequestedByPrincipalId() != requesterID ||
		srRes.GetValue().GetWorkerId() != workerID || srRes.GetValue().GetOpenedAt() == nil {
		t.Fatalf("request provenance: %+v", srRes.GetValue())
	}

	for _, test := range []struct {
		field string
		input *pb.ServiceRequestInput
	}{
		{"request_number", &pb.ServiceRequestInput{RequestNumber: text("replacement-" + reqID)}},
		{"idempotency_key", &pb.ServiceRequestInput{IdempotencyKey: text("replacement-" + reqID)}},
	} {
		if _, err := client.UpdateServiceRequest(ctx, &pb.UpdateServiceRequestRequest{
			Id: reqID, Input: test.input,
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{test.field}},
		}); status.Code(err) != codes.FailedPrecondition {
			t.Errorf("update service request %s = %v, want FailedPrecondition", test.field, err)
		}
	}

	if _, err := client.UpdateServiceRequest(ctx, &pb.UpdateServiceRequestRequest{
		Id: reqID, Input: &pb.ServiceRequestInput{Status: text("cancelled")},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("cancel from CRUD = %v", err)
	}
	if _, err := client.UpdateServiceRequest(ctx, &pb.UpdateServiceRequestRequest{
		Id: reqID, Input: &pb.ServiceRequestInput{DueAt: timestamppb.New(time.Now().UTC().Add(-48 * time.Hour))},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"due_at"}},
	}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("update with deadline before opening = %v", err)
	}

	if _, err := client.UpdateServiceRequest(ctx, &pb.UpdateServiceRequestRequest{
		Id: reqID, Input: &pb.ServiceRequestInput{DueAt: timestamppb.New(time.Now().UTC().Add(48 * time.Hour))},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"due_at"}},
	}); err != nil {
		t.Fatalf("update open service request: %v", err)
	}

	if _, err := client.CreateServiceRequestEvent(ctx, &pb.CreateServiceRequestEventRequest{Input: &pb.ServiceRequestEventInput{
		ServiceRequestId: &reqID, EventType: text("triaged"), IdempotencyKey: text("idem-srev-1-" + reqID),
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("create triage event through generic RPC = %v", err)
	}

	if result, err := client.CreateServiceRequestEvent(ctx, &pb.CreateServiceRequestEventRequest{Input: &pb.ServiceRequestEventInput{
		ServiceRequestId: &reqID, EventType: text("commented"), IdempotencyKey: text("idem-srev-1-" + reqID),
		EvidenceDocumentVersionReferenceId: &docID,
	}}); status.Code(err) != codes.Unavailable {
		evID = result.GetValue().GetId()
		t.Fatalf("create comment with unverified attachment = %v", err)
	}

	evRes, err := client.CreateServiceRequestEvent(ctx, &pb.CreateServiceRequestEventRequest{Input: &pb.ServiceRequestEventInput{
		ServiceRequestId: &reqID, EventType: text("commented"), IdempotencyKey: text("idem-srev-1-" + reqID),
	}})
	if err != nil {
		t.Fatalf("create service request event: %v", err)
	}
	evID = evRes.GetValue().GetId()

	if _, err := client.DeleteServiceRequestEvent(ctx, &common.IDRequest{Id: evID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete service request event = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := client.DeleteServiceRequest(ctx, &common.IDRequest{Id: reqID}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("delete service request with event = %v, want FailedPrecondition", status.Code(err))
	}

	tmpRes, err := client.CreateServiceRequest(ctx, &pb.CreateServiceRequestRequest{Input: &pb.ServiceRequestInput{
		RequestNumber: text("SR-TMP-" + workerID[:8]), WorkerId: &workerID,
		RequestType: text("policy_query"), Status: text("open"),
		RestrictedDetailsDocumentVersionReferenceId: &docID,
		IdempotencyKey: text("idem-sr-tmp-" + workerID),
	}})
	if err != nil {
		t.Fatalf("create temp service request: %v", err)
	}
	tempReqID = tmpRes.GetValue().GetId()
	if _, err := client.DeleteServiceRequest(ctx, &common.IDRequest{Id: tempReqID}); err != nil {
		t.Fatalf("delete empty service request: %v", err)
	}
	tempReqID = ""
	const assignee = "00000000-0000-4000-8000-000000000003"
	cancelDraft, err := client.CreateServiceRequest(ctx, &pb.CreateServiceRequestRequest{Input: &pb.ServiceRequestInput{
		RequestNumber: text("SR-CANCEL-" + workerID[:8]), WorkerId: &workerID,
		RequestType: text("policy_query"), Status: text("open"),
		RestrictedDetailsDocumentVersionReferenceId: &docID,
		IdempotencyKey: text("idem-sr-cancel-" + workerID),
	}})
	if err != nil {
		t.Fatal(err)
	}
	cancelledReqID = cancelDraft.GetValue().GetId()
	cancelRequest := &common.IDRequest{Id: cancelledReqID}
	policy.mu.Lock()
	policy.principalID = assignee
	policy.mu.Unlock()
	if _, err := client.CancelServiceRequest(ctx, cancelRequest); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("cancel by another principal: %v", err)
	}
	policy.mu.Lock()
	policy.principalID = requesterID
	policy.mu.Unlock()
	cancelCodes := make(chan codes.Code, 2)
	startCancel := make(chan struct{})
	for range 2 {
		go func() {
			<-startCancel
			_, err := client.CancelServiceRequest(ctx, cancelRequest)
			cancelCodes <- status.Code(err)
		}()
	}
	close(startCancel)
	firstCode, secondCode := <-cancelCodes, <-cancelCodes
	if !((firstCode == codes.OK && secondCode == codes.FailedPrecondition) ||
		(firstCode == codes.FailedPrecondition && secondCode == codes.OK)) {
		t.Fatalf("concurrent cancellation: codes %v and %v", firstCode, secondCode)
	}
	cancelled, err := client.GetServiceRequestByID(ctx, cancelRequest)
	if err != nil || cancelled.GetValue().GetStatus() != "cancelled" {
		t.Fatalf("persist cancelled request: %+v, %v", cancelled, err)
	}
	var cancelEvents int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").
		Where("aggregate_id = ? AND event_type = ?", cancelledReqID, "hr_service_requests.cancel").
		Count(&cancelEvents).Error; err != nil || cancelEvents != 1 {
		t.Fatalf("cancellation outbox: count=%d, error=%v", cancelEvents, err)
	}
	if _, err := client.CancelServiceRequest(ctx, cancelRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat cancellation: %v", err)
	}
	if _, err := client.AssignServiceRequest(ctx, cancelRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("assign cancelled request: %v", err)
	}
	if _, err := client.UpdateServiceRequest(ctx, &pb.UpdateServiceRequestRequest{
		Id: reqID, Input: &pb.ServiceRequestInput{AssignedPrincipalId: text(assignee)},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"assigned_principal_id"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("assign from CRUD = %v", err)
	}
	policy.mu.Lock()
	policy.principalID = assignee
	policy.mu.Unlock()
	assignRequest := &common.IDRequest{Id: reqID}
	assigned, err := client.AssignServiceRequest(ctx, assignRequest)
	if err != nil || assigned.GetValue().GetStatus() != "in_progress" ||
		assigned.GetValue().GetAssignedPrincipalId() != assignee {
		t.Fatalf("assign request: %+v, %v", assigned, err)
	}
	if _, err := client.UpdateServiceRequest(ctx, &pb.UpdateServiceRequestRequest{
		Id: reqID, Input: &pb.ServiceRequestInput{DueAt: timestamppb.New(time.Now().UTC().Add(72 * time.Hour))},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"due_at"}},
	}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("update assigned service request = %v", err)
	}
	if _, err := client.AssignServiceRequest(ctx, assignRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat assignment = %v", err)
	}
	resolve := &pb.ResolveServiceRequestRequest{Id: reqID, ResolutionDocumentVersionReferenceId: docID}
	if _, err := client.ResolveServiceRequest(ctx, resolve); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("resolve without verified evidence = %v", err)
	}
	resolve.ResolutionDocumentVersionReferenceId = resolutionID
	requestRow := database.WithContext(ctx).Table("hr_service_requests").Where("id = ?", reqID)
	openedAt := assigned.GetValue().GetOpenedAt().AsTime()
	if err := requestRow.Update("opened_at", time.Now().UTC().Add(time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	beforeVerification := resolutionVerifier.calls.Load()
	if _, err := client.ResolveServiceRequest(ctx, resolve); status.Code(err) != codes.FailedPrecondition || resolutionVerifier.calls.Load() != beforeVerification {
		t.Fatalf("resolve before request opening must skip DRM verification: %v", err)
	}
	if err := requestRow.Update("opened_at", openedAt).Error; err != nil {
		t.Fatal(err)
	}
	resolved, err := client.ResolveServiceRequest(ctx, resolve)
	if err != nil || resolved.GetValue().GetStatus() != "resolved" || resolved.GetValue().GetResolvedAt() == nil ||
		resolved.GetValue().GetResolutionDocumentVersionReferenceId() != resolutionID {
		t.Fatalf("resolve HR request: %+v, %v", resolved, err)
	}
	if _, err := client.CloseServiceRequest(ctx, assignRequest); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("close by assignee = %v", err)
	}
	policy.mu.Lock()
	policy.principalID = requesterID
	policy.mu.Unlock()
	resolvedAt := resolved.GetValue().GetResolvedAt().AsTime()
	if err := requestRow.Update("resolved_at", resolvedAt.Add(time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := client.CloseServiceRequest(ctx, assignRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("close before recorded resolution time: %v", err)
	}
	if err := requestRow.Update("resolved_at", resolvedAt).Error; err != nil {
		t.Fatal(err)
	}
	closed, err := client.CloseServiceRequest(ctx, assignRequest)
	if err != nil || closed.GetValue().GetStatus() != "closed" ||
		closed.GetValue().GetResolutionDocumentVersionReferenceId() != resolutionID {
		t.Fatalf("close HR request: %+v, %v", closed, err)
	}
	if _, err := client.CloseServiceRequest(ctx, assignRequest); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("repeat closure = %v", err)
	}
	if _, err := client.CreateServiceRequestEvent(ctx, &pb.CreateServiceRequestEventRequest{Input: &pb.ServiceRequestEventInput{
		ServiceRequestId: &reqID, EventType: text("commented"), IdempotencyKey: text("idem-srev-closed-" + reqID),
	}}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("comment on closed request = %v", err)
	}
	var requestEvents int64
	if err := database.WithContext(ctx).Table("hr_outbox_events").Where("aggregate_id = ?", reqID).Count(&requestEvents).Error; err != nil || requestEvents != 5 {
		t.Fatalf("request outbox events = %d, error = %v", requestEvents, err)
	}

}
