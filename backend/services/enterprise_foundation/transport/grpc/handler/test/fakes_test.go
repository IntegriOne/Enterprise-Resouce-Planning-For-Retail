package handler_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	efv1 "github.com/integrione/erp-retail/backend/services/enterprise_foundation/contract/gen/enterprise_foundation/v1"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/entity"
	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/repository"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/services"
	transportgrpc "github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc"
	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/transport/grpc/handler"
	"github.com/integrione/erp-retail/backend/shared/authcontext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const testBufSize = 1024 * 1024

var testHMACSecret = []byte("test-enterprise-foundation-hmac-secret-32b")

type testHarness struct {
	server        *transportgrpc.Server
	conn          *grpc.ClientConn
	addresses     efv1.AddressServiceClient
	legalEntities efv1.LegalEntityServiceClient
	locations     efv1.LocationServiceClient
	orgUnits      efv1.OrganizationUnitServiceClient
	parties       efv1.PartyServiceClient
	outbox        *memOutboxRepo
	hmacSecret    []byte
	now           time.Time
}

func newTestHarness(t *testing.T) *testHarness {
	t.Helper()

	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	outboxRepo := &memOutboxRepo{}
	addrRepo := newMemAddressRepo()
	leRepo := newMemLegalEntityRepo()
	locRepo := newMemLocationRepo()
	ouRepo := newMemOrgUnitRepo()
	partyRepo := newMemPartyRepo()
	roleRepo := newMemPartyRoleRepo()
	contactRepo := newMemPartyContactRepo()
	paRepo := newMemPartyAddressRepo()

	addrRepo.legalEntities = leRepo
	addrRepo.locations = locRepo
	addrRepo.partyAddresses = paRepo
	partyRepo.roles = roleRepo

	repoReg := &repository.Registry{
		Transactor:        noopTransactor{},
		Addresses:         addrRepo,
		LegalEntities:     leRepo,
		Locations:         locRepo,
		OrganizationUnits: ouRepo,
		Parties:           partyRepo,
		PartyRoles:        roleRepo,
		PartyContacts:     contactRepo,
		PartyAddresses:    paRepo,
		Outbox:            outboxRepo,
	}

	svcReg, err := services.NewRegistry(repoReg, services.Options{
		Clock: fixedClock{now: now},
		IDGen: &seqUUIDGen{},
	})
	if err != nil {
		t.Fatalf("services.NewRegistry failed: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := transportgrpc.NewServer(transportgrpc.ServerConfig{
		ServiceName:               "integrione-enterprise-foundation",
		Address:                   "bufnet",
		EnableReflection:          true,
		InternalGatewayHMACSecret: testHMACSecret,
		ShutdownTimeout:           5 * time.Second,
		Logger:                    logger,
	}, handler.Dependencies{
		Logger:   logger,
		Services: svcReg,
	})
	if err != nil {
		t.Fatalf("transportgrpc.NewServer failed: %v", err)
	}

	lis := bufconn.Listen(testBufSize)
	go func() {
		_ = srv.Serve(lis)
	}()
	t.Cleanup(func() {
		_ = srv.Stop(context.Background())
		_ = lis.Close()
	})

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient failed: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})

	return &testHarness{
		server:        srv,
		conn:          conn,
		addresses:     efv1.NewAddressServiceClient(conn),
		legalEntities: efv1.NewLegalEntityServiceClient(conn),
		locations:     efv1.NewLocationServiceClient(conn),
		orgUnits:      efv1.NewOrganizationUnitServiceClient(conn),
		parties:       efv1.NewPartyServiceClient(conn),
		outbox:        outboxRepo,
		hmacSecret:    testHMACSecret,
		now:           now,
	}
}

