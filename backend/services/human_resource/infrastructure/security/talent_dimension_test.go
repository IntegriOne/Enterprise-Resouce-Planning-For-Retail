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

const talentTestFrameworkID = "11111111-1111-4111-8111-111111111111"

type talentFrameworkStub struct {
	value *entity.TalentFramework
	locks int
}

func (s *talentFrameworkStub) FindByIDForUpdate(_ context.Context, id string) (*entity.TalentFramework, error) {
	s.locks++
	if s.value == nil || id != s.value.ID {
		return nil, nil
	}
	return s.value, nil
}

type talentDimensionStub struct{ rows []entity.TalentDimension }

func (s *talentDimensionStub) SelectWhere(_ context.Context, filters []domainrepo.Filter, _, _ int) ([]entity.TalentDimension, error) {
	if len(filters) != 2 || filters[0].Column != "framework_id" || filters[1].Column != "dimension_code" {
		return nil, hrerrors.ErrInvalidInput
	}
	return s.rows, nil
}

type talentCycleStub struct{ count int64 }

func (s *talentCycleStub) CountWhere(_ context.Context, filters []domainrepo.Filter) (int64, error) {
	if len(filters) != 1 || filters[0].Column != "framework_id" || filters[0].Values[0] != talentTestFrameworkID {
		return 0, hrerrors.ErrInvalidInput
	}
	return s.count, nil
}

func TestTalentDimensionValidatorFreezesUsedFramework(t *testing.T) {
	t.Parallel()
	framework := &talentFrameworkStub{value: &entity.TalentFramework{
		ID: talentTestFrameworkID, FrameworkCode: "LEADERSHIP", FrameworkVersion: 1,
		FrameworkName: "Leadership", RubricDocumentVersionReferenceID: "22222222-2222-4222-8222-222222222222",
		RubricDigest: []byte{1}, EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		ApprovedByPrincipalID: "33333333-3333-4333-8333-333333333333", ApprovedAt: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
	}}
	dimensions := &talentDimensionStub{}
	cycles := &talentCycleStub{}
	validator := &talentDimensionValidator{frameworks: framework, dimensions: dimensions, cycles: cycles}
	value := &entity.TalentDimension{
		ID: "44444444-4444-4444-8444-444444444444", FrameworkID: talentTestFrameworkID,
		DimensionCode: "LEADERSHIP", RatingAnchors: entity.JSON(`{"1":"basic","2":"developing","3":"capable","4":"strong","5":"expert"}`),
	}
	if err := validator.ValidateMutation(t.Context(), "hr_talent_dimensions", nil, value); err != nil {
		t.Fatalf("new dimension on unused framework: %v", err)
	}
	if framework.locks != 1 {
		t.Fatalf("framework lock calls = %d, want 1", framework.locks)
	}
	dimensions.rows = []entity.TalentDimension{{ID: "55555555-5555-4555-8555-555555555555"}}
	if err := validator.ValidateMutation(t.Context(), "hr_talent_dimensions", nil, value); !errors.Is(err, hrerrors.ErrAlreadyExists) {
		t.Fatalf("duplicate dimension code: %v", err)
	}
	dimensions.rows = nil
	cycles.count = 1
	for _, change := range []struct {
		name   string
		before entity.Validatable
		after  entity.Validatable
	}{
		{name: "create", after: value},
		{name: "update", before: value, after: value},
		{name: "delete", before: value},
	} {
		t.Run(change.name, func(t *testing.T) {
			if err := validator.ValidateMutation(t.Context(), "hr_talent_dimensions", change.before, change.after); !errors.Is(err, hrerrors.ErrReferenceConflict) {
				t.Fatalf("used framework mutation: %v", err)
			}
		})
	}
	cycles.count = 0
	moved := *value
	moved.FrameworkID = "66666666-6666-4666-8666-666666666666"
	if err := validator.ValidateMutation(t.Context(), "hr_talent_dimensions", value, &moved); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("dimension moved between frameworks: %v", err)
	}
	framework.value = nil
	if err := validator.ValidateMutation(t.Context(), "hr_talent_dimensions", nil, value); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("missing framework: %v", err)
	}
}
