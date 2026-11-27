package service

import (
	"context"
	"slices"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type PayslipListCheck struct {
	WorkerIDs                    []string
	PrincipalID, PayrollPeriodID string
	Limit, Offset                int
}

// PublishedPayslipLister menerapkan kepemilikan, penerbitan, dan kesiapan DRM sebelum paginasi.
type PublishedPayslipLister interface {
	ListPublishedPayslips(context.Context, PayslipListCheck) (operation.Page[PublishedPayslip], error)
}

// ListMyPayslips membaca halaman slip terbit dan memeriksa kembali referensi HRM.
func (s *Service) ListMyPayslips(ctx context.Context, periodID string, limit, offset int) (operation.Page[PublishedPayslip], error) {
	var empty operation.Page[PublishedPayslip]
	if s == nil || operation.RequireDependency(s.dependencies.Workers) != nil {
		return empty, hrerrors.ErrUnavailable
	}
	if periodID != "" {
		if err := entity.ValidateUUID("payroll_period_id", periodID); err != nil {
			return empty, err
		}
	}
	if limit == 0 {
		limit = domainrepo.DefaultLimit
	}
	if limit < 0 || limit > domainrepo.MaxLimit || offset < 0 {
		return empty, hrerrors.ErrInvalidInput
	}
	actor, workerIDs, err := s.payslipOwner(ctx)
	if err != nil {
		return empty, err
	}
	lister, ok := s.dependencies.PayslipVerifier.(PublishedPayslipLister)
	if !ok || operation.RequireDependency(lister) != nil {
		return empty, hrerrors.ErrUnavailable
	}
	page, err := lister.ListPublishedPayslips(ctx, PayslipListCheck{
		WorkerIDs: slices.Clone(workerIDs), PrincipalID: actor.PrincipalID, PayrollPeriodID: periodID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return empty, operation.OwnerVerificationError(ctx, err)
	}
	expected := min(int64(limit), max(0, page.Total-int64(offset)))
	if page.Total < 0 || int64(len(page.Items)) != expected {
		return empty, hrerrors.ErrReferenceConflict
	}
	if err := s.verifyPayslipPage(ctx, page.Items, workerIDs, periodID); err != nil {
		return empty, err
	}
	checked, currentWorkerIDs, err := s.payslipOwner(ctx)
	if err != nil {
		return empty, err
	}
	if checked != actor || !slices.Equal(workerIDs, currentWorkerIDs) {
		return empty, hrerrors.ErrForbidden
	}
	page.Limit, page.Offset = limit, offset
	return page, nil
}

func (s *Service) verifyPayslipPage(ctx context.Context, items []PublishedPayslip, workerIDs []string, periodID string) error {
	if len(items) == 0 {
		return nil
	}
	now := s.dependencies.Runtime.Clock.Now()
	ids := make([]any, 0, len(items))
	for _, item := range items {
		if entity.ValidateUUID("payroll_result_id", item.ResultID) != nil || !validPublishedPayslip(item, item.ResultID, now) {
			return hrerrors.ErrReferenceConflict
		}
		ids = append(ids, item.ResultID)
	}
	rows, err := s.selectOwnedPayslipResults(ctx, ids, workerIDs)
	if err != nil {
		return err
	}
	if len(rows) != len(items) {
		return hrerrors.ErrReferenceConflict
	}
	indexed := make(map[string]entity.PayrollResult, len(rows))
	for _, row := range rows {
		indexed[strings.ToLower(row.ID)] = row
	}
	ordered := make([]entity.PayrollResult, len(items))
	for i, item := range items {
		key := strings.ToLower(item.ResultID)
		row, ok := indexed[key]
		if !ok || !slices.Contains(workerIDs, strings.ToLower(row.WorkerID)) || !matchesPayslipDocument(item, row) {
			return hrerrors.ErrReferenceConflict
		}
		delete(indexed, key)
		ordered[i] = row
	}
	return s.setPayslipPagePeriods(ctx, items, ordered, periodID)
}

// setPayslipPagePeriods memuat run dan periode dengan dua query per halaman.
func (s *Service) setPayslipPagePeriods(ctx context.Context, items []PublishedPayslip, results []entity.PayrollResult, periodID string) error {
	runIDs := make([]any, 0, len(results))
	seenRuns := make(map[string]bool, len(results))
	for _, result := range results {
		id := strings.ToLower(result.PayrollRunID)
		if !seenRuns[id] {
			runIDs = append(runIDs, id)
			seenRuns[id] = true
		}
	}
	runs, err := s.dependencies.PayrollRun.SelectWhere(ctx, []domainrepo.Filter{{Column: "id", Values: runIDs}}, len(runIDs), 0)
	if err != nil {
		return err
	}
	periodIDs := make([]any, 0, len(runs))
	runIndex := make(map[string]entity.PayrollRun, len(runs))
	seenPeriods := make(map[string]bool, len(runs))
	for _, run := range runs {
		id := strings.ToLower(run.PayrollPeriodID)
		runIndex[strings.ToLower(run.ID)] = run
		if !seenPeriods[id] {
			periodIDs = append(periodIDs, id)
			seenPeriods[id] = true
		}
	}
	if len(runIndex) != len(runIDs) {
		return hrerrors.ErrReferenceConflict
	}
	periods, err := s.dependencies.PayrollPeriod.SelectWhere(ctx, []domainrepo.Filter{{Column: "id", Values: periodIDs}}, len(periodIDs), 0)
	if err != nil {
		return err
	}
	indexed := make(map[string]entity.PayrollPeriod, len(periods))
	for _, period := range periods {
		indexed[strings.ToLower(period.ID)] = period
	}
	for i, result := range results {
		run, ok := runIndex[strings.ToLower(result.PayrollRunID)]
		if !ok || !matchesPayslipRun(items[i], &run, periodID) {
			return hrerrors.ErrReferenceConflict
		}
		period, ok := indexed[strings.ToLower(run.PayrollPeriodID)]
		if !ok || !applyPayslipPeriod(&items[i], period) {
			return hrerrors.ErrReferenceConflict
		}
	}
	return nil
}
