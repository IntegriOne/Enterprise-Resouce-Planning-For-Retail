// Package service memeriksa outcome kursus tanpa memberikan skor kompetensi otomatis.
package service

import (
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// ValidateCourseOutcomes memeriksa pemetaan kursus ke katalog kompetensi berversi.
func ValidateCourseOutcomes(course *entity.LearningCourse, outcomes []entity.LearningCourseCompetency, competencies []entity.Competency) error {
	if err := course.Validate(); err != nil {
		return err
	}
	catalog := map[string]bool{}
	for i := range competencies {
		if err := competencies[i].Validate(); err != nil {
			return err
		}
		key := strings.ToLower(competencies[i].ID)
		if catalog[key] {
			return hrerrors.Invalid("competencies", "duplicate competency")
		}
		catalog[key] = true
	}
	seen := map[string]bool{}
	for i := range outcomes {
		item := &outcomes[i]
		if err := item.Validate(); err != nil {
			return err
		}
		key := strings.ToLower(item.CompetencyID)
		if !strings.EqualFold(item.CourseID, course.ID) || !catalog[key] || seen[key] {
			return hrerrors.Invalid("outcomes", "course outcomes must reference unique catalog competencies")
		}
		seen[key] = true
	}
	return nil
}
