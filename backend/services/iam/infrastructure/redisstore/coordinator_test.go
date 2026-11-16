package redisstore_test

import (
	"context"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/infrastructure/redisstore"
)

func TestCoordinatorRecordRejectsMissingDependencies(t *testing.T) {
	t.Parallel()
	var missing *redisstore.Coordinator
	if _, err := missing.RecordInTx(context.Background(), "event", "user", "id", redisstore.OutboxPayload{}); err == nil {
		t.Fatal("expected error when outbox coordinator is nil")
	}
	coordinator := redisstore.NewCoordinator(nil, nil, nil, nil, nil)
	if _, err := coordinator.RecordInTx(context.Background(), "event", "user", "id", redisstore.OutboxPayload{}); err == nil {
		t.Fatal("expected error when outbox dependencies are missing")
	}
}
