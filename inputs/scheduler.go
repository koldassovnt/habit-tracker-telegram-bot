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

// StartScheduler blocks until ctx is cancelled, checking once a minute whether
// it's time to push the weekly or monthly report to every user.
func StartScheduler(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	var lastWeekly, lastMonthly string // "2006-01-02" guard, in-memory only

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			today := now.Format("2006-01-02")

			if now.Weekday() == time.Monday && now.Hour() == 8 && now.Minute() == 0 && lastWeekly != today {
				lastWeekly = today
				sendWeeklyReports(ctx, bot, store, now)
			}
			if now.Day() == 1 && now.Hour() == 8 && now.Minute() == 0 && lastMonthly != today {
				lastMonthly = today
				sendMonthlyReports(ctx, bot, store, now)
			}
		}
	}
}

// weekRange returns [Mon, Sun] of the week that just ended, given a Monday-08:00 firing.
func weekRange(now time.Time) (start, end time.Time) {
	end = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location()) // yesterday (Sun)
	start = end.AddDate(0, 0, -6)                                                     // Mon before it
	return start, end
}

// monthRange returns [1st, lastDay] of the month that just ended, given a 1st-08:00 firing.
func monthRange(now time.Time) (start, end time.Time) {
	thisMonth1st := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end = thisMonth1st.AddDate(0, 0, -1)                                        // last day of prev month
	start = time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, now.Location())   // 1st of prev month
	return start, end
}

func sendWeeklyReports(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, now time.Time) {
	start, end := weekRange(now)
	title := fmt.Sprintf("📅 Weekly report (%s – %s)", start.Format("Jan 2"), end.Format("Jan 2"))
	sendPeriodReports(ctx, bot, store, start, end, title, weekdayLabel)
}

func sendMonthlyReports(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, now time.Time) {
	start, end := monthRange(now)
	title := fmt.Sprintf("📅 Monthly report (%s)", start.Format("January 2006"))
	sendPeriodReports(ctx, bot, store, start, end, title, dayOfMonthLabel)
}

func sendPeriodReports(ctx context.Context, bot *tgbotapi.BotAPI, store *db.Store, start, end time.Time, title string, dayLabel func(time.Time) string) {
	userIDs, err := store.ListUserIDs(ctx)
	if err != nil {
		log.Printf("scheduler: failed to list users: %v", err)
		return
	}

	for _, uid := range userIDs {
		rows, err := store.PeriodStatus(ctx, uid, start, end)
		if err != nil {
			log.Printf("scheduler: PeriodStatus failed for user %d: %v", uid, err)
			continue
		}
		if len(rows) == 0 {
			continue
		}
		send(bot, tgbotapi.NewMessage(uid, formatPeriodReport(title, rows, dayLabel)))
	}
}

func weekdayLabel(d time.Time) string    { return d.Weekday().String()[:3] } // "Mon".."Sun"
func dayOfMonthLabel(d time.Time) string { return strconv.Itoa(d.Day()) }    // "1".."31"

func formatPeriodReport(title string, rows []db.PeriodLogRow, dayLabel func(time.Time) string) string {
	var b strings.Builder
	b.WriteString(title + "\n")

	var currentCategory, currentHabit string
	var parts []string
	total := 0

	flushHabit := func() {
		if currentHabit != "" {
			fmt.Fprintf(&b, "%s: %s (total %d)\n", currentHabit, strings.Join(parts, ", "), total)
		}
	}

	for _, r := range rows {
		if r.CategoryName != currentCategory {
			flushHabit()
			currentHabit = ""
			currentCategory = r.CategoryName
			fmt.Fprintf(&b, "\n%s\n", currentCategory)
		}
		if r.HabitName != currentHabit {
			flushHabit()
			currentHabit = r.HabitName
			parts = nil
			total = 0
		}
		parts = append(parts, fmt.Sprintf("%s %d", dayLabel(r.Date), r.Count))
		total += r.Count
	}
	flushHabit()

	return b.String()
}
