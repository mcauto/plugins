// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// calc HTTP client types
//
// Command:
// $ goa gen goa.design/plugins/security/examples/calc/calc/design

package client

import (
	calcsvc "goa.design/plugins/security/examples/calc/calc/gen/calc"
)

// LoginUnauthorizedResponseBody is the type of the "calc" service "login"
// endpoint HTTP response body for the "unauthorized" error.
type LoginUnauthorizedResponseBody string

// NewLoginUnauthorized builds a calc service login endpoint unauthorized error.
func NewLoginUnauthorized(body LoginUnauthorizedResponseBody) calcsvc.Unauthorized {
	v := calcsvc.Unauthorized(body)
	return v
}
