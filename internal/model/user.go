package model

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                int    `json:"id"`
	Login             string `json:"login"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
	Name              string `json:"name"`
	Age               int    `json:"age"`
	Gender            string `json:"gender"`
	PhoneNumber       string `json:"phone_number"`
	About             string `json:"about"`
}

func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Login, validation.Required, validation.Length(3, 25)),
		validation.Field(&u.Password, validation.Required, validation.Length(6, 100)),
		validation.Field(&u.Name, validation.Required),
		validation.Field(&u.Age, validation.Required, validation.Min(18), validation.Max(99)),
		validation.Field(&u.Gender, validation.Required, validation.In("male", "female")),
		validation.Field(&u.PhoneNumber, validation.Required, validation.Match(regexp.MustCompile(`\A(\+7|8)9[0-9]{9}\z`))),
	)
}

func (u *User) ClearPassword() {
	u.Password = ""
}

func (u *User) EncryptPassword() error {
	ep, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	u.EncryptedPassword = string(ep)

	return nil
}
