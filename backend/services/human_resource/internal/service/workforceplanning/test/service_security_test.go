package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/workforceplanning"
)

func TestScopeIsAppliedBeforeCountAndPagination(t *testing.T) {
	h, svc, _ := setup(t)
	ctx := context.Background()
	first, err := svc.CreateJobFamily(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	secondInput := input()
	secondInput.FamilyCode = "SECOND"
	if _, err := svc.CreateJobFamily(ctx, secondInput); err != nil {
		t.Fatal(err)
	}
	h.Scope = operation.Scope{Filters: []domainrepo.Filter{{Column: "id", Values: []any{first.ID}}}}
	page, err := svc.GetAllJobFamilies(ctx, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != first.ID {
		t.Fatalf("scope leaked in count/page: %+v", page)
	}
	mismatch, err := svc.FindJobFamilyByFamilyCode(ctx, "SECOND", 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if mismatch.Total != 0 || len(mismatch.Items) != 0 {
		t.Fatal("search filter replaced authorization scope")
	}
	if len(h.LastFilters) != 2 {
		t.Fatal("scope and search must both reach repository")
	}
}

func TestRevocationDuringCountRejectsPage(t *testing.T) {
	t.Parallel()
	h, svc, _ := setup(t)
	if _, err := svc.CreateJobFamily(t.Context(), input()); err != nil {
		t.Fatal(err)
	}
	h.RevokeAfterNextCount = true
	page, err := svc.GetAllJobFamilies(t.Context(), 20, 0)
	if !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("revoked list returned page: %+v, error: %v", page, err)
	}
	if len(page.Items) != 0 || page.Total != 0 {
		t.Fatal("revoked list leaked records or count")
	}
}

func TestConcurrentWritesAreSerializedWithOutbox(t *testing.T) {
	h, svc, store := setup(t)
	ctx := context.Background()
	record, err := svc.CreateJobFamily(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	failures := make(chan error, 16)
	var group sync.WaitGroup
	for i := 0; i < 16; i++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			_, err := svc.UpdateJobFamily(ctx, record.ID, servicePatch(fmt.Sprintf("CODE-%d", index)))
			failures <- err
		}(i)
	}
	close(start)
	group.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(store.Rows) != 1 || len(h.Events) != 17 {
		t.Fatal("lost update/event or duplicate record")
	}
	for index, event := range h.Events {
		if event.AggregateVersion != int64(index+1) {
			t.Fatal("aggregate event versions were not serialized")
		}
	}
	if record.FamilyCode != "TEST" {
		t.Fatal("caller snapshot was mutated")
	}
}

func TestRevocationBeforeCommitRollsBack(t *testing.T) {
	h, svc, store := setup(t)
	h.RevokeAfterNextOutbox = true
	if _, err := svc.CreateJobFamily(context.Background(), input()); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("expected final security recheck: %v", err)
	}
	if len(store.Rows) != 0 || len(h.Events) != 0 {
		t.Fatal("revoked write left committed state")
	}
}

func TestSensitiveStateCannotBeChangedThroughCRUD(t *testing.T) {
	_, svc, _ := setup(t)
	ctx := context.Background()
	planInput := service.CreateWorkforcePlanInput{PlanCode: "PLAN", PlanVersion: 1, LegalEntityReferenceID: mockdb.ID(91), PeriodStart: mockdb.Now(), PeriodEnd: mockdb.Now().AddDate(0, 1, 0)}
	plan, err := svc.CreateWorkforcePlan(ctx, planInput)
	if err != nil {
		t.Fatal(err)
	}
	status := "approved"
	if _, err := svc.UpdateWorkforcePlan(ctx, plan.ID, service.UpdateWorkforcePlanInput{Status: operation.Field[string]{Set: true, Value: &status}}); err == nil {
		t.Fatal("generic CRUD approved a workforce plan")
	}
}
