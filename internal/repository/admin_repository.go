package repository

import (
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type IAdminRepository interface {
}

type AdminRepository struct {
	db      *gorm.DB
	mongodb *mongo.Database
}

func NewAdminRepository(db *gorm.DB, mongodb *mongo.Database) *AdminRepository {
	return &AdminRepository{db: db, mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------
