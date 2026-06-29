package api

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Defaults(t *testing.T) {
	c := New("", "k")
	assert.Equal(t, DefaultBaseURL, c.BaseURL)
	c2 := New("https://x/", "k")
	assert.Equal(t, "https://x", c2.BaseURL) // trailing slash trimmed
}

func TestEncodeSorted_Deterministic(t *testing.T) {
	c := New("https://api.example.com/v1", "k")
	got := c.buildURL("products", map[string][]string{
		"page[size]":   {"100"},
		"filter[a]":    {"1"},
		"page[number]": {"2"},
	})
	// Keys sorted; brackets percent-escaped.
	assert.Equal(t,
		"https://api.example.com/v1/products?filter%5Ba%5D=1&page%5Bnumber%5D=2&page%5Bsize%5D=100",
		got)
}

func TestDryRun_PrintsRedactedCurlWithBearerAndAccept(t *testing.T) {
	var buf bytes.Buffer
	c := New("https://api.lemonsqueezy.com/v1", "secret-token", WithDryRun(true, &buf))
	err := c.doJSON(context.Background(), http.MethodGet, "stores", nil, nil, nil)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "curl -X GET")
	assert.Contains(t, out, "https://api.lemonsqueezy.com/v1/stores")
	assert.Contains(t, out, "Authorization: Bearer "+redacted)
	assert.Contains(t, out, "Accept: "+MediaType)
	assert.NotContains(t, out, "secret-token") // token must be redacted by default
}

func TestDryRun_ShowToken(t *testing.T) {
	var buf bytes.Buffer
	c := New("https://x", "secret-token", WithDryRun(true, &buf))
	c.ShowToken = true
	_ = c.doJSON(context.Background(), http.MethodGet, "stores", nil, nil, nil)
	assert.Contains(t, buf.String(), "secret-token")
}

func TestGetJSON_SendsHeaders(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, MediaType, r.Header.Get("Accept"))
		w.Header().Set("Content-Type", MediaType)
		_, _ = w.Write([]byte(`{"data":{"type":"users","id":"1","attributes":{}}}`))
	})
	var doc singleDoc
	require.NoError(t, c.GetJSON(context.Background(), "users/me", nil, &doc))
	assert.Equal(t, ID("1"), doc.Data.ID)
}

func TestDo_RetriesOn5xxThenSucceeds(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", MediaType)
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	var doc listDoc
	require.NoError(t, c.GetJSON(context.Background(), "products", nil, &doc))
	assert.Equal(t, 2, calls) // one retry
}

func TestDo_DoesNotRetryPOST(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	})
	err := c.doJSON(context.Background(), http.MethodPost, "checkouts", nil, strings.NewReader("{}"), nil)
	require.Error(t, err)
	assert.Equal(t, 1, calls) // POST is not idempotent: no auto-retry
}

func TestShellQuote(t *testing.T) {
	assert.Equal(t, `'a b'`, shellQuote("a b"))
	assert.Equal(t, `'it'\''s'`, shellQuote("it's"))
}
