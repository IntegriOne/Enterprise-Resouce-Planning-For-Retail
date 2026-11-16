package errors

import "fmt"

// ValidationError menyebut field atau invariant tanpa membocorkan nilai PII.
type ValidationError struct{ Field, Rule string }

func (e *ValidationError) Error() string {
	return fmt.Sprintf("hrm.invalid_input: %s: %s", e.Field, e.Rule)
}
func (e *ValidationError) Unwrap() error { return ErrInvalidInput }

// Invalid membentuk error validasi yang dapat diperiksa memakai errors.Is/As.
func Invalid(field, rule string) error { return &ValidationError{Field: field, Rule: rule} }
