package security

import (
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

// LearningValidators memasang pemeriksaan katalog dan relasi kursus yang dimiliki HRM.
func LearningValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil {
		return nil, hrerrors.ErrUnavailable
	}
	for _, dependency := range []any{repos.Worker, repos.LearningCourse, repos.LearningCourseCompetency, repos.LearningEnrollment, repos.Competency, repos.DevelopmentAction} {
		if operation.RequireDependency(dependency) != nil {
			return nil, hrerrors.ErrUnavailable
		}
	}
	return map[string]MutationValidator{
		"hr_learning_courses": &learningCourseValidator{
			courses: repos.LearningCourse, competencies: repos.LearningCourseCompetency,
			enrollments: repos.LearningEnrollment, developmentActions: repos.DevelopmentAction,
		},
		"hr_learning_course_competencies": &learningCourseCompetencyValidator{
			courses: repos.LearningCourse, competencies: repos.Competency, links: repos.LearningCourseCompetency,
		},
		"hr_learning_enrollments": &learningEnrollmentValidator{
			workers: repos.Worker, courses: repos.LearningCourse, enrollments: repos.LearningEnrollment,
		},
	}, nil
}

type learningCourseReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.LearningCourse, error)
	FindByIDForUpdate(context.Context, string) (*entity.LearningCourse, error)
}

type learningReferenceCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type learningCourseValidator struct {
	courses            learningCourseReader
	competencies       learningReferenceCounter
	enrollments        learningReferenceCounter
	developmentActions learningReferenceCounter
}

// ValidateMutation menjaga identitas versi kursus dan referensi konsumennya.
func (v *learningCourseValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_learning_courses" || operation.RequireDependency(v.courses) != nil ||
		operation.RequireDependency(v.competencies) != nil || operation.RequireDependency(v.enrollments) != nil ||
		operation.RequireDependency(v.developmentActions) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.LearningCourse
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.LearningCourse)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return v.rejectIfReferenced(ctx, previous.ID)
	}
	course, ok := after.(*entity.LearningCourse)
	if !ok || course == nil || (previous != nil && previous.ID != course.ID) {
		return hrerrors.ErrInvalidInput
	}
	if course.ProviderPartyReferenceID != nil &&
		(previous == nil || !reflect.DeepEqual(previous.ProviderPartyReferenceID, course.ProviderPartyReferenceID)) {
		return hrerrors.ErrUnavailable
	}
	matches, err := v.courses.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "course_code", Values: []any{course.CourseCode}},
		{Column: "course_version", Values: []any{course.CourseVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != course.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous == nil {
		return nil
	}
	changedDefinition := previous.CourseCode != course.CourseCode || previous.CourseVersion != course.CourseVersion ||
		previous.CourseType != course.CourseType || previous.DeliveryMode != course.DeliveryMode ||
		!reflect.DeepEqual(previous.DurationHours, course.DurationHours) ||
		previous.CredentialAwarded != course.CredentialAwarded || !reflect.DeepEqual(previous.ExpiryMonths, course.ExpiryMonths) ||
		!reflect.DeepEqual(previous.ValidFrom, course.ValidFrom) || !reflect.DeepEqual(previous.ValidTo, course.ValidTo)
	if changedDefinition || (previous.IsActive && !course.IsActive) {
		return v.rejectIfReferenced(ctx, course.ID)
	}
	return nil
}

func (v *learningCourseValidator) rejectIfReferenced(ctx context.Context, courseID string) error {
	filters := []domainrepo.Filter{{Column: "course_id", Values: []any{courseID}}}
	for _, reference := range []learningReferenceCounter{v.competencies, v.enrollments, v.developmentActions} {
		count, err := reference.CountWhere(ctx, filters)
		if err != nil {
			return err
		}
		if count > 0 {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}

type learningLinkReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.LearningCourseCompetency, error)
}

type learningCourseCompetencyValidator struct {
	courses      learningCourseReader
	competencies lockedCompetencyReader
	links        learningLinkReader
}

// ValidateMutation mengunci kursus lalu kompetensi dan menolak pasangan yang sudah dipakai.
func (v *learningCourseCompetencyValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_learning_course_competencies" || operation.RequireDependency(v.courses) != nil ||
		operation.RequireDependency(v.competencies) != nil || operation.RequireDependency(v.links) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.LearningCourseCompetency
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.LearningCourseCompetency)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	if after == nil && previous == nil {
		return hrerrors.ErrInvalidInput
	}
	link := previous
	if after != nil {
		var ok bool
		link, ok = after.(*entity.LearningCourseCompetency)
		if !ok || link == nil || (previous != nil && (previous.CourseID != link.CourseID || previous.CompetencyID != link.CompetencyID)) {
			return hrerrors.ErrInvalidInput
		}
	}
	course, err := v.courses.FindByIDForUpdate(ctx, link.CourseID)
	if err != nil {
		return err
	}
	if course == nil {
		return hrerrors.ErrReferenceConflict
	}
	competency, err := v.competencies.FindByIDForUpdate(ctx, link.CompetencyID)
	if err != nil {
		return err
	}
	if competency == nil {
		return hrerrors.ErrReferenceConflict
	}
	if after == nil {
		return nil
	}
	if !course.IsActive || !competency.IsActive {
		return hrerrors.ErrReferenceConflict
	}
	if previous != nil {
		return nil
	}
	matches, err := v.links.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "course_id", Values: []any{link.CourseID}},
		{Column: "competency_id", Values: []any{link.CompetencyID}},
	}, 1, 0)
	if err != nil {
		return err
	}
	if len(matches) > 0 {
		return hrerrors.ErrAlreadyExists
	}
	return nil
}
