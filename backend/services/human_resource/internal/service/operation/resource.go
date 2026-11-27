package service

import (
	"context"
	"fmt"
	"reflect"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Resource mengoordinasikan akses, transaksi, validasi dan event satu tabel.
type Resource[T any, K any, P Pointer[T]] struct {
	repo Repository[T, K]
	deps Dependencies
	spec Specification[T, K]
}

// NewResource memeriksa wiring secara fail-closed tanpa fallback policy/referensi.
func NewResource[T any, K any, P Pointer[T]](
	repo Repository[T, K],
	deps Dependencies,
	spec Specification[T, K],
) (*Resource[T, K, P], error) {
	for _, dependency := range []struct {
		name  string
		value any
	}{
		{"repository", repo}, {"transactor", deps.Transactor}, {"policy", deps.Policy},
		{"integrity", deps.Integrity}, {"outbox", deps.Outbox},
	} {
		if isNil(dependency.value) {
			return nil, fmt.Errorf("%w: %s %s dependency is required", hrerrors.ErrUnavailable, spec.Table, dependency.name)
		}
	}
	if spec.Table == "" || spec.Domain == "" {
		return nil, hrerrors.ErrInvalidInput
	}
	for _, callback := range []any{
		spec.Key, spec.KeyString, spec.ValidateKey, spec.Initialize,
		spec.Touch, spec.Attributes, spec.Guard, spec.AggregateID,
	} {
		if isNil(callback) {
			return nil, hrerrors.ErrInvalidInput
		}
	}
	if isNil(deps.Clock) {
		deps.Clock = SystemClock{}
	}
	if isNil(deps.IDs) {
		deps.IDs = UUIDGenerator{}
	}
	return &Resource[T, K, P]{repo: repo, deps: deps, spec: spec}, nil
}

func clone[T any](value *T) (*T, error) {
	if value == nil {
		return nil, hrerrors.ErrNotFound
	}
	copy := new(T)
	reflect.ValueOf(copy).Elem().Set(cloneField(reflect.ValueOf(value).Elem()))
	return copy, nil
}

// cloneField menyalin model data tanpa mengubah SQL NULL menjadi literal JSON null.
// Field internal time.Time disalin sebagai nilai; field data publik disalin rekursif.
func cloneField(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.New(value.Type().Elem())
		copy.Elem().Set(cloneField(value.Elem()))
		return copy
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := 0; index < value.Len(); index++ {
			copy.Index(index).Set(cloneField(value.Index(index)))
		}
		return copy
	case reflect.Array:
		copy := reflect.New(value.Type()).Elem()
		for index := 0; index < value.Len(); index++ {
			copy.Index(index).Set(cloneField(value.Index(index)))
		}
		return copy
	case reflect.Struct:
		copy := reflect.New(value.Type()).Elem()
		copy.Set(value)
		for index := 0; index < value.NumField(); index++ {
			if value.Field(index).CanInterface() && copy.Field(index).CanSet() {
				copy.Field(index).Set(cloneField(value.Field(index)))
			}
		}
		return copy
	default:
		return value
	}
}
func checkContext(ctx context.Context) error {
	if ctx == nil {
		return hrerrors.ErrInvalidInput
	}
	return ctx.Err()
}
func (r *Resource[T, K, P]) authorize(ctx context.Context, action string, value *T) (Actor, error) {
	if r == nil {
		return Actor{}, hrerrors.ErrUnavailable
	}
	if err := checkContext(ctx); err != nil {
		return Actor{}, err
	}
	attrs := map[string]string{}
	if value != nil {
		attrs = r.spec.Attributes(value)
	}
	actor, err := r.deps.Policy.Authorize(ctx, Access{Domain: r.spec.Domain, Resource: r.spec.Table, Action: action, Attributes: attrs})
	if err != nil {
		return Actor{}, err
	}
	if err := entity.ValidateUUID("principal_id", actor.PrincipalID); err != nil {
		return Actor{}, hrerrors.ErrForbidden
	}
	if actor.Kind != "human" && actor.Kind != "workload" {
		return Actor{}, hrerrors.ErrForbidden
	}
	if actor.AuthEpoch < 0 || actor.PermissionRevision < 0 {
		return Actor{}, hrerrors.ErrForbidden
	}
	if actor.Kind == "human" {
		if err := entity.ValidateUUID("session_id", actor.SessionID); err != nil {
			return Actor{}, hrerrors.ErrForbidden
		}
	}
	if value != nil && r.spec.CheckActor != nil {
		if err := r.spec.CheckActor(action, value, actor); err != nil {
			return Actor{}, err
		}
	}
	return actor, nil
}
func sameActor(first, second Actor) error {
	if first != second {
		return hrerrors.ErrForbidden
	}
	return nil
}

