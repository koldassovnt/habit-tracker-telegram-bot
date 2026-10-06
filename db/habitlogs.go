package db

import (
	"context"
	"time"
)

func (s *Store) TrackHabit(ctx context.Context, userID, habitID int64) (habitName string, todayCount int, err error) {
	return s.TrackHabitOn(ctx, userID, habitID, time.Now())
}

// TrackHabitOn adds a log for the given day and returns the habit's resulting
// count for that day. A habit may be tracked any number of times per day.
func (s *Store) TrackHabitOn(ctx context.Context, userID, habitID int64, day time.Time) (habitName string, dayCount int, err error) {
	habit, err := s.GetHabit(ctx, userID, habitID)
	if err != nil {
		return "", 0, err
	}

	if _, err := s.pool.Exec(ctx, `
		INSERT INTO habit_logs (habit_id, tracked_at) VALUES ($1, $2::date)
	`, habitID, day); err != nil {
		return "", 0, err
	}

	dayCount, err = s.countLogsOn(ctx, habitID, day)
	return habit.Name, dayCount, err
}

// UntrackHabitOn removes the most recent single log for the given day and
// returns the habit's remaining count for that day. It reports ErrNoLog when
// the habit has nothing tracked on that day.
func (s *Store) UntrackHabitOn(ctx context.Context, userID, habitID int64, day time.Time) (habitName string, dayCount int, err error) {
	habit, err := s.GetHabit(ctx, userID, habitID)
	if err != nil {
		return "", 0, err
	}

	ct, err := s.pool.Exec(ctx, `
		DELETE FROM habit_logs
		WHERE id = (
			SELECT id FROM habit_logs
			WHERE habit_id = $1 AND tracked_at = $2::date
			ORDER BY id DESC
			LIMIT 1
		)
	`, habitID, day)
	if err != nil {
		return "", 0, err
	}
	if ct.RowsAffected() == 0 {
		return habit.Name, 0, ErrNoLog
	}

	dayCount, err = s.countLogsOn(ctx, habitID, day)
	return habit.Name, dayCount, err
}

func (s *Store) countLogsOn(ctx context.Context, habitID int64, day time.Time) (count int, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM habit_logs WHERE habit_id = $1 AND tracked_at = $2::date
	`, habitID, day).Scan(&count)
	return count, err
}

func (s *Store) TodayStatus(ctx context.Context, userID int64) ([]StatusRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.name, h.name, COUNT(hl.id)
		FROM habits h
		JOIN categories c ON c.id = h.category_id
		LEFT JOIN habit_logs hl ON hl.habit_id = h.id AND hl.tracked_at = CURRENT_DATE
		WHERE c.user_id = $1 AND h.actual = true AND c.actual = true
		GROUP BY c.id, h.id, c.name, h.name
		ORDER BY c.name, h.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []StatusRow
	for rows.Next() {
		var r StatusRow
		if err := rows.Scan(&r.CategoryName, &r.HabitName, &r.Count); err != nil {
			return nil, err
		}
		statuses = append(statuses, r)
	}
	return statuses, rows.Err()
}

func (s *Store) PeriodStatus(ctx context.Context, userID int64, from, to time.Time) ([]PeriodLogRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.name, h.name, gs.day::date, COUNT(hl.id)
		FROM habits h
		JOIN categories c ON c.id = h.category_id
		CROSS JOIN generate_series($2::date, $3::date, interval '1 day') AS gs(day)
		LEFT JOIN habit_logs hl ON hl.habit_id = h.id AND hl.tracked_at = gs.day::date
		WHERE c.user_id = $1 AND h.actual = true AND c.actual = true
		GROUP BY c.id, h.id, c.name, h.name, gs.day
		ORDER BY c.name, h.name, gs.day
	`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PeriodLogRow
	for rows.Next() {
		var r PeriodLogRow
		if err := rows.Scan(&r.CategoryName, &r.HabitName, &r.Date, &r.Count); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
