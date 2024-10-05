/*
 * Copyright 2022 CloudWeGo Authors
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
 *
 * MIT License
 *
 * Copyright (c) 2019-present Fenny and Contributors
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.E SOFTWARE.
 *
 * This file may have been modified by CloudWeGo authors. All CloudWeGo
 * Modifications are Copyright 2022 CloudWeGo Authors.
 */

package logrus

import (
	"context"
	"io"

	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"

	"github.com/sirupsen/logrus"
)

var _ logging.NewLogger = (*Logger)(nil)

// Logger otellogrus impl
type Logger struct {
	l *logrus.Logger
}

func (l *Logger) CtxLog(level logging.Level, ctx context.Context, msg string, fields ...logging.CwField) {

	var lv logrus.Level

	switch level {
	case logging.LevelTrace:
		lv = logrus.TraceLevel
	case logging.LevelDebug:
		lv = logrus.DebugLevel
	case logging.LevelInfo:
		lv = logrus.InfoLevel
	case logging.LevelWarn, logging.LevelNotice:
		lv = logrus.WarnLevel
	case logging.LevelError:
		lv = logrus.ErrorLevel
	case logging.LevelFatal:
		lv = logrus.FatalLevel
	default:
		lv = logrus.WarnLevel
	}

	logrusFeilds := convertToLogrusFields(fields...)

	if len(logrusFeilds) > 0 {
		l.l.WithFields(logrusFeilds).WithContext(ctx).Log(lv, msg)
	} else {
		l.l.WithContext(ctx).Log(lv, msg)
	}
}

func (l *Logger) Logw(level logging.Level, msg string, fields ...logging.CwField) {

	var lv logrus.Level

	switch level {
	case logging.LevelTrace:
		lv = logrus.TraceLevel
	case logging.LevelDebug:
		lv = logrus.DebugLevel
	case logging.LevelInfo:
		lv = logrus.InfoLevel
	case logging.LevelWarn, logging.LevelNotice:
		lv = logrus.WarnLevel
	case logging.LevelError:
		lv = logrus.ErrorLevel
	case logging.LevelFatal:
		lv = logrus.FatalLevel
	default:
		lv = logrus.WarnLevel
	}

	logrusFeilds := convertToLogrusFields(fields...)

	if len(logrusFeilds) > 0 {
		l.l.WithFields(logrusFeilds).Log(lv, msg)
	} else {
		l.l.Log(lv, msg)
	}
}

func convertToLogrusFields(fields ...logging.CwField) logrus.Fields {
	logrusFields := make(logrus.Fields)
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}
	return logrusFields
}

// NewLogger create a logger
func NewLogger(opts ...Option) *Logger {
	cfg := defaultConfig()

	// apply options
	for _, opt := range opts {
		opt.apply(cfg)
	}

	// attach hook
	for _, hook := range cfg.hooks {
		cfg.logger.AddHook(hook)
	}

	return &Logger{
		l: cfg.logger,
	}
}

func (l *Logger) Logger() *logrus.Logger {
	return l.l
}

func (l *Logger) SetLevel(level logging.Level) {
	var lv logrus.Level
	switch level {
	case logging.LevelTrace:
		lv = logrus.TraceLevel
	case logging.LevelDebug:
		lv = logrus.DebugLevel
	case logging.LevelInfo:
		lv = logrus.InfoLevel
	case logging.LevelWarn, logging.LevelNotice:
		lv = logrus.WarnLevel
	case logging.LevelError:
		lv = logrus.ErrorLevel
	case logging.LevelFatal:
		lv = logrus.FatalLevel
	default:
		lv = logrus.WarnLevel
	}
	l.l.SetLevel(lv)
}

func (l *Logger) SetOutput(writer io.Writer) {
	l.l.SetOutput(writer)
}
