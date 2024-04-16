package api

import (
	"context"
	"downloader_torrent/api/middleware"
	_ "downloader_torrent/docs"
	"downloader_torrent/internal/handler"
	"downloader_torrent/pkg/response"
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/contrib/fibersentry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

var router *fiber.App

func InitRouter(movieHandler *handler.MovieHandler) {
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

		// Return status code with error message
		//return c.Status(code).SendString(err.Error())
		return response.ResponseError(c, "Internal Error", code)
	}

	router = fiber.New(fiber.Config{
		UnescapePath: true,
		BodyLimit:    100 * 1024 * 1024,
		ErrorHandler: defaultErrorHandler,
	})

	router.Use(helmet.New())
	router.Use(cors.New())
	router.Use(timeoutMiddleware(time.Second * 2))
	router.Use(recover.New())
	// router.Use(logger.New())
	router.Use(compress.New())

	router.Use(fibersentry.New(fibersentry.Config{
		Repanic:         true,
		WaitForDelivery: false,
	}))

	router.Use("/downloads", filesystem.New(filesystem.Config{
		Root:   http.Dir("./downloads"),
		Browse: true,
		//Index:        "index.html",
		//NotFoundFile: "404.html",
		MaxAge: 3600,
	}))

	userRoutes := router.Group("v1/torrent")
	{
		userRoutes.Put("/download/:movieId", middleware.CORSMiddleware, middleware.AuthMiddleware, movieHandler.DownloadTorrent)
		userRoutes.Put("/cancel/:filename", middleware.CORSMiddleware, middleware.AuthMiddleware, movieHandler.CancelDownload)
		userRoutes.Get("/status", middleware.CORSMiddleware, middleware.AuthMiddleware, movieHandler.TorrentStatus)
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
