// Code generated by goa v3.8.4, DO NOT EDIT.
//
// calc service
//
// Command:
// $ goa gen goa.design/plugins/v3/docs/examples/calc/design -o
// $(GOPATH)/src/goa.design/plugins/docs/examples/calc

package calc

import (
	"context"
)

// The calc service performs additions on numbers
type Service interface {
	// Add implements add.
	Add(context.Context, *AddPayload, AddServerStream) (err error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "calc"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [1]string{"add"}

// AddServerStream is the interface a "add" endpoint server stream must satisfy.
type AddServerStream interface {
	// Send streams instances of "int".
	Send(int) error
	// Recv reads instances of "AddStreamingPayload" from the stream.
	Recv() (*AddStreamingPayload, error)
	// Close closes the stream.
	Close() error
}

// AddClientStream is the interface a "add" endpoint client stream must satisfy.
type AddClientStream interface {
	// Send streams instances of "AddStreamingPayload".
	Send(*AddStreamingPayload) error
	// Recv reads instances of "int" from the stream.
	Recv() (int, error)
	// Close closes the stream.
	Close() error
}

// AddPayload is the payload type of the calc service add method.
type AddPayload struct {
	// Left operand
	Left int
	// Right operand
	Right int
}

// AddStreamingPayload is the streaming payload type of the calc service add
// method.
type AddStreamingPayload struct {
	// Left operand
	A int
	// Right operand
	B int
}
