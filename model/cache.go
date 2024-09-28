package model

type CachedUserData struct {
	UserId     int64  `gorm:"column:userId" json:"userId"`
	Username   string `gorm:"column:username" json:"username"`
	PublicName string `gorm:"column:publicName" json:"publicName"`
	//ProfileImages        []ProfileImageDataModel `gorm:"foreignKey:UserId;references:UserId;" json:"profileImages"`
	//NotificationSettings NotificationSettings    `gorm:"foreignKey:UserId;references:UserId;"`
	NotifTokens []string `json:"notifTokens"`
}

type CachedMovieData struct {
	MovieId  string `bson:"_id" json:"movieId"`
	RawTitle string `bson:"rawTitle" json:"rawTitle"`
	Type     string `bson:"type" json:"type"`
	Year     string `bson:"year" json:"year"`
	//Posters  []MoviePoster `bson:"posters" json:"posters"`
}
