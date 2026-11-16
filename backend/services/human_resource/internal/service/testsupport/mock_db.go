// Package testsupport menyediakan mock_db khusus tes service; tidak digunakan runtime HRM.
package testsupport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
	operation "github.com/integrione/erp-retail/backend/services/human_resource/internal/service/operation"
)

type txKey struct{}
type snapshotter interface {
	snapshot() any
	restore(any)
}

// Harness adalah dependency mock eksplisit; policy allow hanya berlaku pada fixture tes ini.
type Harness struct {
	mu                                           sync.Mutex
	idMu                                         sync.Mutex
	sequence                                     int
	stores                                       []snapshotter
	Events                                       []entity.OutboxEvent
	Failures                                     map[string]error
	Scope                                        operation.Scope
	Actor                                        operation.Actor
	DeniedObject                                 string
	LastFilters                                  []domainrepo.Filter
	WriteCalls, IntegrityCalls, TransactionCalls int
	RevokeAfterNextOutbox                        bool
	revoked                                      bool
	RetainLastEvent                              bool
}

func New() *Harness {
	return &Harness{Failures: map[string]error{}, Scope: operation.Scope{All: true}, Actor: operation.Actor{PrincipalID: ID(900), SessionID: ID(901), Kind: "human"}}
}
func ID(n int) string             { return fmt.Sprintf("00000000-0000-0000-0000-%012d", n) }
func Now() time.Time              { return time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC) }
func (h *Harness) Now() time.Time { return Now() }
func (h *Harness) NewID() (string, error) {
	h.idMu.Lock()
	defer h.idMu.Unlock()
	h.sequence++
	return ID(10000 + h.sequence), nil
}
func (h *Harness) lock(ctx context.Context) func() {
	if ctx != nil && ctx.Value(txKey{}) == h {
		return func() {}
	}
	h.mu.Lock()
	return h.mu.Unlock
}
func copyValue[T any](source T) T {
	data, err := json.Marshal(source)
	if err != nil {
		panic(err)
	}
	var target T
	if err := json.Unmarshal(data, &target); err != nil {
		panic(err)
	}
	// MarshalJSON nil milik entity.JSON berbentuk literal null. Fixture memulihkan
	// nil slice sumber agar mock tidak menyembunyikan perbedaan SQL NULL dan JSON null.
	sourceFields := reflect.ValueOf(source)
	targetFields := reflect.ValueOf(&target).Elem()
	if sourceFields.Kind() == reflect.Struct {
		for index := 0; index < sourceFields.NumField(); index++ {
			field := sourceFields.Field(index)
			if field.Kind() == reflect.Slice && field.IsNil() && targetFields.Field(index).CanSet() {
				targetFields.Field(index).SetZero()
			}
		}
	}
	return target
}
func (h *Harness) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	unlock := h.lock(ctx)
	defer unlock()
	h.TransactionCalls++
	if err := h.Failures["begin"]; err != nil {
		return err
	}
	states := make([]any, len(h.stores))
	for i, store := range h.stores {
		states[i] = store.snapshot()
	}
	events := copyValue(h.Events)
	err := fn(context.WithValue(ctx, txKey{}, h))
	if err == nil {
		err = h.Failures["commit"]
	}
	if err != nil {
		for i, store := range h.stores {
			store.restore(states[i])
		}
		h.Events = events
	}
	return err
}
func (h *Harness) Authorize(ctx context.Context, access operation.Access) (operation.Actor, error) {
	unlock := h.lock(ctx)
	defer unlock()
	if h.revoked {
		return operation.Actor{}, hrerrors.ErrForbidden
	}
	if err := h.Failures["authorize"]; err != nil {
		return operation.Actor{}, err
	}
	for _, value := range access.Attributes {
		if h.DeniedObject != "" && value == h.DeniedObject {
			return operation.Actor{}, hrerrors.ErrForbidden
		}
	}
	return h.Actor, nil
}
func (h *Harness) ReadScope(ctx context.Context, _ operation.Access) (operation.Scope, error) {
	unlock := h.lock(ctx)
	defer unlock()
	if err := h.Failures["scope"]; err != nil {
		return operation.Scope{}, err
	}
	return h.Scope, nil
}
func (h *Harness) ValidateMutation(ctx context.Context, _ string, _, _ entity.Validatable) error {
	unlock := h.lock(ctx)
	defer unlock()
	h.IntegrityCalls++
	return h.Failures["integrity"]
}
func (h *Harness) Insert(ctx context.Context, event *entity.OutboxEvent) error {
	unlock := h.lock(ctx)
	defer unlock()
	if ctx.Value(txKey{}) != h {
		return errors.New("outbox outside transaction")
	}
	if err := h.Failures["outbox"]; err != nil {
		return err
	}
	if h.RetainLastEvent {
		h.Events = nil
	}
	h.Events = append(h.Events, copyValue(*event))
	if h.RevokeAfterNextOutbox {
		h.revoked = true
	}
	return nil
}
func (h *Harness) FindByAggregateID(ctx context.Context, id string, limit, offset int) ([]entity.OutboxEvent, error) {
	unlock := h.lock(ctx)
	defer unlock()
	if err := h.Failures["outbox_read"]; err != nil {
		return nil, err
	}
	rows := []entity.OutboxEvent{}
	for _, event := range h.Events {
		if event.AggregateID == id {
			rows = append(rows, copyValue(event))
		}
	}
	if offset >= len(rows) {
		return []entity.OutboxEvent{}, nil
	}
	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[offset:end], nil
}
func (h *Harness) Runtime() operation.Dependencies {
	return operation.Dependencies{Transactor: h, Policy: h, Integrity: h, Outbox: h, Clock: h, IDs: h}
}

