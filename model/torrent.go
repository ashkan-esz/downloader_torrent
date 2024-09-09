package model

import (
	"time"

	"github.com/anacrolix/torrent"
)

type TorrentStatusRes struct {
	DownloadingFiles []*DownloadingFile `json:"downloadingFiles"`
	LocalFiles       []*LocalFile       `json:"localFiles"`
	DiskInfo         *DiskInfo          `json:"diskInfo"`
}

type DownloadingFile struct {
	State          string           `json:"state"`
	Name           string           `json:"name"`
	Size           int64            `json:"size"`
	DownloadedSize int64            `json:"downloadedSize"`
	TorrentUrl     string           `json:"torrentUrl"`
	Torrent        *torrent.Torrent `json:"-"`
	TitleId        string           `json:"titleId"`
	TitleName      string           `json:"titleName"`
	TitleType      string           `json:"titleType"`
	StartTime      time.Time        `json:"startTime"`
	Error          error            `json:"error"`
}

type LocalFile struct {
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	DownloadLink string `json:"downloadLink"`
	StreamLink   string `json:"streamLink"`
}

type StreamStatusRes struct {
	ConvertingFiles []*ConvertingFile `json:"convertingFiles"`
}

type ConvertingFile struct {
	Progress string  `json:"progress"`
	Name     string  `json:"name"`
	Size     int64   `json:"size"`
	Duration float64 `json:"duration"`
}

type DiskInfo struct {
	TotalFilesSizeMb      int64 `json:"totalFilesSizeMb"`
	MaxDownloadSpaceMb    int64 `json:"maxDownloadSpaceMb"`
	DownloadMaxFileSizeMb int64 `json:"DownloadMaxFileSizeMb"`
}
