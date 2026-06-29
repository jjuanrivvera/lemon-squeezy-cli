package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// The License API (https://docs.lemonsqueezy.com/api/license-api) is deliberately NOT
// JSON:API: it accepts form/JSON params and returns a flat object, and the API key is
// optional because license checks are meant to run on a customer's machine. It therefore
// bypasses the generic Resource[T] path and uses a small dedicated client method that still
// reuses auth (when present), the rate limiter, and dry-run.

// LicenseResult is the flat response from an activate/validate/deactivate call.
type LicenseResult struct {
	Activated   *bool          `json:"activated,omitempty"`
	Deactivated *bool          `json:"deactivated,omitempty"`
	Valid       *bool          `json:"valid,omitempty"`
	Error       string         `json:"error,omitempty"`
	LicenseKey  map[string]any `json:"license_key,omitempty"`
	Instance    map[string]any `json:"instance,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
}

// LicenseService drives the License API endpoints.
type LicenseService struct{ c *Client }

// License returns a handle to the License API.
func (c *Client) License() *LicenseService { return &LicenseService{c} }

// Activate registers a new instance for a license key (POST /v1/licenses/activate).
func (s *LicenseService) Activate(ctx context.Context, key, instanceName string) (*LicenseResult, error) {
	return s.call(ctx, "activate", url.Values{"license_key": {key}, "instance_name": {instanceName}})
}

// Validate checks a license key, optionally scoped to an instance (POST /v1/licenses/validate).
func (s *LicenseService) Validate(ctx context.Context, key, instanceID string) (*LicenseResult, error) {
	v := url.Values{"license_key": {key}}
	if instanceID != "" {
		v.Set("instance_id", instanceID)
	}
	return s.call(ctx, "validate", v)
}

// Deactivate removes an instance from a license key (POST /v1/licenses/deactivate).
func (s *LicenseService) Deactivate(ctx context.Context, key, instanceID string) (*LicenseResult, error) {
	return s.call(ctx, "deactivate", url.Values{"license_key": {key}, "instance_id": {instanceID}})
}

// call performs a form-encoded POST to a License API action. POST is not auto-retried (these
// operations are not idempotent). A 4xx still carries a useful {"error":…} body, so it is
// surfaced as an APIError rather than swallowed.
func (s *LicenseService) call(ctx context.Context, action string, form url.Values) (*LicenseResult, error) {
	c := s.c
	fullURL := c.BaseURL + "/licenses/" + action
	body := form.Encode()

	if c.DryRun {
		c.printLicenseCurl(fullURL, body)
		return nil, nil
	}
	if err := c.limiter.wait(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	c.limiter.observe(resp)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp.StatusCode, data)
	}
	var out LicenseResult
	if len(data) > 0 {
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, fmt.Errorf("decode license response: %w", err)
		}
	}
	return &out, nil
}

// printLicenseCurl emits the equivalent curl for a License API call (form body, JSON accept),
// redacting the bearer token unless ShowToken is set.
func (c *Client) printLicenseCurl(fullURL, body string) {
	out := c.DryRunOut
	if out == nil {
		return
	}
	var b strings.Builder
	b.WriteString("curl -X POST ")
	fmt.Fprintf(&b, "%s", shellQuote(fullURL))
	b.WriteString(" -H 'Accept: application/json'")
	if c.apiKey != "" {
		key := c.apiKey
		if !c.ShowToken {
			key = redacted
		}
		fmt.Fprintf(&b, " -H %s", shellQuote("Authorization: Bearer "+key))
	}
	fmt.Fprintf(&b, " -d %s", shellQuote(body))
	_, _ = fmt.Fprintln(out, b.String())
}
