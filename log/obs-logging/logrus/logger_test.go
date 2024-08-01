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
	logging2 "github.com/cloudwego-contrib/obs-opentelemetry/log/logging"
	logrus2 "github.com/cloudwego-contrib/obs-opentelemetry/log/obs-logging/logrus"
	"testing"

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

	logger := logrus2.NewLogger(
		logrus2.WithTraceHookErrorSpanLevel(logrus.WarnLevel),
		logrus2.WithTraceHookLevels(logrus.AllLevels),
		logrus2.WithRecordStackTraceInSpan(true),
	)

	logger.Logger().Info("log from origin logrus")

	logging2.SetLogger(logger)
	logging2.SetLevel(logging2.LevelDebug)

	tracer := otel.Tracer("test otel std logger")
	ctx, span := tracer.Start(ctx, "root")

	logging2.SetLogger(logger)
	logging2.SetLevel(logging2.LevelTrace)

	logging2.Trace("trace")
	logging2.Debug("debug")
	logging2.Info("info")
	logging2.Notice("notice")
	logging2.Warn("warn")
	logging2.Error("error")

	logging2.Tracef("log level: %s", "trace")
	logging2.Debugf("log level: %s", "debug")
	logging2.Infof("log level: %s", "info")
	logging2.Noticef("log level: %s", "notice")
	logging2.Warnf("log level: %s", "warn")
	logging2.Errorf("log level: %s", "error")

	logging2.CtxTracef(ctx, "log level: %s", "trace")
	logging2.CtxDebugf(ctx, "log level: %s", "debug")
	logging2.CtxInfof(ctx, "log level: %s", "info")
	logging2.CtxNoticef(ctx, "log level: %s", "notice")
	logging2.CtxWarnf(ctx, "log level: %s", "warn")
	logging2.CtxErrorf(ctx, "log level: %s", "error")

	span.End()

	ctx, child := tracer.Start(ctx, "child")
	logging2.CtxWarnf(ctx, "foo %s", "bar")
	child.End()

	ctx, errSpan := tracer.Start(ctx, "error")
	logging2.CtxErrorf(ctx, "error %s", "this is a error")
	logging2.Info("no trace context")
	errSpan.End()
}
