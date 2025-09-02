package server

import (
	"context"
	"crypto/subtle"
	"echo_sandbox/internal/qbt"
	"echo_sandbox/internal/server/sse"
	"echo_sandbox/internal/utils"
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

type TarInfo struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Url  string `json:"url"`
}

type TorrentsListResp struct {
	Name       string   `json:"name"`
	Id         string   `json:"id"`
	Tar        *TarInfo `json:"tarInfo,omitempty"`
	Size       int64    `json:"size"`
	Tags       []string `json:"tags"`
	FilesCount int      `json:"filesCount"`
}

type HttpServer struct {
	config       *HttpServerConfig
	echoInstance *echo.Echo
	sseBroker    *sse.Broker[sse.SseEvent]
	qbtClient    qbt.QbtClient
}

func getTarPath(torrInfo *qbt.TorrentInfo) string {
	lastPath := filepath.Base(torrInfo.ContentPath)
	// if torrent was renamed tar name should get by torrent name
	if lastPath != torrInfo.Name {
		lastPath = torrInfo.Name
	}
	return filepath.Join(filepath.Dir(torrInfo.ContentPath), lastPath+".tar")
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
	/*/ serve tars
	staticGroup := e.Group("/tars")
	staticGroup.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   config.TarsDir,
		Browse: true,
	})) */

	// routes
	apiGroup := e.Group("/api")

	// list torrents
	apiGroup.GET("/torrents", func(c echo.Context) error {
		ctx := context.Background()

		// get filesCount opts
		fetchFilesCount := c.QueryParam("filesCount")

		// get torrents list
		err := s.qbtClient.LoginCtx(ctx)
		// TODO: wrap error
		if err != nil {
			return err
		}

		listOpts := &qbt.ListTorrentsOptions{
			Sort:    "completion_on",
			Reverse: true,
		}

		if fetchFilesCount == "true" {
			listOpts.FetchFilesCount = true
		}

		torrents, err := s.qbtClient.ListTarTorrentsCtx(ctx, listOpts)
		// TODO: wrap error
		if err != nil {
			return err
		}

		torrResp := utils.SliceMap(torrents, func(_ int, torr *qbt.TorrentInfo) TorrentsListResp {
			// stat tars for checking that exists
			tarPath := getTarPath(torr)
			stat, err := os.Stat(tarPath)

			resp := TorrentsListResp{
				Name:       torr.Name,
				Id:         torr.Hash,
				Size:       torr.Size,
				Tags:       torr.Tags,
				FilesCount: torr.FilesCount,
			}

			if err == nil {
				tarInfo := &TarInfo{
					Size: stat.Size(),
					Url:  path.Join("/tars", torr.Hash),
					Path: tarPath,
				}
				resp.Tar = tarInfo

			}

			return resp
		})
		return c.JSONPretty(http.StatusOK, torrResp, "  ")
	})

	apiGroup.DELETE("/torrents/:id", func(c echo.Context) error {
		torrHash := c.Param("id")
		err := s.qbtClient.DeleteTorrentsByHash(context.Background(), []string{torrHash}, true)
		if err != nil {
			return err
		}

		return c.NoContent(204)
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
	apiGroup.DELETE("/tars/:id", func(c echo.Context) error {
		id := c.Param("id")
		// TODO: get torrent
		torr, err := s.qbtClient.GetTorrentCtx(context.Background(), id)
		if err != nil {
			return err
		}

		tarPath := getTarPath(torr)

		err = os.Remove(tarPath)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusNoContent)
	})

	// download tar
	e.GET("/tars/:id", func(c echo.Context) error {
		torrId := c.Param("id")
		ctx := context.Background()

		torr, err := s.qbtClient.GetTorrentCtx(ctx, torrId)
		if err != nil {
			return err
		}

		return c.File(getTarPath(torr))
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
