package configs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DbConfigData struct {
	Id                                  primitive.ObjectID `bson:"_id"`
	Title                               string             `bson:"title"`
	CorsAllowedOrigins                  []string           `bson:"corsAllowedOrigins"`
	DisableTestUserRequests             bool               `bson:"disableTestUserRequests"`
	DisableCrawlerForDuration           int                `bson:"disableCrawlerForDuration"`
	DisableCrawlerStart                 int                `bson:"disableCrawlerStart"`
	CrawlerDisabled                     bool               `bson:"crawlerDisabled"`
	DisableCrawler                      bool               `bson:"disableCrawler"`
	DevelopmentFaze                     bool               `bson:"developmentFaze"`
	DevelopmentFazeStart                int                `bson:"developmentFazeStart"`
	MediaFileSizeLimit                  int64              `bson:"mediaFileSizeLimit"`
	ProfileFileSizeLimit                int64              `bson:"profileFileSizeLimit"`
	ProfileImageCountLimit              int64              `bson:"profileImageCountLimit"`
	MediaFileExtensionLimit             string             `bson:"mediaFileExtensionLimit"`
	ProfileImageExtensionLimit          string             `bson:"profileImageExtensionLimit"`
	TorrentDownloadMaxSpaceSize         int64              `bson:"torrentDownloadMaxSpaceSize"`
	TorrentDownloadMaxFileSize          int64              `bson:"torrentDownloadMaxFileSize"`
	TorrentDownloadSpaceThresholdSize   int64              `bson:"torrentDownloadSpaceThresholdSize"`
	TorrentFilesExpireHour              int64              `bson:"torrentFilesExpireHour"`
	TorrentFilesServingConcurrencyLimit int64              `bson:"torrentFilesServingConcurrencyLimit"`
	TorrentDownloadTimeoutMin           int64              `bson:"torrentDownloadTimeoutMin"`
}

var rwm sync.RWMutex
var dbConfigs DbConfigData

func GetDbConfigs() DbConfigData {
	rwm.RLock()
	defer rwm.RUnlock()
	return dbConfigs
}

func LoadDbConfigs(mongodb *mongo.Database) {
	tick := time.NewTicker(15 * time.Minute)
	load(mongodb)
	defer rwm.Unlock()
	for range tick.C {
		load(mongodb)
	}
}

func load(mongodb *mongo.Database) {
	rwm.Lock()
	defer rwm.Unlock()
	err := mongodb.
		Collection("configs").
		FindOne(context.Background(), bson.D{{"title", "server configs"}}).
		Decode(&dbConfigs)
	if err != nil {
		errorMessage := fmt.Sprintf("could not get dbConfig from mongodb: %s", err)
		if configs.PrintErrors {
			log.Println(errorMessage)
		}
		sentry.CaptureException(err)
		//log.Fatalf("could not get dbConfig from mongodb: %s", err)
	}
}
