package middleware

import (
	"testing"
)

func TestNewJwtClaims(t *testing.T) {
	token, err := NewJwtClaims(&JwtBaseInfo{
		UserID:   2,
		TeamID:   1,
		MemberID: 1,
	}).GetToken()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(token)
}
