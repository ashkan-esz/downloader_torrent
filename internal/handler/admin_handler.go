package handler

import (
	"downloader_torrent/internal/service"
	"downloader_torrent/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type IAdminHandler interface {
	FetchDbConfigs(c *fiber.Ctx) error
}

type AdminHandler struct {
	adminService service.IAdminService
}

func NewAdminHandler(adminService service.IAdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

//------------------------------------------
//------------------------------------------

// FetchDbConfigs godoc
//
//	@Summary		Fetch Configs
//	@Description	Reload db configs and dynamic configs.
//	@Tags			Admin
//	@Success		200		{object}	response.ResponseOKModel
//	@Failure		400,401	{object}	response.ResponseErrorModel
//	@Security		BearerAuth
//	@Router			/v1/admin/fetch_configs [get]
func (m *AdminHandler) FetchDbConfigs(c *fiber.Ctx) error {
	err := m.adminService.FetchDbConfigs()
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusInternalServerError)
	}

	return response.ResponseOK(c, "")
}
