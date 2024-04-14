package middleware

import (
	"downloader_torrent/db/redis"
	"downloader_torrent/pkg/response"
	"downloader_torrent/util"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refreshToken", "")
	if refreshToken == "" {
		refreshToken = c.Get("refreshtoken", "")
		if refreshToken == "" {
			refreshToken = c.Get("refreshToken", "")
		}
	}

	if refreshToken == "" {
		return response.ResponseError(c, "Unauthorized, refreshToken not provided", fiber.StatusUnauthorized)
	}

	result, err := redis.GetRedis(c.Context(), "jwtKey:"+refreshToken)
	if result != "" && err != nil && err.Error() != "redis: nil" {
		return response.ResponseError(c, "Unauthorized, refreshToken is in blacklist", fiber.StatusUnauthorized)
	}

	token, claims, err := util.VerifyRefreshToken(refreshToken)
	if err != nil {
		return response.ResponseError(c, "Unauthorized, Invalid refreshToken", fiber.StatusUnauthorized)
	}
	if token == nil || claims == nil {
		return response.ResponseError(c, "Unauthorized, Invalid refreshToken metaData", fiber.StatusUnauthorized)
	}

	//--------------------------------
	//--------------------------------

	accessToken := c.Get("Authorization", "")
	strArr := strings.Split(accessToken, " ")
	if len(strArr) == 2 {
		accessToken = strArr[1]
	} else if len(strArr) == 0 || len(accessToken) < 30 {
		return response.ResponseError(c, "Unauthorized, Invalid accessToken", fiber.StatusUnauthorized)
	}

	token2, claims2, err := util.VerifyToken(accessToken)
	if err != nil {
		return response.ResponseError(c, "Unauthorized, Invalid accessToken", fiber.StatusUnauthorized)
	}
	if token2 == nil || claims2 == nil {
		return response.ResponseError(c, "Unauthorized, Invalid accessToken metaData", fiber.StatusUnauthorized)
	}

	//todo : check role is admin

	c.Locals("refreshToken", refreshToken)
	c.Locals("accessToken", accessToken)
	c.Locals("jwtUserData", claims2)
	return c.Next()
}

func CORSMiddleware(c *fiber.Ctx) error {
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Credentials", "true")
	c.Set("Access-Control-Allow-Headers", "Content-Type,Content-Length")
	c.Set("Access-Control-Allow-Method", "POST, GET, DELETE, PUT")

	if c.Method() == "OPTIONS" {
		//c.AbortWithStatus(204)
		return c.SendStatus(204)
		//return
	}

	return c.Next()
}
