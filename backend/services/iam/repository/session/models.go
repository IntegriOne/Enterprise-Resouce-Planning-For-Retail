// Package session mengimplementasikan repositori PostgreSQL GORM untuk siklus hidup sesi login browser pada tabel sessions.
package session

import "time"

type sessionModel struct {
	ID                string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID            string     `gorm:"column:user_id;type:uuid"`
	TokenDigest       []byte     `gorm:"column:token_digest"`
	CSRFTokenDigest   []byte     `gorm:"column:csrf_token_digest"`
	Generation        int64      `gorm:"column:generation"`
	CSRFKeyVersion    int        `gorm:"column:csrf_key_version"`
	IPAddress         *string    `gorm:"column:ip_address;type:inet"`
	UserAgent         *string    `gorm:"column:user_agent"`
	AuthenticatedAt   time.Time  `gorm:"column:authenticated_at"`
	LastSeenAt        time.Time  `gorm:"column:last_seen_at"`
	IdleExpiresAt     time.Time  `gorm:"column:idle_expires_at"`
	AbsoluteExpiresAt time.Time  `gorm:"column:absolute_expires_at"`
	MFAVerified       bool       `gorm:"column:mfa_verified"`
	MFAVerifiedAt     *time.Time `gorm:"column:mfa_verified_at"`
	EndedAt           *time.Time `gorm:"column:ended_at"`
	EndReason         *string    `gorm:"column:end_reason"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
}

func (sessionModel) TableName() string {
	return "sessions"
}
