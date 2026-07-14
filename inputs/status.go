package inputs

import (
	"fmt"
	"strings"

	"github.com/koldassovnt/habit-tracker-telegram-bot/db"
)

// periodRowsToStatus flattens single-day PeriodStatus output onto the StatusRow
// shape formatStatus renders. Only meaningful when the period is one day.
func periodRowsToStatus(rows []db.PeriodLogRow) []db.StatusRow {
	statuses := make([]db.StatusRow, 0, len(rows))
	for _, r := range rows {
		statuses = append(statuses, db.StatusRow{
			CategoryName: r.CategoryName,
			HabitName:    r.HabitName,
			Count:        r.Count,
		})
	}
	return statuses
}

func formatStatus(title string, rows []db.StatusRow) string {
	var b strings.Builder
	b.WriteString(title + "\n")

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
