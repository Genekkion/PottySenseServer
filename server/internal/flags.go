package internal

import (
	"database/sql"
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/genekkion/PottySenseServer/internal/globals"
	"github.com/genekkion/PottySenseServer/internal/utils"
	"github.com/xuri/excelize/v2"
)

// Parses command line flags
func ParseFlags(db *sql.DB) {
	globals.FLAG_VERBOSE = *flag.Bool("v", false, "Enables verbose mode for debugging")

	adminFlag := flag.String("a", "", "Creates a new admin user with the username provided. Must be used with the -p flag.")
	userFlag := flag.String("u", "", "Creates a new user with the username provided. Must be used with the -p flag.")
	passwordFlag := flag.String("p", "", "Password for user creation. Must be used with the -a or -u flag.")

	fileFlag := flag.String("c", "", "Parses the .xlsx file supplied for client entries and saves to database.")

	flag.Parse()

	if *passwordFlag != "" {
		if *adminFlag != "" && *userFlag != "" {
			log.Println("Only one user can be created at a time using the -a and -u flags. Skipping operation.")
		} else if *adminFlag != "" {
			utils.CreateAdmin(db, *adminFlag, *passwordFlag)
		} else if *userFlag != "" {
			utils.CreateUser(db, *userFlag, *passwordFlag)
		} else {
			log.Println("Password flag -p needs to be used with either -a or -u to create user. Skipping operation.")
		}
		globals.RUN = false
	}

	if *fileFlag != "" {
		ParseFile(*fileFlag, db)
		globals.RUN = false
	}

}

func ParseFile(filePath string, db *sql.DB) {

	file, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Panicln(err)
	}

	rows, err := file.GetRows("Sheet1")
	if err != nil {
		log.Panicln(err)
	}
	file.Close()

	log.Println("Beginning transaction.")
	_, err = db.Exec("BEGIN TRANSACTION")
	if err != nil {
		log.Println("Error beginning transaction.")
		log.Println(err)
	}
	for i, row := range rows[1:] {
		for j, colCell := range row {
			if colCell == "" {
				log.Printf("Missing data at Row %d : Col %c, aborting.\n", i+1, j+'A')
				return
			}
		}

		client := Client{
			FirstName:  row[0],
			LastName:   row[1],
			Gender:     strings.ToLower(row[2]),
			Urination:  row[3],
			Defecation: row[4],
		}

		if !(client.Gender == "male" || client.Gender == "female") {
			log.Printf("Incorrect format for gender at Row %d, aborting.\n", i+1)
			return
		}
		uri, err := strconv.Atoi(client.Urination)
		if err != nil {
			log.Printf("Error parsing urination value at Row %d, aborting.\n", i+1)
			return
		}
		defec, err := strconv.Atoi(client.Defecation)
		if err != nil {
			log.Printf("Error parsing defecation value at Row %d, aborting.\n", i+1)
			return
		}

		_, err = db.Exec(
			`INSERT INTO Clients
            (first_name, last_name,
            gender, urination, defecation)
            VALUES ($1, $2, $3, $4, $5)`,
			client.FirstName, client.LastName,
			client.Gender, uri, defec)
		if err != nil {
			log.Printf("Error saving Row %d, aborting.\n", i+1)
			return
		}
	}
	_, err = db.Exec("COMMIT TRANSACTION")
	if err != nil {
		log.Printf("Error committing transaction.")
		return

	}
	db.Close()
	log.Println("Transaction completed, exiting.")
}
