package api

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_Spacing(t *testing.T) {
	rl := newRateLimiter(100) // 10ms gap
	ctx := context.Background()
	start := time.Now()
	require.NoError(t, rl.wait(ctx)) // first is immediate
	require.NoError(t, rl.wait(ctx)) // second waits ~10ms
	assert.GreaterOrEqual(t, time.Since(start), 8*time.Millisecond)
}

func TestRateLimiter_HalvesOn429(t *testing.T) {
	rl := newRateLimiter(100)
	before := rl.interval
	rl.observe(&http.Response{StatusCode: http.StatusTooManyRequests})
	assert.Greater(t, rl.interval, before)
}

func TestRateLimiter_SlowsWhenBudgetLow(t *testing.T) {
	rl := newRateLimiter(100)
	before := rl.interval
	resp := &http.Response{StatusCode: 200, Header: http.Header{}}
	resp.Header.Set("X-RateLimit-Limit", "100")
	resp.Header.Set("X-RateLimit-Remaining", "5") // 5% left
	rl.observe(resp)
	assert.Greater(t, rl.interval, before)
}

func TestRateLimiter_RelaxesWhenBudgetHigh(t *testing.T) {
	rl := newRateLimiter(100)
	rl.interval = rl.maxGap // pretend we slowed earlier
	resp := &http.Response{StatusCode: 200, Header: http.Header{}}
	resp.Header.Set("X-RateLimit-Limit", "100")
	resp.Header.Set("X-RateLimit-Remaining", "90")
	rl.observe(resp)
	assert.Equal(t, rl.minGap, rl.interval)
}

func TestRateLimiter_WaitCancels(t *testing.T) {
	rl := newRateLimiter(0.001) // very slow => long gap on the 2nd call
	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, rl.wait(context.Background()))
	cancel()
	err := rl.wait(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestHeaderInt(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	assert.Equal(t, -1, headerInt(resp, "X-Missing"))
	resp.Header.Set("X-Num", "abc")
	assert.Equal(t, -1, headerInt(resp, "X-Num"))
	resp.Header.Set("X-Num", "42")
	assert.Equal(t, 42, headerInt(resp, "X-Num"))
}
