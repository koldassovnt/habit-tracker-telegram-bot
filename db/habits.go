package db

import "context"

func (s *Store) ListHabits(ctx context.Context, categoryID int64) ([]Habit, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, category_id, name FROM habits
		WHERE category_id = $1 AND actual = true
		ORDER BY name
	`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []Habit
	for rows.Next() {
		var h Habit
		if err := rows.Scan(&h.ID, &h.CategoryID, &h.Name); err != nil {
			return nil, err
		}
		habits = append(habits, h)
	}
	return habits, rows.Err()
}

func (s *Store) CountHabits(ctx context.Context, categoryID int64) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM habits WHERE category_id = $1 AND actual = true
	`, categoryID).Scan(&count)
	return count, err
}

func (s *Store) GetHabit(ctx context.Context, userID, habitID int64) (Habit, error) {
	var h Habit
	err := s.pool.QueryRow(ctx, `
		SELECT h.id, h.category_id, h.name
		FROM habits h
		JOIN categories c ON c.id = h.category_id
		WHERE h.id = $1 AND c.user_id = $2 AND h.actual = true AND c.actual = true
	`, habitID, userID).Scan(&h.ID, &h.CategoryID, &h.Name)
	if err != nil {
		if isNoRows(err) {
			return Habit{}, ErrNotFound
		}
		return Habit{}, err
	}
	return h, nil
}

func (s *Store) CreateHabit(ctx context.Context, categoryID int64, name string) (Habit, error) {
	h := Habit{CategoryID: categoryID, Name: name}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO habits (category_id, name) VALUES ($1, $2) RETURNING id
	`, categoryID, name).Scan(&h.ID)
	return h, err
}

func (s *Store) RenameHabit(ctx context.Context, userID, habitID int64, name string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE habits h SET name = $1, modified_at = now()
		FROM categories c
		WHERE h.id = $2 AND h.category_id = c.id AND c.user_id = $3
		AND h.actual = true AND c.actual = true
	`, name, habitID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteHabit(ctx context.Context, userID, habitID int64) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE habits h SET actual = false, modified_at = now()
		FROM categories c
		WHERE h.id = $1 AND h.category_id = c.id AND c.user_id = $2
		AND h.actual = true AND c.actual = true
	`, habitID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
