package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResource_List_DecodesEnvelopeAndMeta(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/products", r.URL.Path)
		assert.Equal(t, MediaType, r.Header.Get("Accept"))
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", MediaType)
		_, _ = w.Write([]byte(`{
			"meta":{"page":{"currentPage":1,"lastPage":3,"total":7}},
			"links":{"next":"https://x?page[number]=2"},
			"data":[
				{"type":"products","id":"1","attributes":{"name":"A","store_id":9}},
				{"type":"products","id":"2","attributes":{"name":"B","store_id":9}}
			]}`))
	})
	res := NewResource[product](c, "products")
	items, meta, err := res.List(context.Background(), ListParams{})
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, ID("1"), items[0].ID)
	assert.Equal(t, "A", items[0].Name)
	assert.Equal(t, Int(9), items[0].StoreID)
	assert.Equal(t, 3, meta.Page.LastPage)
	assert.Equal(t, 7, meta.Page.Total)
}

func TestResource_ListAll_WalksByLastPage(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		page := r.URL.Query().Get("page[number]")
		w.Header().Set("Content-Type", MediaType)
		// 3 pages, 1 item each; meta.lastPage=3 drives termination.
		fmt.Fprintf(w, `{"meta":{"page":{"currentPage":%s,"lastPage":3}},"data":[{"type":"products","id":"%s","attributes":{"name":"p"}}]}`, page, page)
	})
	res := NewResource[product](c, "products")
	items, err := res.ListAll(context.Background(), ListParams{})
	require.NoError(t, err)
	assert.Len(t, items, 3)
	assert.Equal(t, 3, calls)
}

func TestResource_ListAll_WalksByNextLink(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", MediaType)
		// No meta.lastPage; terminate when next link disappears (page 2 of 2).
		if calls == 1 {
			_, _ = w.Write([]byte(`{"meta":{"page":{}},"links":{"next":"x"},"data":[{"type":"o","id":"1","attributes":{}}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"meta":{"page":{}},"links":{},"data":[{"type":"o","id":"2","attributes":{}}]}`))
	})
	res := NewResource[product](c, "orders")
	items, err := res.ListAll(context.Background(), ListParams{})
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, 2, calls)
}

func TestResource_Get_WithInclude(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/products/55", r.URL.Path)
		assert.Equal(t, "store,variants", r.URL.Query().Get("include"))
		w.Header().Set("Content-Type", MediaType)
		_, _ = w.Write([]byte(`{"data":{"type":"products","id":"55","attributes":{"name":"Pro"}}}`))
	})
	res := NewResource[product](c, "products")
	got, err := res.Get(context.Background(), "55", "store", "variants")
	require.NoError(t, err)
	assert.Equal(t, ID("55"), got.ID)
	assert.Equal(t, "Pro", got.Name)
}

func TestResource_Create_WrapsEnvelope(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/customers", r.URL.Path)
		assert.Equal(t, MediaType, r.Header.Get("Content-Type"))
		var got map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		data := got["data"].(map[string]any)
		assert.Equal(t, "customers", data["type"])
		assert.Equal(t, "Acme", data["attributes"].(map[string]any)["name"])
		w.Header().Set("Content-Type", MediaType)
		_, _ = w.Write([]byte(`{"data":{"type":"customers","id":"100","attributes":{"name":"Acme"}}}`))
	})
	res := NewResource[product](c, "customers")
	var out product
	err := res.Create(context.Background(), WriteBody{Attributes: map[string]any{"name": "Acme"}}, &out)
	require.NoError(t, err)
	assert.Equal(t, ID("100"), out.ID)
}

func TestResource_Update_PATCHWithID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/subscriptions/9", r.URL.Path)
		var got map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		data := got["data"].(map[string]any)
		assert.Equal(t, "9", data["id"]) // JSON:API requires id in update document
		w.Header().Set("Content-Type", MediaType)
		_, _ = w.Write([]byte(`{"data":{"type":"subscriptions","id":"9","attributes":{"name":"x"}}}`))
	})
	res := NewResource[product](c, "subscriptions")
	var out product
	err := res.Update(context.Background(), "9", WriteBody{Attributes: map[string]any{"pause": nil}}, &out)
	require.NoError(t, err)
	assert.Equal(t, ID("9"), out.ID)
}

func TestResource_Delete(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/webhooks/3", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	res := NewResource[product](c, "webhooks")
	require.NoError(t, res.Delete(context.Background(), "3"))
}

func TestResource_Action_POSTSubPath(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/subscription-items/4/usage", r.URL.Path)
		w.Header().Set("Content-Type", MediaType)
		_, _ = w.Write([]byte(`{"data":{"type":"usage","id":"1","attributes":{}}}`))
	})
	res := NewResource[product](c, "subscription-items")
	var out map[string]any
	err := res.Action(context.Background(), http.MethodPost, "4/usage", nil, map[string]any{"x": 1}, &out)
	require.NoError(t, err)
}

func TestGetOne_Singleton(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/me", r.URL.Path)
		w.Header().Set("Content-Type", MediaType)
		_, _ = w.Write([]byte(`{"data":{"type":"users","id":"7","attributes":{"name":"Jo"}}}`))
	})
	got, err := GetOne[product](context.Background(), c, "users/me", nil)
	require.NoError(t, err)
	assert.Equal(t, ID("7"), got.ID)
	assert.Equal(t, "Jo", got.Name)
}

func TestResource_Get_APIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", MediaType)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"status":"404","title":"Not Found","detail":"No product"}]}`))
	})
	res := NewResource[product](c, "products")
	_, err := res.Get(context.Background(), "999")
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.StatusCode)
	assert.Contains(t, err.Error(), "No product")
}
