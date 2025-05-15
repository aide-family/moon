package password_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moon-monitor/moon/pkg/util/password"
)

// TestGenerateRandomPassword_Success verifies correct password length generation
func TestGenerateRandomPassword_Success(t *testing.T) {
	length := 10
	p := password.GenerateRandomPassword(length)
	assert.Len(t, p, length)
}

// TestGenerateRandomPassword_Randomness verifies password randomness
func TestGenerateRandomPassword_Randomness(t *testing.T) {
	length := 10
	password1 := password.GenerateRandomPassword(length)
	password2 := password.GenerateRandomPassword(length)
	assert.NotEqual(t, password1, password2)
}

// TestGenerateRandomPassword_ZeroLength verifies handling of length = 0
func TestGenerateRandomPassword_ZeroLength(t *testing.T) {
	length := 0
	p := password.GenerateRandomPassword(length)
	assert.Len(t, p, length)
}

// TestGenerateRandomPassword_NegativeLength verifies handling of length < 0
func TestGenerateRandomPassword_NegativeLength(t *testing.T) {
	length := -1
	p := password.GenerateRandomPassword(length)
	assert.Len(t, p, 0)
}
