package db

import "context"

func (s *Store) UpsertUser(ctx context.Context, id int64, username string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (id, username)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE
		SET username = EXCLUDED.username, modified_at = now()
	`, id, username)
	return err
}

func (s *Store) ListUserIDs(ctx context.Context) ([]int64, error) {
	rows, err := s.pool.Query(ctx, `SELECT id FROM users WHERE actual = true`)
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
