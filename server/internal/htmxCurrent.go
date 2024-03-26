package internal

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
)

func (server *Server) htmxCurrent(writer http.ResponseWriter,
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
	tmpl := template.Must(template.ParseFiles("./templates/htmx/current.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
	})
}
func (server *Server) htmxCurrentClients(writer http.ResponseWriter,
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
	// "SELECT * FROM Watch WHERE to_id = $1;"
	db := server.dbStorage
	rows, err := db.Query(
		`SELECT Clients.id, Clients.first_name, Clients.last_name,
        Clients.gender, Clients.urination, Clients.defecation, Clients.last_record
        FROM Watch INNER JOIN Clients
        ON Watch.client_id = Clients.id
        WHERE Watch.to_id = $1`,
		to.Id)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}
	var clients []Client
	// caser := cases.Title(language.English)
	for rows.Next() {
		var client Client
		var urination int
		var defecation int

		rows.Scan(&client.Id, &client.FirstName, &client.LastName,
			&client.Gender, &urination, &defecation,
			&client.LastRecord)
		client.Urination = fmt.Sprintf("%02d:%02d",
			urination/60, urination%60)
		client.Defecation = fmt.Sprintf("%02d:%02d",
			defecation/60, defecation%60)

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

		// client.FirstName = caser.String(client.FirstName)
		// client.LastName = caser.String(client.LastName)

		clients = append(clients, client)
	}
	tmpl := template.Must(template.ParseFiles("./templates/htmx/currentEntry.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"clients":        clients,
	})
}
