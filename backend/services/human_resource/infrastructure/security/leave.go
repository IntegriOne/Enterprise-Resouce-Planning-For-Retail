package security

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
	"github.com/integrione/erp-retail/backend/services/human_resource/repository"
)

type leaveWorkerReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.Worker, error)
}

type leavePolicyReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.LeavePolicy, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.LeavePolicy, error)
}

type leaveEntitlementReader interface {
	FindByIDForUpdate(context.Context, string) (*entity.LeaveEntitlement, error)
}

type leaveLedgerReader interface {
	CountWhere(context.Context, []domainrepo.Filter) (int64, error)
	SelectWhere(context.Context, []domainrepo.Filter, int, int) ([]entity.LeaveLedgerEntry, error)
}

// LeaveValidators menghubungkan policy, hak cuti, dan posting awal dengan repository HRM.
func LeaveValidators(repos *repository.Registry) (map[string]MutationValidator, error) {
	if repos == nil || operation.RequireDependency(repos.Worker) != nil ||
		operation.RequireDependency(repos.LeavePolicy) != nil ||
		operation.RequireDependency(repos.LeaveEntitlement) != nil ||
		operation.RequireDependency(repos.LeaveLedgerEntry) != nil {
		return nil, hrerrors.ErrUnavailable
	}
	return map[string]MutationValidator{
		"hr_leave_policies": &leavePolicyValidator{
			policies: repos.LeavePolicy, entitlements: repos.LeaveEntitlement,
		},
		"hr_leave_entitlements": &leaveEntitlementValidator{
			workers: repos.Worker, policies: repos.LeavePolicy, ledger: repos.LeaveLedgerEntry,
		},
		"hr_leave_ledger_entries": &leaveLedgerValidator{
			entitlements: repos.LeaveEntitlement, policies: repos.LeavePolicy, ledger: repos.LeaveLedgerEntry,
		},
	}, nil
}
