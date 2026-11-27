package outbox

import (
	"math/rand/v2"
	"time"
)

const (
	// DefaultBackoffBase adalah jeda awal default sebelum percobaan pengiriman ulang pertama.
	DefaultBackoffBase = 1 * time.Second
	// DefaultBackoffMax adalah batas maksimum jeda eksponensial antar percobaan pengiriman ulang.
	DefaultBackoffMax = 60 * time.Second
)

// BackoffPolicy mengatur perhitungan jeda eksponensial dengan batas maksimum dan jitter opsional.
type BackoffPolicy struct {
	Base        time.Duration
	Max         time.Duration
	JitterRatio float64
	JitterFn    func(maxJitter time.Duration) time.Duration
}

// DefaultBackoffPolicy mengembalikan kebijakan backoff standar untuk pengiriman outbox Enterprise Foundation.
func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		Base:        DefaultBackoffBase,
		Max:         DefaultBackoffMax,
		JitterRatio: 0.1,
	}
}

func (p BackoffPolicy) normalized() BackoffPolicy {
	base := p.Base
	if base <= 0 {
		base = DefaultBackoffBase
	}
	maxDelay := p.Max
	if maxDelay <= 0 {
		maxDelay = DefaultBackoffMax
	}
	if maxDelay < base {
		maxDelay = base
	}
	ratio := p.JitterRatio
	if ratio < 0 {
		ratio = 0
	} else if ratio > 0.5 {
		ratio = 0.5
	}
	return BackoffPolicy{
		Base:        base,
		Max:         maxDelay,
		JitterRatio: ratio,
		JitterFn:    p.JitterFn,
	}
}

// NextDelay menghitung jeda waktu tunggu untuk nomor percobaan gagal ke-attempt (mulai dari 1).
func (p BackoffPolicy) NextDelay(attempt int) time.Duration {
	cfg := p.normalized()
	if attempt <= 1 {
		return cfg.applyJitter(cfg.Base)
	}

	delay := cfg.Base
	shifts := attempt - 1
	for range shifts {
		if delay >= cfg.Max {
			delay = cfg.Max
			break
		}
		// Lindungi terhadap integer overflow ketika menggandakan durasi backoff.
		if delay > cfg.Max/2 {
			delay = cfg.Max
			break
		}
		delay *= 2
	}
	if delay > cfg.Max {
		delay = cfg.Max
	}
	return cfg.applyJitter(delay)
}

func (p BackoffPolicy) applyJitter(delay time.Duration) time.Duration {
	if p.JitterRatio <= 0 {
		return delay
	}
	maxJitter := time.Duration(float64(delay) * p.JitterRatio)
	if maxJitter <= 0 {
		return delay
	}

	var jitter time.Duration
	if p.JitterFn != nil {
		jitter = p.JitterFn(maxJitter)
	} else {
		jitter = time.Duration(rand.Int64N(int64(maxJitter) + 1))
	}
	if jitter < 0 {
		jitter = 0
	} else if jitter > maxJitter {
		jitter = maxJitter
	}

	result := delay + jitter
	if result > p.Max {
		return p.Max
	}
	return result
}
