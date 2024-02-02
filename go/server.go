package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type Server struct {
	listenAddr   string
	dbStorage    SqliteStorage
	redisStorage RedisStorage
	router       *mux.Router
}

func writeJson(writer http.ResponseWriter, statusCode int, value any) error {
	writer.WriteHeader(statusCode)
	writer.Header().Add("Content-Type", "application/json")

	return json.NewEncoder(writer).Encode(value)
}

func makeHTTPHandleFunc(function apiFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := function(writer, request); err != nil {
			writeJson(writer, http.StatusInternalServerError,
				fmt.Errorf("internal server error"))
		}
	}
}

func (server *Server) template(writer http.ResponseWriter,
	request *http.Request) error {
	return nil
}

func (server *Server) indexHandler(writer http.ResponseWriter,
	request *http.Request) {
	var hxGet string
	var hxReplaceUrl string
	if server.isValidSession(writer, request) {
		hxGet = "/htmx/dashboard"
		hxReplaceUrl = "/dashboard"

	} else {
		hxGet = "/htmx/login"
		hxReplaceUrl = "/login"
	}

	tmpl := template.Must(template.ParseFiles(baseTemplate))
	tmpl.Execute(writer, map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(request),
		"hxGet":          hxGet,
		"hxReplaceUrl":   hxReplaceUrl,
	})
}

func initServer(listenAddr string, dbStorage *SqliteStorage,
	redisStorage *RedisStorage) *Server {

	router := mux.NewRouter()
	server := &Server{
		listenAddr:   listenAddr,
		dbStorage:    *dbStorage,
		redisStorage: *redisStorage,
		router:       router,
	}

	server.addStaticRoutes()

	router.HandleFunc("/", server.indexHandler)

	router.HandleFunc("/login", server.loginHandler)
	router.HandleFunc("/htmx/login", server.htmxLogin)

	router.HandleFunc("/logout", server.logout)

	router.HandleFunc("/dashboard", server.dashboardPage)
	router.HandleFunc("/htmx/dashboard", server.htmxDashboard)

	router.HandleFunc("/htmx/current", server.htmxCurrent)

	router.HandleFunc("/htmx/clients", server.htmxClients)
	router.HandleFunc("/htmx/clients/search", server.htmxClientEntry)
	router.HandleFunc("/htmx/clients/current", server.htmxCurrentClients)
	log.Printf("Server running on: http://localhost%s\n", server.listenAddr)
	return server
}

func (server *Server) run() {
	CSRF := csrf.Protect([]byte(os.Getenv("CSRF_SECRET")),
		csrf.Secure(os.Getenv("IS_PROD") == "true"))

	http.ListenAndServe(server.listenAddr, CSRF(server.router))
}