func (h *testHarness) signedCtx(t *testing.T) context.Context {
	t.Helper()
	caller := authcontext.CallerContext{
		PrincipalID:   "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		SessionID:     "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		UserKind:      "employee",
		AuthEpoch:     1,
		PermRevision:  1,
		RequestID:     "req-test-1",
		ClientIP:      "127.0.0.1",
		UserAgent:     "ef-bufconn-test",
		Authenticated: true,
	}
	return authcontext.AppendToOutgoingGRPCContext(context.Background(), caller, h.hmacSecret, time.Now().UTC())
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type seqUUIDGen struct{ seq atomic.Uint64 }

func (g *seqUUIDGen) NewUUID() (string, error) {
	n := g.seq.Add(1)
	return "10000000-0000-4000-8000-" + padUint12(n), nil
}

func padUint12(n uint64) string {
	digits := make([]byte, 12)
	for i := 11; i >= 0; i-- {
		digits[i] = byte('0' + (n % 10))
		n /= 10
	}
	return string(digits)
}

type noopTransactor struct{}

func (noopTransactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func paginateSlice[T any](items []T, page domainrepo.PageRequest) domainrepo.PageResult[T] {
	norm := page.Normalize()
	total := int64(len(items))
	if norm.Offset >= len(items) {
		return domainrepo.NewPageResult([]T{}, total, norm)
	}
	end := norm.Offset + norm.Limit
	if end > len(items) {
		end = len(items)
	}
	return domainrepo.NewPageResult(items[norm.Offset:end], total, norm)
}

type memOutboxRepo struct {
	mu     sync.RWMutex
	events []*entity.OutboxEvent
}

func (r *memOutboxRepo) Insert(_ context.Context, event *entity.OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
	return nil
}
func (r *memOutboxRepo) SelectByID(context.Context, string) (*entity.OutboxEvent, error) {
	return nil, eferrors.ErrOutboxEventNotFound
}
func (r *memOutboxRepo) SelectByAggregate(context.Context, string, string, domainrepo.PageRequest) (domainrepo.PageResult[*entity.OutboxEvent], error) {
	return domainrepo.PageResult[*entity.OutboxEvent]{}, nil
}
func (r *memOutboxRepo) SelectDuePending(context.Context, time.Time, int) ([]*entity.OutboxEvent, error) {
	return nil, nil
}
func (r *memOutboxRepo) UpdateEffective(context.Context, string, time.Time) error { return nil }
func (r *memOutboxRepo) UpdateRetry(context.Context, string, int, time.Time, string) error {
	return nil
}
func (r *memOutboxRepo) UpdateFailed(context.Context, string, string) error { return nil }
func (r *memOutboxRepo) CountPending(context.Context) (int64, error)        { return 0, nil }
func (r *memOutboxRepo) CountFailed(context.Context) (int64, error)         { return 0, nil }
func (r *memOutboxRepo) DeleteEffectiveBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type memAddressRepo struct {
	mu             sync.RWMutex
	order          []string
	items          map[string]*entity.Address
	legalEntities  *memLegalEntityRepo
	locations      *memLocationRepo
	partyAddresses *memPartyAddressRepo
}

func newMemAddressRepo() *memAddressRepo {
	return &memAddressRepo{items: make(map[string]*entity.Address)}
}

func (r *memAddressRepo) SelectByID(_ context.Context, id string) (*entity.Address, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if item, ok := r.items[id]; ok {
		return item.Clone(), nil
	}
	return nil, eferrors.ErrAddressNotFound
}

func (r *memAddressRepo) SelectByIDs(_ context.Context, ids []string) ([]*entity.Address, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entity.Address, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			out = append(out, item.Clone())
		}
	}
	return out, nil
}

func (r *memAddressRepo) SelectList(_ context.Context, filter domainrepo.AddressFilter) (domainrepo.PageResult[*entity.Address], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := filter.Normalize()
	q := strings.ToLower(f.Query)
	matched := make([]*entity.Address, 0, len(r.order))
	for _, id := range r.order {
		a, ok := r.items[id]
		if !ok {
			continue
		}
		if f.CountryCode != "" && !strings.EqualFold(a.CountryCode, f.CountryCode) {
			continue
		}
		if f.City != "" && !strings.EqualFold(a.City, f.City) {
			continue
		}
		if f.Province != "" && (a.Province == nil || !strings.EqualFold(*a.Province, f.Province)) {
			continue
		}
		if f.PostalCode != "" && (a.PostalCode == nil || !strings.EqualFold(*a.PostalCode, f.PostalCode)) {
			continue
		}
		if q != "" {
			hay := strings.ToLower(a.AddressLine1 + " " + a.City + " " + a.CountryCode)
			if a.Label != nil {
				hay += " " + strings.ToLower(*a.Label)
			}
			if !strings.Contains(hay, q) {
				continue
			}
		}
		matched = append(matched, a.Clone())
	}
	return paginateSlice(matched, f.Page), nil
}

func (r *memAddressRepo) CheckReferenced(_ context.Context, id string) (bool, error) {
	if r.legalEntities != nil && r.legalEntities.referencesAddress(id) {
		return true, nil
	}
	if r.locations != nil && r.locations.referencesAddress(id) {
		return true, nil
	}
	if r.partyAddresses != nil && r.partyAddresses.referencesAddress(id) {
		return true, nil
	}
	return false, nil
}

func (r *memAddressRepo) Insert(_ context.Context, addr *entity.Address) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.order = append(r.order, addr.ID)
	r.items[addr.ID] = addr.Clone()
	return nil
}

func (r *memAddressRepo) Update(_ context.Context, addr *entity.Address) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[addr.ID]; !ok {
		return eferrors.ErrAddressNotFound
	}
	r.items[addr.ID] = addr.Clone()
	return nil
}

func (r *memAddressRepo) DeleteByID(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return eferrors.ErrAddressNotFound
	}
	delete(r.items, id)
	return nil
}

type memLegalEntityRepo struct {
	mu    sync.RWMutex
	order []string
	items map[string]*entity.LegalEntity
}

func newMemLegalEntityRepo() *memLegalEntityRepo {
	return &memLegalEntityRepo{items: make(map[string]*entity.LegalEntity)}
}

func (r *memLegalEntityRepo) referencesAddress(addrID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, le := range r.items {
		if !le.IsDeleted() && le.RegisteredAddressID != nil && *le.RegisteredAddressID == addrID {
			return true
		}
	}
	return false
}

func (r *memLegalEntityRepo) SelectByID(_ context.Context, id string) (*entity.LegalEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if item, ok := r.items[id]; ok && !item.IsDeleted() {
		return item.Clone(), nil
	}
	return nil, eferrors.ErrLegalEntityNotFound
}

func (r *memLegalEntityRepo) SelectByIDs(_ context.Context, ids []string) ([]*entity.LegalEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entity.LegalEntity, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok && !item.IsDeleted() {
			out = append(out, item.Clone())
		}
	}
	return out, nil
}

