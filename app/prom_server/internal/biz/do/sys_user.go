package do

import (
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/password"
)

const TableNameSystemUser = "sys_users"

// SysUser 用户表
type SysUser struct {
	BaseModel
	Username   string           `gorm:"column:username;type:varchar(64);not null;uniqueIndex:idx__username,priority:1;comment:用户名"`
	Nickname   string           `gorm:"column:nickname;type:varchar(64);not null;comment:昵称"`
	Password   string           `gorm:"column:password;type:varchar(255);not null;comment:密码"`
	Email      string           `gorm:"column:email;type:varchar(64);not null;uniqueIndex:idx__email,priority:1;comment:邮箱"`
	Phone      string           `gorm:"column:phone;type:varchar(64);not null;uniqueIndex:idx__phone,priority:1;comment:手机号"`
	Status     vo.Status        `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark     string           `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Avatar     string           `gorm:"column:avatar;type:varchar(255);not null;comment:头像"`
	Salt       string           `gorm:"column:salt;type:varchar(16);not null;comment:盐"`
	Gender     vo.Gender        `gorm:"column:gender;type:tinyint;not null;default:0;comment:性别"`
	Roles      []*SysRole       `gorm:"many2many:sys_user_roles;comment:用户角色"`
	AlarmPages []*PromAlarmPage `gorm:"many2many:sys_user_alarm_pages;comment:用户页面"`
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
func (u *SysUser) GetAlarmPages() []*PromAlarmPage {
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
			Email:     "admin@prometheus.com",
			Phone:     "13800000000",
			Status:    vo.StatusEnabled,
			Remark:    "超级管理员账号",
			Avatar:    "https://img0.baidu.com/it/u=640865303,1189373079&fm=253&fmt=auto&app=138&f=JPEG?w=300&h=300",
			Salt:      "",
			Gender:    vo.GenderMale,
			Roles: []*SysRole{
				{
					BaseModel: BaseModel{ID: 1},
					Remark:    "超级管理员角色",
					Name:      "超级管理员角色",
					Status:    vo.StatusEnabled,
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
