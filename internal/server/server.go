package server

import (
	"crypto/subtle"
	"echo_sandbox/internal/server/sse"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type NotFoundResponse struct {
	Message string `json:"message"`
}

type HttpServer struct {
	config       *HttpServerConfig
	echoInstance *echo.Echo
	sseBroker    *sse.Broker[sse.SseEvent]
}

func NewHttpServer(config *HttpServerConfig) *HttpServer {
	e := echo.New()
	s := &HttpServer{
		config:       config,
		echoInstance: e,
		sseBroker:    sse.NewBroker[sse.SseEvent](),
	}

	go s.sseBroker.Start()

	// middewares
	// basic auth
	if config.BasicAuth {
		e.Use(middleware.BasicAuth(func(username, password string, ctx echo.Context) (bool, error) {
			if subtle.ConstantTimeCompare([]byte(username), []byte(config.User)) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte(config.Password)) == 1 {
				return true, nil
			}

			return false, nil
		}))
	}

	e.Use(middleware.Logger())
	// serve tars
	staticGroup := e.Group("/tars")
	staticGroup.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   config.TarsDir,
		Browse: true,
	}))

	// routes
	apiGroup := e.Group("/api")

	// list tars
	apiGroup.GET("/tars", func(c echo.Context) error {
		// get tars list
		tars, err := filepath.Glob(filepath.Join(config.TarsDir, "*.tar"))
		if err != nil {
			return err
		}

		resp := make([]struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		}, len(tars))

		for idx, tar := range tars {
			resp[idx].Name = filepath.Base(tar)
			resp[idx].Url = path.Join("/tars", filepath.Base(tar))
		}
		return c.JSONPretty(http.StatusOK, resp, "  ")
	})

	apiGroup.POST("/tars", func(c echo.Context) error {
		tarName := c.FormValue("name")

		ev := sse.SseEvent{
			Event: []byte("new-tar"),
			Data:  []byte(tarName),
		}
		s.sseBroker.Pub(ev)
		return nil
	})

	// delete tar
	apiGroup.DELETE("/tars/:name", func(c echo.Context) error {
		name := c.Param("name")

		// check file exists
		tarpath := filepath.Join(config.TarsDir, name)
		_, err := os.Stat(tarpath)

		if os.IsNotExist(err) {
			return c.JSON(http.StatusNotFound, NotFoundResponse{Message: "Given tar not found"})
		}

		err = os.Remove(tarpath)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	})

	// sse
	e.GET("/sse", func(c echo.Context) error {
		log.Printf("SSE client connected, ip: %v", c.RealIP())

		eventChan := make(chan sse.SseEvent)
		s.sseBroker.Subscribe(eventChan)

		w := c.Response()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		for {
			select {
			case <-c.Request().Context().Done():
				log.Printf("SSE client disconnected, ip: %v", c.RealIP())
				s.sseBroker.Unubscribe(eventChan)
				return nil
			case event := <-eventChan:
				// log.Printf("New event received: %v", event)
				if err := event.MarshalTo(w); err != nil {
					return err
				}
				w.Flush()
			}
		}
	})

	// test stub
	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from echo")
	})

	return s
}

func (s *HttpServer) Start() {
	s.echoInstance.Logger.Fatal(s.echoInstance.Start(s.config.Address))
	// return nil
}
