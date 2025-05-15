package bo

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/helper/middleware"
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
