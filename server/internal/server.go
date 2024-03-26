package internal

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gopkg.in/boj/redistore.v1"
)

type Server struct {
	listenAddr        string
	dbStorage         *sql.DB
	redisSessionStore *redistore.RediStore
	router            *mux.Router
	redisStorage      *redis.Client
	telebotAddr       string
}

func InitServer(dbStorage *sql.DB, redisSessionStore *redistore.RediStore,
	redisStorage *redis.Client) *Server {

	listenAddr := ":" + os.Getenv("PORT")
	telebotAddr := "https://api.telegram.org/bot" +
		os.Getenv("TELEGRAM_BOT_TOKEN") + "/sendMessage"

	router := mux.NewRouter()
	server := &Server{
		listenAddr:        listenAddr,
		dbStorage:         dbStorage,
		redisSessionStore: redisSessionStore,
		router:            router,
		telebotAddr:       telebotAddr,
	}

	server.addFileServer()
	server.addRoutes()

	log.Printf("Server running on: http://localhost%s\n", server.listenAddr)
	return server
}

func (server *Server) addRoutes() {
	router := server.router

	router.HandleFunc("/", server.indexHandler)

	router.HandleFunc("/login", server.loginHandler)
	router.HandleFunc("/htmx/login", server.htmxLogin)

	router.HandleFunc("/logout", server.logout)

	router.HandleFunc("/dashboard", server.dashboardPage)
	router.HandleFunc("/htmx/dashboard", server.htmxDashboard)

	router.HandleFunc("/htmx/current", server.htmxCurrent)
	router.HandleFunc("/htmx/current/clients", server.htmxCurrentClients)

	router.HandleFunc("/htmx/clients", server.htmxClients)
	router.HandleFunc("/htmx/clients/search", server.htmxClientEntry)
	router.HandleFunc("/htmx/clients/assign", server.htmxClientAssign)
	router.HandleFunc("/htmx/clients/new", server.htmxClientNewHandler)

	router.HandleFunc("/htmx/accounts", server.htmxAccounts)
	router.HandleFunc("/htmx/accounts/search", server.htmxAccountsSearch)
	router.HandleFunc("/htmx/accounts/{id:[0-9]+}/edit", server.htmxAccountEdit)
	router.HandleFunc("/htmx/accounts/{id:[0-9]+}/save", server.htmxAccountSave)

	router.HandleFunc("/htmx/settings", server.htmxSettings)
	router.HandleFunc("/htmx/settings/details", server.htmxSettingsDetailsSave)
	router.HandleFunc("/htmx/settings/password", server.htmxSettingsPasswordHandler)
}

func (server *Server) Run() {
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
	session, err := server.redisSessionStore.Get(request, "PS-cookie")
	if err != nil {
		return err
	}

	telegram := session.Values["telegram"]
	chatId, err := rClient.Get(context.Background(),
		telegram.(string)).Result()
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

func (server *Server) addFileServer() {
	fileServer := http.FileServer(http.Dir("./static"))
	server.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))
}
