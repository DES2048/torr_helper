package server

import (
	"crypto/subtle"
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
}

func NewHttpServer(config *HttpServerConfig) *HttpServer {
	e := echo.New()
	s := &HttpServer{
		config:       config,
		echoInstance: e,
	}

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

	// test stub
	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from echo")
	})

	return s
}

func (s *HttpServer) Start() {
	s.echoInstance.Logger.Fatal(s.echoInstance.Start(s.config.Address))
	//return nil
}
