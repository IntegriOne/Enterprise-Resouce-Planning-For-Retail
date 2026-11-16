package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/human_resource/config"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/outbox"
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
	OutboxRelay  *outbox.Relay
	pool         *sql.DB
}

// BuildDependencies mewajibkan adapter keamanan dan membangun validator integrity dari repository HRM.
// Resource tanpa validator tetap ditolak tanpa fallback allow-all.
func BuildDependencies(
	ctx context.Context,
	cfg config.Config,
	policy operation.Policy,
	logger *slog.Logger,
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
	personnelValidators, err := security.PersonnelValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM personnel validators: %w", err))
	}
	for resource, validator := range personnelValidators {
		validators[resource] = validator
	}
	positionValidators, err := security.PositionValidators(repos, nil)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM position validators: %w", err))
	}
	for resource, validator := range positionValidators {
		validators[resource] = validator
	}
	competencyValidators, err := security.CompetencyValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM competency validators: %w", err))
	}
	for resource, validator := range competencyValidators {
		validators[resource] = validator
	}
	learningValidators, err := security.LearningValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM learning validators: %w", err))
	}
	for resource, validator := range learningValidators {
		validators[resource] = validator
	}
	benefitValidators, err := security.BenefitsValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM benefits validators: %w", err))
	}
	for resource, validator := range benefitValidators {
		validators[resource] = validator
	}
	compensationValidators, err := security.CompensationValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM compensation validators: %w", err))
	}
	for resource, validator := range compensationValidators {
		validators[resource] = validator
	}
	payComponentValidators, err := security.PayComponentValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM pay component validators: %w", err))
	}
	for resource, validator := range payComponentValidators {
		validators[resource] = validator
	}
	analyticsValidators, err := security.AnalyticsValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM analytics validators: %w", err))
	}
	for resource, validator := range analyticsValidators {
		validators[resource] = validator
	}
	careerValidators, err := security.CareerValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM career validators: %w", err))
	}
	for resource, validator := range careerValidators {
		validators[resource] = validator
	}
	performanceValidators, err := security.PerformanceValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM performance validators: %w", err))
	}
	for resource, validator := range performanceValidators {
		validators[resource] = validator
	}
	talentDimensionValidators, err := security.TalentDimensionValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM talent dimension validators: %w", err))
	}
	for resource, validator := range talentDimensionValidators {
		validators[resource] = validator
	}
	successionValidators, err := security.SuccessionValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM succession validators: %w", err))
	}
	for resource, validator := range successionValidators {
		validators[resource] = validator
	}
	leaveValidators, err := security.LeaveValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM leave validators: %w", err))
	}
	for resource, validator := range leaveValidators {
		validators[resource] = validator
	}
	onboardingValidators, err := security.OnboardingValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM onboarding validators: %w", err))
	}
	for resource, validator := range onboardingValidators {
		validators[resource] = validator
	}
	offboardingValidators, err := security.OffboardingValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM offboarding validators: %w", err))
	}
	for resource, validator := range offboardingValidators {
		validators[resource] = validator
	}
	recruitmentCandidateValidators, err := security.RecruitmentCandidateValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM recruitment candidate validators: %w", err))
	}
	for resource, validator := range recruitmentCandidateValidators {
		validators[resource] = validator
	}
	requisitionValidators, err := security.RecruitmentRequisitionValidators(repos, nil)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM recruitment requisition validators: %w", err))
	}
	for resource, validator := range requisitionValidators {
		validators[resource] = validator
	}
	recruitmentApplicationValidators, err := security.RecruitmentApplicationValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM recruitment application validators: %w", err))
	}
	for resource, validator := range recruitmentApplicationValidators {
		validators[resource] = validator
	}
	recruitmentInterviewValidators, err := security.RecruitmentInterviewValidators(repos)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM recruitment interview validators: %w", err))
	}
	for resource, validator := range recruitmentInterviewValidators {
		validators[resource] = validator
	}
	integrity := security.NewIntegrity(validators)
	services, err := appservice.NewRegistry(repos, operation.Dependencies{Policy: policy, Integrity: integrity})
	if err != nil {
		return fail(fmt.Errorf("initialize HRM services: %w", err))
	}
	deliveryStore, err := outbox.NewStore(handle)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM outbox store: %w", err))
	}
	var publisher outbox.Publisher
	switch cfg.Outbox.Mode {
	case "", "standby":
	case "log":
		publisher = outbox.NewLogPublisher(logger)
	case "webhook":
		publisher, err = outbox.NewWebhookPublisher(cfg.Outbox.WebhookURL, cfg.Outbox.WebhookHMACSecret, cfg.Outbox.PublishTimeout)
		if err != nil {
			return fail(fmt.Errorf("initialize HRM outbox publisher: %w", err))
		}
	default:
		return fail(hrerrors.ErrInvalidInput)
	}
	relay, err := outbox.NewRelay(repos.Transactor, deliveryStore, publisher, outbox.Config{
		Standby:      cfg.Outbox.Mode == "" || cfg.Outbox.Mode == "standby",
		PollInterval: cfg.Outbox.PollInterval, PublishTimeout: cfg.Outbox.PublishTimeout,
		MaxAttempts: cfg.Outbox.MaxAttempts,
	}, logger)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM outbox relay: %w", err))
	}
	return &Dependencies{DB: handle, Repositories: repos, Services: services, OutboxRelay: relay, pool: pool}, nil
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
