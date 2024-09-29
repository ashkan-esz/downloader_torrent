package model

import (
	"errors"
	"slices"
	"time"

	"github.com/anacrolix/torrent"
)

type UserTorrent struct {
	UserId         int64     `gorm:"column:userId;type:integer;not null;primaryKey;uniqueIndex:UserTorrent_userId_key;" swaggerignore:"true"`
	TorrentLeachGb int       `gorm:"column:torrentLeachGb;type:integer;not null;"`
	TorrentSearch  int       `gorm:"column:torrentSearch;type:integer;not null;"`
	FirstUseAt     time.Time `gorm:"column:firstUseAt;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
}

func (UserTorrent) TableName() string {
	return "UserTorrent"
}

//---------------------------------------
//---------------------------------------

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
	Size                int  `json:"size"`
	EnqueueCounter      int  `json:"enqueueCounter"`
	Capacity            int  `json:"capacity"`
	Workers             int  `json:"workers"`
	DequeueWorkersSleep bool `json:"dequeueWorkersSleep"`
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
	UserId         int64            `json:"userId"`
	BotId          string           `json:"botId"`
	ChatId         string           `json:"chatId"`
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

type DownloadRequestInfo struct {
	MovieId      string `json:"movieId"`
	TorrentUrl   string `json:"torrentUrl"`
	IsAdmin      bool   `json:"isAdmin"`
	DownloadNow  bool   `json:"downloadNow"`
	UserId       int64  `json:"userId"`
	IsBotRequest bool   `json:"isBotRequest"`
	BotId        string `json:"botId"`
	ChatId       string `json:"chatId"`
	BotUsername  string `json:"botUsername"`
}

type DownloadRequestRes struct {
	DownloadingFile *DownloadingFile `json:"downloadingFile"`
	Message         string           `json:"message"`
	QueueIndex      int              `json:"queueIndex"`
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
	TorrentUserEnqueueLimit             int     `json:"torrentUserEnqueueLimit"`
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
var ErrBotIsDisabled = errors.New("bot is disabled")
var ErrBotNoPermissionTorrentLeach = errors.New("bot dont have the permission to torrent-leach")
var ErrNoRoleFoundForUser = errors.New("no role found for user")
var ErrReachedTorrentLeachLimit = errors.New("reached torrent leach limit, wait for reset")
var ErrTooManyQueuedDownload = errors.New("too many items added to queue, wait before they end")

func GetErrorCode(err error) int {
	code403 := []error{
		ErrBotIsDisabled,
		ErrBotNoPermissionTorrentLeach,
	}
	code404 := []error{
		ErrNoRoleFoundForUser,
	}
	code409 := []error{}
	code429 := []error{
		ErrReachedTorrentLeachLimit,
		ErrTooManyQueuedDownload,
	}

	if slices.Contains(code403, err) {
		return 403
	}
	if slices.Contains(code404, err) {
		return 404
	}
	if slices.Contains(code409, err) {
		return 409
	}
	if slices.Contains(code429, err) {
		return 429
	}

	return 0
}
