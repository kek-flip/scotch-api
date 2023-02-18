package store

import (
	"github.com/jackc/pgx/v5"
)

type Store struct {
	db             *pgx.Conn
	userRepository *UserRepository
	likeRepository *LikeRepository
}

func NewStore(db *pgx.Conn) *Store {
	return &Store{db: db}
}

func (s *Store) User() *UserRepository {
	if s.userRepository == nil {
		s.userRepository = &UserRepository{s}
	}
	return s.userRepository
}

func (s *Store) Like() *LikeRepository {
	if s.likeRepository == nil {
		s.likeRepository = &LikeRepository{s}
	}
	return s.likeRepository
}
