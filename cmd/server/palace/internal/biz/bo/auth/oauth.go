package auth

import (
	"fmt"

	"github.com/aide-family/moon/pkg/vobj"
)

type IOAuthUser interface {
	fmt.Stringer
	GetOAuthID() uint32
	GetEmail() string
	GetRemark() string
	GetUsername() string
	GetNickname() string
	GetAvatar() string
	GetAPP() vobj.OAuthAPP
}
