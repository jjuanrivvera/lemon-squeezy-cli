package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID_UnmarshalStringAndNumber(t *testing.T) {
	cases := map[string]ID{
		`"abc"`:               "abc",
		`123`:                 "123",
		`9007199254740993`:    "9007199254740993", // > 2^53, must not lose precision
		`null`:                "",
		`""`:                  "",
		`"38b1460a-5104-406"`: "38b1460a-5104-406",
	}
	for in, want := range cases {
		var id ID
		require.NoError(t, json.Unmarshal([]byte(in), &id), "input %s", in)
		assert.Equal(t, want, id, "input %s", in)
	}
	// Round-trips as a string regardless of source.
	b, err := json.Marshal(ID("42"))
	require.NoError(t, err)
	assert.Equal(t, `"42"`, string(b))
}

func TestInt_UnmarshalNumberAndString(t *testing.T) {
	cases := map[string]Int{
		`42`:      42,
		`"42"`:    42,
		`null`:    0,
		`""`:      0,
		`1000000`: 1000000,
		`12.0`:    12,
	}
	for in, want := range cases {
		var n Int
		require.NoError(t, json.Unmarshal([]byte(in), &n), "input %s", in)
		assert.Equal(t, want, n, "input %s", in)
	}
	var n Int
	assert.Error(t, json.Unmarshal([]byte(`"not-a-number"`), &n))
	b, _ := json.Marshal(Int(7))
	assert.Equal(t, "7", string(b))
}

func TestBool_UnmarshalBoolAndString(t *testing.T) {
	cases := map[string]Bool{
		`true`:    true,
		`false`:   false,
		`"true"`:  true,
		`"1"`:     true,
		`"yes"`:   true,
		`"no"`:    false,
		`"0"`:     false,
		`null`:    false,
		`"on"`:    true,
		`"FALSE"`: false,
	}
	for in, want := range cases {
		var b Bool
		require.NoError(t, json.Unmarshal([]byte(in), &b), "input %s", in)
		assert.Equal(t, want, b, "input %s", in)
	}
	out, _ := json.Marshal(Bool(true))
	assert.Equal(t, "true", string(out))
}

func TestMoney_ExactDecimalText(t *testing.T) {
	cases := map[string]Money{
		`999`:      "999",
		`"1999"`:   "1999",
		`0`:        "0",
		`null`:     "",
		`12.34`:    "12.34",
		`"$10.00"`: "$10.00",
	}
	for in, want := range cases {
		var m Money
		require.NoError(t, json.Unmarshal([]byte(in), &m), "input %s", in)
		assert.Equal(t, want, m, "input %s", in)
	}
	b, _ := json.Marshal(Money("1999"))
	assert.Equal(t, `"1999"`, string(b))
}

func TestStringOrSlice(t *testing.T) {
	var s StringOrSlice
	require.NoError(t, json.Unmarshal([]byte(`"x"`), &s))
	assert.Equal(t, StringOrSlice{"x"}, s)
	require.NoError(t, json.Unmarshal([]byte(`["a","b"]`), &s))
	assert.Equal(t, StringOrSlice{"a", "b"}, s)
	require.NoError(t, json.Unmarshal([]byte(`null`), &s))
	assert.Nil(t, s)
}

// Fuzz the flexible decoders: they must never panic on arbitrary JSON-ish input.
func FuzzFlexibleTypes(f *testing.F) {
	for _, seed := range []string{`"x"`, `1`, `12.5`, `null`, `true`, `"true"`, `[]`, `{}`, `1e9`, `""`, `-5`} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in string) {
		var id ID
		_ = json.Unmarshal([]byte(in), &id)
		var n Int
		_ = json.Unmarshal([]byte(in), &n)
		var b Bool
		_ = json.Unmarshal([]byte(in), &b)
		var m Money
		_ = json.Unmarshal([]byte(in), &m)
		var s StringOrSlice
		_ = json.Unmarshal([]byte(in), &s)
	})
}
