package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// APIError is the typed error returned for any non-2xx response. Its Error() appends a
// status-keyed hint so the user sees a next action ("run `lsqueezy auth login`"), not just
// "request failed" — the difference between an actionable CLI and an opaque one.
type APIError struct {
	StatusCode int
	Code       string // domain-specific code if supplied (JSON:API error "code")
	Message    string // human message parsed from the body
	Details    string // any extra detail field
	Body       string // raw body, for --verbose / debugging
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	if msg == "" {
		msg = "request failed"
	}
	if e.Code != "" {
		msg = fmt.Sprintf("%s (code: %s)", msg, e.Code)
	}
	if hint := hintForStatus(e.StatusCode); hint != "" {
		return fmt.Sprintf("HTTP %d: %s — %s", e.StatusCode, msg, hint)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, msg)
}

// hintForStatus maps an HTTP status to a remediation hint. Keep these specific and
// actionable; a vague hint is no better than none.
func hintForStatus(status int) string {
	switch status {
	case http.StatusUnauthorized: // 401
		return "authentication failed; run `lsqueezy auth login` to store a valid API key"
	case http.StatusForbidden: // 403
		return "your API key lacks permission for this resource (check the key's store/test-mode scope)"
	case http.StatusNotFound: // 404
		return "not found; verify the id with `lsqueezy <resource> list`"
	case http.StatusUnprocessableEntity: // 422
		return "validation failed; check required attributes/relationships against the API docs"
	case http.StatusTooManyRequests: // 429
		return "rate limited (300 req/min); slow down or wait before retrying"
	}
	if status >= 500 {
		return "server error, usually transient; retry shortly"
	}
	if status == http.StatusBadRequest { // 400
		return "bad request; check required fields and flag values"
	}
	return ""
}

// IsRetryable reports whether an APIError represents a transient condition worth retrying.
func (e *APIError) IsRetryable() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

// jsonAPIErrors is the JSON:API error document: {"errors":[{"status","title","detail","code"}]}.
type jsonAPIErrors struct {
	Errors []struct {
		Status string `json:"status"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
		Code   string `json:"code"`
	} `json:"errors"`
	// Lemon Squeezy's License API (a non-JSON:API endpoint) returns {"error":"..."} instead.
	Error   string `json:"error"`
	Message string `json:"message"`
}

// parseAPIError extracts a message/code from Lemon Squeezy's error shapes: the JSON:API
// errors[] array (most endpoints) or the flat {"error":…} the License API uses. Multiple
// JSON:API errors are joined so the user sees every validation problem at once.
func parseAPIError(status int, body []byte) *APIError {
	e := &APIError{StatusCode: status, Body: string(body)}
	var doc jsonAPIErrors
	if json.Unmarshal(body, &doc) == nil {
		switch {
		case len(doc.Errors) > 0:
			msgs := make([]string, 0, len(doc.Errors))
			for _, je := range doc.Errors {
				switch {
				case je.Detail != "":
					msgs = append(msgs, je.Detail)
				case je.Title != "":
					msgs = append(msgs, je.Title)
				}
				if e.Code == "" {
					e.Code = je.Code
				}
			}
			e.Message = strings.Join(msgs, "; ")
		case doc.Error != "":
			e.Message = doc.Error
		case doc.Message != "":
			e.Message = doc.Message
		}
	}
	if e.Message == "" {
		if s := strings.TrimSpace(string(body)); s != "" && len(s) < 200 {
			e.Message = s
		}
	}
	return e
}
