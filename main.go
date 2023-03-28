package main

import (
	"embed"
	"flag"
	"fmt"
	. "homepage/assert"
	"homepage/console"
	"io/fs"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"strings"
	"sync"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/gorilla/sessions"
	echopprof "github.com/hiko1129/echo-pprof"

	// echopprof "github.com/hiko1129/echo-pprof"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/oklog/ulid/v2"
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

type Colony [100][100]uint8

func NewColony() *Colony {
	c := Colony{}
	for i := 1; i < len(c)-1; i++ {
		for j := 1; j < len(c[i])-1; j++ {
			c[i][j] = uint8(rand.Int()>>3) & 1
		}
	}
	return &c
}

func (c Colony) Neighours(x, y int) uint8 {
	n := c[y-1][x-1] + c[y-1][x] + c[y-1][x+1]
	n += c[y][x-1] + c[y][x+1]
	n += c[y+1][x-1] + c[y+1][x] + c[y+1][x+1]
	return n
}

func (c Colony) Alive(x, y int) bool {
	return c[y][x] == 1
}

func (c Colony) Next() *Colony {
	g := Colony{}
	for y := 1; y < len(c)-1; y++ {
		for x := 1; x < len(c[x])-1; x++ {
			n := c.Neighours(x, y)
			if c.Alive(x, y) {
				if n == 2 || n == 3 {
					g[y][x] = 1
				}
			} else {
				if n == 3 {
					g[y][x] = 1
				}
			}
		}
	}
	return &g
}

func TimeHandler(c echo.Context) error {
	return c.String(200, fmt.Sprint(time.Now().Format(time.RFC3339)))
}

func PingHandler(c echo.Context) error {
	return c.String(200, "pong")
}

var secured = []string{"/ping"}

func apiKeyValidator(key string, c echo.Context) (bool, error) {
	return key == "zorba", nil
}

func apiKeySkipper(c echo.Context) bool {
	u := c.Request().URL.RequestURI()
	for _, e := range secured {
		if strings.Contains(u, e) {
			return false
		}
	}
	return true
}

func Configure() *echo.Echo {
	e := echo.New()
	e.Debug = false
	e.HideBanner = true
	echopprof.Wrap(e)

	var siteFS fs.FS = site

	if *fLocal {
		siteFS = os.DirFS(T(os.Getwd()))
	}
	e.StaticFS("/", echo.MustSubFS(siteFS, "site"))

	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "header:X-Api-Key",
		Validator: apiKeyValidator,
		Skipper:   apiKeySkipper,
	}))

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

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("!"))))

	e.GET("/time", TimeHandler)

	e.GET("/ping", PingHandler)

	colonies := map[string]*Colony{}
	var lock = sync.RWMutex{}

	load := func(id string) *Colony {
		lock.RLock()
		defer lock.RUnlock()
		if c, ok := colonies[id]; ok {
			log.Info().Str("id", id).Msg("loaded")
			return c
		}
		log.Info().Str("id", id).Msg("generated")
		return NewColony()
	}

	save := func(id string, c *Colony) {
		lock.Lock()
		defer lock.Unlock()
		colonies[id] = c
		log.Info().Str("id", id).Msg("saved")
	}

	e.GET("/life", func(c echo.Context) error {
		s, err := session.Get("session", c)
		if err != nil {
			log.Error().Err(err).Msg("session")
		}

		s.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
		}

		colonyID, found := s.Values["colonyID"].(string)
		if !found || c.QueryParam("restart") != "" {
			log.Info().Str("id", colonyID).Msg("new colony")
			colonyID = ulid.Make().String()
			s.Values["colonyID"] = colonyID
		}

		colony := load(colonyID)
		next := *colony.Next()
		colony = &next
		save(colonyID, colony)

		s.Save(c.Request(), c.Response())

		w := c.Response().Writer
		w.Header().Set("Content-Type", "image/svg+xml")
		g := svg.New(w)

		sx := 5
		sy := 5
		r := sx / 2
		g.Start(sx*len(colony), sy*len(colony[0]))
		for i := 0; i < len(colony); i++ {
			for j := 0; j < len(colony[0]); j++ {
				var color string
				if colony[i][j] == 1 {
					color = "green"
				} else {
					color = "white"
				}
				g.Circle(i*sx+r, j*sy+r, r, "fill: "+color)
			}
		}
		g.End()
		return nil
	})
	return e
}

func serve() {
	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8000"
	}
	hostname := T(os.Hostname())
	log.Info().Str("hostname", hostname).Str("port", port).Send()

	err := Configure().Start(":" + port)
	log.Fatal().Err(err).Send()
}
