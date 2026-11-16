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

func createAccessRequest(in CreateAccessRequestInput) (*entity.AccessRequest, error) {
	v := &entity.AccessRequest{}
	v.WorkerID = in.WorkerID
	v.LifecycleEventID = in.LifecycleEventID
	v.RequestKind = in.RequestKind
	v.RequestedScopeType = in.RequestedScopeType
	v.RequestedScopeReferenceID = in.RequestedScopeReferenceID
	v.ApprovedCatalogReferenceID = in.ApprovedCatalogReferenceID
	v.IAMRequestReferenceID = in.IAMRequestReferenceID
	v.IAMDecisionEventReferenceID = in.IAMDecisionEventReferenceID
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "pending"
	}
	v.LastReconciledAt = in.LastReconciledAt
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchAccessRequest(v *entity.AccessRequest, in UpdateAccessRequestInput) error {
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.LifecycleEventID.Set {
		if in.LifecycleEventID.Value == nil {
			return hrerrors.Invalid("lifecycle_event_id", "must not be null")
		}
		v.LifecycleEventID = *in.LifecycleEventID.Value
	}
	if in.RequestKind.Set {
		if in.RequestKind.Value == nil {
			return hrerrors.Invalid("request_kind", "must not be null")
		}
		v.RequestKind = *in.RequestKind.Value
	}
	if in.RequestedScopeType.Set {
		if in.RequestedScopeType.Value == nil {
			return hrerrors.Invalid("requested_scope_type", "must not be null")
		}
		v.RequestedScopeType = *in.RequestedScopeType.Value
	}
	if in.RequestedScopeReferenceID.Set {
		if in.RequestedScopeReferenceID.Value == nil {
			return hrerrors.Invalid("requested_scope_reference_id", "must not be null")
		}
		v.RequestedScopeReferenceID = *in.RequestedScopeReferenceID.Value
	}
	if in.ApprovedCatalogReferenceID.Set {
		v.ApprovedCatalogReferenceID = in.ApprovedCatalogReferenceID.Value
	}
	if in.IAMRequestReferenceID.Set {
		v.IAMRequestReferenceID = in.IAMRequestReferenceID.Value
	}
	if in.IAMDecisionEventReferenceID.Set {
		v.IAMDecisionEventReferenceID = in.IAMDecisionEventReferenceID.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.LastReconciledAt.Set {
		v.LastReconciledAt = in.LastReconciledAt.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specAccessRequest() operation.Specification[entity.AccessRequest, string] {
	return operation.Specification[entity.AccessRequest, string]{
		Domain: "accessrequest", Table: "hr_access_requests",
		Key:         func(v *entity.AccessRequest) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.AccessRequest) string { return v.ID },
		Initialize: func(v *entity.AccessRequest, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.RequestedByPrincipalID = actor.PrincipalID
			v.RequestedAt = now
		},
		Touch: func(v *entity.AccessRequest, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.AccessRequest) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["requested_scope_reference_id"] = fmt.Sprint(v.RequestedScopeReferenceID)
			if v.ApprovedCatalogReferenceID != nil {
				a["approved_catalog_reference_id"] = fmt.Sprint(*v.ApprovedCatalogReferenceID)
			}
			if v.IAMRequestReferenceID != nil {
				a["iam_request_reference_id"] = fmt.Sprint(*v.IAMRequestReferenceID)
			}
			if v.IAMDecisionEventReferenceID != nil {
				a["iam_decision_event_reference_id"] = fmt.Sprint(*v.IAMDecisionEventReferenceID)
			}
			a["status"] = fmt.Sprint(v.Status)
			a["requested_by_principal_id"] = fmt.Sprint(v.RequestedByPrincipalID)
			return a
		},
		Guard: func(action string, old, next *entity.AccessRequest) error {
			if action == "create" && next.IAMRequestReferenceID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.IAMRequestReferenceID, next.IAMRequestReferenceID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && next.IAMDecisionEventReferenceID != nil {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.IAMDecisionEventReferenceID, next.IAMDecisionEventReferenceID) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "update" && !reflect.DeepEqual(old.Status, next.Status) {
				return hrerrors.ErrWorkflowRequired
			}
			if action == "create" && string(next.Status) != "pending" {
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
