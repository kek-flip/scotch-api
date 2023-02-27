package model_test

import (
	"testing"

	"github.com/kek-flip/scotch-api/internal/model"
	"github.com/stretchr/testify/assert"
)

func testLike(t *testing.T) *model.Like {
	t.Helper()

	return &model.Like{
		UserID:    1,
		LikedUser: 2,
	}
}

func TestLike_Validation(t *testing.T) {
	testCases := []struct {
		name    string
		l       func() *model.Like
		isValid bool
	}{
		{
			name: "Valid like",
			l: func() *model.Like {
				l := testLike(t)
				return l
			},
			isValid: true,
		},
		{
			name: "Same users",
			l: func() *model.Like {
				l := testLike(t)
				l.UserID = l.LikedUser
				return l
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				assert.NoError(t, tc.l().Validate())
			} else {
				assert.Error(t, tc.l().Validate())
			}
		})
	}
}
