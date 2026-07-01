package commands_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- --jq global filter ---------------------------------------------------------------------

func TestE2E_JQ_ListScalars(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonAPI(w, `{"data":[{"type":"stores","id":"1","attributes":{"name":"A"}},{"type":"stores","id":"2","attributes":{"name":"B"}}]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "stores", "list", "-o", "json", "--jq", ".[].id"))
	})
	assert.Contains(t, out, `"1"`)
	assert.Contains(t, out, `"2"`)
	assert.NotContains(t, out, "name") // name filtered away by jq
}

func TestE2E_JQ_ReshapeToTable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonAPI(w, `{"data":[{"type":"orders","id":"9","attributes":{"user_email":"a@b.co","total_formatted":"$10","status":"paid"}}]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "orders", "list", "--no-color", "--jq", "[.[] | {id, email: .user_email}]"))
	})
	assert.Contains(t, out, "EMAIL")
	assert.Contains(t, out, "a@b.co")
	assert.NotContains(t, out, "STATUS") // dropped by the jq projection
}

func TestE2E_JQ_GetScalar(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonAPI(w, `{"data":{"type":"stores","id":"1","attributes":{"name":"My Store","total_sales":42}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "stores", "get", "1", "-o", "json", "--jq", ".total_sales"))
	})
	assert.Contains(t, out, "42")
	assert.NotContains(t, out, "My Store")
}

func TestE2E_JQ_InvalidExpression(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonAPI(w, `{"data":[{"type":"stores","id":"1","attributes":{"name":"A"}}]}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	err := run(t, "stores", "list", "--jq", ".[")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--jq")
}

// --- account selector (--account / hidden --profile alias / env) ----------------------------

func TestE2E_AccountFlag_AuthStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/me", r.URL.Path)
		jsonAPI(w, `{"data":{"type":"users","id":"1","attributes":{"name":"Jo","email":"jo@x.co"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "auth", "status", "--account", "staging"))
	})
	assert.Contains(t, out, "Profile:  staging")
	assert.Contains(t, out, "jo@x.co")
}

func TestE2E_ProfileAlias_StillWorks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonAPI(w, `{"data":{"type":"users","id":"1","attributes":{"name":"Jo","email":"jo@x.co"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	// The legacy --profile flag remains a working (hidden) alias for --account.
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "auth", "status", "--profile", "legacy"))
	})
	assert.Contains(t, out, "Profile:  legacy")
}

func TestE2E_AccountEnv_AuthStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonAPI(w, `{"data":{"type":"users","id":"1","attributes":{"name":"Jo","email":"jo@x.co"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	t.Setenv("LEMONSQUEEZY_ACCOUNT", "from-env")
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "auth", "status"))
	})
	assert.Contains(t, out, "Profile:  from-env")
}

// --- broaden mocked-API read coverage across the resource set -------------------------------

// TestE2E_ResourceListReads drives the generic list path (and each Client accessor + the
// JSON:API flatten) for the resources not otherwise exercised, against one mock server.
func TestE2E_ResourceListReads(t *testing.T) {
	cases := []struct {
		args []string
		path string
		body string
		want string
	}{
		{[]string{"variants", "list"}, "/variants", `{"data":[{"type":"variants","id":"1","attributes":{"name":"V","price":100,"status":"published"}}]}`, "V"},
		{[]string{"prices", "list"}, "/prices", `{"data":[{"type":"prices","id":"1","attributes":{"category":"one_time","scheme":"standard","unit_price":100}}]}`, "one_time"},
		{[]string{"files", "list"}, "/files", `{"data":[{"type":"files","id":"1","attributes":{"name":"f.zip","size_formatted":"1 MB","status":"published"}}]}`, "f.zip"},
		{[]string{"order-items", "list"}, "/order-items", `{"data":[{"type":"order-items","id":"1","attributes":{"order_id":9,"product_name":"P","variant_name":"V","quantity":2}}]}`, "P"},
		{[]string{"subscription-items", "list"}, "/subscription-items", `{"data":[{"type":"subscription-items","id":"1","attributes":{"subscription_id":9,"price_id":3,"quantity":1,"is_usage_based":true}}]}`, "true"},
		{[]string{"subscription-invoices", "list"}, "/subscription-invoices", `{"data":[{"type":"subscription-invoices","id":"1","attributes":{"user_email":"a@b.co","total_formatted":"$9","status":"paid"}}]}`, "a@b.co"},
		{[]string{"usage-records", "list"}, "/usage-records", `{"data":[{"type":"usage-records","id":"1","attributes":{"quantity":5,"action":"increment"}}]}`, "increment"},
		{[]string{"discounts", "list"}, "/discounts", `{"data":[{"type":"discounts","id":"1","attributes":{"name":"D","code":"D10","amount":10,"amount_type":"percent","status":"published"}}]}`, "D10"},
		{[]string{"discount-redemptions", "list"}, "/discount-redemptions", `{"data":[{"type":"discount-redemptions","id":"1","attributes":{"discount_id":2,"order_id":9,"discount_code":"D10","amount":10}}]}`, "D10"},
		{[]string{"license-keys", "list"}, "/license-keys", `{"data":[{"type":"license-keys","id":"1","attributes":{"key_short":"XXXX","user_email":"a@b.co","status":"active"}}]}`, "a@b.co"},
		{[]string{"license-key-instances", "list"}, "/license-key-instances", `{"data":[{"type":"license-key-instances","id":"1","attributes":{"license_key_id":2,"name":"box","identifier":"id1"}}]}`, "box"},
		{[]string{"checkouts", "list"}, "/checkouts", `{"data":[{"type":"checkouts","id":"co_1","attributes":{"url":"https://x/co_1"}}]}`, "co_1"},
		{[]string{"customers", "get", "3"}, "/customers/3", `{"data":{"type":"customers","id":"3","attributes":{"name":"Acme","email":"a@b.co","status":"subscribed"}}}`, "Acme"},
		{[]string{"subscriptions", "get", "9"}, "/subscriptions/9", `{"data":{"type":"subscriptions","id":"9","attributes":{"product_name":"Pro","status":"active"}}}`, "Pro"},
	}
	for _, tc := range cases {
		t.Run(tc.args[0]+"_"+tc.args[1], func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.path, r.URL.Path)
				jsonAPI(w, tc.body)
			}))
			defer srv.Close()
			testEnv(t, srv.URL)
			out := captureStdout(t, func() {
				require.NoError(t, run(t, append(tc.args, "-o", "json")...))
			})
			assert.Contains(t, out, tc.want)
		})
	}
}

// TestE2E_DoctorText exercises the non-JSON doctor renderer end-to-end.
func TestE2E_DoctorText(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonAPI(w, `{"data":{"type":"users","id":"1","attributes":{"email":"x@y.co"}}}`)
	}))
	defer srv.Close()
	testEnv(t, srv.URL)
	out := captureStdout(t, func() { _ = run(t, "doctor") })
	assert.Contains(t, out, "connectivity")
	assert.Contains(t, out, "x@y.co")
}
