package human_resource

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"testing"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc"
)

type sessionStoreFixture struct{ middleware.SessionStorePort }

type personnelFixture struct {
	personnel.PersonnelServiceClient
	calls       int
	pageRequest *common.PageRequest
}

func (f *personnelFixture) GetWorkerByID(_ context.Context, request *common.IDRequest, _ ...grpc.CallOption) (*personnel.WorkerResponse, error) {
	f.calls++
	return &personnel.WorkerResponse{Item: &personnel.Worker{Id: request.Id, WorkerNumber: "W-1", Status: "active"}}, nil
}

func (f *personnelFixture) GetAllWorkers(_ context.Context, request *common.PageRequest, _ ...grpc.CallOption) (*personnel.WorkersResponse, error) {
	f.pageRequest = request
	return &personnel.WorkersResponse{
		Items: []*personnel.Worker{{Id: "10000000-0000-4000-8000-000000000001"}},
		Page:  &common.PageInfo{Total: 2, Limit: request.GetLimit(), Offset: request.GetOffset()},
	}, nil
}

func TestWorkerRequiresReadPermissionBeforeRPC(t *testing.T) {
	t.Parallel()
	auth, err := middleware.NewAuthManager(&sessionStoreFixture{}, middleware.NewSnapshotLoader(nil, 0, 0), middleware.SessionConfig{})
	if err != nil {
		t.Fatal(err)
	}
	client := &personnelFixture{}
	a := NewAdapter(auth, middleware.NewAuthorizer(auth, nil, 0), nil)
	a.personnel = client
	const workerID = "10000000-0000-4000-8000-000000000001"
	sess := &middleware.ResolvedSession{
		Token:    &authsession.SessionTokenRecord{},
		Snapshot: &authz.AccessSnapshot{Status: authz.StatusActive},
		Caller:   authcontext.CallerContext{Authenticated: true, UserKind: "employee"},
	}
	ctx := middleware.WithResolvedSession(t.Context(), sess)
	if _, err := a.Worker(ctx, workerID); err == nil || client.calls != 0 {
		t.Fatal("worker was fetched without a grant")
	}
	if _, err := a.Workers(ctx, 1, 1, nil, nil); err == nil || client.pageRequest != nil {
		t.Fatal("worker list was fetched without a grant")
	}
	sess.Snapshot.Permissions = []authz.PermissionGrant{{
		Code: workerReadPermission, Scope: authz.ScopeGrant{Type: "legal_entity", ID: "20000000-0000-4000-8000-000000000001"},
	}}
	item, err := a.Worker(ctx, workerGIDPrefix+workerID)
	if err != nil || item == nil || item.RawID != workerID || client.calls != 1 {
		t.Fatalf("valid scoped call: item=%+v, calls=%d, error=%v", item, client.calls, err)
	}
	if _, err := a.Workers(ctx, 101, 0, nil, nil); err == nil || client.pageRequest != nil {
		t.Fatal("invalid pagination reached HRM")
	}
	page, err := a.Workers(ctx, 1, 1, nil, nil)
	if err != nil || page == nil || page.TotalCount != 2 || len(page.Items) != 1 ||
		client.pageRequest.GetLimit() != 1 || client.pageRequest.GetOffset() != 1 {
		t.Fatalf("scoped page: page=%+v, request=%+v, error=%v", page, client.pageRequest, err)
	}

}

func TestUpstreamAuthenticationErrorKeepsTransportDetailsPrivate(t *testing.T) {
	sensitive := "private-host:9003 upstream authentication failure"
	err := mapReadError(status.Error(codes.Unauthenticated, sensitive), "worker")
	gw, ok := middleware.AsGatewayError(err)
	if !ok || gw.Code != middleware.CodeUnavailable {
		t.Fatalf("expected unavailable, got %v", err)
	}
	if gw.Err == nil || !strings.Contains(gw.Err.Error(), "Unauthenticated") || strings.Contains(gw.Err.Error(), sensitive) {
		t.Fatalf("expected sanitized transport status, got %v", gw.Err)
	}
}
