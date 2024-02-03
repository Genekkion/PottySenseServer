package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

type Server struct {
	listenAddr        string
	dbStorage         SqliteStorage
	redisSessionStore RedisSessionStore
	router            *mux.Router
	telebotAddr       string
	redisStorage      RedisStorage
}

func initServer(dbStorage *SqliteStorage, redisSessionStore *RedisSessionStore,
	redisStorage *RedisStorage) *Server {

	listenAddr := ":" + os.Getenv("PORT")
	telebotAddr := "https://api.telegram.org/bot" +
		os.Getenv("TELEGRAM_BOT_TOKEN") + "/sendMessage"

	router := mux.NewRouter()
	server := &Server{
		listenAddr:        listenAddr,
		dbStorage:         *dbStorage,
		redisSessionStore: *redisSessionStore,
		router:            router,
		telebotAddr:       telebotAddr,
	}

	addFileServer(server)
	addRoutes(server)

	log.Printf("Server running on: http://localhost%s\n", server.listenAddr)
	return server
}

func addRoutes(server *Server) {
	router := server.router

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

	router.HandleFunc("/htmx/accounts", server.htmxAccounts)
	router.HandleFunc("/htmx/accounts/search", server.htmxAccountsSearch)
	router.HandleFunc("/htmx/accounts/{id:[0-9]+}/edit", server.htmxAccountEdit)
	router.HandleFunc("/htmx/accounts/{id:[0-9]+}/save", server.htmxAccountSave)

	router.HandleFunc("/htmx/settings", server.htmxSettings)
	router.HandleFunc("/htmx/settings/details", server.htmxSettingsDetailsSave)
	router.HandleFunc("/htmx/settings/password", server.htmxSettingsPasswordHandler)
}

func (server *Server) run() {
	CSRF := csrf.Protect([]byte(os.Getenv("CSRF_SECRET")),
		csrf.Secure(os.Getenv("IS_PROD") == "true"))

	http.ListenAndServe(server.listenAddr, CSRF(server.router))
}

func writeJson(writer http.ResponseWriter, statusCode int, value any) error {
	writer.WriteHeader(statusCode)
	writer.Header().Add("Content-Type", "application/json")

	return json.NewEncoder(writer).Encode(value)
}

func (server *Server) indexHandler(writer http.ResponseWriter,
	request *http.Request) {
	var hxGet string
	var hxReplaceUrl string
	if server.isValidSession(request) {
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

func (server *Server) sendTele(request *http.Request, message string) error {
	rClient := server.redisStorage
	session, err := server.redisSessionStore.store.Get(request, "PS-cookie")
	if err != nil {
		return err
	}

	telegram := session.Values["telegram"]
	chatId, err := rClient.get(telegram.(string))
	if err != nil {
		return err
	}
	body, err := json.Marshal(
		map[string]string{
			"chat_id": chatId,
			"text":    message,
		},
	)
	if err != nil {
		return err
	}
	response := bytes.NewBuffer(body)
	_, err = http.Post(server.telebotAddr, "application/json", response)
	return err
}

func addFileServer(server *Server) {
	fileServer := http.FileServer(http.Dir("./static"))
	server.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))
}

