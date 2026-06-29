package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// product mirrors a flat Lemon Squeezy resource struct (id + attributes) used across tests.
type product struct {
	ID      ID     `json:"id"`
	StoreID Int    `json:"store_id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
}

func TestFlattenObject_MergesIDAndType(t *testing.T) {
	obj := resourceObject{
		Type:       "products",
		ID:         "55",
		Attributes: json.RawMessage(`{"store_id":1,"name":"Pro","status":"published"}`),
	}
	m, err := flattenObject(obj)
	require.NoError(t, err)
	assert.JSONEq(t, `"55"`, string(m["id"]))
	assert.JSONEq(t, `"products"`, string(m["type"]))
	assert.JSONEq(t, `1`, string(m["store_id"]))
}

func TestDecodeOne(t *testing.T) {
	obj := resourceObject{
		Type:       "products",
		ID:         "55",
		Attributes: json.RawMessage(`{"store_id":"7","name":"Pro","status":"published"}`),
	}
	got, err := decodeOne[product](obj)
	require.NoError(t, err)
	assert.Equal(t, ID("55"), got.ID)
	assert.Equal(t, Int(7), got.StoreID) // store_id arrived as a string, decoded via Int
	assert.Equal(t, "Pro", got.Name)
}

func TestDecodeList(t *testing.T) {
	objs := []resourceObject{
		{Type: "products", ID: "1", Attributes: json.RawMessage(`{"name":"A"}`)},
		{Type: "products", ID: "2", Attributes: json.RawMessage(`{"name":"B"}`)},
	}
	got, err := decodeList[product](objs)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, ID("2"), got[1].ID)
	assert.Equal(t, "B", got[1].Name)
}

func TestWriteBody_Document(t *testing.T) {
	w := WriteBody{
		Attributes:    map[string]any{"name": "Acme"},
		Relationships: map[string]any{"store": Relationship("stores", "3")},
	}
	doc := w.document("customers")
	b, err := json.Marshal(doc)
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":{"type":"customers","attributes":{"name":"Acme"},"relationships":{"store":{"data":{"type":"stores","id":"3"}}}}}`, string(b))

	// Explicit type and id (update path) are preserved.
	w2 := WriteBody{Type: "subscriptions", ID: "9", Attributes: map[string]any{"cancelled": true}}
	b2, _ := json.Marshal(w2.document("ignored"))
	assert.JSONEq(t, `{"data":{"type":"subscriptions","id":"9","attributes":{"cancelled":true}}}`, string(b2))
}

func TestRelationship_OmitsEmpty(t *testing.T) {
	assert.Nil(t, Relationship("stores", ""))
	assert.NotNil(t, Relationship("stores", "1"))
}
