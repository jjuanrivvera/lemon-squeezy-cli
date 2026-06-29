package api

// This file holds the recurring-billing resources: subscriptions, subscription items,
// subscription invoices, and usage records. See https://docs.lemonsqueezy.com/api.

// Subscription — https://docs.lemonsqueezy.com/api/subscriptions. Read + update + cancel
// (cancel is a DELETE on the subscription).
type Subscription struct {
	ID              ID     `json:"id,omitempty"`
	StoreID         Int    `json:"store_id,omitempty"`
	CustomerID      Int    `json:"customer_id,omitempty"`
	OrderID         Int    `json:"order_id,omitempty"`
	ProductID       Int    `json:"product_id,omitempty"`
	VariantID       Int    `json:"variant_id,omitempty"`
	ProductName     string `json:"product_name,omitempty"`
	VariantName     string `json:"variant_name,omitempty"`
	UserName        string `json:"user_name,omitempty"`
	UserEmail       string `json:"user_email,omitempty"`
	Status          string `json:"status,omitempty"`
	StatusFormatted string `json:"status_formatted,omitempty"`
	CardBrand       string `json:"card_brand,omitempty"`
	CardLastFour    string `json:"card_last_four,omitempty"`
	Cancelled       Bool   `json:"cancelled,omitempty"`
	TrialEndsAt     string `json:"trial_ends_at,omitempty"`
	RenewsAt        string `json:"renews_at,omitempty"`
	EndsAt          string `json:"ends_at,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// Subscriptions returns a typed handle to the /subscriptions resource.
func (c *Client) Subscriptions() *Resource[Subscription] {
	return NewResource[Subscription](c, "subscriptions")
}

// SubscriptionItem — https://docs.lemonsqueezy.com/api/subscription-items. Read + update,
// plus a current-usage action for usage-based items.
type SubscriptionItem struct {
	ID             ID     `json:"id,omitempty"`
	SubscriptionID Int    `json:"subscription_id,omitempty"`
	PriceID        Int    `json:"price_id,omitempty"`
	Quantity       Int    `json:"quantity,omitempty"`
	IsUsageBased   Bool   `json:"is_usage_based,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// SubscriptionItems returns a typed handle to the /subscription-items resource.
func (c *Client) SubscriptionItems() *Resource[SubscriptionItem] {
	return NewResource[SubscriptionItem](c, "subscription-items")
}

// SubscriptionInvoice — https://docs.lemonsqueezy.com/api/subscription-invoices. Read-only,
// plus refund and generate-invoice actions.
type SubscriptionInvoice struct {
	ID              ID     `json:"id,omitempty"`
	StoreID         Int    `json:"store_id,omitempty"`
	SubscriptionID  Int    `json:"subscription_id,omitempty"`
	CustomerID      Int    `json:"customer_id,omitempty"`
	UserName        string `json:"user_name,omitempty"`
	UserEmail       string `json:"user_email,omitempty"`
	BillingReason   string `json:"billing_reason,omitempty"`
	CardBrand       string `json:"card_brand,omitempty"`
	CardLastFour    string `json:"card_last_four,omitempty"`
	Currency        string `json:"currency,omitempty"`
	Status          string `json:"status,omitempty"`
	StatusFormatted string `json:"status_formatted,omitempty"`
	Total           Money  `json:"total,omitempty"`
	TotalFormatted  string `json:"total_formatted,omitempty"`
	Refunded        Bool   `json:"refunded,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// SubscriptionInvoices returns a typed handle to the /subscription-invoices resource.
func (c *Client) SubscriptionInvoices() *Resource[SubscriptionInvoice] {
	return NewResource[SubscriptionInvoice](c, "subscription-invoices")
}

// UsageRecord — https://docs.lemonsqueezy.com/api/usage-records. Read + create.
type UsageRecord struct {
	ID                 ID     `json:"id,omitempty"`
	SubscriptionItemID Int    `json:"subscription_item_id,omitempty"`
	Quantity           Int    `json:"quantity,omitempty"`
	Action             string `json:"action,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

// UsageRecords returns a typed handle to the /usage-records resource.
func (c *Client) UsageRecords() *Resource[UsageRecord] {
	return NewResource[UsageRecord](c, "usage-records")
}