func (r *memLegalEntityRepo) SelectByCode(_ context.Context, code string) (*entity.LegalEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if le := r.items[id]; le != nil && !le.IsDeleted() && strings.EqualFold(le.EntityCode, code) {
			return le.Clone(), nil
		}
	}
	return nil, eferrors.ErrLegalEntityNotFound
}

func (r *memLegalEntityRepo) SelectByTaxIDNumber(_ context.Context, npwp string) (*entity.LegalEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if le := r.items[id]; le != nil && !le.IsDeleted() && le.TaxIDNumber != nil && *le.TaxIDNumber == npwp {
			return le.Clone(), nil
		}
	}
	return nil, eferrors.ErrLegalEntityNotFound
}

func (r *memLegalEntityRepo) SelectByNIB(_ context.Context, nib string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.LegalEntity, 0)
	for _, id := range r.order {
		if le := r.items[id]; le != nil && !le.IsDeleted() && le.NIB != nil && *le.NIB == nib {
			matched = append(matched, le.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memLegalEntityRepo) SelectPrimaryOperating(context.Context) (*entity.LegalEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if le := r.items[id]; le != nil && !le.IsDeleted() && le.IsPrimaryOperatingEntity {
			return le.Clone(), nil
		}
	}
	return nil, eferrors.ErrLegalEntityNotFound
}

func (r *memLegalEntityRepo) SelectList(_ context.Context, filter domainrepo.LegalEntityFilter) (domainrepo.PageResult[*entity.LegalEntity], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := filter.Normalize()
	q := strings.ToLower(f.Query)
	matched := make([]*entity.LegalEntity, 0, len(r.order))
	for _, id := range r.order {
		le := r.items[id]
		if le == nil || (!f.IncludeDeleted && le.IsDeleted()) {
			continue
		}
		if f.ParentLegalEntityID != nil && (le.ParentLegalEntityID == nil || *le.ParentLegalEntityID != *f.ParentLegalEntityID) {
			continue
		}
		if f.Status != nil && le.Status != *f.Status {
			continue
		}
		if f.PrimaryOnly != nil && le.IsPrimaryOperatingEntity != *f.PrimaryOnly {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(le.EntityCode+" "+le.LegalName), q) {
			continue
		}
		matched = append(matched, le.Clone())
	}
	return paginateSlice(matched, f.Page), nil
}

func (r *memLegalEntityRepo) SelectParentID(_ context.Context, id string) (*string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	le, ok := r.items[id]
	if !ok || le.IsDeleted() {
		return nil, eferrors.ErrLegalEntityNotFound
	}
	if le.ParentLegalEntityID == nil {
		return nil, nil
	}
	cp := *le.ParentLegalEntityID
	return &cp, nil
}

func (r *memLegalEntityRepo) SelectChildrenByParentID(_ context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.LegalEntity], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.LegalEntity, 0)
	for _, id := range r.order {
		if le := r.items[id]; le != nil && !le.IsDeleted() && le.ParentLegalEntityID != nil && *le.ParentLegalEntityID == parentID {
			matched = append(matched, le.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memLegalEntityRepo) CountActiveDependents(_ context.Context, id string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, le := range r.items {
		if !le.IsDeleted() && le.ParentLegalEntityID != nil && *le.ParentLegalEntityID == id {
			count++
		}
	}
	return count, nil
}

func (r *memLegalEntityRepo) CheckCodeExists(_ context.Context, code, excludeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, le := range r.items {
		if !le.IsDeleted() && le.ID != excludeID && strings.EqualFold(le.EntityCode, code) {
			return true, nil
		}
	}
	return false, nil
}

func (r *memLegalEntityRepo) CheckTaxIDExists(_ context.Context, npwp, excludeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, le := range r.items {
		if !le.IsDeleted() && le.ID != excludeID && le.TaxIDNumber != nil && *le.TaxIDNumber == npwp {
			return true, nil
		}
	}
	return false, nil
}

func (r *memLegalEntityRepo) Insert(_ context.Context, le *entity.LegalEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.order = append(r.order, le.ID)
	r.items[le.ID] = le.Clone()
	return nil
}

func (r *memLegalEntityRepo) Update(_ context.Context, le *entity.LegalEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.items[le.ID]
	if !ok || existing.IsDeleted() {
		return eferrors.ErrLegalEntityNotFound
	}
	next := le.Clone()
	next.DeletedAt = existing.DeletedAt
	r.items[le.ID] = next
	return nil
}

func (r *memLegalEntityRepo) UpdateClearPrimaryOperatingExcept(_ context.Context, keepID string, updatedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, le := range r.items {
		if le.ID != keepID && le.IsPrimaryOperatingEntity {
			le.IsPrimaryOperatingEntity = false
			le.UpdatedAt = updatedAt
		}
	}
	return nil
}

func (r *memLegalEntityRepo) SoftDeleteByID(_ context.Context, id string, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	le, ok := r.items[id]
	if !ok || le.IsDeleted() {
		return eferrors.ErrLegalEntityNotFound
	}
	le.DeletedAt = &deletedAt
	return nil
}

type memLocationRepo struct {
	mu    sync.RWMutex
	order []string
	items map[string]*entity.Location
}

func newMemLocationRepo() *memLocationRepo {
	return &memLocationRepo{items: make(map[string]*entity.Location)}
}

func (r *memLocationRepo) referencesAddress(addrID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, loc := range r.items {
		if !loc.IsDeleted() && loc.AddressID != nil && *loc.AddressID == addrID {
			return true
		}
	}
	return false
}

func (r *memLocationRepo) SelectByID(_ context.Context, id string) (*entity.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if loc, ok := r.items[id]; ok && !loc.IsDeleted() {
		return loc.Clone(), nil
	}
	return nil, eferrors.ErrLocationNotFound
}

func (r *memLocationRepo) SelectByIDs(_ context.Context, ids []string) ([]*entity.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entity.Location, 0, len(ids))
	for _, id := range ids {
		if loc, ok := r.items[id]; ok && !loc.IsDeleted() {
			out = append(out, loc.Clone())
		}
	}
	return out, nil
}

func (r *memLocationRepo) SelectByCode(_ context.Context, code string) (*entity.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if loc := r.items[id]; loc != nil && !loc.IsDeleted() && strings.EqualFold(loc.LocationCode, code) {
			return loc.Clone(), nil
		}
	}
	return nil, eferrors.ErrLocationNotFound
}

