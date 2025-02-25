package qbt

import (
	"context"
	"errors"
	"fmt"

	"github.com/autobrr/go-qbittorrent"
)

type QbtClient interface {
	LoginCtx(ctx context.Context) error
	DeleteTorrentByNameCtx(ctx context.Context, name string) error
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

func (client *QbtClientWrapper) DeleteTorrentByNameCtx(ctx context.Context, name string) error {
	torrents, err := client.client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{
		Tag: "tar",
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
