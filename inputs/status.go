package inputs

import (
	"fmt"
	"strings"

	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
)

func formatStatus(rows []db.StatusRow) string {
	var b strings.Builder
	b.WriteString("Today's status:\n")

	currentCategory := ""
	for _, r := range rows {
		if r.CategoryName != currentCategory {
			currentCategory = r.CategoryName
			fmt.Fprintf(&b, "\n%s\n", currentCategory)
		}
		if r.Count == 0 {
			fmt.Fprintf(&b, "❌ %s\n", r.HabitName)
		} else {
			fmt.Fprintf(&b, "✅ %s ×%d\n", r.HabitName, r.Count)
		}
	}

	return b.String()
}
