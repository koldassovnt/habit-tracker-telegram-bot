package inputs

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		handleCallback(bot, update.CallbackQuery)
		return
	}

	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	handleCommand(bot, update)
}

func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	switch update.Message.Command() {
	case "managecategory":
		msg := tgbotapi.NewMessage(chatID, "🗂 *Manage Categories*\nWhat would you like to do?")
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = manageCategoryKeyboard()
		send(bot, msg)

	case "managehabit":
		msg := tgbotapi.NewMessage(chatID, "📋 *Manage Habits*\nWhat would you like to do?")
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = manageHabitKeyboard()
		send(bot, msg)

	case "trackhabit":
		send(bot, tgbotapi.NewMessage(chatID, "/trackhabit called"))

	case "todaystatus":
		send(bot, tgbotapi.NewMessage(chatID, "/todaystatus called"))

	default:
		send(bot, tgbotapi.NewMessage(chatID, "I don't know that command. Type /help to see available commands."))
	}
}

func handleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
	// Acknowledge the callback first (removes the loading state on the button)
	bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	switch cb.Data {
	case "category:add":
		send(bot, tgbotapi.NewMessage(cb.Message.Chat.ID, "Adding a new category..."))
	case "category:edit":
		send(bot, tgbotapi.NewMessage(cb.Message.Chat.ID, "Editing a category..."))
	case "category:delete":
		send(bot, tgbotapi.NewMessage(cb.Message.Chat.ID, "Deleting a category..."))

	case "habit:add":
		send(bot, tgbotapi.NewMessage(cb.Message.Chat.ID, "Adding a new habit..."))
	case "habit:edit":
		send(bot, tgbotapi.NewMessage(cb.Message.Chat.ID, "Editing a habit..."))
	case "habit:delete":
		send(bot, tgbotapi.NewMessage(cb.Message.Chat.ID, "Deleting a habit..."))
	}
}

func send(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) {
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
