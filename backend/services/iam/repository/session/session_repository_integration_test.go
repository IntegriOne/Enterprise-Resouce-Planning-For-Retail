//go:build integration

package session_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/integrione/erp-retail/backend/services/iam/db"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/session"
)

func TestSessionUpdatesRejectExpiredRows(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("IAM_TEST_DATABASE_DSN"))
	if os.Getenv("IAM_TEST_DATABASE_DISPOSABLE") != "1" || dsn == "" {
		t.Skip("disposable IAM test database is required")
	}
	parsed, err := pgx.ParseConfig(dsn)
	if err != nil || !strings.HasPrefix(parsed.Database, "iam_disposable_") || !strings.HasSuffix(parsed.Database, "_test") {
		t.Fatal("IAM_TEST_DATABASE_DSN must target a disposable IAM database")
	}
	ctx := context.Background()
	conn, err := db.Init(ctx, dsn, true)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := conn.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	repo, err := session.NewRepository(conn)
	if err != nil {
		t.Fatal(err)
	}

	const userID = "11111111-1111-4111-8111-111111111111"
	if err := conn.Exec("INSERT INTO users (id, display_name) VALUES (?, ?)", userID, "Session audit").Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().Truncate(time.Second)
	cases := []struct {
		name     string
		id       string
		idle     time.Time
		absolute time.Time
		want     error
	}{
		{"active", "22222222-2222-4222-8222-222222222221", now.Add(time.Minute), now.Add(time.Hour), nil},
		{"idle expired", "22222222-2222-4222-8222-222222222222", now, now.Add(time.Hour), iamerrors.ErrSessionIdleExpired},
		{"absolute expired", "22222222-2222-4222-8222-222222222223", now, now, iamerrors.ErrSessionAbsoluteExpired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := conn.Exec(`INSERT INTO sessions
				(id, user_id, token_digest, authenticated_at, last_seen_at, idle_expires_at, absolute_expires_at)
				VALUES (?, ?, ?, ?, ?, ?, ?)`, tc.id, userID, []byte(tc.id), now.Add(-time.Hour), now.Add(-time.Minute), tc.idle, tc.absolute).Error; err != nil {
				t.Fatal(err)
			}
			err := repo.UpdateActivity(ctx, tc.id, now, now.Add(5*time.Minute))
			if !errors.Is(err, tc.want) {
				t.Fatalf("UpdateActivity error = %v, want %v", err, tc.want)
			}
			err = repo.UpdateMFAVerified(ctx, tc.id, now, []byte("rotated-csrf"))
			if !errors.Is(err, tc.want) {
				t.Fatalf("UpdateMFAVerified error = %v, want %v", err, tc.want)
			}
			var state struct {
				MFAVerified bool      `gorm:"column:mfa_verified"`
				IdleExpires time.Time `gorm:"column:idle_expires_at"`
			}
			if err := conn.Table("sessions").Select("mfa_verified", "idle_expires_at").Where("id = ?", tc.id).Take(&state).Error; err != nil {
				t.Fatal(err)
			}
			if tc.want == nil && !state.MFAVerified {
				t.Fatal("active session MFA update was not persisted")
			}
			if tc.want == nil {
				if err := repo.UpdateActivity(ctx, tc.id, now.Add(-time.Second), now.Add(time.Minute)); err != nil {
					t.Fatalf("older activity update failed: %v", err)
				}
				var later struct {
					LastSeenAt    time.Time `gorm:"column:last_seen_at"`
					IdleExpiresAt time.Time `gorm:"column:idle_expires_at"`
				}
				if err := conn.Table("sessions").Select("last_seen_at", "idle_expires_at").Where("id = ?", tc.id).Take(&later).Error; err != nil {
					t.Fatal(err)
				}
				if later.LastSeenAt.Before(now) || later.IdleExpiresAt.Before(now.Add(5*time.Minute)) {
					t.Fatal("older request moved session activity backward")
				}
			}
			if tc.want != nil && (state.MFAVerified || !state.IdleExpires.Equal(tc.idle)) {
				t.Fatal("expired session was modified")
			}
		})
	}
}