// Store menyimpan salinan model; metode mutasi menolak context yang bukan transaksi.
type Store[T any, K comparable] struct {
	Harness *Harness
	Key     func(*T) K
	Rows    map[K]T
}

func NewStore[T any, K comparable](h *Harness, key func(*T) K) *Store[T, K] {
	store := &Store[T, K]{Harness: h, Key: key, Rows: map[K]T{}}
	h.stores = append(h.stores, store)
	return store
}
func (s *Store[T, K]) snapshot() any {
	copy := map[K]T{}
	for key, row := range s.Rows {
		copy[key] = copyValue(row)
	}
	return copy
}
func (s *Store[T, K]) restore(value any) { s.Rows = value.(map[K]T) }
func (s *Store[T, K]) SelectByID(ctx context.Context, key K) (*T, error) {
	unlock := s.Harness.lock(ctx)
	defer unlock()
	if err := s.Harness.Failures["read"]; err != nil {
		return nil, err
	}
	value, exists := s.Rows[key]
	if !exists {
		return nil, hrerrors.ErrNotFound
	}
	value = copyValue(value)
	return &value, nil
}
func (s *Store[T, K]) FindByIDForUpdate(ctx context.Context, key K) (*T, error) {
	if ctx.Value(txKey{}) != s.Harness {
		return nil, errors.New("row lock outside transaction")
	}
	return s.SelectByID(ctx, key)
}
func field(value any, column string) (any, bool) {
	v := reflect.ValueOf(value)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("gorm")
		if !strings.HasPrefix(tag, "column:"+column+";") && tag != "column:"+column {
			continue
		}
		f := v.Field(i)
		if f.Kind() == reflect.Pointer {
			if f.IsNil() {
				return nil, true
			}
			f = f.Elem()
		}
		if f.Kind() == reflect.String {
			return f.String(), true
		}
		return f.Interface(), true
	}
	return nil, false
}
func equal(a, b any) bool {
	if b != nil {
		v := reflect.ValueOf(b)
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				b = nil
			} else {
				b = v.Elem().Interface()
			}
		}
	}
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}
func (s *Store[T, K]) selected(filters []domainrepo.Filter) ([]T, error) {
	rows := []T{}
	for _, value := range s.Rows {
		match := true
		for _, filter := range filters {
			current, exists := field(value, filter.Column)
			if !exists || len(filter.Values) == 0 {
				return nil, hrerrors.ErrInvalidInput
			}
			allowed := false
			for _, candidate := range filter.Values {
				allowed = allowed || equal(current, candidate)
			}
			match = match && allowed
		}
		if match {
			rows = append(rows, copyValue(value))
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		a, b := s.Key(&rows[i]), s.Key(&rows[j])
		av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
		if av.Kind() == reflect.Int64 {
			return av.Int() < bv.Int()
		}
		return fmt.Sprint(a) < fmt.Sprint(b)
	})
	return rows, nil
}
func (s *Store[T, K]) SelectWhere(ctx context.Context, filters []domainrepo.Filter, limit, offset int) ([]T, error) {
	unlock := s.Harness.lock(ctx)
	defer unlock()
	if err := s.Harness.Failures["list"]; err != nil {
		return nil, err
	}
	s.Harness.LastFilters = filters
	rows, err := s.selected(filters)
	if err != nil {
		return nil, err
	}
	if offset >= len(rows) {
		return []T{}, nil
	}
	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[offset:end], nil
}
func (s *Store[T, K]) CountWhere(ctx context.Context, filters []domainrepo.Filter) (int64, error) {
	unlock := s.Harness.lock(ctx)
	defer unlock()
	if err := s.Harness.Failures["count"]; err != nil {
		return 0, err
	}
	rows, err := s.selected(filters)
	return int64(len(rows)), err
}
func (s *Store[T, K]) Insert(ctx context.Context, value *T) error {
	unlock := s.Harness.lock(ctx)
	defer unlock()
	if ctx.Value(txKey{}) != s.Harness {
		return errors.New("insert outside transaction")
	}
	if err := s.Harness.Failures["insert"]; err != nil {
		return err
	}
	key := s.Key(value)
	if _, exists := s.Rows[key]; exists {
		return hrerrors.ErrAlreadyExists
	}
	s.Rows[key] = copyValue(*value)
	s.Harness.WriteCalls++
	return nil
}
func (s *Store[T, K]) Update(ctx context.Context, value *T) error {
	unlock := s.Harness.lock(ctx)
	defer unlock()
	if ctx.Value(txKey{}) != s.Harness {
		return errors.New("update outside transaction")
	}
	if err := s.Harness.Failures["update"]; err != nil {
		return err
	}
	key := s.Key(value)
	if _, exists := s.Rows[key]; !exists {
		return hrerrors.ErrNotFound
	}
	s.Rows[key] = copyValue(*value)
	s.Harness.WriteCalls++
	return nil
}
func (s *Store[T, K]) DeleteByID(ctx context.Context, key K) error {
	unlock := s.Harness.lock(ctx)
	defer unlock()
	if ctx.Value(txKey{}) != s.Harness {
		return errors.New("delete outside transaction")
	}
	if err := s.Harness.Failures["delete"]; err != nil {
		return err
	}
	if _, exists := s.Rows[key]; !exists {
		return hrerrors.ErrNotFound
	}
	delete(s.Rows, key)
	s.Harness.WriteCalls++
	return nil
}