func (r *memLocationRepo) SelectByLegalEntityAndKind(_ context.Context, leID string, kind entity.LocationKind, activeOnly bool, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.Location, 0)
	for _, id := range r.order {
		loc := r.items[id]
		if loc == nil || loc.IsDeleted() || loc.LegalEntityID != leID || loc.Kind != kind {
			continue
		}
		if activeOnly && !loc.IsActive {
			continue
		}
		matched = append(matched, loc.Clone())
	}
	return paginateSlice(matched, page), nil
}

func (r *memLocationRepo) SelectByAddressID(_ context.Context, addrID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.Location, 0)
	for _, id := range r.order {
		if loc := r.items[id]; loc != nil && !loc.IsDeleted() && loc.AddressID != nil && *loc.AddressID == addrID {
			matched = append(matched, loc.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memLocationRepo) SelectList(_ context.Context, filter domainrepo.LocationFilter) (domainrepo.PageResult[*entity.Location], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := filter.Normalize()
	q := strings.ToLower(f.Query)
	matched := make([]*entity.Location, 0)
	for _, id := range r.order {
		loc := r.items[id]
		if loc == nil || (!f.IncludeDeleted && loc.IsDeleted()) {
			continue
		}
		if f.LegalEntityID != "" && loc.LegalEntityID != f.LegalEntityID {
			continue
		}
		if f.ParentLocationID != nil && (loc.ParentLocationID == nil || *loc.ParentLocationID != *f.ParentLocationID) {
			continue
		}
		if f.AddressID != nil && (loc.AddressID == nil || *loc.AddressID != *f.AddressID) {
			continue
		}
		if f.Kind != nil && loc.Kind != *f.Kind {
			continue
		}
		if f.ActiveOnly && !loc.IsActive {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(loc.LocationCode+" "+loc.Name), q) {
			continue
		}
		matched = append(matched, loc.Clone())
	}
	return paginateSlice(matched, f.Page), nil
}

func (r *memLocationRepo) SelectParentID(_ context.Context, id string) (*string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	loc, ok := r.items[id]
	if !ok || loc.IsDeleted() {
		return nil, eferrors.ErrLocationNotFound
	}
	if loc.ParentLocationID == nil {
		return nil, nil
	}
	cp := *loc.ParentLocationID
	return &cp, nil
}

func (r *memLocationRepo) SelectChildrenByParentID(_ context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Location], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.Location, 0)
	for _, id := range r.order {
		if loc := r.items[id]; loc != nil && !loc.IsDeleted() && loc.ParentLocationID != nil && *loc.ParentLocationID == parentID {
			matched = append(matched, loc.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memLocationRepo) CountActiveDependents(_ context.Context, id string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, loc := range r.items {
		if !loc.IsDeleted() && loc.ParentLocationID != nil && *loc.ParentLocationID == id {
			count++
		}
	}
	return count, nil
}

func (r *memLocationRepo) CheckCodeExists(_ context.Context, code, excludeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, loc := range r.items {
		if !loc.IsDeleted() && loc.ID != excludeID && strings.EqualFold(loc.LocationCode, code) {
			return true, nil
		}
	}
	return false, nil
}

func (r *memLocationRepo) Insert(_ context.Context, loc *entity.Location) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.order = append(r.order, loc.ID)
	r.items[loc.ID] = loc.Clone()
	return nil
}

func (r *memLocationRepo) Update(_ context.Context, loc *entity.Location) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.items[loc.ID]
	if !ok || existing.IsDeleted() {
		return eferrors.ErrLocationNotFound
	}
	next := loc.Clone()
	next.DeletedAt = existing.DeletedAt
	r.items[loc.ID] = next
	return nil
}

func (r *memLocationRepo) SoftDeleteByID(_ context.Context, id string, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	loc, ok := r.items[id]
	if !ok || loc.IsDeleted() {
		return eferrors.ErrLocationNotFound
	}
	loc.DeletedAt = &deletedAt
	return nil
}

type memOrgUnitRepo struct {
	mu    sync.RWMutex
	order []string
	items map[string]*entity.OrganizationUnit
}

func newMemOrgUnitRepo() *memOrgUnitRepo {
	return &memOrgUnitRepo{items: make(map[string]*entity.OrganizationUnit)}
}

func (r *memOrgUnitRepo) SelectByID(_ context.Context, id string) (*entity.OrganizationUnit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.items[id]; ok && !u.IsDeleted() {
		return u.Clone(), nil
	}
	return nil, eferrors.ErrOrgUnitNotFound
}

func (r *memOrgUnitRepo) SelectByIDs(_ context.Context, ids []string) ([]*entity.OrganizationUnit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entity.OrganizationUnit, 0, len(ids))
	for _, id := range ids {
		if u, ok := r.items[id]; ok && !u.IsDeleted() {
			out = append(out, u.Clone())
		}
	}
	return out, nil
}

