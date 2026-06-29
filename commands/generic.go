package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/api"
)

// MCP annotation keys (ophis-compatible hints). Set once here so every resource command is
// classified read-only/write/destructive without per-command edits — this drives both the
// MCP tool surface and the agent guard.
const (
	annReadOnly    = "mcp.readOnly"
	annWrite       = "mcp.write"
	annDestructive = "mcp.destructive"
)

// ListFilter maps a CLI flag to a JSON:API filter[key] query param for list commands.
type ListFilter struct {
	Flag  string
	Query string // the filter key, e.g. "store_id" -> filter[store_id]
	Usage string
}

// ExtraCommand is a custom verb (e.g. orders refund) contributed by a resource. Exactly one
// of ReadOnly/Write/Destructive should be set; an unset classification defaults to
// destructive (the safe default for an unannotated mutating verb — §3b).
type ExtraCommand struct {
	Build       func(getClient func(bool) (*api.Client, renderFunc, error)) *cobra.Command
	ReadOnly    bool // list/get/usage — safe to expose freely
	Write       bool // reversible mutation (archive) — needs approval under the agent guard
	Destructive bool // irreversible (refund, cancel, deactivate) — hard-blocked by the guard
}

// renderFunc is the bound render closure handed to resources so they need not import config.
type renderFunc = func(v any, defaultColumns []string) error

// ResourceSpec declares a resource's CLI surface. A new resource = a struct + a Client
// accessor + one RegisterResource call. No shared code changes. Lemon Squeezy is mostly
// read-only, so list/get are always present and the three write verbs are opt-in.
type ResourceSpec[T any] struct {
	Use         string
	Aliases     []string
	Short       string
	New         func(*api.Client) *api.Resource[T]
	Columns     []string
	OrderFields []string
	ListFilters []ListFilter
	Includes    []string // valid JSON:API include values (shown in --include help)

	// Write verbs are opt-in (most Lemon Squeezy resources are read-only). Create/Update use
	// the universal JSON:API write flags (--data/--set/--rel) so adding a writable resource
	// needs zero per-resource flag code.
	CanCreate bool
	CanUpdate bool
	CanDelete bool

	// Extra adds custom verbs (refund, cancel, usage, archive, …).
	Extra []ExtraCommand
}

// registrar is a deferred registration captured at init() and applied to rootCmd in Setup.
type registrar func() *cobra.Command

var registrars []registrar

// RegisterResource is called from a resource package's init(). It is generic over T, so it
// cannot be a method; the closure defers building until the command tree is assembled.
func RegisterResource[T any](spec ResourceSpec[T]) {
	registrars = append(registrars, func() *cobra.Command { return buildResourceCmd(spec) })
}

// RegisterCommand registers a non-generic top-level command group (e.g. the `users` singleton
// or the License API `license` group) built lazily at Setup time, alongside generic resources.
func RegisterCommand(build func() *cobra.Command) {
	registrars = append(registrars, build)
}

// ClientRender exposes the bound client+render factory to resource packages that build custom
// command groups (RegisterCommand) without importing the config/output internals directly.
func ClientRender(requireAuth bool) (*api.Client, renderFunc, error) {
	return clientRenderFactory(requireAuth)
}

// Quiet reports whether --quiet was set, for custom command groups.
func Quiet() bool { return gf.quiet }

// clientRenderFactory adapts getAPIClient + render into the simpler signature resources use.
func clientRenderFactory(requireAuth bool) (*api.Client, renderFunc, error) {
	c, cfg, err := getAPIClient(requireAuth)
	if err != nil {
		return nil, nil, err
	}
	r := func(v any, cols []string) error { return render(cfg, v, cols) }
	return c, r, nil
}

func buildResourceCmd[T any](spec ResourceSpec[T]) *cobra.Command {
	parent := &cobra.Command{
		Use:     spec.Use,
		Aliases: spec.Aliases,
		Short:   spec.Short,
	}

	parent.AddCommand(buildListCmd(spec))
	parent.AddCommand(buildGetCmd(spec))
	if spec.CanCreate {
		parent.AddCommand(buildCreateCmd(spec))
	}
	if spec.CanUpdate {
		parent.AddCommand(buildUpdateCmd(spec))
	}
	if spec.CanDelete {
		parent.AddCommand(buildDeleteCmd(spec))
	}
	for _, ex := range spec.Extra {
		cmd := ex.Build(clientRenderFactory)
		switch {
		case ex.ReadOnly:
			annotate(cmd, annReadOnly)
		case ex.Write:
			annotate(cmd, annWrite)
		default:
			annotate(cmd, annDestructive) // Destructive, or unannotated (safe default)
		}
		parent.AddCommand(cmd)
	}
	return parent
}

func buildListCmd[T any](spec ResourceSpec[T]) *cobra.Command {
	var filterVals map[string]*string
	var include []string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List " + spec.Use,
		Example: fmt.Sprintf("  lsqueezy %s list --limit 25\n  lsqueezy %s list -o json --all",
			spec.Use, spec.Use),
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, r, err := clientRenderFactory(true)
			if err != nil {
				return err
			}
			res := spec.New(c)
			p := listParamsFromFlags(filterVals, include)
			ctx := cmd.Context()
			var items []T
			if gf.all {
				items, err = res.ListAll(ctx, p)
			} else {
				items, _, err = res.List(ctx, p)
			}
			if err != nil {
				return err
			}
			if c.DryRun {
				return nil
			}
			filtered, err := applyClientFilters(items, gf.filter)
			if err != nil {
				return err
			}
			return r(filtered, spec.Columns)
		},
	}
	filterVals = map[string]*string{}
	for _, f := range spec.ListFilters {
		v := new(string)
		cmd.Flags().StringVar(v, f.Flag, "", "filter: "+f.Usage)
		filterVals[f.Query] = v
	}
	if len(spec.Includes) > 0 {
		cmd.Flags().StringSliceVar(&include, "include", nil,
			"embed related resources: "+strings.Join(spec.Includes, ","))
	}
	annotate(cmd, annReadOnly)
	return cmd
}

