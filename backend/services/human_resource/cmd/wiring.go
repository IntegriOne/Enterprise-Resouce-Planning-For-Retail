package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/human_resource/config"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	appservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

// Dependencies berisi hasil wiring eksplisit dari satu database HRM ke seluruh service.
type Dependencies struct {
	DB           *gorm.DB
	Repositories *repository.Registry
	Services     *appservice.Registry
	pool         *sql.DB
}

// BuildDependencies mewajibkan adapter keamanan dan membangun validator integrity dari repository HRM.
// Resource tanpa validator tetap ditolak tanpa fallback allow-all.
func BuildDependencies(
	ctx context.Context,
	cfg config.Config,
	policy operation.Policy,
) (*Dependencies, error) {
	if err := operation.RequireDependency(policy); err != nil {
		return nil, fmt.Errorf("policy dependency: %w", err)
	}
	handle, err := hrdb.Init(ctx, cfg)
	if err != nil {
		return nil, err
	}
	pool, err := handle.DB()
	if err != nil {
		return nil, fmt.Errorf("%w: cannot obtain initialized database pool", hrerrors.ErrUnavailable)
	}
	fail := func(cause error) (*Dependencies, error) {
		if err := pool.Close(); err != nil {
			return nil, errors.Join(cause, fmt.Errorf("%w: close failed during startup", hrerrors.ErrUnavailable))
		}
		return nil, cause
	}
	repos, err := repository.New(handle)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM repositories: %w", err))
	}
	validators, err := security.WorkforcePlanningValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM integrity validators: %w", err))
	}
	integrity := security.NewIntegrity(validators)
	services, err := appservice.NewRegistry(repos, operation.Dependencies{Policy: policy, Integrity: integrity})
	if err != nil {
		return fail(fmt.Errorf("initialize HRM services: %w", err))
	}
	return &Dependencies{DB: handle, Repositories: repos, Services: services, pool: pool}, nil
}

// Close melepaskan pool saat startup gagal atau proses berhenti.
func (d *Dependencies) Close() error {
	if d == nil || d.pool == nil {
		return nil
	}
	if err := d.pool.Close(); err != nil {
		return fmt.Errorf("%w: cannot close HRM database", hrerrors.ErrUnavailable)
	}
	return nil
}
