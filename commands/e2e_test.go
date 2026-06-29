package commands_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/lemon-squeezy-cli/commands"

	// Blank-import the real resources so init() registers them, exactly as main.go does.
	_ "github.com/jjuanrivvera/lemon-squeezy-cli/resources"
)

func init() { commands.Setup() }

func testEnv(t *testing.T, baseURL string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("LEMONSQUEEZY_API_KEY", "test-key")
	if baseURL != "" {
		t.Setenv("LEMONSQUEEZY_BASE_URL", baseURL)
	}
}

// run executes the root command afresh, resetting package-global flag state between runs.
func run(t *testing.T, args ...string) error {
	t.Helper()
	root := commands.Root()
	resetFlags(root)
	root.SetArgs(args)
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	return root.ExecuteContext(context.Background())
}

func resetFlags(root *cobra.Command) {
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				_ = sv.Replace(nil)
			} else {
				_ = f.Value.Set(f.DefValue)
			}
			f.Changed = false
		})
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// jsonAPI writes a JSON:API body with the right content type.
func jsonAPI(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	_, _ = w.Write([]byte(body))
}

func TestE2E_StoresList_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/stores", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/vnd.api+json", r.Header.Get("Accept"))
		jsonAPI(w, `{"meta":{"page":{"lastPage":1}},"data":[{"type":"stores","id":"1","attributes":{"name":"My Store","slug":"my-store","currency":"USD","total_sales":42}}]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "stores", "list", "-o", "json")) })
	assert.Contains(t, out, "My Store")
	assert.Contains(t, out, `"id": "1"`)
}

func TestE2E_StoresList_Table(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonAPI(w, `{"data":[{"type":"stores","id":"1","attributes":{"name":"My Store","slug":"my-store","currency":"USD","total_sales":42}}]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "stores", "list", "--no-color")) })
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "My Store")
}

func TestE2E_StoresList_CSV(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonAPI(w, `{"data":[{"type":"stores","id":"1","attributes":{"name":"My Store","slug":"s","currency":"USD","total_sales":1}}]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "stores", "list", "-o", "csv")) })
	assert.Contains(t, out, "id,name")
}

func TestE2E_StoresList_YAML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonAPI(w, `{"data":[{"type":"stores","id":"1","attributes":{"name":"My Store"}}]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "stores", "list", "-o", "yaml")) })
	assert.Contains(t, out, "name: My Store")
}

func TestE2E_ProductsGet_WithColumns(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/products/55", r.URL.Path)
		jsonAPI(w, `{"data":{"type":"products","id":"55","attributes":{"name":"Pro Plan","status":"published","price":1999}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "products", "get", "55", "-o", "json", "--columns", "id,name")) })
	assert.Contains(t, out, "Pro Plan")
}

func TestE2E_ProductsList_AllPagination(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		page := r.URL.Query().Get("page[number]")
		jsonAPI(w, `{"meta":{"page":{"currentPage":`+page+`,"lastPage":2}},"data":[{"type":"products","id":"`+page+`","attributes":{"name":"p"}}]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "products", "list", "--all", "-o", "json")) })
	assert.Equal(t, 2, calls)
	assert.Contains(t, out, `"id": "1"`)
	assert.Contains(t, out, `"id": "2"`)
}

func TestE2E_ProductsList_Filter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "9", r.URL.Query().Get("filter[store_id]"))
		jsonAPI(w, `{"data":[{"type":"products","id":"1","attributes":{"name":"p","status":"published"}}]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	// client-side --filter also applies on top of the server filter.
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "products", "list", "--store-id", "9", "--filter", "status=published", "-o", "json"))
	})
	assert.Contains(t, out, `"status": "published"`)
}

func TestE2E_CustomersCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := readBody(r)
		assert.Contains(t, body, `"type":"customers"`)
		assert.Contains(t, body, `"name":"Acme"`)
		assert.Contains(t, body, `"store"`)
		jsonAPI(w, `{"data":{"type":"customers","id":"100","attributes":{"name":"Acme","email":"a@b.co","status":"subscribed"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "customers", "create", "--set", "name=Acme", "--set", "email=a@b.co", "--rel", "store=stores:1", "-o", "json"))
	})
	assert.Contains(t, out, `"id": "100"`)
}

func TestE2E_CustomersCreate_DryRun(t *testing.T) {
	testEnv(t, "https://api.lemonsqueezy.com/v1")
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "customers", "create", "--data", `{"name":"Z"}`, "--rel", "store=stores:1", "--dry-run"))
	})
	assert.Contains(t, out, "curl -X POST")
	assert.Contains(t, out, "REDACTED")
	assert.Contains(t, out, "vnd.api+json")
}

