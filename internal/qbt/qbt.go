package qbt

import (
	"context"
	"echo_sandbox/internal/utils"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/autobrr/go-qbittorrent"
)

type QbtClient interface {
	LoginCtx(ctx context.Context) error
	ListTarTorrentsCtx(ctx context.Context, opts *ListTorrentsOptions) ([]*TorrentInfo, error)
	DeleteTorrentsByHash(ctx context.Context, hashes []string, deleteFiles bool) error
	DeleteTorrentByNameCtx(ctx context.Context, name string) error
	GetTorrentCtx(ctx context.Context, hash string) (*TorrentInfo, error)
}

type QbtClientWrapper struct {
	client *qbittorrent.Client
	config *QbtClientConfig
}

type QbtClientConfig struct {
	Host     string `yaml:"Host"`
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
}

type TorrentInfo struct {
	Name        string
	Hash        string
	ContentPath string
	Size        int64
	Tags        []string
	FilesCount  int
}

type ListTorrentsOptions struct {
	Sort            string
	Reverse         bool
	FetchFilesCount bool
}

var ErrTorrentNotFound = errors.New("torrent not found")

func mapTorrentInfo(t qbittorrent.Torrent) *TorrentInfo {
	return &TorrentInfo{
		Name:        t.Name,
		Hash:        t.Hash,
		ContentPath: t.ContentPath,
		Tags:        strings.Split(t.Tags, ", "),
		Size:        t.Size,
	}
}

func NewQbtClientWrapper(config *QbtClientConfig) *QbtClientWrapper {
	client := qbittorrent.NewClient(qbittorrent.Config{
		Host:     config.Host,
		Username: config.Username,
		Password: config.Password,
	})

	return &QbtClientWrapper{
		client: client,
		config: config,
	}
}

func (client *QbtClientWrapper) LoginCtx(ctx context.Context) error {
	return client.client.LoginCtx(ctx)
}

func (client *QbtClientWrapper) ListTarTorrentsCtx(ctx context.Context, opts *ListTorrentsOptions) ([]*TorrentInfo, error) {
	filterOpts := qbittorrent.TorrentFilterOptions{
		Filter: qbittorrent.TorrentFilterCompleted,
		Tag:    "tar",
	}

	if opts != nil {
		filterOpts.Sort = opts.Sort
		filterOpts.Reverse = opts.Reverse
	}

	data, err := client.client.GetTorrentsCtx(ctx, filterOpts)
	if err != nil {
		return nil, err
	}

	// TODO: fetch files count
	var filesCountMap sync.Map

	if opts.FetchFilesCount {
		// for every torrent get its hash
		var wg sync.WaitGroup

		for _, torr := range data {
			wg.Add(1)
			go func(hash string) {
				files, err := client.client.GetFilesInformationCtx(ctx, hash)
				if err != nil {
					return
				}
				filesCount := 0

				for _, file := range *files {
					if file.Progress == 1 {
						filesCount++
					}
				}

				filesCountMap.Store(hash, filesCount)
				wg.Done()
			}(torr.Hash)
		}

		wg.Wait()

	}
	torrInfo := utils.SliceMap(data, func(i int, t qbittorrent.Torrent) *TorrentInfo {
		return mapTorrentInfo(t)
	})

	if opts.FetchFilesCount {
		for _, torr := range torrInfo {
			if count, ok := filesCountMap.Load(torr.Hash); ok {
				torr.FilesCount = count.(int)
			}
		}
	}
	return torrInfo, nil
}

func (client *QbtClientWrapper) GetTorrentCtx(ctx context.Context, hash string) (*TorrentInfo, error) {
	torrs, err := client.client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{
		Hashes: []string{hash},
	})
	if err != nil {
		return nil, err
	}

	if len(torrs) == 0 {
		return nil, ErrTorrentNotFound
	}

	return mapTorrentInfo(torrs[0]), nil
}

func (client *QbtClientWrapper) DeleteTorrentsByHash(ctx context.Context, hashes []string, deleteFiles bool) error {
	return client.client.DeleteTorrentsCtx(ctx, hashes, deleteFiles)
}

func (client *QbtClientWrapper) DeleteTorrentByNameCtx(ctx context.Context, name string) error {
	torrents, err := client.client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{
		Filter: qbittorrent.TorrentFilterCompleted,
		Tag:    "tar",
	})
	if err != nil {
		return fmt.Errorf("failed to get torrents list: %w", err)
	}

	for _, torrrent := range torrents {
		if torrrent.Name == name {
			err := client.client.DeleteTorrents([]string{torrrent.Hash}, true)

			if err != nil {
				return err
			} else {
				return nil
			}
		}
	}
	return ErrTorrentNotFound
}
