package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ITorrentRepository interface {
	CheckTorrentLinkExist(movieId string, torrentUrl string) (*CheckTorrentLinkExistRes, error)
	SaveTorrentLocalLink(movieId string, movieType string, torrentUrl string, localUrl string) error
	RemoveTorrentLocalLink(movieType string, localUrl string) error
	IncrementTorrentLinkDownload(movieType string, localUrl string) error
}

type TorrentRepository struct {
	mongodb *mongo.Database
}

func NewTorrentRepository(mongodb *mongo.Database) *TorrentRepository {
	return &TorrentRepository{mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------

type CheckTorrentLinkExistRes struct {
	Title string `bson:"title"`
	Type  string `bson:"type"`
}

//------------------------------------------
//------------------------------------------

func (m *TorrentRepository) CheckTorrentLinkExist(movieId string, torrentUrl string) (*CheckTorrentLinkExistRes, error) {
	id, err := primitive.ObjectIDFromHex(movieId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": id,
		"$or": []bson.M{
			{"seasons.episodes.torrentLinks.link": torrentUrl},
			{"qualities.torrentLinks.link": torrentUrl},
		},
	}

	projection := bson.M{
		"title": 1,
		"type":  1,
	}

	var result CheckTorrentLinkExistRes
	opts := options.FindOne().SetProjection(projection)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = m.mongodb.
		Collection("movies").
		FindOne(ctx, filter, opts).
		Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func (m *TorrentRepository) SaveTorrentLocalLink(movieId string, movieType string, torrentUrl string, localUrl string) error {
	id, err := primitive.ObjectIDFromHex(movieId)
	if err != nil {
		return err
	}

	isMovie := strings.Contains(movieType, "movie")
	linkQuery := "seasons.episodes.torrentLinks.link"
	updateField := "seasons.$.episodes.$[].torrentLinks.$[item].localLink"
	if !isMovie {
		linkQuery = "qualities.torrentLinks.link"
		updateField = "qualities.$[].torrentLinks.$[item].localLink"
	}

	filter := bson.M{
		"_id":     id,
		linkQuery: torrentUrl,
	}

	update := bson.M{
		"$set": bson.M{
			updateField: localUrl,
		},
	}

	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"item.link": torrentUrl},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err = m.mongodb.
		Collection("movies").
		UpdateOne(ctx, filter, update, opts)

	return err
}

func (m *TorrentRepository) RemoveTorrentLocalLink(movieType string, localUrl string) error {
	isMovie := strings.Contains(movieType, "movie")
	linkQuery := "seasons.episodes.torrentLinks.localLink"
	updateField := "seasons.$.episodes.$[].torrentLinks.$[item].localLink"
	if isMovie {
		linkQuery = "qualities.torrentLinks.localLink"
		updateField = "qualities.$[].torrentLinks.$[item].localLink"
	}

	filter := bson.M{
		linkQuery: localUrl,
	}

	update := bson.M{
		"$set": bson.M{
			updateField: "",
		},
	}

	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"item.localLink": localUrl},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.mongodb.
		Collection("movies").
		UpdateOne(ctx, filter, update, opts)

	return err
}

func (m *TorrentRepository) IncrementTorrentLinkDownload(movieType string, localUrl string) error {
	isMovie := strings.Contains(movieType, "movie")
	linkQuery := "seasons.episodes.torrentLinks.localLink"
	counterQuery := "seasons.$.episodes.$[].torrentLinks.$[item].downloadsCount"
	if isMovie {
		linkQuery = "qualities.torrentLinks.localLink"
		counterQuery = "qualities.$[].torrentLinks.$[item].downloadsCount"
	}

	filter := bson.M{
		linkQuery: localUrl,
	}

	update := bson.M{
		"$inc": bson.M{
			counterQuery: 1,
		},
	}

	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"item.localLink": localUrl},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := m.mongodb.
		Collection("movies").
		UpdateOne(ctx, filter, update, opts)

	if err == nil && res != nil && res.MatchedCount == 0 && movieType == "serial" {
		return m.IncrementTorrentLinkDownload("movie", localUrl)
	}

	return err
}
