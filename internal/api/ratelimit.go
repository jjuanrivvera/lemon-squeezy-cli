package api

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// rateLimiter enforces a minimum interval between requests and adapts to the server's
// quota signals: it slows as the remaining budget depletes and halves the rate on a 429.
// Lemon Squeezy is generous, but a well-behaved client reads quota headers when present.
type rateLimiter struct {
	mu       sync.Mutex
	interval time.Duration // current min gap between requests
	minGap   time.Duration // floor (the configured base rate)
	maxGap   time.Duration // ceiling so a depleted budget can't stall forever
	last     time.Time
}

func newRateLimiter(rps float64) *rateLimiter {
	if rps <= 0 {
		rps = 10 // sane default
	}
	base := time.Duration(float64(time.Second) / rps)
	return &rateLimiter{interval: base, minGap: base, maxGap: 5 * time.Second}
}

// wait blocks until the next request is permitted or ctx is cancelled.
func (r *rateLimiter) wait(ctx context.Context) error {
	r.mu.Lock()
	now := time.Now()
	wait := time.Duration(0)
	if !r.last.IsZero() {
		elapsed := now.Sub(r.last)
		if elapsed < r.interval {
			wait = r.interval - elapsed
		}
	}
	r.last = now.Add(wait)
	r.mu.Unlock()

	if wait <= 0 {
		return nil
	}
	return sleepCtx(ctx, wait)
}

// observe adjusts the interval from a response's rate-limit headers. Halve the rate on a
// 429; otherwise ease back toward the base rate as budget allows.
func (r *rateLimiter) observe(resp *http.Response) {
	if resp == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if resp.StatusCode == http.StatusTooManyRequests {
		r.interval *= 2
		if r.interval > r.maxGap {
			r.interval = r.maxGap
		}
		return
	}

	rem := headerInt(resp, "X-RateLimit-Remaining")
	limit := headerInt(resp, "X-RateLimit-Limit")
	if rem >= 0 && limit > 0 {
		// Below 10% of budget, double the gap; otherwise relax toward the base rate.
		if rem*10 < limit {
			r.interval *= 2
			if r.interval > r.maxGap {
				r.interval = r.maxGap
			}
		} else if r.interval > r.minGap {
			r.interval = r.minGap
		}
	}
}

func headerInt(resp *http.Response, key string) int {
	v := resp.Header.Get(key)
	if v == "" {
		return -1
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return -1
	}
	return n
}
