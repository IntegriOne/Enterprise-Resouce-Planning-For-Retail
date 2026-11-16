package service_test

import (
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/learning"
	f "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/test/fixture"
)

func TestCourseOutcomes(t *testing.T) {
	t.Parallel()
	course := entity.LearningCourse{ID: f.ID(1), CourseCode: "FINANCE", CourseVersion: "v1", CourseName: "Finance", CourseType: "training", DeliveryMode: "online"}
	competency := entity.Competency{ID: f.ID(2), CompetencyCode: "FINANCE", CompetencyVersion: 1, CompetencyName: "Finance", CategoryCode: "technical", ProficiencyScale: entity.JSON(`{}`)}
	outcome := entity.LearningCourseCompetency{CourseID: course.ID, CompetencyID: competency.ID, TargetLevel: 3}
	if err := service.ValidateCourseOutcomes(&course, []entity.LearningCourseCompetency{outcome}, []entity.Competency{competency}); err != nil {
		t.Fatal(err)
	}
	if err := service.ValidateCourseOutcomes(&course, []entity.LearningCourseCompetency{outcome, outcome}, []entity.Competency{competency}); err == nil {
		t.Fatal("duplicate outcome accepted")
	}
}
