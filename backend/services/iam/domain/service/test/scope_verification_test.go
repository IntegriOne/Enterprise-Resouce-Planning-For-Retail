package service_test

import (
	"errors"
	"testing"

	"github.com/integrione/erp-retail/backend/services/iam/domain/entity"
	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	domainservice "github.com/integrione/erp-retail/backend/services/iam/domain/service"
)

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
