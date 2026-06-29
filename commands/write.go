package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/api"
)

// writeFlags registers the universal JSON:API write flags on a create/update command and
// returns a builder that assembles the WriteBody from them. Keeping these generic means a
// writable resource needs no per-resource flag code — the user supplies the exact documented
// attributes, so we never fabricate or hardcode field names.
//
//	--data/-d   attributes as a JSON object (or @file to read it from a file)
//	--set       attribute as key=value, repeatable (value is JSON-parsed, else a string)
//	--rel       relationship as name=type:id, repeatable (e.g. store=stores:1)
func writeFlags(cmd *cobra.Command) func() (api.WriteBody, error) {
	var dataStr string
	var sets, rels []string
	cmd.Flags().StringVarP(&dataStr, "data", "d", "", "attributes as a JSON object, or @file")
	cmd.Flags().StringArrayVar(&sets, "set", nil, "attribute key=value (repeatable)")
	cmd.Flags().StringArrayVar(&rels, "rel", nil, "relationship name=type:id (repeatable), e.g. store=stores:1")

	return func() (api.WriteBody, error) {
		attrs := map[string]any{}
		if dataStr != "" {
			raw := dataStr
			if strings.HasPrefix(dataStr, "@") {
				// User-named file, like `curl -d @file`; this path comes from the invocation,
				// not from API data, so it is not subject to the data-path-confinement rule.
				b, err := os.ReadFile(dataStr[1:]) // #nosec G304 -- explicit user-supplied data file
				if err != nil {
					return api.WriteBody{}, fmt.Errorf("read --data file: %w", err)
				}
				raw = string(b)
			}
			if err := json.Unmarshal([]byte(raw), &attrs); err != nil {
				return api.WriteBody{}, fmt.Errorf("parse --data JSON: %w", err)
			}
		}
		for _, s := range sets {
			k, v, ok := strings.Cut(s, "=")
			if !ok {
				return api.WriteBody{}, fmt.Errorf("invalid --set %q (want key=value)", s)
			}
			attrs[strings.TrimSpace(k)] = coerceValue(v)
		}
		rel := map[string]any{}
		for _, rdef := range rels {
			name, spec, ok := strings.Cut(rdef, "=")
			if !ok {
				return api.WriteBody{}, fmt.Errorf("invalid --rel %q (want name=type:id)", rdef)
			}
			typ, id, ok := strings.Cut(spec, ":")
			if !ok || typ == "" || id == "" {
				return api.WriteBody{}, fmt.Errorf("invalid --rel %q (want name=type:id)", rdef)
			}
			rel[strings.TrimSpace(name)] = api.Relationship(typ, id)
		}

		wb := api.WriteBody{}
		if len(attrs) > 0 {
			wb.Attributes = attrs
		}
		if len(rel) > 0 {
			wb.Relationships = rel
		}
		return wb, nil
	}
}

// coerceValue parses a --set value as JSON so numbers/bools/null/arrays/objects round-trip
// with their real type, falling back to the raw string when it isn't valid JSON.
func coerceValue(v string) any {
	var out any
	if json.Unmarshal([]byte(v), &out) == nil {
		return out
	}
	return v
}
