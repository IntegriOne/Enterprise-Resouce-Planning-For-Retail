package service

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type accountLinkWorkerReader struct {
	domainrepo.RecordRepository[entity.Worker, string]
	worker *entity.Worker
}

func (r accountLinkWorkerReader) SelectByID(context.Context, string) (*entity.Worker, error) {
	return r.worker, nil
}

func TestVerifyWorkerAccountLinkUsesActiveWorkerAndScope(t *testing.T) {
	const id = "00000000-0000-4000-8000-000000000001"
	const legalEntity = "00000000-0000-4000-8000-000000000002"
	const account = "00000000-0000-4000-8000-000000000003"
	linkedPrincipal := account
	worker := entity.Worker{ID: id, PrimaryLegalEntityReferenceID: legalEntity, Status: entity.WorkerStatusActive, IAMPrincipalReferenceID: &linkedPrincipal}
	svc := Service{transferDependencies: Dependencies{Worker: accountLinkWorkerReader{worker: &worker}}}
	if err := svc.VerifyWorkerAccountLink(t.Context(), id, account, legalEntity); err != nil {
		t.Fatalf("active linked worker: %v", err)
	}
	if err := svc.VerifyWorkerAccountLink(t.Context(), id, account, id); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("different legal entity: %v", err)
	}
	if err := svc.VerifyWorkerAccountLink(t.Context(), id, id, legalEntity); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("different account: %v", err)
	}
	worker.Status = entity.WorkerStatusTerminated
	if err := svc.VerifyWorkerAccountLink(t.Context(), id, account, legalEntity); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("inactive worker: %v", err)
	}
}
