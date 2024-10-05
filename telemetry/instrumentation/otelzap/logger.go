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

package otelzap

import (
	"context"
	"errors"

	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"
	cwzap "github.com/cloudwego-contrib/cwgo-pkg/logging/zap"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	cwzap.Logger
	config *config
}

// Ref to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/README.md#json-formats
const (
	traceIDKey    = "trace_id"
	spanIDKey     = "span_id"
	traceFlagsKey = "trace_flags"
)

var extraKeys = []cwzap.ExtraKey{traceIDKey, spanIDKey, traceFlagsKey}

func NewLogger(opts ...Option) *Logger {
	config := defaultConfig()

	// apply options
	for _, opt := range opts {
		opt.apply(config)
	}
	logger := *config.logger
	if config.hasCwZap {
		options := GetOptions(config.cwZap)
		logger = *cwzap.NewLogger(options...)
		extraKeys = append(extraKeys, config.cwZap.extraKeys...)
	}
	logger.PutExtraKeys(extraKeys...)

	return &Logger{
		Logger: logger,
		config: config,
	}
}

func (l *Logger) CtxLog(level logging.Level, ctx context.Context, msg string, fields ...logging.CwField) {
	var zlevel zapcore.Level
	span := trace.SpanFromContext(ctx)

	if span.SpanContext().IsValid() {
		ctx = context.WithValue(ctx, cwzap.ExtraKey(traceIDKey), span.SpanContext().TraceID())
		ctx = context.WithValue(ctx, cwzap.ExtraKey(spanIDKey), span.SpanContext().SpanID())
		ctx = context.WithValue(ctx, cwzap.ExtraKey(traceFlagsKey), span.SpanContext().TraceFlags())

		l.Logger.CtxLog(level, ctx, msg, fields...)
	} else {
		l.Logger.Logw(level, msg, fields...)
	}

	if !span.IsRecording() {
		return
	}

	switch level {
	case logging.LevelDebug, logging.LevelTrace:
		zlevel = zap.DebugLevel
	case logging.LevelInfo:
		zlevel = zap.InfoLevel
	case logging.LevelNotice, logging.LevelWarn:
		zlevel = zap.WarnLevel
	case logging.LevelError:
		zlevel = zap.ErrorLevel
	case logging.LevelFatal:
		zlevel = zap.FatalLevel
	default:
		zlevel = zap.WarnLevel
	}

	// set span status
	if zlevel >= l.config.traceConfig.errorSpanLevel {
		span.SetStatus(codes.Error, "")
		span.RecordError(errors.New(msg), trace.WithStackTrace(l.config.traceConfig.recordStackTraceInSpan))
	}
}

func GetOptions(cwZap cwZap) []cwzap.Option {
	opions := []cwzap.Option{}
	opions = append(opions, cwzap.WithCores(cwzap.CoreConfig{
		Enc: cwZap.coreConfig.Enc,
		Lvl: cwZap.coreConfig.Lvl,
		Ws:  cwZap.coreConfig.Ws,
	}))
	opions = append(opions, cwzap.WithZapOptions(cwZap.zapOpts...))
	opions = append(opions, cwzap.WithCustomFields(cwZap.customFields))
	return opions
}
