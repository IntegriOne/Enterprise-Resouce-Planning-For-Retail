package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// ApproveCompensationRecord mengesahkan proposal tanpa mengaktifkan pembayaran apa pun.
func (s *Service) ApproveCompensationRecord(ctx context.Context, id string) (*entity.CompensationRecord, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.compensationRecord.Command(ctx, id, "approve", func(txCtx context.Context, record *entity.CompensationRecord, actor operation.Actor, now time.Time) (*entity.CompensationRecord, error) {
		if actor.Kind != "human" {
			return nil, hrerrors.ErrForbidden
		}
		if record.Status != entity.CompensationStatusProposed || record.ApprovedAt != nil || record.ApprovedByPrincipalID != nil ||
			(record.EffectiveTo != nil && entity.CompareDates(now.UTC(), *record.EffectiveTo) > 0) {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if err := s.verifyIndependentApprover(txCtx, record.ID, actor.PrincipalID, now); err != nil {
			return nil, err
		}
		record.Status = entity.CompensationStatusApproved
		record.ApprovedAt = &now
		record.ApprovedByPrincipalID = &actor.PrincipalID
		return record, nil
	})
}

func (s *Service) verifyIndependentApprover(ctx context.Context, id, principalID string, now time.Time) error {
	var created bool
	var count, latest int64
	for offset := 0; ; offset += domainrepo.MaxLimit {
		if offset >= 10000 {
			return hrerrors.ErrConflict
		}
		events, err := s.outbox.FindByAggregateID(ctx, id, domainrepo.MaxLimit, offset)
		if err != nil {
			return err
		}
		for _, event := range events {
			if event.AggregateType != "hr_compensation_records" {
				continue
			}
			switch event.EventType {
			case "hr_compensation_records.created":
				if created || event.AggregateVersion != 1 {
					return hrerrors.ErrReferenceConflict
				}
				created = true
			case "hr_compensation_records.updated":
				if event.AggregateVersion <= 1 {
					return hrerrors.ErrReferenceConflict
				}
			default:
				return hrerrors.ErrReferenceConflict
			}
			var payload struct {
				RecordKey string `json:"record_key"`
				ActorID   string `json:"actor_principal_id"`
			}
			if event.SchemaVersion != 1 || event.OccurredAt.IsZero() || event.OccurredAt.After(now) ||
				json.Unmarshal(event.Payload, &payload) != nil || !strings.EqualFold(payload.RecordKey, id) ||
				entity.ValidateUUID("actor_principal_id", payload.ActorID) != nil {
				return hrerrors.ErrReferenceConflict
			}
			if strings.EqualFold(payload.ActorID, principalID) {
				return hrerrors.ErrForbidden
			}
			count++
			latest = max(latest, event.AggregateVersion)
		}
		if len(events) < domainrepo.MaxLimit {
			break
		}
	}
	if !created || count != latest {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}
