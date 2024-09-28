package repository

import (
	"downloader_torrent/model"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type IUserRepository interface {
	GetUserBots(userId int64) ([]model.UserBotDataModel, error)
	GetBotData(botId string) (*model.Bot, error)
	GetUserRoles(userId int64) ([]model.Role, error)
	GetUserRolesWithPermissions(userId int64) ([]model.RoleWithPermissions, error)
	GetUserPermissionsByRoleIds(roleIds []int64) ([]model.Permission, error)
}

type UserRepository struct {
	db      *gorm.DB
	mongodb *mongo.Database
}

func NewUserRepository(db *gorm.DB, mongodb *mongo.Database) *UserRepository {
	return &UserRepository{db: db, mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) GetUserBots(userId int64) ([]model.UserBotDataModel, error) {
	var result []model.UserBotDataModel
	err := r.db.
		Model(&model.UserBotDataModel{}).
		Where("\"userId\" = ? AND notification = true", userId).
		Find(&result).
		Error
	return result, err
}

func (r *UserRepository) GetBotData(botId string) (*model.Bot, error) {
	var result model.Bot
	err := r.db.
		Model(&model.Bot{}).
		Where("\"botId\" = ?", botId).
		Find(&result).
		Error
	return &result, err
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) GetUserRoles(userId int64) ([]model.Role, error) {
	var roles []model.Role

	// Fetch the roles for this user
	if err := r.db.Model(&model.Role{}).
		Joins("JOIN \"UserToRole\" ON \"UserToRole\".\"roleId\" = \"Role\".id").
		Where("\"UserToRole\".\"userId\" = ?", userId).
		Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *UserRepository) GetUserRolesWithPermissions(userId int64) ([]model.RoleWithPermissions, error) {
	roles := []model.RoleWithPermissions{}

	type resType struct {
		model.Role
		model.Permission
	}
	var res []resType

	queryStr := `
		SELECT *
		FROM "UserToRole" ur
        	JOIN "Role" r ON ur."roleId" = r.id
        	JOIN "RoleToPermission" rp ON r.id = rp."roleId"
        	JOIN "Permission" p ON rp."permissionId" = p.id
		WHERE
			ur."userId" = @uid;`

	err := r.db.Raw(queryStr,
		map[string]interface{}{
			"uid": userId,
		}).
		Scan(&res).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t := make([]model.RoleWithPermissions, 0)
			return t, nil
		}
		return nil, err
	}

	for _, item := range res {
		exist := false
		for i2 := range roles {
			if roles[i2].Id == item.Role.Id {
				roles[i2].Permissions = append(roles[i2].Permissions, item.Permission)
				exist = true
			}
		}
		if exist {
			continue
		}
		newRole := model.RoleWithPermissions{
			Id:                  item.Role.Id,
			Name:                item.Role.Name,
			Description:         item.Role.Description,
			TorrentLeachLimitGb: item.Role.TorrentLeachLimitGb,
			TorrentSearchLimit:  item.Role.TorrentSearchLimit,
			BotsNotification:    item.Role.BotsNotification,
			CreatedAt:           item.Role.CreatedAt,
			UpdatedAt:           item.Role.UpdatedAt,
			Permissions:         []model.Permission{item.Permission},
		}
		roles = append(roles, newRole)
	}

	return roles, nil
}

func (r *UserRepository) GetUserPermissionsByRoleIds(roleIds []int64) ([]model.Permission, error) {
	var permissions []model.Permission

	if err := r.db.Model(&model.Permission{}).
		Joins("JOIN \"RoleToPermission\" ON \"RoleToPermission\".\"permissionId\" = \"Permission\".id").
		Where("\"RoleToPermission\".\"roleId\" in ?", roleIds).
		Find(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

//------------------------------------------
//------------------------------------------
