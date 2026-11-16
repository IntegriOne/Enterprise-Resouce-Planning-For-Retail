// Kerangka awal dibuat oleh alat bantu; selanjutnya dipelihara langsung per domain.
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
	Runtime          operation.Dependencies
	LeavePolicy      domainrepo.RecordRepository[entity.LeavePolicy, string]
	LeaveEntitlement domainrepo.RecordRepository[entity.LeaveEntitlement, string]
	LeaveAdjustment  domainrepo.RecordRepository[entity.LeaveAdjustment, string]
	LeaveLedgerEntry domainrepo.RecordRepository[entity.LeaveLedgerEntry, string]
}

// Service mengoordinasikan operasi domain leave.
type Service struct {
	dependencies     Dependencies
	leavePolicy      *operation.Resource[entity.LeavePolicy, string, *entity.LeavePolicy]
	leaveEntitlement *operation.Resource[entity.LeaveEntitlement, string, *entity.LeaveEntitlement]
	leaveAdjustment  *operation.Resource[entity.LeaveAdjustment, string, *entity.LeaveAdjustment]
	leaveLedgerEntry *operation.Resource[entity.LeaveLedgerEntry, string, *entity.LeaveLedgerEntry]
}

// NewService memeriksa seluruh port dan dependency sebelum service dapat dipakai.
func NewService(deps Dependencies) (*Service, error) {
	s := &Service{dependencies: deps}
	var err error
	s.leavePolicy, err = operation.NewResource[entity.LeavePolicy, string, *entity.LeavePolicy](
		deps.LeavePolicy,
		deps.Runtime,
		specLeavePolicy(),
	)
	if err != nil {
		return nil, err
	}
	s.leaveEntitlement, err = operation.NewResource[entity.LeaveEntitlement, string, *entity.LeaveEntitlement](
		deps.LeaveEntitlement,
		deps.Runtime,
		specLeaveEntitlement(),
	)
	if err != nil {
		return nil, err
	}
	s.leaveAdjustment, err = operation.NewResource[entity.LeaveAdjustment, string, *entity.LeaveAdjustment](
		deps.LeaveAdjustment,
		deps.Runtime,
		specLeaveAdjustment(),
	)
	if err != nil {
		return nil, err
	}
	s.leaveLedgerEntry, err = operation.NewResource[entity.LeaveLedgerEntry, string, *entity.LeaveLedgerEntry](
		deps.LeaveLedgerEntry,
		deps.Runtime,
		specLeaveLedgerEntry(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// CreateLeavePolicy menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateLeavePolicy(ctx context.Context, in CreateLeavePolicyInput) (*entity.LeavePolicy, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leavePolicy.Create(ctx, func() (*entity.LeavePolicy, error) { return createLeavePolicy(in) })
}

// UpdateLeavePolicy menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateLeavePolicy(ctx context.Context, key string, in UpdateLeavePolicyInput) (*entity.LeavePolicy, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leavePolicy.Update(ctx, key, func(value *entity.LeavePolicy) error { return patchLeavePolicy(value, in) })
}

// DeleteLeavePolicy menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteLeavePolicy(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.leavePolicy.Delete(ctx, key)
}

// GetLeavePolicyByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetLeavePolicyByID(ctx context.Context, key string) (*entity.LeavePolicy, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leavePolicy.Get(ctx, key)
}

// GetAllLeavePolicies menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllLeavePolicies(ctx context.Context, limit, offset int) (operation.Page[entity.LeavePolicy], error) {
	if s == nil {
		return operation.Page[entity.LeavePolicy]{}, hrerrors.ErrUnavailable
	}
	return s.leavePolicy.List(ctx, nil, limit, offset)
}

// FindLeavePolicyByPolicyCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeavePolicyByPolicyCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeavePolicy], error) {
	if s == nil {
		return operation.Page[entity.LeavePolicy]{}, hrerrors.ErrUnavailable
	}
	return s.leavePolicy.List(ctx, []domainrepo.Filter{{Column: "policy_code", Values: []any{value}}}, limit, offset)
}

// FindLeavePolicyByLeaveTypeCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeavePolicyByLeaveTypeCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeavePolicy], error) {
	if s == nil {
		return operation.Page[entity.LeavePolicy]{}, hrerrors.ErrUnavailable
	}
	return s.leavePolicy.List(ctx, []domainrepo.Filter{{Column: "leave_type_code", Values: []any{value}}}, limit, offset)
}

// FindLeavePolicyByJurisdictionCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeavePolicyByJurisdictionCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeavePolicy], error) {
	if s == nil {
		return operation.Page[entity.LeavePolicy]{}, hrerrors.ErrUnavailable
	}
	return s.leavePolicy.List(ctx, []domainrepo.Filter{{Column: "jurisdiction_code", Values: []any{value}}}, limit, offset)
}

