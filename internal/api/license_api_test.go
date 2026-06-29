package api

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicense_Activate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/licenses/activate", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		_ = r.ParseForm()
		assert.Equal(t, "key-1", r.Form.Get("license_key"))
		assert.Equal(t, "box", r.Form.Get("instance_name"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"activated":true,"instance":{"id":"i1"},"license_key":{"status":"active"}}`))
	})
	res, err := c.License().Activate(context.Background(), "key-1", "box")
	require.NoError(t, err)
	require.NotNil(t, res.Activated)
	assert.True(t, *res.Activated)
}

func TestLicense_Validate_And_Deactivate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/licenses/validate":
			_ = r.ParseForm()
			assert.Equal(t, "inst-9", r.Form.Get("instance_id"))
			_, _ = w.Write([]byte(`{"valid":true}`))
		case "/licenses/deactivate":
			_, _ = w.Write([]byte(`{"deactivated":true}`))
		}
	})
	v, err := c.License().Validate(context.Background(), "key-1", "inst-9")
	require.NoError(t, err)
	require.NotNil(t, v.Valid)
	assert.True(t, *v.Valid)

	d, err := c.License().Deactivate(context.Background(), "key-1", "inst-9")
	require.NoError(t, err)
	require.NotNil(t, d.Deactivated)
	assert.True(t, *d.Deactivated)
}

func TestLicense_ErrorBody(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"license_key not found","valid":false}`))
	})
	_, err := c.License().Validate(context.Background(), "nope", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "license_key not found")
}

func TestLicense_DryRun(t *testing.T) {
	var buf bytes.Buffer
	c := New("https://api.lemonsqueezy.com/v1", "secret", WithDryRun(true, &buf))
	res, err := c.License().Activate(context.Background(), "key-1", "box")
	require.NoError(t, err)
	assert.Nil(t, res)
	out := buf.String()
	assert.Contains(t, out, "curl -X POST")
	assert.Contains(t, out, "/licenses/activate")
	assert.Contains(t, out, "Accept: application/json")
	assert.Contains(t, out, "license_key=key-1")
	assert.NotContains(t, out, "secret") // bearer redacted
}
