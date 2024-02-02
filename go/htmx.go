package main

import (
	"github.com/gorilla/csrf"
	"html/template"
	"log"
	"net/http"
)

type htmxlRoute struct {
	urlSuffix string
	filePath  string
}

// for htmx routes, ensures that users are logged in
// returns username if logged in, empty string otherwise
func (server *Server) secureHtmx(writer http.ResponseWriter,
	request *http.Request) (int, string, error) {
	store := server.redisStorage.store
	session, err := store.Get(request, "PS-cookie")
	if err != nil {
		log.Println("htmx.go: htmxDashboard() - getSession")
		log.Println(err)
		writeJson(writer, http.StatusInternalServerError,
			map[string]interface{}{
				"error": "internal server error",
			},
		)
		return -1, "", err
	}

	id := session.Values["id"]
	username := session.Values["username"]
	if id == nil || username == nil {
		http.Redirect(writer, request, "/login", http.StatusSeeOther)
		return -1, "", nil
	}
	return id.(int), username.(string), nil
}

// /htmx/client/search
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

	id, username, err := server.secureHtmx(writer, request)
	if id == -1 || username == "" || err != nil {
		return
	}

	//searchQuery := "%" + request.FormValue("search") + "%"
	searchQuery := request.FormValue("search") + "%"
	db := server.dbStorage
	rows, err := db.db.Query(
		`SELECT * FROM Clients WHERE 
        first_name LIKE $1 COLLATE NOCASE OR
        last_name LIKE $2 COLLATE NOCASE`,
		searchQuery, searchQuery)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
	}
	var clients []Client
	for rows.Next() {
		var client Client
		rows.Scan(&client.Id, &client.FirstName, &client.LastName,
			&client.Gender, &client.Urination, &client.Defecation,
			&client.LastRecord, &client.ToId)

		clients = append(clients, client)
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/clientEntry.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"clients":        clients,
	})
}

// for htmx dashboard panel, /htmx/dashboard
func (server *Server) htmxDashboard(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "GET" {
		writeJson(writer, http.StatusBadRequest,
			map[string]interface{}{
				"error": "invalid request method",
			},
		)
		return
	}

	id, username, err := server.secureHtmx(writer, request)
	if id == -1 || username == "" || err != nil {
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/dashboard.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"id":             id,
		"username":       username,
	})
}

func (server *Server) htmxLogin(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "GET" {
		writeJson(writer, http.StatusBadRequest,
			map[string]interface{}{
				"error": "invalid request method",
			},
		)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/login.html"))
	err := tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
	})
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
	}
}

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

	id, username, err := server.secureHtmx(writer, request)
	if id == -1 || username == "" || err != nil {
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/clients.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
	})
}

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

	id, username, err := server.secureHtmx(writer, request)
	if id == -1 || username == "" || err != nil {
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

	id, username, err := server.secureHtmx(writer, request)
	if id == -1 || username == "" || err != nil {
		return
	}
	// "SELECT * FROM Watch WHERE to_id = $1;"
	db := server.dbStorage
	rows, err := db.db.Query(
		`SELECT Clients.id, Clients.first_name, Clients.last_name,
        Clients.gender, Clients.urination, Clients.defecation, Clients.last_record
        FROM Watch INNER JOIN Clients
        ON Watch.client_id = Clients.id
        WHERE Watch.to_id = $1`,
		id)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
	}
	var clients []Client
	for rows.Next() {
		var client Client
		rows.Scan(&client.Id, &client.FirstName, &client.LastName,
			&client.Gender, &client.Urination, &client.Defecation,
			&client.LastRecord)

		clients = append(clients, client)
	}
    log.Println("queried", clients)
	tmpl := template.Must(template.ParseFiles("./templates/htmx/clientEntry.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"clients":        clients,
	})
}
