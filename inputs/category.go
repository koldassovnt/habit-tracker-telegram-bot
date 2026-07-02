package inputs

import (
	"context"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
)

const maxCategoriesPerUser = 5

func handleCategoryAdd(bot *tgbotapi.BotAPI, chatID int64) {
	setSession(chatID, session{flow: flowAddCategory})
	send(bot, tgbotapi.NewMessage(chatID, "Send me the name of the new category:"))
}

func handleCategoryEditList(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64) {
	cats, err := store.ListCategories(ctx, userID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if len(cats) == 0 {
		send(bot, tgbotapi.NewMessage(chatID, "You have no categories yet."))
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Choose a category to edit:")
	msg.ReplyMarkup = categoryPickKeyboard(cats, "category:edit:")
	send(bot, msg)
}

func handleCategoryDeleteList(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64) {
	cats, err := store.ListCategories(ctx, userID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if len(cats) == 0 {
		send(bot, tgbotapi.NewMessage(chatID, "You have no categories yet."))
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Choose a category to delete:")
	msg.ReplyMarkup = categoryPickKeyboard(cats, "category:delete:")
	send(bot, msg)
}

func handleCategoryEditPick(bot *tgbotapi.BotAPI, chatID, categoryID int64) {
	setSession(chatID, session{flow: flowRenameCategory, categoryID: categoryID})
	send(bot, tgbotapi.NewMessage(chatID, "Send me the new name for this category:"))
}

func handleCategoryDeletePick(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, categoryID int64) {
	if err := store.DeleteCategory(ctx, userID, categoryID); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			send(bot, tgbotapi.NewMessage(chatID, "Category not found."))
			return
		}
		sendErr(bot, chatID)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, "Category deleted."))
}

func finishAddCategory(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64, name string) {
	count, err := store.CountCategories(ctx, userID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if count >= maxCategoriesPerUser {
		send(bot, tgbotapi.NewMessage(chatID, "You already have the maximum of 5 categories."))
		return
	}
	if _, err := store.CreateCategory(ctx, userID, name); err != nil {
		sendErr(bot, chatID)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, "Category \""+name+"\" created."))
}

func finishRenameCategory(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, categoryID int64, name string) {
	if err := store.RenameCategory(ctx, userID, categoryID, name); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			send(bot, tgbotapi.NewMessage(chatID, "Category not found."))
			return
		}
		sendErr(bot, chatID)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, "Category renamed to \""+name+"\"."))
}
