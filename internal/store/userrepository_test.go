package store_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/kek-flip/scotch-api/internal/model"
	"github.com/kek-flip/scotch-api/internal/store"
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
		PhoneNumber: "+79991234567",
		About:       "example text",
	}
}

func testDb(t *testing.T) *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:admin@localhost:5432/scotch?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	if err = conn.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}

	return conn
}

func TestUserRepository_Create(t *testing.T) {
	u := testUser(t)

	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

	assert.NoError(t, s.User().Create(u))
	assert.NotNil(t, u.ID)

	db.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", u.ID)
}

func TestUserRepository_UniqueLoginConstraint(t *testing.T) {
	u1 := testUser(t)
	u2 := testUser(t)
	u2.PhoneNumber = "+79001234567"

	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

	assert.NoError(t, s.User().Create(u1))
	assert.Error(t, s.User().Create(u2))

	db.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", u1.ID)
}

func TestUserRepository_UniquePhoneNumberConstraint(t *testing.T) {
	u1 := testUser(t)
	u2 := testUser(t)
	u2.Login = "other_valid_login"

	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

	assert.NoError(t, s.User().Create(u1))
	assert.Error(t, s.User().Create(u2))

	db.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", u1.ID)
}

func TestUserRepository_FindById(t *testing.T) {
	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

	original_user := testUser(t)
	err := original_user.EncryptPassword()
	assert.NoError(t, err)

	err = db.QueryRow(
		context.Background(),
		"INSERT INTO users(login, encrypted_password, name, age, gender, city, phone_number, about) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING user_id",
		original_user.Login,
		original_user.EncryptedPassword,
		original_user.Name,
		original_user.Age,
		original_user.Gender,
		original_user.City,
		original_user.PhoneNumber,
		original_user.About,
	).Scan(&original_user.ID)

	assert.NoError(t, err)

	original_user.ClearPassword()

	found_user, err := s.User().FindById(original_user.ID)
	assert.NoError(t, err)
	assert.Equal(t, original_user, found_user)

	db.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", original_user.ID)
}

func TestUserRepository_FindByLogin(t *testing.T) {
	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

	original_user := testUser(t)
	err := original_user.EncryptPassword()
	assert.NoError(t, err)

	err = db.QueryRow(
		context.Background(),
		"INSERT INTO users(login, encrypted_password, name, age, gender, city, phone_number, about) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING user_id",
		original_user.Login,
		original_user.EncryptedPassword,
		original_user.Name,
		original_user.Age,
		original_user.Gender,
		original_user.City,
		original_user.PhoneNumber,
		original_user.About,
	).Scan(&original_user.ID)

	assert.NoError(t, err)

	original_user.ClearPassword()

	found_user, err := s.User().FindByLogin(original_user.Login)
	assert.NoError(t, err)
	assert.Equal(t, original_user, found_user)

	db.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", original_user.ID)
}

func TestUserRepository_DeleteById(t *testing.T) {
	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

	u := testUser(t)
	err := u.EncryptPassword()
	assert.NoError(t, err)

	err = db.QueryRow(
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
	).Scan(&u.ID)

	assert.NoError(t, err)
	assert.NoError(t, s.User().DeleteById(u.ID))
	row := db.QueryRow(context.Background(), "SELECT * FROM users WHERE user_id = $1", u.ID)
	assert.Equal(t, row.Scan(), pgx.ErrNoRows)
}

func TestUserRepository_DeleteByLogin(t *testing.T) {
	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

	u := testUser(t)
	err := u.EncryptPassword()
	assert.NoError(t, err)

	err = db.QueryRow(
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
	).Scan(&u.ID)

	assert.NoError(t, err)
	assert.NoError(t, s.User().DeleteByLogin(u.Login))
	row := db.QueryRow(context.Background(), "SELECT * FROM users WHERE login = $1", u.Login)
	assert.Equal(t, row.Scan(), pgx.ErrNoRows)
}
