package entity

import (
	"encoding/json"
	"strings"
	"time"
)

// OutboxStatus merepresentasikan siklus hidup sinkronisasi event keamanan ke Redis (pending/effective/failed).
type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusEffective OutboxStatus = "effective"
	OutboxStatusFailed    OutboxStatus = "failed"
)

// Jenis event outbox keamanan IAM yang wajib menjaga konsistensi Redis.
const (
	OutboxEventSessionIssued           = "session.issued"
	OutboxEventSessionUpdated          = "session.updated"
	OutboxEventSessionRevoked          = "session.revoked"
	OutboxEventUserSessionsRevoked     = "user.sessions_revoked"
	OutboxEventUserSecurityChanged     = "user.security_changed"
	OutboxEventUserPermissionsChanged  = "user.permissions_changed"
	OutboxEventRolePermissionsChanged  = "role.permissions_changed"
	OutboxEventGroupPermissionsChanged = "group.permissions_changed"
)

// OutboxEvent merepresentasikan entri antrean transaksional pada tabel iam_outbox_events.
type OutboxEvent struct {
	ID            string
	EventType     string
	AggregateType string
	AggregateID   string
	Payload       []byte
	Status        OutboxStatus
	AttemptCount  int
	NextRetryAt   time.Time
	LastError     string
	ProcessedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewOutboxEvent membangun event outbox keamanan baru berstatus pending.
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
	raw := []byte("{}")
	if payloadObj != nil {
		encoded, err := json.Marshal(payloadObj)
		if err != nil {
			return nil, err
		}
		raw = encoded
	}
	ts := normalizeLocalTime(now)
	return &OutboxEvent{
		ID:            strings.TrimSpace(id),
		EventType:     strings.TrimSpace(eventType),
		AggregateType: strings.TrimSpace(aggregateType),
		AggregateID:   strings.TrimSpace(aggregateID),
		Payload:       raw,
		Status:        OutboxStatusPending,
		AttemptCount:  0,
		NextRetryAt:   ts,
		CreatedAt:     ts,
		UpdatedAt:     ts,
	}, nil
}
