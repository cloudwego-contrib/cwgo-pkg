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
	"github.com/cloudwego-contrib/obs-opentelemetry/logging"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ logging.FullLogger = (*Logger)(nil)

type Logger struct {
	l      *zap.Logger
	config *config
}

func NewLogger(opts ...Option) *Logger {
	config := defaultConfig()

	// apply options
	for _, opt := range opts {
		opt.apply(config)
	}

	cores := make([]zapcore.Core, 0, len(config.coreConfigs))
	for _, coreConfig := range config.coreConfigs {
		cores = append(cores, zapcore.NewCore(coreConfig.Enc, coreConfig.Ws, coreConfig.Lvl))
	}

	logger := zap.New(
		zapcore.NewTee(cores[:]...),
		config.zapOpts...)

	return &Logger{
		l:      logger,
		config: config,
	}
}

// GetExtraKeys get extraKeys from logger config
func (l *Logger) GetExtraKeys() []ExtraKey {
	return l.config.extraKeys
}

// PutExtraKeys add extraKeys after init
func (l *Logger) PutExtraKeys(keys ...ExtraKey) {
	for _, k := range keys {
		if !InArray(k, l.config.extraKeys) {
			l.config.extraKeys = append(l.config.extraKeys, k)
		}
	}
}

func (l *Logger) Log(level logging.Level, kvs ...interface{}) {
	sugar := l.l.Sugar()
	switch level {
	case logging.LevelTrace, logging.LevelDebug:
		sugar.Debug(kvs...)
	case logging.LevelInfo:
		sugar.Info(kvs...)
	case logging.LevelNotice, logging.LevelWarn:
		sugar.Warn(kvs...)
	case logging.LevelError:
		sugar.Error(kvs...)
	case logging.LevelFatal:
		sugar.Fatal(kvs...)
	default:
		sugar.Warn(kvs...)
	}
}

func (l *Logger) Logf(level logging.Level, format string, kvs ...interface{}) {
	logger := l.l.Sugar().With()
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
	zLevel := LevelToZapLevel(level)
	if !l.config.coreConfigs[0].Lvl.Enabled(zLevel) {
		return
	}
	zapLogger := l.l
	if len(l.config.extraKeys) > 0 {
		for _, k := range l.config.extraKeys {
			if l.config.extraKeyAsStr {
				v := ctx.Value(string(k))
				if v != nil {
					zapLogger = zapLogger.With(zap.Any(string(k), v))
				}
			} else {
				v := ctx.Value(k)
				if v != nil {
					zapLogger = zapLogger.With(zap.Any(string(k), v))
				}
			}
		}
	}
	log := zapLogger.Sugar()
	switch level {
	case logging.LevelDebug, logging.LevelTrace:
		log.Debugf(format, kvs...)
	case logging.LevelInfo:
		log.Infof(format, kvs...)
	case logging.LevelNotice, logging.LevelWarn:
		log.Warnf(format, kvs...)
	case logging.LevelError:
		log.Errorf(format, kvs...)
	case logging.LevelFatal:
		log.Fatalf(format, kvs...)
	default:
		log.Warnf(format, kvs...)
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
	l.Logf(logging.LevelWarn, format, v...)
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
	lvl := LevelToZapLevel(level)

	l.config.coreConfigs[0].Lvl = lvl

	cores := make([]zapcore.Core, 0, len(l.config.coreConfigs))
	for _, coreConfig := range l.config.coreConfigs {
		cores = append(cores, zapcore.NewCore(coreConfig.Enc, coreConfig.Ws, coreConfig.Lvl))
	}

	logger := zap.New(
		zapcore.NewTee(cores[:]...),
		l.config.zapOpts...)

	l.l = logger
}

func (l *Logger) SetOutput(writer io.Writer) {
	l.config.coreConfigs[0].Ws = zapcore.AddSync(writer)

	cores := make([]zapcore.Core, 0, len(l.config.coreConfigs))
	for _, coreConfig := range l.config.coreConfigs {
		cores = append(cores, zapcore.NewCore(coreConfig.Enc, coreConfig.Ws, coreConfig.Lvl))
	}

	logger := zap.New(
		zapcore.NewTee(cores[:]...),
		l.config.zapOpts...)

	l.l = logger
}

// Logger is used to return an instance of *zap.Logger for custom fields, etc.
func (l *Logger) Logger() *zap.Logger {
	return l.l
}

func (l *Logger) Sync() {
	_ = l.l.Sync()
}
