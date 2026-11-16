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

func createServiceRequest(in CreateServiceRequestInput) (*entity.ServiceRequest, error) {
	v := &entity.ServiceRequest{}
	v.RequestNumber = in.RequestNumber
	v.WorkerID = in.WorkerID
	v.RequestType = in.RequestType
	v.AssignedPrincipalID = in.AssignedPrincipalID
	if in.Status != nil {
		v.Status = *in.Status
	} else {
		v.Status = "open"
	}
	v.RestrictedDetailsDocumentVersionReferenceID = in.RestrictedDetailsDocumentVersionReferenceID
	v.DueAt = in.DueAt
	v.ResolvedAt = in.ResolvedAt
	v.ResolutionDocumentVersionReferenceID = in.ResolutionDocumentVersionReferenceID
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchServiceRequest(v *entity.ServiceRequest, in UpdateServiceRequestInput) error {
	if in.RequestNumber.Set {
		if in.RequestNumber.Value == nil {
			return hrerrors.Invalid("request_number", "must not be null")
		}
		v.RequestNumber = *in.RequestNumber.Value
	}
	if in.WorkerID.Set {
		if in.WorkerID.Value == nil {
			return hrerrors.Invalid("worker_id", "must not be null")
		}
		v.WorkerID = *in.WorkerID.Value
	}
	if in.RequestType.Set {
		if in.RequestType.Value == nil {
			return hrerrors.Invalid("request_type", "must not be null")
		}
		v.RequestType = *in.RequestType.Value
	}
	if in.AssignedPrincipalID.Set {
		v.AssignedPrincipalID = in.AssignedPrincipalID.Value
	}
	if in.Status.Set {
		if in.Status.Value == nil {
			return hrerrors.Invalid("status", "must not be null")
		}
		v.Status = *in.Status.Value
	}
	if in.RestrictedDetailsDocumentVersionReferenceID.Set {
		if in.RestrictedDetailsDocumentVersionReferenceID.Value == nil {
			return hrerrors.Invalid("restricted_details_document_version_reference_id", "must not be null")
		}
		v.RestrictedDetailsDocumentVersionReferenceID = *in.RestrictedDetailsDocumentVersionReferenceID.Value
	}
	if in.DueAt.Set {
		v.DueAt = in.DueAt.Value
	}
	if in.ResolvedAt.Set {
		v.ResolvedAt = in.ResolvedAt.Value
	}
	if in.ResolutionDocumentVersionReferenceID.Set {
		v.ResolutionDocumentVersionReferenceID = in.ResolutionDocumentVersionReferenceID.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specServiceRequest() operation.Specification[entity.ServiceRequest, string] {
	return operation.Specification[entity.ServiceRequest, string]{
		Domain: "hrservice", Table: "hr_service_requests",
		Key:         func(v *entity.ServiceRequest) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.ServiceRequest) string { return v.ID },
		Initialize: func(v *entity.ServiceRequest, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.RequestedByPrincipalID = actor.PrincipalID
			v.OpenedAt = now
		},
		Touch: func(v *entity.ServiceRequest, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.ServiceRequest) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["worker_id"] = fmt.Sprint(v.WorkerID)
			a["requested_by_principal_id"] = fmt.Sprint(v.RequestedByPrincipalID)
			if v.AssignedPrincipalID != nil {
				a["assigned_principal_id"] = fmt.Sprint(*v.AssignedPrincipalID)
			}
			a["status"] = fmt.Sprint(v.Status)
			a["restricted_details_document_version_reference_id"] = fmt.Sprint(v.RestrictedDetailsDocumentVersionReferenceID)
			if v.ResolutionDocumentVersionReferenceID != nil {
				a["resolution_document_version_reference_id"] = fmt.Sprint(*v.ResolutionDocumentVersionReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.ServiceRequest) error {
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
func createServiceRequestEvent(in CreateServiceRequestEventInput) (*entity.ServiceRequestEvent, error) {
	v := &entity.ServiceRequestEvent{}
	v.ServiceRequestID = in.ServiceRequestID
	v.EventType = in.EventType
	v.Decision = in.Decision
	v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID
	v.IdempotencyKey = in.IdempotencyKey
	return v, nil
}
func patchServiceRequestEvent(v *entity.ServiceRequestEvent, in UpdateServiceRequestEventInput) error {
	if in.ServiceRequestID.Set {
		if in.ServiceRequestID.Value == nil {
			return hrerrors.Invalid("service_request_id", "must not be null")
		}
		v.ServiceRequestID = *in.ServiceRequestID.Value
	}
	if in.EventType.Set {
		if in.EventType.Value == nil {
			return hrerrors.Invalid("event_type", "must not be null")
		}
		v.EventType = *in.EventType.Value
	}
	if in.Decision.Set {
		v.Decision = in.Decision.Value
	}
	if in.EvidenceDocumentVersionReferenceID.Set {
		v.EvidenceDocumentVersionReferenceID = in.EvidenceDocumentVersionReferenceID.Value
	}
	if in.IdempotencyKey.Set {
		if in.IdempotencyKey.Value == nil {
			return hrerrors.Invalid("idempotency_key", "must not be null")
		}
		v.IdempotencyKey = *in.IdempotencyKey.Value
	}
	return nil
}
func specServiceRequestEvent() operation.Specification[entity.ServiceRequestEvent, string] {
	return operation.Specification[entity.ServiceRequestEvent, string]{
		Domain: "hrservice", Table: "hr_service_request_events",
		Key:         func(v *entity.ServiceRequestEvent) string { return v.ID },
		KeyString:   func(key string) string { return fmt.Sprint(key) },
		ValidateKey: func(key string) error { return entity.ValidateUUID("id", key) },
		AggregateID: func(v *entity.ServiceRequestEvent) string { return v.ID },
		Initialize: func(v *entity.ServiceRequestEvent, actor operation.Actor, now time.Time, id string) {
			v.ID = id
			v.ActorPrincipalID = actor.PrincipalID
			v.OccurredAt = now
		},
		Touch: func(v *entity.ServiceRequestEvent, actor operation.Actor, now time.Time) {},
		Attributes: func(v *entity.ServiceRequestEvent) map[string]string {
			a := map[string]string{}
			a["id"] = fmt.Sprint(v.ID)
			a["service_request_id"] = fmt.Sprint(v.ServiceRequestID)
			a["actor_principal_id"] = fmt.Sprint(v.ActorPrincipalID)
			if v.EvidenceDocumentVersionReferenceID != nil {
				a["evidence_document_version_reference_id"] = fmt.Sprint(*v.EvidenceDocumentVersionReferenceID)
			}
			return a
		},
		Guard: func(action string, old, next *entity.ServiceRequestEvent) error {
			if action != "create" {
				return hrerrors.ErrWorkflowRequired
			}
			return nil
		},
	}
}
