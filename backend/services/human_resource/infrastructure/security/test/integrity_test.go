package security_test

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	"github.com/integrione/erp-retail/backend/services/human_resource/infrastructure/security"
)

type mutationValidator struct {
	calls int
	err   error
}

func (v *mutationValidator) ValidateMutation(context.Context, string, entity.Validatable, entity.Validatable) error {
	v.calls++
	return v.err
}

func TestIntegrityRejectsMissingOrTypedNilValidator(t *testing.T) {
	t.Parallel()
	var absent *mutationValidator
	for _, validators := range []map[string]security.MutationValidator{nil, {"hr_jobs": absent}} {
		guard := security.NewIntegrity(validators)
		if err := guard.ValidateMutation(t.Context(), "hr_jobs", nil, nil); !errors.Is(err, hrerrors.ErrUnavailable) {
			t.Fatalf("missing validator: %v", err)
		}
	}
}

func TestIntegrityPreservesFailureAndCopiesRegistry(t *testing.T) {
	t.Parallel()
	validator := &mutationValidator{err: hrerrors.ErrReferenceConflict}
	validators := map[string]security.MutationValidator{"hr_jobs": validator}
	guard := security.NewIntegrity(validators)
	delete(validators, "hr_jobs")
	if err := guard.ValidateMutation(t.Context(), "hr_jobs", nil, nil); !errors.Is(err, hrerrors.ErrReferenceConflict) {
		t.Fatalf("reference error lost: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := guard.ValidateMutation(ctx, "hr_jobs", nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
	if validator.calls != 1 {
		t.Fatalf("unexpected validation calls: %d", validator.calls)
	}
}