// Get memeriksa akses sebelum query dan scope objek sebelum mengembalikan data.
func (r *Resource[T, K, P]) Get(ctx context.Context, key K) (*T, error) {
	if r == nil {
		return nil, hrerrors.ErrUnavailable
	}
	if err := r.spec.ValidateKey(key); err != nil {
		return nil, err
	}
	actor, err := r.authorize(ctx, "read", nil)
	if err != nil {
		return nil, err
	}
	scope, err := r.deps.Policy.ReadScope(ctx, Access{Domain: r.spec.Domain, Resource: r.spec.Table, Action: "read"})
	if err != nil {
		return nil, err
	}
	var value *T
	if scope.All && len(scope.Filters) == 0 {
		value, err = r.repo.SelectByID(ctx, key)
		if err != nil {
			return nil, err
		}
	} else {
		if scope.All || len(scope.Filters) == 0 || r.spec.KeyFilters == nil {
			return nil, hrerrors.ErrForbidden
		}
		filters := r.spec.KeyFilters(key)
		if len(filters) == 0 {
			return nil, hrerrors.ErrForbidden
		}
		filters = append(append([]domainrepo.Filter{}, scope.Filters...), filters...)
		matches, err := r.repo.SelectWhere(ctx, filters, 2, 0)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			return nil, hrerrors.ErrNotFound
		}
		if len(matches) != 1 {
			return nil, hrerrors.ErrConflict
		}
		value = &matches[0]
	}
	checked, err := r.authorize(ctx, "read", value)
	if err != nil {
		return nil, err
	}
	if err := sameActor(actor, checked); err != nil {
		return nil, err
	}
	return clone(value)
}

