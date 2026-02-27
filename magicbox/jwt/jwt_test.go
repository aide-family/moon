package jwt_test

import (
	"strings"
	"testing"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/hello"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/aide-family/magicbox/jwt"
)

func TestGetJwtToken(t *testing.T) {
	var id snowflake.ID
	t.Log(id.Int64())
	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		t.Fatalf("NewNode failed: %v", err)
	}
	claims := jwt.NewJwtClaims(&config.JWT{
		Secret: "xxx",
		Expire: durationpb.New(24 * 365 * time.Hour),
		Issuer: "magicbox-test",
	}, jwt.BaseInfo{
		UID:      node.Generate(),
		Username: hello.ID(),
	})
	token, err := claims.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatalf("GenerateToken returned empty token")
	}
	t.Log(strings.Join([]string{cnst.HTTPHeaderBearerPrefix, token}, " "))
}
