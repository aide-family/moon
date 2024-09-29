package auth

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	GiteeUser struct {
		AvatarUrl         string `json:"avatar_url"`
		Bio               string `json:"bio"`
		Blog              string `json:"blog"`
		CreatedAt         string `json:"created_at"`
		Email             string `json:"email"`
		EventsUrl         string `json:"events_url"`
		Followers         uint32 `json:"followers"`
		FollowersUrl      string `json:"followers_url"`
		Following         uint32 `json:"following"`
		FollowingUrl      string `json:"following_url"`
		GistsUrl          string `json:"gists_url"`
		HtmlUrl           string `json:"html_url"`
		Id                uint32 `json:"id"`
		Login             string `json:"login"`
		Name              string `json:"name"`
		OrganizationsUrl  string `json:"organizations_url"`
		PublicGists       uint32 `json:"public_gists"`
		PublicRepos       uint32 `json:"public_repos"`
		ReceivedEventsUrl string `json:"received_events_url"`
		Remark            string `json:"remark"`
		ReposUrl          string `json:"repos_url"`
		Stared            uint32 `json:"stared"`
		StarredUrl        string `json:"starred_url"`
		SubscriptionsUrl  string `json:"subscriptions_url"`
		Type              string `json:"type"`
		UpdatedAt         string `json:"updated_at"`
		Url               string `json:"url"`
		Watched           uint32 `json:"watched"`
		Weibo             string `json:"weibo"`
	}
)

func (g *GiteeUser) GetOAuthID() uint32 {
	return g.Id
}

func (g *GiteeUser) GetEmail() string {
	return g.Email
}

func (g *GiteeUser) GetRemark() string {
	return g.Remark
}

func (g *GiteeUser) GetUsername() string {
	return g.Login
}

func (g *GiteeUser) GetNickname() string {
	return g.Name
}

func (g *GiteeUser) GetAvatar() string {
	return g.AvatarUrl
}

func (g *GiteeUser) GetAPP() vobj.OAuthAPP {
	return vobj.OAuthAPPGitee
}

// String implements fmt.Stringer interface
func (g *GiteeUser) String() string {
	bs, _ := types.Marshal(g)
	return string(bs)
}
