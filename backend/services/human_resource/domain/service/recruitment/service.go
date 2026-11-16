// Package service memeriksa korelasi kandidat, lamaran dan hubungan kerja hasil hiring.
package service

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// ValidateHiring tidak membuat akun IAM atau menyamakan candidate ID dengan worker ID.
func ValidateHiring(application *entity.Application, candidate *entity.Candidate, requisition *entity.Requisition, worker *entity.Worker) error {
	for _, record := range []entity.Validatable{application, candidate, requisition, worker} {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	if application.Status != entity.ApplicationStatusHired || application.HiredWorkerID == nil || !strings.EqualFold(application.CandidateID, candidate.ID) || !strings.EqualFold(application.RequisitionID, requisition.ID) || !strings.EqualFold(*application.HiredWorkerID, worker.ID) {
		return hrerrors.Invalid("hired_worker_id", "hiring must link one application, candidate, requisition and worker")
	}
	if candidate.PersonPartyReferenceID != nil && !strings.EqualFold(*candidate.PersonPartyReferenceID, worker.PersonPartyReferenceID) {
		return hrerrors.Invalid("person_party_reference_id", "hired worker must represent the candidate person")
	}
	return nil
}
