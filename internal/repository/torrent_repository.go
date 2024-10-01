package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ITorrentRepository interface {
	CheckTorrentLinkExist(movieId primitive.ObjectID, torrentUrl string) (*CheckTorrentLinkExistRes, error)
	SaveTorrentLocalLink(movieId primitive.ObjectID, movieType string, torrentUrl string, localUrl string, expireTime int64) error
	RemoveTorrentLocalLink(movieType string, localUrl string) error
	IncrementTorrentLinkDownload(movieType string, localUrl string) error
	GetAllSerialTorrentLocalLinks() (*mongo.Cursor, error)
	GetAllMovieTorrentLocalLinks() (*mongo.Cursor, error)
	FindSerialTorrentLinks(searchList []string) ([]string, error)
	FindMovieTorrentLinks(searchList []string) ([]string, error)
	GetTorrentAutoDownloaderLinks() ([]TorrentAutoDownloaderLinksRes, error)
	UpdateTorrentAutoDownloaderPullDownloadLink(movieId primitive.ObjectID, torrentUrl string) error
	UpdateTorrentAutoDownloaderPullRemoveLink(movieType string, torrentUrl string) error
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

type TorrentAutoDownloaderLinksRes struct {
	Id                   primitive.ObjectID `bson:"_id"`
	Title                string             `bson:"title"`
	Type                 string             `bson:"type"`
	DownloadTorrentLinks []string           `bson:"downloadTorrentLinks"`
	RemoveTorrentLinks   []string           `bson:"removeTorrentLinks"`
}

//------------------------------------------
//------------------------------------------

func (m *TorrentRepository) CheckTorrentLinkExist(movieId primitive.ObjectID, torrentUrl string) (*CheckTorrentLinkExistRes, error) {
	filter := bson.M{
		"_id": movieId,
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

	err := m.mongodb.
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

func (m *TorrentRepository) SaveTorrentLocalLink(movieId primitive.ObjectID, movieType string, torrentUrl string, localUrl string, expireTime int64) error {
	isMovie := strings.Contains(movieType, "movie")
	linkQuery := "seasons.episodes.torrentLinks.link"
	updateField := "seasons.$.episodes.$[].torrentLinks.$[item].localLink"
	updateField2 := "seasons.$.episodes.$[].torrentLinks.$[item].localLinkExpire"
	if isMovie {
		linkQuery = "qualities.torrentLinks.link"
		updateField = "qualities.$[].torrentLinks.$[item].localLink"
		updateField2 = "qualities.$[].torrentLinks.$[item].localLinkExpire"
	}

	filter := bson.M{
		"_id":     movieId,
		linkQuery: torrentUrl,
	}

	update := bson.M{
		"$set": bson.M{
			updateField:  localUrl,
			updateField2: expireTime,
		},
		"$pull": bson.M{
			"downloadTorrentLinks": torrentUrl,
		},
	}

	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"item.link": torrentUrl},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := m.mongodb.
		Collection("movies").
		UpdateOne(ctx, filter, update, opts)

	return err
}

func (m *TorrentRepository) RemoveTorrentLocalLink(movieType string, localUrl string) error {
	isMovie := strings.Contains(movieType, "movie")
	linkQuery := "seasons.episodes.torrentLinks.localLink"
	updateField := "seasons.$.episodes.$[].torrentLinks.$[item].localLink"
	updateField2 := "seasons.$.episodes.$[].torrentLinks.$[item].localLinkExpire"
	if isMovie {
		linkQuery = "qualities.torrentLinks.localLink"
		updateField = "qualities.$[].torrentLinks.$[item].localLink"
		updateField2 = "qualities.$[].torrentLinks.$[item].localLinkExpire"
	}

	filter := bson.M{
		linkQuery: localUrl,
	}

	update := bson.M{
		"$set": bson.M{
			updateField:  "",
			updateField2: 0,
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

//------------------------------------------
//------------------------------------------

func (m *TorrentRepository) GetAllSerialTorrentLocalLinks() (*mongo.Cursor, error) {
	pipeline := mongo.Pipeline{
		{
			{"$unwind",
				bson.D{
					{"path", "$seasons"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		{
			{"$unwind",
				bson.D{
					{"path", "$seasons.episodes"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		{
			{"$unwind",
				bson.D{
					{"path", "$seasons.episodes.torrentLinks"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		{{"$match", bson.D{{"seasons.episodes.torrentLinks.localLink", bson.D{{"$ne", ""}}}}}},
		{{"$project", bson.D{{"localLink", "$seasons.episodes.torrentLinks.localLink"}}}},
	}

	// Execute the aggregation
	cursor, err := m.mongodb.
		Collection("movies").
		Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	return cursor, nil
}

func (m *TorrentRepository) GetAllMovieTorrentLocalLinks() (*mongo.Cursor, error) {
	pipeline := mongo.Pipeline{
		{
			{"$unwind",
				bson.D{
					{"path", "$qualities"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		{
			{"$unwind",
				bson.D{
					{"path", "$qualities.torrentLinks"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		{{"$match", bson.D{{"qualities.torrentLinks.localLink", bson.D{{"$ne", ""}}}}}},
		{{"$project", bson.D{{"localLink", "$qualities.torrentLinks.localLink"}}}},
	}

	// Execute the aggregation
	cursor, err := m.mongodb.
		Collection("movies").
		Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	return cursor, nil
}

func (m *TorrentRepository) FindSerialTorrentLinks(searchList []string) ([]string, error) {
	pipeline := mongo.Pipeline{
		{
			{"$unwind",
				bson.D{
					{"path", "$seasons"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		{
			{"$unwind",
				bson.D{
					{"path", "$seasons.episodes"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		{
			{"$unwind",
				bson.D{
					{"path", "$seasons.episodes.torrentLinks"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		{{"$match", bson.D{{"seasons.episodes.torrentLinks.localLink", bson.D{{"$in", searchList}}}}}},
		{{"$project", bson.D{{"localLink", "$seasons.episodes.torrentLinks.localLink"}}}},
	}

	cursor, err := m.mongodb.Collection("movies").Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var localLinks []string

	var r []map[string]string
	err = cursor.All(context.TODO(), &r)
	if err != nil {
		return nil, err
	}

	for _, el := range r {
		localLinks = append(localLinks, el["localLink"])
	}

	return localLinks, nil
}

func (m *TorrentRepository) FindMovieTorrentLinks(searchList []string) ([]string, error) {
	pipeline := mongo.Pipeline{
		{
			{"$unwind",
				bson.D{
					{"path", "$qualities"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		{
			{"$unwind",
				bson.D{
					{"path", "$qualities.torrentLinks"},
					{"preserveNullAndEmptyArrays", false},
				},
			},
		},
		{{"$match", bson.D{{"qualities.torrentLinks.localLink", bson.D{{"$in", searchList}}}}}},
		{{"$project", bson.D{{"localLink", "$qualities.torrentLinks.localLink"}}}},
	}

	cursor, err := m.mongodb.Collection("movies").Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var localLinks []string

	var r []map[string]string
	err = cursor.All(context.TODO(), &r)
	if err != nil {
		return nil, err
	}

	for _, el := range r {
		localLinks = append(localLinks, el["localLink"])
	}

	return localLinks, nil
}

//------------------------------------------
//------------------------------------------

func (m *TorrentRepository) GetTorrentAutoDownloaderLinks() ([]TorrentAutoDownloaderLinksRes, error) {
	defer func() {
		if r := recover(); r != nil {
			// Convert the panic to an error
			fmt.Printf("recovered from panic: %v\n", r)
		}
	}()

	filter := bson.M{
		"$or": []bson.M{
			{"downloadTorrentLinks": bson.M{"$ne": bson.A{}}},
			{"removeTorrentLinks": bson.M{"$ne": bson.A{}}},
		},
	}

	projection := bson.M{
		"title":                1,
		"type":                 1,
		"update_date":          1,
		"downloadTorrentLinks": 1,
		"removeTorrentLinks":   1,
	}
	opts := options.Find().SetSort(bson.D{{"update_date", -1}}).SetProjection(projection)

	cursor, err := m.mongodb.
		Collection("movies").
		Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var r []bson.M
	err = cursor.All(context.TODO(), &r)
	if err != nil {
		return nil, err
	}

	finalRes := []TorrentAutoDownloaderLinksRes{}
	for _, el := range r {
		dl, err := bsonAToStringSlice(el["downloadTorrentLinks"].(bson.A))
		if err != nil {
			return nil, err
		}
		rl, err := bsonAToStringSlice(el["removeTorrentLinks"].(bson.A))
		if err != nil {
			return nil, err
		}

		finalRes = append(finalRes, TorrentAutoDownloaderLinksRes{
			Id:                   el["_id"].(primitive.ObjectID),
			Title:                el["title"].(string),
			Type:                 el["type"].(string),
			DownloadTorrentLinks: dl,
			RemoveTorrentLinks:   rl,
		})
	}

	return finalRes, nil
}

func (m *TorrentRepository) UpdateTorrentAutoDownloaderPullDownloadLink(movieId primitive.ObjectID, torrentUrl string) error {
	filter := bson.M{
		"_id":                  movieId,
		"downloadTorrentLinks": torrentUrl,
	}

	update := bson.M{
		"$pull": bson.M{
			"downloadTorrentLinks": torrentUrl,
		},
	}

	opts := options.Update()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.mongodb.
		Collection("movies").
		UpdateOne(ctx, filter, update, opts)

	return err
}

func (m *TorrentRepository) UpdateTorrentAutoDownloaderPullRemoveLink(movieType string, torrentUrl string) error {
	isMovie := strings.Contains(movieType, "movie")
	linkQuery := "seasons.episodes.torrentLinks.link"
	updateField := "seasons.$.episodes.$[].torrentLinks.$[item].localLink"
	updateField2 := "seasons.$.episodes.$[].torrentLinks.$[item].localLinkExpire"
	if isMovie {
		linkQuery = "qualities.torrentLinks.link"
		updateField = "qualities.$[].torrentLinks.$[item].localLink"
		updateField2 = "qualities.$[].torrentLinks.$[item].localLinkExpire"
	}

	filter := bson.M{
		linkQuery: torrentUrl,
	}

	update := bson.M{
		"$set": bson.M{
			updateField:  "",
			updateField2: 0,
		},
		"$pull": bson.M{
			"downloadTorrentLinks": torrentUrl,
		},
	}

	opts := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"item.link": torrentUrl},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.mongodb.
		Collection("movies").
		UpdateOne(ctx, filter, update, opts)

	return err
}

//------------------------------------------
//------------------------------------------

func bsonAToStringSlice(a bson.A) ([]string, error) {
	result := make([]string, len(a))
	for i, v := range a {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("element at index %d is not a string", i)
		}
		result[i] = str
	}
	return result, nil
}
