// Code generated by goa v3.7.1, DO NOT EDIT.
//
// calc HTTP server encoders and decoders
//
// Command:
// $ goa gen goa.design/plugins/v3/docs/examples/calc/design -o
// $(GOPATH)/src/goa.design/plugins/docs/examples/calc

package server

import (
	"net/http"
	"strconv"

	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

// DecodeAddRequest returns a decoder for requests sent to the calc add
// endpoint.
func DecodeAddRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			left  int
			right int
			err   error

			params = mux.Vars(r)
		)
		{
			leftRaw := params["left"]
			v, err2 := strconv.ParseInt(leftRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("left", leftRaw, "integer"))
			}
			left = int(v)
		}
		{
			rightRaw := params["right"]
			v, err2 := strconv.ParseInt(rightRaw, 10, strconv.IntSize)
			if err2 != nil {
				err = goa.MergeErrors(err, goa.InvalidFieldTypeError("right", rightRaw, "integer"))
			}
			right = int(v)
		}
		if err != nil {
			return nil, err
		}
		payload := NewAddPayload(left, right)

		return payload, nil
	}
}
