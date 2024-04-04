package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

type botCommandFunc func(tgbotapi.Update) string

// Registers user if authorised.
func (bot *Bot) botCommandStart(update tgbotapi.Update) string {
	var toID int
	err := bot.db.QueryRow(`
		SELECT id
		FROM TOfficers
		WHERE telegram_chat_id = $1
		`, update.Message.Chat.ID).Scan(&toID)
	if err == nil {
		return "Your account has already been registered with PottySense!"
	} else if err != sql.ErrNoRows {
		// Problems connecting to DB
		log.Fatalln(err)
	}

	username := strings.ToLower(update.SentFrom().UserName)

	toIDStr, err := bot.redisCache.Get(context.Background(), username).Result()
	// Only errors here when the cache does not have
	// the info. Means user needs to go on platform
	// to register first
	if err != nil {
		log.Println("botCommandStart(), redis get")
		log.Println(err)
		return "Unauthorized user"
	}

	toID, err = strconv.Atoi(toIDStr)
	if err != nil {
		log.Println("botCommandStart(), atoi")
		log.Println(err)
		return "Error processing your request right now, please try again later!"
	}

	_, err = bot.db.Exec(`
			UPDATE TOfficers
			SET telegram_chat_id = $1
			WHERE id = $2
			`, update.Message.Chat.ID, toID)
	if err != nil {
		log.Println("botCommandStart(), update sql")
		log.Println(err)
		return "Error processing your request right now, please try again later!"
	}

	err = bot.redisCache.Del(context.Background(), username).Err()
	if err != nil {
		log.Println("Error deleting key from cache. Rectify immediately!")
		// Needs to be rectified immediately to prevent
		// odd behaviours
	}
	return "Your account has been registered with PottySense!"
}

func (bot *Bot) botCommandGetAllClients(update tgbotapi.Update) string {
	type ClientData struct {
		id        int
		firstName string
		lastName  string
	}
	rows, err := bot.db.Query(`
		SELECT id, first_name, last_name
		FROM Clients
		ORDER BY id
		`)
	if err != nil {
		log.Println(err)
		return GENERIC_ERROR_MESSAGE
	}

	var clients []ClientData
	for rows.Next() {
		var client ClientData
		err = rows.Scan(&client.id, &client.firstName, &client.lastName)
		if err != nil {
			log.Println(err)
			return GENERIC_ERROR_MESSAGE
		}
		clients = append(clients, client)
	}
	if len(clients) == 0 {
		return "No clients found in the database."
	}
	message := "<b>List of clients</b>\n"
	for _, client := range clients {
		message += fmt.Sprintf("[%d] %s %s\n",
			client.id, client.firstName, client.lastName,
		)
	}

	return message
}

func (bot *Bot) botCommandGetCurrentClients(update tgbotapi.Update) string {
	type ClientData struct {
		id         int
		firstName  string
		lastName   string
		lastRecord time.Time
	}
	rows, err := bot.db.Query(`
		SELECT Clients.id, Clients.first_name,
			Clients.last_name, Clients.last_record
		FROM Clients
			INNER JOIN Track
				ON Clients.id = Track.client_id
			INNER JOIN TOfficers
				ON TOfficers.id = Track.to_id
		WHERE TOfficers.telegram_chat_id = $1
		ORDER BY Clients.id
		`, update.Message.Chat.ID)
	if err != nil {
		log.Println(err)
		return GENERIC_ERROR_MESSAGE
	}
	var clients []ClientData
	for rows.Next() {
		var client ClientData
		err = rows.Scan(
			&client.id,
			&client.firstName,
			&client.lastName,
			&client.lastRecord,
		)
		if err != nil {
			log.Println(err)
			return GENERIC_ERROR_MESSAGE
		}
		log.Println(client)

		clients = append(clients, client)
	}
	if len(clients) == 0 {
		return "You are currently not tracking any clients."
	}

	message := "<b>Currently tracking</b>\n"
	for _, client := range clients {
		message += fmt.Sprintf("[%d] %s %s - %s\n",
			client.id,
			client.firstName,
			client.lastName,
			getTimeElapsedPretty(client.lastRecord),
		)
	}
	return message
}

