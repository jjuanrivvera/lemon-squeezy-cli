package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Resource[T] is the generic CRUD handle. Every resource reuses it; the only per-resource
// code is the struct T and a Client accessor. This is the "generic core, thin resources"
// guarantee: adding a resource never edits this file. T is the flat attributes struct
// (id + attribute fields); the JSON:API envelope is decoded in jsonapi.go.
type Resource[T any] struct {
	client *Client
	path   string // collection path, e.g. "products"
	typ    string // JSON:API resource type, e.g. "products" (used to wrap write bodies)
}

// NewResource builds a typed handle to a collection. The JSON:API type defaults to the path,
// which is the convention across Lemon Squeezy (path "subscription-items" → type
// "subscription-items").
func NewResource[T any](c *Client, path string) *Resource[T] {
	return &Resource[T]{client: c, path: path, typ: path}
}

// WithType overrides the JSON:API resource type when it differs from the path (rare).
func (r *Resource[T]) WithType(typ string) *Resource[T] { r.typ = typ; return r }

// Type returns the JSON:API resource type used when wrapping write bodies.
func (r *Resource[T]) Type() string { return r.typ }

// List fetches one page and returns the items plus the page metadata (for callers that want
// totals/last-page without walking everything).
func (r *Resource[T]) List(ctx context.Context, p ListParams) ([]T, *Meta, error) {
	var doc listDoc
	if err := r.client.doJSON(ctx, http.MethodGet, r.path, p.values(), nil, &doc); err != nil {
		return nil, nil, err
	}
	items, err := decodeList[T](doc.Data)
	if err != nil {
		return nil, nil, err
	}
	return items, &doc.Meta, nil
}

// ListAll walks every page until meta.page reports the last page (or no Next link remains),
// honoring ctx cancellation between pages. Defaults to the max page size to minimize requests.
func (r *Resource[T]) ListAll(ctx context.Context, p ListParams) ([]T, error) {
	if p.PageSize <= 0 {
		p.PageSize = 100 // API maximum: fewest round-trips
	}
	if p.PageNumber <= 0 {
		p.PageNumber = 1
	}
	var all []T
	for {
		var doc listDoc
		if err := r.client.doJSON(ctx, http.MethodGet, r.path, p.values(), nil, &doc); err != nil {
			return all, err
		}
		items, err := decodeList[T](doc.Data)
		if err != nil {
			return all, err
		}
		all = append(all, items...)

		// Stop on the last page per meta, or when the server stops advertising a next link,
		// or when a short/empty page comes back (defensive: meta may be absent on some ops).
		last := doc.Meta.Page.LastPage
		if last > 0 && p.PageNumber >= last {
			break
		}
		if last == 0 && doc.Links.Next == "" {
			break
		}
		if len(items) == 0 {
			break
		}
		p.PageNumber++
		if err := ctx.Err(); err != nil {
			return all, err
		}
	}
	return all, nil
}

// Get fetches a single record by id, optionally embedding related resources via include.
func (r *Resource[T]) Get(ctx context.Context, id string, include ...string) (*T, error) {
	q := url.Values{}
	if len(include) > 0 {
		q.Set("include", joinNonEmpty(include))
	}
	var doc singleDoc
	if err := r.client.doJSON(ctx, http.MethodGet, r.path+"/"+url.PathEscape(id), q, nil, &doc); err != nil {
		return nil, err
	}
	out, err := decodeOne[T](doc.Data)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Create POSTs a JSON:API write document and decodes the created record into out (if non-nil).
// POST is never auto-retried (see retry.go) — re-sending a create could duplicate a record.
func (r *Resource[T]) Create(ctx context.Context, body WriteBody, out *T) error {
	return r.write(ctx, http.MethodPost, r.path, body, out)
}

// Update PATCHes a JSON:API write document (id required in the document per JSON:API) and
// decodes the updated record into out (if non-nil).
func (r *Resource[T]) Update(ctx context.Context, id string, body WriteBody, out *T) error {
	if body.ID == "" {
		body.ID = id
	}
	return r.write(ctx, http.MethodPatch, r.path+"/"+url.PathEscape(id), body, out)
}

// write marshals the JSON:API envelope, performs the request, and decodes the single-resource
// response into out when one is provided.
func (r *Resource[T]) write(ctx context.Context, method, path string, body WriteBody, out *T) error {
	data, err := json.Marshal(body.document(r.typ))
	if err != nil {
		return err
	}
	var doc singleDoc
	target := any(nil)
	if out != nil {
		target = &doc
	}
	if err := r.client.doJSON(ctx, method, path, nil, bytes.NewReader(data), target); err != nil {
		return err
	}
	if out != nil {
		decoded, err := decodeOne[T](doc.Data)
		if err != nil {
			return err
		}
		*out = decoded
	}
	return nil
}

// Delete removes a record by id (204 No Content on success).
func (r *Resource[T]) Delete(ctx context.Context, id string) error {
	err := r.client.doJSON(ctx, http.MethodDelete, r.path+"/"+url.PathEscape(id), nil, nil, nil)
	return err
}

// Action performs a custom verb relative to the collection, decoding the raw response into
// out. Method defaults to GET. subPath is appended to the collection path ("" == the
// collection itself). body, when non-nil, is marshaled verbatim (caller owns the envelope) —
// used for the handful of Lemon Squeezy actions that POST a JSON:API document to a sub-path.
func (r *Resource[T]) Action(ctx context.Context, method, subPath string, query url.Values, body, out any) error {
	if method == "" {
		method = http.MethodGet
	}
	path := r.path
	if subPath != "" {
		path = r.path + "/" + subPath
	}
	var reader *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	if reader == nil {
		err := r.client.doJSON(ctx, method, path, query, nil, out)
		return err
	}
	err := r.client.doJSON(ctx, method, path, query, reader, out)
	return err
}

// ActionOne performs a custom verb that returns a single JSON:API resource (e.g. POST
// /orders/{id}/refund returns the updated order, DELETE /subscriptions/{id} returns the
// cancelled subscription) and decodes it into T. body, when non-nil, is marshaled verbatim.
func (r *Resource[T]) ActionOne(ctx context.Context, method, subPath string, query url.Values, body any) (*T, error) {
	var doc singleDoc
	if err := r.Action(ctx, method, subPath, query, body, &doc); err != nil {
		return nil, err
	}
	out, err := decodeOne[T](doc.Data)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOne decodes a single JSON:API document at an arbitrary path into T. Used by singleton
// reads like `users me` (GET /users/me) that aren't a collection get-by-id.
func GetOne[T any](ctx context.Context, c *Client, path string, query url.Values) (*T, error) {
	var doc singleDoc
	if err := c.doJSON(ctx, http.MethodGet, path, query, nil, &doc); err != nil {
		return nil, err
	}
	out, err := decodeOne[T](doc.Data)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func joinNonEmpty(ss []string) string {
	out := ""
	for _, s := range ss {
		if s == "" {
			continue
		}
		if out != "" {
			out += ","
		}
		out += s
	}
	return out
}
