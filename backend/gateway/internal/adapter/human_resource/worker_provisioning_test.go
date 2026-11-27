package human_resource

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/contract/provisioning"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"github.com/integrione/erp-retail/backend/shared/authsession"
	"github.com/integrione/erp-retail/backend/shared/authz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type workerLinkFixture struct {
	personnel.PersonnelServiceClient
	actorCancel context.CancelFunc
	linkCode    codes.Code
	read        func(context.Context, *common.IDRequest) (*personnel.WorkerResponse, error)
}

func (f workerLinkFixture) GetWorkerByID(ctx context.Context, req *common.IDRequest, _ ...grpc.CallOption) (*personnel.WorkerResponse, error) {
	return f.read(ctx, req)
}

func (f workerLinkFixture) LinkWorkerPrincipal(context.Context, *personnel.LinkWorkerPrincipalRequest, ...grpc.CallOption) (*personnel.WorkerResponse, error) {
	f.actorCancel()
	return nil, status.Error(f.linkCode, "confirmation interrupted")
}

func TestWorkerPrincipalLinkConfirmationIsBoundedAfterCancellation(t *testing.T) {
	t.Parallel()
	const workerID = "10000000-0000-4000-8000-000000000001"
	const partyID = "10000000-0000-4000-8000-000000000002"
	const userID = "10000000-0000-4000-8000-000000000003"
	const operatorID = "10000000-0000-4000-8000-000000000004"
	for _, tc := range []struct {
		name              string
		linkCode          codes.Code
		observedPrincipal string
		waitForDeadline   bool
		wantConflict      bool
	}{
		{name: "already linked after lost response", linkCode: codes.Unavailable, observedPrincipal: userID},
		{name: "deadline", linkCode: codes.Unavailable, waitForDeadline: true},
		{name: "concurrent retry linked same account", linkCode: codes.FailedPrecondition, observedPrincipal: userID},
		{name: "conflict with another account", linkCode: codes.FailedPrecondition, observedPrincipal: partyID, wantConflict: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			auth, err := middleware.NewAuthManager(&sessionStoreFixture{}, middleware.NewSnapshotLoader(nil, 0, 0), middleware.SessionConfig{})
			if err != nil {
				t.Fatal(err)
			}
			adapter := NewAdapter(auth, middleware.NewAuthorizer(auth, nil, 0), nil)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			reads := 0
			adapter.personnel = workerLinkFixture{actorCancel: cancel, linkCode: tc.linkCode, read: func(rpcCtx context.Context, req *common.IDRequest) (*personnel.WorkerResponse, error) {
				reads++
				worker := &personnel.Worker{Id: req.GetId(), PersonPartyReferenceId: partyID, Status: "active"}
				if reads == 2 {
					deadline, exists := rpcCtx.Deadline()
					if !exists || time.Until(deadline) > workerLinkConfirmationTimeout || rpcCtx.Err() != nil {
						t.Fatalf("confirmation did not receive an independent and bounded context: %v, %v", deadline, rpcCtx.Err())
					}
					if tc.waitForDeadline {
						<-rpcCtx.Done()
						return nil, rpcCtx.Err()
					}
					worker.IamPrincipalReferenceId = new(string(tc.observedPrincipal))
				}
				return &personnel.WorkerResponse{Item: worker}, nil
			}}
			global := authz.ScopeGrant{Type: authz.ScopeGlobal}
			ctx = middleware.WithResolvedSession(ctx, &middleware.ResolvedSession{
				Token: &authsession.SessionTokenRecord{MFAAuthenticated: true, IssuedAtUnix: time.Now().Unix()},
				Snapshot: &authz.AccessSnapshot{Status: authz.StatusActive, Permissions: []authz.PermissionGrant{
					{Code: workerLinkPermission, Scope: global},
					{Code: workerReadPermission, Scope: global},
				}},
				Caller: authcontext.CallerContext{PrincipalID: operatorID, Authenticated: true, UserKind: "employee"},
			})
			err = NewWorkerProvisioner(adapter, bytes.Repeat([]byte{1}, 32)).CompleteWorkerPrincipalLink(ctx, workerID, userID)
			if reads != 2 {
				t.Fatalf("link recovery: reads=%d, error=%v", reads, err)
			}
			switch {
			case tc.wantConflict:
				if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeConflict {
					t.Fatalf("expected conflict for another account, got %v", err)
				}
			case tc.waitForDeadline:
				if !errors.Is(err, provisioning.ErrLinkUnconfirmed) {
					t.Fatalf("expected unconfirmed link, got %v", err)
				}
			default:
				if err != nil {
					t.Fatalf("matching account was not recovered: %v", err)
				}
			}
		})
	}
}
