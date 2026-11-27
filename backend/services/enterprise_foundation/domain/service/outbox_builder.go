package service

import (
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

// EntityStatusPayload membawa payload ringkas perubahan status entitas untuk sinkronisasi lintas modul.
type EntityStatusPayload struct {
	EntityID      string `json:"entity_id"`
	LegalEntityID string `json:"legal_entity_id,omitempty"`
	Code          string `json:"code"`
	Status        string `json:"status,omitempty"`
	IsActive      bool   `json:"is_active"`
}

// PartyRolePayload membawa payload perubahan peran bisnis mitra (supplier, vendor, customer_b2b, dll).
type PartyRolePayload struct {
	PartyID     string               `json:"party_id"`
	PartnerCode string               `json:"partner_code"`
	Role        entity.PartyRoleKind `json:"role"`
	IsActive    bool                 `json:"is_active"`
}

// PartyContactPortalPayload membawa payload perubahan tautan portal PIC mitra ke principal IAM.
type PartyContactPortalPayload struct {
	ContactID          string                    `json:"contact_id"`
	PartyID            string                    `json:"party_id"`
	IAMUserID          *string                   `json:"iam_user_id,omitempty"`
	PortalAccessStatus entity.PortalAccessStatus `json:"portal_access_status"`
	IsActive           bool                      `json:"is_active"`
}

// OutboxBuilder merakit entri OutboxEvent baru secara deterministik menggunakan IDGenerator dan Clock.
type OutboxBuilder struct {
	idGen IDGenerator
	clock Clock
}

// NewOutboxBuilder membentuk OutboxBuilder dengan validasi fail-closed atas IDGenerator.
func NewOutboxBuilder(idGen IDGenerator, clock Clock) (*OutboxBuilder, error) {
	if idGen == nil {
		return nil, eferrors.WithField(eferrors.ErrInvalidInput, "id_gen", "id generator is required")
	}
	if clock == nil {
		clock = SystemClock{}
	}
	return &OutboxBuilder{
		idGen: idGen,
		clock: clock,
	}, nil
}

// Build membentuk entity.OutboxEvent berstatus pending siap ditulis pada transaksi yang sama dengan mutasi agregat.
func (b *OutboxBuilder) Build(eventType, aggregateType, aggregateID string, payload any) (*entity.OutboxEvent, error) {
	if b == nil || b.idGen == nil {
		return nil, eferrors.WithField(eferrors.ErrInvalidInput, "outbox_builder", "outbox builder is not initialized")
	}
	eventID, err := b.idGen.NewUUID()
	if err != nil {
		return nil, eferrors.Wrap(err, eferrors.KindUnavailable, "enterprise_foundation.outbox.id_generation_failed", "failed to generate outbox event id")
	}
	return entity.NewOutboxEvent(eventID, eventType, aggregateType, aggregateID, payload, b.clock.Now())
}
