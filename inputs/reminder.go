package inputs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
)

// Reminder hours offered by /reminder, inclusive.
const (
	reminderFirstHour = 6
	reminderLastHour  = 23
)

// handleReminderCommand lists the user's current reminders and offers their
// categories, the first step of picking a habit to set a reminder for.
func handleReminderCommand(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64) {
	reminders, err := store.ListHabitReminders(ctx, userID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}

	var b strings.Builder
	if len(reminders) == 0 {
		b.WriteString("You have no reminders yet.\n")
	} else {
		b.WriteString("⏰ Your reminders:\n")
		for _, r := range reminders {
			fmt.Fprintf(&b, "%02d:00 — %s\n", r.Hour, r.HabitName)
		}
	}
	b.WriteString("\nChoose a category to set or change a habit's reminder:")

	sendCategoryPicker(ctx, bot, store, chatID, userID, "reminder:cat:", b.String())
}

func handleReminderCategoryPick(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, categoryID int64) {
	listHabitsForDayFlow(ctx, bot, store, chatID, userID, categoryID, "reminder:habit:", "Choose a habit:")
}

func handleReminderHabitPick(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, habitID int64) {
	habit, err := store.GetHabit(ctx, userID, habitID)
	if err != nil {
		sendHabitErr(bot, chatID, err)
		return
	}
	hour, ok, err := store.HabitReminderHour(ctx, habitID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}

	text := fmt.Sprintf("%q has no reminder. Pick an hour:", habit.Name)
	if ok {
		text = fmt.Sprintf("%q is reminded at %02d:00. Pick a new hour or turn it off:", habit.Name, hour)
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = reminderHourKeyboard(habitID)
	send(bot, msg)
}

func handleReminderSet(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, habitID int64, hour int) {
	if hour < reminderFirstHour || hour > reminderLastHour {
		return
	}
	habitName, err := store.SetHabitReminder(ctx, userID, habitID, hour)
	if err != nil {
		sendHabitErr(bot, chatID, err)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, fmt.Sprintf("⏰ I'll remind you about %q at %02d:00 every day.", habitName, hour)))
}

func handleReminderOff(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID, habitID int64) {
	habitName, err := store.RemoveHabitReminder(ctx, userID, habitID)
	if err != nil {
		sendHabitErr(bot, chatID, err)
		return
	}
	send(bot, tgbotapi.NewMessage(chatID, fmt.Sprintf("Reminder for %q turned off.", habitName)))
}

func sendHabitErr(bot *tgbotapi.BotAPI, chatID int64, err error) {
	if errors.Is(err, db.ErrNotFound) {
		send(bot, tgbotapi.NewMessage(chatID, "Habit not found."))
		return
	}
	sendErr(bot, chatID)
}

// parseReminderSet reads callback data shaped "reminder:set:<habitID>:<hour>".
func parseReminderSet(data string) (habitID int64, hour int, err error) {
	idPart, hourPart, ok := strings.Cut(strings.TrimPrefix(data, "reminder:set:"), ":")
	if !ok {
		return 0, 0, fmt.Errorf("malformed callback data %q", data)
	}
	if habitID, err = strconv.ParseInt(idPart, 10, 64); err != nil {
		return 0, 0, err
	}
	if hour, err = strconv.Atoi(hourPart); err != nil {
		return 0, 0, err
	}
	return habitID, hour, nil
}

// sendDueReminders sends each user one message with the habits whose reminder
// hour is now's hour and which aren't tracked yet today. Habits already
// tracked are skipped. Every due reminder is marked as handled for the day, so
// a restart inside the hour can't send it twice.
func sendDueReminders(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, now time.Time) {
	due, err := store.DueHabitReminders(ctx, now)
	if err != nil {
		log.Printf("scheduler: failed to list due reminders: %v", err)
		return
	}
	if len(due) == 0 {
		return
	}

	// due is ordered by user, so each user's habits are consecutive.
	var handled []int64
	for start := 0; start < len(due); {
		uid := due[start].UserID
		var untracked []db.Habit
		end := start
		for ; end < len(due) && due[end].UserID == uid; end++ {
			handled = append(handled, due[end].Habit.ID)
			if !due[end].TrackedToday {
				untracked = append(untracked, due[end].Habit)
			}
		}
		start = end

		if len(untracked) > 0 {
			msg := tgbotapi.NewMessage(uid, "⏰ Reminder — not tracked yet today:")
			msg.ReplyMarkup = trackHabitKeyboard(untracked)
			send(bot, msg)
		}
	}

	if err := store.MarkHabitRemindersSent(ctx, handled, now); err != nil {
		log.Printf("scheduler: MarkHabitRemindersSent failed: %v", err)
	}
}
