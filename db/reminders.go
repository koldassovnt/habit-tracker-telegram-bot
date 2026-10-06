package db

import (
	"context"
	"time"
)

// SetReminder turns the user's daily reminder on at the given hour (0-23).
func (s *Store) SetReminder(ctx context.Context, userID int64, hour int) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO reminders (user_id, hour)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE
		SET hour = EXCLUDED.hour, enabled = true
	`, userID, hour)
	return err
}

func (s *Store) DisableReminder(ctx context.Context, userID int64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE reminders SET enabled = false WHERE user_id = $1
	`, userID)
	return err
}

// DueReminders returns the users whose reminder hour is now's hour and who
// have not been handled yet on now's day.
func (s *Store) DueReminders(ctx context.Context, now time.Time) ([]int64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.user_id
		FROM reminders r
		JOIN users u ON u.id = r.user_id
		WHERE r.enabled = true AND u.actual = true AND r.hour = $1
		  AND (r.last_sent_on IS NULL OR r.last_sent_on < $2::date)
	`, now.Hour(), now.Format(time.DateOnly))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) MarkReminderSent(ctx context.Context, userID int64, day time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE reminders SET last_sent_on = $2::date WHERE user_id = $1
	`, userID, day.Format(time.DateOnly))
	return err
}
