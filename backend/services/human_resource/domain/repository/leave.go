package repository

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
)

// LeavePolicyRepository adalah port persistensi hr_leave_policies.
type LeavePolicyRepository interface {
	Repository[entity.LeavePolicy, string]
	InsertLeavePolicies(context.Context, *entity.LeavePolicy) error
	UpdateLeavePolicies(context.Context, *entity.LeavePolicy) error
	FindLeavePoliciesByUUID(context.Context, string) (*entity.LeavePolicy, error)
	FindLeavePoliciesByUUIDForUpdate(context.Context, string) (*entity.LeavePolicy, error)
	DeleteLeavePoliciesByUUID(context.Context, string) error
	GetAllLeavePolicies(context.Context, int, int) ([]entity.LeavePolicy, error)
	CountLeavePolicies(context.Context) (int64, error)
	SelectLeavePoliciesWhere(context.Context, []Filter, int, int) ([]entity.LeavePolicy, error)
	CountLeavePoliciesWhere(context.Context, []Filter) (int64, error)
	FindLeavePoliciesByPolicyCode(context.Context, string, int, int) ([]entity.LeavePolicy, error)
	FindLeavePoliciesByLeaveTypeCode(context.Context, string, int, int) ([]entity.LeavePolicy, error)
	FindLeavePoliciesByJurisdictionCode(context.Context, string, int, int) ([]entity.LeavePolicy, error)
	FindByPolicyCode(ctx context.Context, value string, limit, offset int) ([]entity.LeavePolicy, error)
	FindByLeaveTypeCode(ctx context.Context, value string, limit, offset int) ([]entity.LeavePolicy, error)
	FindByJurisdictionCode(ctx context.Context, value string, limit, offset int) ([]entity.LeavePolicy, error)
}

// LeaveEntitlementRepository adalah port persistensi hr_leave_entitlements.
type LeaveEntitlementRepository interface {
	Repository[entity.LeaveEntitlement, string]
	InsertLeaveEntitlements(context.Context, *entity.LeaveEntitlement) error
	UpdateLeaveEntitlements(context.Context, *entity.LeaveEntitlement) error
	FindLeaveEntitlementsByUUID(context.Context, string) (*entity.LeaveEntitlement, error)
	FindLeaveEntitlementsByUUIDForUpdate(context.Context, string) (*entity.LeaveEntitlement, error)
	DeleteLeaveEntitlementsByUUID(context.Context, string) error
	GetAllLeaveEntitlements(context.Context, int, int) ([]entity.LeaveEntitlement, error)
	CountLeaveEntitlements(context.Context) (int64, error)
	SelectLeaveEntitlementsWhere(context.Context, []Filter, int, int) ([]entity.LeaveEntitlement, error)
	CountLeaveEntitlementsWhere(context.Context, []Filter) (int64, error)
	FindLeaveEntitlementsByWorkerID(context.Context, string, int, int) ([]entity.LeaveEntitlement, error)
	FindLeaveEntitlementsByLeavePolicyID(context.Context, string, int, int) ([]entity.LeaveEntitlement, error)
	FindLeaveEntitlementsByUOMCode(context.Context, string, int, int) ([]entity.LeaveEntitlement, error)
	FindLeaveEntitlementsBySourceWorkforceSnapshotReferenceID(context.Context, *string, int, int) ([]entity.LeaveEntitlement, error)
	FindByWorkerID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveEntitlement, error)
	FindByLeavePolicyID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveEntitlement, error)
	FindByUOMCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveEntitlement, error)
	FindBySourceWorkforceSnapshotReferenceID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveEntitlement, error)
}

