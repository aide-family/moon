package middleware

import (
	"testing"

	"github.com/aide-family/moon/pkg/vobj"
)

func TestNewJwtClaims(t *testing.T) {
	token, err := NewJwtClaims(&JwtBaseInfo{
		User:     1,
		Role:     1,
		Team:     17,
		TeamRole: vobj.RoleSuperAdmin,
	}).GetToken()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(token)
}
