package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient spins up an httptest server with the given handler and returns a Client
// pointed at it. Fast rate limit so tests don't sleep; short retry policy.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New(srv.URL, "test-key", WithHTTPClient(srv.Client()), WithRateLimit(1000))
	c.retry = retryPolicy{MaxRetries: 2, BaseDelay: 1, MaxDelay: 2}
	return c
}
