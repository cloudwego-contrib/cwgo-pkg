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

package zap

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/cloudwego-contrib/obs-opentelemetry/logging"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

// testEncoderConfig encoder config for testing, copy from zap
func testEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "name",
		TimeKey:        "ts",
		CallerKey:      "caller",
		FunctionKey:    "func",
		StacktraceKey:  "stacktrace",
		LineEnding:     "\n",
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// humanEncoderConfig copy from zap
func humanEncoderConfig() zapcore.EncoderConfig {
	cfg := testEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncodeDuration = zapcore.StringDurationEncoder
	return cfg
}

// TestLogger test logger work with opentelemetry
func TestLogger(t *testing.T) {
	ctx := context.Background()

	buf := new(bytes.Buffer)

	shutdown := stdoutProvider(ctx)
	defer shutdown()

	logger := NewLogger(
		WithTraceErrorSpanLevel(zap.WarnLevel),
		WithRecordStackTraceInSpan(true),
	)
	defer func() {
		err := logger.Sync()
		if err != nil {
			return
		}
	}()

	logging.SetLogger(logger)
	logging.SetOutput(buf)
	logging.SetLevel(logging.LevelDebug)

	logger.Info("log from origin zap")
	assert.True(t, strings.Contains(buf.String(), "log from origin zap"))
	buf.Reset()

	tracer := otel.Tracer("test otel std logger")

	ctx, span := tracer.Start(ctx, "root")

	logging.CtxInfof(ctx, "hello %s", "world")
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

// TestLogLevel test SetLevel
func TestLogLevel(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewLogger(
		WithTraceErrorSpanLevel(zap.WarnLevel),
		WithRecordStackTraceInSpan(true),
	)
	defer func() {
		err := logger.Sync()
		if err != nil {
			return
		}
	}()

	// output to buffer
	logger.SetOutput(buf)

	logger.Debug("this is a debug log")
	assert.False(t, strings.Contains(buf.String(), "this is a debug log"))

	logger.SetLevel(logging.LevelDebug)

	logger.Debugf("this is a debug log %s", "msg")
	assert.True(t, strings.Contains(buf.String(), "this is a debug log"))
}

// TestCoreOption test zapcore config option
func TestCoreOption(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewLogger(
		WithCoreEnc(zapcore.NewConsoleEncoder(humanEncoderConfig())),
		WithCoreLevel(zap.NewAtomicLevelAt(zapcore.WarnLevel)),
		WithCoreWs(zapcore.AddSync(buf)),
	)
	defer func() {
		err := logger.Sync()
		if err != nil {
			return
		}
	}()

	logger.SetOutput(buf)

	logger.Debug("this is a debug log")
	// test log level
	assert.False(t, strings.Contains(buf.String(), "this is a debug log"))

	logger.Error("this is a warn log")
	// test log level
	assert.True(t, strings.Contains(buf.String(), "this is a warn log"))
	// test console encoder result
	assert.True(t, strings.Contains(buf.String(), "\tERROR\t"))
}

// TestCoreOption test zapcore config option
func TestZapOption(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewLogger(
		WithZapOptions(zap.AddCaller()),
	)
	defer func() {
		err := logger.Sync()
		if err != nil {
			return
		}
	}()

	logger.SetOutput(buf)

	logger.Debug("this is a debug log")
	assert.False(t, strings.Contains(buf.String(), "this is a debug log"))

	logger.Error("this is a warn log")
	// test caller in log result
	assert.True(t, strings.Contains(buf.String(), "caller"))
}

// TestCtxLogger test kv logger work with ctx
func TestCtxKVLogger(t *testing.T) {
	ctx := context.Background()

	buf := new(bytes.Buffer)

	shutdown := stdoutProvider(ctx)
	defer shutdown()

	logger := NewLogger(
		WithTraceErrorSpanLevel(zap.WarnLevel),
		WithRecordStackTraceInSpan(true),
	)
	defer func() {
		err := logger.Sync()
		if err != nil {
			return
		}
	}()

	logging.SetLogger(logger)
	logging.SetOutput(buf)
	logging.SetLevel(logging.LevelTrace)

	for _, level := range []logging.Level{
		logging.LevelTrace,
		logging.LevelDebug,
		logging.LevelInfo,
		logging.LevelNotice,
		logging.LevelWarn,
		logging.LevelError,
		// logging.LevelFatal,
	} {
		logger.CtxLogf(level, context.Background(), "log from origin zap %s=%s", "k1", "v1")
		println(buf.String())
		assert.True(t, strings.Contains(buf.String(), "log from origin zap"))
		assert.True(t, strings.Contains(buf.String(), "k1"))
		assert.True(t, strings.Contains(buf.String(), "v1"))
		buf.Reset()
	}

	for _, level := range []logging.Level{
		logging.LevelTrace,
		logging.LevelDebug,
		logging.LevelInfo,
		logging.LevelNotice,
		logging.LevelWarn,
		logging.LevelError,
		// logging.LevelFatal,
	} {
		logger.CtxKVLog(context.Background(), level, "log from origin zap", "k1", "v1")
		println(buf.String())
		assert.True(t, strings.Contains(buf.String(), "log from origin zap"))
		assert.True(t, strings.Contains(buf.String(), "k1"))
		assert.True(t, strings.Contains(buf.String(), "v1"))
		buf.Reset()
	}
}

// TestWithCustomFields test WithCustomFileds option.
func TestWithCustomFields(t *testing.T) {
	key := "service_name"
	value := "kitexTracing"
	buf := new(bytes.Buffer)

	t.Run("ctx info", func(t *testing.T) {
		buf.Reset()

		log := NewLogger(WithCustomFields(key, value))
		log.SetOutput(buf)

		ctx := context.Background()
		log.CtxInfof(ctx, "%s log", "extra")

		logStructMap := make(map[string]interface{}, 0)

		err := json.Unmarshal(buf.Bytes(), &logStructMap)
		assert.Nil(t, err)

		ret, ok := logStructMap[key]
		assert.True(t, ok)
		assert.Equal(t, value, ret)

		ret, ok = logStructMap["msg"]
		assert.True(t, ok)
		assert.Equal(t, "extra log", ret)
	})

	t.Run("infof", func(t *testing.T) {
		buf.Reset()

		log := NewLogger(WithCustomFields(key, value))
		log.SetOutput(buf)

		log.Infof("%s log", "extra")

		logStructMap := make(map[string]interface{}, 0)

		err := json.Unmarshal(buf.Bytes(), &logStructMap)
		assert.Nil(t, err)

		ret, ok := logStructMap[key]
		assert.True(t, ok)
		assert.Equal(t, value, ret)

		ret, ok = logStructMap["msg"]
		assert.True(t, ok)
		assert.Equal(t, "extra log", ret)
	})

	t.Run("info", func(t *testing.T) {
		buf.Reset()

		log := NewLogger(WithCustomFields(key, value))
		log.SetOutput(buf)

		log.Info("extra log")

		logStructMap := make(map[string]interface{}, 0)

		err := json.Unmarshal(buf.Bytes(), &logStructMap)
		assert.Nil(t, err)

		ret, ok := logStructMap[key]
		assert.True(t, ok)
		assert.Equal(t, value, ret)

		ret, ok = logStructMap["msg"]
		assert.True(t, ok)
		assert.Equal(t, "extra log", ret)
	})
}

// TestWithExtraKeys test WithExtraKeys option
func TestWithExtraKeys(t *testing.T) {
	buf := new(bytes.Buffer)

	log := NewLogger(WithExtraKeys([]ExtraKey{"requestId"}))
	log.SetOutput(buf)

	ctx := context.WithValue(context.Background(), ExtraKey("requestId"), "123")

	log.CtxInfof(ctx, "%s log", "extra")

	var logStructMap map[string]interface{}

	err := json.Unmarshal(buf.Bytes(), &logStructMap)

	assert.Nil(t, err)

	value, ok := logStructMap["requestId"]

	assert.True(t, ok)
	assert.Equal(t, value, "123")
}

func TestPutExtraKeys(t *testing.T) {
	logger := NewLogger(WithExtraKeys([]ExtraKey{"abc"}))

	assert.Contains(t, logger.GetExtraKeys(), ExtraKey("abc"))
	assert.NotContains(t, logger.GetExtraKeys(), ExtraKey("def"))

	logger.PutExtraKeys("def")
	assert.Contains(t, logger.GetExtraKeys(), ExtraKey("def"))
}

func TestExtraKeyAsStr(t *testing.T) {
	buf := new(bytes.Buffer)
	const v = "value"

	logger := NewLogger(WithExtraKeys([]ExtraKey{"abc"}))

	logger.SetOutput(buf)

	ctx1 := context.TODO()
	ctx1 = context.WithValue(ctx1, "key1", v) //nolint:staticcheck
	logger.CtxErrorf(ctx1, "%s", "error")

	assert.NotContains(t, buf.String(), v)

	buf.Reset()

	strLogger := NewLogger(WithExtraKeys([]ExtraKey{"abc"}), WithExtraKeyAsStr())

	strLogger.SetOutput(buf)

	ctx2 := context.TODO()
	ctx2 = context.WithValue(ctx2, "key2", v) //nolint:staticcheck

	strLogger.CtxErrorf(ctx2, "key2", v)

	assert.Contains(t, buf.String(), v)

	buf.Reset()
}
