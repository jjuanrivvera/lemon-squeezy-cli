package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccessorsHitCorrectPaths verifies every Client resource accessor targets the right
// collection path. A wrong path is a silent, hard-to-spot bug, so it's worth a cheap lock.
func TestAccessorsHitCorrectPaths(t *testing.T) {
	var gotPath string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", MediaType)
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	ctx := context.Background()
	cases := []struct {
		path string
		call func() error
	}{
		{"/stores", func() error { _, _, e := c.Stores().List(ctx, ListParams{}); return e }},
		{"/products", func() error { _, _, e := c.Products().List(ctx, ListParams{}); return e }},
		{"/variants", func() error { _, _, e := c.Variants().List(ctx, ListParams{}); return e }},
		{"/prices", func() error { _, _, e := c.Prices().List(ctx, ListParams{}); return e }},
		{"/files", func() error { _, _, e := c.Files().List(ctx, ListParams{}); return e }},
		{"/customers", func() error { _, _, e := c.Customers().List(ctx, ListParams{}); return e }},
		{"/orders", func() error { _, _, e := c.Orders().List(ctx, ListParams{}); return e }},
		{"/order-items", func() error { _, _, e := c.OrderItems().List(ctx, ListParams{}); return e }},
		{"/subscriptions", func() error { _, _, e := c.Subscriptions().List(ctx, ListParams{}); return e }},
		{"/subscription-items", func() error { _, _, e := c.SubscriptionItems().List(ctx, ListParams{}); return e }},
		{"/subscription-invoices", func() error { _, _, e := c.SubscriptionInvoices().List(ctx, ListParams{}); return e }},
		{"/usage-records", func() error { _, _, e := c.UsageRecords().List(ctx, ListParams{}); return e }},
		{"/discounts", func() error { _, _, e := c.Discounts().List(ctx, ListParams{}); return e }},
		{"/discount-redemptions", func() error { _, _, e := c.DiscountRedemptions().List(ctx, ListParams{}); return e }},
		{"/license-keys", func() error { _, _, e := c.LicenseKeys().List(ctx, ListParams{}); return e }},
		{"/license-key-instances", func() error { _, _, e := c.LicenseKeyInstances().List(ctx, ListParams{}); return e }},
		{"/checkouts", func() error { _, _, e := c.Checkouts().List(ctx, ListParams{}); return e }},
		{"/webhooks", func() error { _, _, e := c.Webhooks().List(ctx, ListParams{}); return e }},
	}
	for _, tc := range cases {
		require.NoError(t, tc.call(), tc.path)
		assert.Equal(t, tc.path, gotPath)
	}
}

func TestMe(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/me", r.URL.Path)
		w.Header().Set("Content-Type", MediaType)
		_, _ = w.Write([]byte(`{"data":{"type":"users","id":"1","attributes":{"email":"a@b.co"}}}`))
	})
	u, err := c.Me(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "a@b.co", u.Email)
}

func TestResourceTypeOverride(t *testing.T) {
	c := New("https://x", "k")
	r := c.Products().WithType("custom-type")
	assert.Equal(t, "custom-type", r.Type())
}

func TestScalarHelpers(t *testing.T) {
	assert.Equal(t, int64(42), Int(42).Int64())
	assert.Equal(t, "9", ID("9").String())
	assert.Equal(t, "1999", Money("1999").String())
}