func getTimeElapsedPretty(timeRecord time.Time) string {
	elapsedTime := time.Since(timeRecord)
	return fmt.Sprintf("%02d:%02d",
		int(elapsedTime.Hours()),
		int(elapsedTime.Minutes())%60,
	)
}

func (bot *Bot) botCommandSearchName(update tgbotapi.Update) string {
	type ClientData struct {
		id        int
		firstName string
		lastName  string
	}
	queries := strings.Split(update.Message.Text, " ")
	// Only accept 1 name query at a time
	if len(queries) != 2 {
		return "Please use the /search command with exactly 1 name after the command."
	}
	query := queries[1]

	rows, err := bot.db.Query(`
        SELECT id, first_name, last_name
        FROM Clients
        WHERE first_name LIKE $1 COLLATE NOCASE
        OR last_name LIKE $1 COLLATE NOCASE
		`, query)
	if err != nil {
		log.Println(err)
		return GENERIC_ERROR_MESSAGE
	}

	var clients []ClientData
	for rows.Next() {
		var client ClientData
		err = rows.Scan(&client.id, &client.firstName, &client.lastName)
		if err != nil {
			log.Println(err)
			return GENERIC_ERROR_MESSAGE
		}
		clients = append(clients, client)
	}
	if len(clients) == 0 {
		return "No clients found with the name \"" + query + "\"."
	}
	message := "<b>List of clients with the name \"" + query + "\"</b>\n"
	for _, client := range clients {
		message += fmt.Sprintf("[%d] %s %s\n",
			client.id, client.firstName, client.lastName,
		)
	}
	return message
}

func (bot *Bot) botCommandGetClient(update tgbotapi.Update) string {
	type ClientData struct {
		id         int
		firstName  string
		lastName   string
		urination  int
		defecation int
		lastRecord time.Time
	}
	queries := strings.Split(update.Message.Text, " ")
	// Only accept 1 name query at a time
	if len(queries) != 2 {
		return "Please use the /id command with exactly 1 id after the command."
	}
	query := queries[1]

	clientId, err := strconv.Atoi(query)
	if err != nil {
		return "Please use the /id command with the numeric id of the client."
	}

	client := ClientData{
		id: clientId,
	}
	err = bot.db.QueryRow(`
        SELECT first_name, last_name,
			urination, defecation, last_record
        FROM Clients
		WHERE id = $1
		`, clientId).Scan(
		&client.firstName,
		&client.lastName,
		&client.urination,
		&client.defecation,
		&client.lastRecord,
	)
	if err == sql.ErrNoRows {
		return "No client found with the id [" + query + "]."
	} else if err != nil {
		log.Println(err)
		return GENERIC_ERROR_MESSAGE
	}

	message := "<b>Client [" + query + "]</b>\n"
	message += fmt.Sprintf("First name: %s\n", client.firstName)
	message += fmt.Sprintf("Last name: %s\n", client.lastName)
	message += fmt.Sprintf("Urination: %s\n", secondsTimeString(client.urination))
	message += fmt.Sprintf("Defecation: %s\n", secondsTimeString(client.defecation))
	message += fmt.Sprintf("Last record: %s\n", getTimeElapsedPretty(client.lastRecord))
	return message
}

func secondsTimeString(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	return time.Time{}.Add(duration).Format("04:05")
}

func (bot *Bot) botCommandTrackClient(update tgbotapi.Update) string {
	queries := strings.Split(update.Message.Text, " ")
	// Only accept 1 name query at a time
	if len(queries) != 2 {
		return "Please use the /track command with exactly 1 id after the command."
	}
	query := queries[1]

	clientId, err := strconv.Atoi(query)
	if err != nil {
		return "Please use the /track command with the numeric id of the client."
	}

	_, err = bot.db.Exec(`
		INSERT OR IGNORE
			INTO Track (to_id, client_id)
		SELECT id, $1
		FROM TOfficers
		WHERE telegram_chat_id = $2
	`, clientId, update.Message.Chat.ID)
	if err != nil {
		return GENERIC_ERROR_MESSAGE
	}
	return "Successfully added to your tracking list!"
}

