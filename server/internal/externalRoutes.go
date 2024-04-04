package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/genekkion/PottySenseServer/internal/globals"
)

// Wraps any http.HandleFunc functions which
// are unprotected by CSRF
func (server *Server) extWrapper(function serverFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Println(request.Header.Get(globals.SECRET_HEADER))
		log.Println(os.Getenv("SECRET_HEADER"))
		if request.Header.Get(globals.SECRET_HEADER) !=
			os.Getenv("SECRET_HEADER") {
			genericUnauthorizedReply(writer)
		} else {
			function(writer, request)
		}

	}
}

func (server *Server) addExternalRoutes() {
	router := server.router
	router.HandleFunc("/ext", server.extWrapper(server.externalHealth))
	router.HandleFunc("/ext/api", server.extWrapper(server.extApiHandler))
	router.HandleFunc("/ext/bot", server.extWrapper(server.extBotHandler))

}

// /ext ALL METHODS
// Mainly to test for server connection
func (server *Server) externalHealth(writer http.ResponseWriter,
	request *http.Request) {
	writeJson(writer, http.StatusOK, map[string]string{
		"message": "Server is up and running!",
	})
}

// /ext/bot
func (server *Server) extBotHandler(writer http.ResponseWriter,
	request *http.Request) {

	switch request.Method {
	case http.MethodPost:
		server.extBotSessionStart(writer, request)
	default:
		genericMethodNotAllowedReply(writer)
	}

}

func (server *Server) extBotSessionStart(writer http.ResponseWriter,
	request *http.Request) {
	type BotMessage struct {
		ClientId int `json:"clientId"`
	}

	var botMessage BotMessage

	err := json.NewDecoder(request.Body).Decode(&botMessage)
	if err != nil {
		log.Println("extBotSessionStart(), decode json")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}
	log.Println("botMessage", botMessage)

	body, err := json.Marshal(
		map[string]string{
			"clientId": fmt.Sprint(botMessage.ClientId),
		},
	)
	if err != nil {
		log.Println("extBotSessionStart(), format string")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	response := bytes.NewBuffer(body)

	postResponse, err := http.Post(
		os.Getenv("PI_ADDR"),
		"application/json",
		response,
	)
	if err != nil {
		log.Println("extBotSessionStart(), post request")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	log.Println(postResponse.StatusCode, postResponse.Body)
	writeJson(writer, http.StatusOK, map[string]interface{}{
		"message": "Bot session started.",
	})
}

// /ext/api
func (server *Server) extApiHandler(writer http.ResponseWriter,
	request *http.Request) {

	switch request.Method {
	case http.MethodPost:
		server.extSendTele(writer, request)
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
func (server *Server) getAllTOTracking(clientId int) []string {
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

// /ext/api/client "POST"
func (server *Server) extSendTele(writer http.ResponseWriter,
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
		genericInternalServerErrorReply(writer)
		return
	}
	log.Println(piMessage)

	chatIDs := server.getAllTOTracking(piMessage.ClientId)
	if len(chatIDs) == 0 {
		writeJson(writer, http.StatusInternalServerError, map[string]string{
			"warning": "No TOs currently tracking this client.",
		})
		return
	}

	var message string
	switch strings.ToLower(piMessage.MessageType) {
	case "alert":
		message = "⚠️ <b>ALERT!</b> ⚠️\n"
	case "notification":
		message = "🔔 <b>Notification!</b> 🔔\n"
	case "complete":
		message = "✅ <b>Complete!</b> ✅\n"
	default:
		message = ""
	}

	message += piMessage.Message

	errCount := 0
	for _, chatId := range chatIDs {
		err := server.sendTeleString(chatId, message)
		if err != nil {
			log.Println(err)
			errCount++
		}
	}
	if errCount == 0 {
		message = "All messages successfuly sent."
	} else {
		message = "Some messages successfuly sent."
	}
	writeJson(writer, http.StatusOK, map[string]string{
		"message": message,
	})
}
