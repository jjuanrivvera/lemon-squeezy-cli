// Package output is the single renderer for every resource: table (default), json, yaml,
// csv. It works by normalizing any value to JSON first, so one code path serves all types
// and the column/filter/sort logic is written once.
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/itchyny/gojq"
	"gopkg.in/yaml.v3"
)

// Format is an output format.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatCSV   Format = "csv"
)

// ParseFormat validates a -o/--output value.
func ParseFormat(s string) (Format, error) {
	switch Format(strings.ToLower(s)) {
	case FormatTable:
		return FormatTable, nil
	case FormatJSON:
		return FormatJSON, nil
	case FormatYAML:
		return FormatYAML, nil
	case FormatCSV:
		return FormatCSV, nil
	default:
		return "", fmt.Errorf("unknown output format %q (want table|json|yaml|csv)", s)
	}
}

// Options control rendering.
type Options struct {
	Format  Format
	Columns []string // explicit column subset/order; empty = default for the format
	NoColor bool
	JQ      string // optional gojq expression applied to the value before rendering
	Writer  io.Writer
}

// shouldColor reports whether ANSI color is appropriate: only on a TTY, never when
// NoColor or the NO_COLOR env var is set.
func (o Options) shouldColor() bool {
	if o.NoColor || os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := o.Writer.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// Render writes v in the requested format. v is typically a slice of structs or a single
// struct. defaultColumns is the resource's canonical column order (used by table/csv when
// Columns is empty).
func Render(v any, defaultColumns []string, o Options) error {
	if o.Writer == nil {
		o.Writer = os.Stdout
	}
	// --jq runs BEFORE format selection so json/yaml/table/csv all render the filtered value.
	// It's the last-mile transform an operator reaches for (e.g. `-o json --jq '.[].id'`).
	if o.JQ != "" {
		filtered, err := applyJQ(o.JQ, v)
		if err != nil {
			return err
		}
		v = filtered
		// The resource's curated default columns describe its native shape; a jq expression
		// reshapes the data, so table/csv must derive columns from the result itself (unless
		// the user pinned --columns). Otherwise a projected {id,email} would still print the
		// original id/status/... headers.
		defaultColumns = nil
	}
	switch o.Format {
	case FormatJSON, "":
		return renderJSON(v, o.Writer)
	case FormatYAML:
		return renderYAML(v, o.Writer)
	case FormatTable:
		return renderTabular(v, defaultColumns, o, false)
	case FormatCSV:
		return renderTabular(v, defaultColumns, o, true)
	default:
		return fmt.Errorf("unsupported format %q", o.Format)
	}
}

// applyJQ runs a gojq program over v and returns the combined result (the single output when
// the program yields one, a slice when it yields many, nil when none). v is normalized to
// plain JSON types first — gojq operates on map/slice/float64/string/bool/nil, not Go structs
// or the resource's named types.
func applyJQ(program string, v any) (any, error) {
	q, err := gojq.Parse(program)
	if err != nil {
		return nil, fmt.Errorf("invalid --jq expression %q: %w", program, err)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var norm any
	if err := json.Unmarshal(b, &norm); err != nil {
		return nil, err
	}
	iter := q.Run(norm)
	var results []any
	for {
		out, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := out.(error); ok {
			return nil, fmt.Errorf("--jq: %w", err)
		}
		results = append(results, out)
	}
	switch len(results) {
	case 0:
		return nil, nil
	case 1:
		return results[0], nil
	default:
		return results, nil
	}
}

func renderJSON(v any, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func renderYAML(v any, w io.Writer) error {
	// Round-trip through JSON so json tags drive YAML keys (yaml.v3 honors yaml tags only).
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var generic any
	if err := json.Unmarshal(b, &generic); err != nil {
		return err
	}
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	return enc.Encode(generic)
}

// toRows normalizes any value to an ordered list of string->string maps plus the union of
// keys. Normalizing via JSON means the formatter never reflects over Go structs directly.
func toRows(v any) ([]map[string]string, []string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, nil, err
	}
	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, err
	}

	var records []map[string]any
	switch t := raw.(type) {
	case []any:
		for _, item := range t {
			if m, ok := item.(map[string]any); ok {
				records = append(records, m)
			} else {
				records = append(records, map[string]any{"value": item})
			}
		}
	case map[string]any:
		records = append(records, t)
	default:
		records = append(records, map[string]any{"value": t})
	}

	rows := make([]map[string]string, 0, len(records))
	keySet := map[string]bool{}
	var keyOrder []string
	for _, rec := range records {
		row := make(map[string]string, len(rec))
		for k, val := range rec {
			if !keySet[k] {
				keySet[k] = true
				keyOrder = append(keyOrder, k)
			}
			row[k] = stringify(val)
		}
		rows = append(rows, row)
	}
	sort.Strings(keyOrder) // deterministic union order when no explicit columns
	return rows, keyOrder, nil
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		// JSON numbers decode as float64; render integers without a trailing .0.
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case bool:
		return fmt.Sprintf("%t", t)
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func renderTabular(v any, defaultColumns []string, o Options, asCSV bool) error {
	rows, unionKeys, err := toRows(v)
	if err != nil {
		return err
	}
	cols := o.Columns
	if len(cols) == 0 {
		if len(defaultColumns) > 0 {
			cols = defaultColumns
		} else {
			cols = unionKeys
		}
	}

	if asCSV {
		return writeCSV(rows, cols, o.Writer)
	}
	return writeTable(rows, cols, o)
}

func writeCSV(rows []map[string]string, cols []string, w io.Writer) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(cols); err != nil {
		return err
	}
	for _, row := range rows {
		rec := make([]string, len(cols))
		for i, c := range cols {
			rec[i] = row[c]
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

const (
	colReset = "\033[0m"
	colBold  = "\033[1m"
)

func writeTable(rows []map[string]string, cols []string, o Options) error {
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = len(c)
	}
	for _, row := range rows {
		for i, c := range cols {
			if l := len(row[c]); l > widths[i] {
				widths[i] = l
			}
		}
	}

	color := o.shouldColor()
	var hb strings.Builder
	for i, c := range cols {
		if color {
			hb.WriteString(colBold)
		}
		hb.WriteString(pad(strings.ToUpper(c), widths[i]))
		if color {
			hb.WriteString(colReset)
		}
		if i < len(cols)-1 {
			hb.WriteString("  ")
		}
	}
	_, _ = fmt.Fprintln(o.Writer, strings.TrimRight(hb.String(), " "))

	for _, row := range rows {
		var rb strings.Builder
		for i, c := range cols {
			rb.WriteString(pad(row[c], widths[i]))
			if i < len(cols)-1 {
				rb.WriteString("  ")
			}
		}
		_, _ = fmt.Fprintln(o.Writer, strings.TrimRight(rb.String(), " "))
	}
	return nil
}

func pad(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
