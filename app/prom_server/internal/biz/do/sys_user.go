package do

import (
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"

	"prometheus-manager/app/prom_server/internal/biz/vobj"
	"prometheus-manager/pkg/util/password"
)

const TableNameSystemUser = "sys_users"

const (
	SysUserFieldUsername          = "username"
	SysUserFieldNickname          = "nickname"
	SysUserFieldPhone             = "phone"
	SysUserFieldEmail             = "email"
	SysUserFieldGender            = "gender"
	SysUserFieldStatus            = "status"
	SysUserFieldRemark            = "remark"
	SysUserFieldAvatar            = "avatar"
	SysUserPreloadFieldRoles      = "Roles"
	SysUserPreloadFieldAlarmPages = "AlarmPages"
)

// SysUserLike 模糊查询
func SysUserLike(keyword string) basescopes.ScopeMethod {
	return basescopes.WhereLikePrefixKeyword(
		keyword,
		SysUserFieldUsername,
		SysUserFieldEmail,
		SysUserFieldPhone,
		SysUserFieldNickname,
		SysUserFieldRemark,
	)
}

// SysUserUserEqName 等于name
func SysUserUserEqName(name string) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(SysUserFieldUsername, name)
}

// SysUserEqEmail 等于email
func SysUserEqEmail(email string) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(SysUserFieldEmail, email)
}

// SysUserEqPhone 等于phone
func SysUserEqPhone(phone string) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(SysUserFieldPhone, phone)
}

// SysUserPreloadRoles 预加载角色
func SysUserPreloadRoles(roleIds ...uint32) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(roleIds) > 0 {
			return db.Preload(SysUserPreloadFieldRoles, basescopes.WhereInColumn(basescopes.BaseFieldID, roleIds...))
		}
		return db.Preload(SysUserPreloadFieldRoles)
	}
}

// SysUserPreloadAlarmPages 预加载报警页面
func SysUserPreloadAlarmPages(alarmPageIds ...uint32) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(alarmPageIds) > 0 {
			return db.Preload(SysUserPreloadFieldAlarmPages, basescopes.WhereInColumn(basescopes.BaseFieldID, alarmPageIds...))
		}
		return db.Preload(SysUserPreloadFieldAlarmPages)
	}
}

// SysUser 用户表
type SysUser struct {
	BaseModel
	Username   string      `gorm:"column:username;type:varchar(64);not null;uniqueIndex:idx__su__username,priority:1;comment:用户名"`
	Nickname   string      `gorm:"column:nickname;type:varchar(64);not null;comment:昵称"`
	Password   string      `gorm:"column:password;type:varchar(255);not null;comment:密码"`
	Email      string      `gorm:"column:email;type:varchar(64);not null;uniqueIndex:idx__su__email,priority:1;comment:邮箱"`
	Phone      string      `gorm:"column:phone;type:varchar(64);not null;uniqueIndex:idx__su__phone,priority:1;comment:手机号"`
	Status     vobj.Status `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark     string      `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Avatar     string      `gorm:"column:avatar;type:varchar(255);not null;comment:头像"`
	Salt       string      `gorm:"column:salt;type:varchar(16);not null;comment:盐"`
	Gender     vobj.Gender `gorm:"column:gender;type:tinyint;not null;default:0;comment:性别"`
	Roles      []*SysRole  `gorm:"many2many:sys_user_roles;comment:用户角色"`
	AlarmPages []*SysDict  `gorm:"many2many:sys_user_alarm_pages;comment:用户页面"`
}

// TableName 表名
func (*SysUser) TableName() string {
	return TableNameSystemUser
}

// GetRoles 获取角色列表
func (u *SysUser) GetRoles() []*SysRole {
	if u == nil {
		return nil
	}
	return u.Roles
}

// GetAlarmPages 获取页面列表
func (u *SysUser) GetAlarmPages() []*SysDict {
	if u == nil {
		return nil
	}
	return u.AlarmPages
}

var once sync.Once

// InitSuperUser 初始化超级管理员账号
func InitSuperUser(db *gorm.DB) (err error) {
	once.Do(func() {
		adminUser := &SysUser{
			BaseModel: BaseModel{ID: 1},
			Username:  "admin",
			Nickname:  "超级管理员",
			Password:  "123456",
			Email:     "1058165620@qq.com",
			Phone:     "13800000000",
			Status:    vobj.StatusEnabled,
			Remark:    "超级管理员账号",
			Avatar:    "https://api.dicebear.com/7.x/miniavs/svg?seed=8",
			Salt:      "",
			Gender:    vobj.GenderMale,
			Roles: []*SysRole{
				{
					BaseModel: BaseModel{ID: 1},
					Remark:    "超级管理员角色",
					Name:      "超级管理员角色",
					Status:    vobj.StatusEnabled,
				},
			},
		}
		adminUser.Salt = password.GenerateSalt()
		adminUser.Password, err = password.GeneratePassword(adminUser.Password, adminUser.Salt)
		if err != nil {
			return
		}
		err = db.Model(&SysUser{}).Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(adminUser).Error
	})
	return err
}
