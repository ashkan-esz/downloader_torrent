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
	UserRepo  repository.IUserRepository
	AdminRepo repository.IAdminRepository
}

func NewAdminService(UserRepo repository.IUserRepository, AdminRepo repository.IAdminRepository) *AdminService {
	service := &AdminService{
		UserRepo:  UserRepo,
		AdminRepo: AdminRepo,
	}

	return service
}

//-----------------------------------------
//-----------------------------------------

func (m *AdminService) FetchDbConfigs() error {
	return configs.FetchMongoDbConfigs(mongodb.MONGODB.GetDB())
}
