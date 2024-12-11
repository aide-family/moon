package middleware

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

type JwtConfigMock struct {
	signKey string
	expire  time.Duration
	issuer  string
}

func (j *JwtConfigMock) GetSignKey() string {
	return j.signKey
}

func (j *JwtConfigMock) GetExpire() *durationpb.Duration {
	return durationpb.New(j.expire)
}

func (j *JwtConfigMock) GetIssuer() string {
	return j.issuer
}

func TestNewJwtClaims(t *testing.T) {
	SetJwtConfig(&JwtConfigMock{
		signKey: "moon-sign_key",
		expire:  time.Second * 3600,
		issuer:  "moon-palace",
	})
	token, err := NewJwtClaims(&JwtBaseInfo{
		UserID:   1,
		TeamID:   1,
		MemberID: 1,
	}).GetToken()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(token)
}
