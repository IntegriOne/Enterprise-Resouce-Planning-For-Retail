package human_resource

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	common "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/common/v1"
	personnel "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/personnel/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func toModelWorkerLifecycleEvent(row *personnel.WorkerLifecycleEvent) (*model.WorkerLifecycleEvent, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker-lifecycle-event response", nil)
	}
	if row.GetId() <= 0 {
		return nil, middleware.ErrUnavailable("invalid human resource worker-lifecycle-event ID", nil)
	}
	rawID := strconv.FormatInt(row.GetId(), 10)
	out := &model.WorkerLifecycleEvent{ID: int64NodeID("worker-lifecycle-event", row.GetId()), RawID: rawID}
	out.WorkerID = row.GetWorkerId()
	out.EventType = row.GetEventType()
	out.OccurredAt = pbTimePtr(row.GetOccurredAt())
	out.EffectiveDate = row.GetEffectiveDate()
	out.PreviousStatus = row.PreviousStatus
	out.NewStatus = row.NewStatus
	out.SourceReference = row.SourceReference
	out.ActorPrincipalID = row.ActorPrincipalId
	if doc := row.GetEventPayload(); doc != nil {
		out.EventPayload = pbJSONString(doc.GetUtf8())
	}
	out.IdempotencyKey = row.GetIdempotencyKey()
	return out, nil
}

// WorkerLifecycleEvent membaca detail worker-lifecycle-event berdasarkan ID.
func (a *Adapter) WorkerLifecycleEvent(ctx context.Context, rawID string) (*model.WorkerLifecycleEvent, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_lifecycle_events", "human_resource.worker_lifecycle_events.read")
	if err != nil {
		return nil, err
	}
	id, err := parsePositiveInt64NodeID(rawID, "worker-lifecycle-event")
	if err != nil {
		return nil, err
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.personnel.GetWorkerLifecycleEventByID(rpcCtx, &common.IdentityIDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "worker-lifecycle-event")
	}
	return toModelWorkerLifecycleEvent(resp.GetItem())
}

// WorkerLifecycleEvents mendaftar dan memfilter worker-lifecycle-event dengan paginasi.
func (a *Adapter) WorkerLifecycleEvents(ctx context.Context, limit, offset int, f_worker_id *string, f_event_type *string, f_previous_status *string, f_new_status *string, f_actor_principal_id *string, f_idempotency_key *string) (*Page[model.WorkerLifecycleEvent], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_lifecycle_events", "human_resource.worker_lifecycle_events.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid worker-lifecycle-event pagination", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_event_type != nil, f_previous_status != nil, f_new_status != nil, f_actor_principal_id != nil, f_idempotency_key != nil)
	if err != nil {
		return nil, err
	}
	var resp *personnel.WorkerLifecycleEventsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerLifecycleEventByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_event_type, "event_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerLifecycleEventByEventType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_previous_status, "previous_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerLifecycleEventByPreviousStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_new_status, "new_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerLifecycleEventByNewStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_actor_principal_id, "actor_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerLifecycleEventByActorPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_idempotency_key, "idempotency_key")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerLifecycleEventByIdempotencyKey(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.personnel.GetAllWorkerLifecycleEvents(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "worker-lifecycle-event")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "worker-lifecycle-event", toModelWorkerLifecycleEvent)
}

func toModelWorkerDocument(row *personnel.WorkerDocument) (*model.WorkerDocument, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker-document response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource worker-document ID", nil)
	}
	out := &model.WorkerDocument{ID: nodeID("worker-document", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.DocumentType = row.GetDocumentType()
	out.DocumentReferenceID = row.GetDocumentReferenceId()
	out.DocumentVersionReferenceID = row.GetDocumentVersionReferenceId()
	out.ContentDigest = pbBase64String(row.GetContentDigest())
	out.DocumentNumberDigest = pbBase64String(row.GetDocumentNumberDigest())
	out.IssuingCountryCode = row.IssuingCountryCode
	out.IssuedOn = row.IssuedOn
	out.ExpiresOn = row.ExpiresOn
	out.VerificationStatus = row.GetVerificationStatus()
	out.VerifiedByPrincipalID = row.VerifiedByPrincipalId
	out.VerifiedAt = pbTimePtr(row.GetVerifiedAt())
	out.RetentionClass = row.GetRetentionClass()
	out.AccessClass = row.GetAccessClass()
	out.CreatedAt = pbTimePtr(row.GetCreatedAt())
	return out, nil
}

// WorkerDocument membaca detail worker-document berdasarkan ID.
func (a *Adapter) WorkerDocument(ctx context.Context, rawID string) (*model.WorkerDocument, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_documents", "human_resource.worker_documents.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "worker-document", "worker-document")
	if err != nil {
		return nil, err
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.personnel.GetWorkerDocumentByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "worker-document")
	}
	return toModelWorkerDocument(resp.GetItem())
}

