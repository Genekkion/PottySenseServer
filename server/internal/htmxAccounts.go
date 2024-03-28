package internal

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/genekkion/PottySenseServer/internal/utils"
	"github.com/gorilla/csrf"
)

// /htmx/accounts
func (server *Server) htmxAccountsHandler(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		server.htmxAccountsPanel(writer, request)
	case http.MethodPost:
		server.htmxAccountsSearch(writer, request)
	default:
		genericMethodNotAllowedReply(writer)
	}

}

// /htmx/accounts "GET"
func (server *Server) htmxAccountsPanel(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	if to.UserType != "admin" {
		genericForbiddenReply(writer)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/accounts.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
		"to":             server.getTOFromCookie(request),
	})
}

// /htmx/accounts "POST"
func (server *Server) htmxAccountsSearch(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	if to.UserType != "admin" {
		genericForbiddenReply(writer)
		return
	}
	err := request.ParseForm()
	if err != nil {
		log.Println("htmxAccountsSearch() - parse form")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	// Add wildcard for autocomplete
	searchQuery := request.FormValue("search") + "%"
	rows, err := server.db.Query(
		`SELECT id, first_name,
			last_name, username,
			type
        FROM TOfficers
		WHERE id != $1
        	AND (first_name LIKE $2 COLLATE NOCASE
        		OR last_name LIKE $2 COLLATE NOCASE
        		OR username LIKE $2 COLLATE NOCASE)
		`, to.Id, searchQuery)

	if err != nil {
		log.Println("htmxAccountsSearch() - db query")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	var accounts []TO
	for rows.Next() {
		var to TO
		rows.Scan(
			&to.Id,
			&to.FirstName,
			&to.LastName,
			&to.Username,
			&to.UserType,
		)
		accounts = append(accounts, to)
	}
	tmpl := template.Must(template.ParseFiles("./templates/htmx/accountEntry.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"csrfToken":      csrf.Token(request),
		"accounts":       accounts,
	})
}

// /htmx/accounts/edit
func (server *Server) htmxAccountsEditHandler(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		server.htmxAccountEditModal(writer, request)
	case http.MethodPut:
		server.htmxAccountEditSave(writer, request)
	default:
		genericMethodNotAllowedReply(writer)
	}
}

// /htmx/accounts/edit "POST"
func (server *Server) htmxAccountEditModal(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	if to.UserType != "admin" {
		genericForbiddenReply(writer)
		return
	}

	err := request.ParseForm()
	if err != nil {
		log.Println("htmxAccountEditModal() - parse form")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/accountEditModal.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"id":             request.FormValue("id"),
		"firstName":      request.FormValue("firstName"),
		"lastName":       request.FormValue("lastName"),
		"username":       request.FormValue("username"),
		"userType":       request.FormValue("userType"),
	})
}

// /htmx/accounts/edit "PUT"
func (server *Server) htmxAccountEditSave(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	if to.UserType != "admin" {
		genericForbiddenReply(writer)
		return
	}

	err := request.ParseForm()
	if err != nil {
		log.Println("htmxAccountsSave() - parse form")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	toId, _ := strconv.Atoi(request.FormValue("id"))
	firstName := request.FormValue("firstName")
	lastName := request.FormValue("lastName")
	username := request.FormValue("username")
	userType := request.FormValue("userType")
	db := server.db
	_, err = db.Exec(
		`UPDATE TOfficers SET
        	first_name = $1,
        	last_name = $2,
        	username = $3,
			type = $4
        WHERE id = $5
		`, firstName, lastName,
		username, userType, toId)

	if err != nil {
		log.Println("htmxAccountsSave() - db update")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return

	}

	if request.FormValue("telegram") != "" {
		err = server.redisStorage.Set(
			request.Context(),
			request.FormValue("telegram"),
			toId,
			0,
		).Err()

		if err != nil {
			log.Println("htmxAccountsSave() - set redis")
			log.Println(err)
			genericInternalServerErrorReply(writer)
			return
		}
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/accountEntrySingle.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"Id":             toId,
		"FirstName":      firstName,
		"LastName":       lastName,
		"Username":       username,
		"UserType":       userType,
	})
}

// /htmx/accounts/new
func (server *Server) htmxAccountsNewHandler(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		server.htmxAccountNewModal(writer, request)
	case http.MethodPost:
		server.htmxAccountNewSave(writer, request)
	default:
		genericMethodNotAllowedReply(writer)
	}
}

// /htmx/accounts/new "GET"
func (server *Server) htmxAccountNewModal(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	if to.UserType != "admin" {
		genericForbiddenReply(writer)
		return
	}

	err := request.ParseForm()
	if err != nil {
		log.Println("htmxAccountNewModal() - parse form")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/htmx/accountNewModal.html"))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
	})
}

// /htmx/accounts/new "POST"
func (server *Server) htmxAccountNewSave(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	if to.UserType != "admin" {
		genericForbiddenReply(writer)
		return
	}

	err := request.ParseForm()
	if err != nil {
		log.Println("htmxAccountsNewSave() - parse form")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	err = utils.CreateUser(
		server.db,
		request.FormValue("firstName"),
		request.FormValue("lastName"),
		request.FormValue("username"),
		utils.SaltPassword(
			request.FormValue("password1")),
		request.FormValue("userType"),
	)
	if err != nil {
		log.Println("htmxAccountsNewSave() - create user")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	if request.FormValue("telegram") != "" {
		var toId int
		err = server.db.QueryRow(`
		SELECT id
		FROM TOfficers
		WHERE username = $1
	`, request.FormValue("username")).Scan(&toId)
		if err != nil {
			log.Println("htmxAccountsNewSave() - query id")
			log.Println(err)
			genericInternalServerErrorReply(writer)
			return
		}

		err = server.redisStorage.Set(
			request.Context(),
			request.FormValue("telegram"),
			toId,
			0,
		).Err()
		if err != nil {
			log.Println("htmxAccountsNewSave() - set redis")
			log.Println(err)
			genericInternalServerErrorReply(writer)
			return
		}
	}

	writer.Header().Set("HX-Trigger", "newAccount")
}
