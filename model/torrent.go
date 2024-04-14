package model

import "github.com/anacrolix/torrent"

type DownloadTorrentReq struct {
}

func (d *DownloadTorrentReq) Validate() string {
	return ""
}

type TorrentStatusRes struct {
	DownloadingFiles []*DownloadingFile `json:"downloadingFiles"`
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
