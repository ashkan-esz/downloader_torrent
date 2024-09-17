package repository

import "go.mongodb.org/mongo-driver/mongo"

type IAdminRepository interface {
}

type AdminRepository struct {
	mongodb *mongo.Database
}

func NewAdminRepository(mongodb *mongo.Database) *AdminRepository {
	return &AdminRepository{mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------
