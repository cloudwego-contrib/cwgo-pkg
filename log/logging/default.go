/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logging

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

var logger = &CwLog{
	logger: &defaultLogger{
		level:  LevelInfo,
		stdlog: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile|log.Lmicroseconds),
	},
}

// SetOutput sets the output of default logger. By default, it is stderr.
func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}

// SetLevel sets the level of logs below which logs will not be output.
// The default log level is LevelTrace.
// Note that this method is not concurrent-safe.
func SetLevel(lv Level) {
	logger.SetLevel(lv)
}

// DefaultLogger return the default logger for Tracing.
func DefaultLogger() NewLogger {
	return logger.logger
}

// SetLogger sets the default logger.
// Note that this method is not concurrent-safe and must not be called
// after the use of DefaultLogger and global functions in this package.
func SetLogger(v NewLogger) {
	logger.SetLogger(v)
}

// Fatal calls the default logger's Fatal method and then os.Exit(1).
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Error calls the default logger's Error method.
func Error(v ...interface{}) {
	logger.Error(v...)
}

// Warn calls the default logger's Warn method.
func Warn(v ...interface{}) {
	logger.Warn(v...)
}

// Notice calls the default logger's Notice method.
func Notice(v ...interface{}) {
	logger.Notice(v...)
}

// Info calls the default logger's Info method.
func Info(v ...interface{}) {
	logger.Info(v...)
}

// Debug calls the default logger's Debug method.
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Trace calls the default logger's Trace method.
func Trace(v ...interface{}) {
	logger.Trace(v...)
}

// Fatalf calls the default logger's Fatalf method and then os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// Errorf calls the default logger's Errorf method.
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Warnf calls the default logger's Warnf method.
func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

// Noticef calls the default logger's Noticef method.
func Noticef(format string, v ...interface{}) {
	logger.Noticef(format, v...)
}

// Infof calls the default logger's Infof method.
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Debugf calls the default logger's Debugf method.
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Tracef calls the default logger's Tracef method.
func Tracef(format string, v ...interface{}) {
	logger.Tracef(format, v...)
}

// CtxFatalf calls the default logger's CtxFatalf method and then os.Exit(1).
func CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	logger.CtxFatalf(ctx, format, v...)
}

// CtxErrorf calls the default logger's CtxErrorf method.
func CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	logger.CtxErrorf(ctx, format, v...)
}

// CtxWarnf calls the default logger's CtxWarnf method.
func CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	logger.CtxWarnf(ctx, format, v...)
}

// CtxNoticef calls the default logger's CtxNoticef method.
func CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	logger.CtxNoticef(ctx, format, v...)
}

// CtxInfof calls the default logger's CtxInfof method.
func CtxInfof(ctx context.Context, format string, v ...interface{}) {
	logger.CtxInfof(ctx, format, v...)
}

// CtxDebugf calls the default logger's CtxDebugf method.
func CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	logger.CtxDebugf(ctx, format, v...)
}

// CtxTracef calls the default logger's CtxTracef method.
func CtxTracef(ctx context.Context, format string, v ...interface{}) {
	logger.CtxTracef(ctx, format, v...)
}

func Fatalw(msg string, fields ...CwField) {
	logger.Fatalw(msg, fields...)
}

func Errorw(msg string, fields ...CwField) {
	logger.Errorw(msg, fields...)
}

func Warnw(msg string, fields ...CwField) {
	logger.Warnw(msg, fields...)
}

func Noticew(msg string, fields ...CwField) {
	logger.Noticew(msg, fields...)
}

func Infow(msg string, fields ...CwField) {
	logger.Infow(msg, fields...)
}

func Debugw(msg string, fields ...CwField) {
	logger.Debugw(msg, fields...)
}

func Tracew(msg string, fields ...CwField) {
	logger.Tracew(msg, fields...)
}

func With(fields ...CwField) {
	logger.WithValue(fields...)
}

func ResetValue() {
	logger.ResetValue()
}

type defaultLogger struct {
	stdlog *log.Logger
	level  Level
}

func (d *defaultLogger) CtxLog(level Level, ctx context.Context, msg string, fields ...CwField) {
	if d.level > level {
		return
	}
	logMessage := level.toString() + "msg:" + msg

	for _, field := range fields {
		if field.Value != nil {
			logMessage += ", " + field.Key + ":" + fmt.Sprint(field.Value)
		} else {
			logMessage += ", " + field.Key
		}
	}

	d.stdlog.Output(4, logMessage)
	if level == LevelFatal {
		os.Exit(1)
	}
}

func (d *defaultLogger) Logw(level Level, msg string, fields ...CwField) {
	if d.level > level {
		return
	}
	logMessage := level.toString() + "msg:" + msg

	for _, field := range fields {
		if field.Value != nil {
			logMessage += ", " + field.Key + ":" + fmt.Sprint(field.Value)
		} else {
			logMessage += ", " + field.Key
		}
	}

	d.stdlog.Output(4, logMessage)
	if level == LevelFatal {
		os.Exit(1)
	}
}

func (d *defaultLogger) SetLevel(level Level) {
	d.level = level
}

func (d *defaultLogger) SetOutput(writer io.Writer) {
	d.stdlog.SetOutput(writer)
}
