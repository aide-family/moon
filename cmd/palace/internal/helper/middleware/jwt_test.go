package middleware_test

import (
	"testing"
	"time"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/helper/middleware"
	"github.com/moon-monitor/moon/pkg/config"
	"google.golang.org/protobuf/types/known/durationpb"
)

func Test_GenJWTToken(t *testing.T) {
	c := &config.JWT{
		SignKey:         "moon-jwt-sign-key",
		Issuer:          "moon.palace",
		Expire:          durationpb.New(time.Hour * 24 * 365),
		AllowOperations: nil,
	}
	claims := middleware.NewJwtClaims(c, &middleware.JwtBaseInfo{
		UserID:   1,
		Username: "admin",
		Nickname: "管理员",
		Avatar:   "",
		Gender:   vobj.GenderMale,
	})
	t.Log(claims.GetToken())
}
