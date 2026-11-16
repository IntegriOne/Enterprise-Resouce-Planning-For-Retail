package pgstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// MapError menjaga klasifikasi errors.Is tanpa menyalin SQL/detail PII ke pesan.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	for _, kind := range []error{hrerrors.ErrInvalidInput, hrerrors.ErrNotFound, hrerrors.ErrAlreadyExists, hrerrors.ErrReferenceConflict, hrerrors.ErrConstraintViolation, hrerrors.ErrConflict, hrerrors.ErrUnavailable, hrerrors.ErrTransactionRequired} {
		if errors.Is(err, kind) {
			return err
		}
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	kind := hrerrors.ErrUnavailable
	var state interface{ SQLState() string }
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		kind = hrerrors.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		kind = hrerrors.ErrAlreadyExists
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		kind = hrerrors.ErrReferenceConflict
	case errors.Is(err, gorm.ErrCheckConstraintViolated):
		kind = hrerrors.ErrConstraintViolation
	case errors.As(err, &state):
		code := state.SQLState()
		switch {
		case code == "23505":
			kind = hrerrors.ErrAlreadyExists
		case code == "23503" || code == "23001":
			kind = hrerrors.ErrReferenceConflict
		case code == "23514" || code == "23P01":
			kind = hrerrors.ErrConstraintViolation
		case code == "40001" || code == "40P01":
			kind = hrerrors.ErrConflict
		case strings.HasPrefix(code, "22") || code == "23502":
			kind = hrerrors.ErrInvalidInput
		}
	}
	return &persistenceError{kind: kind, cause: err}
}

type persistenceError struct{ kind, cause error }

func (e *persistenceError) Error() string        { return fmt.Sprint(e.kind) }
func (e *persistenceError) Unwrap() error        { return e.cause }
func (e *persistenceError) Is(target error) bool { return errors.Is(e.kind, target) }
