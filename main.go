package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_HABIT_TRACKER_TOKEN"))

	if err != nil {
		log.Fatal(err)
	}

	setCommands(bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		go handleUpdate(bot, update)
	}
}

// Set commands on startup
func setCommands(bot *tgbotapi.BotAPI) {
	commands := []tgbotapi.BotCommand{
		{Command: "managecategory", Description: "Add, edit, delete category"},
		{Command: "managehabit", Description: "Add, edit, delete habit"},
		{Command: "trackhabit", Description: "Track a habit"},
		{Command: "todaystatus", Description: "Status of tracked habits for today"},
		{Command: "report", Description: "Generate Report"},
	}

	cfg := tgbotapi.NewSetMyCommands(commands...)
	_, err := bot.Request(cfg)
	if err != nil {
		log.Printf("Failed to set commands: %v", err)
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	if !update.Message.IsCommand() {
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	// Extract the command from the Message.
	switch update.Message.Command() {
	case "managecategory":
		msg.Text = "/managecategory called"
	case "managehabit":
		msg.Text = "/managehabit called"
	case "trackhabit":
		msg.Text = "/trackhabit called"
	case "todaystatus":
		msg.Text = "/todaystatus called"
	case "report":
		msg.Text = "/report called"
	default:
		msg.Text = "I don't know that command"
	}

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}
