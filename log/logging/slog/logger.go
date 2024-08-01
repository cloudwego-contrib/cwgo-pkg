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
	"fmt"
	"github.com/cloudwego-contrib/obs-opentelemetry/log/logging"
	"io"
	"log/slog"
)

const (
	LevelTrace  = slog.Level(-8)
	LevelNotice = slog.Level(2)
	LevelFatal  = slog.Level(12)
)

var _ logging.FullLogger = (*Logger)(nil)

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

// Logger slog impl
type Logger struct {
	l   *slog.Logger
	cfg *config
}

func (l *Logger) Logger() *slog.Logger {
	return l.l
}

func (l *Logger) log(level logging.Level, v ...any) {
	lvl := tranSLevel(level)
	l.l.Log(context.TODO(), lvl, fmt.Sprint(v...))
}

func (l *Logger) logf(level logging.Level, format string, kvs ...any) {
	lvl := tranSLevel(level)
	l.l.Log(context.TODO(), lvl, fmt.Sprintf(format, kvs...))
}

func (l *Logger) ctxLogf(level logging.Level, ctx context.Context, format string, v ...any) {
	lvl := tranSLevel(level)
	l.l.Log(ctx, lvl, fmt.Sprintf(format, v...))
}

func (l *Logger) Trace(v ...any) {
	l.log(logging.LevelTrace, v...)
}

func (l *Logger) Debug(v ...any) {
	l.log(logging.LevelDebug, v...)
}

func (l *Logger) Info(v ...any) {
	l.log(logging.LevelInfo, v...)
}

func (l *Logger) Notice(v ...any) {
	l.log(logging.LevelNotice, v...)
}

func (l *Logger) Warn(v ...any) {
	l.log(logging.LevelWarn, v...)
}

func (l *Logger) Error(v ...any) {
	l.log(logging.LevelError, v...)
}

func (l *Logger) Fatal(v ...any) {
	l.log(logging.LevelFatal, v...)
}

func (l *Logger) Tracef(format string, v ...any) {
	l.logf(logging.LevelTrace, format, v...)
}

func (l *Logger) Debugf(format string, v ...any) {
	l.logf(logging.LevelDebug, format, v...)
}

func (l *Logger) Infof(format string, v ...any) {
	l.logf(logging.LevelInfo, format, v...)
}

func (l *Logger) Noticef(format string, v ...any) {
	l.logf(logging.LevelNotice, format, v...)
}

func (l *Logger) Warnf(format string, v ...any) {
	l.logf(logging.LevelWarn, format, v...)
}

func (l *Logger) Errorf(format string, v ...any) {
	l.logf(logging.LevelError, format, v...)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.logf(logging.LevelFatal, format, v...)
}

func (l *Logger) CtxTracef(ctx context.Context, format string, v ...any) {
	l.ctxLogf(logging.LevelDebug, ctx, format, v...)
}

func (l *Logger) CtxDebugf(ctx context.Context, format string, v ...any) {
	l.ctxLogf(logging.LevelDebug, ctx, format, v...)
}

func (l *Logger) CtxInfof(ctx context.Context, format string, v ...any) {
	l.ctxLogf(logging.LevelInfo, ctx, format, v...)
}

func (l *Logger) CtxNoticef(ctx context.Context, format string, v ...any) {
	l.ctxLogf(logging.LevelNotice, ctx, format, v...)
}

func (l *Logger) CtxWarnf(ctx context.Context, format string, v ...any) {
	l.ctxLogf(logging.LevelWarn, ctx, format, v...)
}

func (l *Logger) CtxErrorf(ctx context.Context, format string, v ...any) {
	l.ctxLogf(logging.LevelError, ctx, format, v...)
}

func (l *Logger) CtxFatalf(ctx context.Context, format string, v ...any) {
	l.ctxLogf(logging.LevelFatal, ctx, format, v...)
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
