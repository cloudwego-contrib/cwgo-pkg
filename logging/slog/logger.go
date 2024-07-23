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
	"github.com/cloudwego-contrib/obs-opentelemetry/logging"
	"io"
	"log/slog"
)

const (
	LevelTrace  = slog.Level(-8)
	LevelNotice = slog.Level(2)
	LevelFatal  = slog.Level(12)
)

// var _ klog.FullLogger = (*Logger)(nil)
var _ logging.FullLogger = (*Logger)(nil)

type Logger struct {
	l      *slog.Logger
	config *config
}

func NewLogger(opts ...Option) *Logger {
	config := defaultConfig()

	for _, opt := range opts {
		opt.apply(config)
	}
	// When user set the handlerOptions level but not set with coreconfig level
	if !config.coreConfig.withLevel && config.coreConfig.withHandlerOptions && config.coreConfig.opt.Level != nil {
		lvl := &slog.LevelVar{}
		lvl.Set(config.coreConfig.opt.Level.Level())
		config.coreConfig.level = lvl
	}
	config.coreConfig.opt.Level = config.coreConfig.level

	var replaceAttrDefined bool
	if config.coreConfig.opt.ReplaceAttr == nil {
		replaceAttrDefined = false
	} else {
		replaceAttrDefined = true
	}

	replaceFunc := config.coreConfig.opt.ReplaceAttr

	replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
		// default replaceAttr level
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
		// append replaceAttr by user
		if replaceAttrDefined {
			return replaceFunc(groups, a)
		} else {
			return a
		}
	}
	config.coreConfig.opt.ReplaceAttr = replaceAttr

	logger := slog.New(NewTraceHandler(config.coreConfig.writer, config.coreConfig.opt, config.traceConfig))
	return &Logger{
		l:      logger,
		config: config,
	}
}

func (l *Logger) Log(level logging.Level, msg string) {
	logger := l.l.With()
	logger.Log(context.TODO(), tranSLevel(level), msg)
}

func (l *Logger) Logf(level logging.Level, format string, kvs ...interface{}) {
	logger := l.l.With()
	msg := getMessage(format, kvs)
	logger.Log(context.TODO(), tranSLevel(level), msg)
}

func (l *Logger) CtxLogf(level logging.Level, ctx context.Context, format string, kvs ...interface{}) {
	logger := l.l.With()
	msg := getMessage(format, kvs)
	logger.Log(ctx, tranSLevel(level), msg)
}

func (l *Logger) Trace(v ...interface{}) {
	l.Log(logging.LevelTrace, fmt.Sprint(v...))
}

func (l *Logger) Debug(v ...interface{}) {
	l.Log(logging.LevelDebug, fmt.Sprint(v...))
}

func (l *Logger) Info(v ...interface{}) {
	l.Log(logging.LevelInfo, fmt.Sprint(v...))
}

func (l *Logger) Notice(v ...interface{}) {
	l.Log(logging.LevelNotice, fmt.Sprint(v...))
}

func (l *Logger) Warn(v ...interface{}) {
	l.Log(logging.LevelWarn, fmt.Sprint(v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.Log(logging.LevelError, fmt.Sprint(v...))
}

func (l *Logger) Fatal(v ...interface{}) {
	l.Log(logging.LevelFatal, fmt.Sprint(v...))
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
	l.Logf(logging.LevelNotice, format, v...)
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
	l.CtxLogf(logging.LevelTrace, ctx, format, v...)
}

func (l *Logger) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(logging.LevelDebug, ctx, format, v...)
}

func (l *Logger) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(logging.LevelInfo, ctx, format, v...)
}

func (l *Logger) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(logging.LevelNotice, ctx, format, v...)
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
	lvl := tranSLevel(level)
	l.config.coreConfig.level.Set(lvl)
}

func (l *Logger) SetOutput(writer io.Writer) {
	log := slog.New(NewTraceHandler(writer, l.config.coreConfig.opt, l.config.traceConfig))
	l.config.coreConfig.writer = writer
	l.l = log
}
