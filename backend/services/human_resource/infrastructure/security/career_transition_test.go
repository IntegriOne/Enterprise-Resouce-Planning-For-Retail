package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

type careerPathReaderStub struct {
	path *entity.CareerPath
}

func (s *careerPathReaderStub) FindByIDForUpdate(context.Context, string) (*entity.CareerPath, error) {
	if s.path == nil {
		return nil, hrerrors.ErrNotFound
	}
	return s.path, nil
}

type careerTransitionStub struct {
	matches []entity.CareerPathTransition
}

func (s *careerTransitionStub) FindByIDForUpdate(context.Context, string) (*entity.CareerPathTransition, error) {
	if len(s.matches) == 0 {
		return nil, hrerrors.ErrNotFound
	}
	copy := s.matches[0]
	return &copy, nil
}

func (s *careerTransitionStub) SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.CareerPathTransition, error) {
	return s.matches, nil
}

type jobMapStub struct {
	jobs map[string]entity.Job
}

func (s *jobMapStub) FindByIDForUpdate(_ context.Context, id string) (*entity.Job, error) {
	job, ok := s.jobs[id]
	if !ok {
		return nil, hrerrors.ErrNotFound
	}
	copy := job
	return &copy, nil
}

func TestCareerPathTransitionValidator(t *testing.T) {
	t.Parallel()
	familyID := "00000000-0000-4000-8000-000000000301"
	otherFamilyID := "00000000-0000-4000-8000-000000000302"
	pathID := "00000000-0000-4000-8000-000000000303"
	fromJobID := "00000000-0000-4000-8000-000000000304"
	toJobID := "00000000-0000-4000-8000-000000000305"
	otherJobID := "00000000-0000-4000-8000-000000000306"
	docID := "00000000-0000-4000-8000-000000000307"

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	path := &entity.CareerPath{
		ID: pathID, PathCode: "CP-1", PathVersion: 1, PathName: "Retail",
		JobFamilyID: familyID, ValidFrom: start, Status: "draft",
	}
	jobs := &jobMapStub{jobs: map[string]entity.Job{
		fromJobID:  {ID: fromJobID, JobFamilyID: familyID, ValidFrom: start, IsActive: true},
		toJobID:    {ID: toJobID, JobFamilyID: familyID, ValidFrom: start, IsActive: true},
		otherJobID: {ID: otherJobID, JobFamilyID: otherFamilyID, ValidFrom: start, IsActive: true},
	}}
	aspirations := &positionCounterStub{}
	v := &careerPathTransitionValidator{
		paths: &careerPathReaderStub{path: path}, jobs: jobs,
		transitions: &careerTransitionStub{}, aspirations: aspirations,
	}

	transition := &entity.CareerPathTransition{
		ID: "00000000-0000-4000-8000-000000000308", CareerPathID: pathID,
		FromJobID: fromJobID, ToJobID: toJobID, MovementType: "promotion",
		MinimumTenureMonths: 12, RequirementsDocumentVersionReferenceID: docID, RequirementsDigest: []byte{1},
	}
	if err := v.ValidateMutation(t.Context(), "hr_career_path_transitions", nil, transition); err != nil {
		t.Fatalf("valid transition rejected: %v", err)
	}

	wrongFamily := *transition
	wrongFamily.ToJobID = otherJobID
	if err := v.ValidateMutation(t.Context(), "hr_career_path_transitions", nil, &wrongFamily); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("cross-family job transition accepted: %v", err)
	}

	aspirations.count = 1
	if err := v.ValidateMutation(t.Context(), "hr_career_path_transitions", transition, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("referenced transition deleted: %v", err)
	}
	aspirations.count = 0

	path.Status = "approved"
	if err := v.ValidateMutation(t.Context(), "hr_career_path_transitions", nil, transition); !errors.Is(err, hrerrors.ErrWorkflowRequired) {
		t.Fatalf("transition created on approved career path: %v", err)
	}
}
