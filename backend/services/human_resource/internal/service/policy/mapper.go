// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"fmt"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

func createPolicy(in CreatePolicyInput) (*entity.Policy, error) {
	v := &entity.Policy{}
	v.PolicyCode = in.PolicyCode
	v.PolicyVersion = in.PolicyVersion
	v.PolicyName = in.PolicyName
	v.CategoryCode = in.CategoryCode
	v.LegalEntityReferenceID = in.LegalEntityReferenceID
	v.EffectiveFrom = in.EffectiveFrom
	v.EffectiveTo = in.EffectiveTo
	v.DocumentVersionReferenceID = in.DocumentVersionReferenceID
	v.ContentDigest = in.ContentDigest
	v.ApprovedByPrincipalID = in.ApprovedByPrincipalID
	v.ApprovedAt = in.ApprovedAt
	return v, nil
}
func patchPolicy(v *entity.Policy, in UpdatePolicyInput) error {
	if in.PolicyCode.Set {
		if in.PolicyCode.Value == nil {
			return hrerrors.Invalid("policy_code", "must not be null")
		}
		v.PolicyCode = *in.PolicyCode.Value
	}
	if in.PolicyVersion.Set {
		if in.PolicyVersion.Value == nil {
			return hrerrors.Invalid("policy_version", "must not be null")
		}
		v.PolicyVersion = *in.PolicyVersion.Value
	}
	if in.PolicyName.Set {
		if in.PolicyName.Value == nil {
			return hrerrors.Invalid("policy_name", "must not be null")
		}
		v.PolicyName = *in.PolicyName.Value
	}
	if in.CategoryCode.Set {
		if in.CategoryCode.Value == nil {
			return hrerrors.Invalid("category_code", "must not be null")
		}
		v.CategoryCode = *in.CategoryCode.Value
	}
	if in.LegalEntityReferenceID.Set {
		if in.LegalEntityReferenceID.Value == nil {
			return hrerrors.Invalid("legal_entity_reference_id", "must not be null")
		}
		v.LegalEntityReferenceID = *in.LegalEntityReferenceID.Value
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
	if in.ApprovedByPrincipalID.Set {
		if in.ApprovedByPrincipalID.Value == nil {
			return hrerrors.Invalid("approved_by_principal_id", "must not be null")
		}
		v.ApprovedByPrincipalID = *in.ApprovedByPrincipalID.Value
	}
	if in.ApprovedAt.Set {
		if in.ApprovedAt.Value == nil {
			return hrerrors.Invalid("approved_at", "must not be null")
		}
		v.ApprovedAt = *in.ApprovedAt.Value
	}
	return nil
}
func specPolicy() operation.Specification[entity.Policy, string] {
	return operation.Specification[entity.Policy, string]{
		Domain: "policy", Table: "hr_policies",
		Key:         func(v *entity.Policy) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.Policy) string { return v.ID },
		Initialize: func(v *entity.Policy, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.ApprovedByPrincipalID = actor.PrincipalID
			v.ApprovedAt = now
		},
		Touch: func(v *entity.Policy, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.Policy) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["legal_entity_reference_id"] = fmt.Sprint(v.LegalEntityReferenceID)
			a["document_version_reference_id"] = fmt.Sprint(v.DocumentVersionReferenceID)
			a["approved_by_principal_id"] = fmt.Sprint(v.ApprovedByPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.Policy) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
func createPolicyAcknowledgement(in CreatePolicyAcknowledgementInput) (*entity.PolicyAcknowledgement, error) {
	v := &entity.PolicyAcknowledgement{}
	v.PolicyID = in.PolicyID
	v.WorkerID = in.WorkerID
	v.DueAt = in.DueAt
	v.AcknowledgedAt = in.AcknowledgedAt
	v.AcknowledgedByPrincipalID = in.AcknowledgedByPrincipalID
	v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID
	return v, nil
}
func patchPolicyAcknowledgement(v *entity.PolicyAcknowledgement, in UpdatePolicyAcknowledgementInput) error {
	if in.PolicyID.Set {
		if in.PolicyID.Value == nil {
			return hrerrors.Invalid("policy_id", "must not be null")
		}
		v.PolicyID = *in.PolicyID.Value
	}
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.DueAt.Set {
		v.DueAt = in.DueAt.Value
	}
	if in.AcknowledgedAt.Set {
		v.AcknowledgedAt = in.AcknowledgedAt.Value
	}
	if in.AcknowledgedByPrincipalID.Set {
		v.AcknowledgedByPrincipalID = in.AcknowledgedByPrincipalID.Value
	}
	if in.EvidenceDocumentVersionReferenceID.Set {
		v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID.Value
	}
	return nil
}
func specPolicyAcknowledgement() operation.Specification[entity.PolicyAcknowledgement, string] {
	return operation.Specification[entity.PolicyAcknowledgement, string]{
		Domain: "policy", Table: "hr_policy_acknowledgements",
		Key:         func(v *entity.PolicyAcknowledgement) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.PolicyAcknowledgement) string { return v.ID },
		Initialize: func(v *entity.PolicyAcknowledgement, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.AssignedAt = now
		},
		Touch: func(v *entity.PolicyAcknowledgement, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.PolicyAcknowledgement) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["policy_id"] = fmt.Sprint(v.PolicyID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			if v.AcknowledgedByPrincipalID != nil {
				a["acknowledged_by_principal_id"] = fmt.Sprint(*v.AcknowledgedByPrincipalID)
			}
			if v.EvidenceDocumentVersionReferenceID != nil {
				a["evidence_document_version_reference_id"] = fmt.Sprint(*v.EvidenceDocumentVersionReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.PolicyAcknowledgement) error {
			if action != "create" && old.AcknowledgedAt != nil {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
