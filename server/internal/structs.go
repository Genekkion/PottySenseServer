package internal

type Client struct {
	Id         int
	FirstName  string
	LastName   string
	Gender     string
	Urination  string
	Defecation string
	LastRecord string
}

type TO struct {
	Id               int
	Username         string
	FirstName        string
	LastName         string
	Telegram         string
    TelegramVerified int
	UserType         string
}
