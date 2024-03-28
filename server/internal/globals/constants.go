package globals

const (
	// Cookie naming scheme
	COOKIE_NAME            = "PS-cookie"
	COOKIE_TO_ID           = "id"             // int
	COOKIE_TO_USERNAME     = "username"       // string
	COOKIE_TO_TELE_CHAT_ID = "telegramChatId" // string
	COOKIE_TO_USER_TYPE    = "userType"       // string

	// File structures
	BASE_TEMPLATE = "./templates/base.html"

	// The threshold for when to display the
	// last record for toilet usage
	LAST_RECORD_THRESHOLD = 6 // in hours
)
