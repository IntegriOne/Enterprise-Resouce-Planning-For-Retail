// Package iamerrors menyediakan tipe error domain terstruktur dan sentinel error untuk layanan IAM.
package iamerrors

import (
	"errors"
	"strings"
)

// Kind mengelompokkan kategori semantik error domain agar mudah dipetakan ke lapisan transport tanpa bergantung pada HTTP/gRPC.
type Kind uint8

const (
	// KindUnknown menandakan error yang belum diklasifikasikan ke kategori domain tertentu.
	KindUnknown Kind = iota
	// KindInvalidInput menandakan pelanggaran validasi masukan atau invarian entitas.
	KindInvalidInput
	// KindNotFound menandakan entitas atau relasi yang diminta tidak ditemukan.
	KindNotFound
	// KindAlreadyExists menandakan duplikasi kunci unik atau relasi yang sudah terdaftar.
	KindAlreadyExists
	// KindUnauthorized menandakan autentikasi gagal, sesi berakhir, atau token tidak sah.
	KindUnauthorized
	// KindForbidden menandakan principal terautentikasi tetapi tidak memiliki izin atau ruang lingkup yang sesuai.
	KindForbidden
	// KindConflict menandakan transisi state atau modifikasi yang bertentangan dengan kondisi terkini.
	KindConflict
	// KindRateLimited menandakan akses dibatasi sementara misalnya akibat penguncian akun.
	KindRateLimited
	// KindUnavailable menandakan ketergantungan kritis tidak tersedia sehingga keputusan fail-closed harus diberlakukan.
	KindUnavailable
)

// String mengembalikan representasi teks stabil untuk setiap Kind.
func (k Kind) String() string {
	switch k {
	case KindInvalidInput:
		return "invalid_input"
	case KindNotFound:
		return "not_found"
	case KindAlreadyExists:
		return "already_exists"
	case KindUnauthorized:
		return "unauthorized"
	case KindForbidden:
		return "forbidden"
	case KindConflict:
		return "conflict"
	case KindRateLimited:
		return "rate_limited"
	case KindUnavailable:
		return "unavailable"
	default:
		return "unknown"
	}
}

// Code merepresentasikan pengenal error domain yang stabil lintas versi API.
type Code string

// Sentinel error dasar untuk setiap kategori Kind utama.
var (
	// ErrInvalidInput adalah sentinel dasar untuk kesalahan validasi dan masukan tidak sah.
	ErrInvalidInput = newBase(KindInvalidInput, "iam.invalid_input", "invalid input")
	// ErrNotFound adalah sentinel dasar untuk entitas yang tidak ditemukan.
	ErrNotFound = newBase(KindNotFound, "iam.not_found", "resource not found")
	// ErrAlreadyExists adalah sentinel dasar untuk konflik duplikasi data unik.
	ErrAlreadyExists = newBase(KindAlreadyExists, "iam.already_exists", "resource already exists")
	// ErrUnauthorized adalah sentinel dasar untuk kegagalan autentikasi atau sesi tidak aktif.
	ErrUnauthorized = newBase(KindUnauthorized, "iam.unauthorized", "unauthorized")
	// ErrForbidden adalah sentinel dasar untuk penolakan otorisasi.
	ErrForbidden = newBase(KindForbidden, "iam.forbidden", "forbidden")
	// ErrConflict adalah sentinel dasar untuk konflik state domain.
	ErrConflict = newBase(KindConflict, "iam.conflict", "state conflict")
	// ErrRateLimited adalah sentinel dasar untuk pembatasan percobaan atau penguncian akun.
	ErrRateLimited = newBase(KindRateLimited, "iam.rate_limited", "rate limited")
	// ErrUnavailable adalah sentinel dasar saat infrastruktur identitas atau otorisasi tidak tersedia.
	ErrUnavailable = newBase(KindUnavailable, "iam.unavailable", "service unavailable")
)

// DomainError membawa konteks kategori, kode stabil, nama field opsional, pesan, serta error penyebab.
type DomainError struct {
	Kind    Kind
	Code    Code
	Message string
	Field   string
	Err     error
	isBase  bool
}

