package store

import (
	"context"

	"github.com/kek-flip/scotch-api/internal/model"
)

type MatchRepository struct {
	s *Store
}

func (r *MatchRepository) Create(m *model.Match) error {
	if err := m.Validate(); err != nil {
		return err
	}

	row := r.s.db.QueryRow(
		context.Background(),
		"INSERT INTO matches(user_1, user_2) VALUES ($1, $2) RETURNING match_id",
		m.User1,
		m.User2,
	)

	return row.Scan(&m.ID)
}

func (r *MatchRepository) FindByUser(userID int) ([]*model.Match, error) {
	matches := make([]*model.Match, 0)

	rows, err := r.s.db.Query(
		context.Background(),
		"SELECT * FROM matches WHERE user_1 = $1 OR user_2 = $1",
		userID,
	)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		m := &model.Match{}
		err := rows.Scan(
			&m.ID,
			&m.User1,
			&m.User2,
		)

		if err != nil {
			return nil, err
		}

		if m.User1 != userID {
			m.User1, m.User2 = m.User2, m.User1
		}

		matches = append(matches, m)
	}

	return matches, nil
}

func (r *MatchRepository) DeleteByUser(id int) error {
	_, err := r.s.db.Exec(
		context.Background(),
		"DELETE FROM matches WHERE user_1 = $1 OR user_2 = $1",
		id,
	)

	return err
}
