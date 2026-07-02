package db

import (
	"context"
	"time"
)

func (s *Store) TrackHabit(ctx context.Context, userID, habitID int64) (todayCount int, err error) {
	if _, err := s.GetHabit(ctx, userID, habitID); err != nil {
		return 0, err
	}

	if _, err := s.pool.Exec(ctx, `INSERT INTO habit_logs (habit_id) VALUES ($1)`, habitID); err != nil {
		return 0, err
	}

	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM habit_logs WHERE habit_id = $1 AND tracked_at = CURRENT_DATE
	`, habitID).Scan(&todayCount)
	return todayCount, err
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
