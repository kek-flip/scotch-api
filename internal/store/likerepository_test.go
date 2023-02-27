package store_test

import (
	"context"
	"testing"

	"github.com/kek-flip/scotch-api/internal/model"
	"github.com/kek-flip/scotch-api/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestLikeRepository_Create(t *testing.T) {
	l := testLike(t)

	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

	assert.NoError(t, s.Like().Create(l))

	db.Exec(context.Background(), "DELETE FROM likes WHERE like_id = $1", l.ID)
	deleteUsers(t, l.UserID, l.LikedUser)
}

func TestLikeRepository_DiffUsersConstraint(t *testing.T) {
	u := testUser(t)
	err := u.EncryptPassword()
	assert.NoError(t, err)

	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

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

	l := &model.Like{
		UserID:    u.ID,
		LikedUser: u.ID,
	}

	assert.Error(t, s.Like().Create(l))

	db.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", u.ID)
}

func TestLikeRepository_MustBeUniquePairConstraint(t *testing.T) {
	l1 := testLike(t)
	l2 := &model.Like{
		UserID:    l1.UserID,
		LikedUser: l1.LikedUser,
	}

	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

	err := db.QueryRow(
		context.Background(),
		"INSERT INTO likes(user_id, liked_user) VALUES ($1, $2) RETURNING like_id",
		l1.UserID,
		l1.LikedUser,
	).Scan(&l1.ID)
	assert.NoError(t, err)

	assert.Error(t, s.Like().Create(l2))

	db.Exec(context.Background(), "DELETE FROM likes WHERE like_id = $1", l1.ID)
	deleteUsers(t, l1.UserID, l1.LikedUser)
}

func TestLikeRepository_FindByID(t *testing.T) {
	original_like := testLike(t)

	db := testDb(t)
	defer db.Close(context.Background())
	s := store.NewStore(db)

	err := db.QueryRow(
		context.Background(),
		"INSERT INTO likes(user_id, liked_user) VALUES ($1, $2) RETURNING like_id",
		original_like.UserID,
		original_like.LikedUser,
	).Scan(&original_like.ID)

	assert.NoError(t, err)

	like, err := s.Like().FindById(original_like.ID)
	assert.NoError(t, err)
	assert.Equal(t, original_like, like)

	db.Exec(context.Background(), "DELETE FROM likes WHERE like_id = $1", original_like.ID)
	deleteUsers(t, original_like.UserID, original_like.LikedUser)
}
