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
	cwzap "github.com/cloudwego-contrib/cwgo-pkg/log/logging/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ExtraKey = cwzap.ExtraKey

type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type traceConfig struct {
	recordStackTraceInSpan bool
	errorSpanLevel         zapcore.Level
}

// cwZap is for compatibility with Kitex otel log

type config struct {
	logger       *cwzap.Logger
	traceConfig  *traceConfig
	options      []cwzap.Option
	customFields []interface{}
}

// defaultConfig default config
func defaultConfig() *config {
	return &config{
		traceConfig: &traceConfig{
			recordStackTraceInSpan: true,
			errorSpanLevel:         zapcore.ErrorLevel,
		},
	}
}

// WithCoreEnc zapcore encoder
func WithCoreEnc(enc zapcore.Encoder) Option {
	return option(func(cfg *config) {
		cfg.options = append(cfg.options, cwzap.WithCoreEnc(enc))
	})
}

// WithCoreWs zapcore write syncer
func WithCoreWs(ws zapcore.WriteSyncer) Option {
	return option(func(cfg *config) {
		cfg.options = append(cfg.options, cwzap.WithCoreWs(ws))
	})
}

// WithCoreLevel zapcore log level
func WithCoreLevel(lvl zap.AtomicLevel) Option {
	return option(func(cfg *config) {
		cfg.options = append(cfg.options, cwzap.WithCoreLevel(lvl))
	})
}

// WithCustomFields record log with the key-value pair.
func WithCustomFields(kv ...interface{}) Option {
	return option(func(cfg *config) {
		cfg.customFields = append(cfg.customFields, kv...)
	})
}

// WithZapOptions add origin zap option
func WithZapOptions(opts ...zap.Option) Option {
	return option(func(cfg *config) {
		cfg.options = append(cfg.options, cwzap.WithZapOptions(opts...))
	})
}

// WithLogger configures logger
func WithLogger(logger *cwzap.Logger) Option {
	return option(func(cfg *config) {
		logger.PutExtraKeys(extraKeys...)
		cfg.logger = logger
	})
}

// WithTraceErrorSpanLevel trace error span level option
func WithTraceErrorSpanLevel(level zapcore.Level) Option {
	return option(func(cfg *config) {
		cfg.traceConfig.errorSpanLevel = level
	})
}

// WithRecordStackTraceInSpan record stack track option
func WithRecordStackTraceInSpan(recordStackTraceInSpan bool) Option {
	return option(func(cfg *config) {
		cfg.traceConfig.recordStackTraceInSpan = recordStackTraceInSpan
	})
}
func WithExtraKeys(keys []ExtraKey) Option {
	return option(func(cfg *config) {
		cfg.options = append(cfg.options, cwzap.WithExtraKeys(keys))
	})
}

func WithExtraKeyAsStr() Option {
	return option(func(cfg *config) {
		cfg.options = append(cfg.options, cwzap.WithExtraKeyAsStr())
	})
}
