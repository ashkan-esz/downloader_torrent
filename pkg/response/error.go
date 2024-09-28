package response

const (
	ServerError = "Server error, try again later"
	//----------------------
	ExceedProfileImage = "Exceeded profile image counts"
	ExceedGenres       = "Exceeded number of genres limit (6)"
	//----------------------
	MovieSourcesNotFound      = "Movie sources not found"
	CrawlerSourceNotFound     = "Crawler source not found"
	CrawlerSourceAlreadyExist = "Crawler source already exist"
	MoviesNotFound            = "Movies not found"
	MovieNotFound             = "Movie not found"
	GenresNotFound            = "Genres not found"
	DocumentsNotFound         = "Documents not found"
	DocumentNotFound          = "Document not found"
	StaffNotFound             = "Staff not found"
	CharacterNotFound         = "Character not found"
	MscNotFound               = "Movie/Staff/Character not found"
	ScNotFound                = "Staff/Character not found"
	BotNotFound               = "Bot not found"
	MessageNotFound           = "Message not found"
	AppNotFound               = "App not found"
	ConfigsDbNotFound         = "Configs from database not found"
	JobNotFound               = "job not found"
	CantRemoveCurrentOrigin   = "Cannot remove current origin from corsAllowedOrigins"
	//----------------------
	UserNotFound         = "Cannot find user"
	SessionNotFound      = "Cannot find session"
	ProfileImageNotFound = "Cannot find profile image"
	EmailNotFound        = "Cannot find user email"
	//----------------------
	InvalidRefreshToken = "Invalid RefreshToken"
	InvalidToken        = "Invalid/Stale Token"
	InvalidDeviceId     = "Invalid deviceId"
	//----------------------
	UserPassNotMatch = "Username and password do not match"
	OldPassNotMatch  = "Old password does not match"
	//----------------------
	BadRequestBody = "Incorrect request body"
	//----------------------
	UserIdAlreadyExist   = "This userId already exists"
	UsernameAlreadyExist = "This username already exists"
	EmailAlreadyExist    = "This email already exists"
	AlreadyExist         = "Already exist"
	AlreadyFollowed      = "Already followed"
	//----------------------
	BotIsDisabled = "This bot is disabled"
	//----------------------
)
