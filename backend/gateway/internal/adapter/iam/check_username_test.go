package iam

import (
	"context"
	"strings"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/integrione/erp-retail/backend/gateway/internal/browserauth"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	"github.com/integrione/erp-retail/backend/gateway/internal/source_connections"
	iamv1 "github.com/integrione/erp-retail/backend/services/iam/contract/gen/iam/v1"
	"github.com/integrione/erp-retail/backend/shared/authz"
)

// TestCheckUsernameAvailability menguji query checkUsername (available, taken, validasi format/email, rate limit per IP, dan IAM outage).
func TestCheckUsernameAvailability(t *testing.T) {
	t.Parallel()

	store := newMemStore()
	rpcCalls := 0
	usersMock := &mockUserClient{
		checkUsernameFn: func(_ context.Context, in *iamv1.CheckUsernameAvailabilityRequest) (*iamv1.CheckUsernameAvailabilityResponse, error) {
			rpcCalls++
			raw := strings.ToLower(strings.TrimSpace(in.GetUsername()))
			switch raw {
			case "fresh.candidate":
				return &iamv1.CheckUsernameAvailabilityResponse{
					Available:          true,
					NormalizedUsername: raw,
					Reason:             "username is available",
				}, nil
			case "taken.staff", "taken.buyer":
				return &iamv1.CheckUsernameAvailabilityResponse{
					Available:          false,
					NormalizedUsername: raw,
					Reason:             "username is already taken",
				}, nil
			default:
				st, _ := status.New(codes.InvalidArgument, "invalid username format").WithDetails(&errdetails.ErrorInfo{
					Reason: "iam.identity.invalid_username",
				})
				return nil, st.Err()
			}
		},
	}

	authMgr, err := middleware.NewAuthManager(
		store,
		middleware.NewSnapshotLoader(dummyRefresher{}, time.Second, 30*time.Second),
		middleware.SessionConfig{CookieSecure: true},
	)
	if err != nil {
		t.Fatalf("NewAuthManager() error = %v", err)
	}
	authzSvc := middleware.NewAuthorizer(authMgr, nil, 10*time.Minute)
	rv := browserauth.NewRedirectValidator([]string{"/", "/identity"})

	adp, err := NewAdapter(
		source_connections.IAMClientSet{Users: usersMock},
		authMgr,
		authzSvc,
		rv,
		Config{LoginRateMax: 3, LoginRateWindow: time.Minute},
	)
	if err != nil {
		t.Fatalf("NewAdapter() error = %v", err)
	}

	anonCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		ClientIP: "198.51.100.10",
	})

	// 1. Username tersedia.
	avail, err := adp.CheckUsername(anonCtx, "  Fresh.Candidate ")
	if err != nil {
		t.Fatalf("CheckUsername(available) error = %v", err)
	}
	if !avail.Available || avail.NormalizedUsername != "fresh.candidate" || avail.Message != "username is available" {
		t.Fatalf("unexpected available payload = %+v", avail)
	}

	// 2. Username terpakai (internal/customer) memulangkan alasan identik tanpa membocorkan PII/user_kind.
	taken, err := adp.CheckUsername(anonCtx, "TAKEN.STAFF")
	if err != nil {
		t.Fatalf("CheckUsername(taken) error = %v", err)
	}
	if taken.Available || taken.NormalizedUsername != "taken.staff" || taken.Message != "username is already taken" {
		t.Fatalf("unexpected taken payload = %+v", taken)
	}

	// 3. Input berformat email ditolak eksplisit sebagai INVALID_ARGUMENT.
	if _, err := adp.CheckUsername(anonCtx, "buyer@integrione.local"); err == nil {
		t.Fatal("expected INVALID_ARGUMENT when passing email format to CheckUsername")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeInvalidArgument {
		t.Fatalf("expected INVALID_ARGUMENT error, got %v", err)
	}

	// 4. Melewati LoginRateMax (=3) untuk IP yang sama -> TOO_MANY_REQUESTS sebelum menembak IAM RPC ke-4.
	callsBeforeLimit := rpcCalls
	if _, err := adp.CheckUsername(anonCtx, "fourth.call"); err == nil {
		t.Fatal("expected TOO_MANY_REQUESTS after exceeding rate limit")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeTooManyRequests {
		t.Fatalf("expected TOO_MANY_REQUESTS, got %v", err)
	}
	if rpcCalls != callsBeforeLimit {
		t.Fatalf("expected rate-limited CheckUsername not to invoke IAM gRPC (%d != %d)", rpcCalls, callsBeforeLimit)
	}

	// 5. IP berbeda tidak ikut terblokir limit IP sebelumnya.
	otherIPCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		ClientIP: "198.51.100.11",
	})
	if res, err := adp.CheckUsername(otherIPCtx, "taken.buyer"); err != nil || res.Available || res.Message != taken.Message {
		t.Fatalf("expected independent IP bucket to succeed with identical non-leaking message, got res=%+v err=%v", res, err)
	}
	_ = authz.ScopeGlobal

	// 6. IAM stub nil / outage -> fail-closed UNAVAILABLE.
	downAdp, err := NewAdapter(
		source_connections.IAMClientSet{},
		authMgr,
		authzSvc,
		rv,
		Config{LoginRateMax: 5, LoginRateWindow: time.Minute},
	)
	if err != nil {
		t.Fatalf("NewAdapter(down) error = %v", err)
	}
	downCtx := middleware.WithRequestMetadata(context.Background(), middleware.RequestMetadata{
		ClientIP: "198.51.100.12",
	})
	if _, err := downAdp.CheckUsername(downCtx, "fresh.candidate"); err == nil {
		t.Fatal("expected UNAVAILABLE when IAM Users client is nil")
	} else if gwErr, ok := middleware.AsGatewayError(err); !ok || gwErr.Code != middleware.CodeUnavailable {
		t.Fatalf("expected UNAVAILABLE, got %v", err)
	}
}
