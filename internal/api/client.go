// Package api is the generic HTTP core for the Lemon Squeezy API: one client + one generic
// Resource[T] power every resource. Lemon Squeezy speaks JSON:API (application/vnd.api+json),
// so the envelope handling (data/attributes/relationships, page-based pagination) lives here
// once — adding a resource never touches this package.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the Lemon Squeezy API v1 root.
	DefaultBaseURL = "https://api.lemonsqueezy.com/v1"
	// MediaType is the JSON:API content type Lemon Squeezy requires on both Accept and
	// Content-Type. A request without it is rejected, so it is set on every call.
	MediaType = "application/vnd.api+json"
	// defaultRPS is the safe base rate. Lemon Squeezy allows 300 requests/minute (5/s); we
	// stay at the ceiling and back off on 429 rather than guess a lower fixed rate.
	defaultRPS = 5
	redacted   = "REDACTED"
)

// Client is the authenticated HTTP client. It is safe for sequential use by the CLI; the
// rate limiter serializes requests internally.
type Client struct {
	BaseURL    string
	apiKey     string
	httpClient *http.Client
	limiter    *rateLimiter
	retry      retryPolicy

	// DryRun, when true, prints the equivalent curl to DryRunOut and performs no request.
	DryRun    bool
	ShowToken bool // reveal the key in dry-run/curl output instead of redacting
	DryRunOut io.Writer

	// Verbose enables request/response logging to VerboseOut.
	Verbose    bool
	VerboseOut io.Writer
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the underlying *http.Client (used by tests).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// WithRateLimit sets the base requests-per-second.
func WithRateLimit(rps float64) Option { return func(c *Client) { c.limiter = newRateLimiter(rps) } }

// WithDryRun toggles dry-run mode and its output sink.
func WithDryRun(on bool, out io.Writer) Option {
	return func(c *Client) { c.DryRun = on; c.DryRunOut = out }
}

// New builds a Client. baseURL defaults to DefaultBaseURL when empty.
func New(baseURL, apiKey string, opts ...Option) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	c := &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		limiter:    newRateLimiter(defaultRPS),
		retry:      defaultRetryPolicy(),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// buildURL joins the base URL, path, and query params deterministically (sorted keys).
func (c *Client) buildURL(path string, query url.Values) string {
	u := c.BaseURL + "/" + strings.TrimLeft(path, "/")
	if len(query) == 0 {
		return u
	}
	return u + "?" + encodeSorted(query)
}

// encodeSorted encodes query params with sorted keys so output (and the dry-run curl) is
// stable across runs — never rely on map iteration order. Keys like "page[size]" are
// percent-escaped by url.QueryEscape, which Lemon Squeezy accepts.
func encodeSorted(v url.Values) string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		for _, val := range v[k] {
			if b.Len() > 0 {
				b.WriteByte('&')
			}
			b.WriteString(url.QueryEscape(k))
			b.WriteByte('=')
			b.WriteString(url.QueryEscape(val))
		}
	}
	return b.String()
}

// Do performs an authenticated request with retry, rate limiting, and dry-run support.
// In dry-run it prints the equivalent curl and returns (nil, nil). The caller owns closing
// resp.Body when a response is returned.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Response, error) {
	var bodyBytes []byte
	if body != nil {
		b, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		bodyBytes = b
	}

	fullURL := c.buildURL(path, query)

	if c.DryRun {
		c.printCurl(method, fullURL, bodyBytes)
		return nil, nil
	}

	var lastErr error
	for attempt := 0; attempt <= c.retry.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, c.retry.backoff(attempt-1)); err != nil {
				return nil, err
			}
		}
		if err := c.limiter.wait(ctx); err != nil {
			return nil, err
		}

		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
		if err != nil {
			return nil, err
		}
		// Bearer auth + JSON:API media type on every request (Lemon Squeezy requires both).
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Accept", MediaType)
		if bodyBytes != nil {
			req.Header.Set("Content-Type", MediaType)
		}

		resp, err := c.httpClient.Do(req)
		c.limiter.observe(resp)

		if shouldRetry(method, resp, err) && attempt < c.retry.MaxRetries {
			if resp != nil {
				_ = resp.Body.Close()
			}
			lastErr = err
			continue
		}
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("request to %s failed after retries", fullURL)
}

// doJSON performs a request and decodes a JSON response into out (if non-nil), turning any
// non-2xx into a typed APIError. JSON:API carries pagination in the body (meta/links), not
// headers, so no header is returned.
func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body io.Reader, out any) error {
	resp, err := c.Do(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	if resp == nil { // dry-run
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if c.Verbose && c.VerboseOut != nil {
		_, _ = fmt.Fprintf(c.VerboseOut, "%s %s -> %d\n", method, path, resp.StatusCode)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, data)
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// GetJSON is a convenience GET-and-decode used by read paths and auth verification.
func (c *Client) GetJSON(ctx context.Context, path string, query url.Values, out any) error {
	return c.doJSON(ctx, http.MethodGet, path, query, nil, out)
}

// printCurl writes the equivalent, shell-escaped curl command, redacting the bearer token
// unless ShowToken is set. Indispensable for debugging and teaching.
func (c *Client) printCurl(method, fullURL string, body []byte) {
	out := c.DryRunOut
	if out == nil {
		return
	}
	key := c.apiKey
	if !c.ShowToken {
		key = redacted
	}
	var b strings.Builder
	b.WriteString("curl -X ")
	b.WriteString(method)
	fmt.Fprintf(&b, " %s", shellQuote(fullURL))
	fmt.Fprintf(&b, " -H %s", shellQuote("Authorization: Bearer "+key))
	fmt.Fprintf(&b, " -H %s", shellQuote("Accept: "+MediaType))
	if len(body) > 0 {
		fmt.Fprintf(&b, " -H %s", shellQuote("Content-Type: "+MediaType))
		fmt.Fprintf(&b, " -d %s", shellQuote(string(body)))
	}
	_, _ = fmt.Fprintln(out, b.String())
}

// shellQuote single-quotes a string for safe pasting into a POSIX shell.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
