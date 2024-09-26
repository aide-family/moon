package auth

import (
	"encoding/json"
	"fmt"

	"github.com/aide-family/moon/api/merr"
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

type OauthLoginParams struct {
	Code    string `json:"code"`
	Email   string `json:"email"`
	OAuthID uint32 `json:"OAuthID"`
}

func NewOAuthRowData(app vobj.OAuthAPP, row string) (IOAuthUser, error) {
	switch app {
	case vobj.OAuthAPPGithub:
		var githubUser GithubUser
		err := json.Unmarshal([]byte(row), &githubUser)
		return &githubUser, err
	case vobj.OAuthAPPGitee:
		var giteeUser GiteeUser
		err := json.Unmarshal([]byte(row), &giteeUser)
		return &giteeUser, err
	default:
		return nil, merr.ErrorI18nNotificationSystemError(nil)
	}
}
