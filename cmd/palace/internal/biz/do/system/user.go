package system

import (
	"encoding/json"
	"strconv"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/password"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.User = (*User)(nil)

const tableNameUser = "sys_users"

type User struct {
	do.BaseModel
	Username string          `gorm:"column:username;type:varchar(64);not null;index:idx__sys_user__username,priority:1;comment:用户名" json:"username"`
	Nickname string          `gorm:"column:nickname;type:varchar(64);not null;comment:昵称" json:"nickname"`
	Password string          `gorm:"column:password;type:varchar(255);not null;comment:密码" json:"-"`
	Email    string          `gorm:"column:email;type:varchar(64);not null;comment:邮箱" json:"email"`
	Phone    string          `gorm:"column:phone;type:varchar(64);not null;comment:手机号" json:"phone"`
	Remark   string          `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Avatar   string          `gorm:"column:avatar;type:varchar(255);not null;comment:头像" json:"avatar"`
	Salt     string          `gorm:"column:salt;type:varchar(128);not null;comment:盐" json:"-"`
	Gender   vobj.Gender     `gorm:"column:gender;type:tinyint(2);not null;comment:性别" json:"gender"`
	Position vobj.Role       `gorm:"column:position;type:tinyint(2);not null;comment:系统默认角色类型" json:"position"`
	Status   vobj.UserStatus `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	Roles    []*Role         `gorm:"many2many:sys_user_roles" json:"roles"`
	Teams    []*Team         `gorm:"many2many:sys_user_teams" json:"teams"`
}

func (u *User) MarshalBinary() (data []byte, err error) {
	if u == nil {
		return nil, nil
	}
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	if u == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, u)
}

func (u *User) UniqueKey() string {
	return strconv.Itoa(int(u.ID))
}

func (u *User) GetUsername() string {
	if u == nil {
		return ""
	}
	return u.Username
}

func (u *User) GetNickname() string {
	if u == nil {
		return ""
	}
	return u.Nickname
}

func (u *User) GetEmail() string {
	if u == nil {
		return ""
	}
	return u.Email
}

func (u *User) GetPhone() string {
	if u == nil {
		return ""
	}
	return u.Phone
}

func (u *User) GetRemark() string {
	if u == nil {
		return ""
	}
	return u.Remark
}

func (u *User) GetPassword() string {
	if u == nil {
		return ""
	}
	return u.Password
}

func (u *User) GetSalt() string {
	if u == nil {
		return ""
	}
	return u.Salt
}

func (u *User) GetGender() vobj.Gender {
	if u == nil {
		return vobj.GenderUnknown
	}
	return u.Gender
}

func (u *User) GetAvatar() string {
	if u == nil {
		return ""
	}
	return u.Avatar
}

func (u *User) GetStatus() vobj.UserStatus {
	if u == nil {
		return vobj.UserStatusUnknown
	}
	return u.Status
}

func (u *User) GetPosition() vobj.Role {
	if u == nil {
		return vobj.RoleUnknown
	}
	return u.Position
}

func (u *User) GetRoles() []do.Role {
	if u == nil {
		return nil
	}
	return slices.Map(u.Roles, func(r *Role) do.Role { return r })
}

func (u *User) GetTeams() []do.Team {
	if u == nil {
		return nil
	}
	return slices.Map(u.Teams, func(t *Team) do.Team { return t })
}

func (u *User) TableName() string {
	return tableNameUser
}

// ValidatePassword validate password
func (u *User) ValidatePassword(p string) bool {
	validate := password.New(p, u.Salt)
	return validate.EQ(u.Password)
}

func (u *User) SetEmail(email string) {
	if u == nil {
		return
	}
	u.Email = email
}
