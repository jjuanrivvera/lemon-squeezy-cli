package api

// Store — see https://docs.lemonsqueezy.com/api/stores. Read-only.
type Store struct {
	ID               ID     `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	Slug             string `json:"slug,omitempty"`
	Domain           string `json:"domain,omitempty"`
	URL              string `json:"url,omitempty"`
	Currency         string `json:"currency,omitempty"`
	Plan             string `json:"plan,omitempty"`
	Country          string `json:"country,omitempty"`
	CountryNice      string `json:"country_nicename,omitempty"`
	TotalSales       Int    `json:"total_sales,omitempty"`
	TotalRevenue     Int    `json:"total_revenue,omitempty"`
	ThirtyDaySales   Int    `json:"thirty_day_sales,omitempty"`
	ThirtyDayRevenue Int    `json:"thirty_day_revenue,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

// Stores returns a typed handle to the /stores resource.
func (c *Client) Stores() *Resource[Store] { return NewResource[Store](c, "stores") }
