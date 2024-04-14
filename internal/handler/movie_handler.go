package handler

import (
	"downloader_torrent/internal/service"
	"downloader_torrent/model"
	"downloader_torrent/pkg/response"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type IMovieHandler interface {
	DownloadTorrent(c *fiber.Ctx) error
	TorrentStatus(c *fiber.Ctx) error
}

type MovieHandler struct {
	movieService service.IMovieService
}

func NewMovieHandler(movieService service.IMovieService) *MovieHandler {
	return &MovieHandler{
		movieService: movieService,
	}
}

//------------------------------------------
//------------------------------------------

// DownloadTorrent godoc
//
//	@Summary		Download Torrent
//	@Description	download from torrent into local storage
//	@Tags			Torrent-Download
//	@Param			movieId		path		string	true	"movieId"
//	@Param			link		query		string	true	"link to torrent file or magnet link"
//	@Success		200			{object}	model.DownloadingFile
//	@Failure		400,401,404	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/torrent/download/:movieId [put]
func (m *MovieHandler) DownloadTorrent(c *fiber.Ctx) error {
	movieId := c.Params("movieId", "")
	if movieId == "" || movieId == ":link" {
		return response.ResponseError(c, "Invalid movieId", fiber.StatusBadRequest)
	}
	link := c.Query("link", "")
	if link == "" || link == ":link" {
		return response.ResponseError(c, "Invalid torrent link", fiber.StatusBadRequest)
	}

	//jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	res, err := m.movieService.DownloadFile(movieId, link)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return response.ResponseError(c, "Torrent link not found in db", fiber.StatusNotFound)
		}
		if strings.HasPrefix(err.Error(), "File") {
			return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return response.ResponseOKWithData(c, res)
}

// TorrentStatus godoc
//
//	@Summary		Torrent Status
//	@Description	get downloading files and storage usage etc.
//	@Tags			Torrent-Download
//	@Success		200		{object}	model.TorrentStatusRes
//	@Failure		400,401	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/torrent/status [get]
func (m *MovieHandler) TorrentStatus(c *fiber.Ctx) error {
	downloadingFiles := m.movieService.GetDownloadingFiles()

	res := model.TorrentStatusRes{
		DownloadingFiles: downloadingFiles,
	}

	return response.ResponseOKWithData(c, res)
}
