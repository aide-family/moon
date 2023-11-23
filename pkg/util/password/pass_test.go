package password

import (
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	salt := GenerateSalt()
	newPass, err := GeneratePassword("123456", salt)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("newPass: ", newPass)
	password, err := DecryptPassword(newPass, salt)
	if err != nil {
		return
	}
	t.Log("d: ", password, " ", salt)

	t.Log("ValidatePassword: ", ValidatePassword("123456", newPass, salt))
	t.Log("ValidatePassword: ", ValidatePassword("12345", newPass, salt))
}

func TestGenerateSalt(t *testing.T) {
	for i := 0; i < 10000; i++ {
		GenerateSalt()
	}
}