func (r *memOrgUnitRepo) SelectByEntityAndCode(_ context.Context, leID, code string) (*entity.OrganizationUnit, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if u := r.items[id]; u != nil && !u.IsDeleted() && u.LegalEntityID == leID && strings.EqualFold(u.UnitCode, code) {
			return u.Clone(), nil
		}
	}
	return nil, eferrors.ErrOrgUnitNotFound
}

func (r *memOrgUnitRepo) SelectByHeadPersonID(_ context.Context, headID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.OrganizationUnit, 0)
	for _, id := range r.order {
		if u := r.items[id]; u != nil && !u.IsDeleted() && u.HeadPersonID != nil && *u.HeadPersonID == headID {
			matched = append(matched, u.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memOrgUnitRepo) SelectByPrimaryLocationID(_ context.Context, locID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.OrganizationUnit, 0)
	for _, id := range r.order {
		if u := r.items[id]; u != nil && !u.IsDeleted() && u.PrimaryLocationID != nil && *u.PrimaryLocationID == locID {
			matched = append(matched, u.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memOrgUnitRepo) SelectList(_ context.Context, filter domainrepo.OrganizationUnitFilter) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := filter.Normalize()
	q := strings.ToLower(f.Query)
	matched := make([]*entity.OrganizationUnit, 0)
	for _, id := range r.order {
		u := r.items[id]
		if u == nil || (!f.IncludeDeleted && u.IsDeleted()) {
			continue
		}
		if f.LegalEntityID != "" && u.LegalEntityID != f.LegalEntityID {
			continue
		}
		if f.ParentUnitID != nil && (u.ParentUnitID == nil || *u.ParentUnitID != *f.ParentUnitID) {
			continue
		}
		if f.PrimaryLocationID != nil && (u.PrimaryLocationID == nil || *u.PrimaryLocationID != *f.PrimaryLocationID) {
			continue
		}
		if f.Kind != nil && u.Kind != *f.Kind {
			continue
		}
		if f.ActiveOnly && !u.IsActive {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(u.UnitCode+" "+u.Name), q) {
			continue
		}
		matched = append(matched, u.Clone())
	}
	return paginateSlice(matched, f.Page), nil
}

func (r *memOrgUnitRepo) SelectParentID(_ context.Context, id string) (*string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.items[id]
	if !ok || u.IsDeleted() {
		return nil, eferrors.ErrOrgUnitNotFound
	}
	if u.ParentUnitID == nil {
		return nil, nil
	}
	cp := *u.ParentUnitID
	return &cp, nil
}

func (r *memOrgUnitRepo) SelectChildrenByParentID(_ context.Context, parentID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.OrganizationUnit], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.OrganizationUnit, 0)
	for _, id := range r.order {
		if u := r.items[id]; u != nil && !u.IsDeleted() && u.ParentUnitID != nil && *u.ParentUnitID == parentID {
			matched = append(matched, u.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memOrgUnitRepo) CountActiveDependents(_ context.Context, id string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, u := range r.items {
		if !u.IsDeleted() && u.ParentUnitID != nil && *u.ParentUnitID == id {
			count++
		}
	}
	return count, nil
}

func (r *memOrgUnitRepo) CheckCodeExistsInEntity(_ context.Context, leID, code, excludeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.items {
		if !u.IsDeleted() && u.ID != excludeID && u.LegalEntityID == leID && strings.EqualFold(u.UnitCode, code) {
			return true, nil
		}
	}
	return false, nil
}

func (r *memOrgUnitRepo) Insert(_ context.Context, u *entity.OrganizationUnit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.order = append(r.order, u.ID)
	r.items[u.ID] = u.Clone()
	return nil
}

func (r *memOrgUnitRepo) Update(_ context.Context, u *entity.OrganizationUnit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.items[u.ID]
	if !ok || existing.IsDeleted() {
		return eferrors.ErrOrgUnitNotFound
	}
	next := u.Clone()
	next.DeletedAt = existing.DeletedAt
	r.items[u.ID] = next
	return nil
}

func (r *memOrgUnitRepo) SoftDeleteByID(_ context.Context, id string, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.items[id]
	if !ok || u.IsDeleted() {
		return eferrors.ErrOrgUnitNotFound
	}
	u.DeletedAt = &deletedAt
	return nil
}

type memPartyRepo struct {
	mu    sync.RWMutex
	order []string
	items map[string]*entity.Party
	roles *memPartyRoleRepo
}

func newMemPartyRepo() *memPartyRepo {
	return &memPartyRepo{items: make(map[string]*entity.Party)}
}

func (r *memPartyRepo) SelectByID(_ context.Context, id string) (*entity.Party, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, ok := r.items[id]; ok && !p.IsDeleted() {
		return p.Clone(), nil
	}
	return nil, eferrors.ErrPartyNotFound
}

func (r *memPartyRepo) SelectByIDs(_ context.Context, ids []string) ([]*entity.Party, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entity.Party, 0, len(ids))
	for _, id := range ids {
		if p, ok := r.items[id]; ok && !p.IsDeleted() {
			out = append(out, p.Clone())
		}
	}
	return out, nil
}

func (r *memPartyRepo) SelectByCode(_ context.Context, code string) (*entity.Party, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if p := r.items[id]; p != nil && !p.IsDeleted() && strings.EqualFold(p.PartnerCode, code) {
			return p.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyNotFound
}

func (r *memPartyRepo) SelectByTaxIDNumber(_ context.Context, npwp string) (*entity.Party, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if p := r.items[id]; p != nil && !p.IsDeleted() && p.TaxIDNumber != nil && *p.TaxIDNumber == npwp {
			return p.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyNotFound
}

func (r *memPartyRepo) SelectByEmail(_ context.Context, email string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.Party], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.Party, 0)
	for _, id := range r.order {
		if p := r.items[id]; p != nil && !p.IsDeleted() && p.Email != nil && strings.EqualFold(*p.Email, email) {
			matched = append(matched, p.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memPartyRepo) SelectList(_ context.Context, filter domainrepo.PartyFilter) (domainrepo.PageResult[*entity.Party], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := filter.Normalize()
	q := strings.ToLower(f.Query)
	matched := make([]*entity.Party, 0)
	for _, id := range r.order {
		p := r.items[id]
		if p == nil || (!f.IncludeDeleted && p.IsDeleted()) {
			continue
		}
		if f.Kind != nil && p.Kind != *f.Kind {
			continue
		}
		if f.Status != nil && p.Status != *f.Status {
			continue
		}
		if f.Role != nil && r.roles != nil && !r.roles.hasActiveRole(p.ID, *f.Role) {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(p.PartnerCode+" "+p.DisplayName), q) {
			continue
		}
		matched = append(matched, p.Clone())
	}
	return paginateSlice(matched, f.Page), nil
}

func (r *memPartyRepo) CheckCodeExists(_ context.Context, code, excludeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.items {
		if !p.IsDeleted() && p.ID != excludeID && strings.EqualFold(p.PartnerCode, code) {
			return true, nil
		}
	}
	return false, nil
}

func (r *memPartyRepo) CheckTaxIDExists(_ context.Context, npwp, excludeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.items {
		if !p.IsDeleted() && p.ID != excludeID && p.TaxIDNumber != nil && *p.TaxIDNumber == npwp {
			return true, nil
		}
	}
	return false, nil
}

func (r *memPartyRepo) Insert(_ context.Context, p *entity.Party) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.order = append(r.order, p.ID)
	r.items[p.ID] = p.Clone()
	return nil
}

func (r *memPartyRepo) Update(_ context.Context, p *entity.Party) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.items[p.ID]
	if !ok || existing.IsDeleted() {
		return eferrors.ErrPartyNotFound
	}
	next := p.Clone()
	next.DeletedAt = existing.DeletedAt
	r.items[p.ID] = next
	return nil
}

func (r *memPartyRepo) SoftDeleteByID(_ context.Context, id string, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.items[id]
	if !ok || p.IsDeleted() {
		return eferrors.ErrPartyNotFound
	}
	p.DeletedAt = &deletedAt
	return nil
}

type memPartyRoleRepo struct {
	mu    sync.RWMutex
	order []string
	items map[string]*entity.PartyRole
}

func newMemPartyRoleRepo() *memPartyRoleRepo {
	return &memPartyRoleRepo{items: make(map[string]*entity.PartyRole)}
}

func (r *memPartyRoleRepo) hasActiveRole(partyID string, kind entity.PartyRoleKind) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, role := range r.items {
		if role.PartyID == partyID && role.Role == kind && role.IsActive {
			return true
		}
	}
	return false
}

func (r *memPartyRoleRepo) SelectByID(_ context.Context, id string) (*entity.PartyRole, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if role, ok := r.items[id]; ok {
		return role.Clone(), nil
	}
	return nil, eferrors.ErrPartyRoleNotFound
}

func (r *memPartyRoleRepo) SelectByPartyID(_ context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.PartyRole, 0)
	for _, id := range r.order {
		if role := r.items[id]; role != nil && role.PartyID == partyID {
			matched = append(matched, role.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memPartyRoleRepo) SelectByPartyAndRole(_ context.Context, partyID string, kind entity.PartyRoleKind) (*entity.PartyRole, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if role := r.items[id]; role != nil && role.PartyID == partyID && role.Role == kind {
			return role.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyRoleNotFound
}

func (r *memPartyRoleRepo) SelectByRoleKind(_ context.Context, kind entity.PartyRoleKind, activeOnly bool, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyRole], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.PartyRole, 0)
	for _, id := range r.order {
		role := r.items[id]
		if role == nil || role.Role != kind {
			continue
		}
		if activeOnly && !role.IsActive {
			continue
		}
		matched = append(matched, role.Clone())
	}
	return paginateSlice(matched, page), nil
}

func (r *memPartyRoleRepo) SelectList(_ context.Context, filter domainrepo.PartyRoleFilter) (domainrepo.PageResult[*entity.PartyRole], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := filter.Normalize()
	q := strings.ToLower(f.Query)
	matched := make([]*entity.PartyRole, 0)
	for _, id := range r.order {
		role := r.items[id]
		if role == nil {
			continue
		}
		if f.PartyID != "" && role.PartyID != f.PartyID {
			continue
		}
		if f.Role != nil && role.Role != *f.Role {
			continue
		}
		if f.ActiveOnly && !role.IsActive {
			continue
		}
		if q != "" && (role.ReferenceNote == nil || !strings.Contains(strings.ToLower(*role.ReferenceNote), q)) {
			continue
		}
		matched = append(matched, role.Clone())
	}
	return paginateSlice(matched, f.Page), nil
}

func (r *memPartyRoleRepo) CheckExists(_ context.Context, partyID string, kind entity.PartyRoleKind) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, role := range r.items {
		if role.PartyID == partyID && role.Role == kind {
			return true, nil
		}
	}
	return false, nil
}

func (r *memPartyRoleRepo) Insert(_ context.Context, role *entity.PartyRole) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.order = append(r.order, role.ID)
	r.items[role.ID] = role.Clone()
	return nil
}

func (r *memPartyRoleRepo) Update(_ context.Context, role *entity.PartyRole) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[role.ID]; !ok {
		return eferrors.ErrPartyRoleNotFound
	}
	r.items[role.ID] = role.Clone()
	return nil
}

func (r *memPartyRoleRepo) DeleteByID(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return eferrors.ErrPartyRoleNotFound
	}
	delete(r.items, id)
	return nil
}

