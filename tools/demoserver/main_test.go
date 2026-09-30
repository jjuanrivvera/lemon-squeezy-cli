package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, h http.Handler, path string) map[string]any {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", path, rec.Code, rec.Body)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("GET %s did not answer JSON: %v", path, err)
	}
	return body
}

// The recording is only as trustworthy as the stand-in behind it: every path the
// tape drives has to answer in the envelope the CLI expects.
func TestDemoHandlerAnswersTheRecordedPaths(t *testing.T) {
	h := demoHandler()
	for _, path := range []string{"/v1/products", "/v1/orders", "/v1/discounts"} {
		body := get(t, h, path)
		if _, ok := body["data"].([]any); !ok {
			t.Errorf("GET %s should answer a JSON:API list: %+v", path, body)
		}
	}
	if list := get(t, h, "/v1/products")["data"].([]any); len(list) != 2 {
		t.Errorf("want the two invented products, got %d", len(list))
	}
	// The discount list starts empty on purpose: the recording fills it, which
	// is what makes the write visibly take effect.
	if list := get(t, h, "/v1/discounts")["data"].([]any); len(list) != 0 {
		t.Errorf("discounts should start empty, got %d", len(list))
	}
}

// A path nobody recorded is a 404, never an empty success that would look like
// a real answer on screen.
func TestDemoHandlerRefusesEverythingElse(t *testing.T) {
	h := demoHandler()
	for _, r := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/v1/stores", nil),
		httptest.NewRequest(http.MethodPost, "/v1/orders", strings.NewReader("{}")),
		httptest.NewRequest(http.MethodGet, "/products", nil), // outside the /v1 prefix
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s %s = %d, want 404", r.Method, r.URL.Path, rec.Code)
		}
	}
}

// Creating a discount has to be visible in the next listing: a write that left
// no trace would make the recording show a result that never happened.
func TestDemoHandlerCreateThenList(t *testing.T) {
	h := demoHandler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/discounts", strings.NewReader(
		`{"data":{"type":"discounts","attributes":{"name":"Demo launch","code":"DEMO20","amount":20,"amount_type":"percent"},
		  "relationships":{"store":{"data":{"type":"stores","id":"1"}}}}}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST discounts = %d: %s", rec.Code, rec.Body)
	}
	var created struct {
		Data struct {
			ID         string         `json:"id"`
			Attributes map[string]any `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	// What was sent comes back, plus the status the API would have assigned.
	if created.Data.Attributes["code"] != "DEMO20" || created.Data.Attributes["status"] != "published" {
		t.Errorf("created discount = %+v", created.Data.Attributes)
	}

	list := get(t, h, "/v1/discounts")["data"].([]any)
	if len(list) != 1 {
		t.Fatalf("the created discount should now be listed, got %d", len(list))
	}
}

func TestDemoHandlerRejectsABodyWithNoAttributes(t *testing.T) {
	h := demoHandler()
	for _, body := range []string{"not json", `{"data":{"type":"discounts"}}`} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/discounts", strings.NewReader(body)))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("POST %q = %d, want 400", body, rec.Code)
		}
	}
}
