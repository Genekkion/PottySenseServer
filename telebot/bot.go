package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

const GENERIC_ERROR_MESSAGE = "Error processing your request right now. Please try again later!"

type Bot struct {
	bot        *tgbotapi.BotAPI
	db         *sql.DB
	redisCache *redis.Client
}

func NewBot(telegramBotToken string, db *sql.DB,
	redisCache *redis.Client) *Bot {
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Println("Error creating bot.")
		log.Fatalln(err)
	}

	bot.Debug = os.Getenv("IS_PROD") == "false"

	return &Bot{
		bot:        bot,
		db:         db,
		redisCache: redisCache,
	}
}

// Wraps all commands requiring the user to be
// authorised on the platform. All except
// "/start" will use this wrapper.
func (bot *Bot) authWrapper(function botCommandFunc) botCommandFunc {
	return func(update tgbotapi.Update) string {
		var id int
		err := bot.db.QueryRow(`
			SELECT id
			FROM TOfficers
			WHERE telegram_chat_id = $1
		`, update.Message.Chat.ID).Scan(&id)
		if err != nil {
			log.Println(err)
			return "Unauthorized user."
		}
		return function(update)
	}
}

// Starts running the bot
func (bot *Bot) Run() {
	log.Println(bot.bot.Self.UserName + " has started polling.")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updatesChannel := bot.bot.GetUpdatesChan(updateConfig)

	for update := range updatesChannel {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		message := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		//message.ParseMode = tgbotapi.ModeMarkdownV2
		message.ParseMode = tgbotapi.ModeHTML
		switch strings.ToLower(update.Message.Command()) {
		case "start":
			message.Text = bot.botCommandStart(update)
		case "clients":
			message.Text = bot.authWrapper(bot.botCommandGetAllClients)(update)
		case "current":
			message.Text = bot.authWrapper(bot.botCommandGetCurrentClients)(update)
		case "search":
			message.Text = bot.authWrapper(bot.botCommandSearchName)(update)
		case "id":
			message.Text = bot.authWrapper(bot.botCommandGetClient)(update)
		case "track":
			message.Text = bot.authWrapper(bot.botCommandTrackClient)(update)
		case "untrack":
			message.Text = bot.authWrapper(bot.botCommandUnTrackClient)(update)
		case "help":
			message.Text = bot.authWrapper(bot.botCommandHelp)(update)
		case "session":
			message.Text = bot.authWrapper(bot.botCommandSessionStart)(update)
		default:
			message.Text = "Error, command not found. Please use /help to get the list of available commands."

		}

		_, err := bot.bot.Send(message)
		if err != nil {
			log.Println("Error sending message")
			log.Println(err)
		}

	}
}

func (bot *Bot) botCommandHelp(update tgbotapi.Update) string {
	message := "<b>List of supported commands:</b>\n"
	message += "<b>1.</b> /start - Register Telegram account\n"
	message += "<b>2.</b> /clients - Get all clients\n"
	message += "<b>3.</b> /current - Get all currently tracked clients\n"
	message += "<b>4.</b> /id - Get the client with the id supplied\n"
	message += "<b>5.</b> /track - Start tracking the client with the id supplied\n"
	message += "<b>6.</b> /untrack - Stop tracking the client with the id supplied\n"
	message += "<b>7.</b> /session - Start a session for the client with the id supplied\n"
	message += "<b>8.</b> /help - List all available commands\n"
	return message
}
