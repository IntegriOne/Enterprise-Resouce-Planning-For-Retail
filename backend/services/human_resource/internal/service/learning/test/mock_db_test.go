// Kerangka tes awal dibuat dengan alat bantu; tes ini dipelihara langsung.
// Fixture dan dependency pada file ini hanya untuk pengujian service.
package service_test

import (
	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	service "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/learning"
	mockdb "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/testsupport"
	"testing"
)

func setup(tb testing.TB) (*mockdb.Harness, *service.Service, *mockdb.Store[entity.LearningCourse, string]) {
	tb.Helper()
	h := mockdb.New()
	pLearningCourse := mockdb.NewStore[entity.LearningCourse, string](h, func(v *entity.LearningCourse) string { return v.ID })
	pLearningEnrollment := mockdb.NewStore[entity.LearningEnrollment, string](h, func(v *entity.LearningEnrollment) string { return v.ID })
	pLearningCourseCompetency := mockdb.NewStore[entity.LearningCourseCompetency, entity.LearningCourseCompetencyKey](h, func(v *entity.LearningCourseCompetency) entity.LearningCourseCompetencyKey {
		return entity.LearningCourseCompetencyKey{CourseID: v.CourseID, CompetencyID: v.CompetencyID}
	})
	svc, err := service.NewService(service.Dependencies{Runtime: h.Runtime(), LearningCourse: pLearningCourse, LearningEnrollment: pLearningEnrollment, LearningCourseCompetency: pLearningCourseCompetency})
	if err != nil {
		tb.Fatal(err)
	}
	return h, svc, pLearningCourse
}
func input() service.CreateLearningCourseInput {
	return service.CreateLearningCourseInput{CourseCode: "TEST", CourseVersion: "TEST", CourseName: "TEST", CourseType: "TEST", DeliveryMode: "TEST"}
}
