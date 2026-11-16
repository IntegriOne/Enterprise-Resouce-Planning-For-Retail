package security

import (
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type compensationBandReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.CompensationBand, error)
}

type compensationRecordCounter interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
}

type compensationBandValidator struct {
	bands   compensationBandReader
	records compensationRecordCounter
}

// CompensationValidators memasang aturan katalog band yang dimiliki HRM.
func CompensationValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.CompensationBand) != nil ||
		operation.RequireDependency(repos.CompensationRecord) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_compensation_bands": &compensationBandValidator{
			bands: repos.CompensationBand, records: repos.CompensationRecord,
		},
	}, nil
}

// ValidateMutation menjaga versi unik dan membekukan band setelah dipakai.
func (v *compensationBandValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_compensation_bands" || operation.RequireDependency(v.bands) != nil ||
		operation.RequireDependency(v.records) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.CompensationBand
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.CompensationBand)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
		if previous.ApprovalReferenceID != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return v.rejectIfUsed(ctx, previous.ID)
	}
	band, ok := after.(*entity.CompensationBand)
	if !ok || band == nil || (previous != nil && previous.ID != band.ID) {
		return hrerrors.ErrInvalidInput
	}
	if band.ApprovalReferenceID != nil {
		return hrerrors.ErrWorkflowRequired
	}
	matches, err := v.bands.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "band_code", Values: []any{band.BandCode}},
		{Column: "band_version", Values: []any{band.BandVersion}},
		{Column: "jurisdiction_code", Values: []any{band.JurisdictionCode}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != band.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	if previous != nil && !reflect.DeepEqual(previous, band) {
		return v.rejectIfUsed(ctx, band.ID)
	}
	return nil
}

func (v *compensationBandValidator) rejectIfUsed(ctx context.Context, bandID string) error {
	count, err := v.records.CountWhere(ctx, []domainrepo.Filter{{Column: "compensation_band_id", Values: []any{bandID}}})
	if err != nil {
		return err
	}
	if count != 0 {
		return hrerrors.ErrReferenceConflict
	}
	return nil
}