// LeaveAdjustmentRepository adalah port persistensi hr_leave_adjustments.
type LeaveAdjustmentRepository interface {
	Repository[entity.LeaveAdjustment, string]
	InsertLeaveAdjustments(context.Context, *entity.LeaveAdjustment) error
	UpdateLeaveAdjustments(context.Context, *entity.LeaveAdjustment) error
	FindLeaveAdjustmentsByUUID(context.Context, string) (*entity.LeaveAdjustment, error)
	FindLeaveAdjustmentsByUUIDForUpdate(context.Context, string) (*entity.LeaveAdjustment, error)
	DeleteLeaveAdjustmentsByUUID(context.Context, string) error
	GetAllLeaveAdjustments(context.Context, int, int) ([]entity.LeaveAdjustment, error)
	CountLeaveAdjustments(context.Context) (int64, error)
	SelectLeaveAdjustmentsWhere(context.Context, []Filter, int, int) ([]entity.LeaveAdjustment, error)
	CountLeaveAdjustmentsWhere(context.Context, []Filter) (int64, error)
	FindLeaveAdjustmentsByEntitlementID(context.Context, string, int, int) ([]entity.LeaveAdjustment, error)
	FindLeaveAdjustmentsByReasonCode(context.Context, string, int, int) ([]entity.LeaveAdjustment, error)
	FindLeaveAdjustmentsByRequestedByPrincipalID(context.Context, string, int, int) ([]entity.LeaveAdjustment, error)
	FindLeaveAdjustmentsByApprovedByPrincipalID(context.Context, *string, int, int) ([]entity.LeaveAdjustment, error)
	FindLeaveAdjustmentsByIdempotencyKey(context.Context, string, int, int) ([]entity.LeaveAdjustment, error)
	FindByEntitlementID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error)
	FindByReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error)
	FindByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error)
	FindByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveAdjustment, error)
	FindByIdempotencyKey(ctx context.Context, value string, limit, offset int) ([]entity.LeaveAdjustment, error)
}

// LeaveLedgerEntryRepository adalah port persistensi hr_leave_ledger_entries.
type LeaveLedgerEntryRepository interface {
	Repository[entity.LeaveLedgerEntry, string]
	InsertLeaveLedgerEntries(context.Context, *entity.LeaveLedgerEntry) error
	UpdateLeaveLedgerEntries(context.Context, *entity.LeaveLedgerEntry) error
	FindLeaveLedgerEntriesByUUID(context.Context, string) (*entity.LeaveLedgerEntry, error)
	FindLeaveLedgerEntriesByUUIDForUpdate(context.Context, string) (*entity.LeaveLedgerEntry, error)
	DeleteLeaveLedgerEntriesByUUID(context.Context, string) error
	GetAllLeaveLedgerEntries(context.Context, int, int) ([]entity.LeaveLedgerEntry, error)
	CountLeaveLedgerEntries(context.Context) (int64, error)
	SelectLeaveLedgerEntriesWhere(context.Context, []Filter, int, int) ([]entity.LeaveLedgerEntry, error)
	CountLeaveLedgerEntriesWhere(context.Context, []Filter) (int64, error)
	FindLeaveLedgerEntriesByEntitlementID(context.Context, string, int, int) ([]entity.LeaveLedgerEntry, error)
	FindLeaveLedgerEntriesByEntryType(context.Context, string, int, int) ([]entity.LeaveLedgerEntry, error)
	FindLeaveLedgerEntriesByAdjustmentID(context.Context, *string, int, int) ([]entity.LeaveLedgerEntry, error)
	FindLeaveLedgerEntriesByReversesEntryID(context.Context, *string, int, int) ([]entity.LeaveLedgerEntry, error)
	FindLeaveLedgerEntriesBySourceManagement(context.Context, string, int, int) ([]entity.LeaveLedgerEntry, error)
	FindLeaveLedgerEntriesBySourceEventID(context.Context, string, int, int) ([]entity.LeaveLedgerEntry, error)
	FindLeaveLedgerEntriesByReasonCode(context.Context, string, int, int) ([]entity.LeaveLedgerEntry, error)
	FindLeaveLedgerEntriesByActorPrincipalID(context.Context, string, int, int) ([]entity.LeaveLedgerEntry, error)
	FindByEntitlementID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error)
	FindByEntryType(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error)
	FindByAdjustmentID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveLedgerEntry, error)
	FindByReversesEntryID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveLedgerEntry, error)
	FindBySourceManagement(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error)
	FindBySourceEventID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error)
	FindByReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error)
	FindByActorPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error)
}
