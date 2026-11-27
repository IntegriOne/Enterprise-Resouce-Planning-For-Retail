package human_resource

import (
	"context"

	"github.com/integrione/erp-retail/backend/gateway/graph/model"
	"github.com/integrione/erp-retail/backend/gateway/internal/middleware"
	policy "github.com/integrione/erp-retail/backend/services/human_resource/contract/gen/human_resource/policy/v1"
)

// AcknowledgePolicy mencatat pengakuan kebijakan perusahaan oleh pekerja aktif.
func (a *Adapter) AcknowledgePolicy(ctx context.Context, rawID string, evidenceDocVerRefID *string) (*model.PolicyAcknowledgement, error) {
	rpcCtx, err := a.authorizeResourceMutation(ctx, "hr_policy_acknowledgements", "acknowledge", "human_resource.policy_acknowledgements.acknowledge")
	if err != nil {
		return nil, err
	}
	id, err := normalizeRequiredUUID(rawID, "policy", "policy")
	if err != nil {
		return nil, err
	}
	docVerID, err := normalizeOptionalUUID(evidenceDocVerRefID, "document-version", "evidence_document_version_reference_id")
	if err != nil {
		return nil, err
	}
	if a.policy == nil {
		return nil, middleware.ErrUnavailable("human resource service unavailable", nil)
	}
	resp, err := a.policy.AcknowledgePolicy(rpcCtx, &policy.AcknowledgePolicyRequest{
		Id:                                 id,
		EvidenceDocumentVersionReferenceId: docVerID,
	})
	if err != nil {
		return nil, mapHRMWriteError(err, "policy-acknowledgement")
	}
	return toModelPolicyAcknowledgement(resp.GetValue())
}
