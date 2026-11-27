// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"fmt"
	"reflect"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func createLearningCourse(in CreateLearningCourseInput) (*entity.LearningCourse, error) {
	v := &entity.LearningCourse{}
	v.CourseCode = in.CourseCode
	v.CourseVersion = in.CourseVersion
	v.CourseName = in.CourseName
	v.CourseType = in.CourseType
	v.ProviderPartyReferenceID = in.ProviderPartyReferenceID
	v.DeliveryMode = in.DeliveryMode
	v.DurationHours = in.DurationHours
	if in.CredentialAwarded != nil {
		v.CredentialAwarded = *in.CredentialAwarded
	} else {
		v.CredentialAwarded = false
	}
	v.ExpiryMonths = in.ExpiryMonths
	v.ValidFrom = in.ValidFrom
	v.ValidTo = in.ValidTo
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	} else {
		v.IsActive = true
	}
	return v, nil
}
func patchLearningCourse(v *entity.LearningCourse, in UpdateLearningCourseInput) error {
	if in.CourseCode.Set {
		if in.CourseCode.Value == nil {
			return hrerrors.Invalid("course_code", "must not be null")
		}
		v.CourseCode = *in.CourseCode.Value
	}
	if in.CourseVersion.Set {
		if in.CourseVersion.Value == nil {
			return hrerrors.Invalid("course_version", "must not be null")
		}
		v.CourseVersion = *in.CourseVersion.Value
	}
	if in.CourseName.Set {
		if in.CourseName.Value == nil {
			return hrerrors.Invalid("course_name", "must not be null")
		}
		v.CourseName = *in.CourseName.Value
	}
	if in.CourseType.Set {
		if in.CourseType.Value == nil {
			return hrerrors.Invalid("course_type", "must not be null")
		}
		v.CourseType = *in.CourseType.Value
	}
	if in.ProviderPartyReferenceID.Set {
		v.ProviderPartyReferenceID = in.ProviderPartyReferenceID.Value
	}
	if in.DeliveryMode.Set {
		if in.DeliveryMode.Value == nil {
			return hrerrors.Invalid("delivery_mode", "must not be null")
		}
		v.DeliveryMode = *in.DeliveryMode.Value
	}
	if in.DurationHours.Set {
		v.DurationHours = in.DurationHours.Value
	}
	if in.CredentialAwarded.Set {
		if in.CredentialAwarded.Value == nil {
			return hrerrors.Invalid("credential_awarded", "must not be null")
		}
		v.CredentialAwarded = *in.CredentialAwarded.Value
	}
	if in.ExpiryMonths.Set {
		v.ExpiryMonths = in.ExpiryMonths.Value
	}
	if in.ValidFrom.Set {
		v.ValidFrom = in.ValidFrom.Value
	}
	if in.ValidTo.Set {
		v.ValidTo = in.ValidTo.Value
	}
	if in.IsActive.Set {
		if in.IsActive.Value == nil {
			return hrerrors.Invalid("is_active", "must not be null")
		}
		v.IsActive = *in.IsActive.Value
	}
	return nil
}
func specLearningCourse() operation.Specification[entity.LearningCourse, string] {
	return operation.Specification[entity.LearningCourse, string]{
		Domain: "learning", Table: "hr_learning_courses",
		Key:         func(v *entity.LearningCourse) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.LearningCourse) string { return v.ID },
		Initialize: func(v *entity.LearningCourse, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.LearningCourse, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.LearningCourse) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			if v.ProviderPartyReferenceID != nil {
				a["provider_party_reference_id"] = fmt.Sprint(*v.ProviderPartyReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.LearningCourse) error {
			return nil
		},
	}
}
func createLearningEnrollment(in CreateLearningEnrollmentInput) (*entity.LearningEnrollment, error) {
	v := &entity.LearningEnrollment{}
	v.WorkerID = in.WorkerID
	v.CourseID = in.CourseID
	if in.EnrollmentStatus != nil {
		v.EnrollmentStatus = *in.EnrollmentStatus
	} else {
		v.EnrollmentStatus = "planned"
	}
	v.AssignedByPrincipalID = in.AssignedByPrincipalID
	v.DueAt = in.DueAt
	v.StartedAt = in.StartedAt
	v.CompletedAt = in.CompletedAt
	v.Score = in.Score
	v.CredentialReferenceID = in.CredentialReferenceID
	v.FundingReferenceID = in.FundingReferenceID
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchLearningEnrollment(v *entity.LearningEnrollment, in UpdateLearningEnrollmentInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.CourseID.Set {
		if in.CourseID.Value == nil {
			return hrerrors.Invalid("course_id", "must not be null")
		}
		v.CourseID = *in.CourseID.Value
	}
	if in.EnrollmentStatus.Set {
		if in.EnrollmentStatus.Value == nil {
			return hrerrors.Invalid("enrollment_status", "must not be null")
		}
		v.EnrollmentStatus = *in.EnrollmentStatus.Value
	}
	if in.AssignedByPrincipalID.Set {
		v.AssignedByPrincipalID = in.AssignedByPrincipalID.Value
	}
	if in.DueAt.Set {
		v.DueAt = in.DueAt.Value
	}
	if in.StartedAt.Set {
		v.StartedAt = in.StartedAt.Value
	}
	if in.CompletedAt.Set {
		v.CompletedAt = in.CompletedAt.Value
	}
	if in.Score.Set {
		v.Score = in.Score.Value
	}
	if in.CredentialReferenceID.Set {
		v.CredentialReferenceID = in.CredentialReferenceID.Value
	}
	if in.FundingReferenceID.Set {
		v.FundingReferenceID = in.FundingReferenceID.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specLearningEnrollment() operation.Specification[entity.LearningEnrollment, string] {
	return operation.Specification[entity.LearningEnrollment, string]{
		Domain: "learning", Table: "hr_learning_enrollments",
		Key:         func(v *entity.LearningEnrollment) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.LearningEnrollment) string { return v.ID },
		Initialize: func(v *entity.LearningEnrollment, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.AssignedAt = now
			v.AssignedByPrincipalID = &actor.PrincipalID
		},
		Touch: func(v *entity.LearningEnrollment, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.LearningEnrollment) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["course_id"] = fmt.Sprint(v.CourseID)
			a["enrollment_status"] = fmt.Sprint(v.EnrollmentStatus)
			if v.AssignedByPrincipalID != nil {
				a["assigned_by_principal_id"] = fmt.Sprint(*v.AssignedByPrincipalID)
			}
			if v.CredentialReferenceID != nil {
				a["credential_reference_id"] = fmt.Sprint(*v.CredentialReferenceID)
			}
			if v.FundingReferenceID != nil {
				a["funding_reference_id"] = fmt.Sprint(*v.FundingReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.LearningEnrollment) error {
			if action == "create" && (next.StartedAt != nil || next.CompletedAt != nil || next.Score != nil || next.CredentialReferenceID != nil) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.EnrollmentStatus, next.EnrollmentStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && (old.WorkerID != next.WorkerID || old.CourseID != next.CourseID ||
				old.IdempotencyKey != next.IdempotencyKey || !reflect.DeepEqual(old.AssignedByPrincipalID, next.AssignedByPrincipalID) ||
				!reflect.DeepEqual(old.StartedAt, next.StartedAt) || !reflect.DeepEqual(old.Score, next.Score) ||
				!reflect.DeepEqual(old.CredentialReferenceID, next.CredentialReferenceID)) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.EnrollmentStatus) != "planned" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.CompletedAt, next.CompletedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch old.EnrollmentStatus {
				case entity.LearningStatusCompleted, entity.LearningStatusFailed, entity.LearningStatusWithdrawn, entity.LearningStatusExpired:
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createLearningCourseCompetency(in CreateLearningCourseCompetencyInput) (*entity.LearningCourseCompetency, error) {
	v := &entity.LearningCourseCompetency{}
	v.CourseID = in.CourseID
	v.CompetencyID = in.CompetencyID
	v.TargetLevel = in.TargetLevel
	return v, nil
}
func patchLearningCourseCompetency(v *entity.LearningCourseCompetency, in UpdateLearningCourseCompetencyInput) error {
	if in.TargetLevel.Set {
		if in.TargetLevel.Value == nil {
			return hrerrors.Invalid("target_level", "must not be null")
		}
		v.TargetLevel = *in.TargetLevel.Value
	}
	return nil
}
func specLearningCourseCompetency() operation.Specification[entity.LearningCourseCompetency, entity.LearningCourseCompetencyKey] {
	return operation.Specification[entity.LearningCourseCompetency, entity.LearningCourseCompetencyKey]{
		Domain: "learning", Table: "hr_learning_course_competencies",
		Key: func(v *entity.LearningCourseCompetency) entity.LearningCourseCompetencyKey {
			return entity.LearningCourseCompetencyKey{CourseID: v.CourseID, CompetencyID: v.CompetencyID}
		},
		KeyString: func(key entity.LearningCourseCompetencyKey) string { return fmt.Sprint(key) },
		ValidateKey: func(key entity.LearningCourseCompetencyKey) error {
			return func() error {
				if err := entity.ValidateUUID("course_id", key.CourseID); err != nil {
					return err
				}
				if err := entity.ValidateUUID("competency_id", key.CompetencyID); err != nil {
					return err
				}
				return nil
			}()
		},
		AggregateID: func(v *entity.LearningCourseCompetency) string { return v.CourseID },
		Initialize: func(v *entity.LearningCourseCompetency, actor operation.Actor, now time.Time, id string) {
		},
		Touch: func(v *entity.LearningCourseCompetency, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.LearningCourseCompetency) map[string]string {
			a := map[string]string{}
			a["course_id"] = fmt.Sprint(v.CourseID)
			a["competency_id"] = fmt.Sprint(v.CompetencyID)
			return a
		},
		Guard: func(action string, old, next *entity.LearningCourseCompetency) error {
			return nil
		},
	}
}
