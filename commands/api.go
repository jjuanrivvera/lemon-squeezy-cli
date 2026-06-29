package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	var body string
	var queries []string
	cmd := &cobra.Command{
		Use:   "api <METHOD> <PATH>",
		Short: "Make a raw authenticated request (escape hatch)",
		Long: `Make a raw authenticated request against the API.

Honors --dry-run (prints the equivalent curl) and --output. Use it for endpoints the
typed resource commands don't cover yet.`,
		Example: "  lsqueezy api GET /products -q 'page[size]=1'\n  lsqueezy api POST /checkouts -d @checkout.json",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := getAPIClient(true)
			if err != nil {
				return err
			}
			method := strings.ToUpper(args[0])
			path := args[1]

			q := url.Values{}
			for _, kv := range queries {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("invalid -q %q (want key=value)", kv)
				}
				q.Add(k, v)
			}

			var reader io.Reader
			if body != "" {
				reader = bytes.NewReader([]byte(body))
			}
			resp, err := c.Do(cmd.Context(), method, path, q, reader)
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
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
			}
			// Pretty-print JSON; pass through anything else verbatim.
			var pretty bytes.Buffer
			if json.Indent(&pretty, data, "", "  ") == nil {
				_, _ = fmt.Fprintln(os.Stdout, pretty.String())
			} else {
				_, _ = fmt.Fprintln(os.Stdout, string(data))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&body, "data", "d", "", "request body (JSON)")
	cmd.Flags().StringArrayVarP(&queries, "query", "q", nil, "query param key=value (repeatable)")
	rootCmd.AddCommand(cmd)
}
