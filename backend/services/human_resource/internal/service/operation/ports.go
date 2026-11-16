// Package service menyediakan mekanisme application service bersama tanpa dependency transport/GORM.
package service

import (
	"context"
	"time"

	"github.com/integrione/erp-retail/backend/services/human_resource/domain/entity"
	domainrepo "github.com/integrione/erp-retail/backend/services/human_resource/domain/repository"
)

// Actor hanya boleh dikembalikan adapter keamanan yang sudah memverifikasi principal/sesi.
type Actor struct {
	PrincipalID, SessionID, Kind  string
	AuthEpoch, PermissionRevision int64
}

// Access membawa atribut minimum objek, tanpa nama, gaji, narasi kasus atau dokumen mentah.
type Access struct {
	Domain, Resource, Action string
	Attributes               map[string]string
}

// Scope adalah keputusan server, bukan filter yang diterima langsung dari browser.
type Scope struct {
	All     bool
	Filters []domainrepo.Filter
}

// Policy wajib memverifikasi permission/scope/SoD dan kesegaran keamanan pada setiap panggilan.
type Policy interface {
	Authorize(context.Context, Access) (Actor, error)
	ReadScope(context.Context, Access) (Scope, error)
}

// Integrity memverifikasi referensi, versi dan invariant lintas objek lewat kontrak pemilik yang sah.
// Snapshot hanya berada pada boundary internal ini; implementasi dilarang mengirim PII ke IAM.
type Integrity interface {
	ValidateMutation(context.Context, string, entity.Validatable, entity.Validatable) error
}

// Clock menyediakan waktu yang dapat dikendalikan dalam tes.
type Clock interface {
	Now() time.Time
}

// IDGenerator menyediakan pengenal yang dapat dikendalikan dalam tes.
type IDGenerator interface {
	NewID() (string, error)
}

// Repository mempertahankan kompatibilitas port operasi dengan kontrak milik domain.
type Repository[T any, K any] = domainrepo.RecordRepository[T, K]

// Outbox menyimpan event dalam transaksi perubahan yang sama, tanpa publikasi jaringan.
type Outbox interface {
	Insert(context.Context, *entity.OutboxEvent) error
	FindByAggregateID(context.Context, string, int, int) ([]entity.OutboxEvent, error)
}

// Dependencies wajib diisi; tidak ada policy allow-all atau referensi sukses palsu bawaan.
type Dependencies struct {
	Transactor domainrepo.Transactor
	Policy     Policy
	Integrity  Integrity
	Outbox     Outbox
	Clock      Clock
	IDs        IDGenerator
}

// Pointer mengikat entity dan method Validate pada model yang sama dengan repository.
type Pointer[T any] interface {
	*T
	entity.Validatable
}

// Specification berisi mapper dan batas lifecycle khusus satu tabel, dibentuk service domain.
type Specification[T any, K any] struct {
	Domain, Table string
	AggregateID   func(*T) string
	Key           func(*T) K
	KeyString     func(K) string
	ValidateKey   func(K) error
	Initialize    func(*T, Actor, time.Time, string)
	Touch         func(*T, Actor, time.Time)
	Attributes    func(*T) map[string]string
	Guard         func(string, *T, *T) error
}

// Page tidak menjanjikan snapshot count/list atomik pada isolasi READ COMMITTED.
type Page[T any] struct {
	Items         []T
	Total         int64
	Limit, Offset int
}

// Field membedakan patch tidak dikirim, null, dan nilai nol/false.
type Field[T any] struct {
	Set   bool
	Value *T
}
