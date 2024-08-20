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
	"log/slog"
	"strings"

	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"
)

// OtelSeverityText convert otelslog level to otel severityText
// ref to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#severity-fields
func OtelSeverityText(lv slog.Level) string {
	s := lv.String()
	if s == "warning" {
		s = "warn"
	}
	return strings.ToUpper(s)
}

// TranSLevel Adapt klog level to teleology level
func TranSLevel(level logging.Level) (lvl slog.Level) {
	switch level {
	case logging.LevelTrace:
		lvl = LevelTrace
	case logging.LevelDebug:
		lvl = slog.LevelDebug
	case logging.LevelInfo:
		lvl = slog.LevelInfo
	case logging.LevelWarn:
		lvl = slog.LevelWarn
	case logging.LevelNotice:
		lvl = LevelNotice
	case logging.LevelError:
		lvl = slog.LevelError
	case logging.LevelFatal:
		lvl = LevelFatal
	default:
		lvl = slog.LevelWarn
	}
	return
}
