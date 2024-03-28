package internal

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/genekkion/PottySenseServer/internal/globals"
	"github.com/genekkion/PottySenseServer/internal/utils"
	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
)

// For internal use, to check if the browser session is valid.
// Returns a boolean value representing the validity of the session.
func (server *Server) isValidSession(request *http.Request) bool {
	store := server.redisSessionStore
	session, err := store.Get(request, globals.COOKIE_NAME)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println(
		session.Values[globals.COOKIE_TO_ID],
		session.Values[globals.COOKIE_TO_USERNAME],
		session.Values[globals.COOKIE_TO_TELE_CHAT_ID],
		session.Values[globals.COOKIE_TO_USER_TYPE],
	)
	return session.Values[globals.COOKIE_TO_ID] != nil
}

// Creates a browser session and saves it to the
// browser cookie. to object supplied must
// have the following fields populated: Id, Username,
// TelegramChatId, UserType.
func (server *Server) createSession(writer http.ResponseWriter,
	request *http.Request, to TO) error {
	store := server.redisSessionStore
	session, err := store.Get(request, globals.COOKIE_NAME)

	if err != nil {
		log.Println(err)
		return err
	}

	session.Values[globals.COOKIE_TO_ID] = to.Id
	session.Values[globals.COOKIE_TO_USERNAME] = to.Username
	session.Values[globals.COOKIE_TO_TELE_CHAT_ID] = to.TelegramChatId
	session.Values[globals.COOKIE_TO_USER_TYPE] = to.UserType
	session.Options.SameSite = http.SameSiteStrictMode

	err = session.Save(request, writer)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// /login
// Only allow "GET"
func (server *Server) loginHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		if server.isValidSession(request) {
			log.Println("Already logged in, redirecting to dashboard")
			http.Redirect(writer, request, "/dashboard", http.StatusSeeOther)
			return
		}

		tmpl := template.Must(
			template.ParseFiles(
				globals.BASE_TEMPLATE,
				"./templates/htmx/loginPanel.html",
			))
		tmpl.ExecuteTemplate(writer, "base", map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
		})

	default:
		genericMethodNotAllowedReply(writer)
	}
}

// /htmx/login
func (server *Server) htmxLoginHandler(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case "GET":
		server.htmxLoginPanel(writer, request)
	case "POST":
		server.htmxLoginForm(writer, request)
	default:
		genericMethodNotAllowedReply(writer)
	}
}

// /htmx/login "GET"
func (server *Server) htmxLoginPanel(writer http.ResponseWriter,
	request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		tmpl := template.Must(template.ParseFiles("./templates/htmx/loginPanel.html"))
		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
		})
	default:
		genericMethodNotAllowedReply(writer)
	}
}

// /htmx/login "POST"
// Request should have form values:
// username, password
func (server *Server) htmxLoginForm(writer http.ResponseWriter,
	request *http.Request) {

	err := request.ParseForm()
	if err != nil {
		log.Println("htmxLoginForm(), parse form")
		log.Println(err)
		return
	}

	username := strings.ToLower(request.FormValue("username"))
	password := utils.SaltPassword(request.FormValue("password"))

	var id int
	var telegramChatId string
	var userType string
	var passwordHash string
	err = server.db.QueryRow(
		`SELECT id, password, 
			telegram_chat_id, type
		FROM TOfficers
		WHERE username = $1
		`, username).Scan(
		&id,
		&passwordHash,
		&telegramChatId,
		&userType,
	)

	if err == sql.ErrNoRows {
		tmpl := template.Must(template.ParseFiles("./templates/htmx/loginForm.html"))
		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
			"errorMessage":   "Invalid username!",
		})
		return
	} else if err != nil {
		tmpl := template.Must(template.ParseFiles("./templates/htmx/loginForm.html"))
		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
			"errorMessage":   "The server is experiencing issues right now, please try again later",
		})
		log.Println(err)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	)
	if err != nil {
		tmpl := template.Must(template.ParseFiles("./templates/htmx/loginForm.html"))
		tmpl.Execute(writer, map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(request),
			"errorMessage":   "Invalid password!",
		})
		return
	}

	server.createSession(writer, request,
		TO{
			Id:             id,
			Username:       username,
			TelegramChatId: telegramChatId,
			UserType:       userType,
		})
	writer.Header().Set("HX-Redirect", "/track")
}

// /logout
func (server *Server) logout(writer http.ResponseWriter, request *http.Request) {
	session, err := server.redisSessionStore.Get(request, globals.COOKIE_NAME)

	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": err.Error(),
			},
		)
	}
	session.Options.MaxAge = -1
	// Clears the cookie
	err = session.Save(request, writer)
	if err != nil {
		writeJson(writer, http.StatusInternalServerError,
			map[string]string{
				"error": err.Error(),
			},
		)
	}
	// Redirects to login page
	http.Redirect(writer, request, "/login", http.StatusSeeOther)
}

// Function prototype for the authWrapper below
type serverFunc func(http.ResponseWriter, *http.Request)

// Wraps any http.HandleFunc functions. Requires the
// browser to be logged in, else defaults to login page.
// Used for ALL possible routes that are exposed.
func (server *Server) authWrapper(function serverFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !server.isValidSession(request) {
			http.Redirect(writer, request, "/login", http.StatusSeeOther)
		} else {
			function(writer, request)
		}
	}
}

// WARN: For internal use only. Only to be
// used WITHIN a route that is auth wrapped to
// guarantee existence of TO details being found
// in cookie.
func (server *Server) getTOFromCookie(request *http.Request) *TO {
	// Will not error here since auth wrapped
	session, _ := server.redisSessionStore.Get(request, globals.COOKIE_NAME)

	return &TO{
		Id:             session.Values[globals.COOKIE_TO_ID].(int),
		Username:       session.Values[globals.COOKIE_TO_USER_TYPE].(string),
		TelegramChatId: session.Values[globals.COOKIE_TO_TELE_CHAT_ID].(string),
		UserType:       session.Values[globals.COOKIE_TO_USER_TYPE].(string),
	}
}
