package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/personnel"
	f "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func TestAssignmentReferences(t *testing.T) {
	t.Parallel()
	worker, position, assignment := f.Worker(), f.Position(), f.Assignment()
	if err := service.ValidateAssignment(&worker, &position, &assignment); err != nil {
		t.Fatal(err)
	}
	position.LegalEntityReferenceID = f.ID(999)
	if err := service.ValidateAssignment(&worker, &position, &assignment); err == nil {
		t.Fatal("foreign employer accepted")
	}
	position = f.Position()
	worker.Status = entity.WorkerStatusTerminated
	if err := service.ValidateAssignment(&worker, &position, &assignment); err == nil {
		t.Fatal("old employment reused")
	}
}

func TestClosePrimaryAssignment(t *testing.T) {
	t.Parallel()
	effective := f.Date()
	nextDay := effective.AddDate(0, 0, 1)
	dayBefore := effective.AddDate(0, 0, -1)
	earlier := dayBefore.AddDate(0, 0, -1)
	for _, scenario := range []struct {
		name   string
		change func(*entity.Worker, *entity.WorkerAssignment)
		valid  bool
	}{
		{"open source", func(_ *entity.Worker, _ *entity.WorkerAssignment) {}, true},
		{"effective on service start", func(w *entity.Worker, _ *entity.WorkerAssignment) { w.ServiceStartDate = &effective }, true},
		{"effective before service start", func(w *entity.Worker, _ *entity.WorkerAssignment) { w.ServiceStartDate = &nextDay }, false},
		{"source ends on previous day", func(_ *entity.Worker, a *entity.WorkerAssignment) { a.ValidTo = &dayBefore }, true},
		{"source ends too early", func(_ *entity.Worker, a *entity.WorkerAssignment) { a.ValidTo = &earlier }, false},
		{"source no longer primary", func(w *entity.Worker, _ *entity.WorkerAssignment) { w.PrimaryAssignmentID = nil }, false},
		{"employment terminated", func(w *entity.Worker, _ *entity.WorkerAssignment) { w.TerminationDate = &dayBefore }, false},
		{"source approval missing", func(_ *entity.Worker, a *entity.WorkerAssignment) { a.ApprovedAt = nil }, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			worker, source := f.Worker(), f.Assignment()
			approver := f.ID(99)
			worker.PrimaryAssignmentID = &source.ID
			source.ApprovedAt, source.ApprovedByPrincipalID = &source.ValidFrom, &approver
			scenario.change(&worker, &source)
			originalEnd := source.ValidTo
			closed, err := service.ClosePrimaryAssignment(&worker, &source, effective)
			if !scenario.valid {
				if err == nil || closed != nil {
					t.Fatalf("invalid source accepted: %+v, %v", closed, err)
				}
				return
			}
			if err != nil || closed == nil || closed.ValidTo == nil || !closed.ValidTo.Equal(dayBefore) {
				t.Fatalf("source should close on previous day: %+v, %v", closed, err)
			}
			if source.ValidTo != originalEnd {
				t.Fatal("calculating replacement dates mutated the original assignment")
			}
		})
	}
}
