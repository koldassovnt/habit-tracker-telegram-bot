package inputs

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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
