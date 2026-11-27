// Package eferrors menyediakan tipe error domain terstruktur dan sentinel error untuk layanan Enterprise Foundation.
package eferrors

import (
	"errors"
	"strings"
)

// Kind mengelompokkan kategori semantik error domain agar mudah dipetakan ke kode gRPC secara konsisten.
type Kind uint8

const (
	// KindUnknown menandakan error yang belum diklasifikasikan ke kategori domain tertentu.
	KindUnknown Kind = iota
	// KindInvalidInput menandakan pelanggaran validasi masukan atau invarian entitas.
	KindInvalidInput
	// KindNotFound menandakan entitas atau relasi yang diminta tidak ditemukan.
	KindNotFound
	// KindAlreadyExists menandakan duplikasi kunci unik atau kode master data yang sudah terdaftar.
	KindAlreadyExists
	// KindUnauthorized menandakan pemanggil tidak terautentikasi atau konteks gateway tidak valid.
	KindUnauthorized
	// KindForbidden menandakan principal terautentikasi tetapi tidak memiliki izin atau ruang lingkup yang sesuai.
	KindForbidden
	// KindConflict menandakan transisi state atau modifikasi yang bertentangan dengan kondisi terkini.
	KindConflict
	// KindRateLimited menandakan akses dibatasi sementara karena batas laju terlampaui.
	KindRateLimited
	// KindUnavailable menandakan ketergantungan kritis tidak tersedia sehingga mode fail-closed diberlakukan.
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

// Sentinel error dasar untuk setiap kategori Kind utama pada Enterprise Foundation.
var (
	// ErrInvalidInput adalah sentinel dasar untuk kesalahan validasi dan masukan tidak sah.
	ErrInvalidInput = newBase(KindInvalidInput, "enterprise_foundation.invalid_input", "invalid input")
	// ErrNotFound adalah sentinel dasar untuk entitas yang tidak ditemukan.
	ErrNotFound = newBase(KindNotFound, "enterprise_foundation.not_found", "resource not found")
	// ErrAlreadyExists adalah sentinel dasar untuk konflik duplikasi data unik.
	ErrAlreadyExists = newBase(KindAlreadyExists, "enterprise_foundation.already_exists", "resource already exists")
	// ErrUnauthorized adalah sentinel dasar untuk konteks pemanggil tidak sah.
	ErrUnauthorized = newBase(KindUnauthorized, "enterprise_foundation.unauthorized", "unauthorized")
	// ErrForbidden adalah sentinel dasar untuk penolakan otorisasi.
	ErrForbidden = newBase(KindForbidden, "enterprise_foundation.forbidden", "forbidden")
	// ErrConflict adalah sentinel dasar untuk konflik state domain.
	ErrConflict = newBase(KindConflict, "enterprise_foundation.conflict", "state conflict")
	// ErrRateLimited adalah sentinel dasar untuk pembatasan laju.
	ErrRateLimited = newBase(KindRateLimited, "enterprise_foundation.rate_limited", "rate limited")
	// ErrUnavailable adalah sentinel dasar saat infrastruktur kritis tidak tersedia.
	ErrUnavailable = newBase(KindUnavailable, "enterprise_foundation.unavailable", "service unavailable")
)

// DomainError membawa kategori, kode stabil, nama field opsional, pesan, serta error penyebab.
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

// Wrap membungkus error penyebab dengan konteks domain tanpa memutus rantai error aslinya.
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
	return &DomainError{
		Kind:    KindOf(err),
		Code:    CodeOf(err),
		Field:   strings.TrimSpace(field),
		Message: strings.TrimSpace(message),
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
			return d.Field + ": " + msg
		}
		return d.Field + ": invalid value"
	}
	if msg != "" {
		return msg
	}
	if d.Code != "" {
		return string(d.Code)
	}
	return d.Kind.String()
}

// Unwrap mengembalikan error penyebab untuk mendukung errors.Is dan errors.As.
func (d *DomainError) Unwrap() error {
	if d == nil {
		return nil
	}
	return d.Err
}

// Is mencocokkan error domain berdasarkan sentinel kategori (isBase) maupun kode domain yang identik.
func (d *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok || d == nil || t == nil {
		return false
	}
	if t.isBase {
		return d.Kind == t.Kind
	}
	if t.Code != "" && d.Code != "" {
		return d.Code == t.Code
	}
	return d.Kind == t.Kind
}

// KindOf mengekstrak kategori Kind dari rantai error bila mengandung DomainError.
func KindOf(err error) Kind {
	if domErr, ok := errors.AsType[*DomainError](err); ok {
		return domErr.Kind
	}
	return KindUnknown
}

// CodeOf mengekstrak kode stabil Code dari rantai error bila mengandung DomainError.
func CodeOf(err error) Code {
	if domErr, ok := errors.AsType[*DomainError](err); ok {
		return domErr.Code
	}
	return ""
}

// FieldOf mengekstrak nama field dari rantai error bila mengandung DomainError.
func FieldOf(err error) string {
	if domErr, ok := errors.AsType[*DomainError](err); ok {
		return domErr.Field
	}
	return ""
}
