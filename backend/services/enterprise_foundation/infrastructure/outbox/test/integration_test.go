package outbox_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	efdb "github.com/integrione/erp-retail/backend/services/enterprise_foundation/db"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/infrastructure/outbox"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/pkg/uuid"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository/pgstore"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services/legalentity"
	"gorm.io/gorm"
)

func resolveLiveTestDSN(t *testing.T) string {
	t.Helper()
	if testDSN := strings.TrimSpace(os.Getenv("EF_TEST_DATABASE_DSN")); testDSN != "" {
		return testDSN
	}
	if devDSN := strings.TrimSpace(os.Getenv("EF_DATABASE_DSN")); devDSN != "" {
		derived, err := efdb.DeriveTestDatabaseDSN(devDSN)
		if err != nil {
			t.Fatalf("failed to derive test dsn: %v", err)
		}
		return derived
	}
	t.Skip("skipping live PostgreSQL outbox integration test: set EF_TEST_DATABASE_DSN or EF_DATABASE_DSN")
	return ""
}

func truncateOutboxTestTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(`TRUNCATE TABLE
		foundation_outbox_events,
		party_addresses,
		party_contacts,
		party_roles,
		parties,
		organization_units,
		locations,
		legal_entities,
		addresses
		RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("truncate tables failed: %v", err)
	}
}

func TestLivePostgreSQLOutboxRelayEndToEnd(t *testing.T) {
	dsn := resolveLiveTestDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := efdb.Init(ctx, dsn, true)
	if err != nil {
		t.Fatalf("db.Init failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() failed: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	truncateOutboxTestTables(t, db)
	t.Cleanup(func() { truncateOutboxTestTables(t, db) })

	repos, err := repository.New(db)
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}

	var (
		mu        sync.Mutex
		received  []outbox.Envelope
		notifyGot = make(chan struct{}, 4)
	)
	webhookSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var env outbox.Envelope
		if unmarshalErr := json.Unmarshal(body, &env); unmarshalErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mu.Lock()
		received = append(received, env)
		mu.Unlock()
		select {
		case notifyGot <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer webhookSrv.Close()

	pub, err := outbox.NewWebhookPublisher(outbox.WebhookConfig{
		Endpoint:   webhookSrv.URL,
		HMACSecret: []byte("0123456789abcdef0123456789abcdef"),
	})
	if err != nil {
		t.Fatalf("NewWebhookPublisher failed: %v", err)
	}

	relay, err := outbox.NewRelay(
		outbox.Dependencies{
			Transactor: repos.Transactor,
			Repo:       repos.Outbox,
			Publisher:  pub,
			Logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
		outbox.Config{
			PollInterval:    10 * time.Second,
			BatchSize:       10,
			MaxAttempts:     5,
			RetentionPeriod: 24 * time.Hour,
		},
	)
	if err != nil {
		t.Fatalf("outbox.NewRelay failed: %v", err)
	}

	// Bungkus OutboxRepository dengan notifier pasca-commit agar service langsung membangunkan relay.
	svcRepos := *repos
	svcRepos.Outbox = outbox.WrapNotifyingRepository(repos.Outbox, relay, pgstore.RegisterAfterCommit)
	svcReg, err := services.NewRegistry(&svcRepos, services.Options{})
	if err != nil {
		t.Fatalf("services.NewRegistry failed: %v", err)
	}

	runCtx, cancelRun := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() {
		runDone <- relay.Run(runCtx)
	}()
	defer func() {
		cancelRun()
		<-runDone
	}()

	le, err := svcReg.LegalEntities.Create(ctx, legalentity.CreateInput{
		EntityCode:               "OUTBOX-PT-01",
		LegalName:                "PT Integrione Outbox Test",
		LegalForm:                ptrString("PT"),
		IsPrimaryOperatingEntity: true,
		Status:                   entity.LegalEntityStatusActive,
		JurisdictionCountryCode:  "ID",
		FiscalYearStartMonth:     1,
	})
	if err != nil {
		t.Fatalf("LegalEntities.Create failed: %v", err)
	}

	select {
	case <-notifyGot:
	case <-time.After(3 * time.Second):
		t.Fatal("expected post-commit wakeup delivery within 3s")
	}

	// Pastikan status baris di PostgreSQL sudah effective.
	var lastPage domainrepo.PageResult[*entity.OutboxEvent]
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		lastPage, err = repos.Outbox.SelectByAggregate(ctx, entity.AggregateTypeLegalEntity, le.ID, domainrepo.PageRequest{Limit: 10})
		if err == nil && len(lastPage.Items) == 1 && lastPage.Items[0].Status == entity.OutboxStatusEffective {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if len(lastPage.Items) != 1 || lastPage.Items[0].Status != entity.OutboxStatusEffective {
		t.Fatalf("expected outbox row effective in PG, got %+v (err=%v)", lastPage.Items, err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 || received[0].AggregateID != le.ID || received[0].EventType != entity.EventTypeLegalEntityCreated {
		t.Fatalf("unexpected webhook envelope received: %+v", received)
	}
}

func TestLivePostgreSQLConcurrentSkipLockedSingleDelivery(t *testing.T) {
	dsn := resolveLiveTestDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := efdb.Init(ctx, dsn, true)
	if err != nil {
		t.Fatalf("db.Init failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() failed: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	truncateOutboxTestTables(t, db)
	t.Cleanup(func() { truncateOutboxTestTables(t, db) })

	repos, err := repository.New(db)
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}

	const totalEvents = 6
	gen := uuid.NewGenerator()
	now := time.Now()
	for i := 0; i < totalEvents; i++ {
		id, err := gen.NewUUID()
		if err != nil {
			t.Fatalf("NewUUID failed: %v", err)
		}
		aggID, err := gen.NewUUID()
		if err != nil {
			t.Fatalf("NewUUID failed: %v", err)
		}
		ev, err := entity.NewOutboxEvent(id, entity.EventTypePartyCreated, entity.AggregateTypeParty, aggID, map[string]any{"idx": i}, now)
		if err != nil {
			t.Fatalf("NewOutboxEvent failed: %v", err)
		}
		if err := repos.Outbox.Insert(ctx, ev); err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	var mu sync.Mutex
	deliveredIDs := make(map[string]int)
	pub := outbox.PublishFunc(func(_ context.Context, ev *entity.OutboxEvent) error {
		time.Sleep(15 * time.Millisecond)
		mu.Lock()
		deliveredIDs[ev.ID]++
		mu.Unlock()
		return nil
	})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	r1, err := outbox.NewRelay(outbox.Dependencies{Transactor: repos.Transactor, Repo: repos.Outbox, Publisher: pub, Logger: logger}, outbox.Config{BatchSize: 2})
	if err != nil {
		t.Fatalf("NewRelay 1 failed: %v", err)
	}
	r2, err := outbox.NewRelay(outbox.Dependencies{Transactor: repos.Transactor, Repo: repos.Outbox, Publisher: pub, Logger: logger}, outbox.Config{BatchSize: 2})
	if err != nil {
		t.Fatalf("NewRelay 2 failed: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = r1.DrainPending(ctx)
	}()
	go func() {
		defer wg.Done()
		_, _ = r2.DrainPending(ctx)
	}()
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(deliveredIDs) != totalEvents {
		t.Fatalf("expected %d unique events delivered, got %d (%v)", totalEvents, len(deliveredIDs), deliveredIDs)
	}
	for id, count := range deliveredIDs {
		if count != 1 {
			t.Fatalf("event %s delivered %d times across concurrent SKIP LOCKED relays", id, count)
		}
	}
}

func ptrString(s string) *string {
	return &s
}
