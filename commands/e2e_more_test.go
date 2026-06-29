package commands_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_OrdersGenerateInvoice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/orders/12/generate-invoice", r.URL.Path)
		assert.Equal(t, "Acme", r.URL.Query().Get("name"))
		assert.Equal(t, "US", r.URL.Query().Get("country"))
		jsonAPI(w, `{"meta":{"urls":{"download_invoice":"https://x/invoice.pdf"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "orders", "generate-invoice", "12", "--name", "Acme", "--country", "US", "-o", "json"))
	})
	assert.Contains(t, out, "download_invoice")
}

func TestE2E_SubInvoiceRefund(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subscription-invoices/55/refund", r.URL.Path)
		jsonAPI(w, `{"data":{"type":"subscription-invoices","id":"55","attributes":{"refunded":true,"status":"refunded"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "subscription-invoices", "refund", "55", "--amount", "100", "-o", "json"))
	})
	assert.Contains(t, out, "refunded")
}

func TestE2E_SubInvoiceGenerateInvoice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subscription-invoices/55/generate-invoice", r.URL.Path)
		jsonAPI(w, `{"meta":{"urls":{"download_invoice":"https://x/i.pdf"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "subscription-invoices", "generate-invoice", "55", "--name", "Z", "-o", "json"))
	})
	assert.Contains(t, out, "download_invoice")
}

func TestE2E_LicenseActivate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/licenses/activate", r.URL.Path)
		body, _ := readBody(r)
		assert.Contains(t, body, "instance_name=my-box")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"activated":true,"instance":{"id":"i1"}}`))
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "license", "activate", "--key", "abc", "--instance-name", "my-box", "-o", "json"))
	})
	assert.Contains(t, out, "activated")
}

func TestE2E_LicenseDeactivate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/licenses/deactivate", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"deactivated":true}`))
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "license", "deactivate", "--key", "abc", "--instance-id", "i1", "-o", "json"))
	})
	assert.Contains(t, out, "deactivated")
}

func TestE2E_LicenseValidate_MissingKey(t *testing.T) {
	testEnv(t, "")
	err := run(t, "license", "validate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--key")
}

func TestE2E_DiscountsCreateDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			jsonAPI(w, `{"data":{"type":"discounts","id":"5","attributes":{"name":"SAVE","code":"SAVE10","amount":10,"amount_type":"percent","status":"published"}}}`)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "discounts", "create", "--data", `{"name":"SAVE","code":"SAVE10","amount":10,"amount_type":"percent"}`, "--rel", "store=stores:1", "-o", "json"))
	})
	assert.Contains(t, out, "SAVE10")
	out2 := captureStdout(t, func() { require.NoError(t, run(t, "discounts", "delete", "5")) })
	assert.Contains(t, out2, "deleted discount 5")
}

func TestE2E_CheckoutsCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/checkouts", r.URL.Path)
		jsonAPI(w, `{"data":{"type":"checkouts","id":"co_1","attributes":{"url":"https://store.lemonsqueezy.com/checkout/co_1"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "checkouts", "create", "--rel", "store=stores:1", "--rel", "variant=variants:2", "-o", "json"))
	})
	assert.Contains(t, out, "checkout/co_1")
}

func TestE2E_UsageRecordsCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/usage-records", r.URL.Path)
		jsonAPI(w, `{"data":{"type":"usage-records","id":"u1","attributes":{"quantity":5,"action":"increment"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "usage-records", "create", "--set", "quantity=5", "--rel", "subscription-item=subscription-items:1", "-o", "json"))
	})
	assert.Contains(t, out, "increment")
}

func TestE2E_LicenseKeysUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/license-keys/7", r.URL.Path)
		jsonAPI(w, `{"data":{"type":"license-keys","id":"7","attributes":{"activation_limit":10}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	require.NoError(t, run(t, "license-keys", "update", "7", "--set", "activation_limit=10", "--quiet"))
}

func TestE2E_WebhooksCreateUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonAPI(w, `{"data":{"type":"webhooks","id":"w1","attributes":{"url":"https://x/h","events":["order_created"]}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	require.NoError(t, run(t, "webhooks", "create", "--data", `{"url":"https://x/h","events":["order_created"]}`, "--rel", "store=stores:1", "--quiet"))
	require.NoError(t, run(t, "webhooks", "update", "w1", "--data", `{"events":["order_refunded"]}`, "--quiet"))
}

func TestE2E_WriteFlags_BadData(t *testing.T) {
	testEnv(t, "https://api.lemonsqueezy.com/v1")
	err := run(t, "customers", "create", "--data", "{bad json}", "--dry-run")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse --data JSON")

	err = run(t, "customers", "create", "--rel", "bogus", "--dry-run")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "want name=type:id")
}
