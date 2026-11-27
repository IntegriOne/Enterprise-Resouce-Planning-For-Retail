package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	payrollpolicy "github.com/integrione/erp-retail/backend/services/human_resource/domain/service/payroll"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// CalculationRequest mengikat kalkulator ke snapshot input dan aturan yang dipilih saat run dibuat.
type CalculationRequest struct {
	RunID, PeriodID, GroupID, RuleSetID, InputSnapshotID    string
	EmployerID, CurrencyCode, JurisdictionCode              string
	EngineName, EngineVersion, RunType, SettlementDirection string
	OriginalRunID, RuleArtifactVersionID                    string
	InputDigest, RuleArtifactDigest                         []byte
	PeriodStart, PeriodEnd, PayDate                         time.Time
	CurrencyScale                                           int16
}

// CalculationResult memuat hasil kalkulator yang terikat pada snapshot dan aturan yang sama.
type CalculationResult struct {
	RunID, RuleSetID, InputSnapshotID             string
	InputDigest, RuleArtifactDigest, OutputDigest []byte
	Results                                       []entity.PayrollResult
	Lines                                         []entity.PayrollResultLine
}

// Calculator menghasilkan snapshot hasil tanpa melakukan perubahan database.
type Calculator interface {
	Calculate(context.Context, CalculationRequest) (CalculationResult, error)
}

// CalculatePayrollRun menyimpan hasil, komponen, digest, dan outbox dalam satu transaksi.
func (s *Service) CalculatePayrollRun(ctx context.Context, id string) (*entity.PayrollRun, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.payrollRun.Command(ctx, id, "calculate", func(txCtx context.Context, run *entity.PayrollRun, actor operation.Actor, now time.Time) (*entity.PayrollRun, error) {
		if actor.Kind != "human" || !strings.EqualFold(run.StartedByPrincipalID, actor.PrincipalID) {
			return nil, hrerrors.ErrForbidden
		}
		if run.Status != entity.PayrollRunStatusDraft || run.CalculatedAt != nil || len(run.OutputDigest) != 0 ||
			run.ApprovedAt != nil || run.ApprovedByPrincipalID != nil || now.Before(run.CreatedAt) ||
			len(run.InputDigest) != sha256.Size {
			return nil, hrerrors.ErrWorkflowRequired
		}
		if operation.RequireDependency(s.dependencies.Calculator) != nil {
			return nil, hrerrors.ErrUnavailable
		}
		period, err := s.dependencies.PayrollPeriod.FindByIDForUpdate(txCtx, run.PayrollPeriodID)
		if err != nil {
			return nil, err
		}
		if period == nil || period.PeriodStatus != "open" {
			return nil, hrerrors.ErrWorkflowRequired
		}
		group, err := s.dependencies.PayrollGroup.FindByIDForUpdate(txCtx, run.PayrollGroupID)
		if err != nil {
			return nil, err
		}
		if group == nil || !group.IsActive {
			return nil, hrerrors.ErrReferenceConflict
		}
		ruleSet, err := s.dependencies.PayrollRuleSet.FindByIDForUpdate(txCtx, run.RuleSetID)
		if err != nil {
			return nil, err
		}
		if ruleSet == nil || len(ruleSet.ArtifactDigest) != sha256.Size ||
			ruleSet.ArtifactDocumentVersionReferenceID == "" ||
			ruleSet.ApprovedAt.IsZero() || ruleSet.ApprovedAt.After(now) ||
			ruleSet.JurisdictionCode != run.JurisdictionCode ||
			entity.CompareDates(period.PeriodStart, ruleSet.EffectiveFrom) < 0 ||
			(ruleSet.EffectiveTo != nil && entity.CompareDates(period.PeriodEnd, *ruleSet.EffectiveTo) > 0) {
			return nil, hrerrors.ErrReferenceConflict
		}
		count, err := s.dependencies.PayrollResult.CountWhere(txCtx, []domainrepo.Filter{
			{Column: "payroll_run_id", Values: []any{run.ID}},
		})
		if err != nil {
			return nil, err
		}
		if count != 0 {
			return nil, hrerrors.ErrConflict
		}
		// Validasikan draf terhadap referensi terkini sebelum memuat input dan menjalankan mesin.
		if err := s.payrollRun.CheckIntegrity(txCtx, run, run); err != nil {
			return nil, err
		}
		request := CalculationRequest{
			RunID: run.ID, PeriodID: run.PayrollPeriodID, GroupID: run.PayrollGroupID, RuleSetID: run.RuleSetID,
			InputSnapshotID: run.InputSnapshotReferenceID, EmployerID: run.EmployingLegalEntityReferenceID,
			CurrencyCode: run.CurrencyCode, JurisdictionCode: run.JurisdictionCode,
			EngineName: run.RuleEngineName, EngineVersion: run.RuleEngineVersion,
			PeriodStart: period.PeriodStart, PeriodEnd: period.PeriodEnd, PayDate: period.PayDate,
			CurrencyScale:         group.CurrencyScale,
			RuleArtifactVersionID: ruleSet.ArtifactDocumentVersionReferenceID, RuleArtifactDigest: bytes.Clone(ruleSet.ArtifactDigest),
			RunType: run.RunType, SettlementDirection: run.SettlementDirection, InputDigest: bytes.Clone(run.InputDigest),
		}
		if run.OriginalRunID != nil {
			request.OriginalRunID = *run.OriginalRunID
		}
		output, err := s.dependencies.Calculator.Calculate(txCtx, request)
		if err != nil {
			return nil, operation.OwnerVerificationError(txCtx, err)
		}
		if !strings.EqualFold(output.RunID, run.ID) || !strings.EqualFold(output.RuleSetID, run.RuleSetID) ||
			!strings.EqualFold(output.InputSnapshotID, run.InputSnapshotReferenceID) ||
			!bytes.Equal(output.InputDigest, run.InputDigest) || !bytes.Equal(output.RuleArtifactDigest, ruleSet.ArtifactDigest) ||
			len(output.OutputDigest) != sha256.Size ||
			len(output.Results) > 10000 || len(output.Lines) > 10000 {
			return nil, hrerrors.ErrReferenceConflict
		}
		run.Status, run.CalculatedAt, run.OutputDigest = entity.PayrollRunStatusCalculated, &now, bytes.Clone(output.OutputDigest)
		for i := range output.Results {
			result := &output.Results[i]
			if result.ResultStatus != "calculated" || result.PaymentReferenceID != nil ||
				result.PayslipDocumentReferenceID != nil || result.PayslipPublicationReferenceID != nil ||
				len(result.CalculationDigest) != sha256.Size {
				return nil, hrerrors.ErrReferenceConflict
			}
			result.CreatedAt = now
		}
		if err := payrollpolicy.ValidateRunResults(run, output.Results, output.Lines); err != nil {
			return nil, err
		}
		for _, result := range output.Results {
			if !validSettlementPrecision(result.NetPayAmount, group.CurrencyScale) {
				return nil, hrerrors.ErrReferenceConflict
			}
		}
		if err := s.validateCalculationReferences(txCtx, run, period, output); err != nil {
			return nil, err
		}
		for i := range output.Results {
			if err := s.dependencies.PayrollResult.Insert(txCtx, &output.Results[i]); err != nil {
				return nil, err
			}
		}
		for i := range output.Lines {
			if err := s.dependencies.PayrollResultLine.Insert(txCtx, &output.Lines[i]); err != nil {
				return nil, err
			}
		}
		return run, nil
	})
}

