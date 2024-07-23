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
	"context"
	"errors"
	"fmt"
	"github.com/kitex-contrib/obs-opentelemetry/logging"
	"io"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ logging.FullLogger = (*Logger)(nil)

// Ref to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/README.md#json-formats
const (
	traceIDKey    = "trace_id"
	spanIDKey     = "span_id"
	traceFlagsKey = "trace_flags"
)

type Logger struct {
	*zap.SugaredLogger
	config *config
}

func NewLogger(opts ...Option) *Logger {
	config := defaultConfig()

	// apply options
	for _, opt := range opts {
		opt.apply(config)
	}

	logger := zap.New(
		zapcore.NewCore(config.coreConfig.enc, config.coreConfig.ws, config.coreConfig.lvl),
		config.zapOpts...)

	return &Logger{
		SugaredLogger: logger.Sugar().With(config.customFields...),
		config:        config,
	}
}

// GetExtraKeys get extraKeys from logger config
func (l *Logger) GetExtraKeys() []ExtraKey {
	return l.config.extraKeys
}

// PutExtraKeys add extraKeys after init
func (l *Logger) PutExtraKeys(keys ...ExtraKey) {
	for _, k := range keys {
		if !inArray(k, l.config.extraKeys) {
			l.config.extraKeys = append(l.config.extraKeys, k)
		}
	}
}

func (l *Logger) Log(level logging.Level, kvs ...interface{}) {
	logger := l.With()

	switch level {
	case logging.LevelTrace, logging.LevelDebug:
		logger.Debug(kvs...)
	case logging.LevelInfo:
		logger.Info(kvs...)
	case logging.LevelNotice, logging.LevelWarn:
		logger.Warn(kvs...)
	case logging.LevelError:
		logger.Error(kvs...)
	case logging.LevelFatal:
		logger.Fatal(kvs...)
	default:
		logger.Warn(kvs...)
	}
}

func (l *Logger) Logf(level logging.Level, format string, kvs ...interface{}) {
	logger := l.With()

	switch level {
	case logging.LevelTrace, logging.LevelDebug:
		logger.Debugf(format, kvs...)
	case logging.LevelInfo:
		logger.Infof(format, kvs...)
	case logging.LevelNotice, logging.LevelWarn:
		logger.Warnf(format, kvs...)
	case logging.LevelError:
		logger.Errorf(format, kvs...)
	case logging.LevelFatal:
		logger.Fatalf(format, kvs...)
	default:
		logger.Warnf(format, kvs...)
	}
}

func (l *Logger) CtxLogf(level logging.Level, ctx context.Context, format string, kvs ...interface{}) {
	var zlevel zapcore.Level
	var sl *zap.SugaredLogger

	span := trace.SpanFromContext(ctx)
	var traceKVs []interface{}
	if span.SpanContext().TraceID().IsValid() {
		traceKVs = append(traceKVs, traceIDKey, span.SpanContext().TraceID())
	}
	if span.SpanContext().SpanID().IsValid() {
		traceKVs = append(traceKVs, spanIDKey, span.SpanContext().SpanID())
	}
	if span.SpanContext().TraceFlags().IsSampled() {
		traceKVs = append(traceKVs, traceFlagsKey, span.SpanContext().TraceFlags())
	}
	if len(traceKVs) > 0 {
		sl = l.With(traceKVs...)
	} else {
		sl = l.With()
	}

	if len(l.config.extraKeys) > 0 {
		for _, k := range l.config.extraKeys {
			if l.config.extraKeyAsStr {
				sl = sl.With(string(k), ctx.Value(string(k)))
			} else {
				sl = sl.With(string(k), ctx.Value(k))
			}
		}
	}

	switch level {
	case logging.LevelDebug, logging.LevelTrace:
		zlevel = zap.DebugLevel
		sl.Debugf(format, kvs...)
	case logging.LevelInfo:
		zlevel = zap.InfoLevel
		sl.Infof(format, kvs...)
	case logging.LevelNotice, logging.LevelWarn:
		zlevel = zap.WarnLevel
		sl.Warnf(format, kvs...)
	case logging.LevelError:
		zlevel = zap.ErrorLevel
		sl.Errorf(format, kvs...)
	case logging.LevelFatal:
		zlevel = zap.FatalLevel
		sl.Fatalf(format, kvs...)
	default:
		zlevel = zap.WarnLevel
		sl.Warnf(format, kvs...)
	}

	if !span.IsRecording() {
		return
	}

	// set span status
	if zlevel >= l.config.traceConfig.errorSpanLevel {
		msg := getMessage(format, kvs)
		span.SetStatus(codes.Error, "")
		span.RecordError(errors.New(msg), trace.WithStackTrace(l.config.traceConfig.recordStackTraceInSpan))
	}
}

