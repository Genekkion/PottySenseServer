package internal

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

func (server *Server) addExternalRoutes() {
	router := server.router
	router.HandleFunc("/ext", server.externalHealth)
	router.HandleFunc("/ext/api", server.externalHandler)
}

// /ext ALL METHODS
// Mainly to test for server connection
func (server *Server) externalHealth(writer http.ResponseWriter,
	request *http.Request) {
	writeJson(writer, http.StatusOK, map[string]string{
		"message": "Server is up and running!",
	})

}

// /ext/api
func (server *Server) externalHandler(writer http.ResponseWriter,
	request *http.Request) {

	switch request.Method {
	case http.MethodPost:
		server.PiSendTO(writer, request)
	default:
		genericMethodNotAllowedReply(writer)
	}
}

// Gets client from the db based on clientId
func (server *Server) getClient(clientId int) Client {
	client := Client{
		Id: clientId,
	}
	err := server.db.QueryRow(
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

// Gets the chatIds for TOs watching for a particular client.
// TOs must have their telegram registered with the bot beforehand
func (server *Server) getAllTOWatching(clientId int) []string {
	rows, err := server.db.Query(
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
	return TOChatIDs
}

/*
	json message format for pi -> server

	{
		"clientId": int,
		"message": "string",
		"messageType" : "alert / message / complete"
	}

*/

// /ext/api/client "POST"
func (server *Server) PiSendTO(writer http.ResponseWriter,
	request *http.Request) {
	type PiMessage struct {
		ClientId    int    `json:"clientId"`
		Message     string `json:"message"`
		MessageType string `json:"messageType"`
	}

	var piMessage PiMessage

	err := json.NewDecoder(request.Body).Decode(&piMessage)
	if err != nil {
		log.Println("PiSendTo(), decode json")
		log.Println(err)
		return
	}
	log.Println(piMessage)

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
