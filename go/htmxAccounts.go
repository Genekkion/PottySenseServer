package main

import (
    "net/http"
    "html/template"
    "github.com/gorilla/csrf"
    "strconv"
)

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
	db := server.dbStorage
	rows, err := db.db.Query(
		`SELECT id, first_name, last_name,
        username, telegram, type
        FROM TOfficers WHERE id != $1
        AND (first_name LIKE $2 COLLATE NOCASE
        OR last_name LIKE $2 COLLATE NOCASE
        OR username LIKE $2 COLLATE NOCASE
        OR telegram LIKE $2 COLLATE NOCASE
        )`,
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
			&to.Username, &to.Telegram, &to.UserType,
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
			Username:  username,
			Telegram:  telegram,
			UserType:  "user",
		},
	})
}
