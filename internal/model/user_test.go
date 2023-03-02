package model_test

import (
	"strings"
	"testing"

	"github.com/kek-flip/scotch-api/internal/model"
	"github.com/stretchr/testify/assert"
)

func testUser(t *testing.T) *model.User {
	t.Helper()

	return &model.User{
		Login:       "valid_login",
		Password:    "valid_password",
		Name:        "valid_name",
		Age:         20,
		Gender:      "male",
		City:        "valid_city",
		PhoneNumber: "+79999999999",
		About:       "example text",
	}
}

func TestUser_Validation(t *testing.T) {
	testCases := []struct {
		name    string
		u       func() *model.User
		isValid bool
	}{
		{
			name: "Valid User",
			u: func() *model.User {
				u := testUser(t)
				return u
			},
			isValid: true,
		},
		{
			name: "Too short login",
			u: func() *model.User {
				u := testUser(t)
				u.Login = strings.Repeat("a", 2)
				return u
			},
			isValid: false,
		},
		{
			name: "Too long login",
			u: func() *model.User {
				u := testUser(t)
				u.Login = strings.Repeat("a", 26)
				return u
			},
			isValid: false,
		},
		{
			name: "Empty login",
			u: func() *model.User {
				u := testUser(t)
				u.Login = ""
				return u
			},
			isValid: false,
		},
		{
			name: "Too short password",
			u: func() *model.User {
				u := testUser(t)
				u.Password = strings.Repeat("a", 5)
				return u
			},
			isValid: false,
		},
		{
			name: "Too long password",
			u: func() *model.User {
				u := testUser(t)
				u.Password = strings.Repeat("a", 101)
				return u
			},
			isValid: false,
		},
		{
			name: "Empty password",
			u: func() *model.User {
				u := testUser(t)
				u.Password = ""
				return u
			},
			isValid: false,
		},
		{
			name: "Empty name",
			u: func() *model.User {
				u := testUser(t)
				u.Name = ""
				return u
			},
			isValid: false,
		},
		{
			name: "Too young user",
			u: func() *model.User {
				u := testUser(t)
				u.Age = 17
				return u
			},
			isValid: false,
		},
		{
			name: "Too old user",
			u: func() *model.User {
				u := testUser(t)
				u.Age = 100
				return u
			},
			isValid: false,
		},
		{
			name: "Invalid gender",
			u: func() *model.User {
				u := testUser(t)
				u.Gender = "invalid"
				return u
			},
			isValid: false,
		},
		{
			name: "Empty gender",
			u: func() *model.User {
				u := testUser(t)
				u.Gender = ""
				return u
			},
			isValid: false,
		},
		{
			name: "Invalid phone number",
			u: func() *model.User {
				u := testUser(t)
				u.PhoneNumber = "invalid"
				return u
			},
			isValid: false,
		},
		{
			name: "Empty phone number",
			u: func() *model.User {
				u := testUser(t)
				u.PhoneNumber = ""
				return u
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				assert.NoError(t, tc.u().Validate())
			} else {
				assert.Error(t, tc.u().Validate())
			}
		})
	}
}

func TestUser_EncryptPassword(t *testing.T) {
	u := testUser(t)
	assert.NoError(t, u.EncryptPassword())
	assert.NotEmpty(t, u.EncryptedPassword)
}
