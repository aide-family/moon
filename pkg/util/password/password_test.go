package password_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moon-monitor/moon/pkg/util/password"
)

func TestHashPassword_Success(t *testing.T) {
	ps := "mySecretPassword"
	hashedPassword, err := password.HashPassword(ps)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	ps := "mySecretPassword"
	hashedPassword, _ := password.HashPassword(ps)
	result := password.CheckPassword(ps, hashedPassword)
	assert.True(t, result)
}

func TestCheckPassword_IncorrectPassword(t *testing.T) {
	ps := "mySecretPassword"
	hashedPassword, _ := password.HashPassword(ps)
	result := password.CheckPassword("wrongPassword", hashedPassword)
	assert.False(t, result)
}

func TestCheckPassword_InvalidHashedPassword(t *testing.T) {
	ps := "mySecretPassword"
	invalidHashedPassword := "invalidHash"
	result := password.CheckPassword(ps, invalidHashedPassword)
	assert.False(t, result)
}

func TestGenerateSalt_Success(t *testing.T) {
	size := 16
	salt, err := password.GenerateSalt(size)
	assert.NoError(t, err)
	assert.Len(t, salt, size)
}

func TestGenerateSalt_Randomness(t *testing.T) {
	size := 16
	salt1, _ := password.GenerateSalt(size)
	salt2, _ := password.GenerateSalt(size)
	assert.NotEqual(t, salt1, salt2)
}

func TestGenerateSalt_ZeroSize(t *testing.T) {
	size := 0
	salt, err := password.GenerateSalt(size)
	assert.NoError(t, err)
	assert.Len(t, salt, size)
}

func TestGenerateSalt_NegativeSize(t *testing.T) {
	size := -1
	salt, err := password.GenerateSalt(size)
	assert.Error(t, err)
	assert.Len(t, salt, 0)
}
