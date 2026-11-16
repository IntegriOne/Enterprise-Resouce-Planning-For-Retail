package leave

import (
	"context"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Operasi bernama objek tidak mengubah aturan append-only ledger pada database.
func (r *LeaveLedgerEntryRepository) InsertLeaveLedgerEntries(ctx context.Context, value *entity.LeaveLedgerEntry) error {
	return r.Insert(ctx, value)
}
func (r *LeaveLedgerEntryRepository) UpdateLeaveLedgerEntries(ctx context.Context, value *entity.LeaveLedgerEntry) error {
	return r.Update(ctx, value)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByUUID(ctx context.Context, uuid string) (*entity.LeaveLedgerEntry, error) {
	return r.FindByID(ctx, uuid)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByUUIDForUpdate(ctx context.Context, uuid string) (*entity.LeaveLedgerEntry, error) {
	return r.FindByIDForUpdate(ctx, uuid)
}
func (r *LeaveLedgerEntryRepository) DeleteLeaveLedgerEntriesByUUID(ctx context.Context, uuid string) error {
	return r.DeleteByID(ctx, uuid)
}
func (r *LeaveLedgerEntryRepository) GetAllLeaveLedgerEntries(ctx context.Context, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.GetAll(ctx, limit, offset)
}
func (r *LeaveLedgerEntryRepository) CountLeaveLedgerEntries(ctx context.Context) (int64, error) {
	return r.Count(ctx)
}
func (r *LeaveLedgerEntryRepository) SelectLeaveLedgerEntriesWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.SelectWhere(ctx, filters, limit, offset)
}
func (r *LeaveLedgerEntryRepository) CountLeaveLedgerEntriesWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	return r.CountWhere(ctx, filters)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByEntitlementID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByEntitlementID(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByEntryType(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByEntryType(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByAdjustmentID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByAdjustmentID(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByReversesEntryID(ctx context.Context, value *string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByReversesEntryID(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesBySourceManagement(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindBySourceManagement(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesBySourceEventID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindBySourceEventID(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByReasonCode(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByReasonCode(ctx, value, limit, offset)
}
func (r *LeaveLedgerEntryRepository) FindLeaveLedgerEntriesByActorPrincipalID(ctx context.Context, value string, limit, offset int) ([]entity.LeaveLedgerEntry, error) {
	return r.FindByActorPrincipalID(ctx, value, limit, offset)
}
