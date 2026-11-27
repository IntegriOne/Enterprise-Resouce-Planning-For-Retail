// Package outbox mengimplementasikan repositori PostgreSQL berbasis GORM untuk tabel foundation_outbox_events.
package outbox

import "time"

type outboxEventModel struct {
	ID            string     `gorm:"column:id;type:uuid;primaryKey"`
	EventType     string     `gorm:"column:event_type"`
	AggregateType string     `gorm:"column:aggregate_type"`
	AggregateID   string     `gorm:"column:aggregate_id;type:uuid"`
	Payload       string     `gorm:"column:payload;type:jsonb"`
	Status        string     `gorm:"column:status"`
	AttemptCount  int        `gorm:"column:attempt_count"`
	NextRetryAt   time.Time  `gorm:"column:next_retry_at"`
	LastError     *string    `gorm:"column:last_error"`
	ProcessedAt   *time.Time `gorm:"column:processed_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (outboxEventModel) TableName() string {
	return "foundation_outbox_events"
}