// WorkerDocuments mendaftar dan memfilter worker-document dengan paginasi.
func (a *Adapter) WorkerDocuments(ctx context.Context, limit, offset int, f_worker_id *string, f_document_type *string, f_document_reference_id *string, f_document_version_reference_id *string, f_issuing_country_code *string, f_verification_status *string, f_verified_by_principal_id *string) (*Page[model.WorkerDocument], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_worker_documents", "human_resource.worker_documents.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid worker-document pagination", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_document_type != nil, f_document_reference_id != nil, f_document_version_reference_id != nil, f_issuing_country_code != nil, f_verification_status != nil, f_verified_by_principal_id != nil)
	if err != nil {
		return nil, err
	}
	var resp *personnel.WorkerDocumentsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerDocumentByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_document_type, "document_type")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerDocumentByDocumentType(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_document_reference_id, "document_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerDocumentByDocumentReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 3:
		v, fErr := validateFilterString(f_document_version_reference_id, "document_version_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerDocumentByDocumentVersionReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 4:
		v, fErr := validateFilterString(f_issuing_country_code, "issuing_country_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerDocumentByIssuingCountryCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 5:
		v, fErr := validateFilterString(f_verification_status, "verification_status")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerDocumentByVerificationStatus(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 6:
		v, fErr := validateFilterString(f_verified_by_principal_id, "verified_by_principal_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindWorkerDocumentByVerifiedByPrincipalID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.personnel.GetAllWorkerDocuments(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "worker-document")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "worker-document", toModelWorkerDocument)
}

// CreateWorkerDocument membuat worker-document baru melalui HRM.
func (a *Adapter) CreateWorkerDocument(ctx context.Context, input model.CreateWorkerDocumentInput) (*model.WorkerDocument, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_worker_documents", "create", "human_resource.worker_documents.create")
	if err != nil {
		return nil, err
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &personnel.WorkerDocumentInput{}
	patch.WorkerId = input.WorkerID
	patch.DocumentType = input.DocumentType
	patch.DocumentReferenceId = input.DocumentReferenceID
	patch.DocumentVersionReferenceId = input.DocumentVersionReferenceID
	if b, bErr := optionalBase64Bytes(input.ContentDigest, "content_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.ContentDigest = b
	}
	if b, bErr := optionalBase64Bytes(input.DocumentNumberDigest, "document_number_digest"); bErr != nil {
		return nil, bErr
	} else if b != nil {
		patch.DocumentNumberDigest = b
	}
	patch.IssuingCountryCode = input.IssuingCountryCode
	patch.IssuedOn = input.IssuedOn
	patch.ExpiresOn = input.ExpiresOn
	patch.VerificationStatus = input.VerificationStatus
	patch.VerifiedByPrincipalId = input.VerifiedByPrincipalID
	patch.VerifiedAt = optionalTimestamp(input.VerifiedAt)
	patch.RetentionClass = input.RetentionClass
	patch.AccessClass = input.AccessClass
	resp, err := a.personnel.CreateWorkerDocument(rpcCtx, &personnel.CreateWorkerDocumentRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "worker-document")
	}
	return toModelWorkerDocument(resp.GetItem())
}

// UpdateWorkerDocument memperbarui worker-document berdasarkan FieldMask.
func (a *Adapter) UpdateWorkerDocument(ctx context.Context, rawID string, input model.UpdateWorkerDocumentInput) (*model.WorkerDocument, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_worker_documents", "update", "human_resource.worker_documents.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "worker-document", "worker-document")
	if err != nil {
		return nil, err
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &personnel.WorkerDocumentInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "document_type", input.DocumentType, &patch.DocumentType, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "document_reference_id", input.DocumentReferenceID, &patch.DocumentReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "document_version_reference_id", input.DocumentVersionReferenceID, &patch.DocumentVersionReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "content_digest", input.ContentDigest, &patch.ContentDigest, true); err != nil {
		return nil, err
	}
	if err := applyBytesMask(mask, "document_number_digest", input.DocumentNumberDigest, &patch.DocumentNumberDigest, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "issuing_country_code", input.IssuingCountryCode, &patch.IssuingCountryCode, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "issued_on", input.IssuedOn, &patch.IssuedOn, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "expires_on", input.ExpiresOn, &patch.ExpiresOn, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "verification_status", input.VerificationStatus, &patch.VerificationStatus, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "verified_by_principal_id", input.VerifiedByPrincipalID, &patch.VerifiedByPrincipalId, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "verified_at", input.VerifiedAt, &patch.VerifiedAt, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "retention_class", input.RetentionClass, &patch.RetentionClass, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "access_class", input.AccessClass, &patch.AccessClass, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("worker-document update is empty", nil)
	}
	resp, err := a.personnel.UpdateWorkerDocument(rpcCtx, &personnel.UpdateWorkerDocumentRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "worker-document")
	}
	return toModelWorkerDocument(resp.GetItem())
}

// DeleteWorkerDocument menghapus worker-document berdasarkan ID.
func (a *Adapter) DeleteWorkerDocument(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_worker_documents", "human_resource.worker_documents.delete", "worker-document", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.personnel.DeleteWorkerDocument(c, req)
	})
}

