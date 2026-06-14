package inputs

import (
	"log"
	"strings"

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
		// todo: replace with real DB fetch
		categories := []Category{
			{ID: "1", Name: "Health"},
			{ID: "2", Name: "Sport"},
			{ID: "3", Name: "Learning"},
		}

		if len(categories) == 0 {
			send(bot, tgbotapi.NewMessage(chatID, "No categories found. Add one with /managecategory"))
			return
		}

		msg := tgbotapi.NewMessage(chatID, "Choose a category:")
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = trackHabitCategoryKeyboard(categories)
		send(bot, msg)

	case "todaystatus":
		send(bot, tgbotapi.NewMessage(chatID, "/todaystatus called")) //todo: return tracked habits for today

	default:
		send(bot, tgbotapi.NewMessage(chatID, "I don't know that command. Type /help to see available commands.")) //todo: need to add /help relization
	}
}

func handleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
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
	default:
		if strings.HasPrefix(cb.Data, "track:category:") {
			categoryID := strings.TrimPrefix(cb.Data, "track:category:")
			send(bot, tgbotapi.NewMessage(cb.Message.Chat.ID, "Selected category: "+categoryID)) // todo: show habits for this category
		}
	}
}

func send(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) {
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
