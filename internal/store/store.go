package store

import (
	"github.com/jackc/pgx/v5"
)

type Store struct {
	db             *pgx.Conn
	userRepository *UserRepository
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
