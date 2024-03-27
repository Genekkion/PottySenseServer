package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func NewSqliteStorage(filepath string) *sql.DB {
	if filepath == "" {
		log.Fatalln("missing filepath for sqlite storage")
	}
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Println("db.go - newSqliteStorage()")
		log.Panic(err)
	}

	testDB(db)
	log.Println("Sqlite connection successfully created.")

	return db
}

func testDB(db *sql.DB) {
	rows, err := db.Query("select * from Tofficers")
	if err != nil {
		log.Println("query issues")
		log.Fatalln(err)
	}

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
	}

	rows.Close()
}
