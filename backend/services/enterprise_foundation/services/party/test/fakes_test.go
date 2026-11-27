package party_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type seqIDGen struct{ seq int }

func (g *seqIDGen) NewUUID() (string, error) {
	g.seq++
	return fmt.Sprintf("cccccccc-cccc-4ccc-8ccc-%012d", g.seq), nil
}

type noopTransactor struct{}

func (noopTransactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type stubOutboxRepo struct {
	events []*entity.OutboxEvent
}

func (r *stubOutboxRepo) Insert(_ context.Context, event *entity.OutboxEvent) error {
	r.events = append(r.events, event)
	return nil
}
func (r *stubOutboxRepo) SelectByID(context.Context, string) (*entity.OutboxEvent, error) {
	return nil, eferrors.ErrOutboxEventNotFound
}
func (r *stubOutboxRepo) SelectByAggregate(context.Context, string, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.OutboxEvent], error) {
	return domainrepo.PageResult[*entity.OutboxEvent]{}, nil
}
func (r *stubOutboxRepo) SelectDuePending(context.Context, time.Time, int) ([]*entity.OutboxEvent, error) {
	return nil, nil
}
func (r *stubOutboxRepo) UpdateEffective(context.Context, string, time.Time) error { return nil }
func (r *stubOutboxRepo) UpdateRetry(context.Context, string, int, time.Time, string) error {
	return nil
}
func (r *stubOutboxRepo) UpdateFailed(context.Context, string, string) error { return nil }
func (r *stubOutboxRepo) CountPending(context.Context) (int64, error)        { return 0, nil }
func (r *stubOutboxRepo) CountFailed(context.Context) (int64, error)         { return 0, nil }
func (r *stubOutboxRepo) DeleteEffectiveBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type stubAddressRepo struct {
	items map[string]*entity.Address
}

func (r *stubAddressRepo) SelectByID(_ context.Context, id string) (*entity.Address, error) {
	if addr, ok := r.items[id]; ok {
		return addr.Clone(), nil
	}
	return nil, eferrors.ErrAddressNotFound
}
func (r *stubAddressRepo) SelectByIDs(context.Context, []string) ([]*entity.Address, error) {
	return nil, nil
}
func (r *stubAddressRepo) SelectList(context.Context, domainrepo.AddressFilter) (domainrepo.PageResult[*entity.Address], error) {
	return domainrepo.PageResult[*entity.Address]{}, nil
}
func (r *stubAddressRepo) CheckReferenced(context.Context, string) (bool, error) { return false, nil }
func (r *stubAddressRepo) Insert(context.Context, *entity.Address) error         { return nil }
func (r *stubAddressRepo) Update(context.Context, *entity.Address) error         { return nil }
func (r *stubAddressRepo) DeleteByID(context.Context, string) error              { return nil }

type memPartyRepo struct {
	order       []string
	items       map[string]*entity.Party
	roles       *memPartyRoleRepo
	insertCalls int
}

func newMemPartyRepo(roles *memPartyRoleRepo) *memPartyRepo {
	return &memPartyRepo{items: make(map[string]*entity.Party), roles: roles}
}

func (r *memPartyRepo) seed(p *entity.Party) {
	r.order = append(r.order, p.ID)
	r.items[p.ID] = p.Clone()
}

func (r *memPartyRepo) SelectByID(_ context.Context, id string) (*entity.Party, error) {
	p, ok := r.items[id]
	if !ok || p.IsDeleted() {
		return nil, eferrors.ErrPartyNotFound
	}
	return p.Clone(), nil
}

func (r *memPartyRepo) SelectByIDs(_ context.Context, ids []string) ([]*entity.Party, error) {
	out := make([]*entity.Party, 0, len(ids))
	for _, id := range ids {
		if p, ok := r.items[id]; ok && !p.IsDeleted() {
			out = append(out, p.Clone())
		}
	}
	return out, nil
}

