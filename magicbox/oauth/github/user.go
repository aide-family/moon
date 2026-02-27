package github

import (
	"encoding/json"
	"strconv"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/oauth"
)

type User struct {
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

func (u *User) Parse() *oauth.OAuth2User {
	return &oauth.OAuth2User{
		OpenID:   strconv.FormatUint(uint64(u.ID), 10),
		Name:     u.Login,
		Nickname: u.Name,
		Email:    u.Email,
		Avatar:   u.AvatarURL,
		App:      config.OAuth2_GITHUB,
		Remark:   u.Bio,
		Raw:      u.GetRaw(),
	}
}

// GetRaw implements [auth.User].
func (u *User) GetRaw() []byte {
	raw, _ := json.Marshal(u)
	return raw
}
