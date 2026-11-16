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
