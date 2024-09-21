package model

import (
	"errors"
	"time"

	"github.com/anacrolix/torrent"
)

type TorrentStatusRes struct {
	DownloadingFiles      []*DownloadingFile  `json:"downloadingFiles"`
	LocalFiles            []*LocalFile        `json:"localFiles"`
	Stats                 *Stats              `json:"stats"`
	Tasks                 *Tasks              `json:"tasks"`
	ActiveDownloadsCounts int64               `json:"activeDownloadsCounts"`
	TorrentClientStats    torrent.ClientStats `json:"torrentClientStats" swaggerignore:"true"`
	DownloadQueueStats    *DownloadQueueStats `json:"downloadQueueStats"`
}

type DownloadQueueStats struct {
	Size           int `json:"size"`
	EnqueueCounter int `json:"enqueueCounter"`
	Capacity       int `json:"capacity"`
	Workers        int `json:"workers"`
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
	Done           bool             `json:"done"`
}

type LocalFile struct {
	Name             string    `json:"name"`
	Size             int64     `json:"size"`
	DownloadLink     string    `json:"downloadLink"`
	StreamLink       string    `json:"streamLink"`
	ExpireTime       time.Time `json:"expireTime"`
	TotalDownloads   *int64    `json:"totalDownloads"`
	ActiveDownloads  *int64    `json:"activeDownloads"`
	LastDownloadTime time.Time `json:"lastDownloadTime"`
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

type Stats struct {
	Configs                       *StatsConfigs `json:"configs"`
	TotalFilesSizeMb              int64         `json:"totalFilesSizeMb"`
	LocalFilesSizeMb              int64         `json:"localFilesSizeMb"`
	DownloadingFilesFinalSizeMb   int64         `json:"downloadingFilesFinalSizeMb"`
	DownloadingFilesCurrentSizeMb int64         `json:"downloadingFilesCurrentSizeMb"`
	RemainingSpaceMb              int64         `json:"remainingSpaceMb"`
	TorrentDownloadTimeoutMin     int64         `json:"torrentDownloadTimeoutMin"`
}

type StatsConfigs struct {
	DownloadSpaceThresholdMb            int64   `json:"downloadSpaceThresholdMb"`
	DownloadSpaceLimitMb                int64   `json:"downloadSpaceLimitMb"`
	DownloadFileSizeLimitMb             int64   `json:"downloadFileSizeLimitMb"`
	TorrentFilesExpireHour              int64   `json:"torrentFilesExpireHour"`
	TorrentFilesServingConcurrencyLimit int64   `json:"torrentFilesServingConcurrencyLimit"`
	TorrentDownloadConcurrencyLimit     int64   `json:"torrentDownloadConcurrencyLimit"`
	TorrentFilesServingDisabled         bool    `json:"torrentFilesServingDisabled"`
	TorrentDownloadDisabled             bool    `json:"torrentDownloadDisabled"`
	TorrentFileExpireDelayFactor        float32 `json:"torrentFileExpireDelayFactor"`
	TorrentFileExpireExtendHour         int64   `json:"torrentFileExpireExtendHour"`
}

type Tasks struct {
	DbsInvalidLocalLinksRemover string `json:"dbsInvalidLocalLinksRemover"`
	InvalidLocalFilesRemover    string `json:"invalidLocalFilesRemover"`
	ExpiredFilesRemover         string `json:"expiredFilesRemover"`
	InCompleteDownloadsRemover  string `json:"inCompleteDownloadsRemover"`
	OrphanMetaFilesRemover      string `json:"orphanMetaFilesRemover"`
	DiskSpaceCleaner            string `json:"diskSpaceCleaner"`
	AutoDownloader              string `json:"autoDownloader"`
}

var ErrTorrentLinkNotFound = errors.New("torrent link not found")
var ErrFileNotFound = errors.New("file not found")
var ErrFileAlreadyExist = errors.New("file already exist")
var ErrTorrentDownloadDisabled = errors.New("torrent downloader is disabled")
var ErrTorrentFilesServingDisabled = errors.New("serving torrent files is disabled")
var ErrTorrentDownloadTimeout = errors.New("torrent download timeout")
var ErrTorrentDownloadInactive = errors.New("torrent download inactivity")
var ErrTorrentDownloadConcurrencyLimit = errors.New("torrent download concurrency exceed, try later")
var ErrEmptyFile = errors.New("file is empty")
var ErrFileSizeExceeded = errors.New("file size exceeds the limit")
var ErrNotEnoughSpace = errors.New("not Enough space left")
var ErrAlreadyDownloading = errors.New("already downloading")
var ErrMaximumDiskUsageExceeded = errors.New("maximum disk usage exceeded")
var ErrExtendLocalFileExpireIsDisabled = errors.New("extending local files expire time is disabled")
