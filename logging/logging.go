package logging

import (
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/context"
)

var logger *slog.Logger

func Setup(runMode string) {
	level := new(slog.LevelVar)
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}

	format := strings.ToLower(os.Getenv("LOG_FORMAT"))
	if format == "" {
		if runMode == "prod" {
			format = "json"
		} else {
			format = "text"
		}
	}

	var handler slog.Handler
	if format == "text" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	logger = slog.New(handler).With("runmode", runMode)
	slog.SetDefault(logger)
}

func Logger() *slog.Logger {
	if logger == nil {
		h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
		logger = slog.New(h)
		slog.SetDefault(logger)
	}
	return logger
}

func StartTimer(ctx *context.Context) {
	ctx.Input.SetData("_startTime", time.Now())
}

func LogRequest(ctx *context.Context) {
	start, _ := ctx.Input.GetData("_startTime").(time.Time)
	if start.IsZero() {
		start = time.Now()
	}
	dur := time.Since(start)
	status := ctx.ResponseWriter.Status
	method := ctx.Input.Method()
	path := ctx.Input.URL()
	ip := clientIP(ctx)
	Logger().Info("http_request",
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.String("ip", ip),
		slog.Duration("duration", dur),
	)
}

func LogIfError(ctx *context.Context) {
	status := ctx.ResponseWriter.Status
	if status < 400 {
		return
	}
	method := ctx.Input.Method()
	path := ctx.Input.URL()
	ip := clientIP(ctx)
	q := sanitizeQuery(ctx.Request.URL.Query())
	Logger().Error("http_error",
		slog.Int("status", status),
		slog.String("method", method),
		slog.String("path", path),
		slog.String("ip", ip),
		slog.Any("query", q),
	)
}

func LogControllerError(ctx *context.Context, message string, err error, fields map[string]interface{}) {
	sanitized := sanitizeFields(fields)
	method := ctx.Input.Method()
	path := ctx.Input.URL()
	ip := clientIP(ctx)
	q := sanitizeQuery(ctx.Request.URL.Query())
	Logger().Error(message,
		slog.String("error", safeErr(err)),
		slog.String("method", method),
		slog.String("path", path),
		slog.String("ip", ip),
		slog.Any("query", q),
		slog.Any("params", sanitized),
	)
}

func clientIP(ctx *context.Context) string {
	ip := ctx.Input.Header("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	return ctx.Input.IP()
}

func sanitizeQuery(v url.Values) map[string]string {
	out := make(map[string]string, len(v))
	for k, vals := range v {
		key := strings.ToLower(k)
		if key == "password" || strings.Contains(key, "pass") || key == "authorization" || strings.Contains(key, "token") {
			out[k] = "[redacted]"
			continue
		}
		if len(vals) > 0 {
			val := vals[0]
			if len(val) > 128 {
				val = val[:128] + "…"
			}
			out[k] = val
		}
	}
	return out
}

func sanitizeFields(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		lk := strings.ToLower(k)
		if lk == "password" || strings.Contains(lk, "pass") || strings.Contains(lk, "token") || lk == "authorization" || strings.Contains(lk, "secret") {
			out[k] = "[redacted]"
			continue
		}
		switch val := v.(type) {
		case string:
			if len(val) > 256 {
				out[k] = val[:256] + "…"
			} else {
				out[k] = val
			}
		default:
			out[k] = v
		}
	}
	return out
}

func safeErr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
