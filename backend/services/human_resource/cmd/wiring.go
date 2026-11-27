package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	stdgrpc "google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/integrione/erp-retail/backend/services/human_resource/config"
	hrdb "github.com/integrione/erp-retail/backend/services/human_resource/db"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	efintegration "github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/integration/enterprise_foundation"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/outbox"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
	appservice "github.com/integrione/erp-retail/backend/services/human_resource/internal/service"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

// Dependencies berisi hasil wiring eksplisit dari satu database HRM ke seluruh service.
type Dependencies struct {
	DB                   *gorm.DB
	Repositories         *repository.Registry
	Services             *appservice.Registry
	OutboxRelay          *outbox.Relay
	EnterpriseFoundation *efintegration.Client
	pool                 *sql.DB
	efConn               *stdgrpc.ClientConn
}

// BuildDependencies mewajibkan adapter keamanan dan membangun validator integrity dari repository HRM.
// Resource tanpa validator tetap ditolak tanpa fallback allow-all.
func BuildDependencies(
	ctx context.Context,
	cfg config.Config,
	policy operation.Policy,
	logger *slog.Logger,
	serviceOptions ...appservice.RegistryOption,
) (*Dependencies, error) {
	return buildDependencies(ctx, cfg, policy, logger, nil, serviceOptions...)
}

func buildDependencies(
	ctx context.Context,
	cfg config.Config,
	policy operation.Policy,
	logger *slog.Logger,
	termination outbox.Publisher,
	serviceOptions ...appservice.RegistryOption,
) (*Dependencies, error) {
	if err := operation.RequireDependency(policy); err != nil {
		return nil, fmt.Errorf("policy dependency: %w", err)
	}
	var (
		efConn *stdgrpc.ClientConn
		err    error
	)
	efClient := efintegration.NewDisabledClient()
	if cfg.EnterpriseFoundation.Enabled {
		efConn, efClient, err = efintegration.Dial(
			cfg.EnterpriseFoundation.Target,
			cfg.EnterpriseFoundation.HMACSecret,
			cfg.EnterpriseFoundation.Timeout,
		)
		if err != nil {
			return nil, fmt.Errorf("initialize enterprise foundation client: %w", err)
		}
	}
	handle, err := hrdb.Init(ctx, cfg)
	if err != nil {
		if efConn != nil {
			_ = efConn.Close()
		}
		return nil, err
	}
	pool, err := handle.DB()
	if err != nil {
		if efConn != nil {
			_ = efConn.Close()
		}
		return nil, fmt.Errorf("%w: cannot obtain initialized database pool", hrerrors.ErrUnavailable)
	}
	fail := func(cause error) (*Dependencies, error) {
		var closeErrs []error
		if err := pool.Close(); err != nil {
			closeErrs = append(closeErrs, fmt.Errorf("%w: close failed during startup", hrerrors.ErrUnavailable))
		}
		if efConn != nil {
			if err := efConn.Close(); err != nil {
				closeErrs = append(closeErrs, err)
			}
		}
		if len(closeErrs) > 0 {
			return nil, errors.Join(append([]error{cause}, closeErrs...)...)
		}
		return nil, cause
	}
	repos, err := repository.New(handle)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM repositories: %w", err))
	}
	validators, err := buildMutationValidators(repos, efClient)
	if err != nil {
		return fail(err)
	}
	integrity := security.NewIntegrity(validators)
	services, err := appservice.NewRegistry(repos, operation.Dependencies{Policy: policy, Integrity: integrity}, serviceOptions...)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM services: %w", err))
	}
	activeGeneral := cfg.Outbox.Mode != "" && cfg.Outbox.Mode != "standby"
	deliveryStore, err := outbox.NewStoreForRoutes(handle, activeGeneral, termination != nil)
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
	if termination != nil {
		publisher = outbox.RouteWorkerTermination(publisher, termination)
	}
	relay, err := outbox.NewRelay(repos.Transactor, deliveryStore, publisher, outbox.Config{
		Standby:      !activeGeneral && termination == nil,
		PollInterval: cfg.Outbox.PollInterval, PublishTimeout: cfg.Outbox.PublishTimeout,
		MaxAttempts: cfg.Outbox.MaxAttempts,
	}, logger)
	if err != nil {
		return fail(fmt.Errorf("initialize HRM outbox relay: %w", err))
	}
	return &Dependencies{
		DB: handle, Repositories: repos, Services: services, OutboxRelay: relay,
		EnterpriseFoundation: efClient, pool: pool, efConn: efConn,
	}, nil
}

// Close melepaskan pool saat startup gagal atau proses berhenti.
func (d *Dependencies) Close() error {
	if d == nil {
		return nil
	}
	var errs []error
	if d.efConn != nil {
		if err := d.efConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if d.pool != nil {
		if err := d.pool.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%w: cannot close HRM database", hrerrors.ErrUnavailable))
		}
	}
	return errors.Join(errs...)
}

