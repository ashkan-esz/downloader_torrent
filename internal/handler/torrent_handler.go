package handler

import (
	"downloader_torrent/internal/service"
	"downloader_torrent/model"
	"downloader_torrent/pkg/response"
	"errors"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type ITorrentHandler interface {
	DownloadTorrent(c *fiber.Ctx) error
	CancelDownload(c *fiber.Ctx) error
	RemoveDownload(c *fiber.Ctx) error
	TorrentStatus(c *fiber.Ctx) error
}

type TorrentHandler struct {
	torrentService service.ITorrentService
}

func NewTorrentHandler(torrentService service.ITorrentService) *TorrentHandler {
	return &TorrentHandler{
		torrentService: torrentService,
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
func (m *TorrentHandler) DownloadTorrent(c *fiber.Ctx) error {
	movieId := c.Params("movieId", "")
	if movieId == "" || movieId == ":link" {
		return response.ResponseError(c, "Invalid movieId", fiber.StatusBadRequest)
	}
	link := c.Query("link", "")
	if link == "" || link == ":link" {
		return response.ResponseError(c, "Invalid torrent link", fiber.StatusBadRequest)
	}

	//jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)
	res, err := m.torrentService.DownloadFile(movieId, link)
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

// CancelDownload godoc
//
//	@Summary		Cancel Download
//	@Description	cancel downloading torrent file.
//	@Tags			Torrent-Download
//	@Param			filename	path		string	true	"filename"
//	@Success		200			{object}	response.ResponseOKModel
//	@Failure		400,401		{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/torrent/cancel/:filename [put]
func (m *TorrentHandler) CancelDownload(c *fiber.Ctx) error {
	filename := c.Params("filename", "")
	if filename == "" || filename == ":filename" {
		return response.ResponseError(c, "Invalid filename", fiber.StatusBadRequest)
	}

	err := m.torrentService.CancelDownload(filename)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return response.ResponseOK(c, "")
}

// RemoveDownload godoc
//
//	@Summary		Remove Download
//	@Description	remove downloaded torrent file.
//	@Tags			Torrent-Download
//	@Param			filename	path		string	true	"filename"
//	@Success		200			{object}	response.ResponseOKModel
//	@Failure		400,401		{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/torrent/remove/:filename [delete]
func (m *TorrentHandler) RemoveDownload(c *fiber.Ctx) error {
	filename := c.Params("filename", "")
	if filename == "" || filename == ":filename" {
		return response.ResponseError(c, "Invalid filename", fiber.StatusBadRequest)
	}

	err := m.torrentService.RemoveDownload(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return response.ResponseError(c, "File not found", fiber.StatusNotFound)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return response.ResponseOK(c, "")
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
func (m *TorrentHandler) TorrentStatus(c *fiber.Ctx) error {
	downloadingFiles := m.torrentService.GetDownloadingFiles()
	localFiles := m.torrentService.GetLocalFiles()

	res := model.TorrentStatusRes{
		DownloadingFiles: downloadingFiles,
		LocalFiles:       localFiles,
	}

	return response.ResponseOKWithData(c, res)
}
