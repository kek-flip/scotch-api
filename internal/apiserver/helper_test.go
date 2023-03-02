package apiserver

import (
	"context"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/kek-flip/scotch-api/internal/model"
	"github.com/kek-flip/scotch-api/internal/store"
)

func testServer(t *testing.T) (*server, *pgx.Conn) {
	t.Helper()

	conn, err := pgx.Connect(context.Background(), "postgres://api:api_password@localhost:5432/scotch?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	store := store.NewStore(conn)
	sessionStore := sessions.NewCookieStore([]byte("key"))

	return newServer(store, sessionStore), conn
}

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
