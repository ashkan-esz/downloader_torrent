package service

import (
	"context"
	"downloader_torrent/db/redis"
	"downloader_torrent/model"
	errorHandler "downloader_torrent/pkg/error"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ICacheService interface {
	GetJwtDataCache(key string) (string, error)
	setJwtDataCache(key string, value string, duration time.Duration) error
	GetRolePermissionsCache(roleIds []int64) ([]string, error)
	SetRolePermissionsCache(roleIds []int64, permissions []string, duration time.Duration) error
	getCachedBotData(botId string) (*model.Bot, error)
	getCachedMultiBotData(botIds []string) ([]model.Bot, error)
	setBotDataCache(botId string, botData *model.Bot)
}

const (
	jwtDataCachePrefix         = "jwtKey:"
	userDataCachePrefix        = "user:"
	movieDataCachePrefix       = "movie:"
	botDataCachePrefix         = "bot:"
	rolePermissionsCachePrefix = "roleIds:"
)

//------------------------------------------
//------------------------------------------

func GetJwtDataCache(key string) (string, error) {
	result, err := redis.GetRedis(context.Background(), jwtDataCachePrefix+key)
	return result, err
}

func setJwtDataCache(key string, value string, duration time.Duration) error {
	err := redis.SetRedis(context.Background(), jwtDataCachePrefix+key, value, duration)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving jwt: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}
	return err
}

//------------------------------------------
//------------------------------------------

func GetRolePermissionsCache(roleIds []int64) ([]string, error) {
	key := int64SliceToString(roleIds, ",")
	result, err := redis.GetRedis(context.Background(), rolePermissionsCachePrefix+key)
	if err != nil && err.Error() != "redis: nil" {
		return nil, nil
	}
	if result != "" {
		var jsonData []string
		err = json.Unmarshal([]byte(result), &jsonData)
		if err != nil {
			return nil, err
		}
		return jsonData, nil
	}
	return nil, err
}

func SetRolePermissionsCache(roleIds []int64, permissions []string, duration time.Duration) error {
	jsonData, err := json.Marshal(permissions)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving permissions: %v", err)
		errorHandler.SaveError(errorMessage, err)
		return err
	}

	key := int64SliceToString(roleIds, ",")
	err = redis.SetRedis(context.Background(), rolePermissionsCachePrefix+key, jsonData, duration)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving permissions: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}
	return err
}

//------------------------------------------
//------------------------------------------

func getCachedBotData(botId string) (*model.Bot, error) {
	result, err := redis.GetRedis(context.Background(), botDataCachePrefix+botId)
	if err != nil && err.Error() != "redis: nil" {
		return nil, nil
	}
	if result != "" {
		var jsonData model.Bot
		err = json.Unmarshal([]byte(result), &jsonData)
		if err != nil {
			return nil, err
		}
		return &jsonData, nil
	}
	return nil, err
}

func getCachedMultiBotData(botIds []string) ([]model.Bot, error) {
	keys := make([]string, len(botIds))
	for i, id := range botIds {
		keys[i] = botDataCachePrefix + id
	}

	result, err := redis.MGetRedis(context.Background(), keys)
	if err != nil && err.Error() != "redis: nil" {
		return nil, nil
	}
	if result != nil {
		cachedData := []model.Bot{}
		for i := range result {
			if result[i] == nil {
				//not found
				continue
			}
			var jsonData model.Bot
			err = json.Unmarshal([]byte(result[i].(string)), &jsonData)
			if err != nil {
				return nil, err
			}
			cachedData = append(cachedData, jsonData)
		}
		return cachedData, nil
	}
	return nil, err
}

func setBotDataCache(botId string, botData *model.Bot) error {
	jsonData, err := json.Marshal(botData)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving botData: %v", err)
		errorHandler.SaveError(errorMessage, err)
		return err
	}
	err = redis.SetRedis(context.Background(), botDataCachePrefix+botId, jsonData, 1*time.Hour)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving botData: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}
	return err
}

//------------------------------------------
//------------------------------------------

func int64SliceToString(nums []int64, delimiter string) string {
	// Create a string slice to hold the converted numbers
	strNums := make([]string, len(nums))

	// Convert each int64 to string
	for i, num := range nums {
		strNums[i] = strconv.FormatInt(num, 10)
	}

	// Join the strings with the specified delimiter
	return strings.Join(strNums, delimiter)
}
