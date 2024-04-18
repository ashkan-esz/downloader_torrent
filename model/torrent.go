package model

import "github.com/anacrolix/torrent"

type TorrentStatusRes struct {
	DownloadingFiles []*DownloadingFile `json:"downloadingFiles"`
	LocalFiles       []*LocalFile       `json:"localFiles"`
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
}

type LocalFile struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	StreamLink string `json:"streamLink"`
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
