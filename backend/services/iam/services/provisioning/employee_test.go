package provisioning

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/iam/domain/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type completionRequests struct {
	domainrepo.EmployeeProvisioningRepository
	request entity.EmployeeProvisioningRequest
	linked  bool
}

func (r *completionRequests) SelectByID(context.Context, string) (*entity.EmployeeProvisioningRequest, error) {
	return &r.request, nil
}

func (r *completionRequests) UpdateLinked(context.Context, string, string, string, time.Time) error {
	r.linked = true
	return nil
}

type completionUsers struct{ domainrepo.UserRepository }

func (completionUsers) SelectAccessSnapshotByUserID(_ context.Context, id string, _ time.Time) (*entity.AccessSnapshot, error) {
	return &entity.AccessSnapshot{
		UserID: id, UserKind: entity.UserKindEmployee, UserStatus: entity.UserStatusActive,
		Permissions: []entity.ScopedPermission{{PermissionCode: "iam.users.write", Scope: entity.GlobalScope()}},
	}, nil
}

type completionClock struct{}

func (completionClock) Now() time.Time { return time.Now() }

type completionVerifier struct {
	err                  error
	worker, account, org string
}

func (v *completionVerifier) VerifyWorkerAccountLink(_ context.Context, worker, account, org string) error {
	v.worker, v.account, v.org = worker, account, org
	return v.err
}

func TestCompleteRequiresCurrentHRMLink(t *testing.T) {
	operator := "10000000-0000-4000-8000-000000000001"
	account := "10000000-0000-4000-8000-000000000002"
	id := "10000000-0000-4000-8000-000000000003"
	requests := &completionRequests{request: entity.EmployeeProvisioningRequest{
		ID: id, Status: entity.EmployeeProvisioningProcessing, ProcessedByUserID: &operator, AccountUserID: &account,
		WorkerID: "10000000-0000-4000-8000-000000000004", LegalEntityID: "10000000-0000-4000-8000-000000000005",
	}}
	ctx := authcontext.WithCaller(t.Context(), authcontext.CallerContext{PrincipalID: operator,
		SessionID: id, UserKind: string(entity.UserKindEmployee), Authenticated: true})
	service := NewService(requests, completionUsers{}, nil, nil, completionClock{})
	if err := service.Complete(ctx, id, account); !errors.Is(err, iamerrors.ErrUnavailable) || requests.linked {
		t.Fatalf("completion without a verifier: %v", err)
	}
	verifier := &completionVerifier{err: iamerrors.ErrConflict}
	service.ConfigureWorkerAccountLinkVerifier(verifier)
	if err := service.Complete(ctx, id, account); !errors.Is(err, iamerrors.ErrConflict) || requests.linked {
		t.Fatalf("completion without a link: %v", err)
	}
	if verifier.worker != requests.request.WorkerID || verifier.account != account || verifier.org != requests.request.LegalEntityID {
		t.Fatal("verification used the wrong provisioning references")
	}
	verifier.err = nil
	if err := service.Complete(ctx, id, account); err != nil || !requests.linked {
		t.Fatalf("completion with a verified link: %v", err)
	}
}
