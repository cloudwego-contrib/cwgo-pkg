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

package slog

import (
	"context"
	"io"
	"log/slog"

	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"
)

const (
	LevelTrace  = slog.Level(-8)
	LevelNotice = slog.Level(2)
	LevelFatal  = slog.Level(12)
)

var _ logging.NewLogger = (*Logger)(nil)

func NewLogger(opts ...Option) *Logger {
	config := defaultConfig()

	// apply options
	for _, opt := range opts {
		opt.apply(config)
	}

	if !config.withLevel && config.withHandlerOptions && config.handlerOptions.Level != nil {
		lvl := &slog.LevelVar{}
		lvl.Set(config.handlerOptions.Level.Level())
		config.level = lvl
	}
	config.handlerOptions.Level = config.level

	var replaceAttrDefined bool
	if config.handlerOptions.ReplaceAttr == nil {
		replaceAttrDefined = false
	} else {
		replaceAttrDefined = true
	}

	replaceFun := config.handlerOptions.ReplaceAttr

	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.LevelKey {
			level := a.Value.Any().(slog.Level)
			switch level {
			case LevelTrace:
				a.Value = slog.StringValue("Trace")
			case slog.LevelDebug:
				a.Value = slog.StringValue("Debug")
			case slog.LevelInfo:
				a.Value = slog.StringValue("Info")
			case LevelNotice:
				a.Value = slog.StringValue("Notice")
			case slog.LevelWarn:
				a.Value = slog.StringValue("Warn")
			case slog.LevelError:
				a.Value = slog.StringValue("Error")
			case LevelFatal:
				a.Value = slog.StringValue("Fatal")
			default:
				a.Value = slog.StringValue("Warn")
			}
		}
		if replaceAttrDefined {
			return replaceFun(groups, a)
		} else {
			return a
		}
	}
	config.handlerOptions.ReplaceAttr = replaceAttr

	return &Logger{
		l:   slog.New(slog.NewJSONHandler(config.output, config.handlerOptions)),
		cfg: config,
	}
}

// Logger otelslog impl
type Logger struct {
	l   *slog.Logger
	cfg *config
}

func (l *Logger) CtxLog(level logging.Level, ctx context.Context, msg string, fields ...logging.CwField) {
	lvl := tranSLevel(level)
	if len(fields) >= 0 {
		l.l.Log(ctx, lvl, msg, convertToSlogFields(fields...)...)
	} else {
		l.l.Log(ctx, lvl, msg)
	}
}

func (l *Logger) Logw(level logging.Level, msg string, fields ...logging.CwField) {
	lvl := tranSLevel(level)
	if len(fields) >= 0 {
		l.l.Log(context.TODO(), lvl, msg, convertToSlogFields(fields...)...)
	} else {
		l.l.Log(context.TODO(), lvl, msg)
	}

}

func convertToSlogFields(fields ...logging.CwField) []any {
	var result []any
	for _, field := range fields {
		result = append(result, field.Key, field.Value)
	}
	return result
}

func (l *Logger) Logger() *slog.Logger {
	return l.l
}

func (l *Logger) SetLevel(level logging.Level) {
	lvl := tranSLevel(level)
	l.cfg.level.Set(lvl)
}

func (l *Logger) SetOutput(writer io.Writer) {
	l.cfg.output = writer
	l.l = slog.New(slog.NewJSONHandler(writer, l.cfg.handlerOptions))
}

func (l *Logger) SetLogger(log *slog.Logger) {
	l.l = log
}

func (l *Logger) GetHandler() *slog.HandlerOptions {
	return l.cfg.handlerOptions
}

func (l *Logger) GetOutput() io.Writer {
	return l.cfg.output
}
