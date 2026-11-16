// Package errors menyediakan klasifikasi error persistensi HRM.
package errors

import "errors"

var (
	ErrUnauthenticated     = errors.New("hrm.unauthenticated")
	ErrForbidden           = errors.New("hrm.forbidden")
	ErrWorkflowRequired    = errors.New("hrm.workflow_required")
	ErrInvalidInput        = errors.New("hrm.invalid_input")
	ErrNotFound            = errors.New("hrm.not_found")
	ErrAlreadyExists       = errors.New("hrm.already_exists")
	ErrReferenceConflict   = errors.New("hrm.reference_conflict")
	ErrConstraintViolation = errors.New("hrm.constraint_violation")
	ErrConflict            = errors.New("hrm.conflict")
	ErrUnavailable         = errors.New("hrm.unavailable")
	ErrTransactionRequired = errors.New("hrm.transaction_required")
)
