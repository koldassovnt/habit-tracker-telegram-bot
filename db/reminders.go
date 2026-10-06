package db

import (
	"context"
	"time"
)

// SetHabitReminder sets the habit's daily reminder hour (0-23). Setting it
// clears last_sent_on, so a reminder moved to a later hour can fire again today.
func (s *Store) SetHabitReminder(ctx context.Context, userID, habitID int64, hour int) (habitName string, err error) {
	habit, err := s.GetHabit(ctx, userID, habitID)
	if err != nil {
		return "", err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO habit_reminders (habit_id, hour)
		VALUES ($1, $2)
		ON CONFLICT (habit_id) DO UPDATE
		SET hour = EXCLUDED.hour, last_sent_on = NULL
	`, habitID, hour)
	return habit.Name, err
}

func (s *Store) RemoveHabitReminder(ctx context.Context, userID, habitID int64) (habitName string, err error) {
	habit, err := s.GetHabit(ctx, userID, habitID)
	if err != nil {
		return "", err
	}

	_, err = s.pool.Exec(ctx, `DELETE FROM habit_reminders WHERE habit_id = $1`, habitID)
	return habit.Name, err
}

// HabitReminderHour reports the habit's reminder hour, and false when it has none.
func (s *Store) HabitReminderHour(ctx context.Context, habitID int64) (hour int, ok bool, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT hour FROM habit_reminders WHERE habit_id = $1
	`, habitID).Scan(&hour)
	if err != nil {
		if isNoRows(err) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return hour, true, nil
}

// ListHabitReminders returns the user's reminders on active habits, earliest hour first.
func (s *Store) ListHabitReminders(ctx context.Context, userID int64) ([]HabitReminder, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT h.name, r.hour
		FROM habit_reminders r
		JOIN habits h ON h.id = r.habit_id
		JOIN categories c ON c.id = h.category_id
		WHERE c.user_id = $1 AND h.actual = true AND c.actual = true
		ORDER BY r.hour, h.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []HabitReminder
	for rows.Next() {
		var r HabitReminder
		if err := rows.Scan(&r.HabitName, &r.Hour); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// DueHabitReminders returns the reminders whose hour is now's hour and which
// have not been handled yet on now's day, ordered by user.
func (s *Store) DueHabitReminders(ctx context.Context, now time.Time) ([]DueReminder, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.user_id, h.id, h.category_id, h.name,
		       EXISTS (
		           SELECT 1 FROM habit_logs hl
		           WHERE hl.habit_id = h.id AND hl.tracked_at = $2::date
		       )
		FROM habit_reminders r
		JOIN habits h ON h.id = r.habit_id
		JOIN categories c ON c.id = h.category_id
		JOIN users u ON u.id = c.user_id
		WHERE r.hour = $1
		  AND (r.last_sent_on IS NULL OR r.last_sent_on < $2::date)
		  AND h.actual = true AND c.actual = true AND u.actual = true
		ORDER BY c.user_id, c.name, h.name
	`, now.Hour(), now.Format(time.DateOnly))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DueReminder
	for rows.Next() {
		var d DueReminder
		if err := rows.Scan(&d.UserID, &d.Habit.ID, &d.Habit.CategoryID, &d.Habit.Name, &d.TrackedToday); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) MarkHabitRemindersSent(ctx context.Context, habitIDs []int64, day time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE habit_reminders SET last_sent_on = $2::date WHERE habit_id = ANY($1)
	`, habitIDs, day.Format(time.DateOnly))
	return err
}
