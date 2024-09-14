package configs

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type ConfigStruct struct {
	Port                      string
	AccessTokenSecret         string
	RefreshTokenSecret        string
	WaitForRedisConnectionSec int
	RedisUrl                  string
	RedisPassword             string
	MongodbDatabaseUrl        string
	MongodbDatabaseName       string
	MainServerAddress         string
	ServerAddress             string
	CorsAllowedOrigins        []string
	SentryDns                 string
	SentryRelease             string
	PrintErrors               bool
	DontConvertMkv            bool
	//DbUrl                        string
	//RabbitMqUrl                  string
	//CloudStorageEndpoint         string
	//CloudStorageWebsiteEndpoint  string
	//CloudStorageAccessKey        string
	//CloudStorageSecretAccessKey  string
	//CloudStorageBucketNamePrefix string
	//Domain                       string
}

var configs = ConfigStruct{}

func GetConfigs() ConfigStruct {
	return configs
}

func LoadEnvVariables() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	configs.Port = os.Getenv("PORT")
	//configs.DbUrl = os.Getenv("POSTGRES_DATABASE_URL")
	configs.AccessTokenSecret = os.Getenv("ACCESS_TOKEN_SECRET")
	configs.RefreshTokenSecret = os.Getenv("REFRESH_TOKEN_SECRET")
	configs.RedisUrl = os.Getenv("REDIS_URL")
	configs.RedisPassword = os.Getenv("REDIS_PASSWORD")
	configs.MongodbDatabaseUrl = os.Getenv("MONGODB_DATABASE_URL")
	configs.MongodbDatabaseName = os.Getenv("MONGODB_DATABASE_NAME")
	configs.MainServerAddress = os.Getenv("MAIN_SERVER_ADDRESS")
	configs.WaitForRedisConnectionSec, _ = strconv.Atoi(os.Getenv("WAIT_REDIS_CONNECTION_SEC"))
	configs.CorsAllowedOrigins = strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), "---")
	for i := range configs.CorsAllowedOrigins {
		configs.CorsAllowedOrigins[i] = strings.TrimSpace(configs.CorsAllowedOrigins[i])
	}
	configs.SentryDns = os.Getenv("SENTRY_DNS")
	configs.SentryRelease = os.Getenv("SENTRY_RELEASE")
	configs.PrintErrors = os.Getenv("PRINT_ERRORS") == "true"
	configs.DontConvertMkv = os.Getenv("DONT_CONVERT_MKV") == "true"
	//configs.CloudStorageEndpoint = os.Getenv("CLOUAD_STORAGE_ENDPOINT")
	//configs.CloudStorageWebsiteEndpoint = os.Getenv("CLOUAD_STORAGE_WEBSITE_ENDPOINT")
	//configs.CloudStorageAccessKey = os.Getenv("CLOUAD_STORAGE_ACCESS_KEY")
	//configs.CloudStorageSecretAccessKey = os.Getenv("CLOUAD_STORAGE_SECRET_ACCESS_KEY")
	//configs.CloudStorageBucketNamePrefix = os.Getenv("BUCKET_NAME_PREFIX")
	configs.ServerAddress = os.Getenv("SERVER_ADDRESS")
	//configs.MainServerAddress = os.Getenv("MAIN_SERVER_ADDRESS")
	//configs.RabbitMqUrl = os.Getenv("RABBITMQ_URL")
	//configs.Domain = os.Getenv("DOMAIN")
}
