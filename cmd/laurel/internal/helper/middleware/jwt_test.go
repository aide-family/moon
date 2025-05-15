package middleware_test

import (
	"testing"
	"time"

	"github.com/moon-monitor/moon/cmd/laurel/internal/helper/middleware"
	"github.com/moon-monitor/moon/pkg/config"
	"google.golang.org/protobuf/types/known/durationpb"
)

func Test_GenJWTToken(t *testing.T) {
	c := &config.JWT{
		SignKey:         "rabbit-sign-key",
		Issuer:          "moon.rabbit",
		Expire:          durationpb.New(time.Hour * 24 * 365),
		AllowOperations: nil,
	}
	claims := middleware.NewJwtClaims(c, "test-token")
	t.Log(claims.GetToken())
}
