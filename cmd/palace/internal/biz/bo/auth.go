package bo

import (
	"encoding/json"

	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/middleware"
)

type Captcha struct {
	ID             string `json:"id"`
	B64s           string `json:"b64s"`
	Answer         string `json:"answer"`
	ExpiredSeconds int64  `json:"expired_seconds"`
}

type CaptchaVerify struct {
	ID     string `json:"id"`
	Answer string `json:"answer"`
	Clear  bool   `json:"clear"`
}

type LoginByPassword struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshToken struct {
	Token  string `json:"token"`
	UserID uint32 `json:"user_id"`
}

type LoginSign struct {
	Base           *middleware.JwtBaseInfo `json:"base"`
	Token          string                  `json:"token"`
	ExpiredSeconds int64                   `json:"expired_seconds"`
}

// FilingInformation represents filing information
type FilingInformation struct {
	URL         string `json:"url"`
	Information string `json:"information"`
}

// VerifyNewPermission represents a new permission verification request
type VerifyNewPermission struct {
	SystemRoleID   uint32        `json:"system_role_id"`
	TeamRoleID     uint32        `json:"team_role_id"`
	TeamID         uint32        `json:"team_id"`
	SystemPosition vobj.Position `json:"system_position"`
	TeamPosition   vobj.Position `json:"team_position"`
}

// UserIdentities represents user identity information
type UserIdentities struct {
	// System position list
	SystemPositions []vobj.Position `json:"system_positions"`
	// System role list
	SystemRoles []*SystemRoleItem `json:"system_roles"`
	// Team list
	Teams []*TeamItem `json:"teams"`
}

// SystemRoleItem represents a system role item
type SystemRoleItem struct {
	ID     uint32            `json:"id"`
	Name   string            `json:"name"`
	Status vobj.GlobalStatus `json:"status"`
}

// TeamItem represents a team item
type TeamItem struct {
	ID        uint32          `json:"id"`
	Name      string          `json:"name"`
	Status    vobj.TeamStatus `json:"status"`
	Positions []vobj.Position `json:"positions"`
	Roles     []*TeamRoleItem `json:"roles"`
}

// TeamRoleItem represents a team role item
type TeamRoleItem struct {
	ID     uint32            `json:"id"`
	Name   string            `json:"name"`
	Status vobj.GlobalStatus `json:"status"`
}

var _ IOAuthUser = (*FeiShuUser)(nil)

// FeiShuUser represents a Feishu user
type FeiShuUser struct {
	Name            string `json:"name"`
	EnName          string `json:"en_name"`
	AvatarURL       string `json:"avatar_url"`
	AvatarThumb     string `json:"avatar_thumb"`
	AvatarMiddle    string `json:"avatar_middle"`
	AvatarBig       string `json:"avatar_big"`
	OpenID          string `json:"open_id"`          // Unique identifier for the user in the current application
	UnionID         string `json:"union_id"`         // Unique identifier for the user in the Feishu open platform
	Email           string `json:"email"`            // User's email
	EnterpriseEmail string `json:"enterprise_email"` // Enterprise email
	ID              string `json:"user_id"`          // User ID (legacy field)
	Mobile          string `json:"mobile"`           // Phone number (with country code)
	TenantKey       string `json:"tenant_key"`       // Enterprise unique identifier
	EmployeeNo      string `json:"employee_no"`      // Employee number

	userID uint32
}

func (f *FeiShuUser) String() string {
	bs, _ := json.Marshal(f)
	return string(bs)
}

func (f *FeiShuUser) GetOpenID() string {
	return f.OpenID
}

func (f *FeiShuUser) GetEmail() string {
	return f.Email
}

func (f *FeiShuUser) GetRemark() string {
	return f.EmployeeNo
}

func (f *FeiShuUser) GetUsername() string {
	return f.Name
}

func (f *FeiShuUser) GetNickname() string {
	return f.EnName
}

func (f *FeiShuUser) GetAvatar() string {
	return f.AvatarURL
}

func (f *FeiShuUser) GetAPP() vobj.OAuthAPP {
	return vobj.OAuthAPPFeiShu
}

func (f *FeiShuUser) WithUserID(userID uint32) IOAuthUser {
	f.userID = userID
	return f
}

func (f *FeiShuUser) GetUserID() uint32 {
	return f.userID
}
