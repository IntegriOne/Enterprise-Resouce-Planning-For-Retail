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

func createWorkerQualification(in CreateWorkerQualificationInput) (*entity.WorkerQualification, error) {
	v := &entity.WorkerQualification{}
	v.WorkerID = in.WorkerID
	v.QualificationType = in.QualificationType
	v.QualificationCode = in.QualificationCode
	v.Title = in.Title
	v.IssuerPartyReferenceID = in.IssuerPartyReferenceID
	v.CredentialReferenceID = in.CredentialReferenceID
	v.ObtainedOn = in.ObtainedOn
	v.ExpiresOn = in.ExpiresOn
	if in.VerificationStatus != nil {
		v.VerificationStatus = *in.VerificationStatus
	} else {
		v.VerificationStatus = "unverified"
	}
	v.VerifiedByPrincipalID = in.VerifiedByPrincipalID
	v.VerifiedAt = in.VerifiedAt
	return v, nil
}
func patchWorkerQualification(v *entity.WorkerQualification, in UpdateWorkerQualificationInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.QualificationType.Set {
		if in.QualificationType.Value == nil {
			return hrerrors.Invalid("qualification_type", "must not be null")
		}
		v.QualificationType = *in.QualificationType.Value
	}
	if in.QualificationCode.Set {
		v.QualificationCode = in.QualificationCode.Value
	}
	if in.Title.Set {
		if in.Title.Value == nil {
			return hrerrors.Invalid("title", "must not be null")
		}
		v.Title = *in.Title.Value
	}
	if in.IssuerPartyReferenceID.Set {
		v.IssuerPartyReferenceID = in.IssuerPartyReferenceID.Value
	}
	if in.CredentialReferenceID.Set {
		v.CredentialReferenceID = in.CredentialReferenceID.Value
	}
	if in.ObtainedOn.Set {
		v.ObtainedOn = in.ObtainedOn.Value
	}
	if in.ExpiresOn.Set {
		v.ExpiresOn = in.ExpiresOn.Value
	}
	if in.VerificationStatus.Set {
		if in.VerificationStatus.Value == nil {
			return hrerrors.Invalid("verification_status", "must not be null")
		}
		v.VerificationStatus = *in.VerificationStatus.Value
	}
	if in.VerifiedByPrincipalID.Set {
		v.VerifiedByPrincipalID = in.VerifiedByPrincipalID.Value
	}
	if in.VerifiedAt.Set {
		v.VerifiedAt = in.VerifiedAt.Value
	}
	return nil
}
func specWorkerQualification() operation.Specification[entity.WorkerQualification, string] {
	return operation.Specification[entity.WorkerQualification, string]{
		Domain: "competency", Table: "hr_worker_qualifications",
		Key:         func(v *entity.WorkerQualification) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.WorkerQualification) string { return v.ID },
		Initialize: func(v *entity.WorkerQualification, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.WorkerQualification, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.WorkerQualification) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			if v.IssuerPartyReferenceID != nil {
				a["issuer_party_reference_id"] = fmt.Sprint(*v.IssuerPartyReferenceID)
			}
			if v.CredentialReferenceID != nil {
				a["credential_reference_id"] = fmt.Sprint(*v.CredentialReferenceID)
			}
			a["verification_status"] = fmt.Sprint(v.VerificationStatus)
			if v.VerifiedByPrincipalID != nil {
				a["verified_by_principal_id"] = fmt.Sprint(*v.VerifiedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.WorkerQualification) error {
			if action == "update" && !reflect.DeepEqual(old.VerificationStatus, next.VerificationStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.VerificationStatus) != "unverified" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.VerifiedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.VerifiedAt, next.VerifiedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.VerificationStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createCompetency(in CreateCompetencyInput) (*entity.Competency, error) {
	v := &entity.Competency{}
	v.CompetencyCode = in.CompetencyCode
	v.CompetencyVersion = in.CompetencyVersion
	v.CompetencyName = in.CompetencyName
	v.CategoryCode = in.CategoryCode
	v.ProficiencyScale = in.ProficiencyScale
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	} else {
		v.IsActive = true
	}
	return v, nil
}
func patchCompetency(v *entity.Competency, in UpdateCompetencyInput) error {
	if in.CompetencyCode.Set {
		if in.CompetencyCode.Value == nil {
			return hrerrors.Invalid("competency_code", "must not be null")
		}
		v.CompetencyCode = *in.CompetencyCode.Value
	}
	if in.CompetencyVersion.Set {
		if in.CompetencyVersion.Value == nil {
			return hrerrors.Invalid("competency_version", "must not be null")
		}
		v.CompetencyVersion = *in.CompetencyVersion.Value
	}
	if in.CompetencyName.Set {
		if in.CompetencyName.Value == nil {
			return hrerrors.Invalid("competency_name", "must not be null")
		}
		v.CompetencyName = *in.CompetencyName.Value
	}
	if in.CategoryCode.Set {
		if in.CategoryCode.Value == nil {
			return hrerrors.Invalid("category_code", "must not be null")
		}
		v.CategoryCode = *in.CategoryCode.Value
	}
	if in.ProficiencyScale.Set {
		if in.ProficiencyScale.Value == nil {
			v.ProficiencyScale = nil
		} else {
			v.ProficiencyScale = *in.ProficiencyScale.Value
		}
	}
	if in.IsActive.Set {
		if in.IsActive.Value == nil {
			return hrerrors.Invalid("is_active", "must not be null")
		}
		v.IsActive = *in.IsActive.Value
	}
	return nil
}
func specCompetency() operation.Specification[entity.Competency, string] {
	return operation.Specification[entity.Competency, string]{
		Domain: "competency", Table: "hr_competencies",
		Key:         func(v *entity.Competency) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.Competency) string { return v.ID },
		Initialize: func(v *entity.Competency, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.Competency, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.Competency) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			return a
		},
		Guard: func(action string, old, next *entity.Competency) error {
			return nil
		},
	}
}
func createJobCompetency(in CreateJobCompetencyInput) (*entity.JobCompetency, error) {
	v := &entity.JobCompetency{}
	v.JobID = in.JobID
	v.CompetencyID = in.CompetencyID
	v.RequiredLevel = in.RequiredLevel
	if in.IsMandatory != nil {
		v.IsMandatory = *in.IsMandatory
	} else {
		v.IsMandatory = true
	}
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	return v, nil
}
func patchJobCompetency(v *entity.JobCompetency, in UpdateJobCompetencyInput) error {
	if in.JobID.Set {
		if in.JobID.Value == nil {
			return hrerrors.Invalid("job_id", "must not be null")
		}
		v.JobID = *in.JobID.Value
	}
	if in.CompetencyID.Set {
		if in.CompetencyID.Value == nil {
			return hrerrors.Invalid("competency_id", "must not be null")
		}
		v.CompetencyID = *in.CompetencyID.Value
	}
	if in.RequiredLevel.Set {
		if in.RequiredLevel.Value == nil {
			return hrerrors.Invalid("required_level", "must not be null")
		}
		v.RequiredLevel = *in.RequiredLevel.Value
	}
	if in.IsMandatory.Set {
		if in.IsMandatory.Value == nil {
			return hrerrors.Invalid("is_mandatory", "must not be null")
		}
		v.IsMandatory = *in.IsMandatory.Value
	}
	if in.EffectiveFrom.Set {
		if in.EffectiveFrom.Value == nil {
			return hrerrors.Invalid("effective_from", "must not be null")
		}
		v.EffectiveFrom = *in.EffectiveFrom.Value
	}
	if in.EffectiveTo.Set {
		v.EffectiveTo = in.EffectiveTo.Value
	}
	return nil
}
func specJobCompetency() operation.Specification[entity.JobCompetency, string] {
	return operation.Specification[entity.JobCompetency, string]{
		Domain: "competency", Table: "hr_job_competencies",
		Key:         func(v *entity.JobCompetency) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.JobCompetency) string { return v.ID },
		Initialize: func(v *entity.JobCompetency, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.JobCompetency, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.JobCompetency) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["job_id"] = fmt.Sprint(v.JobID)
			a["competency_id"] = fmt.Sprint(v.CompetencyID)
			return a
		},
		Guard: func(action string, old, next *entity.JobCompetency) error {
			return nil
		},
	}
}
func createWorkerCompetencyAssessment(in CreateWorkerCompetencyAssessmentInput) (*entity.WorkerCompetencyAssessment, error) {
	v := &entity.WorkerCompetencyAssessment{}
	v.WorkerID = in.WorkerID
	v.CompetencyID = in.CompetencyID
	v.AssessedLevel = in.AssessedLevel
	v.AssessmentMethod = in.AssessmentMethod
	v.AssessedAt = in.AssessedAt
	v.ExpiresOn = in.ExpiresOn
	v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID
	v.EvidenceDigest = in.EvidenceDigest
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchWorkerCompetencyAssessment(v *entity.WorkerCompetencyAssessment, in UpdateWorkerCompetencyAssessmentInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.CompetencyID.Set {
		if in.CompetencyID.Value == nil {
			return hrerrors.Invalid("competency_id", "must not be null")
		}
		v.CompetencyID = *in.CompetencyID.Value
	}
	if in.AssessedLevel.Set {
		if in.AssessedLevel.Value == nil {
			return hrerrors.Invalid("assessed_level", "must not be null")
		}
		v.AssessedLevel = *in.AssessedLevel.Value
	}
	if in.AssessmentMethod.Set {
		if in.AssessmentMethod.Value == nil {
			return hrerrors.Invalid("assessment_method", "must not be null")
		}
		v.AssessmentMethod = *in.AssessmentMethod.Value
	}
	if in.AssessedAt.Set {
		if in.AssessedAt.Value == nil {
			return hrerrors.Invalid("assessed_at", "must not be null")
		}
		v.AssessedAt = *in.AssessedAt.Value
	}
	if in.ExpiresOn.Set {
		v.ExpiresOn = in.ExpiresOn.Value
	}
	if in.EvidenceDocumentVersionReferenceID.Set {
		v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID.Value
	}
	if in.EvidenceDigest.Set {
		if in.EvidenceDigest.Value == nil {
			v.EvidenceDigest = nil
		} else {
			v.EvidenceDigest = *in.EvidenceDigest.Value
		}
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specWorkerCompetencyAssessment() operation.Specification[entity.WorkerCompetencyAssessment, string] {
	return operation.Specification[entity.WorkerCompetencyAssessment, string]{
		Domain: "competency", Table: "hr_worker_competency_assessments",
		Key:         func(v *entity.WorkerCompetencyAssessment) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.WorkerCompetencyAssessment) string { return v.ID },
		Initialize: func(v *entity.WorkerCompetencyAssessment, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.AssessedByPrincipalID = actor.PrincipalID
		},
		Touch: func(v *entity.WorkerCompetencyAssessment, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.WorkerCompetencyAssessment) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["competency_id"] = fmt.Sprint(v.CompetencyID)
			a["assessed_by_principal_id"] = fmt.Sprint(v.AssessedByPrincipalID)
			if v.EvidenceDocumentVersionReferenceID != nil {
				a["evidence_document_version_reference_id"] = fmt.Sprint(*v.EvidenceDocumentVersionReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.WorkerCompetencyAssessment) error {
			return nil
		},
	}
}
