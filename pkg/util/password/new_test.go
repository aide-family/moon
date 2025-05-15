package password_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moon-monitor/moon/pkg/util/password"
)

// TestNew_DefaultSalt verifies password object creation with default salt
func TestNew_DefaultSalt(t *testing.T) {
	pwd := "mySecretPassword"
	passwordObj := password.New(pwd)
	assert.NotNil(t, passwordObj)
	assert.NotEmpty(t, passwordObj.Salt())
	assert.Equal(t, pwd, passwordObj.PValue())
}

// TestNew_CustomSalt verifies password object creation with custom salt
func TestNew_CustomSalt(t *testing.T) {
	pwd := "mySecretPassword"
	customSalt := "customSaltValue"
	passwordObj := password.New(pwd, customSalt)
	assert.NotNil(t, passwordObj)
	assert.Equal(t, customSalt, passwordObj.Salt())
	assert.Equal(t, pwd, passwordObj.PValue())
}

// TestEQ_Success verifies correct hash password matching
func TestEQ_Success(t *testing.T) {
	pwd := "mySecretPassword"
	passwordObj := password.New(pwd)
	hashedPassword, _ := passwordObj.EnValue()
	assert.True(t, passwordObj.EQ(hashedPassword))
}

// TestEQ_Failure verifies incorrect hash password non-matching
func TestEQ_Failure(t *testing.T) {
	pwd := "mySecretPassword"
	passwordObj := password.New(pwd)
	assert.False(t, passwordObj.EQ("wrongHashedPassword"))
}

// TestEQ_EmptyHashedPassword verifies empty hash password non-matching
func TestEQ_EmptyHashedPassword(t *testing.T) {
	pwd := "mySecretPassword"
	passwordObj := password.New(pwd)
	assert.False(t, passwordObj.EQ(""))
}

// TestPValue_Success verifies correct original password return
func TestPValue_Success(t *testing.T) {
	pwd := "mySecretPassword"
	passwordObj := password.New(pwd)
	assert.Equal(t, pwd, passwordObj.PValue())
}

// TestEnValue_Success verifies correct encrypted password
func TestEnValue_Success(t *testing.T) {
	pwd := "mySecretPassword"
	passwordObj := password.New(pwd)
	encryptedPassword, err := passwordObj.EnValue()
	assert.NoError(t, err)
	assert.NotEmpty(t, encryptedPassword)
}

// TestSalt_Success verifies correct salt return
func TestSalt_Success(t *testing.T) {
	pwd := "mySecretPassword"
	passwordObj := password.New(pwd)
	assert.NotEmpty(t, passwordObj.Salt())
}
