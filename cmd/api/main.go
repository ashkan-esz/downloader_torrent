package main

import (
	"downloader_torrent/api"
	"downloader_torrent/configs"
	"downloader_torrent/db"
	"downloader_torrent/db/mongodb"
	"downloader_torrent/db/redis"
	"downloader_torrent/internal/handler"
	"downloader_torrent/internal/repository"
	"downloader_torrent/internal/service"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

// @title						Go Torrent
// @version					2.0
// @description				Torrent service of the downloader_api project.
// @termsOfService				http://swagger.io/terms/
// @contact.name				API Support
// @contact.url				http://www.swagger.io/support
// @contact.email				support@swagger.io
// @license.name				Apache 2.0
// @license.url				http://www.apache.org/licenses/LICENSE-2.0.html
// @host						download.movieTracker.site
// @BasePath					/
// @schemes					https
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
// @Accept						json
// @Produce					json
func main() {
	configs.LoadEnvVariables()

	err := sentry.Init(sentry.ClientOptions{
		Dsn:     configs.GetConfigs().SentryDns,
		Release: configs.GetConfigs().SentryRelease,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1,
		EnableTracing:    true,
		Debug:            true,
		AttachStacktrace: true,
		//BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
		//	if hint.Context != nil {
		//		if c, ok := hint.Context.Value(sentry.RequestContextKey).(*fiber.Ctx); ok {
		//			// You have access to the original Context if it panicked
		//			fmt.Println(utils.ImmutableString(c.Hostname()))
		//		}
		//	}
		//	fmt.Println(event)
		//	return event
		//},
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	go redis.ConnectRedis()

	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}

	mongoDB, err := mongodb.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize mongodb database connection: %s", err)
	}
	go configs.LoadDbConfigs(mongoDB.GetDB())

	//telegramMessageSvc := service.NewTelegramMessageService()

	torrentRep := repository.NewTorrentRepository(mongoDB.GetDB())
	torrentSvc := service.NewTorrentService(torrentRep)
	torrentHandler := handler.NewTorrentHandler(torrentSvc)

	streamSvc := service.NewStreamService(torrentRep)
	streamHandler := handler.NewStreamHandler(streamSvc)

	userRep := repository.NewUserRepository(dbConn.GetDB(), mongoDB.GetDB())
	userSvc := service.NewUserService(userRep)
	userHandler := handler.NewUserHandler(userSvc)

	adminRep := repository.NewAdminRepository(dbConn.GetDB(), mongoDB.GetDB())
	adminSvc := service.NewAdminService(userRep, adminRep)
	adminHandler := handler.NewAdminHandler(adminSvc)

	handlers := &api.Handlers{
		TorrentHandler: torrentHandler,
		StreamHandler:  streamHandler,
		UserHandler:    userHandler,
		AdminHandler:   adminHandler,
		UserRepo:       userRep,
	}

	api.InitRouter(handlers)
	api.Start("0.0.0.0:" + configs.GetConfigs().Port)
}
