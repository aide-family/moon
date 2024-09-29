package auth

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IOAuthUser = (*GithubUser)(nil)

type GithubUser struct {
	AvatarUrl         string `json:"avatar_url"`
	Bio               string `json:"bio"`
	Blog              string `json:"blog"`
	Company           string `json:"company"`
	CreatedAt         string `json:"created_at"`
	Email             string `json:"email"`
	EventsUrl         string `json:"events_url"`
	Followers         uint32 `json:"followers"`
	FollowersUrl      string `json:"followers_url"`
	Following         uint32 `json:"following"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	GravatarId        string `json:"gravatar_id"`
	Hireable          any    `json:"hireable"`
	HtmlUrl           string `json:"html_url"`
	Id                uint32 `json:"id"`
	Location          string `json:"location"`
	Login             string `json:"login"`
	Name              string `json:"name"`
	NodeId            string `json:"node_id"`
	NotificationEmail any    `json:"notification_email"`
	OrganizationsUrl  string `json:"organizations_url"`
	PublicGists       uint32 `json:"public_gists"`
	PublicRepos       uint32 `json:"public_repos"`
	ReceivedEventsUrl string `json:"received_events_url"`
	ReposUrl          string `json:"repos_url"`
	SiteAdmin         bool   `json:"site_admin"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	TwitterUsername   any    `json:"twitter_username"`
	Type              string `json:"type"`
	UpdatedAt         string `json:"updated_at"`
	Url               string `json:"url"`
}

func (g *GithubUser) GetNickname() string {
	return g.Name
}

func (g *GithubUser) GetOAuthID() uint32 {
	return g.Id
}

func (g *GithubUser) GetEmail() string {
	return g.Email
}

func (g *GithubUser) GetRemark() string {
	return g.Bio
}

func (g *GithubUser) GetUsername() string {
	return g.Login
}

func (g *GithubUser) GetAvatar() string {
	return g.AvatarUrl
}

func (g *GithubUser) GetAPP() vobj.OAuthAPP {
	return vobj.OAuthAPPGithub
}

// String implement fmt.Stringer
func (g *GithubUser) String() string {
	bs, _ := types.Marshal(g)
	return string(bs)
}
