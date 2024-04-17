package service

import (
	"downloader_torrent/internal/repository"
)

type IStreamService interface {
	StreamMedia(filename string) error
}

type StreamService struct {
	movieRepo   repository.IMovieRepository
	downloadDir string
}

func NewStreamService(movieRepo repository.IMovieRepository) *StreamService {
	return &StreamService{
		movieRepo:   movieRepo,
		downloadDir: "./downloads/",
	}
}

//------------------------------------------
//------------------------------------------

func (m *StreamService) StreamMedia(filename string) error {
	return nil
}
