// Kerangka awal dibuat dengan alat bantu; kode ini dipelihara langsung per domain.
// Aturan bisnis lintas record tetap dijalankan melalui policy/integrity dan command domain.
package service

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type Dependencies struct {
	Runtime               operation.Dependencies
	Policy                operation.Repository[entity.Policy, string]
	PolicyAcknowledgement operation.Repository[entity.PolicyAcknowledgement, string]
}

// Service mengoordinasikan operasi domain policy.
type Service struct {
	policy                *operation.Resource[entity.Policy, string, *entity.Policy]
	policyAcknowledgement *operation.Resource[entity.PolicyAcknowledgement, string, *entity.PolicyAcknowledgement]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{}
	var err error
	s.policy, err = operation.NewResource[entity.Policy, string, *entity.Policy](deps.Policy, deps.Runtime, specPolicy())
	if err != nil {
		return nil, err
	}
	s.policyAcknowledgement, err = operation.NewResource[entity.PolicyAcknowledgement, string, *entity.PolicyAcknowledgement](deps.PolicyAcknowledgement, deps.Runtime, specPolicyAcknowledgement())
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreatePolicy menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePolicy(ctx context.Context, in CreatePolicyInput) (*entity.Policy, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.policy.Create(ctx, func() (*entity.Policy, error) { return createPolicy(in) })
}

// UpdatePolicy menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePolicy(ctx context.Context, key string, in UpdatePolicyInput) (*entity.Policy, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.policy.Update(ctx, key, func(value *entity.Policy) error { return patchPolicy(value, in) })
}

// DeletePolicy menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePolicy(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.policy.Delete(ctx, key)
}

// GetPolicyByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPolicyByID(ctx context.Context, key string) (*entity.Policy, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.policy.Get(ctx, key)
}

// GetAllPolicies menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPolicies(ctx context.Context, limit, offset int) (operation.Page[entity.Policy], error) {
	if s == nil {
		return operation.Page[entity.Policy]{}, hrerrors.ErrUnavailable
	}
	return s.policy.List(ctx, nil, limit, offset)
}

// FindPolicyByPolicyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPolicyByPolicyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Policy], error) {
	if s == nil {
		return operation.Page[entity.Policy]{}, hrerrors.ErrUnavailable
	}
	return s.policy.List(ctx, []domainrepo.Filter{{Column: "policy_code", Values: []any{value}}}, limit, offset)
}

// FindPolicyByCategoryCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPolicyByCategoryCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Policy], error) {
	if s == nil {
		return operation.Page[entity.Policy]{}, hrerrors.ErrUnavailable
	}
	return s.policy.List(ctx, []domainrepo.Filter{{Column: "category_code", Values: []any{value}}}, limit, offset)
}

// FindPolicyByLegalEntityReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPolicyByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Policy], error) {
	if s == nil {
		return operation.Page[entity.Policy]{}, hrerrors.ErrUnavailable
	}
	return s.policy.List(ctx, []domainrepo.Filter{{Column: "legal_entity_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPolicyByDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPolicyByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Policy], error) {
	if s == nil {
		return operation.Page[entity.Policy]{}, hrerrors.ErrUnavailable
	}
	return s.policy.List(ctx, []domainrepo.Filter{{Column: "document_version_reference_id", Values: []any{value}}}, limit, offset)
}

// FindPolicyByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPolicyByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.Policy], error) {
	if s == nil {
		return operation.Page[entity.Policy]{}, hrerrors.ErrUnavailable
	}
	return s.policy.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// CreatePolicyAcknowledgement menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreatePolicyAcknowledgement(ctx context.Context, in CreatePolicyAcknowledgementInput) (*entity.PolicyAcknowledgement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.policyAcknowledgement.Create(ctx, func() (*entity.PolicyAcknowledgement, error) { return createPolicyAcknowledgement(in) })
}

// UpdatePolicyAcknowledgement menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdatePolicyAcknowledgement(ctx context.Context, key string, in UpdatePolicyAcknowledgementInput) (*entity.PolicyAcknowledgement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.policyAcknowledgement.Update(ctx, key, func(value *entity.PolicyAcknowledgement) error { return patchPolicyAcknowledgement(value, in) })
}

// DeletePolicyAcknowledgement menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeletePolicyAcknowledgement(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.policyAcknowledgement.Delete(ctx, key)
}

// GetPolicyAcknowledgementByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetPolicyAcknowledgementByID(ctx context.Context, key string) (*entity.PolicyAcknowledgement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.policyAcknowledgement.Get(ctx, key)
}

// GetAllPolicyAcknowledgements menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllPolicyAcknowledgements(ctx context.Context, limit, offset int) (operation.Page[entity.PolicyAcknowledgement], error) {
	if s == nil {
		return operation.Page[entity.PolicyAcknowledgement]{}, hrerrors.ErrUnavailable
	}
	return s.policyAcknowledgement.List(ctx, nil, limit, offset)
}

// FindPolicyAcknowledgementByPolicyID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPolicyAcknowledgementByPolicyID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PolicyAcknowledgement], error) {
	if s == nil {
		return operation.Page[entity.PolicyAcknowledgement]{}, hrerrors.ErrUnavailable
	}
	return s.policyAcknowledgement.List(ctx, []domainrepo.Filter{{Column: "policy_id", Values: []any{value}}}, limit, offset)
}

// FindPolicyAcknowledgementByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPolicyAcknowledgementByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.PolicyAcknowledgement], error) {
	if s == nil {
		return operation.Page[entity.PolicyAcknowledgement]{}, hrerrors.ErrUnavailable
	}
	return s.policyAcknowledgement.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindPolicyAcknowledgementByAcknowledgedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPolicyAcknowledgementByAcknowledgedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PolicyAcknowledgement], error) {
	if s == nil {
		return operation.Page[entity.PolicyAcknowledgement]{}, hrerrors.ErrUnavailable
	}
	return s.policyAcknowledgement.List(ctx, []domainrepo.Filter{{Column: "acknowledged_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindPolicyAcknowledgementByEvidenceDocumentVersionReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindPolicyAcknowledgementByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.PolicyAcknowledgement], error) {
	if s == nil {
		return operation.Page[entity.PolicyAcknowledgement]{}, hrerrors.ErrUnavailable
	}
	return s.policyAcknowledgement.List(ctx, []domainrepo.Filter{{Column: "evidence_document_version_reference_id", Values: []any{value}}}, limit, offset)
}