type validatorBuilder struct {
	name  string
	build func() (map[string]security.MutationValidator, error)
}

func buildMutationValidators(
	repos *repository.Registry,
	efClient *efintegration.Client,
) (map[string]security.MutationValidator, error) {
	builders := []validatorBuilder{
		{"workforce planning", func() (map[string]security.MutationValidator, error) {
			return security.WorkforcePlanningValidators(repos)
		}},
		{"workforce plan", func() (map[string]security.MutationValidator, error) {
			return security.WorkforcePlanValidators(repos, efClient)
		}},
		{"personnel", func() (map[string]security.MutationValidator, error) { return security.PersonnelValidators(repos) }},
		{"personnel record", func() (map[string]security.MutationValidator, error) {
			return security.PersonnelRecordValidators(repos, efClient)
		}},
		{"position", func() (map[string]security.MutationValidator, error) {
			return security.PositionValidators(repos, efClient)
		}},
		{"worker", func() (map[string]security.MutationValidator, error) {
			return security.WorkerValidators(repos, efClient)
		}},
		{"competency", func() (map[string]security.MutationValidator, error) { return security.CompetencyValidators(repos) }},
		{"learning", func() (map[string]security.MutationValidator, error) { return security.LearningValidators(repos) }},
		{"benefits", func() (map[string]security.MutationValidator, error) { return security.BenefitsValidators(repos) }},
		{"compensation", func() (map[string]security.MutationValidator, error) { return security.CompensationValidators(repos) }},
		{"pay component", func() (map[string]security.MutationValidator, error) { return security.PayComponentValidators(repos) }},
		{"payroll", func() (map[string]security.MutationValidator, error) {
			return security.PayrollValidators(repos, efClient)
		}},
		{"analytics", func() (map[string]security.MutationValidator, error) { return security.AnalyticsValidators(repos) }},
		{"analytics snapshot", func() (map[string]security.MutationValidator, error) {
			return security.AnalyticsSnapshotValidators(repos, efClient)
		}},
		{"career", func() (map[string]security.MutationValidator, error) { return security.CareerValidators(repos) }},
		{"performance", func() (map[string]security.MutationValidator, error) { return security.PerformanceValidators(repos) }},
		{"talent dimension", func() (map[string]security.MutationValidator, error) {
			return security.TalentDimensionValidators(repos)
		}},
		{"talent", func() (map[string]security.MutationValidator, error) {
			return security.TalentValidators(repos, efClient)
		}},
		{"succession", func() (map[string]security.MutationValidator, error) { return security.SuccessionValidators(repos) }},
		{"leave", func() (map[string]security.MutationValidator, error) { return security.LeaveValidators(repos) }},
		{"onboarding", func() (map[string]security.MutationValidator, error) { return security.OnboardingValidators(repos) }},
		{"offboarding", func() (map[string]security.MutationValidator, error) { return security.OffboardingValidators(repos) }},
		{"recruitment candidate", func() (map[string]security.MutationValidator, error) {
			return security.RecruitmentCandidateValidators(repos, efClient)
		}},
		{"recruitment requisition", func() (map[string]security.MutationValidator, error) {
			return security.RecruitmentRequisitionValidators(repos, efClient)
		}},
		{"recruitment posting", func() (map[string]security.MutationValidator, error) {
			return security.RecruitmentJobPostingValidators(repos)
		}},
		{"recruitment application", func() (map[string]security.MutationValidator, error) {
			return security.RecruitmentApplicationValidators(repos)
		}},
		{"recruitment interview", func() (map[string]security.MutationValidator, error) {
			return security.RecruitmentInterviewValidators(repos)
		}},
		{"recruitment detail", func() (map[string]security.MutationValidator, error) {
			return security.RecruitmentDetailValidators(repos, efClient)
		}},
		{"promotion", func() (map[string]security.MutationValidator, error) {
			return security.PromotionValidators(repos, efClient)
		}},
		{"employee relations", func() (map[string]security.MutationValidator, error) {
			return security.EmployeeRelationsValidators(repos, efClient)
		}},
		{"policy", func() (map[string]security.MutationValidator, error) {
			return security.PolicyValidators(repos, efClient)
		}},
		{"hr service", func() (map[string]security.MutationValidator, error) { return security.HRServiceValidators(repos) }},
		{"engagement", func() (map[string]security.MutationValidator, error) {
			return security.EngagementValidators(repos, efClient)
		}},
		{"access request", func() (map[string]security.MutationValidator, error) { return security.AccessRequestValidators(repos) }},
	}
	validators := map[string]security.MutationValidator{}
	for _, entry := range builders {
		registered, err := entry.build()
		if err != nil {
			return nil, fmt.Errorf("initialize HRM %s validators: %w", entry.name, err)
		}
		for resource, validator := range registered {
			if _, exists := validators[resource]; exists {
				return nil, fmt.Errorf("duplicate HRM validator registration for %s", resource)
			}
			validators[resource] = validator
		}
	}
	return validators, nil
}
