// Copyright 2022 CloudWeGo Authors.
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

package logrus_test

import (
	"context"
	"github.com/cloudwego-contrib/obs-opentelemetry/logging"
	"testing"

	otellogrus "github.com/cloudwego-contrib/obs-opentelemetry/obs-logging/logrus"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func stdoutProvider(ctx context.Context) func() {
	provider := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(provider)

	exp, err := stdouttrace.New()
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
	shutdown := stdoutProvider(ctx)
	defer shutdown()

	logger := otellogrus.NewLogger(
		otellogrus.WithTraceHookErrorSpanLevel(logrus.WarnLevel),
		otellogrus.WithTraceHookLevels(logrus.AllLevels),
		otellogrus.WithRecordStackTraceInSpan(true),
	)

	logger.Logger().Info("log from origin logrus")

	logging.SetLogger(logger)
	logging.SetLevel(logging.LevelDebug)

	tracer := otel.Tracer("test otel std logger")
	ctx, span := tracer.Start(ctx, "root")

	logging.SetLogger(logger)
	logging.SetLevel(logging.LevelTrace)

	logging.Trace("trace")
	logging.Debug("debug")
	logging.Info("info")
	logging.Notice("notice")
	logging.Warn("warn")
	logging.Error("error")

	logging.Tracef("log level: %s", "trace")
	logging.Debugf("log level: %s", "debug")
	logging.Infof("log level: %s", "info")
	logging.Noticef("log level: %s", "notice")
	logging.Warnf("log level: %s", "warn")
	logging.Errorf("log level: %s", "error")

	logging.CtxTracef(ctx, "log level: %s", "trace")
	logging.CtxDebugf(ctx, "log level: %s", "debug")
	logging.CtxInfof(ctx, "log level: %s", "info")
	logging.CtxNoticef(ctx, "log level: %s", "notice")
	logging.CtxWarnf(ctx, "log level: %s", "warn")
	logging.CtxErrorf(ctx, "log level: %s", "error")

	span.End()

	ctx, child := tracer.Start(ctx, "child")
	logging.CtxWarnf(ctx, "foo %s", "bar")
	child.End()

	ctx, errSpan := tracer.Start(ctx, "error")
	logging.CtxErrorf(ctx, "error %s", "this is a error")
	logging.Info("no trace context")
	errSpan.End()
}
