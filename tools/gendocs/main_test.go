package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteIndex(t *testing.T) {
	root := &cobra.Command{Use: "lsqueezy"}
	root.AddCommand(&cobra.Command{Use: "stores", Short: "Browse stores"})
	root.AddCommand(&cobra.Command{Use: "orders", Short: "Manage orders"})
	hidden := &cobra.Command{Use: "secret", Short: "hidden", Hidden: true}
	root.AddCommand(hidden)

	dir := t.TempDir()
	require.NoError(t, writeIndex(root, dir))

	b, err := os.ReadFile(filepath.Join(dir, "index.md"))
	require.NoError(t, err)
	out := string(b)

	assert.Contains(t, out, "# Command Reference")
	assert.Contains(t, out, "[`orders`](lsqueezy_orders.md)")
	assert.Contains(t, out, "[`stores`](lsqueezy_stores.md)")
	// stable, alphabetical: orders before stores
	assert.Less(t, indexOf(out, "orders"), indexOf(out, "stores"))
	// hidden commands are omitted
	assert.NotContains(t, out, "secret")
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
