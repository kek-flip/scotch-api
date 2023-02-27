package store

import (
	"context"
	"fmt"

	"github.com/kek-flip/scotch-api/internal/model"
)

type LikeRepository struct {
	s *Store
}

func (r *LikeRepository) Create(l *model.Like) error {
	if err := l.Validate(); err != nil {
		return err
	}

	row := r.s.db.QueryRow(
		context.Background(),
		"INSERT INTO likes(user_id, liked_user) VALUES ($1, $2) RETURNING like_id",
		l.UserID,
		l.LikedUser,
	)

	return row.Scan(&l.ID)
}

func (r *LikeRepository) find(field string, value interface{}) ([]*model.Like, error) {
	likes := make([]*model.Like, 0)

	rows, err := r.s.db.Query(
		context.Background(),
		fmt.Sprintf("SELECT * FROM likes WHERE %s = $1", field),
		value,
	)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		l := &model.Like{}

		err := rows.Scan(
			&l.ID,
			&l.UserID,
			&l.LikedUser,
		)

		if err != nil {
			return nil, err
		}

		likes = append(likes, l)
	}

	return likes, nil
}

func (r *LikeRepository) FindById(id int) (*model.Like, error) {
	likes, err := r.find("like_id", id)
	return likes[0], err
}

func (r *LikeRepository) FindByUserID(userID int) ([]*model.Like, error) {
	return r.find("user_id", userID)
}
