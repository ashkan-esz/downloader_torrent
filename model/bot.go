package model

type Bot struct {
	BotId                     string    `gorm:"column:botId;type:text;not null;uniqueIndex:Bot_botId_key;primaryKey;" json:"botId"`
	BotName                   string    `gorm:"column:botName;type:text;not null;" json:"botName"`
	BotType                   string    `gorm:"column:botType;type:text;not null;" json:"botType"`
	Disabled                  bool      `gorm:"column:disabled;type:boolean;not null;" json:"disabled"`
	IsOfficial                bool      `gorm:"column:isOfficial;type:boolean;not null;" json:"isOfficial"`
	PermissionToLogin         bool      `gorm:"column:permissionToLogin;type:boolean;not null;" json:"permissionToLogin"`
	PermissionToCrawl         bool      `gorm:"column:permissionToCrawl;type:boolean;not null;" json:"permissionToCrawl"`
	PermissionToTorrentLeech  bool      `gorm:"column:permissionToTorrentLeech;type:boolean;not null;" json:"permissionToTorrentLeech"`
	PermissionToTorrentSearch bool      `gorm:"column:permissionToTorrentSearch;type:boolean;not null;" json:"permissionToTorrentSearch"`
	BotToken                  string    `gorm:"column:botToken;type:text;not null;uniqueIndex:Bot_botToken_key;" json:"botToken"`
	Users                     []UserBot `gorm:"foreignKey:BotId;references:BotId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Bot) TableName() string {
	return "Bot"
}

//------------------------------------
//------------------------------------

type UserBot struct {
	UserId       int64  `gorm:"column:userId;type:integer;not null;uniqueIndex:UserBot_userId_botId_idx;"`
	BotId        string `gorm:"column:botId;type:text;not null;uniqueIndex:UserBot_userId_botId_idx;"`
	Username     string `gorm:"column:username;type:text;not null;"`
	ChatId       string `gorm:"column:chatId;type:text;not null;"`
	Notification bool   `gorm:"column:notification;type:boolean;default:true;not null;"`
}

func (UserBot) TableName() string {
	return "UserBot"
}

//------------------------------------
//------------------------------------

type UserBotDataModel struct {
	UserId       int64  `gorm:"column:userId;type:integer;not null;"`
	BotId        string `gorm:"column:botId;type:text;not null;"`
	Username     string `gorm:"column:username;type:text;not null;"`
	ChatId       string `gorm:"column:chatId;type:text;not null;"`
	Notification bool   `gorm:"column:notification;type:boolean;default:true;not null;"`
}

func (UserBotDataModel) TableName() string {
	return "UserBot"
}
