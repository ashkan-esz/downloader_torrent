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
	MetaFileName   string           `json:"metaFileName"`
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
	Name            string    `json:"name"`
	Size            int64     `json:"size"`
	DownloadLinks   []string  `json:"downloadLinks"`
	StreamLink      string    `json:"streamLink"`
	ExpireTime      time.Time `json:"expireTime"`
	TotalDownloads  *int64    `json:"totalDownloads"`
	ActiveDownloads *int64    `json:"activeDownloads"`
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
	Configs                       *DiskInfoConfigs `json:"configs"`
	TotalFilesSizeMb              int64            `json:"totalFilesSizeMb"`
	LocalFilesSizeMb              int64            `json:"localFilesSizeMb"`
	DownloadingFilesFinalSizeMb   int64            `json:"downloadingFilesFinalSizeMb"`
	DownloadingFilesCurrentSizeMb int64            `json:"downloadingFilesCurrentSizeMb"`
	RemainingSpaceMb              int64            `json:"remainingSpaceMb"`
}

type DiskInfoConfigs struct {
	DownloadSpaceThresholdMb int64 `json:"downloadSpaceThresholdMb"`
	DownloadSpaceLimitMb     int64 `json:"downloadSpaceLimitMb"`
	DownloadFileSizeLimitMb  int64 `json:"downloadFileSizeLimitMb"`
	TorrentFilesExpireHour   int64 `json:"torrentFilesExpireHour"`
}
