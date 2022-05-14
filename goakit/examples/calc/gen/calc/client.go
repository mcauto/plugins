// Code generated by goa v3.7.4, DO NOT EDIT.
//
// calc client
//
// Command:
// $ goa gen goa.design/plugins/v3/goakit/examples/calc/design -o
// $(GOPATH)/src/goa.design/plugins/goakit/examples/calc

package calc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Client is the "calc" service client.
type Client struct {
	AddEndpoint endpoint.Endpoint
}

// NewClient initializes a "calc" service client given the endpoints.
func NewClient(add endpoint.Endpoint) *Client {
	return &Client{
		AddEndpoint: add,
	}
}

// Add calls the "add" endpoint of the "calc" service.
func (c *Client) Add(ctx context.Context, p *AddPayload) (res int, err error) {
	var ires interface{}
	ires, err = c.AddEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(int), nil
}