func (r *memPartyRepo) SelectByCode(_ context.Context, code string) (*entity.Party, error) {
	for _, id := range r.order {
		p := r.items[id]
		if !p.IsDeleted() && p.PartnerCode == code {
			return p.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyNotFound
}

func (r *memPartyRepo) SelectByTaxIDNumber(_ context.Context, taxID string) (*entity.Party, error) {
	for _, id := range r.order {
		p := r.items[id]
		if !p.IsDeleted() && p.TaxIDNumber != nil && *p.TaxIDNumber == taxID {
			return p.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyNotFound
}

func paginateMem[T any](matched []T, page domainrepo.PageRequest) domainrepo.PageResult[T] {
	norm := page.Normalize()
	total := int64(len(matched))
	if norm.Offset >= len(matched) {
		return domainrepo.NewPageResult([]T{}, total, norm)
	}
	end := norm.Offset + norm.Limit
	if end > len(matched) {
		end = len(matched)
	}
	return domainrepo.NewPageResult(matched[norm.Offset:end], total, norm)
}

func (r *memPartyRepo) SelectByEmail(_ context.Context, email string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Party], error) {
	var out []*entity.Party
	for _, id := range r.order {
		p := r.items[id]
		if !p.IsDeleted() && p.Email != nil && strings.EqualFold(*p.Email, email) {
			out = append(out, p.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memPartyRepo) SelectList(_ context.Context, filter domainrepo.PartyFilter) (domainrepo.PageResult[*entity.Party], error) {
	norm := filter.Normalize()
	var matched []*entity.Party
	q := strings.ToLower(strings.TrimSpace(norm.Query))
	for _, id := range r.order {
		p := r.items[id]
		if !norm.IncludeDeleted && p.IsDeleted() {
			continue
		}
		if norm.Kind != nil && p.Kind != *norm.Kind {
			continue
		}
		if norm.Status != nil && p.Status != *norm.Status {
			continue
		}
		if norm.Role != nil && r.roles != nil && !r.roles.hasActiveRole(p.ID, *norm.Role) {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(p.PartnerCode+" "+p.DisplayName), q) {
			continue
		}
		matched = append(matched, p.Clone())
	}
	return paginateMem(matched, norm.Page), nil
}

func (r *memPartyRepo) CheckCodeExists(_ context.Context, code, excludeID string) (bool, error) {
	for _, id := range r.order {
		p := r.items[id]
		if !p.IsDeleted() && p.ID != excludeID && p.PartnerCode == code {
			return true, nil
		}
	}
	return false, nil
}

func (r *memPartyRepo) CheckTaxIDExists(_ context.Context, taxID, excludeID string) (bool, error) {
	for _, id := range r.order {
		p := r.items[id]
		if !p.IsDeleted() && p.ID != excludeID && p.TaxIDNumber != nil && *p.TaxIDNumber == taxID {
			return true, nil
		}
	}
	return false, nil
}

func (r *memPartyRepo) Insert(_ context.Context, p *entity.Party) error {
	r.insertCalls++
	r.seed(p)
	return nil
}

func (r *memPartyRepo) Update(_ context.Context, p *entity.Party) error {
	if _, ok := r.items[p.ID]; !ok {
		return eferrors.ErrPartyNotFound
	}
	r.items[p.ID] = p.Clone()
	return nil
}

func (r *memPartyRepo) SoftDeleteByID(_ context.Context, id string, deletedAt time.Time) error {
	p, ok := r.items[id]
	if !ok {
		return eferrors.ErrPartyNotFound
	}
	p.DeletedAt = &deletedAt
	p.Status = entity.PartyStatusInactive
	p.UpdatedAt = deletedAt
	return nil
}

type memPartyRoleRepo struct {
	order       []string
	items       map[string]*entity.PartyRole
	insertCalls int
}

func newMemPartyRoleRepo() *memPartyRoleRepo {
	return &memPartyRoleRepo{items: make(map[string]*entity.PartyRole)}
}

func (r *memPartyRoleRepo) hasActiveRole(partyID string, kind entity.PartyRoleKind) bool {
	for _, id := range r.order {
		role := r.items[id]
		if role.PartyID == partyID && role.Role == kind && role.IsActive {
			return true
		}
	}
	return false
}

func (r *memPartyRoleRepo) SelectByID(_ context.Context, id string) (*entity.PartyRole, error) {
	role, ok := r.items[id]
	if !ok {
		return nil, eferrors.ErrPartyRoleNotFound
	}
	return role.Clone(), nil
}

func (r *memPartyRoleRepo) SelectList(_ context.Context, filter domainrepo.PartyRoleFilter) (domainrepo.PageResult[*entity.PartyRole], error) {
	norm := filter.Normalize()
	q := strings.ToLower(strings.TrimSpace(norm.Query))
	var matched []*entity.PartyRole
	for _, id := range r.order {
		role := r.items[id]
		if norm.PartyID != "" && role.PartyID != norm.PartyID {
			continue
		}
		if norm.Role != nil && role.Role != *norm.Role {
			continue
		}
		if norm.ActiveOnly && !role.IsActive {
			continue
		}
		if q != "" && (role.ReferenceNote == nil || !strings.Contains(strings.ToLower(*role.ReferenceNote), q)) {
			continue
		}
		matched = append(matched, role.Clone())
	}
	return paginateMem(matched, norm.Page), nil
}

func (r *memPartyRoleRepo) SelectByPartyID(_ context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	var out []*entity.PartyRole
	for _, id := range r.order {
		if role := r.items[id]; role.PartyID == partyID {
			out = append(out, role.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memPartyRoleRepo) SelectByPartyAndRole(_ context.Context, partyID string, kind entity.PartyRoleKind) (*entity.PartyRole, error) {
	for _, id := range r.order {
		if role := r.items[id]; role.PartyID == partyID && role.Role == kind {
			return role.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyRoleNotFound
}

func (r *memPartyRoleRepo) SelectByRoleKind(_ context.Context, kind entity.PartyRoleKind, activeOnly bool, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	var matched []*entity.PartyRole
	for _, id := range r.order {
		role := r.items[id]
		if role.Role != kind {
			continue
		}
		if activeOnly && !role.IsActive {
			continue
		}
		matched = append(matched, role.Clone())
	}
	return paginateMem(matched, page), nil
}

func (r *memPartyRoleRepo) CheckExists(_ context.Context, partyID string, kind entity.PartyRoleKind) (bool, error) {
	for _, id := range r.order {
		if role := r.items[id]; role.PartyID == partyID && role.Role == kind {
			return true, nil
		}
	}
	return false, nil
}

func (r *memPartyRoleRepo) Insert(_ context.Context, role *entity.PartyRole) error {
	r.insertCalls++
	r.order = append(r.order, role.ID)
	r.items[role.ID] = role.Clone()
	return nil
}

func (r *memPartyRoleRepo) Update(_ context.Context, role *entity.PartyRole) error {
	if _, ok := r.items[role.ID]; !ok {
		return eferrors.ErrPartyRoleNotFound
	}
	r.items[role.ID] = role.Clone()
	return nil
}

func (r *memPartyRoleRepo) DeleteByID(_ context.Context, id string) error {
	if _, ok := r.items[id]; !ok {
		return eferrors.ErrPartyRoleNotFound
	}
	delete(r.items, id)
	return nil
}

type memPartyContactRepo struct {
	order       []string
	items       map[string]*entity.PartyContact
	insertCalls int
}

func newMemPartyContactRepo() *memPartyContactRepo {
	return &memPartyContactRepo{items: make(map[string]*entity.PartyContact)}
}

func (r *memPartyContactRepo) SelectByID(_ context.Context, id string) (*entity.PartyContact, error) {
	c, ok := r.items[id]
	if !ok || c.IsDeleted() {
		return nil, eferrors.ErrPartyContactNotFound
	}
	return c.Clone(), nil
}

func (r *memPartyContactRepo) SelectByPartyID(_ context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	var out []*entity.PartyContact
	for _, id := range r.order {
		c := r.items[id]
		if !c.IsDeleted() && c.PartyID == partyID {
			out = append(out, c.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memPartyContactRepo) SelectPrimaryByPartyID(_ context.Context, partyID string) (*entity.PartyContact, error) {
	for _, id := range r.order {
		c := r.items[id]
		if !c.IsDeleted() && c.PartyID == partyID && c.IsPrimary {
			return c.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyContactNotFound
}

func (r *memPartyContactRepo) SelectByIAMUserID(_ context.Context, iamUserID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	var out []*entity.PartyContact
	for _, id := range r.order {
		c := r.items[id]
		if !c.IsDeleted() && c.IAMUserID != nil && *c.IAMUserID == iamUserID {
			out = append(out, c.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memPartyContactRepo) SelectByPartyAndIAMUserID(_ context.Context, partyID, iamUserID string) (*entity.PartyContact, error) {
	for _, id := range r.order {
		c := r.items[id]
		if !c.IsDeleted() && c.PartyID == partyID && c.IAMUserID != nil && *c.IAMUserID == iamUserID {
			return c.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyContactNotFound
}

func (r *memPartyContactRepo) SelectByEmail(_ context.Context, email string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	var out []*entity.PartyContact
	for _, id := range r.order {
		c := r.items[id]
		if !c.IsDeleted() && c.Email != nil && strings.EqualFold(*c.Email, email) {
			out = append(out, c.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memPartyContactRepo) SelectList(_ context.Context, filter domainrepo.PartyContactFilter) (domainrepo.PageResult[*entity.PartyContact], error) {
	norm := filter.Normalize()
	var matched []*entity.PartyContact
	q := strings.ToLower(strings.TrimSpace(norm.Query))
	for _, id := range r.order {
		c := r.items[id]
		if !norm.IncludeDeleted && c.IsDeleted() {
			continue
		}
		if norm.PartyID != "" && c.PartyID != norm.PartyID {
			continue
		}
		if norm.PortalAccessStatus != nil && c.PortalAccessStatus != *norm.PortalAccessStatus {
			continue
		}
		if norm.ActiveOnly && !c.IsActive {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(c.FullName), q) {
			continue
		}
		matched = append(matched, c.Clone())
	}
	return paginateMem(matched, norm.Page), nil
}

func (r *memPartyContactRepo) CheckIAMUserExistsInParty(_ context.Context, partyID, iamUserID, excludeID string) (bool, error) {
	for _, id := range r.order {
		c := r.items[id]
		if !c.IsDeleted() && c.ID != excludeID && c.PartyID == partyID && c.IAMUserID != nil && *c.IAMUserID == iamUserID {
			return true, nil
		}
	}
	return false, nil
}

func (r *memPartyContactRepo) UpdateClearPrimaryByPartyExcept(_ context.Context, partyID, keepID string, updatedAt time.Time) error {
	for _, id := range r.order {
		c := r.items[id]
		if !c.IsDeleted() && c.PartyID == partyID && c.ID != keepID && c.IsPrimary {
			c.IsPrimary = false
			c.UpdatedAt = updatedAt
		}
	}
	return nil
}

func (r *memPartyContactRepo) Insert(_ context.Context, c *entity.PartyContact) error {
	r.insertCalls++
	r.order = append(r.order, c.ID)
	r.items[c.ID] = c.Clone()
	return nil
}

func (r *memPartyContactRepo) Update(_ context.Context, c *entity.PartyContact) error {
	if _, ok := r.items[c.ID]; !ok {
		return eferrors.ErrPartyContactNotFound
	}
	r.items[c.ID] = c.Clone()
	return nil
}

func (r *memPartyContactRepo) SoftDeleteByID(_ context.Context, id string, deletedAt time.Time) error {
	c, ok := r.items[id]
	if !ok {
		return eferrors.ErrPartyContactNotFound
	}
	c.DeletedAt = &deletedAt
	c.IsActive = false
	c.UpdatedAt = deletedAt
	return nil
}

type memPartyAddressRepo struct {
	order       []string
	items       map[string]*entity.PartyAddress
	insertCalls int
}

func newMemPartyAddressRepo() *memPartyAddressRepo {
	return &memPartyAddressRepo{items: make(map[string]*entity.PartyAddress)}
}

func (r *memPartyAddressRepo) SelectByID(_ context.Context, id string) (*entity.PartyAddress, error) {
	pa, ok := r.items[id]
	if !ok {
		return nil, eferrors.ErrPartyAddressNotFound
	}
	return pa.Clone(), nil
}

func (r *memPartyAddressRepo) SelectList(_ context.Context, filter domainrepo.PartyAddressFilter) (domainrepo.PageResult[*entity.PartyAddress], error) {
	norm := filter.Normalize()
	var matched []*entity.PartyAddress
	for _, id := range r.order {
		pa, ok := r.items[id]
		if !ok {
			continue
		}
		if norm.PartyID != "" && pa.PartyID != norm.PartyID {
			continue
		}
		if norm.AddressID != "" && pa.AddressID != norm.AddressID {
			continue
		}
		if norm.Usage != nil && pa.Usage != *norm.Usage {
			continue
		}
		if norm.PrimaryOnly != nil && pa.IsPrimary != *norm.PrimaryOnly {
			continue
		}
		matched = append(matched, pa.Clone())
	}
	return paginateMem(matched, norm.Page), nil
}

func (r *memPartyAddressRepo) SelectByPartyID(_ context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	var out []*entity.PartyAddress
	for _, id := range r.order {
		if pa, ok := r.items[id]; ok && pa.PartyID == partyID {
			out = append(out, pa.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memPartyAddressRepo) SelectByPartyAndUsage(_ context.Context, partyID string, usage entity.AddressUsageKind, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	var out []*entity.PartyAddress
	for _, id := range r.order {
		if pa, ok := r.items[id]; ok && pa.PartyID == partyID && pa.Usage == usage {
			out = append(out, pa.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memPartyAddressRepo) SelectPrimaryByPartyAndUsage(_ context.Context, partyID string, usage entity.AddressUsageKind) (*entity.PartyAddress, error) {
	for _, id := range r.order {
		if pa, ok := r.items[id]; ok && pa.PartyID == partyID && pa.Usage == usage && pa.IsPrimary {
			return pa.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyAddressNotFound
}

func (r *memPartyAddressRepo) SelectByAddressID(_ context.Context, addressID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	var out []*entity.PartyAddress
	for _, id := range r.order {
		if pa, ok := r.items[id]; ok && pa.AddressID == addressID {
			out = append(out, pa.Clone())
		}
	}
	return paginateMem(out, page), nil
}

func (r *memPartyAddressRepo) SelectByPartyAddressUsage(_ context.Context, partyID, addressID string, usage entity.AddressUsageKind) (*entity.PartyAddress, error) {
	for _, id := range r.order {
		if pa, ok := r.items[id]; ok && pa.PartyID == partyID && pa.AddressID == addressID && pa.Usage == usage {
			return pa.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyAddressNotFound
}

func (r *memPartyAddressRepo) UpdateClearPrimaryByPartyUsageExcept(_ context.Context, partyID string, usage entity.AddressUsageKind, keepID string, updatedAt time.Time) error {
	for _, id := range r.order {
		if pa, ok := r.items[id]; ok && pa.PartyID == partyID && pa.Usage == usage && pa.ID != keepID && pa.IsPrimary {
			pa.IsPrimary = false
			pa.UpdatedAt = updatedAt
		}
	}
	return nil
}

func (r *memPartyAddressRepo) Insert(_ context.Context, link *entity.PartyAddress) error {
	r.insertCalls++
	r.order = append(r.order, link.ID)
	r.items[link.ID] = link.Clone()
	return nil
}

func (r *memPartyAddressRepo) Update(_ context.Context, link *entity.PartyAddress) error {
	if _, ok := r.items[link.ID]; !ok {
		return eferrors.ErrPartyAddressNotFound
	}
	r.items[link.ID] = link.Clone()
	return nil
}

func (r *memPartyAddressRepo) DeleteByID(_ context.Context, id string) error {
	if _, ok := r.items[id]; !ok {
		return eferrors.ErrPartyAddressNotFound
	}
	delete(r.items, id)
	return nil
}
