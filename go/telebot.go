package main
/*
import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func initTeleBot() {

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Println("bot.go: setupBot()")
		log.Fatal(err)
	}
	bot.Debug = os.Getenv("IS_PROD") == "false"

	log.Printf("Authorized on account %s", bot.Self.UserName)

	PSBot = bot

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

    PSWG.Done()
	go telebotHandler(updates)
}

func telebotHandler(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		// Skip any non-Message Updates, or non-commands
		if update.Message == nil || !update.Message.IsCommand() {
			log.Println("update skipped")
			continue
		}

		log.Println("update command: ", update.Message.Command())
		switch update.Message.Command() {
		default:
			message := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Unknown command.")
			if _, err := PSBot.Send(message); err != nil {
				log.Println("telebot.go: Start() - unknown command")
				log.Println(err)
			}
		}

	}
}

func (telebot *TeleBot) handleStart(update tgbotapi.Update) {
	bot := telebot.bot
	message := tgbotapi.NewMessage(update.Message.Chat.ID,
		"Your telegram account has been registered! Please return to the PottySense Portal.")
	if _, err := bot.Send(message); err != nil {
		log.Println("telebot.go: handleStart()")
		log.Println(err)
	}
}
*/
