package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

type aspirationWorkerReader struct{ value *entity.Worker }

func (r aspirationWorkerReader) FindByIDForUpdate(context.Context, string) (*entity.Worker, error) {
	return r.value, nil
}

type aspirationJobReader struct{ value *entity.Job }

func (r aspirationJobReader) FindByIDForUpdate(context.Context, string) (*entity.Job, error) {
	return r.value, nil
}

type aspirationTransitionReader struct{ value *entity.CareerPathTransition }

func (r aspirationTransitionReader) FindByIDForUpdate(context.Context, string) (*entity.CareerPathTransition, error) {
	return r.value, nil
}

type aspirationPathReader struct{ value *entity.CareerPath }

func (r aspirationPathReader) FindByIDForUpdate(context.Context, string) (*entity.CareerPath, error) {
	return r.value, nil
}

func TestWorkerCareerAspirationValidator(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC)
	transitionID := "transition-id"
	base := entity.WorkerCareerAspiration{
		ID: "aspiration-id", WorkerID: "worker-id", TargetJobID: "job-id",
		CareerPathTransitionID: &transitionID, RecordedAt: now, RecordedByPrincipalID: "actor-id",
	}
	makeValidator := func() *workerCareerAspirationValidator {
		return &workerCareerAspirationValidator{
			workers: aspirationWorkerReader{value: &entity.Worker{Status: entity.WorkerStatusActive}},
			jobs:    aspirationJobReader{value: &entity.Job{IsActive: true, ValidFrom: now.AddDate(-1, 0, 0)}},
			transitions: aspirationTransitionReader{value: &entity.CareerPathTransition{
				ToJobID: "job-id", CareerPathID: "path-id",
			}},
			paths: aspirationPathReader{value: &entity.CareerPath{
				ApprovedByPrincipalID: "approver-id", ApprovedAt: now.Add(-time.Hour), ValidFrom: now.AddDate(-1, 0, 0),
			}},
		}
	}
	const table = "hr_worker_career_aspirations"
	if err := makeValidator().ValidateMutation(t.Context(), table, nil, &base); err != nil {
		t.Fatalf("valid aspiration rejected: %v", err)
	}
	futureJob := makeValidator()
	futureJob.jobs = aspirationJobReader{value: &entity.Job{IsActive: true, ValidFrom: now.AddDate(1, 0, 0)}}
	if err := futureJob.ValidateMutation(t.Context(), table, nil, &base); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("future job accepted without target date: %v", err)
	}
	wrongTarget := base
	wrongTarget.TargetJobID = "other-job"
	if err := makeValidator().ValidateMutation(t.Context(), table, nil, &wrongTarget); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("mismatched transition target: %v", err)
	}
	unapproved := makeValidator()
	unapproved.paths = aspirationPathReader{value: &entity.CareerPath{ValidFrom: now.AddDate(-1, 0, 0)}}
	if err := unapproved.ValidateMutation(t.Context(), table, nil, &base); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("unapproved path accepted: %v", err)
	}
	location := "external-location"
	withLocation := base
	withLocation.PreferredLocationReferenceID = &location
	if err := makeValidator().ValidateMutation(t.Context(), table, nil, &withLocation); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("unverified external location accepted: %v", err)
	}
	changedOwner := base
	changedOwner.WorkerID = "other-worker"
	if err := makeValidator().ValidateMutation(t.Context(), table, &base, &changedOwner); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("owner changed: %v", err)
	}
	withdrawn := base
	withdrawn.WithdrawnAt = &now
	if err := makeValidator().ValidateMutation(t.Context(), table, nil, &withdrawn); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("aspiration created withdrawn: %v", err)
	}
	if err := makeValidator().ValidateMutation(t.Context(), table, &withdrawn, &base); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("withdrawn aspiration changed: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := makeValidator().ValidateMutation(ctx, table, nil, &base); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}
