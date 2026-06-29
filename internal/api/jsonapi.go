package api

import "encoding/json"

// This file centralizes the JSON:API document envelope so resource types stay flat and
// table-friendly. Lemon Squeezy wraps every record as
//
//	{ "data": { "type": "products", "id": "1", "attributes": {…}, "relationships": {…} } }
//
// and lists add top-level "meta.page" + "links". Decoding flattens id (and relationships)
// into the attributes object before unmarshaling into T, so a resource struct is just the
// attribute fields plus `ID ID json:"id"` — exactly like a plain REST resource.

// resourceObject is one JSON:API record (the contents of "data").
type resourceObject struct {
	Type          string          `json:"type"`
	ID            ID              `json:"id"`
	Attributes    json.RawMessage `json:"attributes,omitempty"`
	Relationships json.RawMessage `json:"relationships,omitempty"`
	Links         json.RawMessage `json:"links,omitempty"`
}

// singleDoc is a single-resource response document.
type singleDoc struct {
	Data     resourceObject  `json:"data"`
	Included json.RawMessage `json:"included,omitempty"`
	Links    Links           `json:"links"`
}

// listDoc is a collection response document.
type listDoc struct {
	Data     []resourceObject `json:"data"`
	Included json.RawMessage  `json:"included,omitempty"`
	Links    Links            `json:"links"`
	Meta     Meta             `json:"meta"`
}

// Links holds the JSON:API pagination links. Next is the signal the --all walker follows.
type Links struct {
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
	Next  string `json:"next,omitempty"`
	Prev  string `json:"prev,omitempty"`
}

// Meta carries the page metadata Lemon Squeezy returns on list endpoints.
type Meta struct {
	Page PageMeta `json:"page"`
}

// PageMeta mirrors Lemon Squeezy's meta.page object (used to decide when --all is done).
type PageMeta struct {
	CurrentPage int `json:"currentPage"`
	From        int `json:"from"`
	To          int `json:"to"`
	PerPage     int `json:"perPage"`
	LastPage    int `json:"lastPage"`
	Total       int `json:"total"`
}

// flattenObject merges a record's id (and relationships) into its attributes object,
// yielding a flat map ready to unmarshal into a resource struct. Relationships are kept
// under the "relationships" key so a struct may opt to capture them, but most don't.
func flattenObject(obj resourceObject) (map[string]json.RawMessage, error) {
	m := map[string]json.RawMessage{}
	if len(obj.Attributes) > 0 {
		if err := json.Unmarshal(obj.Attributes, &m); err != nil {
			return nil, err
		}
	}
	// Attributes never legitimately contain "id"/"type" in JSON:API, so overwriting is safe
	// and gives every record a consistent, string-typed id for rendering and -o id.
	idJSON, err := json.Marshal(obj.ID.String())
	if err != nil {
		return nil, err
	}
	m["id"] = idJSON
	if obj.Type != "" {
		if t, err := json.Marshal(obj.Type); err == nil {
			m["type"] = t
		}
	}
	if len(obj.Relationships) > 0 {
		m["relationships"] = obj.Relationships
	}
	return m, nil
}

// decodeOne flattens a single record and unmarshals it into T.
func decodeOne[T any](obj resourceObject) (T, error) {
	var t T
	m, err := flattenObject(obj)
	if err != nil {
		return t, err
	}
	merged, err := json.Marshal(m)
	if err != nil {
		return t, err
	}
	if err := json.Unmarshal(merged, &t); err != nil {
		return t, err
	}
	return t, nil
}

// decodeList flattens every record in a collection document into a slice of T.
func decodeList[T any](objs []resourceObject) ([]T, error) {
	out := make([]T, 0, len(objs))
	for _, o := range objs {
		t, err := decodeOne[T](o)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// WriteBody is the payload for create/update. The generic core wraps it into a JSON:API
// document ({"data":{"type","id","attributes","relationships"}}), so callers only supply
// the attributes and (when needed) relationships — never the envelope.
type WriteBody struct {
	Type          string         `json:"-"`
	ID            string         `json:"-"`
	Attributes    map[string]any `json:"-"`
	Relationships map[string]any `json:"-"`
}

// writeDoc is the on-the-wire JSON:API write envelope.
type writeDoc struct {
	Data writeData `json:"data"`
}

type writeData struct {
	Type          string         `json:"type"`
	ID            string         `json:"id,omitempty"`
	Attributes    map[string]any `json:"attributes,omitempty"`
	Relationships map[string]any `json:"relationships,omitempty"`
}

// document builds the JSON:API write envelope for a WriteBody, defaulting the type to the
// resource's own type when the caller left it blank.
func (w WriteBody) document(defaultType string) writeDoc {
	t := w.Type
	if t == "" {
		t = defaultType
	}
	return writeDoc{Data: writeData{
		Type:          t,
		ID:            w.ID,
		Attributes:    w.Attributes,
		Relationships: w.Relationships,
	}}
}

// Relationship builds a JSON:API to-one relationship object, the shape Lemon Squeezy expects
// under relationships (e.g. {"store":{"data":{"type":"stores","id":"1"}}}). Returns nil when
// id is empty so optional relationships are omitted cleanly.
func Relationship(typ, id string) map[string]any {
	if id == "" {
		return nil
	}
	return map[string]any{"data": map[string]any{"type": typ, "id": id}}
}
