// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package rpc

import (
	"bytes"
	"context"
	"testing"

	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/encoding/raw"
)

func TestRegisterDispatcher_OK(t *testing.T) {
	RegisterDispatcher(defaultYarpcDispatcher)
}

func makeRequest() *transport.Request {
	return &transport.Request{
		Caller:    "the test suite",
		Service:   "any service",
		Encoding:  raw.Encoding,
		Procedure: "hello",
		Body:      bytes.NewReader([]byte{1, 2, 3}),
	}
}

func makeHandler(err error) transport.UnaryHandler {
	return dummyHandler{
		err: err,
	}
}

type dummyHandler struct {
	err error
}

func (d dummyHandler) Handle(
	ctx context.Context,
	r *transport.Request,
	w transport.ResponseWriter,
) error {
	return d.err
}
