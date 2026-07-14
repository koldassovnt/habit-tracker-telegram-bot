package inputs

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
)

func HandleUpdate(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, update tgbotapi.Update) {
	if user := extractUser(update); user != nil {
		if err := store.UpsertUser(ctx, user.ID, user.UserName); err != nil {
			log.Printf("failed to upsert user: %v", err)
		}
	}

	if update.CallbackQuery != nil {
		handleCallback(ctx, bot, store, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	if update.Message.IsCommand() {
		handleCommand(ctx, bot, store, update)
		return
	}

	if sess, ok := getSession(update.Message.Chat.ID); ok {
		handleSessionInput(ctx, bot, store, update.Message, sess)
		return
	}

	send(bot, helpMessage(update.Message.Chat.ID))
}

func extractUser(update tgbotapi.Update) *tgbotapi.User {
	if update.Message != nil {
		return update.Message.From
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From
	}
	return nil
}

func handleCommand(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

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
		categories, err := store.ListCategories(ctx, userID)
		if err != nil {
			sendErr(bot, chatID)
			return
		}

		if len(categories) == 0 {
			send(bot, tgbotapi.NewMessage(chatID, "No categories found. Add one with /managecategory"))
			return
		}

		msg := tgbotapi.NewMessage(chatID, "Choose a category:")
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = trackHabitCategoryKeyboard(categories)
		send(bot, msg)

	case "trackpast":
		sendCategoryPicker(ctx, bot, store, chatID, userID, "trackpast:cat:", "Choose a category:")

	case "untrack":
		sendCategoryPicker(ctx, bot, store, chatID, userID, "untrack:cat:", "Choose a category:")

	case "paststatus":
		msg := tgbotapi.NewMessage(chatID, "Which day do you want the status for?")
		msg.ReplyMarkup = datePickKeyboard("paststatus:day:", time.Now())
		send(bot, msg)

	case "todaystatus":
		rows, err := store.TodayStatus(ctx, userID)
		if err != nil {
			sendErr(bot, chatID)
			return
		}
		if len(rows) == 0 {
			send(bot, tgbotapi.NewMessage(chatID, "You have no habits yet. Add one with /managehabit"))
			return
		}
		send(bot, tgbotapi.NewMessage(chatID, formatStatus("Today's status:", rows)))

	case "help":
		send(bot, helpMessage(chatID))

	default:
		send(bot, tgbotapi.NewMessage(chatID, "I don't know that command. Type /help to see available commands."))
	}
}

// sendCategoryPicker offers the user's categories as buttons carrying callbackPrefix.
func sendCategoryPicker(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64, callbackPrefix, prompt string) {
	categories, err := store.ListCategories(ctx, userID)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if len(categories) == 0 {
		send(bot, tgbotapi.NewMessage(chatID, "No categories found. Add one with /managecategory"))
		return
	}

	msg := tgbotapi.NewMessage(chatID, prompt)
	msg.ReplyMarkup = categoryPickKeyboard(categories, callbackPrefix)
	send(bot, msg)
}

func handlePastStatus(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, chatID, userID int64, day time.Time) {
	rows, err := store.PeriodStatus(ctx, userID, day, day)
	if err != nil {
		sendErr(bot, chatID)
		return
	}
	if len(rows) == 0 {
		send(bot, tgbotapi.NewMessage(chatID, "You have no habits yet. Add one with /managehabit"))
		return
	}
	title := fmt.Sprintf("Status for %s:", dayLabel(day, time.Now()))
	send(bot, tgbotapi.NewMessage(chatID, formatStatus(title, periodRowsToStatus(rows))))
}

func handleCallback(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, cb *tgbotapi.CallbackQuery) {
	bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	chatID := cb.Message.Chat.ID
	userID := cb.From.ID
	data := cb.Data

	switch {
	case data == "category:add":
		handleCategoryAdd(bot, chatID)
	case data == "category:edit":
		handleCategoryEditList(ctx, bot, store, chatID, userID)
	case data == "category:delete":
		handleCategoryDeleteList(ctx, bot, store, chatID, userID)
	case strings.HasPrefix(data, "category:edit:"):
		id, err := parseID(data, "category:edit:")
		if err == nil {
			handleCategoryEditPick(bot, chatID, id)
		}
	case strings.HasPrefix(data, "category:delete:"):
		id, err := parseID(data, "category:delete:")
		if err == nil {
			handleCategoryDeletePick(ctx, bot, store, chatID, userID, id)
		}

	case data == "habit:add":
		handleHabitAddCategoryList(ctx, bot, store, chatID, userID)
	case strings.HasPrefix(data, "habit:addcat:"):
		id, err := parseID(data, "habit:addcat:")
		if err == nil {
			handleHabitAddPickCategory(bot, chatID, id)
		}
	case data == "habit:edit":
		handleHabitEditCategoryList(ctx, bot, store, chatID, userID)
	case strings.HasPrefix(data, "habit:editcat:"):
		id, err := parseID(data, "habit:editcat:")
		if err == nil {
			handleHabitEditListForCategory(ctx, bot, store, chatID, id)
		}
	case strings.HasPrefix(data, "habit:edit:"):
		id, err := parseID(data, "habit:edit:")
		if err == nil {
			handleHabitEditPick(bot, chatID, id)
		}
	case data == "habit:delete":
		handleHabitDeleteCategoryList(ctx, bot, store, chatID, userID)
	case strings.HasPrefix(data, "habit:deletecat:"):
		id, err := parseID(data, "habit:deletecat:")
		if err == nil {
			handleHabitDeleteListForCategory(ctx, bot, store, chatID, id)
		}
	case strings.HasPrefix(data, "habit:delete:"):
		id, err := parseID(data, "habit:delete:")
		if err == nil {
			handleHabitDeletePick(ctx, bot, store, chatID, userID, id)
		}

	case strings.HasPrefix(data, "track:category:"):
		id, err := parseID(data, "track:category:")
		if err == nil {
			handleTrackCategoryPick(ctx, bot, store, chatID, userID, id)
		}
	case strings.HasPrefix(data, "track:habit:"):
		id, err := parseID(data, "track:habit:")
		if err == nil {
			handleTrackHabitPick(ctx, bot, store, chatID, userID, id)
		}

	case strings.HasPrefix(data, "trackpast:cat:"):
		id, err := parseID(data, "trackpast:cat:")
		if err == nil {
			handleTrackPastCategoryPick(ctx, bot, store, chatID, userID, id)
		}
	case strings.HasPrefix(data, "trackpast:habit:"):
		id, err := parseID(data, "trackpast:habit:")
		if err == nil {
			handleTrackPastHabitPick(bot, chatID, id)
		}
	case strings.HasPrefix(data, "trackpast:day:"):
		id, day, err := parseIDAndDate(data, "trackpast:day:")
		if err == nil {
			handleTrackPastDayPick(ctx, bot, store, chatID, userID, id, day)
		}

	case strings.HasPrefix(data, "untrack:cat:"):
		id, err := parseID(data, "untrack:cat:")
		if err == nil {
			handleUntrackCategoryPick(ctx, bot, store, chatID, userID, id)
		}
	case strings.HasPrefix(data, "untrack:habit:"):
		id, err := parseID(data, "untrack:habit:")
		if err == nil {
			handleUntrackHabitPick(bot, chatID, id)
		}
	case strings.HasPrefix(data, "untrack:day:"):
		id, day, err := parseIDAndDate(data, "untrack:day:")
		if err == nil {
			handleUntrackDayPick(ctx, bot, store, chatID, userID, id, day)
		}

	case strings.HasPrefix(data, "paststatus:day:"):
		day, err := time.ParseInLocation(dateCallbackLayout, strings.TrimPrefix(data, "paststatus:day:"), time.Local)
		if err == nil {
			handlePastStatus(ctx, bot, store, chatID, userID, day)
		}
	}
}

func handleSessionInput(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, msg *tgbotapi.Message, sess session) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	name := strings.TrimSpace(msg.Text)
	clearSession(chatID)

	if name == "" {
		send(bot, tgbotapi.NewMessage(chatID, "Name can't be empty — please start over."))
		return
	}

	switch sess.flow {
	case flowAddCategory:
		finishAddCategory(ctx, bot, store, chatID, userID, name)
	case flowRenameCategory:
		finishRenameCategory(ctx, bot, store, chatID, userID, sess.categoryID, name)
	case flowAddHabit:
		finishAddHabit(ctx, bot, store, chatID, sess.categoryID, name)
	case flowRenameHabit:
		finishRenameHabit(ctx, bot, store, chatID, userID, sess.habitID, name)
	}
}

