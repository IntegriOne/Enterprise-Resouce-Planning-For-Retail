// Package service memeriksa referensi internal hubungan kerja HRM.
package service

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// ValidateAssignment memeriksa satu penempatan terhadap pekerja dan posisi yang dimuat pemanggil.
// Keberadaan master Foundation, kapasitas serentak dan approval tetap diperiksa use case.
func ValidateAssignment(worker *entity.Worker, position *entity.Position, assignment *entity.WorkerAssignment) error {
	for _, record := range []entity.Validatable{worker, position, assignment} {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	if !strings.EqualFold(worker.ID, assignment.WorkerID) || !strings.EqualFold(position.ID, assignment.PositionID) || !strings.EqualFold(position.LegalEntityReferenceID, worker.PrimaryLegalEntityReferenceID) {
		return hrerrors.Invalid("assignment", "worker and employing position references must match")
	}
	if worker.Status == entity.WorkerStatusTerminated || worker.Status == entity.WorkerStatusDeceased {
		return hrerrors.Invalid("worker", "terminal employment requires a new employment relationship")
	}
	if entity.CompareDates(assignment.ValidFrom, position.ValidFrom) < 0 || (position.ValidTo != nil && (assignment.ValidTo == nil || entity.CompareDates(*assignment.ValidTo, *position.ValidTo) > 0)) {
		return hrerrors.Invalid("valid_from", "assignment period must fit position validity")
	}
	return nil
}
