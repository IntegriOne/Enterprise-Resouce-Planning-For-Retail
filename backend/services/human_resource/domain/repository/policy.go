package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// PolicyRepository adalah port persistensi hr_policies.
type PolicyRepository interface {
	Repository[entity.Policy, string]
	InsertPolicies(context.Context, *entity.Policy) error
	UpdatePolicies(context.Context, *entity.Policy) error
	FindPoliciesByUUID(context.Context, string) (*entity.Policy, error)
	FindPoliciesByUUIDForUpdate(context.Context, string) (*entity.Policy, error)
	DeletePoliciesByUUID(context.Context, string) error
	GetAllPolicies(context.Context, int, int) ([]entity.Policy, error)
	CountPolicies(context.Context) (int64, error)
	SelectPoliciesWhere(context.Context, []Filter, int, int) ([]entity.Policy, error)
	CountPoliciesWhere(context.Context, []Filter) (int64, error)
	FindPoliciesByPolicyCode(context.Context, string, int, int) ([]entity.Policy, error)
	FindPoliciesByCategoryCode(context.Context, string, int, int) ([]entity.Policy, error)
	FindPoliciesByLegalEntityReferenceID(context.Context, string, int, int) ([]entity.Policy, error)
	FindPoliciesByDocumentVersionReferenceID(context.Context, string, int, int) ([]entity.Policy, error)
	FindPoliciesByApprovedByPrincipalID(context.Context, string, int, int) ([]entity.Policy, error)
	FindByPolicyCode(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error)
	FindByCategoryCode(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error)
	FindByLegalEntityReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error)
	FindByDocumentVersionReferenceID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error)
	FindByApprovedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.Policy, error)
}

// PolicyAcknowledgementRepository adalah port persistensi hr_policy_acknowledgements.
type PolicyAcknowledgementRepository interface {
	Repository[entity.PolicyAcknowledgement, string]
	InsertPolicyAcknowledgements(context.Context, *entity.PolicyAcknowledgement) error
	UpdatePolicyAcknowledgements(context.Context, *entity.PolicyAcknowledgement) error
	FindPolicyAcknowledgementsByUUID(context.Context, string) (*entity.PolicyAcknowledgement, error)
	FindPolicyAcknowledgementsByUUIDForUpdate(context.Context, string) (*entity.PolicyAcknowledgement, error)
	DeletePolicyAcknowledgementsByUUID(context.Context, string) error
	GetAllPolicyAcknowledgements(context.Context, int, int) ([]entity.PolicyAcknowledgement, error)
	CountPolicyAcknowledgements(context.Context) (int64, error)
	SelectPolicyAcknowledgementsWhere(context.Context, []Filter, int, int) ([]entity.PolicyAcknowledgement, error)
	CountPolicyAcknowledgementsWhere(context.Context, []Filter) (int64, error)
	FindPolicyAcknowledgementsByPolicyID(context.Context, string, int, int) ([]entity.PolicyAcknowledgement, error)
	FindPolicyAcknowledgementsByWorkerID(context.Context, string, int, int) ([]entity.PolicyAcknowledgement, error)
	FindPolicyAcknowledgementsByAcknowledgedByPrincipalID(context.Context, *string, int, int) ([]entity.PolicyAcknowledgement, error)
	FindPolicyAcknowledgementsByEvidenceDocumentVersionReferenceID(context.Context, *string, int, int) ([]entity.PolicyAcknowledgement, error)
	FindByPolicyID(ctx context.Context, value string, limit, offset int) ([]entity.PolicyAcknowledgement, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.PolicyAcknowledgement, error)
	FindByAcknowledgedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.PolicyAcknowledgement, error)
	FindByEvidenceDocumentVersionReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.PolicyAcknowledgement, error)
}
