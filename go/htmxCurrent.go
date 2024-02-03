package main

import (
    "net/http"
    "html/template"
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
	rows, err := db.db.Query(
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
