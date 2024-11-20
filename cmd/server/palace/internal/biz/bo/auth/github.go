package auth

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IOAuthUser = (*GithubUser)(nil)

// GithubUser Github用户
type GithubUser struct {
	AvatarURL         string `json:"avatar_url"`
	Bio               string `json:"bio"`
	Blog              string `json:"blog"`
	Company           string `json:"company"`
	CreatedAt         string `json:"created_at"`
	Email             string `json:"email"`
	EventsURL         string `json:"events_url"`
	Followers         uint32 `json:"followers"`
	FollowersURL      string `json:"followers_url"`
	Following         uint32 `json:"following"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	GravatarID        string `json:"gravatar_id"`
	Hireable          any    `json:"hireable"`
	HTMLURL           string `json:"html_url"`
	ID                uint32 `json:"id"`
	Location          string `json:"location"`
	Login             string `json:"login"`
	Name              string `json:"name"`
	NodeID            string `json:"node_id"`
	NotificationEmail any    `json:"notification_email"`
	OrganizationsURL  string `json:"organizations_url"`
	PublicGists       uint32 `json:"public_gists"`
	PublicRepos       uint32 `json:"public_repos"`
	ReceivedEventsURL string `json:"received_events_url"`
	ReposURL          string `json:"repos_url"`
	SiteAdmin         bool   `json:"site_admin"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	TwitterUsername   any    `json:"twitter_username"`
	Type              string `json:"type"`
	UpdatedAt         string `json:"updated_at"`
	URL               string `json:"url"`
}

// GetNickname 获取昵称
func (g *GithubUser) GetNickname() string {
	return g.Name
}

// GetOAuthID 获取OAuth ID
func (g *GithubUser) GetOAuthID() uint32 {
	return g.ID
}

// GetEmail 获取邮箱
func (g *GithubUser) GetEmail() string {
	return g.Email
}

// GetRemark 获取备注
func (g *GithubUser) GetRemark() string {
	return g.Bio
}

// GetUsername 获取用户名
func (g *GithubUser) GetUsername() string {
	return g.Login
}

// GetAvatar 获取头像
func (g *GithubUser) GetAvatar() string {
	return g.AvatarURL
}

// GetAPP 获取OAuth应用
func (g *GithubUser) GetAPP() vobj.OAuthAPP {
	return vobj.OAuthAPPGithub
}

// String 将Github用户转换为字符串
func (g *GithubUser) String() string {
	bs, _ := types.Marshal(g)
	return string(bs)
}