// List menerapkan scope server dalam SQL sebelum pagination dan count.
func (r *Resource[T, K, P]) List(ctx context.Context, filters []domainrepo.Filter, limit, offset int) (Page[T], error) {
	empty := Page[T]{}
	actor, err := r.authorize(ctx, "list", nil)
	if err != nil {
		return empty, err
	}
	if limit < 0 || offset < 0 {
		return empty, hrerrors.ErrInvalidInput
	}
	if limit == 0 {
		limit = domainrepo.DefaultLimit
	}
	if limit > domainrepo.MaxLimit {
		limit = domainrepo.MaxLimit
	}
	listAccess := Access{Domain: r.spec.Domain, Resource: r.spec.Table, Action: "list"}
	scope, err := r.deps.Policy.ReadScope(ctx, listAccess)
	if err != nil {
		return empty, err
	}
	if (!scope.All && len(scope.Filters) == 0) || (scope.All && len(scope.Filters) > 0) {
		return empty, hrerrors.ErrForbidden
	}
	combined := append(append([]domainrepo.Filter{}, scope.Filters...), filters...)
	for _, filter := range combined {
		if filter.Column == "" || len(filter.Values) == 0 {
			return empty, hrerrors.ErrForbidden
		}
	}
	items, err := r.repo.SelectWhere(ctx, combined, limit, offset)
	if err != nil {
		return empty, err
	}
	if items == nil {
		items = []T{}
	}
	for i := range items {
		checked, err := r.authorize(ctx, "read", &items[i])
		if err != nil {
			return empty, err
		}
		if err := sameActor(actor, checked); err != nil {
			return empty, err
		}
	}
	checked, err := r.authorize(ctx, "list", nil)
	if err != nil {
		return empty, err
	}
	if err := sameActor(actor, checked); err != nil {
		return empty, err
	}
	total, err := r.repo.CountWhere(ctx, combined)
	if err != nil {
		return empty, err
	}
	// Halaman kosong juga membawa total yang perlu dilindungi oleh scope semula.
	currentScope, err := r.deps.Policy.ReadScope(ctx, listAccess)
	if err != nil {
		return empty, err
	}
	if !reflect.DeepEqual(scope, currentScope) {
		return empty, hrerrors.ErrForbidden
	}
	if err := r.finish(ctx, "list", nil, actor); err != nil {
		return empty, err
	}
	return Page[T]{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

// Create membangun entity dari mapper DTO setelah actor terverifikasi.
func (r *Resource[T, K, P]) Create(ctx context.Context, build func() (*T, error)) (*T, error) {
	actor, err := r.authorize(ctx, "create", nil)
	if err != nil {
		return nil, err
	}
	if build == nil {
		return nil, hrerrors.ErrInvalidInput
	}
	value, err := build()
	if err != nil {
		return nil, err
	}
	value, err = clone(value)
	if err != nil {
		return nil, err
	}
	id, err := r.deps.IDs.NewID()
	if err != nil {
		return nil, err
	}
	if err := entity.ValidateUUID("id", id); err != nil {
		return nil, err
	}
	now := r.deps.Clock.Now()
	if now.IsZero() {
		return nil, hrerrors.ErrInvalidInput
	}
	r.spec.Initialize(value, actor, now, id)
	if err := P(value).Validate(); err != nil {
		return nil, err
	}
	if err := r.spec.Guard("create", nil, value); err != nil {
		return nil, err
	}
	err = r.deps.Transactor.WithinTransaction(ctx, func(tc context.Context) error {
		checked, err := r.authorize(tc, "create", value)
		if err != nil {
			return err
		}
		if err := sameActor(actor, checked); err != nil {
			return err
		}
		if err := r.CheckIntegrity(tc, nil, value); err != nil {
			return err
		}
		if err := r.repo.Insert(tc, value); err != nil {
			return err
		}
		if err := r.spec.ValidateKey(r.spec.Key(value)); err != nil {
			return err
		}
		if err := r.emit(tc, "created", value, actor); err != nil {
			return err
		}
		return r.finish(tc, "create", value, actor)
	})
	if err != nil {
		return nil, err
	}
	return clone(value)
}

// Update mengambil lock dan menerapkan patch pada snapshot lengkap yang disalin.
func (r *Resource[T, K, P]) Update(ctx context.Context, key K, patch func(*T) error) (*T, error) {
	actor, err := r.authorize(ctx, "update", nil)
	if err != nil {
		return nil, err
	}
	if err := r.spec.ValidateKey(key); err != nil {
		return nil, err
	}
	if patch == nil {
		return nil, hrerrors.ErrInvalidInput
	}
	var result *T
	err = r.deps.Transactor.WithinTransaction(ctx, func(tc context.Context) error {
		old, err := r.repo.FindByIDForUpdate(tc, key)
		if err != nil {
			return err
		}
		old, err = clone(old)
		if err != nil {
			return err
		}
		checked, err := r.authorize(tc, "update", old)
		if err != nil {
			return err
		}
		if err := sameActor(actor, checked); err != nil {
			return err
		}
		next, err := clone(old)
		if err != nil {
			return err
		}
		if err := patch(next); err != nil {
			return err
		}
		r.spec.Touch(next, actor, r.deps.Clock.Now())
		if !reflect.DeepEqual(r.spec.Key(old), r.spec.Key(next)) {
			return hrerrors.ErrInvalidInput
		}
		if err := P(next).Validate(); err != nil {
			return err
		}
		if err := r.spec.Guard("update", old, next); err != nil {
			return err
		}
		checked, err = r.authorize(tc, "update", next)
		if err != nil {
			return err
		}
		if err := sameActor(actor, checked); err != nil {
			return err
		}
		if err := r.CheckIntegrity(tc, old, next); err != nil {
			return err
		}
		if err := r.repo.Update(tc, next); err != nil {
			return err
		}
		if err := r.emit(tc, "updated", next, actor); err != nil {
			return err
		}
		result = next
		return r.finish(tc, "update", next, actor)
	})
	if err != nil {
		return nil, err
	}
	return clone(result)
}

// Delete tidak memintas pembekuan riwayat dan foreign key milik database.
func (r *Resource[T, K, P]) Delete(ctx context.Context, key K) error {
	actor, err := r.authorize(ctx, "delete", nil)
	if err != nil {
		return err
	}
	if err := r.spec.ValidateKey(key); err != nil {
		return err
	}
	return r.deps.Transactor.WithinTransaction(ctx, func(tc context.Context) error {
		value, err := r.repo.FindByIDForUpdate(tc, key)
		if err != nil {
			return err
		}
		if value == nil {
			return hrerrors.ErrNotFound
		}
		checked, err := r.authorize(tc, "delete", value)
		if err != nil {
			return err
		}
		if err := sameActor(actor, checked); err != nil {
			return err
		}
		if err := r.spec.Guard("delete", value, nil); err != nil {
			return err
		}
		if err := r.CheckIntegrity(tc, value, nil); err != nil {
			return err
		}
		if err := r.repo.DeleteByID(tc, key); err != nil {
			return err
		}
		if err := r.emit(tc, "deleted", value, actor); err != nil {
			return err
		}
		return r.finish(tc, "delete", value, actor)
	})
}

// finish memeriksa ulang actor/revisi keamanan tepat sebelum callback transaksi selesai.
func (r *Resource[T, K, P]) finish(ctx context.Context, action string, value *T, actor Actor) error {
	checked, err := r.authorize(ctx, action, value)
	if err != nil {
		return err
	}
	if err := sameActor(actor, checked); err != nil {
		return err
	}
	return checkContext(ctx)
}

// CheckIntegrity memberi validator salinan agar pemeriksaan dependency tidak mengubah write intent.
func (r *Resource[T, K, P]) CheckIntegrity(ctx context.Context, before, after *T) error {
	if r == nil {
		return hrerrors.ErrUnavailable
	}
	var previous, proposed entity.Validatable
	if before != nil {
		copy, err := clone(before)
		if err != nil {
			return err
		}
		previous = P(copy)
	}
	if after != nil {
		copy, err := clone(after)
		if err != nil {
			return err
		}
		proposed = P(copy)
	}
	return r.deps.Integrity.ValidateMutation(ctx, r.spec.Table, previous, proposed)
}
