package api

import "context"

// User — https://docs.lemonsqueezy.com/api/users. The API only exposes the authenticated
// user via GET /v1/users/me (there is no collection), so this is a singleton, not a Resource[T].
type User struct {
	ID        ID     `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Color     string `json:"color,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// Me fetches the authenticated user (GET /users/me) — the canonical "whoami".
func (c *Client) Me(ctx context.Context) (*User, error) {
	return GetOne[User](ctx, c, "users/me", nil)
}
