package inputs

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
)

// Reminder hours offered by /reminder, inclusive.
const (
	reminderFirstHour = 6
	reminderLastHour  = 23
)

func handleReminderSet(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64, hour int) {
	if hour < reminderFirstHour || hour > reminderLastHour {
		return
	}
	if err := store.SetReminder(ctx, userID, hour); err != nil {
		sendErr(bot, chatID)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, fmt.Sprintf("⏰ Reminder set for %02d:00 every day.", hour)))
}

func handleReminderOff(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64) {
	if err := store.DisableReminder(ctx, userID); err != nil {
		sendErr(bot, chatID)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, "Reminder turned off."))
}

// sendDueReminders messages every user whose reminder hour is now's hour with
// the habits they haven't tracked today. A user with nothing left to track
// gets no message. Either way the user is marked as handled for the day, so a
// restart inside the hour can't send a second reminder.
func sendDueReminders(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, now time.Time) {
	userIDs, err := store.DueReminders(ctx, now)
	if err != nil {
		log.Printf("scheduler: failed to list due reminders: %v", err)
		return
	}

	for _, uid := range userIDs {
		habits, err := store.UntrackedHabits(ctx, uid, now)
		if err != nil {
			log.Printf("scheduler: UntrackedHabits failed for user %d: %v", uid, err)
			continue
		}
		if len(habits) > 0 {
			msg := tgbotapi.NewMessage(uid, "⏰ Not tracked yet today:")
			msg.ReplyMarkup = trackHabitKeyboard(habits)
			send(bot, msg)
		}
		if err := store.MarkReminderSent(ctx, uid, now); err != nil {
			log.Printf("scheduler: MarkReminderSent failed for user %d: %v", uid, err)
		}
	}
}
