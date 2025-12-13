package kalshi

import (
	"context"
	"fmt"
)

// LoginRequest is described here:
// https://trading-api.readme.io/reference/login.
type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// LoginResponse is described here:
// https://trading-api.readme.io/reference/login.
type LoginResponse struct {
	Token  string `json:"token,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

// Login is described here:
// https://trading-api.readme.io/reference/login.
//
// The Client will stay authenticated after Login is called since it stores the
// token in the cookie state.
func (c *Client) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	var resp LoginResponse
	if err := c.request(ctx, request{
		Method:       "POST",
		Endpoint:     "login",
		JSONRequest:  req,
		JSONResponse: &resp,
	}); err != nil {
		return nil, fmt.Errorf("c.request: %w", err)
	}
	return &resp, nil
}

// Logout is described here:
// https://trading-api.readme.io/reference/logout.
func (c *Client) Logout(ctx context.Context) error {
	if err := c.request(ctx, request{
		Method:       "POST",
		Endpoint:     "logout",
		JSONRequest:  nil,
		JSONResponse: nil,
	}); err != nil {
		return fmt.Errorf("c.request: %w", err)
	}

	return nil
}
