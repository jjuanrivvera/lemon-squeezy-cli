package api

// This file holds the transactional resources: customers, orders, order items, checkouts,
// discounts, and discount redemptions. See https://docs.lemonsqueezy.com/api.

// Customer — https://docs.lemonsqueezy.com/api/customers. Supports create/update (and an
// `archive` action that PATCHes status=archived; customers cannot be deleted).
type Customer struct {
	ID              ID     `json:"id,omitempty"`
	StoreID         Int    `json:"store_id,omitempty"`
	Name            string `json:"name,omitempty"`
	Email           string `json:"email,omitempty"`
	Status          string `json:"status,omitempty"`
	StatusFormatted string `json:"status_formatted,omitempty"`
	City            string `json:"city,omitempty"`
	Region          string `json:"region,omitempty"`
	Country         string `json:"country,omitempty"`
	TotalRevenue    Int    `json:"total_revenue_currency,omitempty"`
	MRR             Int    `json:"mrr,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// Customers returns a typed handle to the /customers resource.
func (c *Client) Customers() *Resource[Customer] { return NewResource[Customer](c, "customers") }

// Order — https://docs.lemonsqueezy.com/api/orders. Read-only, plus refund and
// generate-invoice actions.
type Order struct {
	ID              ID     `json:"id,omitempty"`
	StoreID         Int    `json:"store_id,omitempty"`
	CustomerID      Int    `json:"customer_id,omitempty"`
	Identifier      string `json:"identifier,omitempty"`
	OrderNumber     Int    `json:"order_number,omitempty"`
	UserName        string `json:"user_name,omitempty"`
	UserEmail       string `json:"user_email,omitempty"`
	Currency        string `json:"currency,omitempty"`
	Subtotal        Money  `json:"subtotal,omitempty"`
	Total           Money  `json:"total,omitempty"`
	TotalFormatted  string `json:"total_formatted,omitempty"`
	Status          string `json:"status,omitempty"`
	StatusFormatted string `json:"status_formatted,omitempty"`
	Refunded        Bool   `json:"refunded,omitempty"`
	RefundedAt      string `json:"refunded_at,omitempty"`
	TaxName         string `json:"tax_name,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// Orders returns a typed handle to the /orders resource.
func (c *Client) Orders() *Resource[Order] { return NewResource[Order](c, "orders") }

// OrderItem — https://docs.lemonsqueezy.com/api/order-items. Read-only.
type OrderItem struct {
	ID          ID     `json:"id,omitempty"`
	OrderID     Int    `json:"order_id,omitempty"`
	ProductID   Int    `json:"product_id,omitempty"`
	VariantID   Int    `json:"variant_id,omitempty"`
	ProductName string `json:"product_name,omitempty"`
	VariantName string `json:"variant_name,omitempty"`
	Price       Money  `json:"price,omitempty"`
	Quantity    Int    `json:"quantity,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// OrderItems returns a typed handle to the /order-items resource.
func (c *Client) OrderItems() *Resource[OrderItem] { return NewResource[OrderItem](c, "order-items") }

// Checkout — https://docs.lemonsqueezy.com/api/checkouts. Read + create (a custom checkout).
type Checkout struct {
	ID          ID     `json:"id,omitempty"`
	StoreID     Int    `json:"store_id,omitempty"`
	VariantID   Int    `json:"variant_id,omitempty"`
	CustomPrice Int    `json:"custom_price,omitempty"`
	URL         string `json:"url,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	TestMode    Bool   `json:"test_mode,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// Checkouts returns a typed handle to the /checkouts resource.
func (c *Client) Checkouts() *Resource[Checkout] { return NewResource[Checkout](c, "checkouts") }

// Discount — https://docs.lemonsqueezy.com/api/discounts. Read + create + delete (no update).
type Discount struct {
	ID                   ID     `json:"id,omitempty"`
	StoreID              Int    `json:"store_id,omitempty"`
	Name                 string `json:"name,omitempty"`
	Code                 string `json:"code,omitempty"`
	Amount               Int    `json:"amount,omitempty"`
	AmountType           string `json:"amount_type,omitempty"`
	Status               string `json:"status,omitempty"`
	StatusFormatted      string `json:"status_formatted,omitempty"`
	IsLimitedRedemptions Bool   `json:"is_limited_redemptions,omitempty"`
	MaxRedemptions       Int    `json:"max_redemptions,omitempty"`
	Duration             string `json:"duration,omitempty"`
	StartsAt             string `json:"starts_at,omitempty"`
	ExpiresAt            string `json:"expires_at,omitempty"`
	CreatedAt            string `json:"created_at,omitempty"`
	UpdatedAt            string `json:"updated_at,omitempty"`
}

// Discounts returns a typed handle to the /discounts resource.
func (c *Client) Discounts() *Resource[Discount] { return NewResource[Discount](c, "discounts") }

// DiscountRedemption — https://docs.lemonsqueezy.com/api/discount-redemptions. Read-only.
type DiscountRedemption struct {
	ID                 ID     `json:"id,omitempty"`
	DiscountID         Int    `json:"discount_id,omitempty"`
	OrderID            Int    `json:"order_id,omitempty"`
	DiscountName       string `json:"discount_name,omitempty"`
	DiscountCode       string `json:"discount_code,omitempty"`
	DiscountAmount     Int    `json:"discount_amount,omitempty"`
	DiscountAmountType string `json:"discount_amount_type,omitempty"`
	Amount             Int    `json:"amount,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

// DiscountRedemptions returns a typed handle to the /discount-redemptions resource.
func (c *Client) DiscountRedemptions() *Resource[DiscountRedemption] {
	return NewResource[DiscountRedemption](c, "discount-redemptions")
}
