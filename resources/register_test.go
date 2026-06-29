package resources_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/lemon-squeezy-cli/commands"
	_ "github.com/jjuanrivvera/lemon-squeezy-cli/resources"
)

// allResources is the full priority surface derived from the Lemon Squeezy API (the SDK's
// resource modules). The registration test locks the surface so a dropped resource fails CI.
var allResources = []string{
	"stores", "products", "variants", "prices", "files",
	"customers", "orders", "order-items",
	"subscriptions", "subscription-items", "subscription-invoices", "usage-records",
	"discounts", "discount-redemptions",
	"license-keys", "license-key-instances",
	"checkouts", "webhooks",
	"users", "license",
}

func TestAllResourcesRegistered(t *testing.T) {
	root := commands.Setup()
	for _, name := range allResources {
		cmd, _, err := root.Find([]string{name})
		require.NoError(t, err, "resource %q should be reachable", name)
		assert.Equal(t, name, cmd.Name(), "resource %q", name)
	}
}

func TestReadOnlyResourcesHaveNoWriteVerbs(t *testing.T) {
	root := commands.Setup()
	readOnly := []string{"stores", "products", "variants", "prices", "files", "order-items", "discount-redemptions", "license-key-instances"}
	for _, name := range readOnly {
		parent, _, err := root.Find([]string{name})
		require.NoError(t, err)
		for _, sub := range parent.Commands() {
			switch sub.Name() {
			case "create", "update", "delete":
				t.Errorf("read-only resource %q must not expose %q", name, sub.Name())
			}
		}
	}
}

func TestWritableResourcesExposeExpectedVerbs(t *testing.T) {
	root := commands.Setup()
	cases := map[string][]string{
		"webhooks":      {"create", "update", "delete"},
		"discounts":     {"create", "delete"},
		"customers":     {"create", "update", "archive"},
		"subscriptions": {"update", "cancel"},
		"orders":        {"refund", "generate-invoice"},
	}
	for name, verbs := range cases {
		parent, _, err := root.Find([]string{name})
		require.NoError(t, err, name)
		have := map[string]bool{}
		for _, sub := range parent.Commands() {
			have[sub.Name()] = true
		}
		for _, v := range verbs {
			assert.True(t, have[v], "resource %q should expose verb %q", name, v)
		}
	}
}
