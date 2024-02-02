package main

type Client struct {
	Id         int
	FirstName  string
	LastName   string
	Gender     string
	Urination  int
	Defecation int
	LastRecord string
	ToId       int
}

type TO struct {
	Id        int
	Username  string
	FirstName string
	LastName  string
	Telegram  string
	UserType  string
}
