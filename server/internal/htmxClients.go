package internal

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/genekkion/PottySenseServer/internal/globals"
	"github.com/genekkion/PottySenseServer/internal/utils"
	"github.com/gorilla/csrf"
)

// /htmx/clients
func (server *Server) htmxClients(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		server.htmxClientsPanel(writer, request)
	case http.MethodPost:
		server.htmxClientSearch(writer, request)
	case http.MethodPut:
		server.htmxClientTrack(writer, request)
	default:
		genericMethodNotAllowedReply(writer)
	}
}

// /htmx/clients "GET"
func (server *Server) htmxClientsPanel(writer http.ResponseWriter,
	request *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/htmx/clients.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
	})

}

// /htmx/clients "POST"
func (server *Server) htmxClientSearch(writer http.ResponseWriter,
	request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		log.Println("htmxClientSearch() - parse form")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	to := server.getTOFromCookie(request)

	// Add wildcard for autocomplete
	searchQuery := request.FormValue("search") + "%"
	rows, err := server.db.Query(
		`SELECT Clients.id, Clients.first_name,
			Clients.last_name, Clients.gender,
			Clients.urination, Clients.defecation,
        	Clients.last_record, Watch.to_id 
        FROM Clients LEFT JOIN Watch
        	ON Clients.id = Watch.client_id
        		AND Watch.to_id = $1
        WHERE first_name LIKE $2 COLLATE NOCASE
			OR last_name LIKE $2 COLLATE NOCASE
		`, to.Id, searchQuery)
	if err != nil {
		log.Println("htmxClientSearch() - db query")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	type ClientEntry struct {
		Client     Client
		IsAssigned bool
	}

	var entries []ClientEntry
	for rows.Next() {
		var client Client
		// Needs to be nullable in case the TO
		// is NOT watching a particular client
		var checkTo sql.NullInt32

		rows.Scan(
			&client.Id,
			&client.FirstName,
			&client.LastName,
			&client.Gender,
			&client.Urination,
			&client.Defecation,
			&client.LastRecord,
			&checkTo,
		)

		isAssigned := checkTo.Valid

		if time.Since(client.LastRecord).Hours() < globals.LAST_RECORD_THRESHOLD {
			client.PrettyLastRecord = utils.GetTimeElapsedPretty(client.LastRecord)
		} else {
			client.PrettyLastRecord = "nil"
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

// /htmx/clients "PUT"
func (server *Server) htmxClientTrack(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	clientId := request.FormValue("clientId")
	// TODO: change to toTrack from toAssign
	toTrack := request.FormValue("toTrack")
	tmpl := template.Must(template.ParseFiles("./templates/htmx/clientEntryButton.html"))

	// This means to remove tracking
	if toTrack == "false" {
		_, err := server.db.Exec(
			`DELETE FROM Watch
            WHERE to_id = $1
				AND client_id = $2
			`, to.Id, clientId)

		if err != nil {
			log.Println("htmxClientTrack() - db delete")
			log.Println(err)
			genericInternalServerErrorReply(writer)
			return
		}

		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
			"clientId":       clientId,
			//"isAssigned":     false,
			"isTracking": false,
		})
		return
	}

	_, err := server.db.Exec(
		`INSERT OR IGNORE
        INTO Watch (to_id, client_id)
        VALUES ($1, $2)`, to.Id, clientId)
	if err != nil {
		log.Println("htmxClientTrack() - db insert")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"clientId":       clientId,
		//"isAssigned":     true,
		"isTracking": true,
	})
}

// /htmx/clients/new
func (server *Server) htmxClientNewHandler(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		htmxClientNewForm(writer, request, server)
	case http.MethodPost:
		htmxClientNewSave(writer, request, server)
	default:
		genericMethodNotAllowedReply(writer)
	}
}

// /htmx/clients/new "GET"
// Responds with the form to add new client
// TODO: Change to modal
func htmxClientNewForm(writer http.ResponseWriter,
	request *http.Request, _ *Server) {
	tmpl := template.Must(template.ParseFiles("./templates/htmx/clientNew.html"))
	tmpl.Execute(writer,
		map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
		},
	)

}

// /htmx/clients/new "POST"
func htmxClientNewSave(writer http.ResponseWriter,
	request *http.Request, server *Server) {
	err := request.ParseForm()
	if err != nil {
		log.Println("htmxClientNewPost() - parse form")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	firstName := request.FormValue("firstName")
	lastName := request.FormValue("lastName")
	gender := request.FormValue("gender")
	// TODO: add more form values for urination, defecation

	_, err = server.db.Exec(
		`INSERT INTO Clients 
        	(first_name, last_name, gender)
        VALUES ($1, $2, $3)`,
		firstName, lastName, gender)
	if err != nil {
		log.Println("htmxClientNewPost() - db insert")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}
	writer.Header().Add("HX-Trigger", "newClient")
	writeJson(writer, http.StatusCreated, nil)
}
