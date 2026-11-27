package main

import (
	"context"
	"crypto/hmac"

	"github.com/integrione/erp-retail/backend/services/human_resource/contract/provisioning"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/personnel"
)

type workerLinkVerifier struct {
	key []byte
}

func (v workerLinkVerifier) VerifyPrincipalLink(_ context.Context, check personnel.PrincipalLinkCheck) error {
	if len(v.key) < 32 {
		return hrerrors.ErrUnavailable
	}
	expected, err := provisioning.WorkerLinkEvidence(v.key, check.WorkerID, check.PersonPartyID, check.PrincipalID, check.RequestedByPrincipalID)
	if err != nil || !hmac.Equal([]byte(expected), []byte(check.EvidenceID)) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}
