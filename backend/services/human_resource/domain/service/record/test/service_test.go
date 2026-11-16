package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	service "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/record"
)

type mockDB struct {
	mu      sync.Mutex
	writes  int
	failure error
}

func (m *mockDB) Insert(ctx context.Context, value *entity.JobFamily) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failure != nil {
		return m.failure
	}
	m.writes++
	return nil
}
func (m *mockDB) Update(ctx context.Context, value *entity.JobFamily) error {
	return m.Insert(ctx, value)
}

func TestWriterRejectsInvalidBeforePersistence(t *testing.T) {
	port := &mockDB{}
	writer, err := service.NewWriter[entity.JobFamily](port)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	invalid := entity.JobFamily{}
	if err := writer.Insert(ctx, &invalid); !errors.Is(err, hrerrors.ErrInvalidInput) {
		t.Fatalf("invalid input: %v", err)
	}
	if port.writes != 0 {
		t.Fatal("invalid entity reached repository")
	}
	valid := entity.JobFamily{ID: "00000000-0000-0000-0000-000000000001", FamilyCode: "RETAIL", FamilyName: "Retail", JobDomain: "retail"}
	if err := writer.Insert(ctx, &valid); err != nil {
		t.Fatal(err)
	}
	cause := errors.New("storage failed")
	port.failure = cause
	if err := writer.Update(ctx, &valid); err != cause {
		t.Fatalf("repository error lost: %v", err)
	}
	if _, err := service.NewWriter[entity.JobFamily]((*mockDB)(nil)); !errors.Is(err, hrerrors.ErrUnavailable) {
		t.Fatalf("typed nil dependency: %v", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := writer.Insert(canceled, &valid); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}

func TestWriterConcurrentValidation(t *testing.T) {
	port := &mockDB{}
	writer, err := service.NewWriter[entity.JobFamily](port)
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	failures := make(chan error, 16)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			value := entity.JobFamily{ID: "00000000-0000-0000-0000-000000000001", FamilyCode: "RETAIL", FamilyName: "Retail", JobDomain: "retail"}
			failures <- writer.Create(context.Background(), &value)
		}()
	}
	close(start)
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	if port.writes != 16 {
		t.Fatalf("writes=%d", port.writes)
	}
}
