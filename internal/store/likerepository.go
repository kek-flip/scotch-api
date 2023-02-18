package store

import (
	"context"

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
