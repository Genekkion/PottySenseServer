package internal

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

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
        		OR username LIKE $2 COLLATE NOCASE
        		OR telegram LIKE $2 COLLATE NOCASE)
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
		server.htmxAccountForm(writer, request)
	case http.MethodPut:
	default:
		genericMethodNotAllowedReply(writer)

	}

}

// /htmx/accounts/edit "POST"
// TODO: Change to modal
// TODO: Add set telegram handle
func (server *Server) htmxAccountForm(writer http.ResponseWriter,
	request *http.Request) {
	to := server.getTOFromCookie(request)

	if to.UserType != "admin" {
		genericForbiddenReply(writer)
		return
	}

	// TODO: REVAMP THE WHOLE THING

	/*
			vars := mux.Vars(request)
			id := vars["id"]
			var account TO

			db := server.db
			err = db.QueryRow(
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
			tmpl := template.Must(template.ParseFiles("./templates/htmx/accountEditable.html"))
			tmpl.Execute(writer, map[string]interface{}{
				csrf.TemplateTag: csrf.TemplateField(request),
				"csrfToken":      csrf.Token(request),
				"account":        account,
			})
	*/
}

// /htmx/accounts/edit "PUT"
func (server *Server) htmxAccountSave(writer http.ResponseWriter,
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

	toId, err := strconv.Atoi(request.FormValue("id"))
	if err != nil {
		writeJson(writer, http.StatusBadRequest,
			map[string]string{
				"error": "Form value id must be an integer.",
			},
		)
		return
	}
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
        WHERE id = $5
		`, firstName, lastName,
		username, userType, toId)

	if err != nil {
		log.Println("htmxAccountsSave() - db update")
		log.Println(err)
		genericInternalServerErrorReply(writer)
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
			Username:  username,
			UserType:  userType,
		},
	})
}
