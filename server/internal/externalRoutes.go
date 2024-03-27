package internal

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func (server *Server) addExternalRoutes() {
	router := server.router
	router.HandleFunc("/ext/api", server.externalHandler)
	router.HandleFunc("/ext/api/client/{id:[0-9]+}", server.externalHandlerClient)
}

// /ext/api
func (server *Server) externalHandler(writer http.ResponseWriter,
	request *http.Request) {

	switch request.Method {
	case http.MethodGet:
		externalPing(writer, request)
	default:
		http.Error(writer, "Method not allowed!", http.StatusMethodNotAllowed)
	}
}

// /ext/api "GET"
// Test if server is running
func externalPing(writer http.ResponseWriter,
	_ *http.Request) {
	log.Println("get called")
	writeJson(writer, http.StatusOK, "Server is running and open for connections.")
}

// /ext/api/client
func (server *Server) externalHandlerClient(writer http.ResponseWriter,
	request *http.Request) {

	switch request.Method {
	case http.MethodGet:
		server.startSession(writer, request)
	default:
		http.Error(writer, "Method not allowed!", http.StatusMethodNotAllowed)
	}
}

// Gets client from the db based on clientId
func (server *Server) getClient(clientId int) Client {
	client := Client{
		Id: clientId,
	}
	err := server.dbStorage.QueryRow(
		`SELECT first_name, last_name,
			gender, urination,
			defecation, last_record
		FROM Clients
		WHERE id = $1
		`, clientId).Scan(
		&client.FirstName, &client.LastName,
		&client.Gender, &client.Urination,
		&client.Defecation, &client.LastRecord,
	)

	if err != nil {
		log.Println(err)
	}
	return client
}

// /ext/api/client
func (server *Server) startSession(writer http.ResponseWriter,
	request *http.Request) {
	vars := mux.Vars(request)
	id := vars["id"]
	clientId, _ := strconv.Atoi(id)
	client := server.getClient(clientId)

	log.Println("Urination timer started")
	go asyncCallback(
		time.Duration(client.Urination)*time.Second,
		func() {
			// TODO: Callback action after timer is up
			log.Println("Urination timer up!")
			chatIds := server.getAllTOWatching(clientId)

			for _, chatId := range chatIds {
				server.sendTeleString(chatId, id+" urination")
			}
		},
	)

	log.Println("Defecation timer started")
	go asyncCallback(
		time.Duration(client.Defecation)*time.Second,
		func() {
			// TODO: Callback action after timer is up
			log.Println("Defecation timer up!")
		},
	)
}

// Gets the chatIds for TOs watching for a particular client.
// TOs must have their telegram registered with the bot beforehand
func (server *Server) getAllTOWatching(clientId int) []string {
	rows, err := server.dbStorage.Query(
		`SELECT Tofficers.telegram_chat_id
		FROM Tofficers INNER JOIN Watch
		On Tofficers.id = Watch.to_id
		WHERE Watch.client_id = $1
		`, clientId)
	if err != nil {
		log.Println(err)
		return nil
	}
	var TOChatIDs []string

	for rows.Next() {
		var chatId string
		err := rows.Scan(&chatId)
		if err != nil {
			log.Println(err)
			continue
		} else if chatId == "" {
			continue
		}
		TOChatIDs = append(TOChatIDs, chatId)
	}

	log.Println(TOChatIDs)
	return TOChatIDs
}

// Calls the function after the specified duration.
// E.g. go asyncCallback(5 * time.Second, func(){ log.Println("test") })
func asyncCallback(duration time.Duration, function func()) {
	time.Sleep(duration)
	function()
}

// /ext/pi
func (server *Server) PiListener(writer http.ResponseWriter,
	request *http.Request) {

	switch request.Method {

	case http.MethodPost:
		server.PiSendTO(writer, request)

	default:
		http.Error(writer, "Method not allowed!", http.StatusMethodNotAllowed)
	}

}

/*
	json message format for pi -> server

	{
		"clientId": int,
		"message": "string",
		"messageType" : "alert / message / complete"
	}

*/

type PiMessage struct {
	ClientId    int    `json:"clientId"`
	Message     string `json:"message"`
	MessageType string `json:"messageType"`
}

// /ext/pi "POST"
func (server *Server) PiSendTO(writer http.ResponseWriter,
	request *http.Request) {

	err := request.ParseForm()
	if err != nil {
		log.Println("PiSendTo(), parse form")
		log.Println(err)
		return
	}

	var piMessage PiMessage
	err = json.NewDecoder(request.Body).Decode(&piMessage)
	if err != nil {
		log.Println("PiSendTo(), decode json")
		log.Println(err)
		return
	}

	chatIDs := server.getAllTOWatching(piMessage.ClientId)
	var templateFilePath string
	switch piMessage.MessageType {
	// TODO: handle alert message
	case "alert":
		templateFilePath = "/templates/telegram/alert0.md"
	// TODO: handle message
	case "message":

	// TODO: handle complete message
	case "complete":

	}
	
	tmpl := template.Must(template.ParseFiles(templateFilePath))
	for _, chatId := range chatIDs {
		server.sendTeleTemplate(chatId, tmpl)
	}
	
}
