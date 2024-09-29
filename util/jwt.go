package util

import (
	"downloader_torrent/configs"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type MyJwtClaims struct {
	UserId       int64   `json:"userId"`
	Username     string  `json:"username"`
	RoleIds      []int64 `json:"roleIds"`
	GeneratedAt  int64   `json:"generatedAt"`
	ExpiresAt    int64   `json:"expiresAt"`
	IsBotRequest bool    `json:"isBotRequest"`
	jwt.RegisteredClaims
}

type TokenDetail struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

func VerifyToken(tokenString string) (*jwt.Token, *MyJwtClaims, error) {
	claims := MyJwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("wrong signature method")
		}
		return []byte(configs.GetConfigs().AccessTokenSecret), nil
	})

	if err != nil {
		return nil, nil, err
	}

	return token, &claims, nil
}

func VerifyRefreshToken(tokenString string) (*jwt.Token, *MyJwtClaims, error) {
	claims := MyJwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("wrong signature method")
		}
		return []byte(configs.GetConfigs().RefreshTokenSecret), nil
	})

	if err != nil {
		return nil, nil, err
	}

	return token, &claims, nil
}
