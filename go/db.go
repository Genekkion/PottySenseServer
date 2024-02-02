package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type SqliteStorage struct {
	filepath string
	db       *sql.DB
}

func newSqliteStorage() *SqliteStorage {
	filepath := os.Getenv("DATABASE_PATH")
	if filepath == "" {
		log.Println("missing DATABASE_PATH")
		log.Panic()
	}
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Println("db.go - newSqliteStorage()")
		log.Panic(err)
	}

	log.Println("sqlite connection successfully created")

	rows, err := db.Query("select * from Tofficers")
	if err != nil {
		log.Println("query issues")
		log.Panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var first_name string
		var last_name string
		var username string
		var password string
		var telegram string
        var userType string
		err = rows.Scan(&id, &first_name, &last_name,
			&username, &password, &telegram, &userType)
		log.Println("query->", id, first_name, last_name,
			username, password, telegram, userType)
	}

	return &SqliteStorage{
		filepath: filepath,
		db:       db,
	}
}

func (storage *SqliteStorage) close() {
	storage.db.Close()
}

func (storage *SqliteStorage) createAdmin(username string, password string) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(saltPassword(password)),
		bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error creating admin, please try again.")
		log.Panicln(err)
	}

	_, err = storage.db.Exec(
		`INSERT INTO TOfficers (username, password, type)
        VALUES ($1, $2, 'admin')`,
		toLowerCase(username), passwordHash)
	if err != nil {
		log.Println("Error creating admin, please try again.")
		log.Panicln(err)
	}
	log.Println("Successfully created admin account.")
}
