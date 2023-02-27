package store_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/kek-flip/scotch-api/internal/model"
	"github.com/kek-flip/scotch-api/internal/store"
)

func testDb(t *testing.T) *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), "postgres://api:api_password@localhost:5432/scotch?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	if err = conn.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}

	return conn
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

func testLike(t *testing.T) *model.Like {
	t.Helper()

	u1 := testUser(t)
	u2 := testUser(t)
	u2.Login += "1"
	u2.PhoneNumber = "+79999999991"

	db := testDb(t)
	defer db.Close(context.Background())

	s := store.NewStore(db)

	if err := s.User().Create(u1); err != nil {
		t.Fatal(err)
	}
	if err := s.User().Create(u2); err != nil {
		t.Fatal(err)
	}

	return &model.Like{
		UserID:    u1.ID,
		LikedUser: u2.ID,
	}
}

func deleteUsers(t *testing.T, uID1, uID2 int) {
	db := testDb(t)

	if _, err := db.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", uID1); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", uID2); err != nil {
		t.Fatal(err)
	}
}
