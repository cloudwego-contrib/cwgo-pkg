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
	"io"

	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ logging.NewLogger = (*Logger)(nil)

type Logger struct {
	l      *zap.Logger
	config *config
}

func (l *Logger) CtxLog(level logging.Level, ctx context.Context, msg string, fields ...logging.CwField) {
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
	if len(fields) > 0 {
		zapLogger = zapLogger.With(convertToZapFields(fields...)...)
	}
	log := zapLogger.Sugar()
	// Compatible with previous usage methods
	if l.config.customFields != nil {
		log = log.With(l.config.customFields...)
	}

	switch level {
	case logging.LevelDebug, logging.LevelTrace:
		log.Debug(msg)
	case logging.LevelInfo:
		log.Info(msg)
	case logging.LevelNotice, logging.LevelWarn:
		log.Warn(msg)
	case logging.LevelError:
		log.Error(msg)
	case logging.LevelFatal:
		log.Fatal(msg)
	default:
		log.Warn(msg)
	}
}

func (l *Logger) Logw(level logging.Level, msg string, fields ...logging.CwField) {
	zapLogger := l.l
	if len(fields) > 0 {
		zapLogger = zapLogger.With(convertToZapFields(fields...)...)
	}
	sugar := zapLogger.Sugar()

	// Compatible with previous usage methods
	if l.config.customFields != nil {
		sugar = sugar.With(l.config.customFields...)
	}
	switch level {
	case logging.LevelTrace, logging.LevelDebug:
		sugar.Debug(msg)
	case logging.LevelInfo:
		sugar.Info(msg)
	case logging.LevelNotice, logging.LevelWarn:
		sugar.Warn(msg)
	case logging.LevelError:
		sugar.Error(msg)
	case logging.LevelFatal:
		sugar.Fatal(msg)
	default:
		sugar.Warn(msg)
	}
}

func convertToZapFields(fields ...logging.CwField) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value) // 你可以根据实际类型选择合适的 zap.Field 方法
	}
	return zapFields
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
