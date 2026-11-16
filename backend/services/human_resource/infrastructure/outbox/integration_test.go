//go:build integration

package outbox

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository/pgstore"
)

func TestDeliveryStoreWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("HRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("HRM_TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Scheme != "postgres" || !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatal("integration test requires a dedicated *_test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open HRM test database failed")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := hrdb.RunMigrationsUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	event := validEvent()
	if err := db.WithContext(ctx).Exec("DELETE FROM hr_outbox_events WHERE id = ?", event.ID).Error; err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Exec("DELETE FROM hr_outbox_events WHERE id = ?", event.ID).Error }()
	if err := db.WithContext(ctx).Create(event).Error; err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pgstore.NewTransactor(db)
	if err != nil {
		t.Fatal(err)
	}
	locked := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- tx.WithinTransaction(ctx, func(tc context.Context) error {
			items, err := store.SelectDue(tc, time.Now().UTC().Add(time.Second), 1, 20)
			if err != nil {
				return err
			}
			if len(items) != 1 || items[0].ID != event.ID {
				return fmt.Errorf("first claimant got %d events", len(items))
			}
			close(locked)
			select {
			case <-release:
			case <-tc.Done():
				return tc.Err()
			}
			return store.MarkPublished(tc, event.ID, time.Now().UTC(), 1)
		})
	}()
	select {
	case <-locked:
	case err := <-firstDone:
		t.Fatalf("first claimant failed: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	secondErr := tx.WithinTransaction(ctx, func(tc context.Context) error {
		items, err := store.SelectDue(tc, time.Now().UTC().Add(time.Second), 1, 20)
		if err != nil {
			return err
		}
		if len(items) != 0 {
			return fmt.Errorf("second claimant received %d locked events", len(items))
		}
		return nil
	})
	close(release)
	firstErr := <-firstDone
	if secondErr != nil {
		t.Fatal(secondErr)
	}
	if firstErr != nil {
		t.Fatal(firstErr)
	}
	var published time.Time
	if err := db.WithContext(ctx).Raw("SELECT published_at FROM hr_outbox_events WHERE id = ?", event.ID).Scan(&published).Error; err != nil {
		t.Fatal(err)
	}
	if published.IsZero() {
		t.Fatal("published event was not persisted")
	}
}
