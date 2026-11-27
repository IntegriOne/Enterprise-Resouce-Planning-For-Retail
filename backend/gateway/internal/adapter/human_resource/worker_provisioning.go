package human_resource

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"github.com/integrione/erp-retail/backend/services/human_resource/contract/provisioning"
)

const workerLinkPermission = "iam.users.write"
const workerLinkConfirmationTimeout = 3 * time.Second

type WorkerProvisioner struct {
	adapter *Adapter
	key     []byte
}

func NewWorkerProvisioner(adapter *Adapter, key []byte) *WorkerProvisioner {
	return &WorkerProvisioner{adapter: adapter, key: bytes.Clone(key)}
}

func (p *WorkerProvisioner) readWorker(ctx context.Context, workerID string) (*personnel.Worker, error) {
	readCtx, err := p.adapter.authorizeRead(ctx, workerReadPermission)
	if err != nil {
		return nil, err
	}
	response, err := p.adapter.personnel.GetWorkerByID(readCtx, &common.IDRequest{Id: workerID})
	if err != nil {
		return nil, mapReadError(err, "worker")
	}
	worker := response.GetItem()
	if worker.GetId() != workerID || worker.GetPersonPartyReferenceId() == "" {
		return nil, middleware.ErrUnavailable("invalid human resource worker response", nil)
	}
	return worker, nil
}

func (p *WorkerProvisioner) PrepareWorkerPrincipalLink(ctx context.Context, workerID string) (string, error) {
	if p.adapter.personnel == nil || len(p.key) < 32 {
		return "", middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	if _, err := p.adapter.authorizeSensitiveEmployee(ctx, workerLinkPermission); err != nil {
		return "", err
	}
	worker, err := p.readWorker(ctx, workerID)
	if err != nil {
		return "", err
	}
	if worker.GetStatus() != "active" || worker.GetIamPrincipalReferenceId() != "" {
		return "", middleware.ErrConflict("worker is ineligible for account provisioning", nil)
	}
	return worker.GetId(), nil
}

func (p *WorkerProvisioner) CompleteWorkerPrincipalLink(ctx context.Context, workerID, userID string) error {
	rpcCtx, err := p.adapter.authorizeSensitiveEmployee(ctx, workerLinkPermission)
	if err != nil {
		return err
	}
	worker, err := p.readWorker(ctx, workerID)
	if err != nil {
		return err
	}
	if worker.GetIamPrincipalReferenceId() == userID {
		return nil
	}
	session, _ := middleware.ResolvedSessionFromContext(rpcCtx)
	evidenceID, err := provisioning.WorkerLinkEvidence(p.key, workerID, worker.GetPersonPartyReferenceId(), userID, session.Caller.PrincipalID)
	if err != nil {
		return middleware.ErrUnavailable("invalid worker provisioning evidence", err)
	}
	response, rpcErr := p.adapter.personnel.LinkWorkerPrincipal(rpcCtx, &personnel.LinkWorkerPrincipalRequest{
		WorkerId: workerID, PrincipalId: userID, EvidenceId: evidenceID,
	})
	if _, err = workerMutationResponse(response, rpcErr, workerID, "active"); err == nil &&
		response.GetItem().GetIamPrincipalReferenceId() == userID {
		return nil
	}
	gwErr, classified := middleware.AsGatewayError(err)
	conflict := classified && gwErr.Code == middleware.CodeConflict
	if classified && !conflict && gwErr.Code != middleware.CodeUnavailable {
		return err
	}
	// Periksa apakah tautan sudah tersimpan sebelum respons hilang atau permintaan lain menang.
	confirmCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), workerLinkConfirmationTimeout)
	defer cancel()
	linked, readErr := p.readWorker(confirmCtx, workerID)
	if readErr == nil && linked.GetIamPrincipalReferenceId() == userID {
		return nil
	}
	if conflict {
		return err
	}
	return middleware.ErrUnavailable("worker account link could not be confirmed", errors.Join(provisioning.ErrLinkUnconfirmed, err))
}