func (l *Logger) Trace(v ...interface{}) {
	l.Log(logging.LevelTrace, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.Log(logging.LevelDebug, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.Log(logging.LevelInfo, v...)
}

func (l *Logger) Notice(v ...interface{}) {
	l.Log(logging.LevelNotice, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.Log(logging.LevelWarn, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.Log(logging.LevelError, v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.Log(logging.LevelFatal, v...)
}

func (l *Logger) Tracef(format string, v ...interface{}) {
	l.Logf(logging.LevelTrace, format, v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Logf(logging.LevelDebug, format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Logf(logging.LevelInfo, format, v...)
}

func (l *Logger) Noticef(format string, v ...interface{}) {
	l.Logf(logging.LevelInfo, format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Logf(logging.LevelWarn, format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Logf(logging.LevelError, format, v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Logf(logging.LevelFatal, format, v...)
}

func (l *Logger) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(logging.LevelDebug, ctx, format, v...)
}

func (l *Logger) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(logging.LevelDebug, ctx, format, v...)
}

func (l *Logger) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(logging.LevelInfo, ctx, format, v...)
}

func (l *Logger) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(logging.LevelWarn, ctx, format, v...)
}

func (l *Logger) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(logging.LevelWarn, ctx, format, v...)
}

func (l *Logger) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(logging.LevelError, ctx, format, v...)
}

func (l *Logger) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(logging.LevelFatal, ctx, format, v...)
}

func (l *Logger) SetLevel(level logging.Level) {
	var lvl zapcore.Level
	switch level {
	case logging.LevelTrace, logging.LevelDebug:
		lvl = zap.DebugLevel
	case logging.LevelInfo:
		lvl = zap.InfoLevel
	case logging.LevelWarn, logging.LevelNotice:
		lvl = zap.WarnLevel
	case logging.LevelError:
		lvl = zap.ErrorLevel
	case logging.LevelFatal:
		lvl = zap.FatalLevel
	default:
		lvl = zap.WarnLevel
	}
	l.config.coreConfig.lvl.SetLevel(lvl)
}

func (l *Logger) SetOutput(writer io.Writer) {
	ws := zapcore.AddSync(writer)
	log := zap.New(
		zapcore.NewCore(l.config.coreConfig.enc, ws, l.config.coreConfig.lvl),
		l.config.zapOpts...,
	)
	l.config.coreConfig.ws = ws
	l.SugaredLogger = log.Sugar().With(l.config.customFields...)
}

// Logger is used to return an instance of *zap.Logger for custom fields, etc.
func (l *Logger) Logger() *zap.Logger {
	return l.SugaredLogger.Desugar()
}

func (l *Logger) CtxKVLog(ctx context.Context, level logging.Level, format string, kvs ...interface{}) {
	if len(kvs) == 0 || len(kvs)%2 != 0 {
		l.Warn(fmt.Sprint("Keyvalues must appear in pairs:", kvs))
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().TraceID().IsValid() {
		kvs = append(kvs, traceIDKey, span.SpanContext().TraceID())
	}
	if span.SpanContext().SpanID().IsValid() {
		kvs = append(kvs, spanIDKey, span.SpanContext().SpanID())
	}
	if span.SpanContext().TraceFlags().IsSampled() {
		kvs = append(kvs, traceFlagsKey, span.SpanContext().TraceFlags())
	}

	var zlevel zapcore.Level
	zl := l.With()
	switch level {
	case logging.LevelDebug, logging.LevelTrace:
		zlevel = zap.DebugLevel
		zl.Debugw(format, kvs...)
	case logging.LevelInfo:
		zlevel = zap.InfoLevel
		zl.Infow(format, kvs...)
	case logging.LevelNotice, logging.LevelWarn:
		zlevel = zap.WarnLevel
		zl.Warnw(format, kvs...)
	case logging.LevelError:
		zlevel = zap.ErrorLevel
		zl.Errorw(format, kvs...)
	case logging.LevelFatal:
		zlevel = zap.FatalLevel
		zl.Fatalw(format, kvs...)
	default:
		zlevel = zap.WarnLevel
		zl.Warnw(format, kvs...)
	}

	if !span.IsRecording() {
		return
	}

	// set span status
	if zlevel >= l.config.traceConfig.errorSpanLevel {
		msg := getMessage(format, kvs)
		span.SetStatus(codes.Error, "")
		span.RecordError(errors.New(msg), trace.WithStackTrace(l.config.traceConfig.recordStackTraceInSpan))
	}
}
