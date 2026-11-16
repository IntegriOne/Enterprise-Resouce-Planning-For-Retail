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

func createEmployeeRelationsCase(in CreateEmployeeRelationsCaseInput) (*entity.EmployeeRelationsCase, error) {
	v := &entity.EmployeeRelationsCase{}
	v.CaseNumber = in.CaseNumber
	v.CaseType = in.CaseType
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "open"
	}
	v.SubjectWorkerID = in.SubjectWorkerID
	v.ReportingPartyReferenceID = in.ReportingPartyReferenceID
	v.AssignedInvestigatorPrincipalID = in.AssignedInvestigatorPrincipalID
	v.DueAt = in.DueAt
	v.ResolvedAt = in.ResolvedAt
	v.OutcomeCode = in.OutcomeCode
	v.RestrictedSummaryReferenceID = in.RestrictedSummaryReferenceID
	if in.ConfidentialityClass != nil {
		v.ConfidentialityClass = *in.ConfidentialityClass
	} else {
		v.ConfidentialityClass = "highly_restricted"
	}
	v.LegalHoldReferenceID = in.LegalHoldReferenceID
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchEmployeeRelationsCase(v *entity.EmployeeRelationsCase, in UpdateEmployeeRelationsCaseInput) error {
	if in.CaseNumber.Set {
		if in.CaseNumber.Value == nil {
			return hrerrors.Invalid("case_number", "must not be null")
		}
		v.CaseNumber = *in.CaseNumber.Value
	}
	if in.CaseType.Set {
		if in.CaseType.Value == nil {
			return hrerrors.Invalid("case_type", "must not be null")
		}
		v.CaseType = *in.CaseType.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.SubjectWorkerID.Set {
		v.SubjectWorkerID = in.SubjectWorkerID.Value
	}
	if in.ReportingPartyReferenceID.Set {
		v.ReportingPartyReferenceID = in.ReportingPartyReferenceID.Value
	}
	if in.AssignedInvestigatorPrincipalID.Set {
		v.AssignedInvestigatorPrincipalID = in.AssignedInvestigatorPrincipalID.Value
	}
	if in.DueAt.Set {
		v.DueAt = in.DueAt.Value
	}
	if in.ResolvedAt.Set {
		v.ResolvedAt = in.ResolvedAt.Value
	}
	if in.OutcomeCode.Set {
		v.OutcomeCode = in.OutcomeCode.Value
	}
	if in.RestrictedSummaryReferenceID.Set {
		if in.RestrictedSummaryReferenceID.Value == nil {
			return hrerrors.Invalid("restricted_summary_reference_id", "must not be null")
		}
		v.RestrictedSummaryReferenceID = *in.RestrictedSummaryReferenceID.Value
	}
	if in.ConfidentialityClass.Set {
		if in.ConfidentialityClass.Value == nil {
			return hrerrors.Invalid("confidentiality_class", "must not be null")
		}
		v.ConfidentialityClass = *in.ConfidentialityClass.Value
	}
	if in.LegalHoldReferenceID.Set {
		v.LegalHoldReferenceID = in.LegalHoldReferenceID.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specEmployeeRelationsCase() operation.Specification[entity.EmployeeRelationsCase, string] {
	return operation.Specification[entity.EmployeeRelationsCase, string]{
		Domain: "employeerelations", Table: "hr_employee_relations_cases",
		Key:         func(v *entity.EmployeeRelationsCase) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.EmployeeRelationsCase) string { return v.ID },
		Initialize: func(v *entity.EmployeeRelationsCase, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.OpenedAt = now
		},
		Touch: func(v *entity.EmployeeRelationsCase, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.EmployeeRelationsCase) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["status"] = fmt.Sprint(v.Status)
			if v.SubjectWorkerID != nil {
				a["subject_worker_id"] = fmt.Sprint(*v.SubjectWorkerID)
			}
			if v.ReportingPartyReferenceID != nil {
				a["reporting_party_reference_id"] = fmt.Sprint(*v.ReportingPartyReferenceID)
			}
			if v.AssignedInvestigatorPrincipalID != nil {
				a["assigned_investigator_principal_id"] = fmt.Sprint(*v.AssignedInvestigatorPrincipalID)
			}
			a["restricted_summary_reference_id"] = fmt.Sprint(v.RestrictedSummaryReferenceID)
			if v.LegalHoldReferenceID != nil {
				a["legal_hold_reference_id"] = fmt.Sprint(*v.LegalHoldReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.EmployeeRelationsCase) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "open" {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.Status) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createEmployeeRelationsEvent(in CreateEmployeeRelationsEventInput) (*entity.EmployeeRelationsEvent, error) {
	v := &entity.EmployeeRelationsEvent{}
	v.CaseID = in.CaseID
	v.EventType = in.EventType
	v.EventDigest = in.EventDigest
	v.EventReferenceID = in.EventReferenceID
	return v, nil
}
func patchEmployeeRelationsEvent(v *entity.EmployeeRelationsEvent, in UpdateEmployeeRelationsEventInput) error {
	if in.CaseID.Set {
		if in.CaseID.Value == nil {
			return hrerrors.Invalid("case_id", "must not be null")
		}
		v.CaseID = *in.CaseID.Value
	}
	if in.EventType.Set {
		if in.EventType.Value == nil {
			return hrerrors.Invalid("event_type", "must not be null")
		}
		v.EventType = *in.EventType.Value
	}
	if in.EventDigest.Set {
		if in.EventDigest.Value == nil {
			v.EventDigest = nil
		} else {
			v.EventDigest = *in.EventDigest.Value
		}
	}
	if in.EventReferenceID.Set {
		v.EventReferenceID = in.EventReferenceID.Value
	}
	return nil
}
func specEmployeeRelationsEvent() operation.Specification[entity.EmployeeRelationsEvent, int64] {
	return operation.Specification[entity.EmployeeRelationsEvent, int64]{
		Domain: "employeerelations", Table: "hr_employee_relations_events",
		Key:       func(v *entity.EmployeeRelationsEvent) int64 { return v.ID },
		KeyString: func(key int64) string { return fmt.Sprint(key) },
		ValidateKey: func(key int64) error {
			return func() error {
				if key <= 0 {
					return hrerrors.ErrInvalidInput
				}
				return nil
			}()
		},
		AggregateID: func(v *entity.EmployeeRelationsEvent) string { return v.CaseID },
		Initialize: func(v *entity.EmployeeRelationsEvent, actor operation.Actor, now time.Time, id string) {
			v.OccurredAt = now
			v.ActorPrincipalID = &actor.PrincipalID
		},
		Touch: func(v *entity.EmployeeRelationsEvent, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.EmployeeRelationsEvent) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["case_id"] = fmt.Sprint(v.CaseID)
			if v.ActorPrincipalID != nil {
				a["actor_principal_id"] = fmt.Sprint(*v.ActorPrincipalID)
			}
			if v.EventReferenceID != nil {
				a["event_reference_id"] = fmt.Sprint(*v.EventReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.EmployeeRelationsEvent) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createCollectiveAgreement(in CreateCollectiveAgreementInput) (*entity.CollectiveAgreement, error) {
	v := &entity.CollectiveAgreement{}
	v.AgreementCode = in.AgreementCode
	v.AgreementVersion = in.AgreementVersion
	v.LegalEntityReferenceID = in.LegalEntityReferenceID
	v.RepresentativePartyReferenceID = in.RepresentativePartyReferenceID
	v.JurisdictionCode = in.JurisdictionCode
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	v.SignedAt = in.SignedAt
	v.DocumentVersionReferenceID = in.DocumentVersionReferenceID
	v.ContentDigest = in.ContentDigest
	return v, nil
}
func patchCollectiveAgreement(v *entity.CollectiveAgreement, in UpdateCollectiveAgreementInput) error {
	if in.AgreementCode.Set {
		if in.AgreementCode.Value == nil {
			return hrerrors.Invalid("agreement_code", "must not be null")
		}
		v.AgreementCode = *in.AgreementCode.Value
	}
	if in.AgreementVersion.Set {
		if in.AgreementVersion.Value == nil {
			return hrerrors.Invalid("agreement_version", "must not be null")
		}
		v.AgreementVersion = *in.AgreementVersion.Value
	}
	if in.LegalEntityReferenceID.Set {
		if in.LegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("legal_entity_reference_id", "must not be null")
		}
		v.LegalEntityReferenceID = *in.LegalEntityReferenceID.Value
	}
	if in.RepresentativePartyReferenceID.Set {
		if in.RepresentativePartyReferenceID.Value == nil {
			return hrerrors.Invalid("representative_party_reference_id", "must not be null")
		}
		v.RepresentativePartyReferenceID = *in.RepresentativePartyReferenceID.Value
	}
	if in.JurisdictionCode.Set {
		if in.JurisdictionCode.Value == nil {
			return hrerrors.Invalid("jurisdiction_code", "must not be null")
		}
		v.JurisdictionCode = *in.JurisdictionCode.Value
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
	if in.SignedAt.Set {
		if in.SignedAt.Value == nil {
			return hrerrors.Invalid("signed_at", "must not be null")
		}
		v.SignedAt = *in.SignedAt.Value
	}
	if in.DocumentVersionReferenceID.Set {
		if in.DocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("document_version_reference_id", "must not be null")
		}
		v.DocumentVersionReferenceID = *in.DocumentVersionReferenceID.Value
	}
	if in.ContentDigest.Set {
		if in.ContentDigest.Value == nil {
			v.ContentDigest = nil
		} else {
			v.ContentDigest = *in.ContentDigest.Value
		}
	}
	return nil
}
func specCollectiveAgreement() operation.Specification[entity.CollectiveAgreement, string] {
	return operation.Specification[entity.CollectiveAgreement, string]{
		Domain: "employeerelations", Table: "hr_collective_agreements",
		Key:         func(v *entity.CollectiveAgreement) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.CollectiveAgreement) string { return v.ID },
		Initialize: func(v *entity.CollectiveAgreement, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.CollectiveAgreement, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.CollectiveAgreement) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["legal_entity_reference_id"] = fmt.Sprint(v.LegalEntityReferenceID)
			a["representative_party_reference_id"] = fmt.Sprint(v.RepresentativePartyReferenceID)
			a["document_version_reference_id"] = fmt.Sprint(v.DocumentVersionReferenceID)
			return a
		},
		Guard: func(action string, old, next *entity.CollectiveAgreement) error {
			return nil
		},
	}
}
