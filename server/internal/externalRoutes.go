package internal

import (
	"bytes"
	"context"
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
	case http.MethodDelete:
		server.extBotSessionCancel(writer, request)
	default:
		genericMethodNotAllowedReply(writer)
	}

}

// /ext/bot "POST"
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

	var urination int
	var defecation int
	server.db.QueryRow(`
		SELECT urination, defecation
		FROM Clients
		WHERE id = $1
	`, botMessage.ClientId).Scan(
		&urination,
		&defecation,
	)

	body, err := json.Marshal(
		map[string]interface{}{
			"clientId":   botMessage.ClientId,
			"urination":  urination,
			"defecation": defecation,
			// "businessType": bot
		},
	)
	if err != nil {
		log.Println("extBotSessionStart(), make json")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	response := bytes.NewBuffer(body)

	postResponse, err := http.Post(
		"http://"+os.Getenv("PI_ADDR")+"/api",
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

func (server *Server) getToilet(clientId int) string {
	result, err := server.redisStorage.Get(
		context.Background(),
		"client-"+fmt.Sprint(clientId),
	).Int()
	if err != nil {
		log.Println("getToilet()")
		log.Println(err)
		return ""
	}

	return globals.TOILETS_URL[result]
}

// /ext/bot "DELETE"
func (server *Server) extBotSessionCancel(writer http.ResponseWriter,
	request *http.Request) {
	// type BotMessage struct {
	// 	ClientId int `json:"clientId"`
	// }

	// var botMessage BotMessage

	// err := json.NewDecoder(request.Body).Decode(&botMessage)
	// if err != nil {
	// 	log.Println("extBotSessionCancel(), decode json")
	// 	log.Println(err)
	// 	genericInternalServerErrorReply(writer)
	// 	return
	// }
	// log.Println("botMessage", botMessage)

	// serverUrl := server.getToilet(botMessage.ClientId)
	// if serverUrl == "" {
	// 	log.Println("Error getting toilet url")
	// 	genericInternalServerErrorReply(writer)
	// 	return
	// }

	deleteRequest, err := http.NewRequest(
		http.MethodDelete,
		"http://"+os.Getenv("PI_ADDR")+"/api",
		nil,
	)
	if err != nil {
		log.Println("extBotSessionCancel(), create request")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	response, err := http.DefaultClient.Do(deleteRequest)
	if err != nil {
		log.Println("extBotSessionCancel(), delete request")
		log.Println(err)
		genericInternalServerErrorReply(writer)
		return
	}

	log.Println(response.StatusCode, response.Body)
	writeJson(writer, http.StatusOK, map[string]interface{}{
		"message": "Bot session cancelled.",
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

// Gets the chatIds for TOs Tracking for a particular client.
// TOs must have their telegram registered with the bot beforehand
func (server *Server) getAllTOTracking(clientId int) []string {
	rows, err := server.db.Query(
		`SELECT Tofficers.telegram_chat_id
		FROM Tofficers INNER JOIN Track
		On Tofficers.id = Track.to_id
		WHERE Track.client_id = $1
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
		IsSilent    bool   `json:"silentMessage"`
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
	messageType := strings.ToLower(piMessage.MessageType)
	isSilent := true
	switch messageType {
	case "alert":
		message = "⚠️ <b>ALERT!</b> ⚠️\n"
		isSilent = false
	case "notification":
		message = "🔔 <b>Notification!</b> 🔔\n"
		isSilent = false
	case "complete":
		message = "✅ <b>Complete!</b> ✅\n"
		isSilent = false
	default:
		message = ""
	}

	message += piMessage.Message

	errCount := 0
	for _, chatId := range chatIDs {
		err := server.sendTeleString(chatId, message, isSilent)
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
