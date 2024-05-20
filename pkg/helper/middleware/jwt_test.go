package middleware

import (
	"testing"
)

func TestNewJwtClaims(t *testing.T) {
	token, err := NewJwtClaims(&JwtBaseInfo{
		User:     1,
		Role:     1,
		Team:     1,
		TeamRole: 1,
	}).GetToken()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(token)
}
