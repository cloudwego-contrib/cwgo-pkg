// Copyright 2023 CloudWeGo Authors.
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

package otelslog

import (
	"io"
	"log/slog"

	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"
	cwslog "github.com/cloudwego-contrib/cwgo-pkg/log/logging/slog"
)

const (
	LevelTrace  = slog.Level(-8)
	LevelNotice = slog.Level(2)
	LevelFatal  = slog.Level(12)
)

type Logger struct {
	cwslog.Logger
	config *config
}

func NewLogger(opts ...Option) *Logger {
	cfg := defaultConfig()

	for _, opt := range opts {
		opt.apply(cfg)
	}
	logger := &Logger{
		Logger: *cfg.logger,
		config: cfg,
	}
	logger.setTraceLogger()
	return logger
}

func (l *Logger) setTraceLogger() {
	log := slog.New(NewTraceHandler(l.GetOutput(), l.config.logger.GetHandler(), l.config.traceConfig))
	l.Logger.SetLogger(log)
}

func (l *Logger) SetOutput(writer io.Writer) {
	log := slog.New(NewTraceHandler(writer, l.config.logger.GetHandler(), l.config.traceConfig))
	l.config.logger.SetOutput(writer)
	l.Logger.SetLogger(log)
}

func (l *Logger) SetLevel(level logging.Level) {
	l.Logger.SetLevel(level)
}
