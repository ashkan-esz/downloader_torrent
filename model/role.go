package model

import "time"

// sync with :: https://github.com/ashkan-esz/downloader_api/blob/master/src/data/db/admin/roleAndPermissionsDbMethods.js

type UserToRole struct {
	UserId int64 `gorm:"column:userId;type:integer;primaryKey"`
	RoleId int64 `gorm:"column:roleId;type:integer;primaryKey"`
}

func (UserToRole) TableName() string {
	return "UserToRole"
}

//------------------------------------------
//------------------------------------------

type Role struct {
	Id                  int64     `gorm:"column:id;type:serial;autoIncrement;primaryKey;"`
	Name                string    `gorm:"column:name;type:text;not null;uniqueIndex:Role_name_key"`
	Description         string    `gorm:"column:description;type:text;default:\"\";not null;"`
	TorrentLeachLimitGb int       `gorm:"column:torrentLeachLimitGb;type:integer;not null;"`
	TorrentSearchLimit  int       `gorm:"column:torrentSearchLimit;type:integer;not null;"`
	BotsNotification    bool      `gorm:"column:botsNotification;type:boolean;not null;default:false;"`
	CreatedAt           time.Time `gorm:"column:createdAt;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	UpdatedAt           time.Time `gorm:"column:updatedAt;type:timestamp(3);not null;"`

	Users       []UserToRole       `gorm:"foreignKey:RoleId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Permissions []RoleToPermission `gorm:"foreignKey:RoleId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Role) TableName() string {
	return "Role"
}

//------------------------------------------
//------------------------------------------

type RoleToPermission struct {
	RoleId       int64 `gorm:"column:roleId;type:integer;primaryKey"`
	PermissionId int64 `gorm:"column:permissionId;type:integer;primaryKey"`
}

func (RoleToPermission) TableName() string {
	return "RoleToPermission"
}

//------------------------------------------
//------------------------------------------

type Permission struct {
	Id          int64     `gorm:"column:id;type:serial;autoIncrement;primaryKey;"`
	Name        string    `gorm:"column:name;type:text;not null;uniqueIndex:Permission_name_key"`
	Description string    `gorm:"column:description;type:text;default:\"\";not null;"`
	CreatedAt   time.Time `gorm:"column:createdAt;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	UpdatedAt   time.Time `gorm:"column:updatedAt;type:timestamp(3);not null;"`

	roles []RoleToPermission `gorm:"foreignKey:Id;references:PermissionId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Permission) TableName() string {
	return "Permission"
}

//------------------------------------------
//------------------------------------------

type RoleWithPermissions struct {
	Id                  int64     `gorm:"column:id;type:serial;autoIncrement;primaryKey;"`
	Name                string    `gorm:"column:name;type:text;not null;uniqueIndex:Role_name_key"`
	Description         string    `gorm:"column:description;type:text;default:\"\";not null;"`
	TorrentLeachLimitGb int       `gorm:"column:torrentLeachLimitGb;type:integer;not null;"`
	TorrentSearchLimit  int       `gorm:"column:torrentSearchLimit;type:integer;not null;"`
	BotsNotification    bool      `gorm:"column:botsNotification;type:boolean;not null;default:false;"`
	CreatedAt           time.Time `gorm:"column:createdAt;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	UpdatedAt           time.Time `gorm:"column:updatedAt;type:timestamp(3);not null;"`
	Permissions         []Permission
}

//------------------------------------------
//------------------------------------------

type UserRolePermissionRes struct {
	RolesWithPermissions []RoleWithPermissions `json:"rolesWithPermissions"`
	Roles                []Role                `json:"roles"`
}

//------------------------------------------
//------------------------------------------

type DefaultRoleId int64
type DefaultRoleName string

const (
	MainAdmin        DefaultRoleId   = 0
	DefaultAdmin     DefaultRoleId   = 1
	DefaultUser      DefaultRoleId   = 2
	TestUser         DefaultRoleId   = 3
	DefaultBot       DefaultRoleId   = 4
	MainAdminRole    DefaultRoleName = "main_admin_role"
	DefaultAdminRole DefaultRoleName = "default_admin_role"
	DefaultUserRole  DefaultRoleName = "default_user_role"
	TestUserRole     DefaultRoleName = "test_user_role"
	DefaultBotRole   DefaultRoleName = "default_bot_role"
)