func newBase(kind Kind, code Code, message string) *DomainError {
	return &DomainError{
		Kind:    kind,
		Code:    code,
		Message: message,
		isBase:  true,
	}
}

func newSentinel(kind Kind, code Code, message string) *DomainError {
	return &DomainError{
		Kind:    kind,
		Code:    code,
		Message: message,
	}
}

// New membuat error domain baru tanpa error penyebab internal.
func New(kind Kind, code Code, message string) *DomainError {
	return &DomainError{
		Kind:    kind,
		Code:    code,
		Message: strings.TrimSpace(message),
	}
}

// Wrap membungkus error penyebab dengan konteks domain tanpa menghilangkan rantai error aslinya.
func Wrap(err error, kind Kind, code Code, message string) *DomainError {
	if err == nil {
		return nil
	}
	return &DomainError{
		Kind:    kind,
		Code:    code,
		Message: strings.TrimSpace(message),
		Err:     err,
	}
}

// NewValidation membuat error validasi masukan dengan nama field spesifik.
func NewValidation(field string, code Code, message string) *DomainError {
	return &DomainError{
		Kind:    KindInvalidInput,
		Code:    code,
		Field:   strings.TrimSpace(field),
		Message: strings.TrimSpace(message),
	}
}

// WithField membungkus sentinel atau error domain dengan konteks nama field dan pesan tambahan.
func WithField(err error, field, message string) *DomainError {
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(message)
	if msg == "" {
		if domErr, ok := errors.AsType[*DomainError](err); ok {
			msg = strings.TrimSpace(domErr.Message)
		}
	}
	return &DomainError{
		Kind:    KindOf(err),
		Code:    CodeOf(err),
		Field:   strings.TrimSpace(field),
		Message: msg,
		Err:     err,
	}
}

// Error memformat pesan error dalam huruf kecil tanpa tanda baca penutup sesuai konvensi Go.
func (d *DomainError) Error() string {
	if d == nil {
		return ""
	}
	msg := d.Message
	if d.Field != "" {
		if msg != "" {
			msg = d.Field + ": " + msg
		} else {
			msg = d.Field
		}
	}
	if d.Err != nil {
		errStr := d.Err.Error()
		if errStr != "" {
			if msg == "" {
				return errStr
			}
			// Hindari pengulangan pesan yang identik saat WithField membungkus sentinel dengan pesan yang sama.
			if errStr != d.Message {
				return msg + ": " + errStr
			}
		}
	}
	if msg != "" {
		return msg
	}
	if d.Code != "" {
		return string(d.Code)
	}
	return d.Kind.String()
}

// Unwrap membuka error penyebab di bawahnya agar kompatibel dengan errors.Is dan errors.AsType.
func (d *DomainError) Unwrap() error {
	if d == nil {
		return nil
	}
	return d.Err
}

// Is mencocokkan error domain terhadap sentinel basis berdasarkan Kind atau terhadap error spesifik dengan Code yang sama.
func (d *DomainError) Is(target error) bool {
	if d == nil || target == nil {
		return false
	}
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	if t.isBase && d.Kind != KindUnknown && d.Kind == t.Kind {
		return true
	}
	return d.Code != "" && d.Code == t.Code && d.Kind == t.Kind
}

// KindOf mengekstrak Kind dari rantai error atau mengembalikan KindUnknown bila bukan DomainError.
func KindOf(err error) Kind {
	if err == nil {
		return KindUnknown
	}
	de, ok := errors.AsType[*DomainError](err)
	if !ok {
		return KindUnknown
	}
	return de.Kind
}

// CodeOf mengekstrak Code dari rantai error atau mengembalikan string kosong bila tidak tersedia.
func CodeOf(err error) Code {
	if err == nil {
		return ""
	}
	de, ok := errors.AsType[*DomainError](err)
	if !ok {
		return ""
	}
	return de.Code
}

// FieldOf mengekstrak nama field dari rantai error atau mengembalikan string kosong bila tidak tersedia.
func FieldOf(err error) string {
	if err == nil {
		return ""
	}
	de, ok := errors.AsType[*DomainError](err)
	if !ok {
		return ""
	}
	return de.Field
}
