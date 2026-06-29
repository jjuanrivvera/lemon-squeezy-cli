package api

// This file holds the read-only catalog resources: products, variants, prices, and files.
// See https://docs.lemonsqueezy.com/api.

// Product — https://docs.lemonsqueezy.com/api/products
type Product struct {
	ID              ID     `json:"id,omitempty"`
	StoreID         Int    `json:"store_id,omitempty"`
	Name            string `json:"name,omitempty"`
	Slug            string `json:"slug,omitempty"`
	Description     string `json:"description,omitempty"`
	Status          string `json:"status,omitempty"`
	StatusFormatted string `json:"status_formatted,omitempty"`
	Price           Money  `json:"price,omitempty"`
	PriceFormatted  string `json:"price_formatted,omitempty"`
	FromPrice       Money  `json:"from_price,omitempty"`
	ToPrice         Money  `json:"to_price,omitempty"`
	BuyNowURL       string `json:"buy_now_url,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// Products returns a typed handle to the /products resource.
func (c *Client) Products() *Resource[Product] { return NewResource[Product](c, "products") }

// Variant — https://docs.lemonsqueezy.com/api/variants
type Variant struct {
	ID             ID     `json:"id,omitempty"`
	ProductID      Int    `json:"product_id,omitempty"`
	Name           string `json:"name,omitempty"`
	Slug           string `json:"slug,omitempty"`
	Description    string `json:"description,omitempty"`
	Price          Money  `json:"price,omitempty"`
	IsSubscription Bool   `json:"is_subscription,omitempty"`
	Interval       string `json:"interval,omitempty"`
	IntervalCount  Int    `json:"interval_count,omitempty"`
	HasFreeTrial   Bool   `json:"has_free_trial,omitempty"`
	Status         string `json:"status,omitempty"`
	Sort           Int    `json:"sort,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// Variants returns a typed handle to the /variants resource.
func (c *Client) Variants() *Resource[Variant] { return NewResource[Variant](c, "variants") }

// Price — https://docs.lemonsqueezy.com/api/prices
type Price struct {
	ID                  ID     `json:"id,omitempty"`
	VariantID           Int    `json:"variant_id,omitempty"`
	Category            string `json:"category,omitempty"`
	Scheme              string `json:"scheme,omitempty"`
	UsageAggregation    string `json:"usage_aggregation,omitempty"`
	UnitPrice           Money  `json:"unit_price,omitempty"`
	UnitPriceDecimal    Money  `json:"unit_price_decimal,omitempty"`
	SetupFee            Money  `json:"setup_fee,omitempty"`
	PackageSize         Int    `json:"package_size,omitempty"`
	TierMode            string `json:"tier_mode,omitempty"`
	RenewalIntervalUnit string `json:"renewal_interval_unit,omitempty"`
	RenewalIntervalQty  Int    `json:"renewal_interval_quantity,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

// Prices returns a typed handle to the /prices resource.
func (c *Client) Prices() *Resource[Price] { return NewResource[Price](c, "prices") }

// File — https://docs.lemonsqueezy.com/api/files
type File struct {
	ID            ID     `json:"id,omitempty"`
	VariantID     Int    `json:"variant_id,omitempty"`
	Identifier    string `json:"identifier,omitempty"`
	Name          string `json:"name,omitempty"`
	Extension     string `json:"extension,omitempty"`
	DownloadURL   string `json:"download_url,omitempty"`
	Size          Int    `json:"size,omitempty"`
	SizeFormatted string `json:"size_formatted,omitempty"`
	Version       string `json:"version,omitempty"`
	Sort          Int    `json:"sort,omitempty"`
	Status        string `json:"status,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

// Files returns a typed handle to the /files resource.
func (c *Client) Files() *Resource[File] { return NewResource[File](c, "files") }
