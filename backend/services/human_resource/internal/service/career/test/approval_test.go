package service_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/career"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
)

func TestCareerPathDraftAndIndependentApproval(t *testing.T) {
	t.Parallel()
	h := mockdb.New()
	plan := mockdb.NewStore[entity.DevelopmentPlan, string](h, func(v *entity.DevelopmentPlan) string { return v.ID })
	action := mockdb.NewStore[entity.DevelopmentAction, string](h, func(v *entity.DevelopmentAction) string { return v.ID })
	path := mockdb.NewStore[entity.CareerPath, string](h, func(v *entity.CareerPath) string { return v.ID })
	transition := mockdb.NewStore[entity.CareerPathTransition, string](h, func(v *entity.CareerPathTransition) string { return v.ID })
	aspiration := mockdb.NewStore[entity.WorkerCareerAspiration, string](h, func(v *entity.WorkerCareerAspiration) string { return v.ID })
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), DevelopmentPlan: plan,
		DevelopmentAction: action, CareerPath: path, CareerPathTransition: transition, WorkerCareerAspiration: aspiration})
	if err != nil {
		t.Fatal(err)
	}
	created, err := svc.CreateCareerPath(t.Context(), service.CreateCareerPathInput{
		PathCode: "STORE", PathVersion: 1, PathName: "Store career",
		JobFamilyID: mockdb.ID(20), ValidFrom: mockdb.Now(),
	})
	if err != nil || created.Status != "draft" || created.CreatedByPrincipalID != h.Actor.PrincipalID ||
		created.ApprovedAt != nil || created.ApprovedByPrincipalID != nil || len(h.Events) != 1 {
		t.Fatalf("create draft: %+v, %v", created, err)
	}
	if _, err := svc.ApproveCareerPath(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrForbidden) {
		t.Fatalf("creator approved own path: %v", err)
	}
	if len(h.Events) != 1 {
		t.Fatal("rejected approval emitted event")
	}
	h.Actor.PrincipalID = mockdb.ID(902)
	approved, err := svc.ApproveCareerPath(t.Context(), created.ID)
	if err != nil || approved.Status != "approved" || approved.ApprovedByPrincipalID == nil ||
		*approved.ApprovedByPrincipalID != h.Actor.PrincipalID || approved.ApprovedAt == nil || len(h.Events) != 2 {
		t.Fatalf("approve career path: %+v, %v", approved, err)
	}
	if _, err := svc.ApproveCareerPath(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("repeat approval: %v", err)
	}
	if _, err := svc.UpdateCareerPath(t.Context(), created.ID, service.UpdateCareerPathInput{}); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("update approved path: %v", err)
	}
	if err := svc.DeleteCareerPath(t.Context(), created.ID); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("delete approved path: %v", err)
	}
	if len(h.Events) != 2 {
		t.Fatal("rejected mutation emitted event")
	}
}
