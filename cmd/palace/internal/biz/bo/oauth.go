package bo

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/validate"
)

type IOAuthUser interface {
	fmt.Stringer
	GetOpenID() string
	GetEmail() string
	GetRemark() string
	GetUsername() string
	GetNickname() string
	GetAvatar() string
	GetAPP() vobj.OAuthAPP
	WithUserID(userID uint32) IOAuthUser
	GetUserID() uint32
}

type LoginWithEmailParams struct {
	Code string
	do.User
	SendEmailFun SendEmailFun
}

type VerifyEmailParams struct {
	Email        string
	SendEmailFun SendEmailFun
}

type VerifyEmailCodeParams struct {
	Email string
	Code  string
}

type OAuthLoginParams struct {
	APP          vobj.OAuthAPP
	From         vobj.OAuthFrom
	Code         string
	Email        string
	OpenID       string
	Token        string
	SendEmailFun SendEmailFun
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
		return nil, merr.ErrorInternalServer("unknown oauth app")
	}
}

var _ IOAuthUser = (*GithubUser)(nil)

// GithubUser represents a GitHub user
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

	userID uint32
}

func (g *GithubUser) WithUserID(userID uint32) IOAuthUser {
	g.userID = userID
	return g
}

func (g *GithubUser) GetUserID() uint32 {
	return g.userID
}

// GetNickname gets the nickname
func (g *GithubUser) GetNickname() string {
	return g.Name
}

// GetOpenID gets the OpenID
func (g *GithubUser) GetOpenID() string {
	return strconv.FormatUint(uint64(g.ID), 10)
}

// GetEmail gets the email
func (g *GithubUser) GetEmail() string {
	if err := validate.CheckEmail(g.Email); err != nil {
		return ""
	}
	return g.Email
}

// GetRemark gets the remark
func (g *GithubUser) GetRemark() string {
	return g.Bio
}

// GetUsername gets the username
func (g *GithubUser) GetUsername() string {
	return g.Login
}

// GetAvatar gets the avatar
func (g *GithubUser) GetAvatar() string {
	return g.AvatarURL
}

// GetAPP gets the OAuth application
func (g *GithubUser) GetAPP() vobj.OAuthAPP {
	return vobj.OAuthAPPGithub
}

// String converts the GitHub user to a string
func (g *GithubUser) String() string {
	bs, _ := json.Marshal(g)
	return string(bs)
}

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

	userID uint32
}

func (g *GiteeUser) WithUserID(userID uint32) IOAuthUser {
	g.userID = userID
	return g
}

func (g *GiteeUser) GetUserID() uint32 {
	return g.userID
}

// GetOpenID gets the OpenID
func (g *GiteeUser) GetOpenID() string {
	return strconv.FormatUint(uint64(g.ID), 10)
}

// GetEmail gets the email
func (g *GiteeUser) GetEmail() string {
	if err := validate.CheckEmail(g.Email); err != nil {
		return ""
	}
	return g.Email
}

// GetRemark gets the remark
func (g *GiteeUser) GetRemark() string {
	return g.Remark
}

// GetUsername gets the username
func (g *GiteeUser) GetUsername() string {
	return g.Login
}

// GetNickname gets the nickname
func (g *GiteeUser) GetNickname() string {
	return g.Name
}

// GetAvatar gets the avatar
func (g *GiteeUser) GetAvatar() string {
	return g.AvatarURL
}

// GetAPP gets the OAuth application
func (g *GiteeUser) GetAPP() vobj.OAuthAPP {
	return vobj.OAuthAPPGitee
}

// String converts the Gitee user to a string
func (g *GiteeUser) String() string {
	bs, _ := json.Marshal(g)
	return string(bs)
}
