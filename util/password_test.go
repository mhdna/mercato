package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := RandomString(12)

	hash, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	err = CheckPassword(password, hash)
	require.NoError(t, err)

	wrongPassword := RandomString(12)
	err = CheckPassword(wrongPassword, hash)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := RandomString(12)
	hash1, err := HashPassword(password)
	require.NoError(t, err)

	hash2, err := HashPassword(password)
	require.NoError(t, err)

	require.NotEqual(t, hash1, hash2)
}
