// Package service memeriksa referensi internal hubungan kerja HRM.
package service

import (
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// ValidateAssignmentManager memeriksa identitas dan masa kerja manajer untuk seluruh periode penempatan.
func ValidateAssignmentManager(worker, manager *entity.Worker, managerID string, from time.Time, to *time.Time) error {
	if worker == nil || manager == nil || !strings.EqualFold(manager.ID, managerID) ||
		strings.EqualFold(manager.ID, worker.ID) || manager.Status != entity.WorkerStatusActive ||
		!strings.EqualFold(manager.PrimaryLegalEntityReferenceID, worker.PrimaryLegalEntityReferenceID) ||
		(manager.HireDate != nil && entity.CompareDates(from, *manager.HireDate) < 0) ||
		(manager.ServiceStartDate != nil && entity.CompareDates(from, *manager.ServiceStartDate) < 0) ||
		(manager.TerminationDate != nil && (to == nil || entity.CompareDates(*to, *manager.TerminationDate) > 0)) {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}

// ClosePrimaryAssignment mengakhiri penempatan primer lama sehari sebelum penempatan pengganti.
func ClosePrimaryAssignment(worker *entity.Worker, source *entity.WorkerAssignment, effectiveDate time.Time) (*entity.WorkerAssignment, error) {
	endDate := effectiveDate.AddDate(0, 0, -1)
	if worker == nil || source == nil || worker.Status != entity.WorkerStatusActive ||
		worker.TerminationDate != nil ||
		(worker.HireDate != nil && entity.CompareDates(effectiveDate, *worker.HireDate) < 0) ||
		(worker.ServiceStartDate != nil && entity.CompareDates(effectiveDate, *worker.ServiceStartDate) < 0) ||
		worker.PrimaryAssignmentID == nil || !strings.EqualFold(*worker.PrimaryAssignmentID, source.ID) ||
		!strings.EqualFold(worker.ID, source.WorkerID) || source.AssignmentType != "primary" ||
		source.CancelledAt != nil || source.ApprovedAt == nil || source.ApprovedByPrincipalID == nil ||
		entity.CompareDates(effectiveDate, source.ValidFrom) <= 0 ||
		(source.ValidTo != nil && entity.CompareDates(endDate, *source.ValidTo) > 0) {
		return nil, hrerrors.ErrReferenceConflict
	}
	closed := *source
	closed.ValidTo = &endDate
	return &closed, nil
}

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
