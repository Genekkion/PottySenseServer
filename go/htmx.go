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

	db := server.dbStorage
	rows, err := db.db.Query("SELECT * FROM Clients")
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

	tmpl := template.Must(template.ParseFiles("./templates/htmx/clients.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"clients":        clients,
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
	})
}
