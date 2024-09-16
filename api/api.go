package api

import (
	"context"
	"downloader_torrent/api/middleware"
	"downloader_torrent/configs"
	_ "downloader_torrent/docs"
	"downloader_torrent/internal/handler"
	services "downloader_torrent/internal/service"
	"downloader_torrent/pkg/response"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/contrib/fibersentry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/gofiber/template/html/v2"
)

var router *fiber.App

func InitRouter(torrentHandler *handler.TorrentHandler, streamHandler *handler.StreamHandler) {
	var defaultErrorHandler = func(c *fiber.Ctx, err error) error {
		// Status code defaults to 500
		code := fiber.StatusInternalServerError

		// Retrieve the custom status code if it's a *fiber.Error
		var e *fiber.Error
		if errors.As(err, &e) {
			code = e.Code
		}

		// Set Content-Type: text/plain; charset=utf-8
		c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)

		if !strings.Contains(err.Error(), "/favicon.ico") && code >= 500 {
			fmt.Println(err.Error())
		}

		// Return status code with error message
		//return c.Status(code).SendString(err.Error())
		return response.ResponseError(c, "Internal Error", code)
	}

	engine := html.New("./templates", ".tpl")
	router = fiber.New(fiber.Config{
		UnescapePath: true,
		BodyLimit:    100 * 1024 * 1024,
		ErrorHandler: defaultErrorHandler,
		Views:        engine,
	})

	router.Use(helmet.New())
	router.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return middleware.LocalhostRegex.MatchString(origin) ||
				slices.Index(configs.GetConfigs().CorsAllowedOrigins, origin) != -1
		},
		AllowCredentials: true,
	}))
	router.Use(timeoutMiddleware(time.Second * 10))
	router.Use(recover.New())
	// router.Use(logger.New())
	router.Use(compress.New())

	router.Use(fibersentry.New(fibersentry.Config{
		Repanic:         true,
		WaitForDelivery: false,
	}))

	router.Static("/direct_download", "downloads", fiber.Static{
		Compress:      false,
		ByteRange:     true,
		Browse:        true,
		Download:      false,
		Index:         "",
		CacheDuration: 0,
		MaxAge:        3600,
		ModifyResponse: func(c *fiber.Ctx) error {
			filenames := strings.Split(c.Path(), "/")
			if services.TorrentSvc != nil {
				services.TorrentSvc.DecrementFileDownloadCount(filenames[len(filenames)-1])
			}
			return nil
		},
		Next: func(c *fiber.Ctx) bool {
			filenames := strings.Split(c.Path(), "/")
			filename := filenames[len(filenames)-1]
			if !services.TorrentSvc.CheckServingLocalFile(filename) {
				return true
			}
			if services.TorrentSvc != nil {
				_ = services.TorrentSvc.IncrementFileDownloadCount(filename)
			}
			return false
		},
	})

	//router.Use("/downloads", filesystem.New(filesystem.Config{
	//	Root:   http.Dir("./downloads"),
	//	Browse: true,
	//	//Index:        "index.html",
	//	//NotFoundFile: "404.html",
	//	MaxAge: 3600,
	//}))

	// This allows clients to request specific parts of a file, which is useful for resumable downloads or streaming media.
	router.Get("/partial_download/:filename", torrentHandler.ServeLocalFile)

	torrentRoutes := router.Group("v1/torrent")
	{
		torrentRoutes.Put("/download/:movieId", middleware.AuthMiddleware, torrentHandler.DownloadTorrent)
		torrentRoutes.Put("/cancel/:filename", middleware.AuthMiddleware, torrentHandler.CancelDownload)
		torrentRoutes.Delete("/remove/:filename", middleware.AuthMiddleware, torrentHandler.RemoveDownload)
		torrentRoutes.Get("/status", middleware.AuthMiddleware, torrentHandler.TorrentStatus)
	}

	streamRoutes := router.Group("v1/stream")
	{
		streamRoutes.Get("/status", middleware.AuthMiddleware, streamHandler.StreamStatus)
		streamRoutes.Get("/:filename", func(c *fiber.Ctx) error {
			filename := c.Params("filename", "")
			noConversion := c.QueryBool("noConversion", false)
			crf := c.QueryInt("crf", 30)
			return c.Render("index", fiber.Map{
				"Filename":     filename,
				"NoConversion": noConversion,
				"Crf":          crf,
			})
		})
		streamRoutes.Get("/play/:filename", streamHandler.StreamMedia)
	}

	router.Get("/", HealthCheck)
	router.Get("/metrics", monitor.New())

	router.Get("/swagger/*", swagger.HandlerDefault) // default
}

func Start(addr string) error {
	return router.Listen(addr)
}

func timeoutMiddleware(timeout time.Duration) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		// wrap the request context with a timeout
		ctx, cancel := context.WithTimeout(c.Context(), timeout)

		defer func() {
			// check if context timeout was reached
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {

				// write response and abort the request
				//c.Writer.WriteHeader(fiber.StatusGatewayTimeout)
				c.SendStatus(fiber.StatusGatewayTimeout)
				//c.Abort()
			}

			//cancel to clear resources after finished
			cancel()
		}()

		// replace request with context wrapped request
		//c.Request = c.Request.WithContext(ctx)
		return c.Next()
	}
}

// HealthCheck godoc
//
//	@Summary		Show the status of server.
//	@Description	get the status of server.
//	@Tags			System
//	@Success		200	{object}	map[string]interface{}
//	@Router			/ [get]
func HealthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	if err := c.JSON(res); err != nil {
		return err
	}

	return nil
}
