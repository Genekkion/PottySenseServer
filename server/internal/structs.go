package internal

import "time"

type Client struct {
	Id               int
	FirstName        string
	LastName         string
	Gender           string
	Urination        int
	Defecation       int
	LastRecord       time.Time
	PrettyLastRecord string
}

type TO struct {
	Id             int
	Username       string
	FirstName      string
	LastName       string
	TelegramChatId string
	UserType       string
}
