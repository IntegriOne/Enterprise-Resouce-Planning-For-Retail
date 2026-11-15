package service_test

import (
	"errors"
	"strings"
	"testing"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

func TestPasswordPolicyValidation(t *testing.T) {
	t.Parallel()

	policy := service.NewDefaultPasswordPolicy()

	cases := []struct {
		name      string
		password  string
		forbidden []string
		wantErr   error
	}{
		{
			name:      "valid strong password",
			password:  "Tr4ck!ngWarehouse#26",
			forbidden: []string{"budi.santoso@integrione.id", "budi.santoso"},
			wantErr:   nil,
		},
		{
			name:     "too short",
			password: "Ab1!short",
			wantErr:  iamerrors.ErrPasswordTooShort,
		},
		{
			name:     "too long",
			password: "Aa1!" + strings.Repeat("x", 130),
			wantErr:  iamerrors.ErrPasswordTooLong,
		},
		{
			name:     "missing uppercase",
			password: "lowercaseonly123!",
			wantErr:  iamerrors.ErrPasswordTooWeak,
		},
		{
			name:     "missing digit",
			password: "NoDigitsHere!Secure",
			wantErr:  iamerrors.ErrPasswordTooWeak,
		},
		{
			name:     "missing symbol",
			password: "NoSymbolHere1234",
			wantErr:  iamerrors.ErrPasswordTooWeak,
		},
		{
			name:     "contains nul byte",
			password: "ValidPass1!\x00extra",
			wantErr:  iamerrors.ErrPasswordTooWeak,
		},
		{
			name:      "contains email local part case-insensitively",
			password:  "MyBudiSantoso#99",
			forbidden: []string{"budi.santoso@integrione.id", "budisantoso"},
			wantErr:   iamerrors.ErrPasswordContainsIdentity,
		},
		{
			name:      "contains display name token",
			password:  "KaryawanPratama#2026",
			forbidden: []string{"Budi Pratama"},
			wantErr:   iamerrors.ErrPasswordContainsIdentity,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := policy.Validate(tc.password, tc.forbidden...)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Validate() = %v, want %v", err, tc.wantErr)
			}
		})
	}

	custom := service.NewPasswordPolicy(service.PasswordPolicyConfig{MinLength: 8, MaxLength: 64, RequireLower: true})
	if err := custom.Validate("simplepass"); err != nil {
		t.Fatalf("custom PasswordPolicy Validate() unexpected error: %v", err)
	}
	fallback := service.NewPasswordPolicy(service.PasswordPolicyConfig{})
	if err := fallback.Validate("short"); !errors.Is(err, iamerrors.ErrPasswordTooShort) {
		t.Fatalf("fallback PasswordPolicy Validate(short) = %v, want ErrPasswordTooShort", err)
	}
}
