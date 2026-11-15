package browserauth

import (
	"context"
	"net/http"
	"strings"
	"sync"
)

type effectsContextKey struct{}

// ResponseEffects menampung perubahan cookie dan HTTP response header yang harus ditulis sebelum payload GraphQL dikirim.
type ResponseEffects struct {
	mu      sync.Mutex
	cookies []*http.Cookie
	headers http.Header
}

// NewResponseEffects membuat holder ResponseEffects kosong untuk satu siklus permintaan HTTP.
func NewResponseEffects() *ResponseEffects {
	return &ResponseEffects{
		headers: make(http.Header),
	}
}

// WithResponseEffects menyematkan ResponseEffects ke dalam context permintaan.
func WithResponseEffects(ctx context.Context, effects *ResponseEffects) context.Context {
	if effects == nil {
		return ctx
	}
	return context.WithValue(ctx, effectsContextKey{}, effects)
}

// EffectsFromContext mengambil ResponseEffects dari context bila tersedia.
func EffectsFromContext(ctx context.Context) *ResponseEffects {
	if ctx == nil {
		return nil
	}
	eff, _ := ctx.Value(effectsContextKey{}).(*ResponseEffects)
	return eff
}

// AddCookie menjadwalkan pengiriman header Set-Cookie pada respons HTTP akhir dengan deduplikasi berdasarkan nama cookie.
func (e *ResponseEffects) AddCookie(c *http.Cookie) {
	if e == nil || c == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	filtered := make([]*http.Cookie, 0, len(e.cookies)+1)
	for _, existing := range e.cookies {
		if existing != nil && existing.Name == c.Name {
			continue
		}
		filtered = append(filtered, existing)
	}
	e.cookies = append(filtered, c)
}

// SetHeader menjadwalkan penulisan header respons HTTP (misal X-CSRF-Token atau Location) secara aman tanpa karakter kontrol.
func (e *ResponseEffects) SetHeader(key, value string) {
	if e == nil || strings.TrimSpace(key) == "" {
		return
	}
	cleanVal := strings.TrimSpace(value)
	cleanVal = strings.ReplaceAll(cleanVal, "\r", "")
	cleanVal = strings.ReplaceAll(cleanVal, "\n", "")
	e.mu.Lock()
	defer e.mu.Unlock()
	e.headers.Set(key, cleanVal)
}

// ApplyTo menyalin seluruh cookie dan header ke http.ResponseWriter target.
func (e *ResponseEffects) ApplyTo(w http.ResponseWriter) {
	if e == nil || w == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	target := w.Header()
	for k, vals := range e.headers {
		for _, v := range vals {
			target.Set(k, v)
		}
	}
	for _, c := range e.cookies {
		http.SetCookie(w, c)
	}
}

// EffectResponseWriter membungkus http.ResponseWriter agar ResponseEffects otomatis diterapkan tepat sebelum header HTTP terkirim.
type EffectResponseWriter struct {
	http.ResponseWriter
	effects *ResponseEffects
	once    sync.Once
}

// NewEffectResponseWriter membentuk pembungkus writer yang menerapkan efek cookie dan header sebelum respons ditulis.
func NewEffectResponseWriter(w http.ResponseWriter, effects *ResponseEffects) *EffectResponseWriter {
	return &EffectResponseWriter{
		ResponseWriter: w,
		effects:        effects,
	}
}

func (w *EffectResponseWriter) flushEffects() {
	w.once.Do(func() {
		if w.effects != nil {
			w.effects.ApplyTo(w.ResponseWriter)
		}
	})
}

// WriteHeader menerapkan seluruh efek header/cookie sebelum mengirimkan kode status HTTP.
func (w *EffectResponseWriter) WriteHeader(statusCode int) {
	w.flushEffects()
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write menerapkan seluruh efek header/cookie sebelum menulis body respons HTTP pertama kali.
func (w *EffectResponseWriter) Write(p []byte) (int, error) {
	w.flushEffects()
	return w.ResponseWriter.Write(p)
}
