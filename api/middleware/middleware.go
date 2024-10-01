package middleware

import (
	"downloader_torrent/internal/repository"
	services "downloader_torrent/internal/service"
	"downloader_torrent/pkg/response"
	"downloader_torrent/util"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(c *fiber.Ctx) error {
	isBotRequest := c.Get("isbotrequest", "") == "true"
	if !isBotRequest {
		isBotRequest = c.Get("isBotRequest", "") == "true"
	}

	if !isBotRequest {
		// bots dont send refreshToken
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

		result, err := services.GetJwtDataCache(refreshToken)
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

		c.Locals("refreshToken", refreshToken)
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

	if isBotRequest != claims2.IsBotRequest {
		return response.ResponseError(c, "Unauthorized, Conflict for jwt.isBotRequest !== headers.isBotRequest", fiber.StatusUnauthorized)
	}

	c.Locals("accessToken", accessToken)
	c.Locals("jwtUserData", claims2)
	return c.Next()
}

func IsAuthRefreshToken(c *fiber.Ctx) error {
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

	result, err := services.GetJwtDataCache(refreshToken)
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

	c.Locals("refreshToken", refreshToken)
	c.Locals("jwtUserData", claims)
	return c.Next()
}

func CheckUserPermission(userRepo *repository.UserRepository, needPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		jwtUserData := c.Locals("jwtUserData").(*util.MyJwtClaims)

		permissions, _ := services.GetRolePermissionsCache(jwtUserData.RoleIds)
		if permissions == nil || len(permissions) == 0 {
			permissionsModel, err := userRepo.GetUserPermissionsByRoleIds(jwtUserData.RoleIds)
			if err != nil {
				return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
			}

			permissions = []string{}
			for _, p := range permissionsModel {
				permissions = append(permissions, p.Name)
			}

			_ = services.SetRolePermissionsCache(jwtUserData.RoleIds, permissions, 1*time.Minute)
		}

		if !slices.Contains(permissions, needPermission) {
			message := fmt.Sprintf("Forbidden, need ([%v]) permissions", needPermission)
			return response.ResponseError(c, message, fiber.StatusForbidden)
		}

		c.Locals("permissions", permissions)
		return c.Next()
	}
}

var (
	LocalhostRegex = regexp.MustCompile(`(?i)^(https?://)?localhost(:\d{4})?$`)
)
