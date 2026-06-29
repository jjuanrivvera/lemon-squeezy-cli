package api

// Webhook — https://docs.lemonsqueezy.com/api/webhooks. Full CRUD. The `events` array lists
// the event names the endpoint subscribes to (e.g. order_created, subscription_updated).
type Webhook struct {
	ID         ID            `json:"id,omitempty"`
	StoreID    Int           `json:"store_id,omitempty"`
	URL        string        `json:"url,omitempty"`
	Events     StringOrSlice `json:"events,omitempty"`
	LastSentAt string        `json:"last_sent_at,omitempty"`
	CreatedAt  string        `json:"created_at,omitempty"`
	UpdatedAt  string        `json:"updated_at,omitempty"`
}

// Webhooks returns a typed handle to the /webhooks resource.
func (c *Client) Webhooks() *Resource[Webhook] { return NewResource[Webhook](c, "webhooks") }
