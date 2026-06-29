package api

// This file holds the licensing resources (the JSON:API ones). The non-JSON:API License API
// (activate/validate/deactivate) lives in license_api.go.

// LicenseKey — https://docs.lemonsqueezy.com/api/license-keys. Read + update (disable/limit).
type LicenseKey struct {
	ID              ID     `json:"id,omitempty"`
	StoreID         Int    `json:"store_id,omitempty"`
	CustomerID      Int    `json:"customer_id,omitempty"`
	OrderID         Int    `json:"order_id,omitempty"`
	OrderItemID     Int    `json:"order_item_id,omitempty"`
	ProductID       Int    `json:"product_id,omitempty"`
	UserName        string `json:"user_name,omitempty"`
	UserEmail       string `json:"user_email,omitempty"`
	Key             string `json:"key,omitempty"`
	KeyShort        string `json:"key_short,omitempty"`
	ActivationLimit Int    `json:"activation_limit,omitempty"`
	InstancesCount  Int    `json:"instances_count,omitempty"`
	Disabled        Bool   `json:"disabled,omitempty"`
	Status          string `json:"status,omitempty"`
	StatusFormatted string `json:"status_formatted,omitempty"`
	ExpiresAt       string `json:"expires_at,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

// LicenseKeys returns a typed handle to the /license-keys resource.
func (c *Client) LicenseKeys() *Resource[LicenseKey] {
	return NewResource[LicenseKey](c, "license-keys")
}

// LicenseKeyInstance — https://docs.lemonsqueezy.com/api/license-key-instances. Read-only.
type LicenseKeyInstance struct {
	ID           ID     `json:"id,omitempty"`
	LicenseKeyID Int    `json:"license_key_id,omitempty"`
	Identifier   string `json:"identifier,omitempty"`
	Name         string `json:"name,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// LicenseKeyInstances returns a typed handle to the /license-key-instances resource.
func (c *Client) LicenseKeyInstances() *Resource[LicenseKeyInstance] {
	return NewResource[LicenseKeyInstance](c, "license-key-instances")
}
