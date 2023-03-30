package apiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kek-flip/scotch-api/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestServer_handlerSessionCreate(t *testing.T) {
	s, conn := testServer(t)
	defer conn.Close(context.Background())

	testCases := []struct {
		name         string
		data         func() string
		expectedCode int
		isValid      bool
	}{
		{
			name: "Valid user",
			data: func() string {
				u := map[string]string{
					"login":    "valid_login",
					"password": "valid_password",
				}
				d, _ := json.Marshal(u)
				return string(d)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid format",
			data: func() string {
				return "Invalid json"
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid login",
			data: func() string {
				u := map[string]string{
					"login":    "invalid",
					"password": "valid_password",
				}
				d, _ := json.Marshal(u)
				return string(d)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Invalid password",
			data: func() string {
				u := map[string]string{
					"login":    "valid_login",
					"password": "invalid",
				}
				d, _ := json.Marshal(u)
				return string(d)
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	u := testUser(t)
	ud, err := json.Marshal(u)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(ud))
	assert.NoError(t, err)

	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
	json.NewDecoder(rec.Body).Decode(u)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/sessions", bytes.NewReader([]byte(tc.data())))
			assert.NoError(t, err)

			s.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedCode, rec.Code)

			if tc.expectedCode == http.StatusOK {
				assert.NotEmpty(t, rec.Header()["Set-Cookie"])
			}
		})
	}

	_, err = conn.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", u.ID)
	assert.NoError(t, err)
}

func TestServer_handlerUserCreate(t *testing.T) {
	s, conn := testServer(t)
	defer conn.Close(context.Background())

	testCases := []struct {
		name         string
		data         func() string
		isValid      bool
		expectedCode int
	}{
		{
			name: "Valid user",
			data: func() string {
				u := testUser(t)
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      true,
			expectedCode: http.StatusCreated,
		},
		{
			name: "Invalid format",
			data: func() string {
				return "Invalid json"
			},
			isValid:      false,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid user data: too short login",
			data: func() string {
				u := testUser(t)
				u.Login = strings.Repeat("a", 2)
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: too long login",
			data: func() string {
				u := testUser(t)
				u.Login = strings.Repeat("a", 26)
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: empty login",
			data: func() string {
				u := testUser(t)
				u.Login = ""
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: too short password",
			data: func() string {
				u := testUser(t)
				u.Password = strings.Repeat("a", 2)
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: too long paswsword",
			data: func() string {
				u := testUser(t)
				u.Password = strings.Repeat("a", 101)
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: empty password",
			data: func() string {
				u := testUser(t)
				u.Password = ""
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: empty name",
			data: func() string {
				u := testUser(t)
				u.Name = ""
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: too young user",
			data: func() string {
				u := testUser(t)
				u.Age = 17
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: too old user",
			data: func() string {
				u := testUser(t)
				u.Age = 100
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: invalid gender",
			data: func() string {
				u := testUser(t)
				u.Gender = "invalid"
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: empty name",
			data: func() string {
				u := testUser(t)
				u.Gender = ""
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: invalid phone number",
			data: func() string {
				u := testUser(t)
				u.Gender = "invalid"
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "Invalid user data: empty phone number",
			data: func() string {
				u := testUser(t)
				u.PhoneNumber = ""
				d, _ := json.Marshal(u)
				return string(d)
			},
			isValid:      false,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodPost, "/users", strings.NewReader(tc.data()))
			assert.NoError(t, err)

			s.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedCode, rec.Code)

			if tc.isValid {
				u := &model.User{}
				json.NewDecoder(rec.Body).Decode(u)

				assert.NotEmpty(t, u.ID)
				assert.Empty(t, u.Password)
				assert.Empty(t, u.EncryptedPassword)

				_, err = conn.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", u.ID)
				assert.NoError(t, err)
			}
		})
	}
}

func TestServer_handlerCurrentUser(t *testing.T) {
	s, conn := testServer(t)
	defer conn.Close(context.Background())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/current", nil)

	s.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestServer_handlerUser(t *testing.T) {
	s, conn := testServer(t)
	defer conn.Close(context.Background())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)

	s.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestServer_handlerLikeCreate(t *testing.T) {
	s, conn := testServer(t)
	defer conn.Close(context.Background())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/likes", nil)

	s.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestServer_handlerLikeLiked(t *testing.T) {
	s, conn := testServer(t)
	defer conn.Close(context.Background())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/liked", nil)

	s.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
