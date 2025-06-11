package bo

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/util/validate"
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

type OAuthLoginParams struct {
	APP          vobj.OAuthAPP `json:"app"`
	Code         string        `json:"code"`
	Email        string        `json:"email"`
	OpenID       string        `json:"openID"`
	Token        string        `json:"token"`
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
		return nil, merr.ErrorInternalServerError("unknown oauth app")
	}
}

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

	userID uint32
}

func (g *GithubUser) WithUserID(userID uint32) IOAuthUser {
	g.userID = userID
	return g
}

func (g *GithubUser) GetUserID() uint32 {
	return g.userID
}

// GetNickname 获取昵称
func (g *GithubUser) GetNickname() string {
	return g.Name
}

// GetOpenID 获取OpenID
func (g *GithubUser) GetOpenID() string {
	return strconv.FormatUint(uint64(g.ID), 10)
}

// GetEmail 获取邮箱
func (g *GithubUser) GetEmail() string {
	if err := validate.CheckEmail(g.Email); err != nil {
		return ""
	}
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

// GetOpenID 获取OpenID
func (g *GiteeUser) GetOpenID() string {
	return strconv.FormatUint(uint64(g.ID), 10)
}

// GetEmail 获取邮箱
func (g *GiteeUser) GetEmail() string {
	if err := validate.CheckEmail(g.Email); err != nil {
		return ""
	}
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
	bs, _ := json.Marshal(g)
	return string(bs)
}
