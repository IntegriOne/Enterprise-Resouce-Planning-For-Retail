package service_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	efservice "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/service"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type fakeIDGen struct {
	id  string
	err error
}

func (g fakeIDGen) NewUUID() (string, error) {
	if g.err != nil {
		return "", g.err
	}
	return g.id, nil
}

func TestOutboxBuilder_Build(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 11, 15, 14, 0, 0, 0, time.UTC)

	t.Run("requires non-nil id generator", func(t *testing.T) {
		t.Parallel()
		_, err := efservice.NewOutboxBuilder(nil, fixedClock{now: fixedNow})
		if !errors.Is(err, eferrors.ErrInvalidInput) {
			t.Fatalf("NewOutboxBuilder(nil) error = %v, want ErrInvalidInput", err)
		}
	})

	t.Run("builds deterministic outbox event with fixed clock", func(t *testing.T) {
		t.Parallel()
		builder, err := efservice.NewOutboxBuilder(fakeIDGen{id: svcUUID1}, fixedClock{now: fixedNow})
		if err != nil {
			t.Fatalf("NewOutboxBuilder failed: %v", err)
		}

		payload := efservice.PartyRolePayload{
			PartyID:     svcUUID2,
			PartnerCode: "pt-kain",
			Role:        entity.PartyRoleKindVendor,
			IsActive:    true,
		}
		ev, err := builder.Build(entity.EventTypePartyRoleChanged, entity.AggregateTypePartyRole, svcUUID3, payload)
		if err != nil {
			t.Fatalf("Build() unexpected error: %v", err)
		}
		if ev.ID != svcUUID1 || ev.AggregateID != svcUUID3 || ev.Status != entity.OutboxStatusPending {
			t.Errorf("unexpected outbox event: %+v", ev)
		}
		if !ev.CreatedAt.Equal(fixedNow) {
			t.Errorf("CreatedAt = %v, want %v", ev.CreatedAt, fixedNow)
		}

		var decoded efservice.PartyRolePayload
		if err := json.Unmarshal(ev.Payload, &decoded); err != nil {
			t.Fatalf("json.Unmarshal(Payload) failed: %v", err)
		}
		if decoded != payload {
			t.Errorf("decoded payload = %+v, want %+v", decoded, payload)
		}
	})

	t.Run("propagates id generation error as KindUnavailable", func(t *testing.T) {
		t.Parallel()
		genErr := errors.New("entropy exhausted")
		builder, err := efservice.NewOutboxBuilder(fakeIDGen{err: genErr}, nil)
		if err != nil {
			t.Fatalf("NewOutboxBuilder(nil clock) should default to SystemClock, got %v", err)
		}
		_, err = builder.Build(entity.EventTypePartyCreated, entity.AggregateTypeParty, svcUUID2, nil)
		if !errors.Is(err, genErr) || eferrors.KindOf(err) != eferrors.KindUnavailable {
			t.Errorf("Build error = (%v, kind=%v), want wrapping genErr with KindUnavailable", err, eferrors.KindOf(err))
		}
	})

	t.Run("nil builder fails closed", func(t *testing.T) {
		t.Parallel()
		var nilBuilder *efservice.OutboxBuilder
		_, err := nilBuilder.Build(entity.EventTypePartyCreated, entity.AggregateTypeParty, svcUUID2, nil)
		if !errors.Is(err, eferrors.ErrInvalidInput) {
			t.Errorf("nil Build error = %v, want ErrInvalidInput", err)
		}
	})
}
