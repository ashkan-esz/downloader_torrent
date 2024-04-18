package repository

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IMovieRepository interface {
	CheckTorrentLinkExist(movieId string, torrentUrl string) (*CheckTorrentLinkExistRes, error)
	SaveTorrentLocalLink(movieId string, movieType string, torrentUrl string, localUrl string) error
	RemoveTorrentLocalLink(movieType string, localUrl string) error
}

type MovieRepository struct {
	mongodb *mongo.Database
}

func NewMovieRepository(mongodb *mongo.Database) *MovieRepository {
	return &MovieRepository{mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------

type CheckTorrentLinkExistRes struct {
	Title string `bson:"title"`
	Type  string `bson:"type"`
}

//------------------------------------------
//------------------------------------------

func (m *MovieRepository) CheckTorrentLinkExist(movieId string, torrentUrl string) (*CheckTorrentLinkExistRes, error) {
	id, err := primitive.ObjectIDFromHex(movieId)
	if err != nil {
		return nil, err
	}

	var result CheckTorrentLinkExistRes
	opts := options.FindOne().SetProjection(bson.D{
		{"title", 1},
		{"type", 1},
	})
	err = m.mongodb.
		Collection("movies").
		FindOne(context.TODO(),
			bson.D{
				{"_id", id},
				{"$or", []interface{}{
					bson.D{{"seasons.episodes.torrentLinks.link", torrentUrl}},
					bson.D{{"qualities.torrentLinks.link", torrentUrl}},
				}},
			}, opts).
		Decode(&result)

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *MovieRepository) SaveTorrentLocalLink(movieId string, movieType string, torrentUrl string, localUrl string) error {
	id, err := primitive.ObjectIDFromHex(movieId)
	if err != nil {
		return err
	}

	linkQuery := "seasons.episodes.torrentLinks.link"
	if strings.Contains(movieType, "movie") {
		linkQuery = "qualities.torrentLinks.link"
	}
	filter := bson.D{
		{"_id", id},
		{linkQuery, torrentUrl},
	}
	update := bson.D{
		{"$set", bson.D{
			{"seasons.$.episodes.$[].torrentLinks.$[item].localLink", localUrl},
			{"qualities.$[].torrentLinks.$[item].localLink", localUrl},
		}},
	}
	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"item.link": torrentUrl},
		},
	})

	_, err = m.mongodb.
		Collection("movies").
		UpdateOne(context.TODO(), filter, update, opts)

	return err
}

func (m *MovieRepository) RemoveTorrentLocalLink(movieType string, localUrl string) error {
	linkQuery := "seasons.episodes.torrentLinks.localLink"
	if strings.Contains(movieType, "movie") {
		linkQuery = "qualities.torrentLinks.localLink"
	}
	filter := bson.D{
		{linkQuery, localUrl},
	}
	update := bson.D{
		{"$set", bson.D{
			{"seasons.$.episodes.$[].torrentLinks.$[item].localLink", ""},
			{"qualities.$[].torrentLinks.$[item].localLink", ""},
		}},
	}
	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"item.localLink": localUrl},
		},
	})

	_, err := m.mongodb.
		Collection("movies").
		UpdateOne(context.TODO(), filter, update, opts)

	return err
}
