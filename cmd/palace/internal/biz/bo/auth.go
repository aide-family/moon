package bo

import (
	"encoding/json"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/middleware"
)

type Captcha struct {
	Id             string `json:"id"`
	B64s           string `json:"b64s"`
	Answer         string `json:"answer"`
	ExpiredSeconds int64  `json:"expired_seconds"`
}

type CaptchaVerify struct {
	Id     string `json:"id"`
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

// FilingInformation 备案信息
type FilingInformation struct {
	URL         string `json:"url"`
	Information string `json:"information"`
}

// VerifyNewPermission 验证新权限请求
type VerifyNewPermission struct {
	SystemRoleID   uint32    `json:"system_role_id"`
	TeamRoleID     uint32    `json:"team_role_id"`
	TeamID         uint32    `json:"team_id"`
	SystemPosition vobj.Role `json:"system_position"`
	TeamPosition   vobj.Role `json:"team_position"`
}

// UserIdentities 用户身份信息
type UserIdentities struct {
	// 系统职位列表
	SystemPositions []vobj.Role `json:"system_positions"`
	// 系统角色列表
	SystemRoles []*SystemRoleItem `json:"system_roles"`
	// 团队列表
	Teams []*TeamItem `json:"teams"`
}

// SystemRoleItem 系统角色项
type SystemRoleItem struct {
	ID     uint32            `json:"id"`
	Name   string            `json:"name"`
	Status vobj.GlobalStatus `json:"status"`
}

// TeamItem 团队项
type TeamItem struct {
	ID        uint32          `json:"id"`
	Name      string          `json:"name"`
	Status    vobj.TeamStatus `json:"status"`
	Positions []vobj.Role     `json:"positions"`
	Roles     []*TeamRoleItem `json:"roles"`
}

// TeamRoleItem 团队角色项
type TeamRoleItem struct {
	ID     uint32            `json:"id"`
	Name   string            `json:"name"`
	Status vobj.GlobalStatus `json:"status"`
}

var _ IOAuthUser = (*FeiShuUser)(nil)

// FeiShuUser 飞书用户
type FeiShuUser struct {
	Name            string `json:"name"`
	EnName          string `json:"en_name"`
	AvatarURL       string `json:"avatar_url"`
	AvatarThumb     string `json:"avatar_thumb"`
	AvatarMiddle    string `json:"avatar_middle"`
	AvatarBig       string `json:"avatar_big"`
	OpenID          string `json:"open_id"`          // 用户在当前应用中的唯一标识
	UnionID         string `json:"union_id"`         // 用户在飞书开放平台中的唯一标识
	Email           string `json:"email"`            // 用户邮箱
	EnterpriseEmail string `json:"enterprise_email"` // 企业邮箱
	ID              string `json:"user_id"`          // 用户ID（旧版字段）
	Mobile          string `json:"mobile"`           // 手机号（带国际码）
	TenantKey       string `json:"tenant_key"`       // 企业唯一标识
	EmployeeNo      string `json:"employee_no"`      // 工号

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
