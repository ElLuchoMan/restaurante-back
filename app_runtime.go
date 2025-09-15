package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	httpSwagger "github.com/swaggo/http-swagger"

	appcors "restaurante/internal/middleware/cors"
	"restaurante/logging"
)

func setupAndRun() {
	runMode := web.BConfig.RunMode
	logging.Setup(runMode)

	web.InsertFilter("*", web.BeforeRouter, logging.StartTimer)
	web.InsertFilter("*", web.FinishRouter, logging.LogRequest)
	web.InsertFilter("*", web.FinishRouter, logging.LogIfError)

	web.Get("/healthz", func(ctx *context.Context) {
		ctx.Output.SetStatus(200)
		_ = ctx.Output.Body([]byte("ok"))
	})
	web.Get("/readyz", func(ctx *context.Context) {
		status := 200
		if sqlDB, err := getSQLPinger(); err == nil && sqlDB != nil {
			if err := sqlDB.Ping(); err != nil {
				status = http.StatusServiceUnavailable
			}
		}
		ctx.Output.SetStatus(status)
		_ = ctx.Output.Body([]byte(http.StatusText(status)))
	})

	if v := os.Getenv("MAX_BODY_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			web.BConfig.MaxMemory = n
		}
	}
	if v := os.Getenv("MULTIPART_MAX_MEMORY_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			web.BConfig.CopyRequestBody = false
			web.BConfig.MaxMemory = int64(n) * 1024 * 1024
		}
	}

	_ = runMode
	web.InsertFilter("/*", web.BeforeRouter, appcors.CORS())

	web.InsertFilter("*", web.BeforeRouter, setStaticHeaders)

	web.Handler("/swagger/*", httpSwagger.WrapHandler)

	if os.Getenv("SKIP_CRON") != "1" && dbReady {
		go generarNominaAutomatica()
	}

	if os.Getenv("SKIP_WEB_RUN") != "1" {
		webRun()
	}
}
