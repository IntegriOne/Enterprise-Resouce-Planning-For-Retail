package outbox

import (
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
)

func outboxToModel(e *entity.OutboxEvent) (outboxEventModel, error) {
	if e == nil {
		return outboxEventModel{}, eferrors.WithField(eferrors.ErrInvalidInput, "outbox_event", "outbox event must not be nil")
	}
	id, err := pgstore.ValidateUUIDParam("id", e.ID)
	if err != nil {
		return outboxEventModel{}, err
	}
	aggID, err := pgstore.ValidateUUIDParam("aggregate_id", e.AggregateID)
	if err != nil {
		return outboxEventModel{}, err
	}
	eventType, err := entity.ValidateRequiredText(
		e.EventType,
		entity.MaxOutboxEventTypeLength,
		eferrors.WithField(eferrors.ErrInvalidInput, "event_type", "event type must be non-empty and within length bounds"),
	)
	if err != nil {
		return outboxEventModel{}, err
	}
	aggType, err := entity.ValidateRequiredText(
		e.AggregateType,
		entity.MaxOutboxAggregateTypeLength,
		eferrors.WithField(eferrors.ErrInvalidInput, "aggregate_type", "aggregate type must be non-empty and within length bounds"),
	)
	if err != nil {
		return outboxEventModel{}, err
	}
	if !e.Status.IsValid() {
		return outboxEventModel{}, eferrors.ErrInvalidOutboxStatus
	}
	if e.AttemptCount < 0 {
		return outboxEventModel{}, eferrors.WithField(eferrors.ErrInvalidInput, "attempt_count", "attempt count cannot be negative")
	}
	payload, err := entity.ValidateJSONObject("payload", e.Payload)
	if err != nil {
		return outboxEventModel{}, err
	}

	return outboxEventModel{
		ID:            id,
		EventType:     eventType,
		AggregateType: aggType,
		AggregateID:   aggID,
		Payload:       string(payload),
		Status:        e.Status.String(),
		AttemptCount:  e.AttemptCount,
		NextRetryAt:   pgstore.MustLocalNow(e.NextRetryAt),
		LastError:     pgstore.CloneStringPtr(e.LastError),
		ProcessedAt:   pgstore.NormalizeLocalTimePtr(e.ProcessedAt),
		CreatedAt:     pgstore.MustLocalNow(e.CreatedAt),
		UpdatedAt:     pgstore.MustLocalNow(e.UpdatedAt),
	}, nil
}

func outboxToDomain(m *outboxEventModel) (*entity.OutboxEvent, error) {
	if m == nil {
		return nil, eferrors.ErrOutboxEventNotFound
	}
	id, err := pgstore.ValidateUUIDParam("id", m.ID)
	if err != nil {
		return nil, err
	}
	aggID, err := pgstore.ValidateUUIDParam("aggregate_id", m.AggregateID)
	if err != nil {
		return nil, err
	}
	status, err := entity.ParseOutboxStatus(m.Status)
	if err != nil {
		return nil, err
	}
	payload, err := entity.ValidateJSONObject("payload", []byte(m.Payload))
	if err != nil {
		return nil, err
	}

	return &entity.OutboxEvent{
		ID:            id,
		EventType:     m.EventType,
		AggregateType: m.AggregateType,
		AggregateID:   aggID,
		Payload:       payload,
		Status:        status,
		AttemptCount:  m.AttemptCount,
		NextRetryAt:   pgstore.NormalizeLocalTime(m.NextRetryAt),
		LastError:     pgstore.CloneStringPtr(m.LastError),
		ProcessedAt:   pgstore.NormalizeLocalTimePtr(m.ProcessedAt),
		CreatedAt:     pgstore.NormalizeLocalTime(m.CreatedAt),
		UpdatedAt:     pgstore.NormalizeLocalTime(m.UpdatedAt),
	}, nil
}
