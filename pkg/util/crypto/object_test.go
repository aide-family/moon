package crypto_test

import (
	"strings"
	"testing"

	"github.com/moon-monitor/moon/pkg/util/crypto"
)

type User struct {
	Name     string
	NickName string
}

func TestObject_Scan_Success(t *testing.T) {
	user := &User{Name: "moon", NickName: "moon"}
	object := crypto.Object[*User]{Data: user}
	val, err := object.Value()
	if err != nil {
		t.Fatal(err)
	}

	var origin string
	switch val.(type) {
	case string:
		origin = val.(string)
	case []byte:
		origin = string(val.([]byte))
	default:
		t.Fatal("Unexpected type")
	}

	t.Logf("val: %v", origin)
	var userObject crypto.Object[*User]
	if err := userObject.Scan(val); err != nil {
		t.Fatal(err)
	}
	t.Logf("user: %v", userObject.Get())
	gotUser := userObject.Get()
	if !strings.EqualFold(gotUser.Name, user.Name) {
		t.Errorf("Expected '%v', got '%v'", gotUser.Name, user.Name)
	}
	if !strings.EqualFold(gotUser.NickName, user.NickName) {
		t.Errorf("Expected '%v', got '%v'", gotUser.NickName, user.NickName)
	}
}
