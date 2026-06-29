package api

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsIdempotent(t *testing.T) {
	assert.True(t, isIdempotent(http.MethodGet))
	assert.True(t, isIdempotent(http.MethodDelete))
	assert.True(t, isIdempotent(http.MethodPut))
	assert.False(t, isIdempotent(http.MethodPost))
	assert.False(t, isIdempotent(http.MethodPatch))
}

func TestShouldRetry(t *testing.T) {
	cases := []struct {
		name   string
		method string
		status int
		err    error
		want   bool
	}{
		{"GET 503", http.MethodGet, 503, nil, true},
		{"GET 429", http.MethodGet, 429, nil, true},
		{"GET 200", http.MethodGet, 200, nil, false},
		{"POST 503", http.MethodPost, 503, nil, false},
		{"GET network err", http.MethodGet, 0, errors.New("conn reset"), true},
		{"GET ctx canceled", http.MethodGet, 0, context.Canceled, false},
		{"POST network err", http.MethodPost, 0, errors.New("x"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var resp *http.Response
			if tc.status != 0 {
				resp = &http.Response{StatusCode: tc.status}
			}
			assert.Equal(t, tc.want, shouldRetry(tc.method, resp, tc.err))
		})
	}
}

func TestBackoffWithinBounds(t *testing.T) {
	p := retryPolicy{MaxRetries: 5, BaseDelay: 10 * time.Millisecond, MaxDelay: 100 * time.Millisecond}
	for i := 0; i < 10; i++ {
		d := p.backoff(i)
		assert.GreaterOrEqual(t, d, time.Duration(0))
		assert.LessOrEqual(t, d, p.MaxDelay)
	}
}

func TestSleepCtxCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := sleepCtx(ctx, time.Hour)
	require.ErrorIs(t, err, context.Canceled)
}

func TestSleepCtxCompletes(t *testing.T) {
	require.NoError(t, sleepCtx(context.Background(), time.Millisecond))
}
