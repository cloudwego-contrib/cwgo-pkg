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

package otellogrus

import (
	"github.com/sirupsen/logrus"
)

// Option otellogrus hook option
type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	logger *logrus.Logger
	hooks  []logrus.Hook

	traceHookConfig *TraceHookConfig
}

func defaultConfig() *config {
	// std logger
	stdLogger := logrus.StandardLogger()
	// default json format
	stdLogger.SetFormatter(new(logrus.JSONFormatter))

	return &config{
		logger: logrus.StandardLogger(),
		hooks:  []logrus.Hook{},
		traceHookConfig: &TraceHookConfig{
			recordStackTraceInSpan: true,
			enableLevels:           logrus.AllLevels,
			errorSpanLevel:         logrus.ErrorLevel,
		},
	}
}

// WithLogger configures logger
func WithLogger(logger *logrus.Logger) Option {
	return option(func(cfg *config) {
		cfg.logger = logger
	})
}

// WithHook configures hook
func WithHook(hook logrus.Hook) Option {
	return option(func(cfg *config) {
		cfg.hooks = append(cfg.hooks, hook)
	})
}

// WithTraceHookConfig configures trace hook config
func WithTraceHookConfig(hookConfig *TraceHookConfig) Option {
	return option(func(cfg *config) {
		cfg.traceHookConfig = hookConfig
	})
}

// WithTraceHookLevels configures hook levels
func WithTraceHookLevels(levels []logrus.Level) Option {
	return option(func(cfg *config) {
		cfg.traceHookConfig.enableLevels = levels
	})
}

// WithTraceHookErrorSpanLevel configures trace hook error span level
func WithTraceHookErrorSpanLevel(level logrus.Level) Option {
	return option(func(cfg *config) {
		cfg.traceHookConfig.errorSpanLevel = level
	})
}

// WithRecordStackTraceInSpan configures whether record stack trace in span
func WithRecordStackTraceInSpan(recordStackTraceInSpan bool) Option {
	return option(func(cfg *config) {
		cfg.traceHookConfig.recordStackTraceInSpan = recordStackTraceInSpan
	})
}
