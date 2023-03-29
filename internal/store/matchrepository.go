package store

import (
	"context"
	"fmt"

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

func (r *MatchRepository) find(field string, value interface{}) ([]*model.Match, error) {
	matches := make([]*model.Match, 0)

	rows, err := r.s.db.Query(
		context.Background(),
		fmt.Sprintf("SELECT * FROM matches WHERE %s = $1", field),
		value,
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

		matches = append(matches, m)
	}

	return matches, nil
}

func (r *MatchRepository) FindById(id int) (*model.Match, error) {
	matches, err := r.find("match_id", id)
	return matches[0], err
}

func (r *MatchRepository) FindByUser(userId int) ([]*model.Match, error) {
	matchesBy1, err1 := r.find("user_1", userId)
	matchesBy2, err2 := r.find("user_2", userId)

	if err1 != nil && err2 != nil {
		return nil, err1
	}

	for i := range matchesBy2 {
		matchesBy2[i].User1, matchesBy2[i].User2 = matchesBy2[i].User2, matchesBy2[i].User1
	}

	return append(matchesBy1, matchesBy2...), nil
}

func (r *MatchRepository) DeleteByUser(id int) error {
	_, err := r.s.db.Exec(
		context.Background(),
		"DELETE FROM matches WHERE user_1 = $1 OR user_2 = $1",
		id,
	)

	return err
}
