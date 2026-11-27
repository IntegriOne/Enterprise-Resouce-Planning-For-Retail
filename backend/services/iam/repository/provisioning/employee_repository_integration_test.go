//go:build integration

package provisioning_test

import (
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/integrione/erp-retail/backend/services/iam/db"
	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/repository/provisioning"
)

func TestEmployeeProvisioningReceiptAndConcurrentClaim(t *testing.T) {
	dsn := os.Getenv("IAM_TEST_DATABASE_DSN")
	if dsn == "" || os.Getenv("IAM_TEST_DATABASE_DISPOSABLE") != "1" {
		t.Skip("disposable IAM test database is required")
	}
	config, err := pgx.ParseConfig(dsn)
	if err != nil || !strings.HasPrefix(config.Database, "iam_disposable_") || !strings.HasSuffix(config.Database, "_test") {
		t.Fatal("IAM_TEST_DATABASE_DSN must target a disposable IAM database")
	}
	ctx := t.Context()
	conn, err := db.Init(ctx, dsn, true)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := conn.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	repo, err := provisioning.NewEmployeeRepository(conn)
	if err != nil {
		t.Fatal(err)
	}

	const operator = "10000000-0000-4000-8000-000000000001"
	if err := conn.Exec("INSERT INTO users (id, user_kind, display_name) VALUES (?, 'employee', ?)", operator, "Operator fixture").Error; err != nil {
		t.Fatal(err)
	}
	if err := conn.Exec("DROP TABLE employee_provisioning_requests").Error; err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := db.RunMigrationsUp(ctx, pool); err != nil {
			t.Fatalf("reconcile installed IAM baseline: %v", err)
		}
	}
	var versions int64
	if err := conn.Table("schema_migrations").Count(&versions).Error; err != nil || versions != 1 {
		t.Fatalf("migration versions=%d error=%v", versions, err)
	}
	var count int64
	if err := conn.Table("users").Where("id = ?", operator).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("existing IAM users=%d error=%v", count, err)
	}
	request := entity.EmployeeProvisioningRequest{
		HRMRequestID:           "10000000-0000-4000-8000-000000000002",
		WorkerID:               "10000000-0000-4000-8000-000000000003",
		LegalEntityID:          "10000000-0000-4000-8000-000000000004",
		RequestedByPrincipalID: operator,
		WorkerNumber:           "EMP-023",
		CatalogRoleID:          "10000000-0000-4000-8000-000000000006",
		RequestedAt:            time.Now(),
	}
	receipt, err := repo.InsertReceipt(ctx, request)
	if err != nil || receipt.Status != entity.EmployeeProvisioningPending {
		t.Fatalf("receipt=%+v err=%v", receipt, err)
	}
	repeated, err := repo.InsertReceipt(ctx, request)
	if err != nil || repeated.ID != receipt.ID {
		t.Fatalf("repeat=%+v err=%v", repeated, err)
	}
	duplicate := request
	duplicate.HRMRequestID = "10000000-0000-4000-8000-000000000005"
	if _, err := repo.InsertReceipt(ctx, duplicate); !errors.Is(err, iamerrors.ErrAlreadyExists) {
		t.Fatalf("second open request for worker: %v", err)
	}
	mismatch := request
	mismatch.LegalEntityID = duplicate.HRMRequestID
	if _, err := repo.InsertReceipt(ctx, mismatch); !errors.Is(err, iamerrors.ErrConflict) {
		t.Fatalf("changed payload on replay: %v", err)
	}

	start := make(chan struct{})
	outcomes := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Go(func() {
			<-start
			_, err := repo.UpdateClaim(ctx, receipt.ID, operator, time.Now())
			outcomes <- err
		})
	}
	close(start)
	workers.Wait()
	close(outcomes)
	var successes, conflicts int
	for err := range outcomes {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, iamerrors.ErrConflict):
			conflicts++
		default:
			t.Fatalf("unexpected claim result: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	claimed, err := repo.SelectByID(ctx, receipt.ID)
	if err != nil || claimed.Status != entity.EmployeeProvisioningProcessing || claimed.ProcessedByUserID == nil || *claimed.ProcessedByUserID != operator {
		t.Fatalf("claimed=%+v err=%v", claimed, err)
	}
}
