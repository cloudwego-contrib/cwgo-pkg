// Copyright 2024 CloudWeGo Authors.
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

package otelzerolog

import (
	cwzerolog "github.com/cloudwego-contrib/cwgo-pkg/logging/zerolog"
)

type Logger struct {
	*cwzerolog.Logger
	config *config
}

// Ref to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/README.md#json-formats
const (
	traceIDKey    = "trace_id"
	spanIDKey     = "span_id"
	traceFlagsKey = "trace_flags"
)

func NewLogger(opts ...Option) *Logger {
	cfg := defaultConfig()

	// apply options
	for _, opt := range opts {
		opt.apply(cfg)
	}
	logger := *cfg.logger
	if cfg.zeroLogger != nil {
		logger = *cwzerolog.From(*cfg.zeroLogger)
	}

	zerologLogger := logger.Unwrap().
		Hook(cfg.defaultZerologHookFn())

	return &Logger{
		Logger: cwzerolog.From(zerologLogger),
		config: cfg,
	}
}
