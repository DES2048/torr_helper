package server

import (
	"context"
	"echo_sandbox/internal/qbt"
	"echo_sandbox/internal/utils"
	"echo_sandbox/internal/utils/tar"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

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

type ApiHandler struct {
	config    *HttpServerConfig
	qbtClient qbt.QbtClient
}

func NewApiHandler(config *HttpServerConfig, qbtClient qbt.QbtClient) *ApiHandler {
	return &ApiHandler{
		config:    config,
		qbtClient: qbtClient,
	}
}

func (h *ApiHandler) ListTorrents(c echo.Context) error {
	ctx := context.Background()

	// get filesCount opts
	fetchFilesCount := c.QueryParam("filesCount")

	// get torrents list
	err := h.qbtClient.LoginCtx(ctx)
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

	torrents, err := h.qbtClient.ListTarTorrentsCtx(ctx, listOpts)
	// TODO: wrap error
	if err != nil {
		return err
	}

	torrResp := utils.SliceMap(torrents, func(_ int, torr *qbt.TorrentInfo) TorrentsListResp {
		// stat tars for checking that exists
		tarPath, err := tar.GetTarPath(torr, h.config.TarsDirs...)
		stat, _ := os.Stat(tarPath)

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
}

func (h *ApiHandler) DeleteTorrent(c echo.Context) error {
	torrHash := c.Param("id")
	err := h.qbtClient.DeleteTorrentsByHash(context.Background(), []string{torrHash}, true)
	if err != nil {
		return err
	}

	return c.NoContent(204)
}

func (h *ApiHandler) MakeTar(c echo.Context) error {
	id := c.Param("id")
	// TODO: get torrent
	torr, err := h.qbtClient.GetTorrentCtx(context.Background(), id)
	if err != nil {
		return err
	}

	// make tar

	lastPath := filepath.Base(torr.ContentPath)
	// if torrent was renamed tar name should get by torrent name
	if lastPath != torr.Name {
		lastPath = torr.Name
	}

	// Write File or directory to a tar file.
	tarPath := filepath.Join(h.config.TarCreateDir, lastPath+".tar")

	go tar.CreateTar(tarPath, torr.ContentPath, filepath.Base(torr.ContentPath))
	return c.NoContent(http.StatusNoContent)
}

func (h *ApiHandler) DeleteTar(c echo.Context) error {
	id := c.Param("id")
	// TODO: get torrent
	torr, err := h.qbtClient.GetTorrentCtx(context.Background(), id)
	if err != nil {
		return err
	}

	tarPath, _ := tar.GetTarPath(torr, h.config.TarsDirs...)

	err = os.Remove(tarPath)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
