package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sample struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Origin string `json:"origin"`
}

var rows = []sample{
	{"1", "Bengal", "US"},
	{"2", "Siamese", "TH"},
}

func TestParseFormat(t *testing.T) {
	for _, f := range []string{"table", "json", "yaml", "csv", "JSON"} {
		_, err := ParseFormat(f)
		require.NoError(t, err)
	}
	_, err := ParseFormat("xml")
	assert.Error(t, err)
}

func TestRender_JSON(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(rows, nil, Options{Format: FormatJSON, Writer: &buf}))
	assert.Contains(t, buf.String(), `"name": "Bengal"`)
}

func TestRender_YAML(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(rows, nil, Options{Format: FormatYAML, Writer: &buf}))
	assert.Contains(t, buf.String(), "name: Bengal")
}

func TestRender_CSV(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(rows, []string{"id", "name", "origin"}, Options{Format: FormatCSV, Writer: &buf}))
	out := buf.String()
	assert.Contains(t, out, "id,name,origin")
	assert.Contains(t, out, "1,Bengal,US")
}

func TestRender_Table(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(rows, []string{"id", "name", "origin"}, Options{Format: FormatTable, Writer: &buf, NoColor: true}))
	out := buf.String()
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "Bengal")
	// No ANSI color codes when NoColor + non-TTY.
	assert.NotContains(t, out, "\033[")
}

func TestRender_TableColumnsSubset(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(rows, []string{"id", "name", "origin"}, Options{Format: FormatTable, Columns: []string{"name"}, Writer: &buf, NoColor: true}))
	out := buf.String()
	assert.Contains(t, out, "NAME")
	assert.NotContains(t, out, "ORIGIN")
}

func TestRender_SingleObject(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(rows[0], []string{"id", "name"}, Options{Format: FormatTable, Writer: &buf, NoColor: true}))
	assert.Contains(t, buf.String(), "Bengal")
}

func TestRender_DefaultColumnsFromUnion(t *testing.T) {
	var buf bytes.Buffer
	// No defaultColumns => fall back to the sorted union of keys.
	require.NoError(t, Render(rows, nil, Options{Format: FormatCSV, Writer: &buf}))
	header := strings.SplitN(buf.String(), "\n", 2)[0]
	assert.Equal(t, "id,name,origin", header)
}

func TestStringify(t *testing.T) {
	assert.Equal(t, "", stringify(nil))
	assert.Equal(t, "5", stringify(float64(5)))
	assert.Equal(t, "1.5", stringify(float64(1.5)))
	assert.Equal(t, "true", stringify(true))
	assert.Equal(t, "hi", stringify("hi"))
	assert.Equal(t, `["a","b"]`, stringify([]any{"a", "b"}))
}

func TestRender_NumberFormatting(t *testing.T) {
	var buf bytes.Buffer
	type rec struct {
		Value int `json:"value"`
	}
	require.NoError(t, Render([]rec{{42}}, []string{"value"}, Options{Format: FormatCSV, Writer: &buf}))
	assert.Contains(t, buf.String(), "42")
	assert.NotContains(t, buf.String(), "42.0")
}

func TestApplyJQ(t *testing.T) {
	// One field per row -> a slice of scalars.
	got, err := applyJQ(".[].name", rows)
	require.NoError(t, err)
	assert.Equal(t, []any{"Bengal", "Siamese"}, got)

	// A single object field -> one scalar.
	got, err = applyJQ(".name", rows[0])
	require.NoError(t, err)
	assert.Equal(t, "Bengal", got)

	// Reshape into new objects: the [...] collector yields a single array output.
	got, err = applyJQ("[.[] | {n: .name}]", rows)
	require.NoError(t, err)
	arr, ok := got.([]any)
	require.True(t, ok)
	assert.Len(t, arr, 2)
}

func TestApplyJQ_NoMatch(t *testing.T) {
	got, err := applyJQ(".missing", rows[0])
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestApplyJQ_InvalidExpression(t *testing.T) {
	_, err := applyJQ(".[", rows)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --jq expression")
}

func TestApplyJQ_RuntimeError(t *testing.T) {
	// Indexing a string with a field is a gojq runtime error.
	_, err := applyJQ(".name.oops", rows[0])
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--jq")
}

func TestRender_WithJQ_JSON(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(rows, nil, Options{Format: FormatJSON, JQ: ".[].id", Writer: &buf}))
	out := buf.String()
	assert.Contains(t, out, `"1"`)
	assert.Contains(t, out, `"2"`)
	assert.NotContains(t, out, "Bengal") // name filtered out
}

func TestRender_WithJQ_Table(t *testing.T) {
	var buf bytes.Buffer
	// Reshape then render as a table: the jq output drives the columns.
	require.NoError(t, Render(rows, nil, Options{
		Format: FormatTable, NoColor: true, JQ: "[.[] | {id, name}]", Writer: &buf,
	}))
	out := buf.String()
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "Bengal")
	assert.NotContains(t, out, "ORIGIN")
}

func TestRender_WithJQ_Error(t *testing.T) {
	var buf bytes.Buffer
	err := Render(rows, nil, Options{Format: FormatJSON, JQ: "..[", Writer: &buf})
	require.Error(t, err)
}
