/*
 * Copyright 2024 CloudWeGo Authors
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

package zerolog

import (
	"errors"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rs/zerolog"
)

var _ klog.FullLogger = (*KLogger)(nil)

type KLogger struct {
	*Logger
}

func (l *KLogger) SetLevel(level klog.Level) {
	var lv hlog.Level
	switch level {
	case klog.LevelTrace:
		lv = hlog.LevelTrace
	case klog.LevelDebug:
		lv = hlog.LevelDebug
	case klog.LevelInfo:
		lv = hlog.LevelInfo
	case klog.LevelWarn:
		lv = hlog.LevelWarn
	case klog.LevelNotice:
		lv = hlog.LevelNotice
	case klog.LevelError:
		lv = hlog.LevelError
	case klog.LevelFatal:
		lv = hlog.LevelFatal
	default:
		lv = hlog.LevelWarn
	}
	l.Logger.SetLevel(lv)
}

func NewK(options ...Opt) *KLogger {
	return &KLogger{New(options...)}
}

// From returns a new Logger instance using an existing logger
func FromK(log zerolog.Logger, options ...Opt) *KLogger {
	return &KLogger{From(log, options...)}
}

func GetKLogger() (KLogger, error) {
	defaultLogger := klog.DefaultLogger()

	if logger, ok := defaultLogger.(*KLogger); ok {
		return *logger, nil
	}

	return KLogger{}, errors.New("klog.DefaultLogger is not a zerolog logger")
}
