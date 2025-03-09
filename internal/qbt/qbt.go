package qbt

import (
	"context"
	"echo_sandbox/internal/utils"
	"errors"
	"fmt"

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
	Host     string
	Username string
	Password string
}

type TorrentInfo struct {
	Name        string
	Hash        string
	ContentPath string
}

type ListTorrentsOptions struct {
	Sort    string
	Reverse bool
}

var ErrTorrentNotFound = errors.New("torrent not found")

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
	fulterOpts := qbittorrent.TorrentFilterOptions{
		Filter: qbittorrent.TorrentFilterCompleted,
		Tag:    "tar",
	}

	if opts != nil {
		fulterOpts.Sort = opts.Sort
		fulterOpts.Reverse = opts.Reverse
	}

	data, err := client.client.GetTorrentsCtx(ctx, fulterOpts)
	if err != nil {
		return nil, err
	}

	torrInfo := utils.SliceMap(data, func(i int, t qbittorrent.Torrent) *TorrentInfo {
		return &TorrentInfo{
			Name:        t.Name,
			Hash:        t.Hash,
			ContentPath: t.ContentPath,
		}
	})

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

	return &TorrentInfo{
		Name:        torrs[0].Name,
		Hash:        torrs[0].Hash,
		ContentPath: torrs[0].ContentPath,
	}, nil
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
