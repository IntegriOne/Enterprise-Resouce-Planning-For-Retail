package outbox_test

import (
	"testing"
	"time"

	"github.com/integrione/erp-retail/backend/services/enterprise_foundation/infrastructure/outbox"
)

func TestBackoffPolicyNextDelay(t *testing.T) {
	t.Parallel()

	policy := outbox.BackoffPolicy{
		Base:        2 * time.Second,
		Max:         30 * time.Second,
		JitterRatio: 0,
	}

	cases := []struct {
		name    string
		attempt int
		want    time.Duration
	}{
		{name: "attempt zero clamps to base", attempt: 0, want: 2 * time.Second},
		{name: "attempt one returns base", attempt: 1, want: 2 * time.Second},
		{name: "attempt two doubles", attempt: 2, want: 4 * time.Second},
		{name: "attempt three quadruples", attempt: 3, want: 8 * time.Second},
		{name: "attempt four yields 16s", attempt: 4, want: 16 * time.Second},
		{name: "attempt five caps at max", attempt: 5, want: 30 * time.Second},
		{name: "large attempt remains overflow-safe", attempt: 100, want: 30 * time.Second},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := policy.NextDelay(tc.attempt); got != tc.want {
				t.Fatalf("got delay %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBackoffPolicyNormalizationAndJitter(t *testing.T) {
	t.Parallel()

	t.Run("zero values normalize to package defaults", func(t *testing.T) {
		t.Parallel()
		var zero outbox.BackoffPolicy
		if got := zero.NextDelay(1); got != outbox.DefaultBackoffBase {
			t.Fatalf("got base %v, want %v", got, outbox.DefaultBackoffBase)
		}
		if got := zero.NextDelay(10); got != outbox.DefaultBackoffMax {
			t.Fatalf("got max %v, want %v", got, outbox.DefaultBackoffMax)
		}
	})

	t.Run("jitter respects max bound and deterministic JitterFn", func(t *testing.T) {
		t.Parallel()
		policy := outbox.BackoffPolicy{
			Base:        10 * time.Second,
			Max:         12 * time.Second,
			JitterRatio: 0.5,
			JitterFn: func(maxJitter time.Duration) time.Duration {
				return maxJitter
			},
		}
		if got := policy.NextDelay(1); got != 12*time.Second {
			t.Fatalf("expected capped delay 12s, got %v", got)
		}
	})
}
