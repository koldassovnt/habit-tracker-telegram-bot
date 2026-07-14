package inputs

import (
	"context"
	"errors"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
)

func handleTrackPastCategoryPick(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, categoryID int64) {
	listHabitsForDayFlow(ctx, bot, store, chatID, userID, categoryID, "trackpast:habit:", "Choose a habit to track for a past day:")
}

func handleUntrackCategoryPick(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, categoryID int64) {
	listHabitsForDayFlow(ctx, bot, store, chatID, userID, categoryID, "untrack:habit:", "Choose a habit to untrack:")
}

// listHabitsForDayFlow validates the category belongs to the user, then offers
// its habits as buttons carrying callbackPrefix.
func listHabitsForDayFlow(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, categoryID int64, callbackPrefix, prompt string) {
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

	msg := tgbotapi.NewMessage(chatID, prompt)
	msg.ReplyMarkup = habitPickKeyboard(habits, callbackPrefix)
	send(bot, msg)
}

func handleTrackPastHabitPick(bot *tgbotapi.BotAPI, chatID, habitID int64) {
	askForDay(bot, chatID, habitID, "trackpast:day:", "Which day should I track it for?")
}

func handleUntrackHabitPick(bot *tgbotapi.BotAPI, chatID, habitID int64) {
	askForDay(bot, chatID, habitID, "untrack:day:", "Which day should I remove it from?")
}

func askForDay(bot *tgbotapi.BotAPI, chatID, habitID int64, callbackPrefix, prompt string) {
	msg := tgbotapi.NewMessage(chatID, prompt)
	msg.ReplyMarkup = datePickKeyboard(fmt.Sprintf("%s%d:", callbackPrefix, habitID), time.Now())
	send(bot, msg)
}

func handleTrackPastDayPick(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, habitID int64, day time.Time) {
	habitName, count, err := store.TrackHabitOn(ctx, userID, habitID, day)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			send(bot, tgbotapi.NewMessage(chatID, "Habit not found."))
			return
		}
		sendErr(bot, chatID)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, fmt.Sprintf("Tracked %q for %s! (%d that day)", habitName, dayLabel(day, time.Now()), count)))
}

func handleUntrackDayPick(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, habitID int64, day time.Time) {
	habitName, count, err := store.UntrackHabitOn(ctx, userID, habitID, day)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrNotFound):
			send(bot, tgbotapi.NewMessage(chatID, "Habit not found."))
		case errors.Is(err, db.ErrNoLog):
			send(bot, tgbotapi.NewMessage(chatID, fmt.Sprintf("Nothing tracked for %q on %s.", habitName, dayLabel(day, time.Now()))))
		default:
			sendErr(bot, chatID)
		}
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, fmt.Sprintf("Removed one %q from %s. (%d left that day)", habitName, dayLabel(day, time.Now()), count)))
}