// validateCalculationReferences memeriksa katalog payroll sebelum hasil dapat tersimpan.
func (s *Service) validateCalculationReferences(ctx context.Context, run *entity.PayrollRun, period *entity.PayrollPeriod, output CalculationResult) error {
	for _, result := range output.Results {
		profile, err := s.dependencies.PayrollProfile.FindByIDForUpdate(ctx, result.PayrollProfileReferenceID)
		if err != nil {
			return err
		}
		if profile == nil || !profile.IsActive ||
			!strings.EqualFold(profile.WorkerID, result.WorkerID) ||
			!strings.EqualFold(profile.PayrollGroupID, run.PayrollGroupID) ||
			!strings.EqualFold(profile.EmployingLegalEntityReferenceID, run.EmployingLegalEntityReferenceID) ||
			profile.CurrencyCode != run.CurrencyCode || profile.PayrollJurisdictionCode != run.JurisdictionCode ||
			!overlapsPayrollPeriod(profile.EffectiveFrom, profile.EffectiveTo, period) {
			return hrerrors.ErrReferenceConflict
		}
	}
	components := make(map[string]*entity.PayComponent)
	for _, line := range output.Lines {
		key := strings.ToLower(line.ComponentID)
		component := components[key]
		if component == nil {
			var err error
			component, err = s.dependencies.PayComponent.FindByIDForUpdate(ctx, line.ComponentID)
			if err != nil {
				return err
			}
			if component == nil || !component.IsActive || component.JurisdictionCode != run.JurisdictionCode ||
				(component.CurrencyCode != nil && *component.CurrencyCode != run.CurrencyCode) ||
				(component.RuleSetID != nil && !strings.EqualFold(*component.RuleSetID, run.RuleSetID)) ||
				!overlapsPayrollPeriod(component.EffectiveFrom, component.EffectiveTo, period) {
				return hrerrors.ErrReferenceConflict
			}
			components[key] = component
		}
		if component.ComponentCode != line.ComponentCodeSnapshot ||
			component.ComponentName != line.ComponentNameSnapshot ||
			component.ComponentType != line.ComponentTypeSnapshot {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}

func overlapsPayrollPeriod(from time.Time, to *time.Time, period *entity.PayrollPeriod) bool {
	return entity.CompareDates(from, period.PeriodEnd) <= 0 &&
		(to == nil || entity.CompareDates(*to, period.PeriodStart) >= 0)
}
