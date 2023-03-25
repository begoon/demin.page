package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"

	. "homepage/assert"
	"homepage/console"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

func main() {
	flag.Parse()
	console.Init()
	serve()
}

var fLocal = flag.Bool("local", false, "use local files")

//go:embed site/*
var site embed.FS

func serve() {
	e := echo.New()
	e.Debug = false
	e.HideBanner = true
	e.DisableHTTP2 = true

	var siteFS fs.FS = site

	if *fLocal {
		siteFS = os.DirFS(T(os.Getwd()))
	}
	e.StaticFS("/", echo.MustSubFS(siteFS, "site"))

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisablePrintStack: true,
	}))
	e.Use(middleware.LoggerWithConfig(
		middleware.LoggerConfig{
			Skipper: func(c echo.Context) bool {
				return c.Request().URL.Path == "/time"
			},
			Format: strings.Join([]string{
				`${time_rfc3339_nano}`,
				`${method} ${uri}`,
				`${status} ${latency_human} ${bytes_out}`,
			}, " - ") + "\n",
			CustomTimeFormat: "2006-01-02 15:04:05",
		},
	))

	e.Use(middleware.CORS())

	e.GET("/time", func(c echo.Context) error {
		return c.String(200, fmt.Sprint(time.Now().Format(time.RFC3339)))
	})

	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8000"
	}
	hostname := T(os.Hostname())
	log.Info().Str("hostname", hostname).Str("port", port).Send()

	err := e.Start(":" + port)
	log.Fatal().Err(err).Send()
}
