package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

type htmxlRoute struct {
	urlSuffix string
	filePath  string
}

// for htmx routes, ensures that users are logged in
// returns username if logged in, empty string otherwise
func (server *Server) secureHtmx(writer http.ResponseWriter,
	request *http.Request) (TO, error) {
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
		return TO{}, err
	}

	id := session.Values["id"]
	username := session.Values["username"]
	userType := session.Values["userType"]
	telegram := session.Values["telegram"]
	if id == nil || username == nil {
		http.Redirect(writer, request, "/login", http.StatusSeeOther)
		return TO{}, nil
	}
	return TO{
		Id:       id.(int),
		Username: username.(string),
		Telegram: telegram.(string),
		UserType: userType.(string),
	}, nil
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

	to, err := server.secureHtmx(writer, request)
	if err != nil {
		return
	}
	tmpl := template.Must(template.ParseFiles("./templates/htmx/dashboard.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"id":             to.Id,
		"username":       to.Username,
		"userType":       to.UserType,
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

func (server *Server) htmxAccounts(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "GET" {
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
	} else if to.UserType != "admin" {
		writeJson(writer, http.StatusUnauthorized,
			map[string]string{
				"error": "unauthorized",
			},
		)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/accounts.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
	})
}

// /htmx/accounts/search
func (server *Server) htmxAccountsSearch(writer http.ResponseWriter,
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
	} else if to.UserType != "admin" {
		writeJson(writer, http.StatusUnauthorized,
			map[string]string{
				"error": "unauthorized",
			},
		)
		return
	}

	searchQuery := request.FormValue("search") + "%"
	log.Println("sq", searchQuery)
	db := server.dbStorage
	rows, err := db.db.Query(
		`SELECT id, first_name, last_name,
        telegram, type
        FROM TOfficers WHERE id != $1
        AND (first_name LIKE $2 COLLATE NOCASE
        OR last_name LIKE $3 COLLATE NOCASE)`,
		to.Id, searchQuery, searchQuery)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	var accounts []TO
	for rows.Next() {
		var to TO

		rows.Scan(
			&to.Id, &to.FirstName, &to.LastName,
			&to.Telegram, &to.UserType,
		)
		accounts = append(accounts, to)
	}
	log.Println("accounts", accounts)
	tmpl := template.Must(template.ParseFiles("./templates/htmx/accountEntry.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
		"accounts":       accounts,
	})
}

func (server *Server) htmxAccountEdit(writer http.ResponseWriter,
	request *http.Request) {
	if request.Method != "GET" {
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
	} else if to.UserType != "admin" {
		writeJson(writer, http.StatusUnauthorized,
			map[string]string{
				"error": "unauthorized",
			},
		)
		return
	}

	vars := mux.Vars(request)
	id := vars["id"]
	var account TO

	db := server.dbStorage
	err = db.db.QueryRow(
		`SELECT id, first_name, last_name,
        telegram, type
        FROM TOfficers WHERE id = $1`,
		id).Scan(&account.Id, &account.FirstName,
		&account.LastName, &account.Telegram, &account.UserType)

	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
	}
	log.Println("acc", account)
	tmpl := template.Must(template.ParseFiles("./templates/htmx/accountEditable.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
		"account":        account,
	})
}

func (server *Server) htmxAccountSave(writer http.ResponseWriter,
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
	} else if to.UserType != "admin" {
		writeJson(writer, http.StatusUnauthorized,
			map[string]string{
				"error": "unauthorized",
			},
		)
		return
	}

	vars := mux.Vars(request)
	id := vars["id"]
	var toId int
	firstName := request.FormValue("firstName")
	lastName := request.FormValue("lastName")
	username := request.FormValue("username")
	telegram := request.FormValue("telegram")
	db := server.dbStorage
	_, err = db.db.Exec(
		`UPDATE TOfficers SET
        first_name = $1,
        last_name = $2,
        username = $3,
        telegram = $4
        WHERE id = $5`,
		firstName, lastName, username, telegram, id)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		log.Println("here2")
		log.Println(err)
		return
	} else if toId, err = strconv.Atoi(id); err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}
	tmpl := template.Must(template.ParseFiles("./templates/htmx/accountEntrySingle.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
		"account": TO{
			Id:        toId,
			FirstName: firstName,
			LastName:  lastName,
            Username: username,
			Telegram:  telegram,
			UserType:  "user",
		},
	})
}
