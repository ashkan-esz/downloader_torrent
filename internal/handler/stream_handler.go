package handler

import (
	"downloader_torrent/internal/service"
	"downloader_torrent/pkg/response"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type IStreamHandler interface {
	StreamMedia(c *fiber.Ctx) error
}

type StreamHandler struct {
	streamService service.IStreamService
}

func NewStreamHandler(streamService service.IStreamService) *StreamHandler {
	return &StreamHandler{
		streamService: streamService,
	}
}

//------------------------------------------
//------------------------------------------

// StreamMedia godoc
//
//	@Summary		Stream Media
//	@Description	stream downloaded file.
//	@Tags			Stream-Media
//	@Success		200		{object}	response.ResponseOKModel
//	@Failure		400,401	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/stream/play/:filename [get]
func (m *StreamHandler) StreamMedia(c *fiber.Ctx) error {
	filename := c.Params("filename", "")
	if filename == "" || filename == ":filename" {
		return response.ResponseError(c, "Invalid filename", fiber.StatusBadRequest)
	}

	filePath := "./downloads/" + filename

	// Open the video file
	file, err := os.Open(filePath)
	if err != nil {
		errorMessage := fmt.Sprintf("Error opening video file: %v", err)
		return response.ResponseError(c, errorMessage, fiber.StatusInternalServerError)
	}
	defer file.Close()

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting file information: %v", err)
		return response.ResponseError(c, errorMessage, fiber.StatusInternalServerError)
	}

	// get the file mime informations
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))

	// get file size
	fileSize := fileInfo.Size()

	rangeResult, err := c.Range(int(fileSize))
	if err != nil {
		errorMessage := fmt.Sprintf("Invalid Range Header: %v", err)
		return response.ResponseError(c, errorMessage, fiber.StatusInternalServerError)
	}

	if len(rangeResult.Ranges) != 0 {
		start := int64(rangeResult.Ranges[0].Start)
		end := int64(rangeResult.Ranges[0].End)

		// Setting required response headers
		c.Set(fiber.HeaderContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		c.Set(fiber.HeaderContentLength, strconv.FormatInt(end-start+1, 10))
		c.Set(fiber.HeaderContentType, mimeType)
		c.Set(fiber.HeaderAcceptRanges, "bytes")
		c.Status(fiber.StatusPartialContent)

		// Seek to the start position
		_, seekErr := file.Seek(start, io.SeekStart)
		if seekErr != nil {
			errorMessage := fmt.Sprintf("Error seeking to start position: %v", seekErr)
			return response.ResponseError(c, errorMessage, fiber.StatusInternalServerError)
		}

		// Copy the specified range of bytes to the response
		_, copyErr := io.CopyN(c.Response().BodyWriter(), file, end-start+1)
		if copyErr != nil {
			errorMessage := fmt.Sprintf("Error copying bytes to response: %v", copyErr)
			return response.ResponseError(c, errorMessage, fiber.StatusInternalServerError)
		}
	} else {
		// If no Range header is present, serve the entire video
		c.Set("Content-Length", strconv.FormatInt(fileSize, 10))

		_, copyErr := io.Copy(c.Response().BodyWriter(), file)
		if copyErr != nil {
			errorMessage := fmt.Sprintf("Error copying entire file to response: %v", copyErr)
			return response.ResponseError(c, errorMessage, fiber.StatusInternalServerError)
		}
	}

	return nil
}
