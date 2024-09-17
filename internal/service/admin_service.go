package service

import (
	"downloader_torrent/configs"
	"downloader_torrent/db/mongodb"
	"downloader_torrent/internal/repository"
)

type IAdminService interface {
	FetchDbConfigs() error
}

type AdminService struct {
	AdminRepo repository.IAdminRepository
}

func NewAdminService(AdminRepo repository.IAdminRepository) *AdminService {
	service := &AdminService{
		AdminRepo: AdminRepo,
	}

	return service
}

//-----------------------------------------
//-----------------------------------------

func (m *AdminService) FetchDbConfigs() error {
	return configs.FetchMongoDbConfigs(mongodb.MONGODB.GetDB())
}
