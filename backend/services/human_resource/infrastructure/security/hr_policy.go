package security

import (
	"context"
	"reflect"
	"strings"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
)

type policyReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Policy, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.Policy, error)
}

type policyAcknowledgementReader interface {
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.PolicyAcknowledgement, error)
}

type policyValidator struct {
	policies   policyReader
	foundation FoundationVerifier
}

type policyAcknowledgementValidator struct {
	policies policyReader
	workers  lockedWorkerReader
	acks     policyAcknowledgementReader
}

// PolicyValidators mendaftarkan validator integritas kebijakan HR dan bukti pengakuan pekerja.
func PolicyValidators(repos *repository.Registry, foundation FoundationVerifier) (map[string]MutationValidator, error) {
	if repos == nil ||
		operation.RequireDependency(repos.Policy) != nil ||
		operation.RequireDependency(repos.PolicyAcknowledgement) != nil ||
		operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(foundation) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_policies": &policyValidator{
			policies:   repos.Policy,
			foundation: foundation,
		},
		"hr_policy_acknowledgements": &policyAcknowledgementValidator{
			policies: repos.Policy,
			workers:  repos.Worker,
			acks:     repos.PolicyAcknowledgement,
		},
	}, nil
}

// ValidateMutation membekukan kebijakan terbit dan menjaga keunikan versi kebijakan per entitas legal.
func (v *policyValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_policies" ||
		operation.RequireDependency(v.policies) != nil ||
		operation.RequireDependency(v.foundation) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if before == nil && after == nil {
		return hrerrors.ErrInvalidInput
	}
	if before != nil || after == nil {
		return hrerrors.ErrWorkflowRequired
	}
	policy, ok := after.(*entity.Policy)
	if !ok || policy == nil {
		return hrerrors.ErrInvalidInput
	}
	if !supportedPolicyCategory(policy.CategoryCode) {
		return hrerrors.ErrInvalidInput
	}
	caller, ok := authcontext.FromContext(ctx)
	if !ok || !caller.Authenticated {
		return hrerrors.ErrUnauthenticated
	}
	if !strings.EqualFold(caller.PrincipalID, policy.ApprovedByPrincipalID) {
		return hrerrors.ErrForbidden
	}
	if err := verifyFoundationLegalEntity(ctx, v.foundation, policy.LegalEntityReferenceID); err != nil {
		return err
	}
	matches, err := v.policies.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "legal_entity_reference_id", Values: []any{policy.LegalEntityReferenceID}},
		{Column: "policy_code", Values: []any{policy.PolicyCode}},
		{Column: "policy_version", Values: []any{policy.PolicyVersion}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != policy.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

// ValidateMutation memastikan pengakuan kebijakan hanya dicatat oleh pekerja pemilik melalui command acknowledge.
func (v *policyAcknowledgementValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_policy_acknowledgements" ||
		operation.RequireDependency(v.policies) != nil ||
		operation.RequireDependency(v.workers) != nil ||
		operation.RequireDependency(v.acks) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.PolicyAcknowledgement
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.PolicyAcknowledgement)
		if !ok || previous == nil || previous.AcknowledgedAt != nil {
			return hrerrors.ErrWorkflowRequired
		}
	}
	if after == nil {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		return nil
	}
	ack, ok := after.(*entity.PolicyAcknowledgement)
	if !ok || ack == nil || (previous != nil && previous.ID != ack.ID) {
		return hrerrors.ErrInvalidInput
	}
	acknowledging := previous != nil && previous.AcknowledgedAt == nil && ack.AcknowledgedAt != nil
	if acknowledging {
		if ack.AcknowledgedByPrincipalID == nil {
			return hrerrors.ErrWorkflowRequired
		}
		caller, ok := authcontext.FromContext(ctx)
		if !ok || !caller.Authenticated {
			return hrerrors.ErrUnauthenticated
		}
		if !strings.EqualFold(caller.PrincipalID, *ack.AcknowledgedByPrincipalID) {
			return hrerrors.ErrForbidden
		}
		unchanged := *ack
		unchanged.AcknowledgedAt = nil
		unchanged.AcknowledgedByPrincipalID = nil
		unchanged.EvidenceDocumentVersionReferenceID = previous.EvidenceDocumentVersionReferenceID
		if !reflect.DeepEqual(previous, &unchanged) {
			return hrerrors.ErrWorkflowRequired
		}
	} else {
		if ack.AcknowledgedAt != nil || ack.AcknowledgedByPrincipalID != nil {
			return hrerrors.ErrWorkflowRequired
		}
		if previous != nil && (previous.PolicyID != ack.PolicyID ||
			previous.WorkerID != ack.WorkerID || !previous.AssignedAt.Equal(ack.AssignedAt)) {
			return hrerrors.ErrWorkflowRequired
		}
	}
	worker, err := v.workers.FindByIDForUpdate(ctx, ack.WorkerID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if worker == nil || worker.Status != entity.WorkerStatusActive {
		return hrerrors.ErrReferenceConflict
	}
	if acknowledging {
		if worker.IAMPrincipalReferenceID == nil {
			return hrerrors.ErrReferenceConflict
		}
		if !strings.EqualFold(*worker.IAMPrincipalReferenceID, *ack.AcknowledgedByPrincipalID) {
			return hrerrors.ErrForbidden
		}
	}
	// Bukti baru perlu dibuktikan oleh DRM sebelum tercatat sebagai bagian dari pengakuan.
	if ack.EvidenceDocumentVersionReferenceID != nil &&
		(previous == nil || previous.EvidenceDocumentVersionReferenceID == nil ||
			!strings.EqualFold(*ack.EvidenceDocumentVersionReferenceID, *previous.EvidenceDocumentVersionReferenceID)) {
		return hrerrors.ErrUnavailable
	}
	policy, err := v.policies.FindByIDForUpdate(ctx, ack.PolicyID)
	if err != nil {
		return reviewReferenceError(err)
	}
	if policy == nil || !strings.EqualFold(policy.LegalEntityReferenceID, worker.PrimaryLegalEntityReferenceID) ||
		(previous == nil && ack.AssignedAt.Before(policy.ApprovedAt)) ||
		(acknowledging && ack.AcknowledgedAt.Before(policy.ApprovedAt)) ||
		(policy.EffectiveTo != nil && entity.CompareDates(*policy.EffectiveTo, ack.AssignedAt) < 0) {
		return hrerrors.ErrReferenceConflict
	}
	matches, err := v.acks.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "policy_id", Values: []any{ack.PolicyID}},
		{Column: "worker_id", Values: []any{ack.WorkerID}},
	}, 2, 0)
	if err != nil {
		return err
	}
	for _, existing := range matches {
		if existing.ID != ack.ID {
			return hrerrors.ErrAlreadyExists
		}
	}
	return nil
}

func supportedPolicyCategory(value string) bool {
	switch value {
	case "conduct", "inclusion", "wellbeing", "remote_work", "grievance", "other":
		return true
	default:
		return false
	}
}