func (bot *Bot) botCommandUnTrackClient(update tgbotapi.Update) string {
	queries := strings.Split(update.Message.Text, " ")
	// Only accept 1 name query at a time
	if len(queries) != 2 {
		return "Please use the /untrack command with exactly 1 id after the command."
	}
	query := queries[1]

	clientId, err := strconv.Atoi(query)
	if err != nil {
		return "Please use the /untrack command with the numeric id of the client."
	}

	_, err = bot.db.Exec(`
		DELETE FROM Track
		WHERE client_id = $1
			AND to_id IN 
				(SELECT id FROM TOfficers
					WHERE telegram_chat_id = $2)
	`, clientId, update.Message.Chat.ID)
	if err != nil {
		return GENERIC_ERROR_MESSAGE
	}
	return "Successfully removed from your tracking list!"
}

// WARN: Currently hardcoded for one toilet
func (bot *Bot) botCommandSessionStart(update tgbotapi.Update) string {
	queries := strings.Split(update.Message.Text, " ")
	// Only accept 1 name query at a time
	if len(queries) != 2 {
		return "Please use the /session command with exactly 1 id after the command."
	}
	query := queries[1]

	clientId, err := strconv.Atoi(query)
	if err != nil {
		return "Please use the /session command with the numeric id of the client."
	}

	body, err := json.Marshal(
		map[string]int{
			"clientId": clientId,
		},
	)
	if err != nil {
		return GENERIC_ERROR_MESSAGE
	}

	postRequest, err := http.NewRequest(
		http.MethodPost,
		"http://"+os.Getenv("SERVER_ADDR")+"/ext/bot",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return GENERIC_ERROR_MESSAGE
	}
	postRequest.Header.Set("X-PS-Header", os.Getenv("SECRET_HEADER"))

	postResponse, err := http.DefaultClient.Do(postRequest)
	if err != nil {
		log.Println(err)
		return GENERIC_ERROR_MESSAGE
	}
	log.Println("serverResponse", postResponse.StatusCode)

	// WARN: Harcoded for single toilet with id 1
	err = bot.redisCache.Set(context.Background(),
		"client-"+fmt.Sprint(clientId),
		1,
		0,
	).Err()
	if err != nil {
		log.Println(err)
		return GENERIC_ERROR_MESSAGE
	}

	return "Successfully started the session!"
}

// WARN: Currently hardcoded for one toilet
func (bot *Bot) botCommandSessionCancel(update tgbotapi.Update) string {

	queries := strings.Split(update.Message.Text, " ")
	// Only accept 1 name query at a time
	if len(queries) != 2 {
		return "Please use the /session command with exactly 1 id after the command."
	}
	query := queries[1]

	clientId, err := strconv.Atoi(query)
	if err != nil {
		return "Please use the /session command with the numeric id of the client."
	}

	_, err = bot.redisCache.Get(context.Background(),
		"client-"+fmt.Sprint(clientId)).Result()
	if err == redis.Nil {
		log.Println(err)
		return "No session found for client with id " + fmt.Sprint(clientId)
	} else if err != nil {
		log.Println(err)
		return GENERIC_ERROR_MESSAGE
	}

	body, err := json.Marshal(
		map[string]int{
			"clientId": clientId,
		},
	)
	if err != nil {
		return GENERIC_ERROR_MESSAGE
	}

	postRequest, err := http.NewRequest(
		http.MethodPost,
		"http://"+os.Getenv("SERVER_ADDR")+"/ext/bot",
		bytes.NewBuffer(body),
	)
	postRequest.Header.Set("X-PS-Header", os.Getenv("SECRET_HEADER"))
	if err != nil {
		return GENERIC_ERROR_MESSAGE
	}

	postResponse, err := http.DefaultClient.Do(postRequest)
	if err != nil {
		return GENERIC_ERROR_MESSAGE
	}

	log.Println("serverResponse", postResponse.StatusCode)

	return "Successfully started the session!"
}