type memPartyContactRepo struct {
	mu    sync.RWMutex
	order []string
	items map[string]*entity.PartyContact
}

func newMemPartyContactRepo() *memPartyContactRepo {
	return &memPartyContactRepo{items: make(map[string]*entity.PartyContact)}
}

func (r *memPartyContactRepo) SelectByID(_ context.Context, id string) (*entity.PartyContact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.items[id]; ok && !c.IsDeleted() {
		return c.Clone(), nil
	}
	return nil, eferrors.ErrPartyContactNotFound
}

func (r *memPartyContactRepo) SelectByPartyID(_ context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.PartyContact, 0)
	for _, id := range r.order {
		if c := r.items[id]; c != nil && !c.IsDeleted() && c.PartyID == partyID {
			matched = append(matched, c.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memPartyContactRepo) SelectPrimaryByPartyID(_ context.Context, partyID string) (*entity.PartyContact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if c := r.items[id]; c != nil && !c.IsDeleted() && c.PartyID == partyID && c.IsPrimary {
			return c.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyContactNotFound
}

func (r *memPartyContactRepo) SelectByIAMUserID(_ context.Context, iamUserID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.PartyContact, 0)
	for _, id := range r.order {
		if c := r.items[id]; c != nil && !c.IsDeleted() && c.IAMUserID != nil && *c.IAMUserID == iamUserID {
			matched = append(matched, c.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memPartyContactRepo) SelectByPartyAndIAMUserID(_ context.Context, partyID, iamUserID string) (*entity.PartyContact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if c := r.items[id]; c != nil && !c.IsDeleted() && c.PartyID == partyID && c.IAMUserID != nil && *c.IAMUserID == iamUserID {
			return c.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyContactNotFound
}

func (r *memPartyContactRepo) SelectByEmail(_ context.Context, email string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyContact], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.PartyContact, 0)
	for _, id := range r.order {
		if c := r.items[id]; c != nil && !c.IsDeleted() && c.Email != nil && strings.EqualFold(*c.Email, email) {
			matched = append(matched, c.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memPartyContactRepo) SelectList(_ context.Context, filter domainrepo.PartyContactFilter) (domainrepo.PageResult[*entity.PartyContact], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := filter.Normalize()
	q := strings.ToLower(f.Query)
	matched := make([]*entity.PartyContact, 0)
	for _, id := range r.order {
		c := r.items[id]
		if c == nil || (!f.IncludeDeleted && c.IsDeleted()) {
			continue
		}
		if f.PartyID != "" && c.PartyID != f.PartyID {
			continue
		}
		if f.ContactRole != "" && !strings.EqualFold(c.ContactRole, f.ContactRole) {
			continue
		}
		if f.PortalAccessStatus != nil && c.PortalAccessStatus != *f.PortalAccessStatus {
			continue
		}
		if f.ActiveOnly && !c.IsActive {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(c.FullName+" "+c.ContactRole), q) {
			continue
		}
		matched = append(matched, c.Clone())
	}
	return paginateSlice(matched, f.Page), nil
}

func (r *memPartyContactRepo) CheckIAMUserExistsInParty(_ context.Context, partyID, iamUserID, excludeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.items {
		if !c.IsDeleted() && c.ID != excludeID && c.PartyID == partyID && c.IAMUserID != nil && *c.IAMUserID == iamUserID {
			return true, nil
		}
	}
	return false, nil
}

func (r *memPartyContactRepo) UpdateClearPrimaryByPartyExcept(_ context.Context, partyID, keepID string, updatedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.items {
		if !c.IsDeleted() && c.PartyID == partyID && c.ID != keepID && c.IsPrimary {
			c.IsPrimary = false
			c.UpdatedAt = updatedAt
		}
	}
	return nil
}

func (r *memPartyContactRepo) Insert(_ context.Context, c *entity.PartyContact) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.order = append(r.order, c.ID)
	r.items[c.ID] = c.Clone()
	return nil
}

func (r *memPartyContactRepo) Update(_ context.Context, c *entity.PartyContact) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.items[c.ID]
	if !ok || existing.IsDeleted() {
		return eferrors.ErrPartyContactNotFound
	}
	next := c.Clone()
	next.DeletedAt = existing.DeletedAt
	r.items[c.ID] = next
	return nil
}

func (r *memPartyContactRepo) SoftDeleteByID(_ context.Context, id string, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.items[id]
	if !ok || c.IsDeleted() {
		return eferrors.ErrPartyContactNotFound
	}
	c.DeletedAt = &deletedAt
	return nil
}

type memPartyAddressRepo struct {
	mu    sync.RWMutex
	order []string
	items map[string]*entity.PartyAddress
}

func newMemPartyAddressRepo() *memPartyAddressRepo {
	return &memPartyAddressRepo{items: make(map[string]*entity.PartyAddress)}
}

func (r *memPartyAddressRepo) referencesAddress(addrID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, link := range r.items {
		if link.AddressID == addrID {
			return true
		}
	}
	return false
}

func (r *memPartyAddressRepo) SelectByID(_ context.Context, id string) (*entity.PartyAddress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if link, ok := r.items[id]; ok {
		return link.Clone(), nil
	}
	return nil, eferrors.ErrPartyAddressNotFound
}

func (r *memPartyAddressRepo) SelectByPartyID(_ context.Context, partyID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.PartyAddress, 0)
	for _, id := range r.order {
		if link := r.items[id]; link != nil && link.PartyID == partyID {
			matched = append(matched, link.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memPartyAddressRepo) SelectByPartyAndUsage(_ context.Context, partyID string, usage entity.AddressUsageKind, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.PartyAddress, 0)
	for _, id := range r.order {
		if link := r.items[id]; link != nil && link.PartyID == partyID && link.Usage == usage {
			matched = append(matched, link.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memPartyAddressRepo) SelectPrimaryByPartyAndUsage(_ context.Context, partyID string, usage entity.AddressUsageKind) (*entity.PartyAddress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if link := r.items[id]; link != nil && link.PartyID == partyID && link.Usage == usage && link.IsPrimary {
			return link.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyAddressNotFound
}

func (r *memPartyAddressRepo) SelectByAddressID(_ context.Context, addrID string, page domainrepo.PageRequest) (domainrepo.PageResult[*entity.PartyAddress], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matched := make([]*entity.PartyAddress, 0)
	for _, id := range r.order {
		if link := r.items[id]; link != nil && link.AddressID == addrID {
			matched = append(matched, link.Clone())
		}
	}
	return paginateSlice(matched, page), nil
}

func (r *memPartyAddressRepo) SelectByPartyAddressUsage(_ context.Context, partyID, addrID string, usage entity.AddressUsageKind) (*entity.PartyAddress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, id := range r.order {
		if link := r.items[id]; link != nil && link.PartyID == partyID && link.AddressID == addrID && link.Usage == usage {
			return link.Clone(), nil
		}
	}
	return nil, eferrors.ErrPartyAddressNotFound
}

func (r *memPartyAddressRepo) SelectList(_ context.Context, filter domainrepo.PartyAddressFilter) (domainrepo.PageResult[*entity.PartyAddress], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := filter.Normalize()
	matched := make([]*entity.PartyAddress, 0)
	for _, id := range r.order {
		link := r.items[id]
		if link == nil {
			continue
		}
		if f.PartyID != "" && link.PartyID != f.PartyID {
			continue
		}
		if f.AddressID != "" && link.AddressID != f.AddressID {
			continue
		}
		if f.Usage != nil && link.Usage != *f.Usage {
			continue
		}
		if f.PrimaryOnly != nil && link.IsPrimary != *f.PrimaryOnly {
			continue
		}
		matched = append(matched, link.Clone())
	}
	return paginateSlice(matched, f.Page), nil
}

func (r *memPartyAddressRepo) UpdateClearPrimaryByPartyUsageExcept(_ context.Context, partyID string, usage entity.AddressUsageKind, keepID string, updatedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, link := range r.items {
		if link.PartyID == partyID && link.Usage == usage && link.ID != keepID && link.IsPrimary {
			link.IsPrimary = false
			link.UpdatedAt = updatedAt
		}
	}
	return nil
}

func (r *memPartyAddressRepo) Insert(_ context.Context, link *entity.PartyAddress) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.order = append(r.order, link.ID)
	r.items[link.ID] = link.Clone()
	return nil
}

func (r *memPartyAddressRepo) Update(_ context.Context, link *entity.PartyAddress) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[link.ID]; !ok {
		return eferrors.ErrPartyAddressNotFound
	}
	r.items[link.ID] = link.Clone()
	return nil
}

func (r *memPartyAddressRepo) DeleteByID(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return eferrors.ErrPartyAddressNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *memOrgUnitRepo) SelectByIDForUpdate(ctx context.Context, id string) (*entity.OrganizationUnit, error) {
	return r.SelectByID(ctx, id)
}
