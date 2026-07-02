package inputs

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
)

func manageCategoryKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Add", "category:add"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Edit", "category:edit"),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Delete", "category:delete"),
		),
	)
}

func manageHabitKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Add", "habit:add"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Edit", "habit:edit"),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Delete", "habit:delete"),
		),
	)
}

func trackHabitCategoryKeyboard(categories []db.Category) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, cat := range categories {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(cat.Name, "track:category:"+strconv.FormatInt(cat.ID, 10)),
		)
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func trackHabitKeyboard(habits []db.Habit) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, h := range habits {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.Name, "track:habit:"+strconv.FormatInt(h.ID, 10)),
		)
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// categoryPickKeyboard lists categories as buttons whose callback data is
// callbackPrefix + the category ID (e.g. "category:edit:" or "habit:addcat:").
func categoryPickKeyboard(categories []db.Category, callbackPrefix string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, cat := range categories {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(cat.Name, callbackPrefix+strconv.FormatInt(cat.ID, 10)),
		)
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// habitPickKeyboard lists habits as buttons whose callback data is
// callbackPrefix + the habit ID (e.g. "habit:edit:" or "habit:delete:").
func habitPickKeyboard(habits []db.Habit, callbackPrefix string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, h := range habits {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(h.Name, callbackPrefix+strconv.FormatInt(h.ID, 10)),
		)
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
