package server

import (
	"context"
	"crypto/subtle"
	"echo_sandbox/internal/qbt"
	"echo_sandbox/internal/server/sse"
	"echo_sandbox/internal/utils/tar"
	"log"
	"net/http"

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
	qbtClient    qbt.QbtClient
}

func NewHttpServer(config *HttpServerConfig, qbtClient qbt.QbtClient) *HttpServer {
	e := echo.New()
	s := &HttpServer{
		config:       config,
		echoInstance: e,
		sseBroker:    sse.NewBroker[sse.SseEvent](),
		qbtClient:    qbtClient,
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

	// routes
	apiGroup := e.Group("/api")

	apiHandler := NewApiHandler(config, qbtClient)

	// list torrents
	apiGroup.GET("/torrents", apiHandler.ListTorrents)

	apiGroup.DELETE("/torrents/:id", apiHandler.DeleteTorrent)

	apiGroup.POST("/make-tar/:id", apiHandler.MakeTar)

	// delete tar
	apiGroup.DELETE("/tars/:id", apiHandler.DeleteTar)

	// download tar
	e.GET("/tars/:id", func(c echo.Context) error {
		torrId := c.Param("id")
		ctx := context.Background()

		torr, err := s.qbtClient.GetTorrentCtx(ctx, torrId)
		if err != nil {
			return err
		}
		tarPath, err := tar.GetTarPath(torr, config.TarsDirs...)
		if err != nil {
			return err
		}
		return c.File(tarPath)
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

	apiGroup.POST("/tars", func(c echo.Context) error {
		tarName := c.FormValue("name")

		ev := sse.SseEvent{
			Event: []byte("new-tar"),
			Data:  []byte(tarName),
		}
		s.sseBroker.Pub(ev)
		return nil
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