// CreateLeaveEntitlement menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateLeaveEntitlement(ctx context.Context, in CreateLeaveEntitlementInput) (*entity.LeaveEntitlement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.Create(ctx, func() (*entity.LeaveEntitlement, error) { return createLeaveEntitlement(in) })
}

// UpdateLeaveEntitlement menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateLeaveEntitlement(ctx context.Context, key string, in UpdateLeaveEntitlementInput) (*entity.LeaveEntitlement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.Update(ctx, key, func(value *entity.LeaveEntitlement) error { return patchLeaveEntitlement(value, in) })
}

// DeleteLeaveEntitlement menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteLeaveEntitlement(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.Delete(ctx, key)
}

// GetLeaveEntitlementByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetLeaveEntitlementByID(ctx context.Context, key string) (*entity.LeaveEntitlement, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.Get(ctx, key)
}

// GetAllLeaveEntitlements menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllLeaveEntitlements(ctx context.Context, limit, offset int) (operation.Page[entity.LeaveEntitlement], error) {
	if s == nil {
		return operation.Page[entity.LeaveEntitlement]{}, hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.List(ctx, nil, limit, offset)
}

// FindLeaveEntitlementByWorkerID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveEntitlementByWorkerID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveEntitlement], error) {
	if s == nil {
		return operation.Page[entity.LeaveEntitlement]{}, hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.List(ctx, []domainrepo.Filter{{Column: "worker_id", Values: []any{value}}}, limit, offset)
}

// FindLeaveEntitlementByLeavePolicyID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveEntitlementByLeavePolicyID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveEntitlement], error) {
	if s == nil {
		return operation.Page[entity.LeaveEntitlement]{}, hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.List(ctx, []domainrepo.Filter{{Column: "leave_policy_id", Values: []any{value}}}, limit, offset)
}

// FindLeaveEntitlementByUOMCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveEntitlementByUOMCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveEntitlement], error) {
	if s == nil {
		return operation.Page[entity.LeaveEntitlement]{}, hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.List(ctx, []domainrepo.Filter{{Column: "uom_code", Values: []any{value}}}, limit, offset)
}

// FindLeaveEntitlementBySourceWorkforceSnapshotReferenceID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveEntitlementBySourceWorkforceSnapshotReferenceID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.LeaveEntitlement], error) {
	if s == nil {
		return operation.Page[entity.LeaveEntitlement]{}, hrerrors.ErrUnavailable
	}
	return s.leaveEntitlement.List(ctx, []domainrepo.Filter{{Column: "source_workforce_snapshot_reference_id", Values: []any{value}}}, limit, offset)
}

// CreateLeaveAdjustment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateLeaveAdjustment(ctx context.Context, in CreateLeaveAdjustmentInput) (*entity.LeaveAdjustment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.Create(ctx, func() (*entity.LeaveAdjustment, error) { return createLeaveAdjustment(in) })
}

// UpdateLeaveAdjustment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateLeaveAdjustment(ctx context.Context, key string, in UpdateLeaveAdjustmentInput) (*entity.LeaveAdjustment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.Update(ctx, key, func(value *entity.LeaveAdjustment) error { return patchLeaveAdjustment(value, in) })
}

// DeleteLeaveAdjustment menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteLeaveAdjustment(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.Delete(ctx, key)
}

// GetLeaveAdjustmentByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetLeaveAdjustmentByID(ctx context.Context, key string) (*entity.LeaveAdjustment, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.Get(ctx, key)
}

