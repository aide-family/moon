package auth

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// GiteeUser Gitee用户
	GiteeUser struct {
		AvatarURL         string `json:"avatar_url"`
		Bio               string `json:"bio"`
		Blog              string `json:"blog"`
		CreatedAt         string `json:"created_at"`
		Email             string `json:"email"`
		EventsURL         string `json:"events_url"`
		Followers         uint32 `json:"followers"`
		FollowersURL      string `json:"followers_url"`
		Following         uint32 `json:"following"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		HTMLURL           string `json:"html_url"`
		ID                uint32 `json:"id"`
		Login             string `json:"login"`
		Name              string `json:"name"`
		OrganizationsURL  string `json:"organizations_url"`
		PublicGists       uint32 `json:"public_gists"`
		PublicRepos       uint32 `json:"public_repos"`
		ReceivedEventsURL string `json:"received_events_url"`
		Remark            string `json:"remark"`
		ReposURL          string `json:"repos_url"`
		Stared            uint32 `json:"stared"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		Type              string `json:"type"`
		UpdatedAt         string `json:"updated_at"`
		URL               string `json:"url"`
		Watched           uint32 `json:"watched"`
		Weibo             string `json:"weibo"`
	}
)

// GetOAuthID 获取OAuth ID
func (g *GiteeUser) GetOAuthID() uint32 {
	return g.ID
}

// GetEmail 获取邮箱
func (g *GiteeUser) GetEmail() string {
	return g.Email
}

// GetRemark 获取备注
func (g *GiteeUser) GetRemark() string {
	return g.Remark
}

// GetUsername 获取用户名
func (g *GiteeUser) GetUsername() string {
	return g.Login
}

// GetNickname 获取昵称
func (g *GiteeUser) GetNickname() string {
	return g.Name
}

// GetAvatar 获取头像
func (g *GiteeUser) GetAvatar() string {
	return g.AvatarURL
}

// GetAPP 获取OAuth应用
func (g *GiteeUser) GetAPP() vobj.OAuthAPP {
	return vobj.OAuthAPPGitee
}

// String 将Gitee用户转换为字符串
func (g *GiteeUser) String() string {
	bs, _ := types.Marshal(g)
	return string(bs)
}
