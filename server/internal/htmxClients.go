package internal

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
)

func (server *Server) htmxClients(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "GET" {
		writeJson(writer, http.StatusBadRequest,
			map[string]interface{}{
				"error": "invalid request method",
			},
		)
		return
	}

	_, err := server.secureHtmx(writer, request)
	if err != nil {
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/clients.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
	})
}

type ClientEntry struct {
	Client     Client
	IsAssigned bool
}

func (server *Server) htmxClientEntry(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "POST" {
		writeJson(writer, http.StatusBadRequest,
			map[string]interface{}{
				"error": "invalid request method",
			},
		)
		return
	}

	to, err := server.secureHtmx(writer, request)
	if err != nil {
		return
	}

	searchQuery := request.FormValue("search") + "%"
	db := server.dbStorage
	rows, err := db.Query(
		`SELECT Clients.id, Clients.first_name, Clients.last_name,
        Clients.gender, Clients.urination, Clients.defecation,
        Clients.last_record, Watch.to_id 
        FROM Clients LEFT JOIN Watch
        ON Clients.id = Watch.client_id
        AND Watch.to_id = $1
        WHERE 
        first_name LIKE $2 COLLATE NOCASE OR
        last_name LIKE $3 COLLATE NOCASE`,
		to.Id, searchQuery, searchQuery)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		log.Println(err)
		return
	}
	var entries []ClientEntry
	// caser := cases.Title(language.English)
	for rows.Next() {
		var client Client
		var checkTo sql.NullInt32
		//var urination int
		//var defecation int
		rows.Scan(&client.Id, &client.FirstName, &client.LastName,
			&client.Gender, &client.Urination, &client.Defecation,
			&client.LastRecord, &checkTo)
		/*
			client.Urination = fmt.Sprintf("%02d:%02d",
				urination/60, urination%60)
			client.Defecation = fmt.Sprintf("%02d:%02d",
				defecation/60, defecation%60)
		*/
		isAssigned := checkTo.Valid

		// client.FirstName = caser.String(client.FirstName)
		// client.LastName = caser.String(client.LastName)
		startTime, err := time.Parse(time.RFC3339, client.LastRecord)
		if err != nil {
			writeJson(writer, http.StatusInternalServerError,
				map[string]string{
					"error": "internal server error",
				},
			)
			return
		}
		currentTime := time.Now()
		elapsedTime := currentTime.Sub(startTime)
		if elapsedTime.Hours() <= 6 {
			client.LastRecord = fmt.Sprintf("%02d:%02d",
				int(elapsedTime.Hours()), int(elapsedTime.Minutes())%60)
		} else {
			client.LastRecord = "nil"
		}
		entries = append(entries, ClientEntry{
			Client:     client,
			IsAssigned: isAssigned,
		})
	}
	tmpl := template.Must(template.ParseFiles("./templates/htmx/clientEntry.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"entries":        entries,
	})
}

func (server *Server) htmxClientAssign(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "POST" {
		writeJson(writer, http.StatusBadRequest,
			map[string]interface{}{
				"error": "invalid request method",
			},
		)
		return
	}

	to, err := server.secureHtmx(writer, request)
	if err != nil {
		return
	}

	clientId := request.FormValue("clientId")
	toAssign := request.FormValue("toAssign")
	tmpl := template.Must(template.ParseFiles("./templates/htmx/clientEntryButton.html"))
	db := server.dbStorage
	if toAssign == "false" {
		_, err = db.Exec(
			`DELETE FROM Watch
            WHERE to_id = $1 AND client_id = $2`,
			to.Id, clientId)
		if err != nil {
			writeJson(writer, http.StatusInternalServerError,
				map[string]string{
					"error": "internal server error",
				},
			)
			return
		}

		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
			"clientId":       clientId,
			"isAssigned":     false,
		})
		return
	}

	_, err = db.Exec(
		`INSERT OR IGNORE
        INTO Watch (to_id, client_id)
        VALUES ($1, $2)`, to.Id, clientId)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"clientId":       clientId,
		"isAssigned":     true,
	})
}

func (server *Server) htmxClientNewHandler(writer http.ResponseWriter,
	request *http.Request) {
	_, err := server.secureHtmx(writer, request)
	if err != nil {
		return
	}

	if request.Method == "GET" {
		htmxClientNewGet(writer, request, server)
		return
	} else if request.Method == "POST" {
		htmxClientNewPost(writer, request, server)
		return
	}
	writeJson(writer, http.StatusBadRequest,
		map[string]interface{}{
			"error": "invalid request method",
		},
	)
}

func htmxClientNewGet(writer http.ResponseWriter,
	request *http.Request, _ *Server) {
	tmpl := template.Must(template.ParseFiles("./templates/htmx/clientNew.html"))
	tmpl.Execute(writer,
		map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
		},
	)

}

func htmxClientNewPost(writer http.ResponseWriter,
	request *http.Request, server *Server) {
	firstName := request.FormValue("firstName")
	lastName := request.FormValue("lastName")
	gender := request.FormValue("gender")

	db := server.dbStorage
	_, err := db.Exec(
		`INSERT INTO Clients 
        (first_name, last_name, gender)
        VALUES ($1, $2, $3)`,
		firstName, lastName, gender)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}
	writer.Header().Add("HX-Trigger", "newClient")
	writeJson(writer, http.StatusCreated, nil)
}