// GetAllLeaveAdjustments menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllLeaveAdjustments(ctx context.Context, limit, offset int) (operation.Page[entity.LeaveAdjustment], error) {
	if s == nil {
		return operation.Page[entity.LeaveAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.List(ctx, nil, limit, offset)
}

// FindLeaveAdjustmentByEntitlementID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveAdjustmentByEntitlementID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveAdjustment], error) {
	if s == nil {
		return operation.Page[entity.LeaveAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.List(ctx, []domainrepo.Filter{{Column: "entitlement_id", Values: []any{value}}}, limit, offset)
}

// FindLeaveAdjustmentByReasonCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveAdjustmentByReasonCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveAdjustment], error) {
	if s == nil {
		return operation.Page[entity.LeaveAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.List(ctx, []domainrepo.Filter{{Column: "reason_code", Values: []any{value}}}, limit, offset)
}

// FindLeaveAdjustmentByRequestedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveAdjustmentByRequestedByPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveAdjustment], error) {
	if s == nil {
		return operation.Page[entity.LeaveAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.List(ctx, []domainrepo.Filter{{Column: "requested_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindLeaveAdjustmentByApprovedByPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveAdjustmentByApprovedByPrincipalID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.LeaveAdjustment], error) {
	if s == nil {
		return operation.Page[entity.LeaveAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.List(ctx, []domainrepo.Filter{{Column: "approved_by_principal_id", Values: []any{value}}}, limit, offset)
}

// FindLeaveAdjustmentByIdempotencyKey menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveAdjustmentByIdempotencyKey(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveAdjustment], error) {
	if s == nil {
		return operation.Page[entity.LeaveAdjustment]{}, hrerrors.ErrUnavailable
	}
	return s.leaveAdjustment.List(ctx, []domainrepo.Filter{{Column: "idempotency_key", Values: []any{value}}}, limit, offset)
}

// CreateLeaveLedgerEntry menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) CreateLeaveLedgerEntry(ctx context.Context, in CreateLeaveLedgerEntryInput) (*entity.LeaveLedgerEntry, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.Create(ctx, func() (*entity.LeaveLedgerEntry, error) { return createLeaveLedgerEntry(in) })
}

// UpdateLeaveLedgerEntry menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) UpdateLeaveLedgerEntry(ctx context.Context, key string, in UpdateLeaveLedgerEntryInput) (*entity.LeaveLedgerEntry, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.Update(ctx, key, func(value *entity.LeaveLedgerEntry) error { return patchLeaveLedgerEntry(value, in) })
}

// DeleteLeaveLedgerEntry menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) DeleteLeaveLedgerEntry(ctx context.Context, key string) error {
	if s == nil {
		return hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.Delete(ctx, key)
}

// GetLeaveLedgerEntryByID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetLeaveLedgerEntryByID(ctx context.Context, key string) (*entity.LeaveLedgerEntry, error) {
	if s == nil {
		return nil, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.Get(ctx, key)
}

// GetAllLeaveLedgerEntries menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) GetAllLeaveLedgerEntries(ctx context.Context, limit, offset int) (operation.Page[entity.LeaveLedgerEntry], error) {
	if s == nil {
		return operation.Page[entity.LeaveLedgerEntry]{}, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.List(ctx, nil, limit, offset)
}

// FindLeaveLedgerEntryByEntitlementID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveLedgerEntryByEntitlementID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveLedgerEntry], error) {
	if s == nil {
		return operation.Page[entity.LeaveLedgerEntry]{}, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.List(ctx, []domainrepo.Filter{{Column: "entitlement_id", Values: []any{value}}}, limit, offset)
}

// FindLeaveLedgerEntryByEntryType menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveLedgerEntryByEntryType(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveLedgerEntry], error) {
	if s == nil {
		return operation.Page[entity.LeaveLedgerEntry]{}, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.List(ctx, []domainrepo.Filter{{Column: "entry_type", Values: []any{value}}}, limit, offset)
}

// FindLeaveLedgerEntryByAdjustmentID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveLedgerEntryByAdjustmentID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.LeaveLedgerEntry], error) {
	if s == nil {
		return operation.Page[entity.LeaveLedgerEntry]{}, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.List(ctx, []domainrepo.Filter{{Column: "adjustment_id", Values: []any{value}}}, limit, offset)
}

// FindLeaveLedgerEntryByReversesEntryID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveLedgerEntryByReversesEntryID(ctx context.Context, value *string, limit, offset int) (operation.Page[entity.LeaveLedgerEntry], error) {
	if s == nil {
		return operation.Page[entity.LeaveLedgerEntry]{}, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.List(ctx, []domainrepo.Filter{{Column: "reverses_entry_id", Values: []any{value}}}, limit, offset)
}

// FindLeaveLedgerEntryBySourceManagement menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveLedgerEntryBySourceManagement(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveLedgerEntry], error) {
	if s == nil {
		return operation.Page[entity.LeaveLedgerEntry]{}, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.List(ctx, []domainrepo.Filter{{Column: "source_management", Values: []any{value}}}, limit, offset)
}

// FindLeaveLedgerEntryBySourceEventID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveLedgerEntryBySourceEventID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveLedgerEntry], error) {
	if s == nil {
		return operation.Page[entity.LeaveLedgerEntry]{}, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.List(ctx, []domainrepo.Filter{{Column: "source_event_id", Values: []any{value}}}, limit, offset)
}

// FindLeaveLedgerEntryByReasonCode menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveLedgerEntryByReasonCode(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveLedgerEntry], error) {
	if s == nil {
		return operation.Page[entity.LeaveLedgerEntry]{}, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.List(ctx, []domainrepo.Filter{{Column: "reason_code", Values: []any{value}}}, limit, offset)
}

// FindLeaveLedgerEntryByActorPrincipalID menjalankan operasi aplikasi dengan validasi dan scope server.
func (s *Service) FindLeaveLedgerEntryByActorPrincipalID(ctx context.Context, value string, limit, offset int) (operation.Page[entity.LeaveLedgerEntry], error) {
	if s == nil {
		return operation.Page[entity.LeaveLedgerEntry]{}, hrerrors.ErrUnavailable
	}
	return s.leaveLedgerEntry.List(ctx, []domainrepo.Filter{{Column: "actor_principal_id", Values: []any{value}}}, limit, offset)
}