func parseID(data, prefix string) (int64, error) {
	return strconv.ParseInt(strings.TrimPrefix(data, prefix), 10, 64)
}

// parseIDAndDate reads callback data shaped "<prefix><id>:<YYYY-MM-DD>".
func parseIDAndDate(data, prefix string) (int64, time.Time, error) {
	rest := strings.TrimPrefix(data, prefix)
	idPart, dayPart, ok := strings.Cut(rest, ":")
	if !ok {
		return 0, time.Time{}, fmt.Errorf("malformed callback data %q", data)
	}

	id, err := strconv.ParseInt(idPart, 10, 64)
	if err != nil {
		return 0, time.Time{}, err
	}

	day, err := time.ParseInLocation(dateCallbackLayout, dayPart, time.Local)
	if err != nil {
		return 0, time.Time{}, err
	}
	return id, day, nil
}

func helpMessage(chatID int64) tgbotapi.MessageConfig {
	text := `🛠 *Available Commands*

/managecategory — Add, edit, delete category
/managehabit — Add, edit, delete habit
/trackhabit — Track a habit
/trackpast — Track a habit for a past day
/untrack — Remove a tracked habit
/todaystatus — Status of tracked habits for today
/paststatus — Status of tracked habits for a past day
/help — Show this help message`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	return msg
}

// telegramMessageLimit is Telegram's max character count for a single message.
const telegramMessageLimit = 4096

func send(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) {
	chunks := splitMessage(msg.Text, telegramMessageLimit)
	for i, chunk := range chunks {
		part := msg
		part.Text = chunk
		if i != len(chunks)-1 {
			part.ReplyMarkup = nil // only the last chunk gets any keyboard
		}
		if _, err := bot.Send(part); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
}

// splitMessage breaks text into chunks of at most limit characters, preferring
// to break on line boundaries so a single habit/day entry doesn't get cut in half.
func splitMessage(text string, limit int) []string {
	if len(text) <= limit {
		return []string{text}
	}

	var chunks []string
	var b strings.Builder

	for _, line := range strings.SplitAfter(text, "\n") {
		for len(line) > limit {
			if b.Len() > 0 {
				chunks = append(chunks, b.String())
				b.Reset()
			}
			chunks = append(chunks, line[:limit])
			line = line[limit:]
		}
		if b.Len()+len(line) > limit && b.Len() > 0 {
			chunks = append(chunks, b.String())
			b.Reset()
		}
		b.WriteString(line)
	}
	if b.Len() > 0 {
		chunks = append(chunks, b.String())
	}
	return chunks
}

func sendErr(bot *tgbotapi.BotAPI, chatID int64) {
	send(bot, tgbotapi.NewMessage(chatID, "Something went wrong, please try again."))
}
