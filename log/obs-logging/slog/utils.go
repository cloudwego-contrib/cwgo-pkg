package slog

import (
	"fmt"
	"github.com/cloudwego-contrib/obs-opentelemetry/log/logging"
	"log/slog"
	"strings"
)

// get format msg
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

// OtelSeverityText convert slog level to otel severityText
// ref to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#severity-fields
func OtelSeverityText(lv slog.Level) string {
	s := lv.String()
	if s == "warning" {
		s = "warn"
	}
	return strings.ToUpper(s)
}

// Adapt klog level to slog level
func tranSLevel(level logging.Level) (lvl slog.Level) {
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
