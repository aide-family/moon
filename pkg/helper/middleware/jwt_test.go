package middleware

import (
	"testing"
)

func TestNewJwtClaims(t *testing.T) {
	token, err := NewJwtClaims(&JwtBaseInfo{
		UserID: 1,
		TeamID: 5,
	}).GetToken()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(token)
}
