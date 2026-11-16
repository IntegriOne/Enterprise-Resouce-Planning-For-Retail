package service

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"time"

	hrerrors "github.com/integrione/erp-retail/backend/services/human_resource/domain/errors"
)

// SystemClock menggunakan waktu server UTC, bukan tanggal dari browser.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// UUIDGenerator menghasilkan UUID v4 dengan sumber acak kriptografis standar.
type UUIDGenerator struct{}

func (UUIDGenerator) NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate uuid: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
func isNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return v.IsNil()
	}
	return false
}

// RequireDependency menolak interface kosong maupun typed nil sebelum memanggil adapter opsional.
func RequireDependency(value any) error {
	if isNil(value) {
		return hrerrors.ErrUnavailable
	}
	return nil
}
