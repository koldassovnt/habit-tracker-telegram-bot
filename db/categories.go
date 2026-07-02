package db

import "context"

func (s *Store) ListCategories(ctx context.Context, userID int64) ([]Category, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, user_id FROM categories
		WHERE user_id = $1 AND actual = true
		ORDER BY name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.UserID); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (s *Store) CountCategories(ctx context.Context, userID int64) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM categories WHERE user_id = $1 AND actual = true
	`, userID).Scan(&count)
	return count, err
}

func (s *Store) GetCategory(ctx context.Context, userID, categoryID int64) (Category, error) {
	var c Category
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, user_id FROM categories
		WHERE id = $1 AND user_id = $2 AND actual = true
	`, categoryID, userID).Scan(&c.ID, &c.Name, &c.UserID)
	if err != nil {
		if isNoRows(err) {
			return Category{}, ErrNotFound
		}
		return Category{}, err
	}
	return c, nil
}

func (s *Store) CreateCategory(ctx context.Context, userID int64, name string) (Category, error) {
	c := Category{Name: name, UserID: userID}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO categories (name, user_id) VALUES ($1, $2) RETURNING id
	`, name, userID).Scan(&c.ID)
	return c, err
}

func (s *Store) RenameCategory(ctx context.Context, userID, categoryID int64, name string) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE categories SET name = $1, modified_at = now()
		WHERE id = $2 AND user_id = $3 AND actual = true
	`, name, categoryID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteCategory(ctx context.Context, userID, categoryID int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ct, err := tx.Exec(ctx, `
		UPDATE categories SET actual = false, modified_at = now()
		WHERE id = $1 AND user_id = $2 AND actual = true
	`, categoryID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}

	if _, err := tx.Exec(ctx, `
		UPDATE habits SET actual = false, modified_at = now()
		WHERE category_id = $1
	`, categoryID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
