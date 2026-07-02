package inputs

import (
	"context"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
)

func handleTrackCategoryPick(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, categoryID int64) {
	if _, err := store.GetCategory(ctx, userID, categoryID); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			send(bot, tgbotapi.NewMessage(chatID, "Category not found."))
			return
		}
		sendErr(bot, chatID)
		return
	}

	habits, err := store.ListHabits(ctx, categoryID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if len(habits) == 0 {
		send(bot, tgbotapi.NewMessage(chatID, "No habits in this category. Add one with /managehabit"))
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Choose a habit to track:")
	msg.ReplyMarkup = trackHabitKeyboard(habits)
	send(bot, msg)
}

func handleTrackHabitPick(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, habitID int64) {
	count, err := store.TrackHabit(ctx, userID, habitID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			send(bot, tgbotapi.NewMessage(chatID, "Habit not found."))
			return
		}
		sendErr(bot, chatID)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, fmt.Sprintf("Tracked! (%d today)", count)))
}
