package auth

import (
	"encoding/json"
)

type GithubLoginResponse struct {
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

// String implement fmt.Stringer
func (g *GithubLoginResponse) String() string {
	bs, _ := json.Marshal(g)
	return string(bs)
}
