package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

type stubScopeOwnerVerifier struct {
	err       error
	lastScope entity.Scope
	calls     int
}

func (s *stubScopeOwnerVerifier) VerifyScope(_ context.Context, scope entity.Scope) error {
	s.calls++
	s.lastScope = scope
	return s.err
}

func TestRequireScopeOwnerVerification(t *testing.T) {
	t.Parallel()
	id := "11111111-1111-4111-8111-111111111111"
	tests := []struct {
		name    string
		scope   entity.Scope
		wantErr bool
	}{
		{name: "global", scope: entity.GlobalScope()},
		{name: "self", scope: entity.Scope{Type: entity.ScopeTypeSelf, ID: &id}},
		{name: "legal entity", scope: entity.Scope{Type: entity.ScopeTypeLegalEntity, ID: &id}, wantErr: true},
		{name: "facility", scope: entity.Scope{Type: entity.ScopeTypeFacility, ID: &id}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domainservice.RequireScopeOwnerVerification(tt.scope)
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected scope error: %v", err)
			}
			if tt.wantErr {
				var domainErr *iamerrors.DomainError
				if !errors.As(err, &domainErr) || domainErr.Kind != iamerrors.KindUnavailable {
					t.Fatalf("scope error = %v, want unavailable", err)
				}
			}
		})
	}
}

func TestVerifyScope(t *testing.T) {
	t.Parallel()
	id := "11111111-1111-4111-8111-111111111111"

	tests := []struct {
		name      string
		verifier  domainservice.ScopeOwnerVerifier
		scope     entity.Scope
		wantCalls int
		wantErr   error
	}{
		{
			name:     "invalid scope rejected before verifier",
			verifier: &stubScopeOwnerVerifier{},
			scope:    entity.Scope{Type: entity.ScopeTypeLegalEntity},
			wantErr:  iamerrors.ErrInvalidScope,
		},
		{
			name:     "global scope bypasses external verifier",
			verifier: &stubScopeOwnerVerifier{err: errors.New("should not be called")},
			scope:    entity.GlobalScope(),
		},
		{
			name:     "self scope bypasses external verifier",
			verifier: &stubScopeOwnerVerifier{err: errors.New("should not be called")},
			scope:    entity.Scope{Type: entity.ScopeTypeSelf, ID: &id},
		},
		{
			name:     "nil verifier fails closed for external scope",
			verifier: nil,
			scope:    entity.Scope{Type: entity.ScopeTypeOrganizationUnit, ID: &id},
			wantErr:  iamerrors.ErrScopeOwnerVerificationUnavailable,
		},
		{
			name:      "eligible external scope verified cleanly",
			verifier:  &stubScopeOwnerVerifier{},
			scope:     entity.Scope{Type: entity.ScopeTypeLegalEntity, ID: &id},
			wantCalls: 1,
		},
		{
			name:      "ineligible external scope returns verifier error",
			verifier:  &stubScopeOwnerVerifier{err: iamerrors.ErrScopeTargetNotEligible},
			scope:     entity.Scope{Type: entity.ScopeTypeVendorParty, ID: &id},
			wantCalls: 1,
			wantErr:   iamerrors.ErrScopeTargetNotEligible,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domainservice.VerifyScope(context.Background(), tt.verifier, tt.scope)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("VerifyScope() error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("VerifyScope() unexpected error = %v", err)
			}
			if stub, ok := tt.verifier.(*stubScopeOwnerVerifier); ok && stub.calls != tt.wantCalls {
				t.Fatalf("verifier calls = %d, want %d", stub.calls, tt.wantCalls)
			}
		})
	}
}