func toModelEmergencyContact(row *personnel.EmergencyContact) (*model.EmergencyContact, error) {
	if row == nil {
		return nil, middleware.ErrUnavailable("invalid human resource emergency-contact response", nil)
	}
	if id, err := uuid.Parse(row.GetId()); err != nil || id == uuid.Nil {
		return nil, middleware.ErrUnavailable("invalid human resource emergency-contact ID", nil)
	}
	out := &model.EmergencyContact{ID: nodeID("emergency-contact", row.GetId()), RawID: row.GetId()}
	out.WorkerID = row.GetWorkerId()
	out.ContactPartyReferenceID = row.GetContactPartyReferenceId()
	out.RelationshipCode = row.GetRelationshipCode()
	out.PriorityOrder = int(row.GetPriorityOrder())
	out.AuthorizationConfirmedAt = pbTimePtr(row.GetAuthorizationConfirmedAt())
	out.IsActive = row.GetIsActive()
	return out, nil
}

// EmergencyContact membaca detail emergency-contact berdasarkan ID.
func (a *Adapter) EmergencyContact(ctx context.Context, rawID string) (*model.EmergencyContact, error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_emergency_contacts", "human_resource.emergency_contacts.read")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "emergency-contact", "emergency-contact")
	if err != nil {
		return nil, err
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.personnel.GetEmergencyContactByID(rpcCtx, &common.IDRequest{Id: id})
	if err != nil {
		return nil, mapReadError(err, "emergency-contact")
	}
	return toModelEmergencyContact(resp.GetItem())
}

// EmergencyContacts mendaftar dan memfilter emergency-contact dengan paginasi.
func (a *Adapter) EmergencyContacts(ctx context.Context, limit, offset int, f_worker_id *string, f_contact_party_reference_id *string, f_relationship_code *string) (*Page[model.EmergencyContact], error) {
	rpcCtx, err := a.authorizeResourceRead(ctx, "hr_emergency_contacts", "human_resource.emergency_contacts.read")
	if err != nil {
		return nil, err
	}
	if !validPagination(limit, offset) {
		return nil, middleware.ErrInvalidArgument("invalid emergency-contact pagination", nil)
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	idx, err := pickSingleFilter(f_worker_id != nil, f_contact_party_reference_id != nil, f_relationship_code != nil)
	if err != nil {
		return nil, err
	}
	var resp *personnel.EmergencyContactsResponse
	switch idx {
	case 0:
		v, fErr := validateFilterString(f_worker_id, "worker_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmergencyContactByWorkerID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 1:
		v, fErr := validateFilterString(f_contact_party_reference_id, "contact_party_reference_id")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmergencyContactByContactPartyReferenceID(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	case 2:
		v, fErr := validateFilterString(f_relationship_code, "relationship_code")
		if fErr != nil {
			return nil, fErr
		}
		resp, err = a.personnel.FindEmergencyContactByRelationshipCode(rpcCtx, &common.StringFinderRequest{Value: &v, Limit: int32(limit), Offset: int32(offset)})
	default:
		resp, err = a.personnel.GetAllEmergencyContacts(rpcCtx, &common.PageRequest{Limit: int32(limit), Offset: int32(offset)})
	}
	if err != nil {
		return nil, mapReadError(err, "emergency-contact")
	}
	return mapRelayPage(resp.GetPage(), resp.GetItems(), limit, offset, "emergency-contact", toModelEmergencyContact)
}

// CreateEmergencyContact membuat emergency-contact baru melalui HRM.
func (a *Adapter) CreateEmergencyContact(ctx context.Context, input model.CreateEmergencyContactInput) (*model.EmergencyContact, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_emergency_contacts", "create", "human_resource.emergency_contacts.create")
	if err != nil {
		return nil, err
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &personnel.EmergencyContactInput{}
	patch.WorkerId = input.WorkerID
	patch.ContactPartyReferenceId = input.ContactPartyReferenceID
	patch.RelationshipCode = input.RelationshipCode
	patch.PriorityOrder = optionalInt32Ptr(input.PriorityOrder)
	patch.AuthorizationConfirmedAt = optionalTimestamp(input.AuthorizationConfirmedAt)
	patch.IsActive = input.IsActive
	resp, err := a.personnel.CreateEmergencyContact(rpcCtx, &personnel.CreateEmergencyContactRequest{Input: patch})
	if err != nil {
		return nil, mapHRMWriteError(err, "emergency-contact")
	}
	return toModelEmergencyContact(resp.GetItem())
}

// UpdateEmergencyContact memperbarui emergency-contact berdasarkan FieldMask.
func (a *Adapter) UpdateEmergencyContact(ctx context.Context, rawID string, input model.UpdateEmergencyContactInput) (*model.EmergencyContact, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_emergency_contacts", "update", "human_resource.emergency_contacts.update")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "emergency-contact", "emergency-contact")
	if err != nil {
		return nil, err
	}
	if a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	patch := &personnel.EmergencyContactInput{}
	mask := &fieldmaskpb.FieldMask{}
	if err := applyStringMask(mask, "worker_id", input.WorkerID, &patch.WorkerId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "contact_party_reference_id", input.ContactPartyReferenceID, &patch.ContactPartyReferenceId, true); err != nil {
		return nil, err
	}
	if err := applyStringMask(mask, "relationship_code", input.RelationshipCode, &patch.RelationshipCode, true); err != nil {
		return nil, err
	}
	if err := applyInt32Mask(mask, "priority_order", input.PriorityOrder, &patch.PriorityOrder, true); err != nil {
		return nil, err
	}
	if err := applyTimestampMask(mask, "authorization_confirmed_at", input.AuthorizationConfirmedAt, &patch.AuthorizationConfirmedAt, true); err != nil {
		return nil, err
	}
	if err := applyBoolMask(mask, "is_active", input.IsActive, &patch.IsActive, true); err != nil {
		return nil, err
	}
	if len(mask.Paths) == 0 {
		return nil, middleware.ErrInvalidArgument("emergency-contact update is empty", nil)
	}
	resp, err := a.personnel.UpdateEmergencyContact(rpcCtx, &personnel.UpdateEmergencyContactRequest{Id: id, Input: patch, UpdateMask: mask})
	if err != nil {
		return nil, mapHRMWriteError(err, "emergency-contact")
	}
	return toModelEmergencyContact(resp.GetItem())
}

// DeleteEmergencyContact menghapus emergency-contact berdasarkan ID.
func (a *Adapter) DeleteEmergencyContact(ctx context.Context, rawID string) (*bool, error) {
	if a == nil || a.personnel == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	return deleteByUUID(ctx, a, rawID, "hr_emergency_contacts", "human_resource.emergency_contacts.delete", "emergency-contact", func(c context.Context, req *common.IDRequest) (*common.EmptyResponse, error) {
		return a.personnel.DeleteEmergencyContact(c, req)
	})
}
