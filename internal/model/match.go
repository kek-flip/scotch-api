package model

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation"
)

type Match struct {
	ID    int `json:"id,omitempty"`
	User1 int `json:"user_1"`
	User2 int `json:"user_2"`
}

func (m *Match) Validate() error {
	err := validation.ValidateStruct(
		m,
		validation.Field(&m.User1, validation.Required),
		validation.Field(&m.User2, validation.Required),
	)

	if m.User1 == m.User2 {
		err = errors.New("user_1 cannot equal user_2")
	}

	return err
}
