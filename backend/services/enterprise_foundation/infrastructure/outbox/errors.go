package outbox

import (
	"errors"
	"strings"
	"unicode/utf8"

	eferrors "github.com/integrione/erp-retail/backend/services/enterprise_foundation/domain/errors"
)

const maxErrorMessageRunes = 1000

// Kesalahan infrastruktur outbox standar untuk inisialisasi, rute penerbitan, dan konfigurasi webhook.
var (
	// ErrPublisherUnavailable menandakan relay aktif dijalankan tanpa Publisher eksternal yang valid.
	ErrPublisherUnavailable = eferrors.New(
		eferrors.KindUnavailable,
		eferrors.Code("enterprise_foundation.outbox.publisher_unavailable"),
		"outbox publisher is not configured",
	)
	// ErrRelayNotInitialized menandakan dependensi wajib relay outbox (transactor atau repository) belum tersedia.
	ErrRelayNotInitialized = eferrors.New(
		eferrors.KindUnavailable,
		eferrors.Code("enterprise_foundation.outbox.relay_not_initialized"),
		"outbox relay dependencies are not initialized",
	)
	// ErrRouteNotFound menandakan tidak ada rute penerbit yang terdaftar untuk tipe event tertentu.
	ErrRouteNotFound = eferrors.New(
		eferrors.KindNotFound,
		eferrors.Code("enterprise_foundation.outbox.route_not_found"),
		"outbox publisher route not found for event type",
	)
	// ErrInvalidWebhookURL menandakan URL target webhook outbox tidak valid atau tidak memenuhi kebijakan skema.
	ErrInvalidWebhookURL = eferrors.NewValidation(
		"webhook_url",
		eferrors.Code("enterprise_foundation.outbox.invalid_webhook_url"),
		"invalid outbox webhook url",
	)
)

// PermanentError menandai kesalahan pengiriman event outbox yang bersifat permanen dan tidak boleh dicoba ulang.
type PermanentError struct {
	Err error
}

// Error mengembalikan deskripsi error permanen.
func (e *PermanentError) Error() string {
	if e == nil || e.Err == nil {
		return "permanent outbox publish error"
	}
	return e.Err.Error()
}

// Unwrap mengembalikan error penyebab asli untuk mendukung errors.Is dan errors.As.
func (e *PermanentError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// MarkPermanent membungkus error sebagai kegagalan permanen sehingga relay langsung memindahkan event ke status failed.
func MarkPermanent(err error) error {
	if err == nil {
		return nil
	}
	var existing *PermanentError
	if errors.As(err, &existing) {
		return err
	}
	return &PermanentError{Err: err}
}

// IsPermanent memeriksa apakah error merupakan kegagalan pengiriman permanen yang tidak perlu retry.
func IsPermanent(err error) bool {
	if err == nil {
		return false
	}
	var permanent *PermanentError
	return errors.As(err, &permanent)
}

func normalizeErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	trimmed := strings.TrimSpace(err.Error())
	if trimmed == "" {
		return "outbox publish failed"
	}
	// Batasi jumlah rune agar konsisten dengan batas penyimpanan kolom last_error pada tabel outbox.
	if utf8.RuneCountInString(trimmed) > maxErrorMessageRunes {
		runes := []rune(trimmed)
		return string(runes[:maxErrorMessageRunes])
	}
	return trimmed
}
