// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func createWorker(in CreateWorkerInput) (*entity.Worker, error) {
	if in.IAMPrincipalReferenceID != nil {
		return nil, hrerrors.ErrWorkflowRequired
	}
	v := &entity.Worker{}
	v.WorkerNumber = in.WorkerNumber
	v.PersonPartyReferenceID = in.PersonPartyReferenceID
	v.WorkerCategory = in.WorkerCategory
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "preboarding"
	}
	v.HireDate = in.HireDate
	v.ServiceStartDate = in.ServiceStartDate
	v.TerminationDate = in.TerminationDate
	v.TerminationReasonCode = in.TerminationReasonCode
	v.PrimaryLegalEntityReferenceID = in.PrimaryLegalEntityReferenceID
	v.PrimaryAssignmentID = in.PrimaryAssignmentID
	return v, nil
}
func patchWorker(v *entity.Worker, in UpdateWorkerInput) error {
	if in.IAMPrincipalReferenceID.Set {
		return hrerrors.ErrWorkflowRequired
	}
	if in.WorkerNumber.Set {
		if in.WorkerNumber.Value == nil {
			return hrerrors.Invalid("worker_number", "must not be null")
		}
		v.WorkerNumber = *in.WorkerNumber.Value
	}
	if in.PersonPartyReferenceID.Set {
		if in.PersonPartyReferenceID.Value == nil {
			return hrerrors.Invalid("person_party_reference_id", "must not be null")
		}
		v.PersonPartyReferenceID = *in.PersonPartyReferenceID.Value
	}
	if in.WorkerCategory.Set {
		if in.WorkerCategory.Value == nil {
			return hrerrors.Invalid("worker_category", "must not be null")
		}
		v.WorkerCategory = *in.WorkerCategory.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.HireDate.Set {
		v.HireDate = in.HireDate.Value
	}
	if in.ServiceStartDate.Set {
		v.ServiceStartDate = in.ServiceStartDate.Value
	}
	if in.TerminationDate.Set {
		v.TerminationDate = in.TerminationDate.Value
	}
	if in.TerminationReasonCode.Set {
		v.TerminationReasonCode = in.TerminationReasonCode.Value
	}
	if in.PrimaryLegalEntityReferenceID.Set {
		if in.PrimaryLegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("primary_legal_entity_reference_id", "must not be null")
		}
		v.PrimaryLegalEntityReferenceID = *in.PrimaryLegalEntityReferenceID.Value
	}
	if in.PrimaryAssignmentID.Set {
		v.PrimaryAssignmentID = in.PrimaryAssignmentID.Value
	}
	return nil
}
func specWorker() operation.Specification[entity.Worker, string] {
	return operation.Specification[entity.Worker, string]{
		Domain: "personnel", Table: "hr_workers",
		Key:         func(v *entity.Worker) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		KeyFilters:  func(key string) []domainrepo.Filter { return []domainrepo.Filter{{Column: "id", Values: []any{key}}} },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.Worker) string { return v.ID },
		Initialize: func(v *entity.Worker, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedAt = now
			v.UpdatedAt = now
			v.CreatedByPrincipalID = &actor.PrincipalID
		},
		Touch: func(v *entity.Worker, actor operation.Actor, now time.Time) { v.UpdatedAt = now },
		Attributes: func(v *entity.Worker) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["person_party_reference_id"] = fmt.Sprint(v.PersonPartyReferenceID)
			a["worker_category"] = fmt.Sprint(v.WorkerCategory)
			a["status"] = fmt.Sprint(v.Status)
			a["primary_legal_entity_reference_id"] = fmt.Sprint(v.PrimaryLegalEntityReferenceID)
			if v.PrimaryAssignmentID != nil {
				a["primary_assignment_id"] = fmt.Sprint(*v.PrimaryAssignmentID)
			}
			if v.IAMPrincipalReferenceID != nil {
				a["iam_principal_reference_id"] = fmt.Sprint(*v.IAMPrincipalReferenceID)
			}
			if v.CreatedByPrincipalID != nil {
				a["created_by_principal_id"] = fmt.Sprint(*v.CreatedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.Worker) error {
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "preboarding" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.PrimaryAssignmentID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.PrimaryAssignmentID, next.PrimaryAssignmentID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.IAMPrincipalReferenceID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.IAMPrincipalReferenceID, next.IAMPrincipalReferenceID) {
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
func createEmploymentContract(in CreateEmploymentContractInput) (*entity.EmploymentContract, error) {
	v := &entity.EmploymentContract{}
	v.WorkerID = in.WorkerID
	v.ContractNumber = in.ContractNumber
	v.ContractType = in.ContractType
	if in.ContractStatus != nil {
		v.ContractStatus = *in.ContractStatus
	} else {
		v.ContractStatus = "draft"
	}
	v.EmployingLegalEntityReferenceID = in.EmployingLegalEntityReferenceID
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	v.SignedAt = in.SignedAt
	v.SignedDocumentReferenceID = in.SignedDocumentReferenceID
	v.SignedDocumentVersionReferenceID = in.SignedDocumentVersionReferenceID
	v.SignedContentDigest = in.SignedContentDigest
	v.ProbationEndDate = in.ProbationEndDate
	v.NoticePeriodDays = in.NoticePeriodDays
	v.WorkCountryCode = in.WorkCountryCode
	v.GoverningJurisdictionCode = in.GoverningJurisdictionCode
	v.CollectiveAgreementID = in.CollectiveAgreementID
	return v, nil
}
func patchEmploymentContract(v *entity.EmploymentContract, in UpdateEmploymentContractInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.ContractNumber.Set {
		if in.ContractNumber.Value == nil {
			return hrerrors.Invalid("contract_number", "must not be null")
		}
		v.ContractNumber = *in.ContractNumber.Value
	}
	if in.ContractType.Set {
		if in.ContractType.Value == nil {
			return hrerrors.Invalid("contract_type", "must not be null")
		}
		v.ContractType = *in.ContractType.Value
	}
	if in.ContractStatus.Set {
		if in.ContractStatus.Value == nil {
			return hrerrors.Invalid("contract_status", "must not be null")
		}
		v.ContractStatus = *in.ContractStatus.Value
	}
	if in.EmployingLegalEntityReferenceID.Set {
		if in.EmployingLegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("employing_legal_entity_reference_id", "must not be null")
		}
		v.EmployingLegalEntityReferenceID = *in.EmployingLegalEntityReferenceID.Value
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
		v.SignedAt = in.SignedAt.Value
	}
	if in.SignedDocumentReferenceID.Set {
		v.SignedDocumentReferenceID = in.SignedDocumentReferenceID.Value
	}
	if in.SignedDocumentVersionReferenceID.Set {
		v.SignedDocumentVersionReferenceID = in.SignedDocumentVersionReferenceID.Value
	}
	if in.SignedContentDigest.Set {
		if in.SignedContentDigest.Value == nil {
			v.SignedContentDigest = nil
		} else {
			v.SignedContentDigest = *in.SignedContentDigest.Value
		}
	}
	if in.ProbationEndDate.Set {
		v.ProbationEndDate = in.ProbationEndDate.Value
	}
	if in.NoticePeriodDays.Set {
		v.NoticePeriodDays = in.NoticePeriodDays.Value
	}
	if in.WorkCountryCode.Set {
		if in.WorkCountryCode.Value == nil {
			return hrerrors.Invalid("work_country_code", "must not be null")
		}
		v.WorkCountryCode = *in.WorkCountryCode.Value
	}
	if in.GoverningJurisdictionCode.Set {
		if in.GoverningJurisdictionCode.Value == nil {
			return hrerrors.Invalid("governing_jurisdiction_code", "must not be null")
		}
		v.GoverningJurisdictionCode = *in.GoverningJurisdictionCode.Value
	}
	if in.CollectiveAgreementID.Set {
		v.CollectiveAgreementID = in.CollectiveAgreementID.Value
	}
	return nil
}
func specEmploymentContract() operation.Specification[entity.EmploymentContract, string] {
	return operation.Specification[entity.EmploymentContract, string]{
		Domain: "personnel", Table: "hr_employment_contracts",
		Key:         func(v *entity.EmploymentContract) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.EmploymentContract) string { return v.ID },
		Initialize: func(v *entity.EmploymentContract, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedAt = now
		},
		Touch: func(v *entity.EmploymentContract, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.EmploymentContract) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["contract_status"] = fmt.Sprint(v.ContractStatus)
			a["employing_legal_entity_reference_id"] = fmt.Sprint(v.EmployingLegalEntityReferenceID)
			if v.SignedDocumentReferenceID != nil {
				a["signed_document_reference_id"] = fmt.Sprint(*v.SignedDocumentReferenceID)
			}
			if v.SignedDocumentVersionReferenceID != nil {
				a["signed_document_version_reference_id"] = fmt.Sprint(*v.SignedDocumentVersionReferenceID)
			}
			if v.CollectiveAgreementID != nil {
				a["collective_agreement_id"] = fmt.Sprint(*v.CollectiveAgreementID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.EmploymentContract) error {
			if action == "update" && !reflect.DeepEqual(old.ContractStatus, next.ContractStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.ContractStatus) != "draft" {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				switch string(old.ContractStatus) {
				case "approved", "posted", "paid", "finalized", "applied", "closed", "archived", "terminated", "deceased", "completed", "accepted", "issued", "signed", "hired", "processed", "rejected", "cancelled", "withdrawn", "superseded":
					return hrerrors.ErrWorkflowRequired
				}
			}
			return nil
		},
	}
}
func createWorkerAssignment(in CreateWorkerAssignmentInput) (*entity.WorkerAssignment, error) {
	v := &entity.WorkerAssignment{}
	v.WorkerID = in.WorkerID
	v.PositionID = in.PositionID
	v.ManagerWorkerID = in.ManagerWorkerID
	v.AssignmentType = in.AssignmentType
	v.EmploymentStatusSnapshot = in.EmploymentStatusSnapshot
	if in.FTEPercent != nil {
		v.FTEPercent = *in.FTEPercent
	} else {
		v.FTEPercent = entity.Decimal("100")
	}
	v.ValidFrom = in.ValidFrom
	v.ValidTo = in.ValidTo
	v.AssignmentReason = in.AssignmentReason
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	return v, nil
}
func patchWorkerAssignment(v *entity.WorkerAssignment, in UpdateWorkerAssignmentInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.PositionID.Set {
		if in.PositionID.Value == nil {
			return hrerrors.Invalid("position_id", "must not be null")
		}
		v.PositionID = *in.PositionID.Value
	}
	if in.ManagerWorkerID.Set {
		v.ManagerWorkerID = in.ManagerWorkerID.Value
	}
	if in.AssignmentType.Set {
		if in.AssignmentType.Value == nil {
			return hrerrors.Invalid("assignment_type", "must not be null")
		}
		v.AssignmentType = *in.AssignmentType.Value
	}
	if in.EmploymentStatusSnapshot.Set {
		v.EmploymentStatusSnapshot = in.EmploymentStatusSnapshot.Value
	}
	if in.FTEPercent.Set {
		if in.FTEPercent.Value == nil {
			return hrerrors.Invalid("fte_percent", "must not be null")
		}
		v.FTEPercent = *in.FTEPercent.Value
	}
	if in.ValidFrom.Set {
		if in.ValidFrom.Value == nil {
			return hrerrors.Invalid("valid_from", "must not be null")
		}
		v.ValidFrom = *in.ValidFrom.Value
	}
	if in.ValidTo.Set {
		v.ValidTo = in.ValidTo.Value
	}
	if in.AssignmentReason.Set {
		v.AssignmentReason = in.AssignmentReason.Value
	}
	if in.ApprovedByPrincipalID.Set {
		v.ApprovedByPrincipalID = in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		v.ApprovedAt = in.ApprovedAt.Value
	}
	return nil
}
func specWorkerAssignment() operation.Specification[entity.WorkerAssignment, string] {
	return operation.Specification[entity.WorkerAssignment, string]{
		Domain: "personnel", Table: "hr_worker_assignments",
		Key:         func(v *entity.WorkerAssignment) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.WorkerAssignment) string { return v.ID },
		Initialize: func(v *entity.WorkerAssignment, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			if actor.Kind == "human" {
				v.RequestedByPrincipalID = &actor.PrincipalID
				v.RequestedAt = &now
			}
		},
		Touch: func(v *entity.WorkerAssignment, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.WorkerAssignment) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["position_id"] = fmt.Sprint(v.PositionID)
			if v.RequestedByPrincipalID != nil {
				a["requested_by_principal_id"] = *v.RequestedByPrincipalID
			}
			if v.ManagerWorkerID != nil {
				a["manager_worker_id"] = fmt.Sprint(*v.ManagerWorkerID)
			}
			if v.ApprovedByPrincipalID != nil {
				a["approved_by_principal_id"] = fmt.Sprint(*v.ApprovedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.WorkerAssignment) error {
			if action == "create" && (next.RequestedByPrincipalID == nil || next.RequestedAt == nil) {
				return hrerrors.ErrForbidden
			}
			if action == "create" && next.EmploymentStatusSnapshot != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedByPrincipalID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedByPrincipalID, next.ApprovedByPrincipalID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.ApprovedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.ApprovedAt, next.ApprovedAt) {
				return hrerrors.ErrWorkflowRequired
			}
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createWorkerLifecycleEvent(in CreateWorkerLifecycleEventInput) (*entity.WorkerLifecycleEvent, error) {
	v := &entity.WorkerLifecycleEvent{}
	v.WorkerID = in.WorkerID
	v.EventType = in.EventType
	v.OccurredAt = in.OccurredAt
	v.EffectiveDate = in.EffectiveDate
	v.PreviousStatus = in.PreviousStatus
	v.NewStatus = in.NewStatus
	v.SourceReference = in.SourceReference
	if in.EventPayload != nil {
		v.EventPayload = in.EventPayload
	} else {
		v.EventPayload = entity.JSON(`{}`)
	}
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchWorkerLifecycleEvent(v *entity.WorkerLifecycleEvent, in UpdateWorkerLifecycleEventInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.EventType.Set {
		if in.EventType.Value == nil {
			return hrerrors.Invalid("event_type", "must not be null")
		}
		v.EventType = *in.EventType.Value
	}
	if in.OccurredAt.Set {
		if in.OccurredAt.Value == nil {
			return hrerrors.Invalid("occurred_at", "must not be null")
		}
		v.OccurredAt = *in.OccurredAt.Value
	}
	if in.EffectiveDate.Set {
		if in.EffectiveDate.Value == nil {
			return hrerrors.Invalid("effective_date", "must not be null")
		}
		v.EffectiveDate = *in.EffectiveDate.Value
	}
	if in.PreviousStatus.Set {
		v.PreviousStatus = in.PreviousStatus.Value
	}
	if in.NewStatus.Set {
		v.NewStatus = in.NewStatus.Value
	}
	if in.SourceReference.Set {
		v.SourceReference = in.SourceReference.Value
	}
	if in.EventPayload.Set {
		if in.EventPayload.Value == nil {
			v.EventPayload = nil
		} else {
			v.EventPayload = *in.EventPayload.Value
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
func specWorkerLifecycleEvent() operation.Specification[entity.WorkerLifecycleEvent, int64] {
	return operation.Specification[entity.WorkerLifecycleEvent, int64]{
		Domain: "personnel", Table: "hr_worker_lifecycle_events",
		Key:       func(v *entity.WorkerLifecycleEvent) int64 { return v.ID },
		KeyString: func(key int64) string { return fmt.Sprint(key) },
		ValidateKey: func(key int64) error {
			return func() error {
				if key <= 0 {
					return hrerrors.ErrInvalidInput
				}
				return nil
			}()
		},
		AggregateID: func(v *entity.WorkerLifecycleEvent) string { return v.WorkerID },
		Initialize: func(v *entity.WorkerLifecycleEvent, actor operation.Actor, now time.Time, id string) {
			v.ActorPrincipalID = &actor.PrincipalID
		},
		Touch: func(v *entity.WorkerLifecycleEvent, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.WorkerLifecycleEvent) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			if v.PreviousStatus != nil {
				a["previous_status"] = fmt.Sprint(*v.PreviousStatus)
			}
			if v.NewStatus != nil {
				a["new_status"] = fmt.Sprint(*v.NewStatus)
			}
			if v.ActorPrincipalID != nil {
				a["actor_principal_id"] = fmt.Sprint(*v.ActorPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.WorkerLifecycleEvent) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.PreviousStatus, next.PreviousStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.NewStatus, next.NewStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createWorkerDocument(in CreateWorkerDocumentInput) (*entity.WorkerDocument, error) {
	v := &entity.WorkerDocument{}
	v.WorkerID = in.WorkerID
	v.DocumentType = in.DocumentType
	v.DocumentReferenceID = in.DocumentReferenceID
	v.DocumentVersionReferenceID = in.DocumentVersionReferenceID
	v.ContentDigest = in.ContentDigest
	v.DocumentNumberDigest = in.DocumentNumberDigest
	v.IssuingCountryCode = in.IssuingCountryCode
	v.IssuedOn = in.IssuedOn
	v.ExpiresOn = in.ExpiresOn
	if in.VerificationStatus != nil {
		v.VerificationStatus = *in.VerificationStatus
	} else {
		v.VerificationStatus = "pending"
	}
	v.VerifiedByPrincipalID = in.VerifiedByPrincipalID
	v.VerifiedAt = in.VerifiedAt
	v.RetentionClass = in.RetentionClass
	if in.AccessClass != nil {
		v.AccessClass = *in.AccessClass
	} else {
		v.AccessClass = "restricted"
	}
	return v, nil
}
func patchWorkerDocument(v *entity.WorkerDocument, in UpdateWorkerDocumentInput) error {
	previousDocumentID := v.DocumentReferenceID
	previousVersionID := v.DocumentVersionReferenceID
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.DocumentType.Set {
		if in.DocumentType.Value == nil {
			return hrerrors.Invalid("document_type", "must not be null")
		}
		v.DocumentType = *in.DocumentType.Value
	}
	if in.DocumentReferenceID.Set {
		if in.DocumentReferenceID.Value == nil {
			return hrerrors.Invalid("document_reference_id", "must not be null")
		}
		v.DocumentReferenceID = *in.DocumentReferenceID.Value
	}
	if in.DocumentVersionReferenceID.Set {
		if in.DocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("document_version_reference_id", "must not be null")
		}
		v.DocumentVersionReferenceID = *in.DocumentVersionReferenceID.Value
	}
	documentChanged := !strings.EqualFold(previousDocumentID, v.DocumentReferenceID)
	versionChanged := !strings.EqualFold(previousVersionID, v.DocumentVersionReferenceID)
	if documentChanged && !versionChanged {
		return hrerrors.Invalid("document_version_reference_id", "a new version is required when replacing the document")
	}
	if versionChanged && !in.ContentDigest.Set {
		return hrerrors.Invalid("content_digest", "must be supplied when replacing the document version")
	}
	if in.ContentDigest.Set {
		if in.ContentDigest.Value == nil {
			v.ContentDigest = nil
		} else {
			v.ContentDigest = *in.ContentDigest.Value
		}
	}
	if in.DocumentNumberDigest.Set {
		if in.DocumentNumberDigest.Value == nil {
			v.DocumentNumberDigest = nil
		} else {
			v.DocumentNumberDigest = *in.DocumentNumberDigest.Value
		}
	}
	if in.IssuingCountryCode.Set {
		v.IssuingCountryCode = in.IssuingCountryCode.Value
	}
	if in.IssuedOn.Set {
		v.IssuedOn = in.IssuedOn.Value
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
	if in.RetentionClass.Set {
		if in.RetentionClass.Value == nil {
			return hrerrors.Invalid("retention_class", "must not be null")
		}
		v.RetentionClass = *in.RetentionClass.Value
	}
	if in.AccessClass.Set {
		if in.AccessClass.Value == nil {
			return hrerrors.Invalid("access_class", "must not be null")
		}
		v.AccessClass = *in.AccessClass.Value
	}
	return nil
}
func specWorkerDocument() operation.Specification[entity.WorkerDocument, string] {
	return operation.Specification[entity.WorkerDocument, string]{
		Domain: "personnel", Table: "hr_worker_documents",
		Key:         func(v *entity.WorkerDocument) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.WorkerDocument) string { return v.ID },
		Initialize: func(v *entity.WorkerDocument, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.CreatedAt = now
		},
		Touch: func(v *entity.WorkerDocument, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.WorkerDocument) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["document_reference_id"] = fmt.Sprint(v.DocumentReferenceID)
			a["document_version_reference_id"] = fmt.Sprint(v.DocumentVersionReferenceID)
			a["verification_status"] = fmt.Sprint(v.VerificationStatus)
			if v.VerifiedByPrincipalID != nil {
				a["verified_by_principal_id"] = fmt.Sprint(*v.VerifiedByPrincipalID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.WorkerDocument) error {
			if action == "update" && !reflect.DeepEqual(old.VerificationStatus, next.VerificationStatus) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.VerificationStatus) != "pending" {
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
func createEmergencyContact(in CreateEmergencyContactInput) (*entity.EmergencyContact, error) {
	v := &entity.EmergencyContact{}
	v.WorkerID = in.WorkerID
	v.ContactPartyReferenceID = in.ContactPartyReferenceID
	v.RelationshipCode = in.RelationshipCode
	v.PriorityOrder = in.PriorityOrder
	v.AuthorizationConfirmedAt = in.AuthorizationConfirmedAt
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	} else {
		v.IsActive = true
	}
	return v, nil
}
func patchEmergencyContact(v *entity.EmergencyContact, in UpdateEmergencyContactInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.ContactPartyReferenceID.Set {
		if in.ContactPartyReferenceID.Value == nil {
			return hrerrors.Invalid("contact_party_reference_id", "must not be null")
		}
		if !in.AuthorizationConfirmedAt.Set && !strings.EqualFold(v.ContactPartyReferenceID, *in.ContactPartyReferenceID.Value) {
			v.AuthorizationConfirmedAt = nil
		}
		v.ContactPartyReferenceID = *in.ContactPartyReferenceID.Value
	}
	if in.RelationshipCode.Set {
		if in.RelationshipCode.Value == nil {
			return hrerrors.Invalid("relationship_code", "must not be null")
		}
		v.RelationshipCode = *in.RelationshipCode.Value
	}
	if in.PriorityOrder.Set {
		if in.PriorityOrder.Value == nil {
			return hrerrors.Invalid("priority_order", "must not be null")
		}
		v.PriorityOrder = *in.PriorityOrder.Value
	}
	if in.AuthorizationConfirmedAt.Set {
		v.AuthorizationConfirmedAt = in.AuthorizationConfirmedAt.Value
	}
	if in.IsActive.Set {
		if in.IsActive.Value == nil {
			return hrerrors.Invalid("is_active", "must not be null")
		}
		v.IsActive = *in.IsActive.Value
	}
	return nil
}
func specEmergencyContact() operation.Specification[entity.EmergencyContact, string] {
	return operation.Specification[entity.EmergencyContact, string]{
		Domain: "personnel", Table: "hr_emergency_contacts",
		Key:         func(v *entity.EmergencyContact) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.EmergencyContact) string { return v.ID },
		Initialize: func(v *entity.EmergencyContact, actor operation.Actor, now time.Time, id string) {
			v.ID = id
		},
		Touch: func(v *entity.EmergencyContact, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.EmergencyContact) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["contact_party_reference_id"] = fmt.Sprint(v.ContactPartyReferenceID)
			return a
		},
		Guard: func(action string, old, next *entity.EmergencyContact) error {
			return nil
		},
	}
}
