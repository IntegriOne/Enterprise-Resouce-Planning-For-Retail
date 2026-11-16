package security

import (
	"context"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type leavePolicyValidator struct {
	policies     leavePolicyReader
	entitlements interface {
		CountWhere(context.Context, []domainrepo.Filter) (int64, error)
	}
}

// ValidateMutation membatasi policy aktif pada metode annual grant dan membekukan versi yang dipakai.
func (v *leavePolicyValidator) ValidateMutation(ctx context.Context, table string, before, after entity.Validatable) error {
	if v == nil || table != "hr_leave_policies" || operation.RequireDependency(v.policies) != nil ||
		operation.RequireDependency(v.entitlements) != nil {
		return hrerrors.ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var previous *entity.LeavePolicy
	if before != nil {
		var ok bool
		previous, ok = before.(*entity.LeavePolicy)
		if !ok || previous == nil {
			return hrerrors.ErrInvalidInput
		}
	}
	var policy *entity.LeavePolicy
	if after != nil {
		var ok bool
		policy, ok = after.(*entity.LeavePolicy)
		if !ok || policy == nil || (previous != nil && previous.ID != policy.ID) {
			return hrerrors.ErrInvalidInput
		}
	} else {
		if previous == nil {
			return hrerrors.ErrInvalidInput
		}
		policy = previous
	}
	if previous != nil {
		count, err := v.entitlements.CountWhere(ctx, []domainrepo.Filter{{Column: "leave_policy_id", Values: []any{policy.ID}}})
		if err != nil {
			return err
		}
		if count > 0 && (after == nil || !reflect.DeepEqual(previous, policy)) {
			return hrerrors.ErrReferenceConflict
		}
	}
	if after == nil {
		return nil
	}
	if policy.MinimumServiceDays != nil && *policy.MinimumServiceDays < 0 {
		return hrerrors.Invalid("minimum_service_days", "must be nonnegative")
	}
	if policy.AccrualMethod != "annual_grant" || policy.AnnualEntitlement == nil {
		return hrerrors.ErrWorkflowRequired
	}
	matches, err := v.policies.SelectWhere(ctx, []domainrepo.Filter{
		{Column: "policy_code", Values: []any{policy.PolicyCode}},
		{Column: "policy_version", Values: []any{policy.PolicyVersion}},
		{Column: "jurisdiction_code", Values: []any{policy.JurisdictionCode}},
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
