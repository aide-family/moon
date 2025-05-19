package do

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
)

type GetUserFun func(id uint32) User

type GetTeamFun func(id uint32) Team

type GetTeamMemberFun func(id uint32) TeamMember

type HasTableFun func(teamId uint32, tableName string) bool

type CacheTableFlag func(teamId uint32, tableName string) error

var registerGetUserFuncOnce sync.Once
var getUser GetUserFun

func RegisterGetUserFunc(getUserFunc GetUserFun) {
	registerGetUserFuncOnce.Do(func() {
		getUser = getUserFunc
	})
}

func GetUser(id uint32) User {
	return getUser(id)
}

var registerGetTeamFuncOnce sync.Once
var getTeam GetTeamFun

func RegisterGetTeamFunc(getTeamFunc GetTeamFun) {
	registerGetTeamFuncOnce.Do(func() {
		getTeam = getTeamFunc
	})
}

func GetTeam(id uint32) Team {
	return getTeam(id)
}

var registerGetTeamMemberFuncOnce sync.Once
var getTeamMember GetTeamMemberFun

func RegisterGetTeamMemberFunc(getTeamMemberFunc GetTeamMemberFun) {
	registerGetTeamMemberFuncOnce.Do(func() {
		getTeamMember = getTeamMemberFunc
	})
}

func GetTeamMember(id uint32) TeamMember {
	return getTeamMember(id)
}

var registerHasTableFuncOnce sync.Once
var hasTable HasTableFun = nil

func RegisterHasTableFunc(hasTableFunc HasTableFun) {
	registerHasTableFuncOnce.Do(func() {
		hasTable = hasTableFunc
	})
}

var registerCacheTableFlagOnce sync.Once
var cacheTableFlag CacheTableFlag = nil

func RegisterCacheTableFlag(cacheTableFlagFunc CacheTableFlag) {
	registerCacheTableFlagOnce.Do(func() {
		cacheTableFlag = cacheTableFlagFunc
	})
}

type ORMModel interface {
	sql.Scanner
	driver.Valuer
}

type Base interface {
	GetID() uint32
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() soft_delete.DeletedAt
	GetContext() context.Context
	WithContext(context.Context)
}

type Creator interface {
	Base
	GetCreatorID() uint32
	GetCreator() User
	GetCreatorMember() TeamMember
}

type TeamBase interface {
	Creator
	GetTeamID() uint32
	GetTeam() Team
}

var _ Base = (*BaseModel)(nil)

// BaseModel gorm base model
type BaseModel struct {
	ctx context.Context `gorm:"-"`

	ID        uint32                `gorm:"column:id;primaryKey;autoIncrement" json:"id,omitempty"`
	CreatedAt time.Time             `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at,omitempty"`
	UpdatedAt time.Time             `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at,omitempty"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;" json:"deleted_at,omitempty"`
}

func (u *BaseModel) GetID() uint32 {
	if u == nil {
		return 0
	}
	return u.ID
}

func (u *BaseModel) GetCreatedAt() time.Time {
	if u == nil {
		return time.Time{}
	}
	return u.CreatedAt
}

func (u *BaseModel) GetUpdatedAt() time.Time {
	if u == nil {
		return time.Time{}
	}
	return u.UpdatedAt
}

func (u *BaseModel) GetDeletedAt() soft_delete.DeletedAt {
	if u == nil {
		return 0
	}
	return u.DeletedAt
}

// WithContext set context
func (u *BaseModel) WithContext(ctx context.Context) {
	u.ctx = ctx
}

// GetContext get context
func (u *BaseModel) GetContext() context.Context {
	if u.ctx == nil {
		panic("context is nil")
	}
	return u.ctx
}

var _ Creator = (*CreatorModel)(nil)

type CreatorModel struct {
	BaseModel
	CreatorID uint32 `gorm:"column:creator;type:int unsigned;not null;comment:创建者" json:"creator_id,omitempty"`
}

func (u *CreatorModel) GetCreatorMember() TeamMember {
	if u == nil {
		return nil
	}
	return getTeamMember(u.GetCreatorID())
}

func (u *CreatorModel) GetCreatorID() uint32 {
	if u == nil {
		return 0
	}
	return u.CreatorID
}

func (u *CreatorModel) GetCreator() (user User) {
	defer func() {
		fmt.Println("get creator", user)
	}()
	if u == nil {
		return nil
	}
	return getUser(u.GetCreatorID())
}

func (u *CreatorModel) BeforeCreate(tx *gorm.DB) (err error) {
	var exist bool
	u.CreatorID, exist = permission.GetUserIDByContext(u.GetContext())
	if !exist || u.CreatorID == 0 {
		return merr.ErrorInternalServer("user id not found")
	}
	tx.WithContext(u.GetContext())
	return
}

var _ TeamBase = (*TeamModel)(nil)

type TeamModel struct {
	CreatorModel
	TeamID uint32 `gorm:"column:team_id;type:int unsigned;not null;comment:团队ID" json:"team_id,omitempty"`
}

func (u *TeamModel) GetTeam() Team {
	if u == nil {
		return nil
	}
	return getTeam(u.GetTeamID())
}

func (u *TeamModel) GetTeamID() uint32 {
	if u == nil {
		return 0
	}
	return u.TeamID
}

func (u *TeamModel) BeforeCreate(tx *gorm.DB) (err error) {
	var exist bool
	if u.TeamID <= 0 {
		u.TeamID, exist = permission.GetTeamIDByContext(u.GetContext())
		if !exist || u.TeamID == 0 {
			return merr.ErrorInternalServer("team id not found")
		}
	}
	u.CreatorID, exist = permission.GetUserIDByContext(u.GetContext())
	if !exist || u.CreatorID == 0 {
		return merr.ErrorInternalServer("user id not found")
	}
	tx.WithContext(u.GetContext())
	return
}

func HasTable(teamId uint32, tx *gorm.DB, tableName string) bool {
	if hasTable == nil {
		if tx == nil {
			return false
		}
		if !tx.Migrator().HasTable(tableName) {
			return false
		}
		if cacheTableFlag != nil {
			_ = cacheTableFlag(teamId, tableName)
		}
		return true
	}
	return hasTable(teamId, tableName)
}

func CreateTable(teamId uint32, tx *gorm.DB, tableName string, model any) error {
	if err := tx.Table(tableName).AutoMigrate(model); err != nil {
		return err
	}
	if cacheTableFlag == nil {
		return nil
	}
	return cacheTableFlag(teamId, tableName)
}

// GetPreviousMonday 返回给定日期所在周的周一
func GetPreviousMonday(t time.Time) time.Time {
	// 计算与周一的偏移量
	offset := int(time.Monday - t.Weekday())
	if offset > 0 { // 如果当天是周日(Weekday=0)，则 offset=1，需要减去7天
		offset -= 7
	}
	return t.AddDate(0, 0, offset)
}
