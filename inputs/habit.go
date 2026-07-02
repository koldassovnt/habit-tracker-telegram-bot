package inputs

import (
	"context"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
)

const maxHabitsPerCategory = 10

func handleHabitAddCategoryList(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64) {
	cats, err := store.ListCategories(ctx, userID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if len(cats) == 0 {
		send(bot, tgbotapi.NewMessage(chatID, "Add a category first with /managecategory."))
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Which category should this habit belong to?")
	msg.ReplyMarkup = categoryPickKeyboard(cats, "habit:addcat:")
	send(bot, msg)
}

func handleHabitEditCategoryList(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64) {
	cats, err := store.ListCategories(ctx, userID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if len(cats) == 0 {
		send(bot, tgbotapi.NewMessage(chatID, "You have no categories yet."))
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Which category is the habit in?")
	msg.ReplyMarkup = categoryPickKeyboard(cats, "habit:editcat:")
	send(bot, msg)
}

func handleHabitDeleteCategoryList(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64) {
	cats, err := store.ListCategories(ctx, userID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if len(cats) == 0 {
		send(bot, tgbotapi.NewMessage(chatID, "You have no categories yet."))
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Which category is the habit in?")
	msg.ReplyMarkup = categoryPickKeyboard(cats, "habit:deletecat:")
	send(bot, msg)
}

func handleHabitAddPickCategory(bot *tgbotapi.BotAPI, chatID, categoryID int64) {
	setSession(chatID, session{flow: flowAddHabit, categoryID: categoryID})
	send(bot, tgbotapi.NewMessage(chatID, "Send me the name of the new habit:"))
}

func handleHabitEditListForCategory(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, categoryID int64) {
	habits, err := store.ListHabits(ctx, categoryID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if len(habits) == 0 {
		send(bot, tgbotapi.NewMessage(chatID, "No habits in this category yet."))
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Choose a habit to edit:")
	msg.ReplyMarkup = habitPickKeyboard(habits, "habit:edit:")
	send(bot, msg)
}

func handleHabitDeleteListForCategory(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, categoryID int64) {
	habits, err := store.ListHabits(ctx, categoryID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if len(habits) == 0 {
		send(bot, tgbotapi.NewMessage(chatID, "No habits in this category yet."))
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Choose a habit to delete:")
	msg.ReplyMarkup = habitPickKeyboard(habits, "habit:delete:")
	send(bot, msg)
}

func handleHabitEditPick(bot *tgbotapi.BotAPI, chatID, habitID int64) {
	setSession(chatID, session{flow: flowRenameHabit, habitID: habitID})
	send(bot, tgbotapi.NewMessage(chatID, "Send me the new name for this habit:"))
}

func handleHabitDeletePick(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, habitID int64) {
	if err := store.DeleteHabit(ctx, userID, habitID); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			send(bot, tgbotapi.NewMessage(chatID, "Habit not found."))
			return
		}
		sendErr(bot, chatID)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, "Habit deleted."))
}

func finishAddHabit(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, categoryID int64, name string) {
	count, err := store.CountHabits(ctx, categoryID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if count >= maxHabitsPerCategory {
		send(bot, tgbotapi.NewMessage(chatID, "This category already has the maximum of 10 habits."))
		return
	}
	if _, err := store.CreateHabit(ctx, categoryID, name); err != nil {
		sendErr(bot, chatID)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, "Habit \""+name+"\" created."))
}

func finishRenameHabit(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, habitID int64, name string) {
	if err := store.RenameHabit(ctx, userID, habitID, name); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			send(bot, tgbotapi.NewMessage(chatID, "Habit not found."))
			return
		}
		sendErr(bot, chatID)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, "Habit renamed to \""+name+"\"."))
}
