package internal

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/genekkion/PottySenseServer/internal/globals"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gopkg.in/boj/redistore.v1"
)

type Server struct {
	listenAddr        string
	db                *sql.DB
	redisSessionStore *redistore.RediStore
	router            *mux.Router
	redisStorage      *redis.Client
	telebotAddr       string
}

func InitServer(dbStorage *sql.DB,
	redisSessionStore *redistore.RediStore,
	redisStorage *redis.Client) *Server {

	listenAddr := os.Getenv("SERVER_ADDR")
	telebotAddr := "https://api.telegram.org/bot" +
		os.Getenv("TELEGRAM_BOT_TOKEN") + "/sendMessage"

	router := mux.NewRouter()
	server := &Server{
		listenAddr:        listenAddr,
		db:                dbStorage,
		redisSessionStore: redisSessionStore,
		redisStorage:      redisStorage,
		router:            router,
		telebotAddr:       telebotAddr,
	}

	server.addFileServer()
	server.addInternalRoutes()
	server.addExternalRoutes()

	log.Printf("Server running on: http://%s\n", server.listenAddr)
	return server
}

// Adds all internal routes
func (server *Server) addInternalRoutes() {
	router := server.router

	router.HandleFunc("/", server.indexHandler)

	router.HandleFunc("/login", server.loginHandler)
	router.HandleFunc("/htmx/login", server.htmxLoginHandler)

	router.HandleFunc("/logout", server.logout)

	router.HandleFunc("/track", server.authWrapper(server.dashboardTrack))
	router.HandleFunc("/htmx/track", server.authWrapper(server.htmxTrackingHandler))

	router.HandleFunc("/clients", server.authWrapper(server.dashboardClients))
	router.HandleFunc("/htmx/clients", server.authWrapper(server.htmxClients))
	router.HandleFunc("/htmx/clients/new", server.authWrapper(server.htmxClientNewHandler))

	router.HandleFunc("/accounts", server.authWrapper(server.dashboardAccounts))
	router.HandleFunc("/htmx/accounts", server.authWrapper(server.htmxAccountsHandler))
	router.HandleFunc("/htmx/accounts/edit", server.authWrapper(server.htmxAccountsEditHandler))
	router.HandleFunc("/htmx/accounts/new", server.authWrapper(server.htmxAccountsNewHandler))

	router.HandleFunc("/settings", server.authWrapper(server.dashboardSettings))
	router.HandleFunc("/htmx/settings", server.authWrapper(server.htmxSettingsHandler))
	router.HandleFunc("/htmx/settings/password", server.authWrapper(server.htmxSettingsPasswordHandler))
}

// Starts the server
func (server *Server) Run() {
	CSRF := csrf.Protect([]byte(os.Getenv("CSRF_SECRET")),
		csrf.Secure(os.Getenv("IS_PROD") == "true"))

	server.router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter,
			request *http.Request) {
			served := false
			for _, route := range globals.UNPROTECTED_ROUTES {
				if request.URL.Path == route {
					served = true
					handler.ServeHTTP(writer, request)
					break
				}
			}
			if !served {
				CSRF(handler).ServeHTTP(writer, request)
			}
		})
	})
	http.ListenAndServe(server.listenAddr, server.router)
}

// Writes json to the writer
func writeJson(writer http.ResponseWriter, statusCode int, value any) error {
	writer.WriteHeader(statusCode)
	writer.Header().Add("Content-Type", "application/json")

	return json.NewEncoder(writer).Encode(value)
}

// Handles the "/" route
func (server *Server) indexHandler(writer http.ResponseWriter,
	request *http.Request) {
	if server.isValidSession(request) {
		http.Redirect(writer, request,
			globals.DEFAULT_DASHBOARD_ROUTE, http.StatusSeeOther)
	} else {
		http.Redirect(writer, request,
			"/login", http.StatusSeeOther)
	}
}

// Sends telegram message to a specified chatId.
// Accepts the template to apply
func (server *Server) sendTeleTemplate(chatId string,
	tmpl *template.Template) error {

	var stringBuffer bytes.Buffer
	err := tmpl.Execute(&stringBuffer, "")
	if err != nil {
		log.Println("sendTeleTemplate(), execute template")
		log.Println(err)
		return err
	}

	body, err := json.Marshal(
		map[string]string{
			"chat_id":    chatId,
			"text":       stringBuffer.String(),
			"parse_mode": "HTML",
		},
	)
	if err != nil {
		return err
	}
	response := bytes.NewBuffer(body)
	_, err = http.Post(
		server.telebotAddr,
		"application/json",
		response,
	)
	return err
}

// Sends telegram message to a specified chatId
func (server *Server) sendTeleString(chatId string, message string, isSilent bool) error {
	body, err := json.Marshal(
		map[string]interface{}{
			"chat_id":              chatId,
			"text":                 message,
			"parse_mode":           "HTML",
			"disable_notification": isSilent,
		},
	)
	if err != nil {
		return err
	}
	response := bytes.NewBuffer(body)
	_, err = http.Post(
		server.telebotAddr,
		"application/json",
		response,
	)
	return err
}

// Serves static files
func (server *Server) addFileServer() {
	fileServer := http.FileServer(http.Dir("./static"))
	server.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))
}

// Generic reply json for methods which are not allowed
func genericMethodNotAllowedReply(writer http.ResponseWriter) {
	writeJson(writer, http.StatusMethodNotAllowed, map[string]string{
		"error": "Method not allowed.",
	})
}

// Generic reply json for server errors
func genericInternalServerErrorReply(writer http.ResponseWriter) {
	writeJson(writer, http.StatusInternalServerError, map[string]string{
		"error": "Internal server error.",
	})
}

// Generic reply json for forbbiden
func genericForbiddenReply(writer http.ResponseWriter) {
	writeJson(writer, http.StatusForbidden, map[string]string{
		"error": "Forbidden.",
	})
}

// Generic reply json for unauthorized access
func genericUnauthorizedReply(writer http.ResponseWriter) {
	writeJson(writer, http.StatusUnauthorized, map[string]string{
		"error": "Unauthorized.",
	})
}
