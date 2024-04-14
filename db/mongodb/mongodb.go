package mongodb

import (
	"context"
	"downloader_torrent/configs"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDatabase struct {
	Db     *mongo.Database
	client *mongo.Client
}

var MONGODB *MongoDatabase

func NewDatabase() (*MongoDatabase, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opts := options.Client().ApplyURI(configs.GetConfigs().MongodbDatabaseUrl)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	MONGODB = &MongoDatabase{
		client: client,
		Db:     client.Database(configs.GetConfigs().MongodbDatabaseName),
	}
	return &MongoDatabase{
		client: client,
		Db:     client.Database(configs.GetConfigs().MongodbDatabaseName),
	}, nil
}

func (d *MongoDatabase) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := d.client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

func (d *MongoDatabase) GetDB() *mongo.Database {
	return d.Db
}
