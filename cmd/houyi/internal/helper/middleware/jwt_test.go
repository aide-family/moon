package middleware_test

import (
	"testing"
	"time"

	"github.com/aide-family/moon/cmd/houyi/internal/helper/middleware"
	"github.com/aide-family/moon/pkg/config"
	"google.golang.org/protobuf/types/known/durationpb"
)

func Test_GenJWTToken(t *testing.T) {
	c := &config.JWT{
		SignKey:         "houyi-sign-key",
		Issuer:          "moon.houyi",
		Expire:          durationpb.New(time.Hour * 24 * 365),
		AllowOperations: nil,
	}
	claims := middleware.NewJwtClaims(c, "test-token")
	t.Log(claims.GetToken())
}
