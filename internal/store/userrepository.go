package store

import (
	"context"

	"github.com/kek-flip/scotch-api/internal/model"
)

type UserRepository struct {
	s *Store
}

func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.EncryptPassword(); err != nil {
		return err
	}

	row := r.s.db.QueryRow(
		context.Background(),
		"INSERT INTO users(login, encrypted_password, name, age, gender, city, phone_number, about) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING user_id",
		u.Login,
		u.EncryptedPassword,
		u.Name,
		u.Age,
		u.Gender,
		u.City,
		u.PhoneNumber,
		u.About,
	)

	return row.Scan(&u.ID)
}

func (r *UserRepository) find(field string, value interface{}) (*model.User, error) {
	u := &model.User{}
	err := r.s.db.QueryRow(
		context.Background(),
		"SELECT * FROM users WHERE $1 = $2",
		field,
		value,
	).Scan(
		&u.ID,
		&u.Login,
		&u.EncryptedPassword,
		&u.Name,
		&u.Age,
		&u.Gender,
		&u.City,
		&u.PhoneNumber,
		&u.About,
	)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) FindById(id int) (*model.User, error) {
	return r.find("user_id", id)
}

func (r *UserRepository) FindByLogin(login string) (*model.User, error) {
	return r.find("login", login)
}

// TODO: write func Update()

func (r *UserRepository) delete(field string, value interface{}) error {
	_, err := r.s.db.Exec(context.Background(), "DELETE FROM users WHERE $1 = $2", field, value)
	return err
}

func (r *UserRepository) DeleteById(id int) error {
	return r.delete("user_id", id)
}

func (r *UserRepository) DeleteByLogin(login string) error {
	return r.delete("login", login)
}
