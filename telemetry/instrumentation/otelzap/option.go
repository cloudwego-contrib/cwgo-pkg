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
	cwzap "github.com/cloudwego-contrib/cwgo-pkg/logging/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
type cwZap struct {
	customFields []interface{}
	extraKeys    []cwzap.ExtraKey
	coreConfig   cwzap.CoreConfig
	zapOpts      []zap.Option
}

type config struct {
	logger      *cwzap.Logger
	traceConfig *traceConfig
	cwZap       cwZap
	hasCwZap    bool
}

// defaultConfig default config
func defaultConfig() *config {
	return &config{
		traceConfig: &traceConfig{
			recordStackTraceInSpan: true,
			errorSpanLevel:         zapcore.ErrorLevel,
		},
		logger: cwzap.NewLogger(),
	}
}

// WithCoreEnc zapcore encoder
func WithCoreEnc(enc zapcore.Encoder) Option {
	return option(func(cfg *config) {
		cfg.cwZap.coreConfig.Enc = enc
		cfg.hasCwZap = true
	})
}

// WithCoreWs zapcore write syncer
func WithCoreWs(ws zapcore.WriteSyncer) Option {
	return option(func(cfg *config) {
		cfg.cwZap.coreConfig.Ws = ws
		cfg.hasCwZap = true
	})
}

// WithCoreLevel zapcore log level
func WithCoreLevel(lvl zap.AtomicLevel) Option {
	return option(func(cfg *config) {
		cfg.cwZap.coreConfig.Lvl = lvl
		cfg.hasCwZap = true
	})
}

// WithCustomFields record log with the key-value pair.
func WithCustomFields(kv ...interface{}) Option {
	return option(func(cfg *config) {
		cfg.cwZap.customFields = append(cfg.cwZap.customFields, kv...)
		cfg.hasCwZap = true
	})
}

// WithZapOptions add origin zap option
func WithZapOptions(opts ...zap.Option) Option {
	return option(func(cfg *config) {
		cfg.cwZap.zapOpts = append(cfg.cwZap.zapOpts, opts...)
		cfg.hasCwZap = true
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
