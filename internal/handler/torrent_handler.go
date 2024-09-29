package handler

import (
	"downloader_torrent/internal/service"
	"downloader_torrent/model"
	"downloader_torrent/pkg/response"
	"downloader_torrent/util"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type ITorrentHandler interface {
	ServeLocalFile(c *fiber.Ctx) error
	DownloadTorrent(c *fiber.Ctx) error
	CancelDownload(c *fiber.Ctx) error
	RemoveDownload(c *fiber.Ctx) error
	ExtendLocalFileExpireTime(c *fiber.Ctx) error
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

// ServeLocalFile godoc
//
//	@Summary		serve file
//	@Description	serve files downloaded from torrent
//	@Tags			Serve-Files
//	@Param			filename	path		string	true	"filename"
//	@Param			Range		header		string	true	"download/stream range"
//	@Success		200			{object}	response.ResponseOKModel
//	@Failure		400,401,404	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/partial_download/:filename [get]
func (m *TorrentHandler) ServeLocalFile(c *fiber.Ctx) error {
	filename := c.Params("filename", "")
	if filename == "" || filename == ":filename" {
		return response.ResponseError(c, "Invalid filename", fiber.StatusBadRequest)
	}

	//return c.SendFile("./downloads/" + filename)

	if !m.torrentService.CheckServingLocalFile(filename) {
		return response.ResponseError(c, model.ErrTorrentFilesServingDisabled.Error(), fiber.StatusServiceUnavailable)
	}

	if !m.torrentService.CheckConcurrentServingLimit() {
		return response.ResponseError(c, "Server is busy", fiber.StatusServiceUnavailable)
	}

	file, err := os.Open("./downloads/" + filename)
	if err != nil {
		if os.IsNotExist(err) {
			return response.ResponseError(c, "File not found", fiber.StatusNotFound)
		}
		errorMessage := fmt.Sprintf("Error opening file [%v]: %v", filename, err)
		return response.ResponseError(c, errorMessage, fiber.StatusInternalServerError)
	}
	defer file.Close()

	_ = m.torrentService.IncrementFileDownloadCount(filename)
	defer m.torrentService.DecrementFileDownloadCount(filename)

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	rangeHeader := c.Get("Range")
	if rangeHeader != "" {
		// Parse range header and serve partial content
		// Implement range parsing logic here
	}

	// Serve full file if no range header
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Length", strconv.FormatInt(fileSize, 10))
	err = c.SendFile(file.Name())
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}
	return nil
}

// DownloadTorrent godoc
//
//	@Summary		Download Torrent
//	@Description	download from torrent into local storage
//	@Tags			Torrent-Download
//	@Param			movieId		path		string	true	"movieId"
//	@Param			link		query		string	true	"link to torrent file or magnet link"
//	@Param			downloadNow	query		boolean	true	"starts downloading now. need permission 'admin_manage_torrent' to work"
//	@Success		200			{object}	model.DownloadRequestRes
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
	downloadNow := c.QueryBool("downloadNow", false)

	permissions := c.Locals("permissions").([]string)
	jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)

	info := &model.DownloadRequestInfo{
		MovieId:      movieId,
		TorrentUrl:   link,
		IsAdmin:      slices.Contains(permissions, "admin_manage_torrent"),
		DownloadNow:  downloadNow,
		UserId:       jwtUserData.UserId,
		IsBotRequest: jwtUserData.IsBotRequest,
		BotId:        jwtUserData.BotId,
		ChatId:       jwtUserData.ChatId,
		BotUsername:  jwtUserData.BotUsername,
	}

	res, err := m.torrentService.HandleDownloadTorrentRequest(info)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return response.ResponseError(c, "Torrent link not found in db", fiber.StatusNotFound)
		}
		if strings.HasPrefix(err.Error(), "File") {
			return response.ResponseError(c, err.Error(), fiber.StatusBadRequest)
		}
		if errors.Is(err, model.ErrAlreadyDownloading) {
			return response.ResponseError(c, err.Error(), fiber.StatusConflict)
		}

		code := model.GetErrorCode(err)
		if code != 0 {
			return response.ResponseError(c, err.Error(), code)
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

// ExtendLocalFileExpireTime godoc
//
//	@Summary		Extend Expire Time
//	@Description	add time to expiration of local files
//	@Tags			Torrent-Download
//	@Param			filename	path		string	true	"filename"
//	@Success		200			{object}	response.ResponseOKWithDataModel
//	@Failure		400,401,404	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/torrent/extend_expire_time/:filename [put]
func (m *TorrentHandler) ExtendLocalFileExpireTime(c *fiber.Ctx) error {
	filename := c.Params("filename", "")
	if filename == "" || filename == ":filename" {
		return response.ResponseError(c, "Invalid filename", fiber.StatusBadRequest)
	}

	newTime, err := m.torrentService.ExtendLocalFileExpireTime(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return response.ResponseError(c, "File not found", fiber.StatusNotFound)
		}
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return response.ResponseOKWithData(c, newTime)
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
	res := m.torrentService.GetTorrentStatus()

	return response.ResponseOKWithData(c, res)
}
