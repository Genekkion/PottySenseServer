package main

import (
    "net/http"
    "html/template"
    "github.com/gorilla/csrf"
)


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

	_, err := server.secureHtmx(writer, request)
	if err != nil {
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
			&client.LastRecord)

		clients = append(clients, client)
	}
	tmpl := template.Must(template.ParseFiles("./templates/htmx/clientEntry.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"clients":        clients,
	})
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

