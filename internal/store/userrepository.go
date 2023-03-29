package store

import (
	"context"
	"fmt"

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
		`INSERT INTO users(login, encrypted_password, name, age, gender, city, phone_number, about) 
			VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING user_id`,
		u.Login, u.EncryptedPassword, u.Name, u.Age, u.Gender, u.City, u.PhoneNumber, u.About,
	)

	return row.Scan(&u.ID)
}

func (r *UserRepository) FindByFilter(currentUser, minAge, MaxAge int, gender, city string) ([]*model.User, error) {
	users := make([]*model.User, 0)

	rows, err := r.s.db.Query(
		context.Background(),
		"SELECT * FROM users WHERE age BETWEEN $1 AND $2 AND gender = $3 AND city = $4",
		minAge, MaxAge, gender, city,
	)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		u := &model.User{}

		err = rows.Scan(
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

		if u.ID == currentUser {
			continue
		}

		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) find(field string, value interface{}) (*model.User, error) {
	u := &model.User{}
	err := r.s.db.QueryRow(
		context.Background(),
		fmt.Sprintf("SELECT * FROM users WHERE %s = $1", field),
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

func (r *UserRepository) Update(u *model.User) error {
	_, err := r.s.db.Exec(
		context.Background(),
		`UPDATE users SET 
			login = $1, 
			name = $2, 
			age = $3, 
			gender = $4, 
			city = $5, 
			phone_number = $6, 
			about = $7
		WHERE user_id = $8`,
		u.Login, u.Name, u.Age, u.Gender, u.City, u.PhoneNumber, u.About,
		u.ID,
	)

	return err
}

func (r *UserRepository) delete(field string, value interface{}) error {
	_, err := r.s.db.Exec(context.Background(), fmt.Sprintf("DELETE FROM users WHERE %s = $1", field), value)
	return err
}

func (r *UserRepository) DeleteById(id int) error {
	return r.delete("user_id", id)
}

func (r *UserRepository) DeleteByLogin(login string) error {
	return r.delete("login", login)
}
