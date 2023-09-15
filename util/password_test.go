package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := RandomString(6)

	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	err = ComparePasswords(password, hashedPassword)
	require.NoError(t, err)
}

func TestWrongPassword(t *testing.T) {
	password := RandomString(6)
	hashedPassword, _ := HashPassword(password)

	err := ComparePasswords(RandomString(6), hashedPassword)
	require.Error(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
}

func TestHashPasswordShouldBeDifferent(t *testing.T) {
	password1 := RandomString(6)

	hashedPassword1, _ := HashPassword(password1)
	hashedPassword2, _ := HashPassword(password1)

	require.NotEqual(t, hashedPassword1, hashedPassword2)
}
