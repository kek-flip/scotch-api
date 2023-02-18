package model

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation"
)

type Like struct {
	ID        int `json:"id"`
	UserID    int `json:"user_id"`
	LikedUser int `json:"liked_user"`
}

func (l *Like) Validate() error {
	err := validation.ValidateStruct(
		l,
		validation.Field(&l.UserID, validation.Required, validation.Min(1)),
		validation.Field(&l.LikedUser, validation.Required, validation.Min(1)),
	)

	if l.UserID == l.LikedUser {
		err = errors.New("user_id cannot equal liked_user")
	}

	return err
}
