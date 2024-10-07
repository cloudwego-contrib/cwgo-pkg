// Copyright 2023 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelslog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"
	cwslog "github.com/cloudwego-contrib/cwgo-pkg/log/logging/slog"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func stdoutProvider(ctx context.Context) func() {
	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)

	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		panic(err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	provider.RegisterSpanProcessor(bsp)

	return func() {
		if err := provider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}
}

func TestLogger(t *testing.T) {
	ctx := context.Background()

	buf := new(bytes.Buffer)

	shutdown := stdoutProvider(ctx)
	defer shutdown()

	logger := NewLogger(
		WithTraceErrorSpanLevel(slog.LevelWarn),
		WithRecordStackTraceInSpan(true),
	)

	logging.SetLogger(logger)
	logging.SetOutput(buf)
	logging.SetLevel(logging.LevelDebug)

	logger.Logw(logging.LevelInfo, "log from origin otelslog")
	assert.True(t, strings.Contains(buf.String(), "log from origin otelslog"))
	buf.Reset()

	tracer := otel.Tracer("test otel std logger")

	ctx, span := tracer.Start(ctx, "root")

	logging.CtxInfof(ctx, "hello %s", "you")
	assert.True(t, strings.Contains(buf.String(), "trace_id"))
	assert.True(t, strings.Contains(buf.String(), "span_id"))
	assert.True(t, strings.Contains(buf.String(), "trace_flags"))

	buf.Reset()

	span.End()

	ctx, child := tracer.Start(ctx, "child")

	logging.CtxWarnf(ctx, "foo %s", "bar")

	logging.CtxTracef(ctx, "trace %s", "this is a trace log")
	logging.CtxDebugf(ctx, "debug %s", "this is a debug log")
	logging.CtxInfof(ctx, "info %s", "this is a info log")
	logging.CtxNoticef(ctx, "notice %s", "this is a notice log")
	logging.CtxWarnf(ctx, "warn %s", "this is a warn log")
	logging.CtxErrorf(ctx, "error %s", "this is a error log")

	child.End()

	_, errSpan := tracer.Start(ctx, "error")

	logging.Info("no trace context")

	errSpan.End()
}

func TestLogLevel(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewLogger(
		WithTraceErrorSpanLevel(slog.LevelWarn),
		WithRecordStackTraceInSpan(true),
	)

	// output to buffer
	logger.SetOutput(buf)
	logging.SetLogger(logger)
	logging.Debug("this is a debug log")
	assert.False(t, strings.Contains(buf.String(), "this is a debug log"))

	logger.SetLevel(logging.LevelDebug)

	logging.Debugf("this is a debug log %s", "msg")

	assert.True(t, strings.Contains(buf.String(), "this is a debug log"))
}

func TestLogOption(t *testing.T) {
	buf := new(bytes.Buffer)

	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)
	logger := NewLogger(
		WithLogger(
			cwslog.NewLogger(
				cwslog.WithLevel(lvl),
				cwslog.WithOutput(buf),
				cwslog.WithHandlerOptions(&slog.HandlerOptions{
					AddSource: false,
					ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
						if a.Key == slog.MessageKey {
							msg := a.Value.Any().(string)
							msg = strings.ReplaceAll(msg, "log", "new log")
							a.Value = slog.StringValue(msg)
						}
						return a
					},
				}),
			)),
		WithTraceErrorSpanLevel(slog.LevelWarn),
		WithRecordStackTraceInSpan(true),
	)
	logging.SetLogger(logger)
	logging.Debug("this is a debug log")
	assert.True(t, strings.Contains(buf.String(), "this is a debug new log"))
}