func TestE2E_CustomersArchive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/customers/7", r.URL.Path)
		body, _ := readBody(r)
		assert.Contains(t, body, `"status":"archived"`)
		jsonAPI(w, `{"data":{"type":"customers","id":"7","attributes":{"name":"Z","status":"archived"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "customers", "archive", "7", "-o", "json")) })
	assert.Contains(t, out, "archived")
}

func TestE2E_SubscriptionsUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		body, _ := readBody(r)
		assert.Contains(t, body, `"id":"9"`)
		jsonAPI(w, `{"data":{"type":"subscriptions","id":"9","attributes":{"status":"active"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	require.NoError(t, run(t, "subscriptions", "update", "9", "--set", "pause=null", "--quiet"))
}

func TestE2E_SubscriptionsCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/subscriptions/9", r.URL.Path)
		jsonAPI(w, `{"data":{"type":"subscriptions","id":"9","attributes":{"status":"cancelled"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "subscriptions", "cancel", "9", "-o", "json")) })
	assert.Contains(t, out, "cancelled")
}

func TestE2E_OrdersRefund(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/orders/12/refund", r.URL.Path)
		body, _ := readBody(r)
		assert.Contains(t, body, `"amount":500`)
		jsonAPI(w, `{"data":{"type":"orders","id":"12","attributes":{"status":"refunded","refunded":true}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "orders", "refund", "12", "--amount", "500", "-o", "json")) })
	assert.Contains(t, out, "refunded")
}

func TestE2E_SubscriptionItemUsage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subscription-items/4/current-usage", r.URL.Path)
		jsonAPI(w, `{"meta":{"period_start":"2026-01-01","quantity":10}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "subscription-items", "current-usage", "4", "-o", "json")) })
	assert.Contains(t, out, "quantity")
}

func TestE2E_WebhooksDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "webhooks", "delete", "3")) })
	assert.Contains(t, out, "deleted webhook 3")
}

func TestE2E_LicenseValidate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/licenses/validate", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		body, _ := readBody(r)
		assert.Contains(t, body, "license_key=abc")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":true,"license_key":{"status":"active"}}`))
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "license", "validate", "--key", "abc", "-o", "json")) })
	assert.Contains(t, out, "valid")
}

func TestE2E_UsersMe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/me", r.URL.Path)
		jsonAPI(w, `{"data":{"type":"users","id":"1","attributes":{"name":"Jo","email":"jo@x.co"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { require.NoError(t, run(t, "users", "me", "-o", "json")) })
	assert.Contains(t, out, "jo@x.co")
}

func TestE2E_APIRaw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/stores", r.URL.Path)
		jsonAPI(w, `{"data":[]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	require.NoError(t, run(t, "api", "GET", "/stores"))
}

func TestE2E_AuthStatus_NoKey(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // no key set
	out := captureStdout(t, func() { require.NoError(t, run(t, "auth", "status")) })
	assert.Contains(t, out, "no key stored")
}

func TestE2E_Version(t *testing.T) {
	out := captureStdout(t, func() { require.NoError(t, run(t, "version")) })
	assert.Contains(t, strings.ToLower(out), "lsqueezy")
}

func TestE2E_ConfigSetViewUse(t *testing.T) {
	testEnv(t, "")
	require.NoError(t, run(t, "config", "set", "output", "json"))
	require.NoError(t, run(t, "config", "set", "base_url", "https://api.lemonsqueezy.com/v1"))
	out := captureStdout(t, func() { require.NoError(t, run(t, "config", "view")) })
	assert.Contains(t, out, "base_url")
	require.NoError(t, run(t, "config", "path"))
	require.NoError(t, run(t, "config", "list-profiles"))
}

func TestE2E_Alias(t *testing.T) {
	testEnv(t, "")
	require.NoError(t, run(t, "alias", "set", "ords", "orders list"))
	out := captureStdout(t, func() { require.NoError(t, run(t, "alias", "list")) })
	assert.Contains(t, out, "ords")
	require.NoError(t, run(t, "alias", "remove", "ords"))
}

func TestE2E_DoctorJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonAPI(w, `{"data":{"type":"users","id":"1","attributes":{"email":"x@y.co"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { _ = run(t, "doctor", "--json") })
	assert.Contains(t, out, "connectivity")
}

func TestE2E_APIError_Hint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		jsonAPI(w, `{"errors":[{"status":"404","detail":"Order not found"}]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	err := run(t, "orders", "get", "999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Order not found")
	assert.Contains(t, err.Error(), "verify the id")
}

func TestE2E_CompletionGenerates(t *testing.T) {
	out := captureStdout(t, func() { require.NoError(t, run(t, "completion", "bash")) })
	assert.Contains(t, out, "lsqueezy")
}

func readBody(r *http.Request) (string, error) {
	var b bytes.Buffer
	_, err := b.ReadFrom(r.Body)
	return b.String(), err
}
