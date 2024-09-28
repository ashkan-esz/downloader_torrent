package service

import (
	"downloader_torrent/internal/repository"
	"time"
)

type IUserService interface {
}

type UserService struct {
	userRepo repository.IUserRepository
	timeout  time.Duration
}

func NewUserService(userRepo repository.IUserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		timeout:  time.Duration(2) * time.Second,
	}
}

//------------------------------------------
//------------------------------------------
