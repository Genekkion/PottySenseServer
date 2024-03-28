package utils

import (
	"database/sql"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// filepath := os.Getenv("DATABASE_PATH")
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
		err = rows.Scan(
			&id,
			&first_name,
			&last_name,
			&username,
			&password,
			&telegram,
			&userType,
		)
		if err != nil {
			log.Fatalln(err)
		}
	}

	rows.Close()
}

func CreateAdmin(db *sql.DB,
	username string, password string) {
	log.Println("HERER")
	log.Println(username)
	log.Println(password)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(
		SaltPassword(password)),
		bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error creating admin, please try again.")
		log.Fatalln(err)
	}
	var id sql.NullInt32
	err = db.QueryRow(
		`SELECT id FROM TOFFICERS WHERE username = $1`,
		strings.ToLower(username),
	).Scan(&id)
	if err == nil {
		log.Println("Another account with this username already exists. Please use another username.")
		return
	} else if err != sql.ErrNoRows {
		log.Println("Error creating admin, please try again.")
		log.Fatalln(err)
	}

	_, err = db.Exec(
		`INSERT INTO TOfficers (username, password, type)
        VALUES ($1, $2, 'admin')`,
		strings.ToLower(username),
		string(passwordHash))
	if err != nil {
		log.Println("Error creating admin, please try again.")
		log.Fatalln(err)
	}
	log.Println("Successfully created admin account.")
}

func CreateUser(db *sql.DB,
	username string, password string) {

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(SaltPassword(password)),
		bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error creating user, please try again.")
		log.Fatalln(err)
	}

	var id sql.NullInt32
	err = db.QueryRow(
		`SELECT id FROM TOFFICERS WHERE username = $1`,
		strings.ToLower(username),
	).Scan(&id)
	if err == nil {
		log.Println("Another account with this username already exists. Please use another username.")
		return
	} else if err != sql.ErrNoRows {
		log.Println("Error creating user, please try again.")
		log.Fatalln(err)
	}

	_, err = db.Exec(
		`INSERT INTO TOfficers (username, password, type)
		VALUES ($1, $2, 'admin')`,
		strings.ToLower(username),
		passwordHash)
	if err != nil {
		log.Println("Error creating user, please try again.")
		log.Fatalln(err)
	}
	log.Println("Successfully created user account.")
}

func SaltPassword(password string) string {
	return "cS46O" + password + "$1aY"
}
