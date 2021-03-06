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

package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/auth"
	"go.uber.org/fx/config"
	"go.uber.org/fx/internal/fxcontext"
	"go.uber.org/fx/service"
	"go.uber.org/fx/tracing"
	"go.uber.org/fx/ulog"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	jconfig "github.com/uber/jaeger-client-go/config"
)

var (
	_respOK   = &http.Response{StatusCode: http.StatusOK}
	_req      = httptest.NewRequest("", "http://localhost", nil)
	errClient = errors.New("client test error")
)

func TestExecutionChain(t *testing.T) {
	execChain := newExecutionChain([]Filter{}, getNopClient())
	resp, err := execChain.Do(fxcontext.New(context.Background(), service.NopHost()), _req)
	assert.NoError(t, err)
	assert.Equal(t, _respOK, resp)
}

func TestExecutionChainFilters(t *testing.T) {
	execChain := newExecutionChain(
		[]Filter{tracingFilter()}, getNopClient(),
	)
	ctx := fx.NopContext
	resp, err := execChain.Do(ctx, _req)
	assert.NoError(t, err)
	assert.Equal(t, _respOK, resp)
}

func TestExecutionChainFiltersError(t *testing.T) {
	execChain := newExecutionChain(
		[]Filter{tracingFilter()}, getErrorClient(),
	)
	resp, err := execChain.Do(fx.NopContext, _req)
	assert.Error(t, err)
	assert.Equal(t, errClient, err)
	assert.Nil(t, resp)
}

func withOpentracingSetup(t *testing.T, registerFunc auth.RegisterFunc, fn func(tracer opentracing.Tracer)) {
	tracer, closer, err := tracing.InitGlobalTracer(&jconfig.Configuration{}, "Test", ulog.NopLogger, tally.NullStatsReporter)
	defer closer.Close()
	assert.NotNil(t, closer)
	require.NoError(t, err)

	_serviceName = "test_service"
	auth.UnregisterClient()
	defer auth.UnregisterClient()
	auth.RegisterClient(registerFunc)
	fn(tracer)
}

func TestExecutionChainFilters_AuthContextPropagation(t *testing.T) {
	withOpentracingSetup(t, nil, func(tracer opentracing.Tracer) {
		execChain := newExecutionChain(
			[]Filter{authenticationFilter(fakeAuthInfo{})}, getContextPropogationClient(t),
		)
		span := tracer.StartSpan("test_method")
		span.SetBaggageItem(auth.ServiceAuth, _serviceName)
		ctx := &fxcontext.Context{
			Context: opentracing.ContextWithSpan(context.Background(), span),
		}
		resp, err := execChain.Do(ctx, _req)
		assert.NoError(t, err)
		assert.Equal(t, _respOK, resp)
	})
}

func TestExecutionChainFilters_AuthContextPropagationFailure(t *testing.T) {
	withOpentracingSetup(t, auth.FakeFailureClient, func(tracer opentracing.Tracer) {
		execChain := newExecutionChain(
			[]Filter{authenticationFilter(fakeAuthInfo{})}, getContextPropogationClient(t),
		)
		span := tracer.StartSpan("test_method")
		span.SetBaggageItem(auth.ServiceAuth, _serviceName)
		ctx := &fxcontext.Context{
			Context: opentracing.ContextWithSpan(context.Background(), span),
		}
		resp, err := execChain.Do(ctx, _req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

type fakeAuthInfo struct {
	yaml []byte
}

func (f fakeAuthInfo) Config() config.Provider {
	return config.NewYAMLProviderFromBytes(f.yaml)
}

func (f fakeAuthInfo) Logger() ulog.Log {
	return ulog.NopLogger
}

func getContextPropogationClient(t *testing.T) BasicClient {
	return BasicClientFunc(
		func(ctx fx.Context, req *http.Request) (resp *http.Response, err error) {
			span := opentracing.SpanFromContext(ctx)
			assert.NotNil(t, span)
			assert.Equal(t, _serviceName, span.BaggageItem(auth.ServiceAuth))
			return _respOK, nil
		},
	)
}

func getNopClient() BasicClient {
	return BasicClientFunc(
		func(ctx fx.Context, req *http.Request) (resp *http.Response, err error) {
			return _respOK, nil
		},
	)
}

func getErrorClient() BasicClient {
	return BasicClientFunc(
		func(ctx fx.Context, req *http.Request) (resp *http.Response, err error) {
			return nil, errClient
		},
	)
}
