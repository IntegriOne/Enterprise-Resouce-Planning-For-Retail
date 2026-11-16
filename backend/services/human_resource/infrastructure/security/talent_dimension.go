package security

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	talentdomain "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/talent"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type talentFrameworkReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.TalentFramework, error)
}

type talentDimensionReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.TalentDimension, error)
}

type talentReviewCycleCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type talentDimensionValidator struct {
	frameworks talentFrameworkReader
	dimensions talentDimensionReader
	cycles     talentReviewCycleCounter
}

// TalentDimensionValidators memasang validasi dimensi pada framework HRM yang sudah disahkan.
func TalentDimensionValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.TalentFramework) != nil ||
		operation.RequireDependency(repos.TalentDimension) != nil || operation.RequireDependency(repos.TalentReviewCycle) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_talent_dimensions": &talentDimensionValidator{
			frameworks: repos.TalentFramework, dimensions: repos.TalentDimension, cycles: repos.TalentReviewCycle,
		},
	}, nil
}

// ValidateMutation membekukan rubrik setelah framework dipakai oleh siklus review.
func (v *talentDimensionValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_talent_dimensions" || operation.RequireDependency(v.frameworks) != nil ||
		operation.RequireDependency(v.dimensions) != nil || operation.RequireDependency(v.cycles) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.TalentDimension
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.TalentDimension)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	var dimension *entity.TalentDimension
	if after != nil {
		var ok bool
		dimension, ok = after.(*entity.TalentDimension)
		if !ok || dimension == nil || (previous != nil && (previous.ID != dimension.ID || previous.FrameworkID != dimension.FrameworkID)) {
			return hrerrors.ErrInvalidInput
		}
		if err := talentdomain.ValidateDimensionAnchors(dimension.RatingAnchors); err != nil {
			return err
		}
	} else {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		dimension = previous
	}
	framework, err := v.frameworks.FindByIDForUpdate(ctx, dimension.FrameworkID)
	if err != nil {
		return err
	}
	if framework == nil {
		return hrerrors.ErrReferenceConflict
	}
	if err := framework.Validate(); err != nil {
		return err
	}
	count, err := v.cycles.CountWhere(ctx, []domainrepo.Filter{{Column: "framework_id", Values: []any{framework.ID}}})
	if err != nil {
		return err
	}
	if count > 0 {
		return hrerrors.ErrReferenceConflict
	}
	if after == nil {
		return nil
	}
	matches, err := v.dimensions.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "framework_id", Values: []any{framework.ID}},
		{Column: "dimension_code", Values: []any{dimension.DimensionCode}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != dimension.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}
