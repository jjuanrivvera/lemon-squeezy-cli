package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Flexible JSON types absorb the shape drift real APIs exhibit. Lemon Squeezy is mostly
// consistent (it ships a strict JSON:API), but ids appear as strings at the envelope level
// and as numbers inside attributes (store_id, product_id, …), money is integer cents, and a
// few fields are sometimes null. These types decode all of those without per-field special
// cases and render deterministically.

// ID unmarshals from a JSON string OR number and always marshals back as a string. JSON:API
// resource ids are strings, but related ids inside attributes (store_id, order_id, …) come
// back as integers; one flexible type renders every id consistently and avoids float64
// precision loss above 2^53.
type ID string

func (id *ID) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*id = ""
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*id = ID(s)
		return nil
	}
	// Number: json.Number keeps large integers in exact textual form (no 2^53 rounding).
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*id = ID(n.String())
	return nil
}

func (id ID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

func (id ID) String() string { return string(id) }

// Int accepts a JSON number OR a numeric string and stores an int64. Decoding via json.Number
// (Int64 before Float64) avoids the >2^53 precision loss a float64 decode would cause; NaN/Inf
// and non-numeric strings are rejected rather than silently zeroed.
type Int int64

func (n *Int) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*n = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if strings.TrimSpace(s) == "" {
			*n = 0
			return nil
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer %q: %w", s, err)
		}
		*n = Int(v)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(b, &num); err != nil {
		return err
	}
	if v, err := num.Int64(); err == nil {
		*n = Int(v)
		return nil
	}
	// Fall back to float for values written with a decimal point, rejecting NaN/Inf.
	f, err := num.Float64()
	if err != nil {
		return err
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return fmt.Errorf("invalid number %q", num.String())
	}
	*n = Int(int64(f))
	return nil
}

func (n Int) MarshalJSON() ([]byte, error) { return json.Marshal(int64(n)) }

func (n Int) Int64() int64 { return int64(n) }

// Bool accepts a real JSON bool OR the string forms "true"/"1"/"yes" (and their negatives).
// Some webhook/test payloads stringify booleans; this normalizes them.
type Bool bool

func (b *Bool) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*b = false
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "true", "1", "yes", "y", "on":
			*b = true
		default:
			*b = false
		}
		return nil
	}
	var v bool
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*b = Bool(v)
	return nil
}

func (b Bool) MarshalJSON() ([]byte, error) { return json.Marshal(bool(b)) }

// Money holds a monetary amount as exact decimal text, never float64, so cent/precision is
// never lost. Lemon Squeezy reports amounts as integer cents; this accepts a number or a
// numeric string and preserves the exact digits.
type Money string

func (m *Money) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*m = ""
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*m = Money(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	// Reject NaN/Inf, which json.Number permits as text but money must never be.
	if f, err := n.Float64(); err == nil && (math.IsNaN(f) || math.IsInf(f, 0)) {
		return fmt.Errorf("invalid money value %q", n.String())
	}
	*m = Money(n.String())
	return nil
}

func (m Money) MarshalJSON() ([]byte, error) { return json.Marshal(string(m)) }

func (m Money) String() string { return string(m) }

// StringOrSlice accepts a JSON string ("x") or an array of strings (["x","y"]) and
// normalizes to a slice — a common shape drift (e.g. webhook event lists).
type StringOrSlice []string

func (s *StringOrSlice) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*s = nil
		return nil
	}
	if b[0] == '[' {
		var arr []string
		if err := json.Unmarshal(b, &arr); err != nil {
			return err
		}
		*s = arr
		return nil
	}
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	*s = []string{str}
	return nil
}
