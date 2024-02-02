package main

import (
	"database/sql"
	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
)

func (server *Server) isValidSession(writer http.ResponseWriter,
	request *http.Request) bool {
	store := server.redisStorage.store
	session, err := store.Get(request, "PS-cookie")
	if err != nil {
		log.Println("Server.go: isValidSession() - getSession")
		log.Println(err)
		return false
	}
	return session.Values["id"] != nil
}

func (server *Server) createSession(writer http.ResponseWriter,
	request *http.Request, id int, username string,
	telegram string, userType string) error {
	store := server.redisStorage.store
	session, err := store.Get(request, "PS-cookie")

	if err != nil {
		log.Println("Server.go: createSession() - getSession")
		log.Println(err)
		return err
	}

	session.Values["id"] = id
	session.Values["username"] = username
	session.Values["telegram"] = telegram
	session.Values["userType"] = userType
	session.Options.SameSite = http.SameSiteStrictMode

	err = session.Save(request, writer)
	if err != nil {
		log.Println("Server.go: createSession() - saveSession")
		log.Println(err)
		return err
	}
	return nil
}
func toLowerCase(str string) string {
	var lowerCase string
	for _, char := range str {
		if char >= 'A' && char <= 'Z' {
			lowerCase += string(char + 32)
		} else {
			lowerCase += string(char)
		}
	}
	return lowerCase
}

func saltPassword(password string) string {
	return "cS46O" + password + "$1aY"
}

func (server *Server) loginHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		loginPage(writer, request, server)
		break
	case "POST":
		loginApi(writer, request, server)
		break
	default:
		writeJson(writer, http.StatusBadRequest,
			map[string]string{
				"error": "method not allowed",
			},
		)
		return

	}
}

func loginPage(writer http.ResponseWriter,
	request *http.Request, server *Server) {
	if request.Method != "GET" {
		writeJson(writer, http.StatusMethodNotAllowed,
			map[string]string{
				"error": "method not allowed",
			},
		)
		return
	} else if server.isValidSession(writer, request) {
		log.Println("redirecting to dashboard")
		http.Redirect(writer, request, "/dashboard", http.StatusSeeOther)
		return
	}
	log.Println("trying to render", baseTemplate)
	tmpl := template.Must(template.ParseFiles(baseTemplate))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"hxGet":          "/htmx/login",
		"hxReplaceUrl":   "/login",
	})
	log.Println("render /login")
}

func loginApi(writer http.ResponseWriter,
	request *http.Request, server *Server) {

	username := toLowerCase(request.FormValue("username"))
	password := saltPassword(request.FormValue("password"))

	var id int
	var telegram string
	var userType string
	var passwordHash string
	err := server.dbStorage.db.QueryRow(
		"SELECT id, password, telegram, type FROM TOfficers WHERE username = $1",
		username).Scan(&id, &passwordHash, &telegram, &userType)

	if err == sql.ErrNoRows {
		writeJson(writer, http.StatusUnauthorized,
			map[string]string{
				"error": "invalid username",
			},
		)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		writeJson(writer, http.StatusUnauthorized,
			map[string]string{
				"error": "invalid password",
			},
		)
		return
	}

	server.createSession(writer, request, id, username, telegram, userType)
	http.Redirect(writer, request, "/dashboard", http.StatusSeeOther)
}

func returnError(writer http.ResponseWriter, request *http.Request,
	err error) {
	writeJson(writer, http.StatusInternalServerError,
		map[string]string{
			"error": err.Error(),
		},
	)
}

func (server *Server) logout(writer http.ResponseWriter, request *http.Request) {
	store := server.redisStorage.store
	session, err := store.Get(request, "PS-cookie")

	if err != nil {
		returnError(writer, request, err)
	}

	session.Options.MaxAge = -1

	err = session.Save(request, writer)
	if err != nil {
		returnError(writer, request, err)
	}

	http.Redirect(writer, request, "/login", http.StatusSeeOther)
}
