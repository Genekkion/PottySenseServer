package main

import (
	"database/sql"
	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
)

func (server *Server) isValidSession(request *http.Request) bool {
	store := server.redisSessionStore.store
	session, err := store.Get(request, "PS-cookie")
	if err != nil {
		log.Println(err)
		return false
	}
	return session.Values["id"] != nil
}

func (server *Server) createSession(writer http.ResponseWriter,
	request *http.Request, to TO) error {
	store := server.redisSessionStore.store
	session, err := store.Get(request, "PS-cookie")

	if err != nil {
		log.Println(err)
		return err
	}

	session.Values["id"] = to.Id
	session.Values["username"] = to.Username
	session.Values["telegram"] = to.Telegram
	session.Values["userType"] = to.UserType
	session.Options.SameSite = http.SameSiteStrictMode

	err = session.Save(request, writer)
	if err != nil {
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
	} else if server.isValidSession(request) {
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

	server.createSession(writer, request,
		TO{
			Id:       id,
			Username: username,
			Telegram: telegram,
			UserType: userType,
		})
	http.Redirect(writer, request, "/dashboard", http.StatusSeeOther)
}

func (server *Server) logout(writer http.ResponseWriter, request *http.Request) {
	store := server.redisSessionStore.store
	session, err := store.Get(request, "PS-cookie")

	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": err.Error(),
			},
		)
	}

	session.Options.MaxAge = -1

	err = session.Save(request, writer)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": err.Error(),
			},
		)
	}

	http.Redirect(writer, request, "/login", http.StatusSeeOther)
}

func (server *Server) secureHtmx(writer http.ResponseWriter,
	request *http.Request) (TO, error) {
	store := server.redisSessionStore.store
	session, err := store.Get(request, "PS-cookie")
	if err != nil {
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
