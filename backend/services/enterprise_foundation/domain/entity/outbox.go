package entity

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// Konstanta nama agregat dan jenis event outbox Enterprise Foundation untuk sinkronisasi lintas layanan.
const (
	AggregateTypeLegalEntity      = "legal_entity"
	AggregateTypeOrganizationUnit = "organization_unit"
	AggregateTypeLocation         = "location"
	AggregateTypeParty            = "party"
	AggregateTypePartyRole        = "party_role"
	AggregateTypePartyContact     = "party_contact"

	EventTypeLegalEntityCreated        = "legal_entity.created"
	EventTypeLegalEntityUpdated        = "legal_entity.updated"
	EventTypeLegalEntityStatusChanged  = "legal_entity.status_changed"
	EventTypeOrgUnitCreated            = "organization_unit.created"
	EventTypeOrgUnitUpdated            = "organization_unit.updated"
	EventTypeOrgUnitStatusChanged      = "organization_unit.status_changed"
	EventTypeLocationCreated           = "location.created"
	EventTypeLocationUpdated           = "location.updated"
	EventTypeLocationStatusChanged     = "location.status_changed"
	EventTypePartyCreated              = "party.created"
	EventTypePartyUpdated              = "party.updated"
	EventTypePartyStatusChanged        = "party.status_changed"
	EventTypePartyRoleChanged          = "party.role_changed"
	EventTypePartyContactPortalChanged = "party.contact_portal_changed"
)

// OutboxEvent merepresentasikan entri antrean transaksional pada tabel foundation_outbox_events.
type OutboxEvent struct {
	ID            string
	EventType     string
	AggregateType string
	AggregateID   string
	Payload       []byte
	Status        OutboxStatus
	AttemptCount  int
	NextRetryAt   time.Time
	LastError     *string
	ProcessedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewOutboxEvent membangun entri event outbox baru berstatus pending dengan payload JSON tervalidasi.
func NewOutboxEvent(
	id, eventType, aggregateType, aggregateID string,
	payloadObj any,
	now time.Time,
) (*OutboxEvent, error) {
	if err := ValidateUUID("id", id); err != nil {
		return nil, err
	}
	if err := ValidateUUID("aggregate_id", aggregateID); err != nil {
		return nil, err
	}
	trimmedEventType, err := ValidateRequiredText(
		eventType,
		MaxOutboxEventTypeLength,
		eferrors.NewValidation("event_type", eferrors.Code("enterprise_foundation.outbox.invalid_event_type"), "event type must be non-empty and within length bounds"),
	)
	if err != nil {
		return nil, err
	}
	trimmedAggType, err := ValidateRequiredText(
		aggregateType,
		MaxOutboxAggregateTypeLength,
		eferrors.NewValidation("aggregate_type", eferrors.Code("enterprise_foundation.outbox.invalid_aggregate_type"), "aggregate type must be non-empty and within length bounds"),
	)
	if err != nil {
		return nil, err
	}

	raw := []byte("{}")
	if payloadObj != nil {
		switch v := payloadObj.(type) {
		case []byte:
			validated, err := ValidateJSONObject("payload", v)
			if err != nil {
				return nil, err
			}
			raw = validated
		case json.RawMessage:
			validated, err := ValidateJSONObject("payload", v)
			if err != nil {
				return nil, err
			}
			raw = validated
		default:
			encoded, err := json.Marshal(payloadObj)
			if err != nil {
				return nil, eferrors.Wrap(err, eferrors.KindInvalidInput, "enterprise_foundation.outbox.invalid_payload", "failed to encode outbox payload")
			}
			validated, err := ValidateJSONObject("payload", encoded)
			if err != nil {
				return nil, err
			}
			raw = validated
		}
	}

	ts := normalizeLocalTime(now)
	return &OutboxEvent{
		ID:            NormalizeUUID(id),
		EventType:     trimmedEventType,
		AggregateType: trimmedAggType,
		AggregateID:   NormalizeUUID(aggregateID),
		Payload:       raw,
		Status:        OutboxStatusPending,
		AttemptCount:  0,
		NextRetryAt:   ts,
		CreatedAt:     ts,
		UpdatedAt:     ts,
	}, nil
}

// MarkEffective menandai event outbox berhasil dikirim pada waktu processedAt.
func (e *OutboxEvent) MarkEffective(processedAt time.Time) error {
	if e == nil {
		return eferrors.ErrOutboxEventNotFound
	}
	ts := normalizeLocalTime(processedAt)
	e.Status = OutboxStatusEffective
	e.LastError = nil
	e.ProcessedAt = &ts
	e.UpdatedAt = ts
	return nil
}

// ScheduleRetry menambah penghitung percobaan dan menjadwalkan ulang pengiriman event.
func (e *OutboxEvent) ScheduleRetry(nextRetryAt time.Time, errMsg string, now time.Time) error {
	if e == nil {
		return eferrors.ErrOutboxEventNotFound
	}
	ts := normalizeLocalTime(now)
	retryAt := normalizeLocalTime(nextRetryAt)
	if retryAt.Before(ts) {
		retryAt = ts
	}
	e.Status = OutboxStatusPending
	e.AttemptCount++
	e.NextRetryAt = retryAt
	e.LastError = normalizeOutboxError(errMsg)
	e.UpdatedAt = ts
	return nil
}

// MarkFailed menandai event outbox gagal permanen setelah melampaui batas percobaan atau kesalahan fatal.
func (e *OutboxEvent) MarkFailed(errMsg string, now time.Time) error {
	if e == nil {
		return eferrors.ErrOutboxEventNotFound
	}
	ts := normalizeLocalTime(now)
	e.Status = OutboxStatusFailed
	e.AttemptCount++
	e.LastError = normalizeOutboxError(errMsg)
	e.UpdatedAt = ts
	return nil
}

// PayloadCopy mengembalikan salinan slice payload JSON agar caller tidak mengubah slice internal.
func (e *OutboxEvent) PayloadCopy() []byte {
	if e == nil || len(e.Payload) == 0 {
		return []byte("{}")
	}
	return bytes.Clone(e.Payload)
}

// Clone mengembalikan salinan mendalam OutboxEvent.
func (e *OutboxEvent) Clone() *OutboxEvent {
	if e == nil {
		return nil
	}
	return &OutboxEvent{
		ID:            e.ID,
		EventType:     e.EventType,
		AggregateType: e.AggregateType,
		AggregateID:   e.AggregateID,
		Payload:       e.PayloadCopy(),
		Status:        e.Status,
		AttemptCount:  e.AttemptCount,
		NextRetryAt:   e.NextRetryAt,
		LastError:     cloneStringPtr(e.LastError),
		ProcessedAt:   cloneTimePtr(e.ProcessedAt),
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func normalizeOutboxError(errMsg string) *string {
	trimmed := strings.TrimSpace(errMsg)
	if trimmed == "" {
		return nil
	}
	// Batasi panjang pesan error outbox agar tetap wajar saat dipersistensikan.
	const maxOutboxErrRunes = 1000
	if utf8.RuneCountInString(trimmed) > maxOutboxErrRunes {
		runes := []rune(trimmed)
		trimmed = string(runes[:maxOutboxErrRunes])
	}
	return &trimmed
}
