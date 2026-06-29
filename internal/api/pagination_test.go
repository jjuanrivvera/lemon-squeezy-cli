package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListParams_Values(t *testing.T) {
	p := ListParams{
		PageSize:   50,
		PageNumber: 2,
		Sort:       "-createdAt",
		Include:    []string{"store", "variants"},
		Filters:    map[string]string{"store_id": "9", "status": ""},
	}
	v := p.values()
	assert.Equal(t, "50", v.Get("page[size]"))
	assert.Equal(t, "2", v.Get("page[number]"))
	assert.Equal(t, "-createdAt", v.Get("sort"))
	assert.Equal(t, "store,variants", v.Get("include"))
	assert.Equal(t, "9", v.Get("filter[store_id]"))
	_, hasEmpty := v["filter[status]"]
	assert.False(t, hasEmpty, "empty filter values are omitted")
}

func TestListParams_Empty(t *testing.T) {
	v := ListParams{}.values()
	assert.Empty(t, v, "a zero ListParams sends no query params")
}
