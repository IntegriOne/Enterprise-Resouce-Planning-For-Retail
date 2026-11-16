// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/onboarding"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func TestStartOnboardingPlanRequiresTaskAndPreservesAudit(t *testing.T) {
	t.Parallel()
	h := mockdb.New()
	plans := mockdb.NewStore[entity.OnboardingPlan, string](h, func(v *entity.OnboardingPlan) string { return v.ID })
	tasks := mockdb.NewStore[entity.OnboardingTask, string](h, func(v *entity.OnboardingTask) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), OnboardingPlan: plans, OnboardingTask: tasks})
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	plan, err := svc.CreateOnboardingPlan(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartOnboardingPlan(ctx, plan.ID); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("start without tasks: %v", err)
	}
	if len(h.Events) != 1 {
		t.Fatal("rejected start must not create an outbox event")
	}
	_, err = svc.CreateOnboardingTask(ctx, service.CreateOnboardingTaskInput{
		OnboardingPlanID: plan.ID, TaskCode: "ORIENTATION", TaskName: "Orientation", TaskOwnerType: "worker",
	})
	if err != nil {
		t.Fatal(err)
	}
	started, err := svc.StartOnboardingPlan(ctx, plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if started.Status != "in_progress" || started.StartedAt == nil || started.StartedByPrincipalID == nil ||
		started.CompletedAt != nil || started.CompletedByPrincipalID != nil {
		t.Fatalf("invalid started plan: %+v", started)
	}
	if _, err := svc.StartOnboardingPlan(ctx, plan.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeated start: %v", err)
	}
	if len(h.Events) != 3 || plans.Rows[plan.ID].Status != "in_progress" {
		t.Fatal("start must persist once with its outbox event")
	}
}

func TestServiceCRUDAndPagination(t *testing.T) {
	t.Parallel()
	h, svc, store := setup(t)
	ctx := context.Background()
	created, err := svc.CreateOnboardingPlan(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Rows) != 1 || len(h.Events) != 1 || h.IntegrityCalls != 1 {
		t.Fatal("create must atomically persist record and outbox after integrity validation")
	}
	key := created.ID
	loaded, err := svc.GetOnboardingPlanByID(ctx, key)
	if err != nil || loaded == nil {
		t.Fatalf("get: %v", err)
	}
	replacement := "UPDATED"
	_, err = svc.UpdateOnboardingPlan(ctx, key, servicePatch(replacement))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err = svc.GetOnboardingPlanByID(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.PlanCode != replacement {
		t.Fatal("patch value was not persisted")
	}
	page, err := svc.GetAllOnboardingPlans(ctx, 1, 0)
	if err != nil || len(page.Items) != 1 || page.Total != 1 || page.Limit != 1 {
		t.Fatalf("page: %+v %v", page, err)
	}
	empty, err := svc.GetAllOnboardingPlans(ctx, 1, 1)
	if err != nil || empty.Items == nil || len(empty.Items) != 0 {
		t.Fatalf("empty page: %+v %v", empty, err)
	}
	_, err = svc.GetAllOnboardingPlans(ctx, -1, 0)
	if !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatal("negative pagination accepted")
	}
	if err := svc.DeleteOnboardingPlan(ctx, key); err != nil {
		t.Fatal(err)
	}
	if len(store.Rows) != 0 || len(h.Events) != 3 {
		t.Fatal("delete must persist alongside its event")
	}
}
func servicePatch(value string) service.UpdateOnboardingPlanInput {
	return service.UpdateOnboardingPlanInput{PlanCode: operation.Field[string]{Set: true, Value: &value}}
}
func TestServiceFailureAndRollback(t *testing.T) {
	t.Parallel()
	cases := []string{"authorize", "integrity", "insert", "outbox", "outbox_read", "commit"}
	for _, stage := range cases {
		t.Run(stage, func(t *testing.T) {
			h, svc, store := setup(t)
			cause := errors.New("injected failure")
			h.Failures[stage] = cause
			result, err := svc.CreateOnboardingPlan(context.Background(), input())
			if err != cause || result != nil {
				t.Fatalf("error propagation: %v", err)
			}
			if len(store.Rows) != 0 || len(h.Events) != 0 {
				t.Fatal("failed operation left committed data")
			}
			if stage == "authorize" && (h.WriteCalls != 0 || h.TransactionCalls != 0) {
				t.Fatal("denied operation reached database transaction")
			}
		})
	}
}
func TestServiceCancellationAndScope(t *testing.T) {
	t.Parallel()
	h, svc, store := setup(t)
	ctx := context.Background()
	created, err := svc.CreateOnboardingPlan(ctx, input())
	if err != nil {
		t.Fatal(err)
	}
	key := created.ID
	h.DeniedObject = created.ID
	if _, err := svc.GetOnboardingPlanByID(ctx, key); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("wrong object scope: %v", err)
	}
	h.DeniedObject = ""
	h.Scope = operation.Scope{}
	if _, err := svc.GetAllOnboardingPlans(ctx, 20, 0); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("missing scope: %v", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := svc.CreateOnboardingPlan(canceled, input()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	if len(store.Rows) != 1 || len(h.Events) != 1 {
		t.Fatal("denied calls changed committed data")
	}
}
