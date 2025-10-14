package cors

import (
	"net/http"
	"os"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func CORS() func(*context.Context) {
	defaultAllowedOrigins := []string{
		"capacitor://localhost",
		"http://localhost",
		"http://localhost:4200",
		"https://localhost",
	}
	for i := range defaultAllowedOrigins {
		defaultAllowedOrigins[i] = strings.ToLower(defaultAllowedOrigins[i])
	}

	allowMethods := "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	allowHeaders := "Content-Type, Authorization, X-Correlation-Id"
	exposeHeaders := "X-Correlation-Id"

	return func(ctx *context.Context) {
		origin := strings.ToLower(ctx.Input.Header("Origin"))

		runMode := web.BConfig.RunMode
		allowedOriginsEnv := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
		allowAll := (runMode != "prod" && allowedOriginsEnv == "")

		if origin != "" {
			if allowAll {
				ctx.Output.Header("Access-Control-Allow-Origin", "*")
				ctx.Output.Header("Access-Control-Allow-Methods", allowMethods)
				ctx.Output.Header("Access-Control-Allow-Headers", allowHeaders)
				ctx.Output.Header("Access-Control-Expose-Headers", exposeHeaders)
			} else {
				allowed := defaultAllowedOrigins
				if allowedOriginsEnv != "" {
					parts := strings.Split(allowedOriginsEnv, ",")
					allowed = allowed[:0]
					for _, p := range parts {
						p = strings.ToLower(strings.TrimSpace(p))
						if p != "" {
							allowed = append(allowed, p)
						}
					}
				}
				if isAllowedOrigin(origin, allowed) {
					ctx.Output.Header("Access-Control-Allow-Origin", origin)
					ctx.Output.Header("Vary", "Origin")
					ctx.Output.Header("Access-Control-Allow-Methods", allowMethods)
					ctx.Output.Header("Access-Control-Allow-Headers", allowHeaders)
					ctx.Output.Header("Access-Control-Expose-Headers", exposeHeaders)
				}
			}
		}

		if ctx.Input.Is("OPTIONS") {
			ctx.Output.SetStatus(http.StatusNoContent)
			_ = ctx.Output.Body([]byte{})
			return
		}
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	for _, o := range allowed {
		if origin == o {
			return true
		}
	}
	return false
}
