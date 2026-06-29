package api

import (
	"net/url"
	"strconv"
	"strings"
)

// ListParams drives a single list request. Lemon Squeezy uses JSON:API page-based
// pagination (page[size] 1-100, default 10; page[number] 1-based), resource filters
// (filter[key]=value), sparse includes (include=a,b), and an optional sort field.
type ListParams struct {
	PageSize   int               // items per page (1-100); 0 lets the API default (10)
	PageNumber int               // 1-based page index; 0 == first page
	Filters    map[string]string // filter[key]=value (resource-specific)
	Include    []string          // JSON:API related resources to embed
	Sort       string            // sort field; prefix "-" for descending
}

// values renders the params as url.Values, omitting zero/empty fields so the request stays
// minimal and deterministic. Bracketed keys (page[size], filter[store_id]) are percent-
// escaped on encode, which Lemon Squeezy accepts.
func (p ListParams) values() url.Values {
	v := url.Values{}
	if p.PageSize > 0 {
		v.Set("page[size]", strconv.Itoa(p.PageSize))
	}
	if p.PageNumber > 0 {
		v.Set("page[number]", strconv.Itoa(p.PageNumber))
	}
	for k, val := range p.Filters {
		if val != "" {
			v.Set("filter["+k+"]", val)
		}
	}
	if len(p.Include) > 0 {
		v.Set("include", strings.Join(p.Include, ","))
	}
	if p.Sort != "" {
		v.Set("sort", p.Sort)
	}
	return v
}
