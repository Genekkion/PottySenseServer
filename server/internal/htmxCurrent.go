package internal

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/genekkion/PottySenseServer/internal/globals"
	"github.com/genekkion/PottySenseServer/internal/utils"
	"github.com/gorilla/csrf"
)

// /htmx/current
func (server *Server) htmxTrackingHandler(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		server.htmxTrackingPanel(writer, request)
	case http.MethodPost:
		server.htmxTrackingLoad(writer, request)
	default:
		genericMethodNotAllowedReply(writer)

	}
}

// /htmx/current "GET"
func (server *Server) htmxTrackingPanel(writer http.ResponseWriter,
	request *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/htmx/current.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
	})
}

// /htmx/current "POST"
// Lazy loading for currently tracked clients
// NOTE: Potentiall combine with the panel
func (server *Server) htmxTrackingLoad(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	rows, err := server.db.Query(
		`SELECT Clients.id, Clients.first_name,
			Clients.last_name, Clients.gender,
			Clients.urination, Clients.defecation,
			Clients.last_record
        FROM Watch
		INNER JOIN Clients
        	ON Watch.client_id = Clients.id
        WHERE Watch.to_id = $1`,
		to.Id)
	if err != nil {
		log.Println("htmxTrackingLoad() - db query")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	var clients []Client

	for rows.Next() {
		var client Client

		rows.Scan(
			&client.Id,
			&client.FirstName,
			&client.LastName,
			&client.Gender,
			&client.Urination,
			&client.Defecation,
			&client.LastRecord,
		)

		if time.Since(client.LastRecord).Hours() < globals.LAST_RECORD_THRESHOLD {
			client.PrettyLastRecord = utils.GetTimeElapsedPretty(client.LastRecord)
		} else {
			client.PrettyLastRecord = "nil"
		}

		clients = append(clients, client)
	}
	tmpl := template.Must(template.ParseFiles("./templates/htmx/currentEntry.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"clients":        clients,
	})
}