func listParamsFromFlags(filters map[string]*string, include []string) api.ListParams {
	p := api.ListParams{
		PageSize:   gf.limit,
		PageNumber: gf.page,
		Sort:       gf.sort,
		Include:    include,
		Filters:    map[string]string{},
	}
	for q, v := range filters {
		if v != nil && *v != "" {
			p.Filters[q] = *v
		}
	}
	return p
}

// applyClientFilters keeps only items whose JSON fields match every field=value pair in
// filters. Filtering happens client-side (post-fetch) so it works uniformly for any
// resource without per-resource code. An empty filter list passes everything through.
func applyClientFilters[T any](items []T, filters []string) ([]T, error) {
	if len(filters) == 0 {
		return items, nil
	}
	want := map[string]string{}
	for _, f := range filters {
		k, v, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --filter %q (want field=value)", f)
		}
		want[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	var out []T
	for _, it := range items {
		b, err := json.Marshal(it)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		match := true
		for k, v := range want {
			if fmt.Sprintf("%v", m[k]) != v {
				match = false
				break
			}
		}
		if match {
			out = append(out, it)
		}
	}
	return out, nil
}

func buildGetCmd[T any](spec ResourceSpec[T]) *cobra.Command {
	var include []string
	cmd := &cobra.Command{
		Use:     "get <id>",
		Short:   "Get a single " + singular(spec.Use) + " by id",
		Example: fmt.Sprintf("  lsqueezy %s get 42 -o yaml", spec.Use),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, r, err := clientRenderFactory(true)
			if err != nil {
				return err
			}
			item, err := spec.New(c).Get(cmd.Context(), args[0], include...)
			if err != nil {
				return err
			}
			if c.DryRun {
				return nil
			}
			return r(item, spec.Columns)
		},
	}
	if len(spec.Includes) > 0 {
		cmd.Flags().StringSliceVar(&include, "include", nil,
			"embed related resources: "+strings.Join(spec.Includes, ","))
	}
	annotate(cmd, annReadOnly)
	return cmd
}

func buildCreateCmd[T any](spec ResourceSpec[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a " + singular(spec.Use),
		Long: "Create a " + singular(spec.Use) + ".\n\n" +
			"Supply attributes with --data (raw JSON object or @file) and/or repeated --set\n" +
			"key=value, and relationships with repeated --rel name=type:id. The JSON:API\n" +
			"envelope (type/attributes/relationships) is added for you.",
		Example: fmt.Sprintf(
			"  lsqueezy %s create --data '{\"name\":\"Acme\"}' --rel store=stores:1\n"+
				"  lsqueezy %s create --set name=Acme --set email=a@b.co --rel store=stores:1 --dry-run",
			spec.Use, spec.Use),
	}
	build := writeFlags(cmd)
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		c, r, err := clientRenderFactory(true)
		if err != nil {
			return err
		}
		body, err := build()
		if err != nil {
			return err
		}
		var out T
		if err := spec.New(c).Create(cmd.Context(), body, &out); err != nil {
			return err
		}
		if c.DryRun || gf.quiet {
			return nil
		}
		return r(&out, spec.Columns)
	}
	annotate(cmd, annWrite)
	return cmd
}

func buildUpdateCmd[T any](spec ResourceSpec[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a " + singular(spec.Use) + " by id",
		Example: fmt.Sprintf(
			"  lsqueezy %s update 42 --set name=NewName\n  lsqueezy %s update 42 --data '{\"status\":\"archived\"}' --dry-run",
			spec.Use, spec.Use),
		Args: cobra.ExactArgs(1),
	}
	build := writeFlags(cmd)
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		c, r, err := clientRenderFactory(true)
		if err != nil {
			return err
		}
		body, err := build()
		if err != nil {
			return err
		}
		var out T
		if err := spec.New(c).Update(cmd.Context(), args[0], body, &out); err != nil {
			return err
		}
		if c.DryRun || gf.quiet {
			return nil
		}
		return r(&out, spec.Columns)
	}
	annotate(cmd, annWrite)
	return cmd
}

func buildDeleteCmd[T any](spec ResourceSpec[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <id>",
		Short:   "Delete a " + singular(spec.Use) + " by id",
		Example: fmt.Sprintf("  lsqueezy %s delete 42\n  lsqueezy %s delete 42 --dry-run", spec.Use, spec.Use),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := clientRenderFactory(true)
			if err != nil {
				return err
			}
			if err := spec.New(c).Delete(cmd.Context(), args[0]); err != nil {
				return err
			}
			if c.DryRun {
				return nil
			}
			if !gf.quiet {
				fmt.Printf("deleted %s %s\n", singular(spec.Use), args[0])
			}
			return nil
		},
	}
	annotate(cmd, annDestructive)
	return cmd
}

// annotate stamps an MCP classification annotation on a command (set once, in the builder).
func annotate(cmd *cobra.Command, kind string) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[kind] = "true"
}

// singular is a crude depluralizer good enough for our resource names.
func singular(s string) string {
	switch {
	case strings.HasSuffix(s, "ies"):
		return s[:len(s)-3] + "y"
	case len(s) > 1 && s[len(s)-1] == 's':
		return s[:len(s)-1]
	default:
		return s
	}
}
