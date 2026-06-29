package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAPIError_JSONAPIArray(t *testing.T) {
	body := []byte(`{"errors":[{"status":"422","title":"Unprocessable","detail":"name is required","code":"invalid"},{"detail":"email is required"}]}`)
	e := parseAPIError(http.StatusUnprocessableEntity, body)
	assert.Equal(t, 422, e.StatusCode)
	assert.Equal(t, "invalid", e.Code)
	assert.Contains(t, e.Message, "name is required")
	assert.Contains(t, e.Message, "email is required") // both joined
}

func TestParseAPIError_LicenseAPIFlatError(t *testing.T) {
	// The License API (non-JSON:API) returns {"error":"..."}.
	e := parseAPIError(http.StatusBadRequest, []byte(`{"error":"license_key not found","valid":false}`))
	assert.Equal(t, "license_key not found", e.Message)
}

func TestParseAPIError_PlainBodyFallback(t *testing.T) {
	e := parseAPIError(http.StatusBadGateway, []byte("upstream timeout"))
	assert.Equal(t, "upstream timeout", e.Message)
}

func TestAPIError_HintsByStatus(t *testing.T) {
	cases := map[int]string{
		401: "auth login",
		403: "permission",
		404: "verify the id",
		422: "validation failed",
		429: "rate limited",
		500: "server error",
	}
	for status, want := range cases {
		e := &APIError{StatusCode: status, Message: "x"}
		assert.Contains(t, e.Error(), want, "status %d", status)
	}
}

func TestAPIError_IncludesCode(t *testing.T) {
	e := &APIError{StatusCode: 422, Message: "bad", Code: "invalid_field"}
	assert.Contains(t, e.Error(), "code: invalid_field")
}

func TestAPIError_IsRetryable(t *testing.T) {
	assert.True(t, (&APIError{StatusCode: 429}).IsRetryable())
	assert.True(t, (&APIError{StatusCode: 503}).IsRetryable())
	assert.False(t, (&APIError{StatusCode: 404}).IsRetryable())
}
