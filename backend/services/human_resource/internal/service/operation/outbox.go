package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"math"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// emit menyimpan payload minimum. Unique aggregate-version menjadi pagar konflik konkuren.
// Retensi outbox tidak boleh menghapus sumber version terakhir sebelum strategi version persisten tersedia.
func (r *Resource[T, K, P]) emit(ctx context.Context, action string, value *T, actor Actor) error {
	aggregateID := r.spec.AggregateID(value)
	if err := entity.ValidateUUID("aggregate_id", aggregateID); err != nil {
		return err
	}
	var latest int64
	for offset := 0; ; offset += 100 {
		if offset >= 10000 {
			return hrerrors.ErrConflict
		}
		events, err := r.deps.Outbox.FindByAggregateID(ctx, aggregateID, 100, offset)
		if err != nil {
			return err
		}
		for _, event := range events {
			if event.AggregateType == r.spec.Table && event.AggregateVersion > latest {
				latest = event.AggregateVersion
			}
		}
		if len(events) < 100 {
			break
		}
	}
	if latest == math.MaxInt64 {
		return hrerrors.ErrConflict
	}
	id, err := r.deps.IDs.NewID()
	if err != nil {
		return err
	}
	envelope := map[string]any{
		"aggregate_id":       aggregateID,
		"record_key":         r.spec.KeyString(r.spec.Key(value)),
		"change_kind":        action,
		"actor_principal_id": actor.PrincipalID,
	}
	if r.spec.EventDetails != nil {
		if details := r.spec.EventDetails(action, value); len(details) != 0 {
			envelope["details"] = details
		}
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(payload)
	now := r.deps.Clock.Now()
	event := &entity.OutboxEvent{
		ID:               id,
		AggregateType:    r.spec.Table,
		AggregateID:      aggregateID,
		AggregateVersion: latest + 1,
		EventType:        r.spec.Table + "." + action,
		SchemaVersion:    1,
		Payload:          entity.JSON(payload),
		PayloadDigest:    digest[:],
		OccurredAt:       now,
		AvailableAt:      now,
	}
	if err := event.Validate(); err != nil {
		return err
	}
	return r.deps.Outbox.Insert(ctx, event)
}
