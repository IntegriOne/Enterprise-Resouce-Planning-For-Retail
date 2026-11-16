package security

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

// Integrity menolak aktivasi mutasi sebelum validator resource tersedia.
// Setiap validator harus memeriksa referensi pemilik dan mengunci parent dalam context transaksi yang sama.
type Integrity struct {
	validators map[string]MutationValidator
}

// MutationValidator adalah kontrak pemeriksaan satu agregat yang sudah diimplementasikan.
type MutationValidator interface {
	ValidateMutation(context.Context, string, entity.Validatable, entity.Validatable) error
}

// NewIntegrity menyalin registry validator tanpa fallback sukses untuk resource yang belum didukung.
func NewIntegrity(validators map[string]MutationValidator) *Integrity {
	registered := make(map[string]MutationValidator, len(validators))
	for resource, validator := range validators {
		registered[resource] = validator
	}
	return &Integrity{validators: registered}
}

// ValidateMutation meneruskan pemeriksaan hanya kepada pemilik aturan resource.
func (i *Integrity) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if i == nil {
		return hrerrors.ErrUnavailable
	}
	validator, ok := i.validators[table]
	if !ok || operation.RequireDependency(validator) != nil {
		return hrerrors.ErrUnavailable
	}
	return validator.ValidateMutation(ctx, table, before, after)
}
