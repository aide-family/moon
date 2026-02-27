package gitee

import (
	"encoding/json"
	"strconv"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/oauth"
)

type GiteeUser struct {
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

func (g *GiteeUser) Parse() *oauth.OAuth2User {
	return &oauth.OAuth2User{
		OpenID:   strconv.FormatUint(uint64(g.ID), 10),
		Name:     g.Login,
		Nickname: g.Name,
		Email:    g.Email,
		Avatar:   g.AvatarURL,
		App:      config.OAuth2_GITEE,
		Remark:   g.Remark,
		Raw:      g.GetRaw(),
	}
}

// GetRaw implements [auth.User].
func (g *GiteeUser) GetRaw() []byte {
	raw, _ := json.Marshal(g)
	return raw
}
