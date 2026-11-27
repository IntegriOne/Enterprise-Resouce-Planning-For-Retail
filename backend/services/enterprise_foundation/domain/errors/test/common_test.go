package eferrors_test

import (
	"errors"
	"testing"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

func TestDomainError_MatchingAndWrapping(t *testing.T) {
	t.Parallel()

	rootCause := errors.New("sql duplicate key")
	wrapped := eferrors.Wrap(
		rootCause,
		eferrors.KindAlreadyExists,
		"enterprise_foundation.legal_entity_code_exists",
		"legal entity code already exists",
	)

	if !errors.Is(wrapped, eferrors.ErrAlreadyExists) {
		t.Fatal("expected wrapped error to match ErrAlreadyExists category sentinel")
	}
	if !errors.Is(wrapped, rootCause) {
		t.Fatal("expected wrapped error to preserve rootCause in Unwrap chain")
	}
	if got := eferrors.KindOf(wrapped); got != eferrors.KindAlreadyExists {
		t.Fatalf("got Kind=%v, expected KindAlreadyExists", got)
	}
	if got := eferrors.CodeOf(wrapped); got != "enterprise_foundation.legal_entity_code_exists" {
		t.Fatalf("got Code=%q, expected enterprise_foundation.legal_entity_code_exists", got)
	}

	validationErr := eferrors.NewValidation("code", "enterprise_foundation.invalid_code", "code is required")
	if !errors.Is(validationErr, eferrors.ErrInvalidInput) {
		t.Fatal("expected validation error to match ErrInvalidInput")
	}
	if got := validationErr.Error(); got != "code: code is required" {
		t.Fatalf("got Error()=%q, expected %q", got, "code: code is required")
	}
}
